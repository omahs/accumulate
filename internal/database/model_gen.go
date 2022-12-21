// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package database

// GENERATED BY go run ./tools/cmd/gen-model. DO NOT EDIT.

//lint:file-ignore S1008,U1000 generated code

import (
	"encoding/hex"
	"strconv"
	"sync"

	"gitlab.com/accumulatenetwork/accumulate/internal/database/record"
	"gitlab.com/accumulatenetwork/accumulate/internal/errors"
	"gitlab.com/accumulatenetwork/accumulate/internal/logging"
	"gitlab.com/accumulatenetwork/accumulate/pkg/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	"gitlab.com/accumulatenetwork/accumulate/smt/storage"
)

type Batch struct {
	logger      logging.OptionalLogger
	store       record.Store
	done        bool
	writable    bool
	id          string
	nextChildId int64
	parent      *Batch
	kvstore     storage.KeyValueTxn
	bptEntries  map[storage.Key][32]byte

	account        map[accountKey]*Account
	account_mu     sync.RWMutex
	transaction    map[transactionKey]*Transaction
	transaction_mu sync.RWMutex
	systemData     map[systemDataKey]*SystemData
	systemData_mu  sync.RWMutex
}

type accountKey struct {
	Url [32]byte
}

func keyForAccount(url *url.URL) accountKey {
	return accountKey{record.MapKeyUrl(url)}
}

type transactionKey struct {
	Hash [32]byte
}

func keyForTransaction(hash [32]byte) transactionKey {
	return transactionKey{hash}
}

type systemDataKey struct {
	Partition string
}

func keyForSystemData(partition string) systemDataKey {
	return systemDataKey{partition}
}

func (c *Batch) getAccount(url *url.URL) *Account {
	return getOrCreateMap(&c.account, &c.account_mu, keyForAccount(url), func() *Account {
		v := new(Account)
		v.logger = c.logger
		v.store = c.store
		v.key = record.Key{}.Append("Account", url)
		v.parent = c
		v.label = "account" + " " + url.RawString()
		return v
	})
}

func (c *Batch) getTransaction(hash [32]byte) *Transaction {
	return getOrCreateMap(&c.transaction, &c.transaction_mu, keyForTransaction(hash), func() *Transaction {
		v := new(Transaction)
		v.logger = c.logger
		v.store = c.store
		v.key = record.Key{}.Append("Transaction", hash)
		v.parent = c
		v.label = "transaction" + " " + hex.EncodeToString(hash[:])
		return v
	})
}

func (c *Batch) SystemData(partition string) *SystemData {
	return getOrCreateMap(&c.systemData, &c.systemData_mu, keyForSystemData(partition), func() *SystemData {
		v := new(SystemData)
		v.logger = c.logger
		v.store = c.store
		v.key = record.Key{}.Append("SystemData", partition)
		v.parent = c
		v.label = "system data" + " " + partition
		return v
	})
}

func (c *Batch) Resolve(key record.Key) (record.Record, record.Key, error) {
	if len(key) == 0 {
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for batch")
	}

	switch key[0] {
	case "Account":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for batch")
		}
		url, okUrl := key[1].(*url.URL)
		if !okUrl {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for batch")
		}
		v := c.getAccount(url)
		return v, key[2:], nil
	case "Transaction":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for batch")
		}
		hash, okHash := key[1].([32]byte)
		if !okHash {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for batch")
		}
		v := c.getTransaction(hash)
		return v, key[2:], nil
	case "SystemData":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for batch")
		}
		partition, okPartition := key[1].(string)
		if !okPartition {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for batch")
		}
		v := c.SystemData(partition)
		return v, key[2:], nil
	default:
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for batch")
	}
}

func (c *Batch) IsDirty() bool {
	if c == nil {
		return false
	}

	for _, v := range c.account {
		if v.IsDirty() {
			return true
		}
	}
	for _, v := range c.transaction {
		if v.IsDirty() {
			return true
		}
	}
	for _, v := range c.systemData {
		if v.IsDirty() {
			return true
		}
	}

	return false
}

