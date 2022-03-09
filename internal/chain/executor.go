package chain

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tendermint/tendermint/libs/log"
	"gitlab.com/accumulatenetwork/accumulate/config"
	"gitlab.com/accumulatenetwork/accumulate/internal/abci"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/indexing"
	"gitlab.com/accumulatenetwork/accumulate/internal/logging"
	"gitlab.com/accumulatenetwork/accumulate/internal/routing"
	"gitlab.com/accumulatenetwork/accumulate/internal/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	"gitlab.com/accumulatenetwork/accumulate/smt/storage"
	"gitlab.com/accumulatenetwork/accumulate/smt/storage/memory"
	"gitlab.com/accumulatenetwork/accumulate/types"
	"gitlab.com/accumulatenetwork/accumulate/types/api/transactions"
)

const chainWGSize = 4

type Executor struct {
	ExecutorOptions

	executors map[types.TxType]TxExecutor
	governor  *governor
	logger    log.Logger

	blockLeader bool
	blockIndex  int64
	blockTime   time.Time
	blockBatch  *database.Batch
	blockMeta   blockMetadata

	validatorsUpdates []abci.ValidatorUpdate
}

var _ abci.Chain = (*Executor)(nil)

type ExecutorOptions struct {
	DB      *database.Database
	Logger  log.Logger
	Key     ed25519.PrivateKey
	Router  routing.Router
	Network config.Network

	isGenesis bool
}

func newExecutor(opts ExecutorOptions, executors ...TxExecutor) (*Executor, error) {
	m := new(Executor)
	m.ExecutorOptions = opts
	m.executors = map[types.TxType]TxExecutor{}

	if opts.Logger != nil {
		m.logger = opts.Logger.With("module", "executor")
	}

	if !m.isGenesis {
		m.governor = newGovernor(opts)
	}

	for _, x := range executors {
		if _, ok := m.executors[x.Type()]; ok {
			panic(fmt.Errorf("duplicate executor for %d", x.Type()))
		}
		m.executors[x.Type()] = x
	}

	batch := m.DB.Begin(false)
	defer batch.Discard()

	var height int64
	ledger := protocol.NewInternalLedger()
	err := batch.Account(m.Network.NodeUrl(protocol.Ledger)).GetStateAs(ledger)
	switch {
	case err == nil:
		height = ledger.Index
	case errors.Is(err, storage.ErrNotFound):
		height = 0
	default:
		return nil, err
	}

	anchor, err := batch.GetMinorRootChainAnchor(&m.Network)
	if err != nil {
		return nil, err
	}

	m.logInfo("Loaded", "height", height, "hash", logging.AsHex(anchor))
	return m, nil
}

func (m *Executor) logDebug(msg string, keyVals ...interface{}) {
	if m.logger != nil {
		m.logger.Debug(msg, keyVals...)
	}
}

func (m *Executor) logInfo(msg string, keyVals ...interface{}) {
	if m.logger != nil {
		m.logger.Info(msg, keyVals...)
	}
}

func (m *Executor) logError(msg string, keyVals ...interface{}) {
	if m.logger != nil {
		m.logger.Error(msg, keyVals...)
	}
}

func (m *Executor) Start() error {
	return m.governor.Start()
}

func (m *Executor) Stop() error {
	return m.governor.Stop()
}

