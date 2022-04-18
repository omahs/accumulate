package chain

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"

	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/logging"
	"gitlab.com/accumulatenetwork/accumulate/internal/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	"gitlab.com/accumulatenetwork/accumulate/smt/storage"
)

type stateOperation interface {
	// Execute executes the operation and returns any chains that should be
	// created via a synthetic transaction.
	Execute(*stateCache) ([]protocol.Account, error)
}

type createRecords struct {
	records []protocol.Account
}

// Create queues a record for a synthetic chain create transaction. Will panic
// if called by a synthetic transaction. Will panic if the record is a
// transaction.
func (m *stateCache) Create(record ...protocol.Account) {
	if m.txType.IsSynthetic() {
		panic("Called StateManager.Create from a synthetic transaction!")
	}
	for _, r := range record {
		m.chains[r.GetUrl().AccountID32()] = r
	}

	m.operations = append(m.operations, &createRecords{record})
}

func (op *createRecords) Execute(st *stateCache) ([]protocol.Account, error) {
	return op.records, nil
}

type updateRecord struct {
	url    *url.URL
	record protocol.Account
}

// Update queues a record for storage in the database. The queued update will
// fail if the record does not already exist, unless it is created by a
// synthetic transaction, or the record is a transaction.
func (m *stateCache) Update(record ...protocol.Account) {
	for _, r := range record {
		m.chains[r.GetUrl().AccountID32()] = r
		m.operations = append(m.operations, &updateRecord{r.GetUrl(), r})
	}
}

func (op *updateRecord) Execute(st *stateCache) ([]protocol.Account, error) {
	// Update: update an existing record. Non-synthetic transactions are
	// not allowed to create accounts, so we must check if the record
	// already exists. The record may have been added to the DB
	// transaction already, so in order to actually know if the record
	// exists on disk, we have to use GetPersistentEntry.

	rec := st.batch.Account(op.url)
	_, err := rec.GetState()
	switch {
	case err == nil:
		// If the record already exists, update it

	case !errors.Is(err, storage.ErrNotFound):
		// Handle unexpected errors
		return nil, fmt.Errorf("failed to check for an existing record: %v", err)

	case st.txType.IsSynthetic() || st.txType.IsInternal():
		// Synthetic and internal transactions are allowed to create accounts

	default:
		// Non-synthetic transactions are NOT allowed to create accounts
		return nil, fmt.Errorf("cannot create an account in a non-synthetic transaction")
	}

	record := st.batch.Account(op.url)
	err = record.PutState(op.record)
	if err != nil {
		return nil, fmt.Errorf("failed to update state of %q: %v", op.url, err)
	}

	return nil, st.state.ChainUpdates.AddChainEntry(st.batch, op.url, protocol.MainChain, protocol.ChainTypeTransaction, st.txHash[:], 0, 0)
}

type updateTxStatus struct {
	txid   []byte
	status *protocol.TransactionStatus
}

func (m *StateManager) UpdateStatus(txid []byte, status *protocol.TransactionStatus) {
	m.operations = append(m.operations, &updateTxStatus{txid, status})
}

func (u *updateTxStatus) Execute(st *stateCache) ([]protocol.Account, error) {
	// Load the transaction status & state
	tx := st.batch.Transaction(u.txid)
	curStatus, err := tx.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("retrieving status for transaction %s failed: %w", logging.AsHex(u.txid), err)
	}

	// Only update the status when changed
	if curStatus != nil && statusEqual(curStatus, u.status) {
		return nil, nil
	}

	err = tx.PutStatus(u.status)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

type updateSignator struct {
	url    *url.URL
	record protocol.Account
}

func (m *stateCache) UpdateSignator(record protocol.Account) error {
	// Load the previous state of the record
	rec := m.batch.Account(record.GetUrl())
	old, err := rec.GetState()
	if err != nil {
		return fmt.Errorf("failed to load state for %q", record.GetUrl())
	}

	// Check that the nonce is the only thing that changed
	switch record.Type() {
	case protocol.AccountTypeLiteTokenAccount:
		old, new := old.(*protocol.LiteTokenAccount), record.(*protocol.LiteTokenAccount)
		old.LastUsedOn = new.LastUsedOn
		old.CreditBalance = new.CreditBalance
		if !old.Equal(new) {
			return fmt.Errorf("attempted to change more than the nonce and the credit balance")
		}

	case protocol.AccountTypeKeyPage:
		old, new := old.(*protocol.KeyPage), record.(*protocol.KeyPage)
		old.CreditBalance = new.CreditBalance
		for i := 0; i < len(old.Keys) && i < len(new.Keys); i++ {
			old.Keys[i].LastUsedOn = new.Keys[i].LastUsedOn
		}
		if !old.Equal(new) {
			return fmt.Errorf("attempted to change more than a nonce and the credit balance")
		}

	default:
		return fmt.Errorf("account type %d is not a signator", old.Type())
	}

	m.chains[record.GetUrl().AccountID32()] = record
	m.operations = append(m.operations, &updateSignator{record.GetUrl(), record})
	return nil
}

func (op *updateSignator) Execute(st *stateCache) ([]protocol.Account, error) {
	record := st.batch.Account(op.url)
	err := record.PutState(op.record)
	if err != nil {
		return nil, fmt.Errorf("failed to update state of %q: %v", op.url, err)
	}

	return nil, st.state.ChainUpdates.AddChainEntry(st.batch, op.url, protocol.SignatureChain, protocol.ChainTypeTransaction, st.txHash[:], 0, 0)
}