func (c *Batch) baseCommit() error {
	if c == nil {
		return nil
	}

	var err error
	for _, v := range c.account {
		commitField(&err, v)
	}
	for _, v := range c.transaction {
		commitField(&err, v)
	}
	for _, v := range c.systemData {
		commitField(&err, v)
	}

	return err
}

type Account struct {
	logger logging.OptionalLogger
	store  record.Store
	key    record.Key
	label  string
	parent *Batch

	url                       *record.Value[*url.URL]
	url_mu                    sync.RWMutex
	main                      *record.Value[protocol.Account]
	main_mu                   sync.RWMutex
	pending                   *record.Set[*url.TxID]
	pending_mu                sync.RWMutex
	syntheticForAnchor        map[accountSyntheticForAnchorKey]*record.Set[*url.TxID]
	syntheticForAnchor_mu     sync.RWMutex
	directory                 *record.Set[*url.URL]
	directory_mu              sync.RWMutex
	mainChain                 *Chain2
	mainChain_mu              sync.RWMutex
	scratchChain              *Chain2
	scratchChain_mu           sync.RWMutex
	signatureChain            *Chain2
	signatureChain_mu         sync.RWMutex
	rootChain                 *Chain2
	rootChain_mu              sync.RWMutex
	anchorSequenceChain       *Chain2
	anchorSequenceChain_mu    sync.RWMutex
	majorBlockChain           *Chain2
	majorBlockChain_mu        sync.RWMutex
	syntheticSequenceChain    map[accountSyntheticSequenceChainKey]*Chain2
	syntheticSequenceChain_mu sync.RWMutex
	anchorChain               map[accountAnchorChainKey]*AccountAnchorChain
	anchorChain_mu            sync.RWMutex
	chains                    *record.Set[*protocol.ChainMetadata]
	chains_mu                 sync.RWMutex
	syntheticAnchors          *record.Set[[32]byte]
	syntheticAnchors_mu       sync.RWMutex
	data                      *AccountData
	data_mu                   sync.RWMutex
}

type accountSyntheticForAnchorKey struct {
	Anchor [32]byte
}

func keyForAccountSyntheticForAnchor(anchor [32]byte) accountSyntheticForAnchorKey {
	return accountSyntheticForAnchorKey{anchor}
}

type accountSyntheticSequenceChainKey struct {
	Partition string
}

func keyForAccountSyntheticSequenceChain(partition string) accountSyntheticSequenceChainKey {
	return accountSyntheticSequenceChainKey{partition}
}

type accountAnchorChainKey struct {
	Partition string
}

func keyForAccountAnchorChain(partition string) accountAnchorChainKey {
	return accountAnchorChainKey{partition}
}

func (c *Account) getUrl() *record.Value[*url.URL] {
	return getOrCreateField(&c.url, &c.url_mu, func() *record.Value[*url.URL] {
		return record.NewValue(c.logger.L, c.store, c.key.Append("Url"), c.label+" "+"url", false, record.Wrapped(record.UrlWrapper))
	})
}

func (c *Account) Main() *record.Value[protocol.Account] {
	return getOrCreateField(&c.main, &c.main_mu, func() *record.Value[protocol.Account] {
		return record.NewValue(c.logger.L, c.store, c.key.Append("Main"), c.label+" "+"main", false, record.Union(protocol.UnmarshalAccount))
	})
}

func (c *Account) Pending() *record.Set[*url.TxID] {
	return getOrCreateField(&c.pending, &c.pending_mu, func() *record.Set[*url.TxID] {
		return record.NewSet(c.logger.L, c.store, c.key.Append("Pending"), c.label+" "+"pending", record.Wrapped(record.TxidWrapper), record.CompareTxid)
	})
}

