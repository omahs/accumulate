package chain

import (
	"bytes"
	"crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
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
)

type Executor struct {
	ExecutorOptions

	executors map[protocol.TransactionType]TxExecutor
	governor  *governor
	logger    log.Logger

	blockMeta  BlockMeta
	blockState BlockState
	blockBatch *database.Batch

	// oldBlockMeta blockMetadata
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
	m.executors = map[protocol.TransactionType]TxExecutor{}

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
	var ledger *protocol.InternalLedger
	err := batch.Account(m.Network.NodeUrl(protocol.Ledger)).GetStateAs(&ledger)
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

	m.blockMeta = BlockMeta{
		Index: protocol.GenesisBlock,
		Time:  time,
	}
	m.blockState = BlockState{}
	m.blockBatch = m.DB.Begin(true)
	defer m.blockBatch.Discard()

	env := new(protocol.Envelope)
	env.Transaction = new(protocol.Transaction)
	env.Transaction.Header.Principal = protocol.AcmeUrl()
	env.Transaction.Body = new(protocol.InternalGenesis)
	env.Signatures = []protocol.Signature{&protocol.InternalSignature{Network: m.Network.NodeUrl()}}

	st, err := NewStateManager(m.blockBatch, m.Network.NodeUrl(), env)
	if err != nil {
		return nil, err
	}
	defer st.Discard()

	if st.Origin != nil {
		return nil, errors.New("already initialized")
	}
	st.logger.L = m.logger

	status := &protocol.TransactionStatus{Delivered: true}
	err = m.blockBatch.Transaction(env.GetTxHash()).Put(env.Transaction, status, nil)
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

	err = st.Commit()
	if err != nil {
		return nil, err
	}

	m.blockState.Merge(&st.blockState)

	// Process synthetic transactions generated by the validator
	produced := st.blockState.ProducedTxns
	st.Reset()
	err = m.addSynthTxns(&st.stateCache, produced)
	if err != nil {
		return nil, err
	}
	err = st.Commit()
	if err != nil {
		return nil, err
	}

	return m.Commit()
}

func (m *Executor) InitChain(data []byte, time time.Time) ([]byte, error) {
	if m.isGenesis {
		panic("Cannot call InitChain on a genesis txn executor")
	}

	// Check if InitChain already happened
	var anchor []byte
	var err error
	err = m.DB.View(func(batch *database.Batch) error {
		anchor, err = batch.GetMinorRootChainAnchor(&m.Network)
		return err
	})
	if err != nil {
		return nil, err
	}
	if len(anchor) > 0 {
		return anchor, nil
	}

	// Load the genesis state (JSON) into an in-memory key-value store
	src := memory.New(nil)
	err = src.UnmarshalJSON(data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal app state: %v", err)
	}

	// Load the root anchor chain so we can verify the system state
	srcBatch := database.New(src, nil).Begin(false)
	defer srcBatch.Discard()
	srcAnchor, err := srcBatch.GetMinorRootChainAnchor(&m.Network)
	if err != nil {
		return nil, fmt.Errorf("failed to load root anchor chain from app state: %v", err)
	}

	// Dump the genesis state into the key-value store
	batch := m.DB.Begin(true)
	defer batch.Discard()
	err = batch.Import(src)
	if err != nil {
		return nil, fmt.Errorf("failed to import database: %v", err)
	}

	// Commit the database batch
	err = batch.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to load app state into database: %v", err)
	}

	// Recreate the batch to reload the BPT
	batch = m.DB.Begin(false)
	defer batch.Discard()

	anchor, err = batch.GetMinorRootChainAnchor(&m.Network)
	if err != nil {
		return nil, err
	}

	// Make sure the database BPT root hash matches what we found in the genesis state
	if !bytes.Equal(srcAnchor, anchor) {
		panic(fmt.Errorf("Root chain anchor from state DB does not match the app state\nWant: %X\nGot:  %X", srcAnchor, anchor))
	}

	err = m.governor.DidCommit(batch, true, BlockMeta{Index: protocol.GenesisBlock, Time: time, IsLeader: true}, BlockState{}, nil)
	if err != nil {
		return nil, err
	}

	return anchor, nil
}

