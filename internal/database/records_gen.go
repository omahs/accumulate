package database

// GENERATED BY go run ./tools/cmd/gen-record. DO NOT EDIT.

import (
	"strings"

	"gitlab.com/accumulatenetwork/accumulate/internal/database/record"
	"gitlab.com/accumulatenetwork/accumulate/internal/errors"
	"gitlab.com/accumulatenetwork/accumulate/internal/logging"
	"gitlab.com/accumulatenetwork/accumulate/internal/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	"gitlab.com/accumulatenetwork/accumulate/smt/managed"
	"gitlab.com/accumulatenetwork/accumulate/smt/storage"
)

type ChangeSet struct {
	store    record.Store
	logger   logging.OptionalLogger
	done     bool
	writable bool
	id       uint64
	nextId   uint64
	kvStore  storage.KeyValueTxn

	account     map[storage.Key]*Account
	transaction map[storage.Key]*Transaction
}

func (c *ChangeSet) Account(url *url.URL) *Account {
	return getOrCreateMap(&c.account, record.Key{}.Append("Account", url), func() *Account {
		v := new(Account)
		v.store = c.store
		v.key = record.Key{}.Append("Account", url)
		v.container = c
		return v
	})
}

func (c *ChangeSet) Resolve(key record.Key) (record.Record, record.Key, error) {
	switch key[0] {
	case "Account":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for change set")
		}
		url, okUrl := key[1].(*url.URL)
		if okUrl {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for change set")
		}
		v := c.Account(url)
		return v, key[2:], nil
	case "Transaction":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for change set")
		}
		hash, okHash := key[1].([]byte)
		if okHash {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for change set")
		}
		v := c.Transaction(hash)
		return v, key[2:], nil
	default:
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for change set")
	}
}

func (c *ChangeSet) IsDirty() bool {
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

	return false
}

func (c *ChangeSet) dirtyChains() []*managed.Chain {
	if c == nil {
		return nil
	}

	var chains []*managed.Chain

	for _, v := range c.account {
		chains = append(chains, v.dirtyChains()...)
	}

	return chains
}

func (c *ChangeSet) baseCommit() error {
	if c == nil {
		return nil
	}

	for _, v := range c.account {
		if err := v.Commit(); err != nil {
			return errors.Wrap(errors.StatusUnknown, err)
		}
	}
	for _, v := range c.transaction {
		if err := v.Commit(); err != nil {
			return errors.Wrap(errors.StatusUnknown, err)
		}
	}

	return nil
}

type Account struct {
	store     record.Store
	key       record.Key
	container *ChangeSet

	state                         *record.Wrapped[protocol.Account]
	pending                       *record.Set[*url.TxID]
	syntheticForAnchor            map[storage.Key]*record.Set[*url.TxID]
	mainChain                     *managed.Chain
	signatureChain                *managed.Chain
	rootChain                     *AccountRootChain
	syntheticChain                *managed.Chain
	anchorChain                   map[storage.Key]*AccountAnchorChain
	mainIndexChain                *MajorMinorIndexChain
	signatureIndexChain           *MajorMinorIndexChain
	syntheticIndexChain           *MajorMinorIndexChain
	rootIndexChain                *MajorMinorIndexChain
	anchorIndexChain              map[storage.Key]*MajorMinorIndexChain
	syntheticProducedChain        map[storage.Key]*managed.Chain
	chains                        *record.Set[*protocol.ChainMetadata]
	syntheticAnchors              *record.Set[[32]byte]
	directory                     *record.Counted[*url.URL]
	data                          *AccountData
	blockChainUpdates             *record.Set[*ChainUpdate]
	producedSyntheticTransactions *record.Set[*BlockStateSynthTxnEntry]
}

func (c *Account) State() *record.Wrapped[protocol.Account] {
	return getOrCreateField(&c.state, func() *record.Wrapped[protocol.Account] {
		return record.NewWrapped(c.store, c.key.Append("State"), "account %[2]v state", false, record.NewWrapper(record.UnionWrapper(protocol.UnmarshalAccount)))
	})
}