func (c *Account) SyntheticForAnchor(anchor [32]byte) *record.Set[*url.TxID] {
	return getOrCreateMap(&c.syntheticForAnchor, &c.syntheticForAnchor_mu, keyForAccountSyntheticForAnchor(anchor), func() *record.Set[*url.TxID] {
		return record.NewSet(c.logger.L, c.store, c.key.Append("SyntheticForAnchor", anchor), c.label+" "+"synthetic for anchor"+" "+hex.EncodeToString(anchor[:]), record.Wrapped(record.TxidWrapper), record.CompareTxid)
	})
}

func (c *Account) Directory() *record.Set[*url.URL] {
	return getOrCreateField(&c.directory, &c.directory_mu, func() *record.Set[*url.URL] {
		return record.NewSet(c.logger.L, c.store, c.key.Append("Directory"), c.label+" "+"directory", record.Wrapped(record.UrlWrapper), record.CompareUrl)
	})
}

func (c *Account) MainChain() *Chain2 {
	return getOrCreateField(&c.mainChain, &c.mainChain_mu, func() *Chain2 {
		return newChain2(c, c.logger.L, c.store, c.key.Append("MainChain"), "main", c.label+" "+"main chain")
	})
}

func (c *Account) ScratchChain() *Chain2 {
	return getOrCreateField(&c.scratchChain, &c.scratchChain_mu, func() *Chain2 {
		return newChain2(c, c.logger.L, c.store, c.key.Append("ScratchChain"), "scratch", c.label+" "+"scratch chain")
	})
}

func (c *Account) SignatureChain() *Chain2 {
	return getOrCreateField(&c.signatureChain, &c.signatureChain_mu, func() *Chain2 {
		return newChain2(c, c.logger.L, c.store, c.key.Append("SignatureChain"), "signature", c.label+" "+"signature chain")
	})
}

func (c *Account) RootChain() *Chain2 {
	return getOrCreateField(&c.rootChain, &c.rootChain_mu, func() *Chain2 {
		return newChain2(c, c.logger.L, c.store, c.key.Append("RootChain"), "root", c.label+" "+"root chain")
	})
}

func (c *Account) AnchorSequenceChain() *Chain2 {
	return getOrCreateField(&c.anchorSequenceChain, &c.anchorSequenceChain_mu, func() *Chain2 {
		return newChain2(c, c.logger.L, c.store, c.key.Append("AnchorSequenceChain"), "anchor-sequence", c.label+" "+"anchor sequence chain")
	})
}

func (c *Account) MajorBlockChain() *Chain2 {
	return getOrCreateField(&c.majorBlockChain, &c.majorBlockChain_mu, func() *Chain2 {
		return newChain2(c, c.logger.L, c.store, c.key.Append("MajorBlockChain"), "major-block", c.label+" "+"major block chain")
	})
}

func (c *Account) getSyntheticSequenceChain(partition string) *Chain2 {
	return getOrCreateMap(&c.syntheticSequenceChain, &c.syntheticSequenceChain_mu, keyForAccountSyntheticSequenceChain(partition), func() *Chain2 {
		return newChain2(c, c.logger.L, c.store, c.key.Append("SyntheticSequenceChain", partition), "synthetic-sequence(%[4]v)", c.label+" "+"synthetic sequence chain"+" "+partition)
	})
}

func (c *Account) getAnchorChain(partition string) *AccountAnchorChain {
	return getOrCreateMap(&c.anchorChain, &c.anchorChain_mu, keyForAccountAnchorChain(partition), func() *AccountAnchorChain {
		v := new(AccountAnchorChain)
		v.logger = c.logger
		v.store = c.store
		v.key = c.key.Append("AnchorChain", partition)
		v.parent = c
		v.label = c.label + " " + "anchor chain" + " " + partition
		return v
	})
}

func (c *Account) Chains() *record.Set[*protocol.ChainMetadata] {
	return getOrCreateField(&c.chains, &c.chains_mu, func() *record.Set[*protocol.ChainMetadata] {
		return record.NewSet(c.logger.L, c.store, c.key.Append("Chains"), c.label+" "+"chains", record.Struct[protocol.ChainMetadata](), func(u, v *protocol.ChainMetadata) int { return u.Compare(v) })
	})
}