// BeginBlock implements ./abci.Chain
func (m *Executor) BeginBlock(req abci.BeginBlockRequest) (resp abci.BeginBlockResponse, err error) {
	m.logDebug("Begin block", "height", req.Height, "leader", req.IsLeader, "time", req.Time)

	// Set the block metadata
	m.blockMeta = BlockMeta{
		IsLeader: req.IsLeader,
		Index:    req.Height,
		Time:     req.Time,
	}

	// Reset the block state variable
	m.blockState = BlockState{}

	// Create a new batch for the block
	m.blockBatch = m.DB.Begin(true)

	defer func() {
		if err != nil {
			m.blockBatch.Discard()
		}
	}()

	// Reset the block state
	err = indexing.BlockState(m.blockBatch, m.Network.NodeUrl(protocol.Ledger)).Clear()
	if err != nil {
		return abci.BeginBlockResponse{}, nil
	}

	// Load the ledger state
	ledger := m.blockBatch.Account(m.Network.NodeUrl(protocol.Ledger))
	var ledgerState *protocol.InternalLedger
	err = ledger.GetStateAs(&ledgerState)
	switch {
	case err == nil:
		// Make sure the block index is increasing
		if ledgerState.Index >= m.blockMeta.Index {
			panic(fmt.Errorf("Current height is %d but the next block height is %d!", ledgerState.Index, m.blockMeta.Index))
		}

	case m.isGenesis && errors.Is(err, storage.ErrNotFound):
		// OK

	default:
		return abci.BeginBlockResponse{}, fmt.Errorf("cannot load ledger: %w", err)
	}

	// Reset transient values
	ledgerState.Index = m.blockMeta.Index
	ledgerState.Timestamp = m.blockMeta.Time

	err = ledger.PutState(ledgerState)
	if err != nil {
		return abci.BeginBlockResponse{}, fmt.Errorf("cannot write ledger: %w", err)
	}

	//store votes from previous block, choosing to marshal as json to make it easily viewable by explorers
	data, err := json.Marshal(req.CommitInfo)
	if err != nil {
		m.logger.Error("cannot marshal voting info data as json")
	} else {
		wd := protocol.WriteData{}
		wd.Entry.Data = append(wd.Entry.Data, data)

		err := m.processInternalDataTransaction(protocol.Votes, &wd)
		if err != nil {
			m.logger.Error(fmt.Sprintf("error processing internal vote transaction, %v", err))
		}
	}

	//capture evidence of maleficence if any occurred
	if req.Evidence != nil {
		data, err := json.Marshal(req.Evidence)
		if err != nil {
			m.logger.Error("cannot marshal evidence as json")
		} else {
			wd := protocol.WriteData{}
			wd.Entry.Data = append(wd.Entry.Data, data)

			err := m.processInternalDataTransaction(protocol.Evidence, &wd)
			if err != nil {
				m.logger.Error(fmt.Sprintf("error processing internal evidence transaction, %v", err))
			}
		}
	}

	return abci.BeginBlockResponse{}, nil
}