func (c *Account) Pending() *record.Set[*url.TxID] {
	return getOrCreateField(&c.pending, func() *record.Set[*url.TxID] {
		return record.NewSet(c.store, c.key.Append("Pending"), "account %[2]v pending", record.NewWrapperSlice(record.TxidWrapper), record.CompareTxid)
	})
}

func (c *Account) SyntheticForAnchor(anchor [32]byte) *record.Set[*url.TxID] {
	return getOrCreateMap(&c.syntheticForAnchor, c.key.Append("SyntheticForAnchor", anchor), func() *record.Set[*url.TxID] {
		return record.NewSet(c.store, c.key.Append("SyntheticForAnchor", anchor), "account %[2]v synthetic for anchor %[4]x", record.NewWrapperSlice(record.TxidWrapper), record.CompareTxid)
	})
}

func (c *Account) MainChain() *managed.Chain {
	return getOrCreateField(&c.mainChain, func() *managed.Chain {
		return managed.NewChain(c.store, c.key.Append("MainChain"), markPower, managed.ChainTypeTransaction, "main", "account %[2]v main chain")
	})
}

func (c *Account) SignatureChain() *managed.Chain {
	return getOrCreateField(&c.signatureChain, func() *managed.Chain {
		return managed.NewChain(c.store, c.key.Append("SignatureChain"), markPower, managed.ChainTypeTransaction, "signature", "account %[2]v signature chain")
	})
}

func (c *Account) RootChain() *AccountRootChain {
	return getOrCreateField(&c.rootChain, func() *AccountRootChain {
		v := new(AccountRootChain)
		v.store = c.store
		v.key = c.key.Append("RootChain")
		v.container = c
		return v
	})
}

func (c *Account) SyntheticChain() *managed.Chain {
	return getOrCreateField(&c.syntheticChain, func() *managed.Chain {
		return managed.NewChain(c.store, c.key.Append("SyntheticChain"), markPower, managed.ChainTypeTransaction, "synthetic", "account %[2]v synthetic chain")
	})
}

func (c *Account) AnchorChain(subnetId string) *AccountAnchorChain {
	return getOrCreateMap(&c.anchorChain, c.key.Append("AnchorChain", subnetId), func() *AccountAnchorChain {
		v := new(AccountAnchorChain)
		v.store = c.store
		v.key = c.key.Append("AnchorChain", subnetId)
		v.container = c
		return v
	})
}

func (c *Account) MainIndexChain() *MajorMinorIndexChain {
	return getOrCreateField(&c.mainIndexChain, func() *MajorMinorIndexChain {
		return newMajorMinorIndexChain(c.store, c.key.Append("MainIndexChain"), "main-index", "account %[2]v main index chain")
	})
}

func (c *Account) SignatureIndexChain() *MajorMinorIndexChain {
	return getOrCreateField(&c.signatureIndexChain, func() *MajorMinorIndexChain {
		return newMajorMinorIndexChain(c.store, c.key.Append("SignatureIndexChain"), "signature-index", "account %[2]v signature index chain")
	})
}

func (c *Account) SyntheticIndexChain() *MajorMinorIndexChain {
	return getOrCreateField(&c.syntheticIndexChain, func() *MajorMinorIndexChain {
		return newMajorMinorIndexChain(c.store, c.key.Append("SyntheticIndexChain"), "synthetic-index", "account %[2]v synthetic index chain")
	})
}

func (c *Account) RootIndexChain() *MajorMinorIndexChain {
	return getOrCreateField(&c.rootIndexChain, func() *MajorMinorIndexChain {
		return newMajorMinorIndexChain(c.store, c.key.Append("RootIndexChain"), "root-index", "account %[2]v root index chain")
	})
}