func (c *Account) SyntheticAnchors() *record.Set[[32]byte] {
	return getOrCreateField(&c.syntheticAnchors, &c.syntheticAnchors_mu, func() *record.Set[[32]byte] {
		return record.NewSet(c.logger.L, c.store, c.key.Append("SyntheticAnchors"), c.label+" "+"synthetic anchors", record.Wrapped(record.HashWrapper), record.CompareHash)
	})
}

func (c *Account) Data() *AccountData {
	return getOrCreateField(&c.data, &c.data_mu, func() *AccountData {
		v := new(AccountData)
		v.logger = c.logger
		v.store = c.store
		v.key = c.key.Append("Data")
		v.parent = c
		v.label = c.label + " " + "data"
		return v
	})
}

func (c *Account) Resolve(key record.Key) (record.Record, record.Key, error) {
	if len(key) == 0 {
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for account")
	}

	switch key[0] {
	case "Url":
		return c.getUrl(), key[1:], nil
	case "Main":
		return c.Main(), key[1:], nil
	case "Pending":
		return c.Pending(), key[1:], nil
	case "SyntheticForAnchor":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for account")
		}
		anchor, okAnchor := key[1].([32]byte)
		if !okAnchor {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for account")
		}
		v := c.SyntheticForAnchor(anchor)
		return v, key[2:], nil
	case "Directory":
		return c.Directory(), key[1:], nil
	case "MainChain":
		return c.MainChain(), key[1:], nil
	case "ScratchChain":
		return c.ScratchChain(), key[1:], nil
	case "SignatureChain":
		return c.SignatureChain(), key[1:], nil
	case "RootChain":
		return c.RootChain(), key[1:], nil
	case "AnchorSequenceChain":
		return c.AnchorSequenceChain(), key[1:], nil
	case "MajorBlockChain":
		return c.MajorBlockChain(), key[1:], nil
	case "SyntheticSequenceChain":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for account")
		}
		partition, okPartition := key[1].(string)
		if !okPartition {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for account")
		}
		v := c.getSyntheticSequenceChain(partition)
		return v, key[2:], nil
	case "AnchorChain":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for account")
		}
		partition, okPartition := key[1].(string)
		if !okPartition {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for account")
		}
		v := c.getAnchorChain(partition)
		return v, key[2:], nil
	case "Chains":
		return c.Chains(), key[1:], nil
	case "SyntheticAnchors":
		return c.SyntheticAnchors(), key[1:], nil
	case "Data":
		return c.Data(), key[1:], nil
	default:
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for account")
	}
}

func (c *Account) IsDirty() bool {
	if c == nil {
		return false
	}

	if fieldIsDirty(c.url) {
		return true
	}
	if fieldIsDirty(c.main) {
		return true
	}
	if fieldIsDirty(c.pending) {
		return true
	}
	for _, v := range c.syntheticForAnchor {
		if v.IsDirty() {
			return true
		}
	}
	if fieldIsDirty(c.directory) {
		return true
	}
	if fieldIsDirty(c.mainChain) {
		return true
	}
	if fieldIsDirty(c.scratchChain) {
		return true
	}
	if fieldIsDirty(c.signatureChain) {
		return true
	}
	if fieldIsDirty(c.rootChain) {
		return true
	}
	if fieldIsDirty(c.anchorSequenceChain) {
		return true
	}
	if fieldIsDirty(c.majorBlockChain) {
		return true
	}
	for _, v := range c.syntheticSequenceChain {
		if v.IsDirty() {
			return true
		}
	}
	for _, v := range c.anchorChain {
		if v.IsDirty() {
			return true
		}
	}
	if fieldIsDirty(c.chains) {
		return true
	}
	if fieldIsDirty(c.syntheticAnchors) {
		return true
	}
	if fieldIsDirty(c.data) {
		return true
	}

	return false
}