// EndBlock implements ./abci.Chain
func (m *Executor) EndBlock(req abci.EndBlockRequest) abci.EndBlockResponse {
	return abci.EndBlockResponse{
		ValidatorsUpdates: m.blockState.ValidatorsUpdates,
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
	var ledgerState *protocol.InternalLedger
	err := ledger.GetStateAs(&ledgerState)
	if err != nil {
		return nil, err
	}

	//set active oracle from pending
	ledgerState.ActiveOracle = ledgerState.PendingOracle

	var anchorTxn *protocol.SyntheticAnchor
	if !force && m.blockState.Empty() {
		m.logInfo("Committed empty transaction")
		m.blockBatch.Discard()
	} else {
		m.logInfo("Committing",
			"height", m.blockMeta.Index,
			"delivered", m.blockState.Delivered,
			"signed", m.blockState.SynthSigned,
			"sent", m.blockState.SynthSent,
			"updated", len(m.blockState.ChainUpdates),
			"submitted", len(m.blockState.ProducedTxns))
		t := time.Now()

		anchorTxn, err = m.doCommit(ledgerState)
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

		m.logInfo("Committed", "height", m.blockMeta.Index, "duration", time.Since(t))
	}

	// Get a clean batch
	batch := m.DB.Begin(false)
	defer batch.Discard()

	if !m.isGenesis {
		err := m.governor.DidCommit(batch, false, m.blockMeta, m.blockState, anchorTxn)
		if err != nil {
			return nil, err
		}
	}

	//return anchor from minor root anchor chain
	anchor, err := batch.GetMinorRootChainAnchor(&m.Network)
	if err != nil {
		return nil, err
	}

	return anchor, nil
}

// updateOraclePrice reads the oracle from the oracle account and updates the
// value on the ledger.
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
	if e.Data == nil {
		return fmt.Errorf("no data in oracle data account")
	}
	err = json.Unmarshal(e.Data[0], &o)
	if err != nil {
		return fmt.Errorf("cannot unmarshal oracle data entry %x", e.Data)
	}

	if o.Price == 0 {
		return fmt.Errorf("invalid oracle price, must be > 0")
	}

	ledgerState.PendingOracle = o.Price
	return nil
}