func (m *Executor) Genesis(time time.Time, callback func(st *StateManager) error) ([]byte, error) {
	var err error

	if !m.isGenesis {
		panic("Cannot call Genesis on a node txn executor")
	}

	m.blockIndex = 1
	m.blockTime = time
	m.blockBatch = m.DB.Begin(true)
	defer m.blockBatch.Discard()

	env := new(transactions.Envelope)
	env.Transaction = new(transactions.Transaction)
	env.Transaction.Origin = protocol.AcmeUrl()
	env.Transaction.Body = new(protocol.InternalGenesis)

	st, err := NewStateManager(m.blockBatch, m.Network.NodeUrl(), env)
	if err != nil {
		return nil, err
	}
	if st.Origin != nil {
		return nil, errors.New("already initialized")
	}
	st.logger.L = m.logger

	status := &protocol.TransactionStatus{Delivered: true}
	err = m.blockBatch.Transaction(env.GetTxHash()).Put(env, status, nil)
	if err != nil {
		return nil, err
	}

	err = indexing.BlockState(m.blockBatch, m.Network.NodeUrl(protocol.Ledger)).Clear()
	if err != nil {
		return nil, err
	}

	err = callback(st)
	if err != nil {
		return nil, err
	}

	submitted, err := st.Commit()
	if err != nil {
		return nil, err
	}

	// Process synthetic transactions generated by the validator
	st.Reset()
	err = m.addSynthTxns(&st.stateCache, submitted)
	if err != nil {
		return nil, err
	}
	_, err = st.Commit()
	if err != nil {
		return nil, err
	}

	return m.Commit()
}

func (m *Executor) InitChain(data []byte, time time.Time, blockIndex int64) ([]byte, error) {
	if m.isGenesis {
		panic("Cannot call InitChain on a genesis txn executor")
	}

	// Check if InitChain already happened
	var rootHash []byte
	err := m.DB.View(func(batch *database.Batch) error {
		_, err := batch.Account(m.Network.NodeUrl(protocol.Ledger)).GetState()
		if err != nil {
			return err
		}

		rootHash = batch.BptRootHash()
		return nil
	})
	if err == nil {
		return rootHash, nil
	}
	if !errors.Is(err, storage.ErrNotFound) {
		return nil, err
	}

	// Load the genesis state (JSON) into an in-memory key-value store
	src := memory.New(nil)
	err = src.UnmarshalJSON(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal app state: %v", err)
	}

	// Load the root anchor chain so we can verify the system state
	srcBatch := database.New(src, nil).Begin(true)
	defer srcBatch.Discard()
	srcAnchor, err := srcBatch.GetMinorRootChainAnchor(&m.Network)
	if err != nil {
		return nil, fmt.Errorf("failed to load root anchor chain from app state: %v", err)
	}

	// Dump the genesis state into the key-value store
	batch := m.DB.Begin(true)
	defer batch.Discard()
	batch.Import(src)

	// Commit the database batch
	err = batch.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to load app state into database: %v", err)
	}

	// Recreate the batch to reload the BPT
	batch = m.DB.Begin(true)
	defer batch.Discard()

	anchor, err := batch.GetMinorRootChainAnchor(&m.Network)
	if err != nil {
		return nil, err
	}

	// Make sure the database BPT root hash matches what we found in the genesis state
	if !bytes.Equal(srcAnchor, anchor) {
		panic(fmt.Errorf("Root chain anchor from state DB does not match the app state\nWant: %X\nGot:  %X", srcAnchor, anchor))
	}

	err = m.governor.DidCommit(batch, true, true, blockIndex, time)
	if err != nil {
		return nil, err
	}

	return anchor, nil
}