func (c *Account) baseCommit() error {
	if c == nil {
		return nil
	}

	var err error
	commitField(&err, c.url)
	commitField(&err, c.main)
	commitField(&err, c.pending)
	for _, v := range c.syntheticForAnchor {
		commitField(&err, v)
	}
	commitField(&err, c.directory)
	commitField(&err, c.mainChain)
	commitField(&err, c.scratchChain)
	commitField(&err, c.signatureChain)
	commitField(&err, c.rootChain)
	commitField(&err, c.anchorSequenceChain)
	commitField(&err, c.majorBlockChain)
	for _, v := range c.syntheticSequenceChain {
		commitField(&err, v)
	}
	for _, v := range c.anchorChain {
		commitField(&err, v)
	}
	commitField(&err, c.chains)
	commitField(&err, c.syntheticAnchors)
	commitField(&err, c.data)

	return err
}

type AccountAnchorChain struct {
	logger logging.OptionalLogger
	store  record.Store
	key    record.Key
	label  string
	parent *Account

	root    *Chain2
	root_mu sync.RWMutex
	bpt     *Chain2
	bpt_mu  sync.RWMutex
}

func (c *AccountAnchorChain) Root() *Chain2 {
	return getOrCreateField(&c.root, &c.root_mu, func() *Chain2 {
		return newChain2(c, c.logger.L, c.store, c.key.Append("Root"), "anchor(%[4]v)-root", c.label+" "+"root")
	})
}

func (c *AccountAnchorChain) BPT() *Chain2 {
	return getOrCreateField(&c.bpt, &c.bpt_mu, func() *Chain2 {
		return newChain2(c, c.logger.L, c.store, c.key.Append("BPT"), "anchor(%[4]v)-bpt", c.label+" "+"bpt")
	})
}

func (c *AccountAnchorChain) Resolve(key record.Key) (record.Record, record.Key, error) {
	if len(key) == 0 {
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for anchor chain")
	}

	switch key[0] {
	case "Root":
		return c.Root(), key[1:], nil
	case "BPT":
		return c.BPT(), key[1:], nil
	default:
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for anchor chain")
	}
}

func (c *AccountAnchorChain) IsDirty() bool {
	if c == nil {
		return false
	}

	if fieldIsDirty(c.root) {
		return true
	}
	if fieldIsDirty(c.bpt) {
		return true
	}

	return false
}

func (c *AccountAnchorChain) Commit() error {
	if c == nil {
		return nil
	}

	var err error
	commitField(&err, c.root)
	commitField(&err, c.bpt)

	return err
}

type AccountData struct {
	logger logging.OptionalLogger
	store  record.Store
	key    record.Key
	label  string
	parent *Account

	entry          *record.Counted[[32]byte]
	entry_mu       sync.RWMutex
	transaction    map[accountDataTransactionKey]*record.Value[[32]byte]
	transaction_mu sync.RWMutex
}

type accountDataTransactionKey struct {
	EntryHash [32]byte
}

func keyForAccountDataTransaction(entryHash [32]byte) accountDataTransactionKey {
	return accountDataTransactionKey{entryHash}
}

func (c *AccountData) Entry() *record.Counted[[32]byte] {
	return getOrCreateField(&c.entry, &c.entry_mu, func() *record.Counted[[32]byte] {
		return record.NewCounted(c.logger.L, c.store, c.key.Append("Entry"), c.label+" "+"entry", record.WrappedFactory(record.HashWrapper))
	})
}

func (c *AccountData) Transaction(entryHash [32]byte) *record.Value[[32]byte] {
	return getOrCreateMap(&c.transaction, &c.transaction_mu, keyForAccountDataTransaction(entryHash), func() *record.Value[[32]byte] {
		return record.NewValue(c.logger.L, c.store, c.key.Append("Transaction", entryHash), c.label+" "+"transaction"+" "+hex.EncodeToString(entryHash[:]), false, record.Wrapped(record.HashWrapper))
	})
}

func (c *AccountData) Resolve(key record.Key) (record.Record, record.Key, error) {
	if len(key) == 0 {
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for data")
	}

	switch key[0] {
	case "Entry":
		return c.Entry(), key[1:], nil
	case "Transaction":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for data")
		}
		entryHash, okEntryHash := key[1].([32]byte)
		if !okEntryHash {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for data")
		}
		v := c.Transaction(entryHash)
		return v, key[2:], nil
	default:
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for data")
	}
}