func (c *Account) AnchorIndexChain(subnetId string) *MajorMinorIndexChain {
	return getOrCreateMap(&c.anchorIndexChain, c.key.Append("AnchorIndexChain", subnetId), func() *MajorMinorIndexChain {
		return newMajorMinorIndexChain(c.store, c.key.Append("AnchorIndexChain", subnetId), "anchor-index(%[4]v)", "account %[2]v anchor index chain %[4]v")
	})
}

func (c *Account) SyntheticProducedChain(subnetId string) *managed.Chain {
	return getOrCreateMap(&c.syntheticProducedChain, c.key.Append("SyntheticProducedChain", subnetId), func() *managed.Chain {
		return managed.NewChain(c.store, c.key.Append("SyntheticProducedChain", subnetId), markPower, managed.ChainTypeIndex, "synthetic-produced(%[4]v)", "account %[2]v synthetic produced chain %[4]v")
	})
}

func (c *Account) Chains() *record.Set[*protocol.ChainMetadata] {
	return getOrCreateField(&c.chains, func() *record.Set[*protocol.ChainMetadata] {
		new := func() (v *protocol.ChainMetadata) { return new(protocol.ChainMetadata) }
		cmp := func(u, v *protocol.ChainMetadata) int { return u.Compare(v) }
		return record.NewSet(c.store, c.key.Append("Chains"), "account %[2]v chains", record.NewSlice(new), cmp)
	})
}

func (c *Account) SyntheticAnchors() *record.Set[[32]byte] {
	return getOrCreateField(&c.syntheticAnchors, func() *record.Set[[32]byte] {
		return record.NewSet(c.store, c.key.Append("SyntheticAnchors"), "account %[2]v synthetic anchors", record.NewWrapperSlice(record.HashWrapper), record.CompareHash)
	})
}

func (c *Account) Directory() *record.Counted[*url.URL] {
	return getOrCreateField(&c.directory, func() *record.Counted[*url.URL] {

		return record.NewCounted(c.store, c.key.Append("Directory"), "account %[2]v directory", record.NewCountableWrapped(record.UrlWrapper))
	})
}

func (c *Account) Data() *AccountData {
	return getOrCreateField(&c.data, func() *AccountData {
		v := new(AccountData)
		v.store = c.store
		v.key = c.key.Append("Data")
		v.container = c
		return v
	})
}

func (c *Account) BlockChainUpdates() *record.Set[*ChainUpdate] {
	return getOrCreateField(&c.blockChainUpdates, func() *record.Set[*ChainUpdate] {
		new := func() (v *ChainUpdate) { return new(ChainUpdate) }
		cmp := func(u, v *ChainUpdate) int { return u.Compare(v) }
		return record.NewSet(c.store, c.key.Append("BlockChainUpdates"), "account %[2]v block chain updates", record.NewSlice(new), cmp)
	})
}

func (c *Account) ProducedSyntheticTransactions() *record.Set[*BlockStateSynthTxnEntry] {
	return getOrCreateField(&c.producedSyntheticTransactions, func() *record.Set[*BlockStateSynthTxnEntry] {
		new := func() (v *BlockStateSynthTxnEntry) { return new(BlockStateSynthTxnEntry) }
		cmp := func(u, v *BlockStateSynthTxnEntry) int { return u.Compare(v) }
		return record.NewSet(c.store, c.key.Append("ProducedSyntheticTransactions"), "account %[2]v produced synthetic transactions", record.NewSlice(new), cmp)
	})
}