// BeginBlock implements ./abci.Chain
func (m *Executor) BeginBlock(req abci.BeginBlockRequest) (resp abci.BeginBlockResponse, err error) {
	m.logDebug("Begin block", "height", req.Height, "leader", req.IsLeader, "time", req.Time)

	m.blockLeader = req.IsLeader
	m.blockIndex = req.Height
	m.blockTime = req.Time
	m.blockBatch = m.DB.Begin(true)
	m.blockMeta = blockMetadata{}
	m.validatorsUpdates = m.validatorsUpdates[:0]

	defer func() {
		if err != nil {
			m.blockBatch.Discard()
		}
	}()

	m.governor.DidBeginBlock(req.IsLeader, req.Height, req.Time)

	// Reset the block state
	err = indexing.BlockState(m.blockBatch, m.Network.NodeUrl(protocol.Ledger)).Clear()
	if err != nil {
		return abci.BeginBlockResponse{}, nil
	}

	// Load the ledger state
	ledger := m.blockBatch.Account(m.Network.NodeUrl(protocol.Ledger))
	ledgerState := protocol.NewInternalLedger()
	err = ledger.GetStateAs(ledgerState)
	switch {
	case err == nil:
		// Make sure the block index is increasing
		if ledgerState.Index >= m.blockIndex {
			panic(fmt.Errorf("Current height is %d but the next block height is %d!", ledgerState.Index, m.blockIndex))
		}

	case m.isGenesis && errors.Is(err, storage.ErrNotFound):
		// OK

	default:
		return abci.BeginBlockResponse{}, fmt.Errorf("cannot load ledger: %w", err)
	}

	// Reset transient values
	ledgerState.Index = m.blockIndex
	ledgerState.Timestamp = m.blockTime
	ledgerState.Updates = nil
	ledgerState.Synthetic.Produced = nil

	err = ledger.PutState(ledgerState)
	if err != nil {
		return abci.BeginBlockResponse{}, fmt.Errorf("cannot write ledger: %w", err)
	}

	//store votes from previous block, choosing to marshal as json to make it easily viewable by explorers

	data, err := json.Marshal(&req.CommitInfo)
	if err != nil {
		m.logger.Error("cannot marshal voting info data")
	} else {

		wd := protocol.WriteData{}
		wd.Entry.Data = data

		err := m.processInternalDataTransaction(protocol.Votes, &wd)
		if err != nil {
			m.logger.Error(fmt.Sprintf("error processing internal vote transaction, %v", err))
		}
	}

	return abci.BeginBlockResponse{}, nil
}

// EndBlock implements ./abci.Chain
func (m *Executor) EndBlock(req abci.EndBlockRequest) abci.EndBlockResponse {
	return abci.EndBlockResponse{
		ValidatorsUpdates: m.validatorsUpdates,
	}
}

// Commit implements ./abci.Chain
func (m *Executor) Commit() ([]byte, error) {
	return m.commit(false)
}

func (m *Executor) commit(force bool) ([]byte, error) {
	// Discard changes if commit fails
	defer m.blockBatch.Discard()

	// Load the ledger
	ledger := m.blockBatch.Account(m.Network.NodeUrl(protocol.Ledger))
	ledgerState := protocol.NewInternalLedger()
	err := ledger.GetStateAs(ledgerState)
	if err != nil {
		return nil, err
	}

	//set active oracle from pending
	ledgerState.ActiveOracle = ledgerState.PendingOracle

	// Deduplicate the update list
	updatedMap := make(map[string]bool, len(ledgerState.Updates))
	updatedSlice := make([]protocol.AnchorMetadata, 0, len(ledgerState.Updates))
	for _, u := range ledgerState.Updates {
		s := strings.ToLower(fmt.Sprintf("%s#chain/%s", u.Account, u.Name))
		if updatedMap[s] {
			continue
		}

		updatedSlice = append(updatedSlice, u)
		updatedMap[s] = true
	}
	ledgerState.Updates = updatedSlice

	if !force && m.blockMeta.Empty() && len(updatedSlice) == 0 && len(ledgerState.Synthetic.Produced) == 0 {
		m.logInfo("Committed empty transaction")
		m.blockBatch.Discard()
	} else {
		m.logInfo("Committing", "height", m.blockIndex, "delivered", m.blockMeta.Delivered, "signed", m.blockMeta.SynthSigned, "sent", m.blockMeta.SynthSent, "updated", len(updatedSlice), "submitted", len(ledgerState.Synthetic.Produced))
		t := time.Now()

		err := m.doCommit(ledgerState)
		if err != nil {
			return nil, err
		}

		// Write the updated ledger
		err = ledger.PutState(ledgerState)
		if err != nil {
			return nil, err
		}

		err = m.blockBatch.Commit()
		if err != nil {
			return nil, err
		}

		m.logInfo("Committed", "height", m.blockIndex, "duration", time.Since(t))
	}

	// Get a clean batch
	batch := m.DB.Begin(false)
	defer batch.Discard()

	if !m.isGenesis {
		err := m.governor.DidCommit(batch, m.blockLeader, false, m.blockIndex, m.blockTime)
		if err != nil {
			return nil, err
		}
	}

	//return anchor from minor root anchor chain
	anchor, err := m.blockBatch.GetMinorRootChainAnchor(&m.Network)
	if err != nil {
		return nil, err
	}

	return anchor, nil
}