func (c *AccountData) IsDirty() bool {
	if c == nil {
		return false
	}

	if fieldIsDirty(c.entry) {
		return true
	}
	for _, v := range c.transaction {
		if v.IsDirty() {
			return true
		}
	}

	return false
}

func (c *AccountData) Commit() error {
	if c == nil {
		return nil
	}

	var err error
	commitField(&err, c.entry)
	for _, v := range c.transaction {
		commitField(&err, v)
	}

	return err
}

type Transaction struct {
	logger logging.OptionalLogger
	store  record.Store
	key    record.Key
	label  string
	parent *Batch

	main          *record.Value[*SigOrTxn]
	main_mu       sync.RWMutex
	status        *record.Value[*protocol.TransactionStatus]
	status_mu     sync.RWMutex
	produced      *record.Set[*url.TxID]
	produced_mu   sync.RWMutex
	signatures    map[transactionSignaturesKey]*record.Value[*sigSetData]
	signatures_mu sync.RWMutex
	chains        *record.Set[*TransactionChainEntry]
	chains_mu     sync.RWMutex
}

type transactionSignaturesKey struct {
	Signer [32]byte
}

func keyForTransactionSignatures(signer *url.URL) transactionSignaturesKey {
	return transactionSignaturesKey{record.MapKeyUrl(signer)}
}

func (c *Transaction) Main() *record.Value[*SigOrTxn] {
	return getOrCreateField(&c.main, &c.main_mu, func() *record.Value[*SigOrTxn] {
		return record.NewValue(c.logger.L, c.store, c.key.Append("Main"), c.label+" "+"main", false, record.Struct[SigOrTxn]())
	})
}

func (c *Transaction) Status() *record.Value[*protocol.TransactionStatus] {
	return getOrCreateField(&c.status, &c.status_mu, func() *record.Value[*protocol.TransactionStatus] {
		return record.NewValue(c.logger.L, c.store, c.key.Append("Status"), c.label+" "+"status", true, record.Struct[protocol.TransactionStatus]())
	})
}

func (c *Transaction) Produced() *record.Set[*url.TxID] {
	return getOrCreateField(&c.produced, &c.produced_mu, func() *record.Set[*url.TxID] {
		return record.NewSet(c.logger.L, c.store, c.key.Append("Produced"), c.label+" "+"produced", record.Wrapped(record.TxidWrapper), record.CompareTxid)
	})
}

func (c *Transaction) getSignatures(signer *url.URL) *record.Value[*sigSetData] {
	return getOrCreateMap(&c.signatures, &c.signatures_mu, keyForTransactionSignatures(signer), func() *record.Value[*sigSetData] {
		return record.NewValue(c.logger.L, c.store, c.key.Append("Signatures", signer), c.label+" "+"signatures"+" "+signer.RawString(), true, record.Struct[sigSetData]())
	})
}

func (c *Transaction) Chains() *record.Set[*TransactionChainEntry] {
	return getOrCreateField(&c.chains, &c.chains_mu, func() *record.Set[*TransactionChainEntry] {
		return record.NewSet(c.logger.L, c.store, c.key.Append("Chains"), c.label+" "+"chains", record.Struct[TransactionChainEntry](), func(u, v *TransactionChainEntry) int { return u.Compare(v) })
	})
}

func (c *Transaction) Resolve(key record.Key) (record.Record, record.Key, error) {
	if len(key) == 0 {
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for transaction")
	}

	switch key[0] {
	case "Main":
		return c.Main(), key[1:], nil
	case "Status":
		return c.Status(), key[1:], nil
	case "Produced":
		return c.Produced(), key[1:], nil
	case "Signatures":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for transaction")
		}
		signer, okSigner := key[1].(*url.URL)
		if !okSigner {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for transaction")
		}
		v := c.getSignatures(signer)
		return v, key[2:], nil
	case "Chains":
		return c.Chains(), key[1:], nil
	default:
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for transaction")
	}
}