func (c *Account) Resolve(key record.Key) (record.Record, record.Key, error) {
	switch key[0] {
	case "State":
		return c.state, key[1:], nil
	case "Pending":
		return c.pending, key[1:], nil
	case "SyntheticForAnchor":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for account")
		}
		anchor, okAnchor := key[1].([32]byte)
		if okAnchor {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for account")
		}
		v := c.SyntheticForAnchor(anchor)
		return v, key[2:], nil
	case "MainChain":
		return c.mainChain, key[1:], nil
	case "SignatureChain":
		return c.signatureChain, key[1:], nil
	case "RootChain":
		return c.rootChain, key[1:], nil
	case "SyntheticChain":
		return c.syntheticChain, key[1:], nil
	case "AnchorChain":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for account")
		}
		subnetId, okSubnetId := key[1].(string)
		if okSubnetId {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for account")
		}
		v := c.AnchorChain(subnetId)
		return v, key[2:], nil
	case "MainIndexChain":
		return c.mainIndexChain, key[1:], nil
	case "SignatureIndexChain":
		return c.signatureIndexChain, key[1:], nil
	case "SyntheticIndexChain":
		return c.syntheticIndexChain, key[1:], nil
	case "RootIndexChain":
		return c.rootIndexChain, key[1:], nil
	case "AnchorIndexChain":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for account")
		}
		subnetId, okSubnetId := key[1].(string)
		if okSubnetId {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for account")
		}
		v := c.AnchorIndexChain(subnetId)
		return v, key[2:], nil
	case "SyntheticProducedChain":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for account")
		}
		subnetId, okSubnetId := key[1].(string)
		if okSubnetId {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for account")
		}
		v := c.SyntheticProducedChain(subnetId)
		return v, key[2:], nil
	case "Chains":
		return c.chains, key[1:], nil
	case "SyntheticAnchors":
		return c.syntheticAnchors, key[1:], nil
	case "Directory":
		return c.directory, key[1:], nil
	case "Data":
		return c.data, key[1:], nil
	case "BlockChainUpdates":
		return c.blockChainUpdates, key[1:], nil
	case "ProducedSyntheticTransactions":
		return c.producedSyntheticTransactions, key[1:], nil
	default:
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for account")
	}
}

func (c *Account) IsDirty() bool {
	if c == nil {
		return false
	}

	if c.state.IsDirty() {
		return true
	}
	if c.pending.IsDirty() {
		return true
	}
	for _, v := range c.syntheticForAnchor {
		if v.IsDirty() {
			return true
		}
	}
	if c.mainChain.IsDirty() {
		return true
	}
	if c.signatureChain.IsDirty() {
		return true
	}
	if c.rootChain.IsDirty() {
		return true
	}
	if c.syntheticChain.IsDirty() {
		return true
	}
	for _, v := range c.anchorChain {
		if v.IsDirty() {
			return true
		}
	}
	if c.mainIndexChain.IsDirty() {
		return true
	}
	if c.signatureIndexChain.IsDirty() {
		return true
	}
	if c.syntheticIndexChain.IsDirty() {
		return true
	}
	if c.rootIndexChain.IsDirty() {
		return true
	}
	for _, v := range c.anchorIndexChain {
		if v.IsDirty() {
			return true
		}
	}
	for _, v := range c.syntheticProducedChain {
		if v.IsDirty() {
			return true
		}
	}
	if c.chains.IsDirty() {
		return true
	}
	if c.syntheticAnchors.IsDirty() {
		return true
	}
	if c.directory.IsDirty() {
		return true
	}
	if c.data.IsDirty() {
		return true
	}
	if c.blockChainUpdates.IsDirty() {
		return true
	}
	if c.producedSyntheticTransactions.IsDirty() {
		return true
	}

	return false
}

func (c *Account) resolveChain(name string) (*managed.Chain, bool) {
	switch {
	case name == "main":
		return c.MainChain(), true

	case name == "signature":
		return c.SignatureChain(), true

	case strings.HasPrefix(name, "root-"):
		return c.RootChain().resolveChain(name[len("root-"):])

	case name == "synthetic":
		return c.SyntheticChain(), true

	case strings.HasPrefix(name, "anchor("):
		name = name[len("anchor("):]
		i := strings.Index(name, ")")
		if i < 0 {
			return nil, false
		}

		params := strings.Split(name[:i], ",")
		name = name[i+1:]
		if len(params) != 1 {
			return nil, false
		}
		paramSubnetId, err := record.ParseString(params[0])
		if err != nil {
			return nil, false
		}

		return c.AnchorChain(paramSubnetId).resolveChain(name)

	case strings.HasPrefix(name, "synthetic-produced("):
		name = name[len("synthetic-produced("):]
		i := strings.Index(name, ")")
		if i < 0 {
			return nil, false
		}

		params := strings.Split(name[:i], ",")
		name = name[i+1:]
		if len(params) != 1 {
			return nil, false
		}
		paramSubnetId, err := record.ParseString(params[0])
		if err != nil {
			return nil, false
		}

		return c.SyntheticProducedChain(paramSubnetId), true

	default:
		return nil, false
	}
}