func (m *Executor) updateOraclePrice(ledgerState *protocol.InternalLedger) error {
	data, err := m.blockBatch.Account(protocol.PriceOracle()).Data()
	if err != nil {
		return fmt.Errorf("cannot retrieve oracle data entry: %v", err)
	}
	_, e, err := data.GetLatest()
	if err != nil {
		return fmt.Errorf("cannot retrieve latest oracle data entry: data batch at height %d: %v", data.Height(), err)
	}

	o := protocol.AcmeOracle{}
	err = json.Unmarshal(e.Data, &o)
	if err != nil {
		return fmt.Errorf("cannot unmarshal oracle data entry %x", e.Data)
	}

	if o.Price == 0 {
		return fmt.Errorf("invalid oracle price, must be > 0")
	}

	ledgerState.PendingOracle = o.Price
	return nil
}

func (m *Executor) doCommit(ledgerState *protocol.InternalLedger) error {
	// Load the main chain of the minor root
	ledgerUrl := m.Network.NodeUrl(protocol.Ledger)
	ledger := m.blockBatch.Account(ledgerUrl)
	rootChain, err := ledger.Chain(protocol.MinorRootChain, protocol.ChainTypeAnchor)
	if err != nil {
		return err
	}

	// Pending transaction-chain index entries
	type txChainIndexEntry struct {
		indexing.TransactionChainEntry
		Txid []byte
	}
	txChainEntries := make([]*txChainIndexEntry, 0, len(ledgerState.Updates))

	// Process chain updates
	accountSeen := map[string]bool{}
	updates := ledgerState.Updates
	ledgerState.Updates = make([]protocol.AnchorMetadata, 0, len(updates))
	for _, u := range updates {
		// Do not create root chain or BPT entries for the ledger
		if ledgerUrl.Equal(u.Account) {
			continue
		}

		ledgerState.Updates = append(ledgerState.Updates, u)
		m.logDebug("Updated a chain", "url", fmt.Sprintf("%s#chain/%s", u.Account, u.Name))

		indexIndex, didIndex, err := m.commitChainUpdate(&u, rootChain, accountSeen)
		if err != nil {
			return err
		}

		// Add a pending transaction-chain index update
		if didIndex && u.Type == protocol.ChainTypeTransaction {
			e := new(txChainIndexEntry)
			e.Txid = u.Entry
			e.Account = u.Account
			e.Chain = u.Name
			e.ChainIndex = uint64(indexIndex)
			// e.Block = uint64(m.blockIndex)
			// e.ChainEntry = u.Index
			// e.ChainAnchor = uint64(accountChain.Height()) - 1
			// e.RootEntry = uint64(rootIndex)
			txChainEntries = append(txChainEntries, e)
		}
	}

	// If dn/oracle was updated, update the ledger's oracle value, but only if
	// we're on the DN - mirroring can cause dn/oracle to be updated on the BVN
	if accountSeen[protocol.PriceOracleAuthority] && m.Network.LocalSubnetID == protocol.Directory {
		// If things go south here, don't return and error, instead, just log one
		err := m.updateOraclePrice(ledgerState)
		if err != nil {
			m.logError(fmt.Sprintf("%v", err))
		}
	}

	// Add the synthetic transaction chain to the root chain
	var synthIndex uint64
	if len(ledgerState.Synthetic.Produced) > 0 {
		synthIndex, err = m.commitSynthChainUpdate(ledger, ledgerUrl, ledgerState, rootChain)
		if err != nil {
			return err
		}
	}

	// Add the BPT to the root chain
	err = m.commitBptUpdate(ledgerState, rootChain)
	if err != nil {
		return err
	}

	// Index the root chain
	rootIndexIndex, err := addIndexChainEntry(ledger, protocol.MinorRootIndexChain, &protocol.IndexEntry{
		Source:     uint64(rootChain.Height() - 1),
		BlockIndex: uint64(m.blockIndex),
		BlockTime:  &m.blockTime,
	})
	if err != nil {
		return err
	}

	// Update the transaction-chain index
	for _, e := range txChainEntries {
		e.AnchorIndex = rootIndexIndex
		err = indexing.TransactionChain(m.blockBatch, e.Txid).Add(&e.TransactionChainEntry)
		if err != nil {
			return err
		}
	}

	// Add transaction-chain index entries for synthetic transactions
	blockState, err := indexing.BlockState(m.blockBatch, ledgerUrl).Get()
	if err != nil {
		return err
	}

	for _, e := range blockState.ProducedSynthTxns {
		err = indexing.TransactionChain(m.blockBatch, e.Transaction).Add(&indexing.TransactionChainEntry{
			Account:     ledgerUrl,
			Chain:       protocol.SyntheticChain,
			ChainIndex:  synthIndex,
			AnchorIndex: rootIndexIndex,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Executor) commitChainUpdate(update *protocol.AnchorMetadata, rootChain *database.Chain, seen map[string]bool) (indexIndex uint64, didIndex bool, err error) {
	// Anchor and index the chain
	account := m.blockBatch.Account(update.Account)
	indexIndex, didIndex, err = addChainAnchor(rootChain, account, update.Account, update.Name, update.Type)
	if err != nil {
		return 0, false, err
	}

	// Once for each account
	s := strings.ToLower(update.Account.String())
	if seen[s] {
		return indexIndex, didIndex, nil
	}
	seen[s] = true

	// Load the state
	state, err := account.GetState()
	if err != nil {
		return 0, false, err
	}

	// Marshal it
	data, err := state.MarshalBinary()
	if err != nil {
		return 0, false, err
	}

	// Hash it
	var hashes []byte
	h := sha256.Sum256(data)
	hashes = append(hashes, h[:]...)

	// Load the object metadata
	objMeta, err := account.GetObject()
	if err != nil {
		return 0, false, err
	}

	// For each chain
	for _, chainMeta := range objMeta.Chains {
		// Load the chain
		recordChain, err := account.ReadChain(chainMeta.Name)
		if err != nil {
			return 0, false, err
		}

		// Get the anchor
		anchor := recordChain.Anchor()
		h := sha256.Sum256(anchor)
		hashes = append(hashes, h[:]...)
	}

	// Write the hash of the hashes to the BPT
	account.PutBpt(sha256.Sum256(hashes))

	return indexIndex, didIndex, nil
}

func (m *Executor) commitSynthChainUpdate(ledger *database.Account, ledgerUrl *url.URL, ledgerState *protocol.InternalLedger, rootChain *database.Chain) (indexIndex uint64, err error) {
	indexIndex, _, err = addChainAnchor(rootChain, ledger, ledgerUrl, protocol.SyntheticChain, protocol.ChainTypeTransaction)
	if err != nil {
		return 0, err
	}

	ledgerState.Updates = append(ledgerState.Updates, protocol.AnchorMetadata{
		ChainMetadata: protocol.ChainMetadata{
			Name: protocol.SyntheticChain,
			Type: protocol.ChainTypeTransaction,
		},
		Account: ledgerUrl,
		// Index:   uint64(synthChain.Height() - 1),
	})

	return indexIndex, nil
}

func (m *Executor) commitBptUpdate(ledgerState *protocol.InternalLedger, rootChain *database.Chain) error {
	m.blockBatch.UpdateBpt()

	ledgerState.Updates = append(ledgerState.Updates, protocol.AnchorMetadata{
		ChainMetadata: protocol.ChainMetadata{
			Name: "bpt",
		},
		Account: m.Network.NodeUrl(),
		Index:   uint64(m.blockIndex - 1),
	})

	return rootChain.AddEntry(m.blockBatch.BptRootHash(), false)
}