func (c *Transaction) IsDirty() bool {
	if c == nil {
		return false
	}

	if fieldIsDirty(c.main) {
		return true
	}
	if fieldIsDirty(c.status) {
		return true
	}
	if fieldIsDirty(c.produced) {
		return true
	}
	for _, v := range c.signatures {
		if v.IsDirty() {
			return true
		}
	}
	if fieldIsDirty(c.chains) {
		return true
	}

	return false
}

func (c *Transaction) Commit() error {
	if c == nil {
		return nil
	}

	var err error
	commitField(&err, c.main)
	commitField(&err, c.status)
	commitField(&err, c.produced)
	for _, v := range c.signatures {
		commitField(&err, v)
	}
	commitField(&err, c.chains)

	return err
}

type SystemData struct {
	logger logging.OptionalLogger
	store  record.Store
	key    record.Key
	label  string
	parent *Batch

	syntheticIndexIndex    map[systemDataSyntheticIndexIndexKey]*record.Value[uint64]
	syntheticIndexIndex_mu sync.RWMutex
}

type systemDataSyntheticIndexIndexKey struct {
	Block uint64
}

func keyForSystemDataSyntheticIndexIndex(block uint64) systemDataSyntheticIndexIndexKey {
	return systemDataSyntheticIndexIndexKey{block}
}

func (c *SystemData) SyntheticIndexIndex(block uint64) *record.Value[uint64] {
	return getOrCreateMap(&c.syntheticIndexIndex, &c.syntheticIndexIndex_mu, keyForSystemDataSyntheticIndexIndex(block), func() *record.Value[uint64] {
		return record.NewValue(c.logger.L, c.store, c.key.Append("SyntheticIndexIndex", block), c.label+" "+"synthetic index index"+" "+strconv.FormatUint(block, 10), false, record.Wrapped(record.UintWrapper))
	})
}

func (c *SystemData) Resolve(key record.Key) (record.Record, record.Key, error) {
	if len(key) == 0 {
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for system data")
	}

	switch key[0] {
	case "SyntheticIndexIndex":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for system data")
		}
		block, okBlock := key[1].(uint64)
		if !okBlock {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for system data")
		}
		v := c.SyntheticIndexIndex(block)
		return v, key[2:], nil
	default:
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for system data")
	}
}

func (c *SystemData) IsDirty() bool {
	if c == nil {
		return false
	}

	for _, v := range c.syntheticIndexIndex {
		if v.IsDirty() {
			return true
		}
	}

	return false
}

func (c *SystemData) Commit() error {
	if c == nil {
		return nil
	}

	var err error
	for _, v := range c.syntheticIndexIndex {
		commitField(&err, v)
	}

	return err
}

func getOrCreateField[T any](ptr **T, mu *sync.RWMutex, create func() *T) *T {
	mu.RLock()
	v := *ptr
	mu.RUnlock()

	if v != nil {
		return v
	}

	mu.Lock()
	defer mu.Unlock()
	if *ptr != nil {
		return *ptr
	}

	*ptr = create()
	return *ptr
}

func getOrCreateMap[T any, K comparable](ptr *map[K]T, mu *sync.RWMutex, key K, create func() T) T {
	mu.RLock()
	v, ok := (*ptr)[key]
	mu.RUnlock()
	if ok {
		return v
	}

	mu.Lock()
	defer mu.Unlock()
	if *ptr == nil {
		*ptr = map[K]T{}
	}

	if v, ok := (*ptr)[key]; ok {
		return v
	}

	v = create()
	(*ptr)[key] = v
	return v
}

func commitField[T any, PT record.RecordPtr[T]](lastErr *error, field PT) {
	if *lastErr != nil || field == nil {
		return
	}

	*lastErr = field.Commit()
}

func fieldIsDirty[T any, PT record.RecordPtr[T]](field PT) bool {
	return field != nil && field.IsDirty()
}