func (c *Account) dirtyChains() []*managed.Chain {
	if c == nil {
		return nil
	}

	var chains []*managed.Chain

	if c.mainChain.IsDirty() {
		chains = append(chains, c.mainChain)
	}
	if c.signatureChain.IsDirty() {
		chains = append(chains, c.signatureChain)
	}
	chains = append(chains, c.rootChain.dirtyChains()...)
	if c.syntheticChain.IsDirty() {
		chains = append(chains, c.syntheticChain)
	}
	for _, v := range c.anchorChain {
		chains = append(chains, v.dirtyChains()...)
	}
	for _, v := range c.syntheticProducedChain {
		if v.IsDirty() {
			chains = append(chains, v)
		}
	}

	return chains
}

func (c *Account) baseCommit() error {
	if c == nil {
		return nil
	}

	if err := c.state.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.pending.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	for _, v := range c.syntheticForAnchor {
		if err := v.Commit(); err != nil {
			return errors.Wrap(errors.StatusUnknown, err)
		}
	}
	if err := c.mainChain.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.signatureChain.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.rootChain.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.syntheticChain.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	for _, v := range c.anchorChain {
		if err := v.Commit(); err != nil {
			return errors.Wrap(errors.StatusUnknown, err)
		}
	}
	if err := c.mainIndexChain.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.signatureIndexChain.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.syntheticIndexChain.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.rootIndexChain.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	for _, v := range c.anchorIndexChain {
		if err := v.Commit(); err != nil {
			return errors.Wrap(errors.StatusUnknown, err)
		}
	}
	for _, v := range c.syntheticProducedChain {
		if err := v.Commit(); err != nil {
			return errors.Wrap(errors.StatusUnknown, err)
		}
	}
	if err := c.chains.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.syntheticAnchors.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.directory.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.data.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.blockChainUpdates.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.producedSyntheticTransactions.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}

	return nil
}

type AccountRootChain struct {
	store     record.Store
	key       record.Key
	container *Account

	minor *managed.Chain
	major *managed.Chain
}

func (c *AccountRootChain) Minor() *managed.Chain {
	return getOrCreateField(&c.minor, func() *managed.Chain {
		return managed.NewChain(c.store, c.key.Append("Minor"), markPower, managed.ChainTypeAnchor, "root-minor", "account %[2]v root chain minor")
	})
}

func (c *AccountRootChain) Major() *managed.Chain {
	return getOrCreateField(&c.major, func() *managed.Chain {
		return managed.NewChain(c.store, c.key.Append("Major"), markPower, managed.ChainTypeAnchor, "root-major", "account %[2]v root chain major")
	})
}

func (c *AccountRootChain) Resolve(key record.Key) (record.Record, record.Key, error) {
	switch key[0] {
	case "Minor":
		return c.minor, key[1:], nil
	case "Major":
		return c.major, key[1:], nil
	default:
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for root chain")
	}
}

func (c *AccountRootChain) IsDirty() bool {
	if c == nil {
		return false
	}

	if c.minor.IsDirty() {
		return true
	}
	if c.major.IsDirty() {
		return true
	}

	return false
}

func (c *AccountRootChain) resolveChain(name string) (*managed.Chain, bool) {
	switch {
	case name == "minor":
		return c.Minor(), true

	case name == "major":
		return c.Major(), true

	default:
		return nil, false
	}
}