type addDataEntry struct {
	url          *url.URL
	liteStateRec protocol.Account
	hash         []byte
	entry        *protocol.DataEntry
}

//UpdateData will cache a data associated with a DataAccount chain.
//the cache data will not be stored directly in the state but can be used
//upstream for storing a chain in the state database.
func (m *stateCache) UpdateData(record protocol.Account, entryHash []byte, dataEntry *protocol.DataEntry) {
	var stateRec protocol.Account

	if record.Type() == protocol.AccountTypeLiteDataAccount {
		stateRec = record
	}

	m.operations = append(m.operations, &addDataEntry{record.GetUrl(), stateRec, entryHash, dataEntry})
}

func (op *addDataEntry) Execute(st *stateCache) ([]protocol.Account, error) {
	// Add entry to data chain
	record := st.batch.Account(op.url)

	// Add lite record to data chain if applicable
	if op.liteStateRec != nil {
		_, err := record.GetState()
		if err != nil {
			//if we have no state, store it
			err = record.PutState(op.liteStateRec)
			if err != nil {
				return nil, err
			}
		}
	}

	// Add entry to data chain
	data, err := record.Data()
	if err != nil {
		return nil, fmt.Errorf("failed to load data chain of %q: %v", op.url, err)
	}

	index := data.Height()
	err = data.Put(op.hash, op.entry)
	if err != nil {
		return nil, fmt.Errorf("failed to add entry to data chain of %q: %v", op.url, err)
	}

	err = st.state.ChainUpdates.DidAddChainEntry(st.batch, op.url, protocol.DataChain, protocol.ChainTypeData, op.hash, uint64(index), 0, 0)
	if err != nil {
		return nil, err
	}

	// Add TX to main chain
	return nil, st.state.ChainUpdates.AddChainEntry(st.batch, op.url, protocol.MainChain, protocol.ChainTypeTransaction, st.txHash[:], 0, 0)
}

type addChainEntryOp struct {
	account     *url.URL
	name        string
	typ         protocol.ChainType
	entry       []byte
	sourceIndex uint64
	sourceBlock uint64
}

func (m *stateCache) AddChainEntry(u *url.URL, name string, typ protocol.ChainType, entry []byte, sourceIndex, sourceBlock uint64) error {
	// The main and pending chain cannot be updated this way
	switch strings.ToLower(name) {
	case protocol.MainChain, protocol.SignatureChain:
		return fmt.Errorf("invalid operation: cannot update %s chain with AddChainEntry", name)
	}

	// Check if the chain is valid
	_, err := m.batch.Account(u).Chain(name, typ)
	if err != nil {
		return fmt.Errorf("failed to load %s#chain/%s: %v", u, name, err)
	}

	m.operations = append(m.operations, &addChainEntryOp{u, name, typ, entry, sourceIndex, sourceBlock})
	return nil
}

func (op *addChainEntryOp) Execute(st *stateCache) ([]protocol.Account, error) {
	return nil, st.state.ChainUpdates.AddChainEntry(st.batch, op.account, op.name, op.typ, op.entry, op.sourceIndex, op.sourceBlock)
}

type writeIndex struct {
	index *database.Value
	put   bool
	value []byte
}

func (c *stateCache) RecordIndex(u *url.URL, key ...interface{}) *writeIndex {
	return c.getIndex(c.batch.Account(u).Index(key...))
}

func (c *stateCache) TxnIndex(id []byte, key ...interface{}) *writeIndex {
	return c.getIndex(c.batch.Transaction(id).Index(key...))
}

func (c *stateCache) getIndex(i *database.Value) *writeIndex {
	op, ok := c.indices[i.Key()]
	if ok {
		return op
	}

	op = &writeIndex{i, false, nil}
	c.indices[i.Key()] = op
	c.operations = append(c.operations, op)
	return op
}

func (op *writeIndex) Get() ([]byte, error) {
	if !op.put {
		return op.index.Get()
	}
	return op.value, nil
}

func (op *writeIndex) Put(data []byte) error {
	op.value, op.put = data, true
	return nil
}

func (op *writeIndex) Execute(st *stateCache) ([]protocol.Account, error) {
	if !op.put {
		return nil, nil
	}
	return nil, op.index.Put(op.value)
}

type signTransaction struct {
	txid       []byte
	signatures []protocol.Signature
}

func (m *stateCache) SignTransaction(txid []byte, signatures ...protocol.Signature) {
	m.operations = append(m.operations, &signTransaction{
		txid:       txid,
		signatures: signatures,
	})
}

func (op *signTransaction) Execute(st *stateCache) ([]protocol.Account, error) {
	t := st.batch.Transaction(op.txid)
	for _, sig := range op.signatures {
		// Add the signature
		_, err := t.AddSignature(sig)
		if err != nil {
			return nil, err
		}

		// Ensure it's stored locally. This shouldn't be necessary, but *shrug*
		// it works.
		data, err := sig.MarshalBinary()
		if err != nil {
			return nil, err
		}
		hash := sha256.Sum256(data)
		err = st.batch.Transaction(hash[:]).PutState(&database.SigOrTxn{Signature: sig, Hash: *(*[32]byte)(op.txid)})
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

type addSyntheticTxns struct {
	txid  []byte
	synth [32]byte
}

func (m *stateCache) AddSyntheticTxn(txid []byte, synth [32]byte) {
	m.operations = append(m.operations, &addSyntheticTxns{
		txid:  txid,
		synth: synth,
	})
}

func (op *addSyntheticTxns) Execute(st *stateCache) ([]protocol.Account, error) {
	return nil, st.batch.Transaction(op.txid).AddSyntheticTxns(op.synth)
}