func (m *Executor) doCommit(ledgerState *protocol.InternalLedger) (*protocol.SyntheticAnchor, error) {
	// Load the main chain of the minor root
	ledgerUrl := m.Network.NodeUrl(protocol.Ledger)
	ledger := m.blockBatch.Account(ledgerUrl)
	rootChain, err := ledger.Chain(protocol.MinorRootChain, protocol.ChainTypeAnchor)
	if err != nil {
		return nil, err
	}

	// Pending transaction-chain index entries
	type txChainIndexEntry struct {
		indexing.TransactionChainEntry
		Txid []byte
	}
	txChainEntries := make([]*txChainIndexEntry, 0, len(m.blockState.ChainUpdates))

	// Process chain updates
	accountSeen := map[string]bool{}
	for _, u := range m.blockState.ChainUpdates {
		// Do not create root chain or BPT entries for the ledger
		if ledgerUrl.Equal(u.Account) {
			continue
		}

		// Anchor and index the chain
		m.logDebug("Updated a chain", "url", fmt.Sprintf("%s#chain/%s", u.Account, u.Name))
		account := m.blockBatch.Account(u.Account)
		indexIndex, didIndex, err := addChainAnchor(rootChain, account, u.Account, u.Name, u.Type)
		if err != nil {
			return nil, err
		}

		// Once for each account
		s := strings.ToLower(u.Account.String())
		if !accountSeen[s] {
			accountSeen[s] = true
			err = m.updateAccountBPT(account)
			if err != nil {
				return nil, err
			}
		}

		// Add a pending transaction-chain index update
		if didIndex && u.Type == protocol.ChainTypeTransaction {
			e := new(txChainIndexEntry)
			e.Txid = u.Entry
			e.Account = u.Account
			e.Chain = u.Name
			e.ChainIndex = uint64(indexIndex)
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
	var synthIndexIndex uint64
	if len(m.blockState.ProducedTxns) > 0 {
		synthIndexIndex, err = m.anchorSynthChain(ledger, ledgerUrl, ledgerState, rootChain)
		if err != nil {
			return nil, err
		}
	}

	// Add the BPT to the root chain
	err = m.anchorBPT(ledgerState, rootChain)
	if err != nil {
		return nil, err
	}

	// Index the root chain
	rootIndexIndex, err := addIndexChainEntry(ledger, protocol.MinorRootIndexChain, &protocol.IndexEntry{
		Source:     uint64(rootChain.Height() - 1),
		BlockIndex: uint64(m.blockMeta.Index),
		BlockTime:  &m.blockMeta.Time,
	})
	if err != nil {
		return nil, err
	}

	// Update the transaction-chain index
	for _, e := range txChainEntries {
		e.AnchorIndex = rootIndexIndex
		err = indexing.TransactionChain(m.blockBatch, e.Txid).Add(&e.TransactionChainEntry)
		if err != nil {
			return nil, err
		}
	}

	// Add transaction-chain index entries for synthetic transactions
	blockState, err := indexing.BlockState(m.blockBatch, ledgerUrl).Get()
	if err != nil {
		return nil, err
	}

	for _, e := range blockState.ProducedSynthTxns {
		err = indexing.TransactionChain(m.blockBatch, e.Transaction).Add(&indexing.TransactionChainEntry{
			Account:     ledgerUrl,
			Chain:       protocol.SyntheticChain,
			ChainIndex:  synthIndexIndex,
			AnchorIndex: rootIndexIndex,
		})
		if err != nil {
			return nil, err
		}
	}

	err = m.buildSynthReceipts(rootChain.Anchor(), int64(synthIndexIndex), int64(rootIndexIndex))
	if err != nil {
		return nil, err
	}

	return m.buildAnchorTxn(ledgerState, rootChain)
}

// updateAccountBPT updates the BPT entry of an account.
func (m *Executor) updateAccountBPT(account *database.Account) (err error) {
	// Load the state
	entry, err := account.StateHash()
	if err != nil {
		return err
	}

	account.PutBpt(*(*[32]byte)(entry))

	return nil
}

// anchorSynthChain anchors the synthetic transaction chain.
func (m *Executor) anchorSynthChain(ledger *database.Account, ledgerUrl *url.URL, ledgerState *protocol.InternalLedger, rootChain *database.Chain) (indexIndex uint64, err error) {
	indexIndex, _, err = addChainAnchor(rootChain, ledger, ledgerUrl, protocol.SyntheticChain, protocol.ChainTypeTransaction)
	if err != nil {
		return 0, err
	}

	m.blockState.DidUpdateChain(ChainUpdate{
		Name:    protocol.SyntheticChain,
		Type:    protocol.ChainTypeTransaction,
		Account: ledgerUrl,
		// Index:   uint64(synthChain.Height() - 1),
	})

	return indexIndex, nil
}

// anchorBPT anchors the BPT after ensuring any pending changes have been flushed.
func (m *Executor) anchorBPT(ledgerState *protocol.InternalLedger, rootChain *database.Chain) error {
	root, err := m.blockBatch.CommitBpt()
	if err != nil {
		return err
	}

	m.blockState.DidUpdateChain(ChainUpdate{
		Name:    "bpt",
		Account: m.Network.NodeUrl(),
		Index:   uint64(m.blockMeta.Index - 1),
	})

	return rootChain.AddEntry(root, false)
}

// buildSynthReceipts builds partial receipts for produced synthetic
// transactions and stores them in the synthetic transaction ledger.
func (m *Executor) buildSynthReceipts(rootAnchor []byte, synthIndexIndex, rootIndexIndex int64) error {
	ledger := m.blockBatch.Account(m.Network.SyntheticLedger())
	ledgerState := new(protocol.InternalSyntheticLedger)
	err := ledger.GetStateAs(&ledgerState)
	if err != nil {
		return fmt.Errorf("unable to load the synthetic transaction ledger: %w", err)
	}

	synthChain, err := m.blockBatch.Account(m.Network.Ledger()).ReadChain(protocol.SyntheticChain)
	if err != nil {
		return fmt.Errorf("unable to load synthetic transaction chain: %w", err)
	}

	height := synthChain.Height()
	offset := height - int64(len(m.blockState.ProducedTxns))
	for i, txn := range m.blockState.ProducedTxns {
		if txn.Type() == protocol.TransactionTypeSyntheticAnchor || txn.Type() == protocol.TransactionTypeSyntheticMirror {
			// Do not generate a receipt for the anchor
			continue
		}

		entry := new(protocol.SyntheticLedgerEntry)
		entry.TransactionHash = *(*[32]byte)(txn.GetHash())
		entry.RootAnchor = *(*[32]byte)(rootAnchor)
		entry.SynthIndex = uint64(offset) + uint64(i)
		entry.RootIndexIndex = uint64(rootIndexIndex)
		entry.SynthIndexIndex = uint64(synthIndexIndex)
		entry.NeedsReceipt = true
		ledgerState.Pending = append(ledgerState.Pending, entry)
		m.logDebug("Adding synthetic transaction to the ledger", "hash", logging.AsHex(txn.GetHash()), "type", txn.Type(), "anchor", logging.AsHex(rootAnchor), "module", "synthetic")
	}

	err = ledger.PutState(ledgerState)
	if err != nil {
		return fmt.Errorf("unable to store the synthetic transaction ledger: %w", err)
	}

	return nil
}

// buildAnchorTxn builds the anchor transaction for the block.
func (m *Executor) buildAnchorTxn(ledger *protocol.InternalLedger, rootChain *database.Chain) (*protocol.SyntheticAnchor, error) {
	txn := new(protocol.SyntheticAnchor)
	txn.Source = m.Network.NodeUrl()
	txn.RootIndex = uint64(rootChain.Height() - 1)
	txn.RootAnchor = *(*[32]byte)(rootChain.Anchor())
	txn.Block = uint64(m.blockMeta.Index)
	txn.AcmeBurnt, ledger.AcmeBurnt = ledger.AcmeBurnt, *big.NewInt(int64(0))
	if m.Network.Type == config.Directory {
		txn.AcmeOraclePrice = ledger.PendingOracle
	}

	// TODO This is pretty inefficient; we're constructing a receipt for every
	// anchor. If we were more intelligent about it, we could send just the
	// Merkle state and a list of transactions, though we would need that for
	// the root chain and each anchor chain.

	anchorUrl := m.Network.NodeUrl(protocol.AnchorPool)
	anchor := m.blockBatch.Account(anchorUrl)
	for _, update := range m.blockState.ChainUpdates {
		// Is it an anchor chain?
		if update.Type != protocol.ChainTypeAnchor {
			continue
		}

		// Does it belong to our anchor pool?
		if !update.Account.Equal(anchorUrl) {
			continue
		}

		indexChain, err := anchor.ReadIndexChain(update.Name, false)
		if err != nil {
			return nil, fmt.Errorf("unable to load minor index chain of intermediate anchor chain %s: %w", update.Name, err)
		}

		from, to, anchorIndex, err := getRangeFromIndexEntry(indexChain, uint64(indexChain.Height())-1)
		if err != nil {
			return nil, fmt.Errorf("unable to load range from minor index chain of intermediate anchor chain %s: %w", update.Name, err)
		}

		rootReceipt, err := rootChain.Receipt(int64(anchorIndex), rootChain.Height()-1)
		if err != nil {
			return nil, fmt.Errorf("unable to build receipt for the root chain: %w", err)
		}

		anchorChain, err := anchor.ReadChain(update.Name)
		if err != nil {
			return nil, fmt.Errorf("unable to load intermediate anchor chain %s: %w", update.Name, err)
		}

		for i := from; i <= to; i++ {
			anchorReceipt, err := anchorChain.Receipt(int64(i), int64(to))
			if err != nil {
				return nil, fmt.Errorf("unable to build receipt for intermediate anchor chain %s: %w", update.Name, err)
			}

			receipt, err := anchorReceipt.Combine(rootReceipt)
			if err != nil {
				return nil, fmt.Errorf("unable to build receipt for intermediate anchor chain %s: %w", update.Name, err)
			}

			r := protocol.ReceiptFromManaged(receipt)
			txn.Receipts = append(txn.Receipts, *r)
			m.logDebug("Build receipt for an anchor", "chain", update.Name, "anchor", logging.AsHex(r.Start), "block", m.blockMeta.Index, "height", i, "module", "synthetic")
		}
	}

	return txn, nil
}