func (c *AccountRootChain) dirtyChains() []*managed.Chain {
	if c == nil {
		return nil
	}

	var chains []*managed.Chain

	if c.minor.IsDirty() {
		chains = append(chains, c.minor)
	}
	if c.major.IsDirty() {
		chains = append(chains, c.major)
	}

	return chains
}

func (c *AccountRootChain) Commit() error {
	if c == nil {
		return nil
	}

	if err := c.minor.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.major.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}

	return nil
}

type AccountAnchorChain struct {
	store     record.Store
	key       record.Key
	container *Account

	root *managed.Chain
	bpt  *managed.Chain
}

func (c *AccountAnchorChain) Root() *managed.Chain {
	return getOrCreateField(&c.root, func() *managed.Chain {
		return managed.NewChain(c.store, c.key.Append("Root"), markPower, managed.ChainTypeAnchor, "anchor(%[4]v)-root", "account %[2]v anchor chain %[4]v root")
	})
}

func (c *AccountAnchorChain) BPT() *managed.Chain {
	return getOrCreateField(&c.bpt, func() *managed.Chain {
		return managed.NewChain(c.store, c.key.Append("BPT"), markPower, managed.ChainTypeAnchor, "anchor(%[4]v)-bpt", "account %[2]v anchor chain %[4]v bpt")
	})
}

func (c *AccountAnchorChain) Resolve(key record.Key) (record.Record, record.Key, error) {
	switch key[0] {
	case "Root":
		return c.root, key[1:], nil
	case "BPT":
		return c.bpt, key[1:], nil
	default:
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for anchor chain")
	}
}

func (c *AccountAnchorChain) IsDirty() bool {
	if c == nil {
		return false
	}

	if c.root.IsDirty() {
		return true
	}
	if c.bpt.IsDirty() {
		return true
	}

	return false
}

func (c *AccountAnchorChain) resolveChain(name string) (*managed.Chain, bool) {
	switch {
	case name == "root":
		return c.Root(), true

	case name == "bpt":
		return c.BPT(), true

	default:
		return nil, false
	}
}

func (c *AccountAnchorChain) dirtyChains() []*managed.Chain {
	if c == nil {
		return nil
	}

	var chains []*managed.Chain

	if c.root.IsDirty() {
		chains = append(chains, c.root)
	}
	if c.bpt.IsDirty() {
		chains = append(chains, c.bpt)
	}

	return chains
}

func (c *AccountAnchorChain) Commit() error {
	if c == nil {
		return nil
	}

	if err := c.root.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.bpt.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}

	return nil
}

type AccountData struct {
	store     record.Store
	key       record.Key
	container *Account

	entry       *record.Counted[[32]byte]
	transaction map[storage.Key]*record.Wrapped[[32]byte]
}

func (c *AccountData) Entry() *record.Counted[[32]byte] {
	return getOrCreateField(&c.entry, func() *record.Counted[[32]byte] {

		return record.NewCounted(c.store, c.key.Append("Entry"), "account %[2]v data entry", record.NewCountableWrapped(record.HashWrapper))
	})
}

func (c *AccountData) Transaction(entryHash [32]byte) *record.Wrapped[[32]byte] {
	return getOrCreateMap(&c.transaction, c.key.Append("Transaction", entryHash), func() *record.Wrapped[[32]byte] {
		return record.NewWrapped(c.store, c.key.Append("Transaction", entryHash), "account %[2]v data transaction %[5]x", false, record.NewWrapper(record.HashWrapper))
	})
}

func (c *AccountData) Resolve(key record.Key) (record.Record, record.Key, error) {
	switch key[0] {
	case "Entry":
		return c.entry, key[1:], nil
	case "Transaction":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for data")
		}
		entryHash, okEntryHash := key[1].([32]byte)
		if okEntryHash {
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

	if c.entry.IsDirty() {
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

	if err := c.entry.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	for _, v := range c.transaction {
		if err := v.Commit(); err != nil {
			return errors.Wrap(errors.StatusUnknown, err)
		}
	}

	return nil
}

type Transaction struct {
	store     record.Store
	key       record.Key
	container *ChangeSet

	value            *record.Value[*protocol.Transaction]
	status           *record.Value[*protocol.TransactionStatus]
	produced         *record.Set[*url.TxID]
	systemSignatures *SystemSignatureSet
	signatures       map[storage.Key]*VersionedSignatureSet
	signers          *record.Set[*url.URL]
	chains           *record.Set[*TransactionChainEntry]
	signature        *record.Wrapped[protocol.Signature]
	txID             *record.Wrapped[*url.TxID]
}

func (c *Transaction) Value() *record.Value[*protocol.Transaction] {
	return getOrCreateField(&c.value, func() *record.Value[*protocol.Transaction] {
		new := func() (v *protocol.Transaction) { return new(protocol.Transaction) }
		return record.NewValue(c.store, c.key.Append("Value"), "transaction %[2]x value", false, new)
	})
}

func (c *Transaction) Status() *record.Value[*protocol.TransactionStatus] {
	return getOrCreateField(&c.status, func() *record.Value[*protocol.TransactionStatus] {
		new := func() (v *protocol.TransactionStatus) { return new(protocol.TransactionStatus) }
		return record.NewValue(c.store, c.key.Append("Status"), "transaction %[2]x status", false, new)
	})
}

func (c *Transaction) Produced() *record.Set[*url.TxID] {
	return getOrCreateField(&c.produced, func() *record.Set[*url.TxID] {
		return record.NewSet(c.store, c.key.Append("Produced"), "transaction %[2]x produced", record.NewWrapperSlice(record.TxidWrapper), record.CompareTxid)
	})
}

func (c *Transaction) SystemSignatures() *SystemSignatureSet {
	return getOrCreateField(&c.systemSignatures, func() *SystemSignatureSet {
		return newSystemSignatureSet(c.store, c.key.Append("SystemSignatures"), "transaction(%[2]x)-system-signatures", "transaction %[2]x system signatures")
	})
}

func (c *Transaction) Signers() *record.Set[*url.URL] {
	return getOrCreateField(&c.signers, func() *record.Set[*url.URL] {
		return record.NewSet(c.store, c.key.Append("Signers"), "transaction %[2]x signers", record.NewWrapperSlice(record.UrlWrapper), record.CompareUrl)
	})
}

func (c *Transaction) Chains() *record.Set[*TransactionChainEntry] {
	return getOrCreateField(&c.chains, func() *record.Set[*TransactionChainEntry] {
		new := func() (v *TransactionChainEntry) { return new(TransactionChainEntry) }
		cmp := func(u, v *TransactionChainEntry) int { return u.Compare(v) }
		return record.NewSet(c.store, c.key.Append("Chains"), "transaction %[2]x chains", record.NewSlice(new), cmp)
	})
}

func (c *Transaction) Signature() *record.Wrapped[protocol.Signature] {
	return getOrCreateField(&c.signature, func() *record.Wrapped[protocol.Signature] {
		return record.NewWrapped(c.store, c.key.Append("Signature"), "transaction %[2]x signature", false, record.NewWrapper(record.UnionWrapper(protocol.UnmarshalSignature)))
	})
}

func (c *Transaction) TxID() *record.Wrapped[*url.TxID] {
	return getOrCreateField(&c.txID, func() *record.Wrapped[*url.TxID] {
		return record.NewWrapped(c.store, c.key.Append("TxID"), "transaction %[2]x tx id", false, record.NewWrapper(record.TxidWrapper))
	})
}

func (c *Transaction) Resolve(key record.Key) (record.Record, record.Key, error) {
	switch key[0] {
	case "Value":
		return c.value, key[1:], nil
	case "Status":
		return c.status, key[1:], nil
	case "Produced":
		return c.produced, key[1:], nil
	case "SystemSignatures":
		return c.systemSignatures, key[1:], nil
	case "Signatures":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for transaction")
		}
		signer, okSigner := key[1].(*url.URL)
		if okSigner {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for transaction")
		}
		v := c.Signatures(signer)
		return v, key[2:], nil
	case "Signers":
		return c.signers, key[1:], nil
	case "Chains":
		return c.chains, key[1:], nil
	case "Signature":
		return c.signature, key[1:], nil
	case "TxID":
		return c.txID, key[1:], nil
	default:
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for transaction")
	}
}

func (c *Transaction) IsDirty() bool {
	if c == nil {
		return false
	}

	if c.value.IsDirty() {
		return true
	}
	if c.status.IsDirty() {
		return true
	}
	if c.produced.IsDirty() {
		return true
	}
	if c.systemSignatures.IsDirty() {
		return true
	}
	for _, v := range c.signatures {
		if v.IsDirty() {
			return true
		}
	}
	if c.signers.IsDirty() {
		return true
	}
	if c.chains.IsDirty() {
		return true
	}
	if c.signature.IsDirty() {
		return true
	}
	if c.txID.IsDirty() {
		return true
	}

	return false
}

func (c *Transaction) baseCommit() error {
	if c == nil {
		return nil
	}

	if err := c.value.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.status.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.produced.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.systemSignatures.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	for _, v := range c.signatures {
		if err := v.Commit(); err != nil {
			return errors.Wrap(errors.StatusUnknown, err)
		}
	}
	if err := c.signers.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.chains.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.signature.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.txID.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}

	return nil
}

type MajorMinorIndexChain struct {
	store record.Store
	key   record.Key
	name  string
	label string

	minor *managed.Chain
	major *managed.Chain
}

func (c *MajorMinorIndexChain) Minor() *managed.Chain {
	return getOrCreateField(&c.minor, func() *managed.Chain {
		return managed.NewChain(c.store, record.Key{}.Append("Minor"), markPower, managed.ChainTypeIndex, c.name+"-minor", c.label+" minor")
	})
}

func (c *MajorMinorIndexChain) Major() *managed.Chain {
	return getOrCreateField(&c.major, func() *managed.Chain {
		return managed.NewChain(c.store, record.Key{}.Append("Major"), markPower, managed.ChainTypeIndex, c.name+"-major", c.label+" major")
	})
}

func (c *MajorMinorIndexChain) Resolve(key record.Key) (record.Record, record.Key, error) {
	switch key[0] {
	case "Minor":
		return c.minor, key[1:], nil
	case "Major":
		return c.major, key[1:], nil
	default:
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for major minor index chain")
	}
}

func (c *MajorMinorIndexChain) IsDirty() bool {
	if c == nil {
		return false
	}

	if c.minor.IsDirty() {
		return true
	}
	if c.major.IsDirty() {
		return true
	}

	return false
}

func (c *MajorMinorIndexChain) resolveChain(name string) (*managed.Chain, bool) {
	switch {
	case name == "minor":
		return c.Minor(), true

	case name == "major":
		return c.Major(), true

	default:
		return nil, false
	}
}

func (c *MajorMinorIndexChain) dirtyChains() []*managed.Chain {
	if c == nil {
		return nil
	}

	var chains []*managed.Chain

	if c.minor.IsDirty() {
		chains = append(chains, c.minor)
	}
	if c.major.IsDirty() {
		chains = append(chains, c.major)
	}

	return chains
}

func (c *MajorMinorIndexChain) Commit() error {
	if c == nil {
		return nil
	}

	if err := c.minor.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	if err := c.major.Commit(); err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}

	return nil
}

func getOrCreateField[T any](ptr **T, create func() *T) *T {
	if *ptr != nil {
		return *ptr
	}

	*ptr = create()
	return *ptr
}

func getOrCreateMap[T any](ptr *map[storage.Key]*T, key record.Key, create func() *T) *T {
	if *ptr == nil {
		*ptr = map[storage.Key]*T{}
	}

	k := key.Hash()
	if v, ok := (*ptr)[k]; ok {
		return v
	}

	v := create()
	(*ptr)[k] = v
	return v
}
