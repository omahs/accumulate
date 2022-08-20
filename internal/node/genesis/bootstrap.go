package genesis

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strings"
	"time"

	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	"gitlab.com/accumulatenetwork/accumulate/internal/block"
	"gitlab.com/accumulatenetwork/accumulate/internal/core"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/database/smt/storage"
	"gitlab.com/accumulatenetwork/accumulate/internal/database/smt/storage/memory"
	"gitlab.com/accumulatenetwork/accumulate/internal/encoding"
	"gitlab.com/accumulatenetwork/accumulate/internal/execute"
	"gitlab.com/accumulatenetwork/accumulate/internal/node/config"
	"gitlab.com/accumulatenetwork/accumulate/internal/node/routing"
	"gitlab.com/accumulatenetwork/accumulate/pkg/errors"
	"gitlab.com/accumulatenetwork/accumulate/pkg/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	"golang.org/x/sync/errgroup"
)

type InitOpts struct {
	PartitionId     string
	NetworkType     config.NetworkType
	GenesisTime     time.Time
	Logger          log.Logger
	FactomAddresses func() (io.Reader, error)
	GenesisGlobals  *core.GlobalValues
	OperatorKeys    [][]byte
}

func Init(snapshot io.WriteSeeker, opts InitOpts) ([]byte, error) {
	store := memory.New(opts.Logger.With("module", "storage"))
	b := &bootstrap{
		InitOpts:    opts,
		kvdb:        store,
		db:          database.New(store, opts.Logger.With("module", "database")),
		dataRecords: make([]DataRecord, 0),
		records:     make([]protocol.Account, 0),
	}

	// Build the routing table
	var bvns []string
	for _, partition := range opts.GenesisGlobals.Network.Partitions {
		if partition.PartitionID != protocol.Directory {
			bvns = append(bvns, partition.PartitionID)
		}
	}
	b.routingTable = new(protocol.RoutingTable)
	b.routingTable.Routes = routing.BuildSimpleTable(bvns)
	b.routingTable.Overrides = make([]protocol.RouteOverride, 1, len(opts.GenesisGlobals.Network.Partitions)+1)
	b.routingTable.Overrides[0] = protocol.RouteOverride{Account: protocol.AcmeUrl(), Partition: protocol.Directory}
	for _, partition := range opts.GenesisGlobals.Network.Partitions {
		u := protocol.PartitionUrl(partition.PartitionID)
		b.routingTable.Overrides = append(b.routingTable.Overrides, protocol.RouteOverride{Account: u, Partition: partition.PartitionID})
	}

	// Create the router
	var err error
	b.router, err = routing.NewStaticRouter(b.routingTable, nil)
	if err != nil {
		return nil, errors.Wrap(errors.StatusUnknownError, err)
	}

	exec, err := block.NewGenesisExecutor(b.db, opts.Logger, &config.Describe{
		NetworkType: opts.NetworkType,
		PartitionId: opts.PartitionId,
	}, b.router)
	if err != nil {
		return nil, errors.Wrap(errors.StatusUnknownError, err)
	}

	// Capture background tasks
	errg := new(errgroup.Group)
	exec.Background = func(f func()) { errg.Go(func() error { f(); return nil }) }

	b.block = new(block.Block)
	b.block.Index = protocol.GenesisBlock

	b.block.Time = b.GenesisTime
	b.block.Batch = b.db.Begin(true)
	defer b.block.Batch.Discard()

	err = exec.Genesis(b.block, b)
	if err != nil {
		return nil, errors.Wrap(errors.StatusUnknownError, err)
	}

	err = b.block.Batch.Commit()
	if err != nil {
		return nil, errors.Wrap(errors.StatusUnknownError, err)
	}

	// Wait for background tasks
	err = errg.Wait()
	if err != nil {
		return nil, errors.Wrap(errors.StatusUnknownError, err)
	}

	batch := b.db.Begin(false)
	defer batch.Discard()
	err = exec.SaveSnapshot(batch, snapshot)
	if err != nil {
		return nil, errors.Wrap(errors.StatusUnknownError, err)
	}

	return batch.BptRoot(), nil
}

type bootstrap struct {
	InitOpts
	networkAuthority *url.URL
	partition        config.NetworkUrl
	localAuthority   *url.URL

	kvdb         storage.KeyValueStore
	db           *database.Database
	block        *block.Block
	urls         []*url.URL
	records      []protocol.Account
	dataRecords  []DataRecord
	router       routing.Router
	routingTable *protocol.RoutingTable
	globals      *core.GlobalValues
}

type DataRecord struct {
	Account *url.URL
	Entry   protocol.DataEntry
}

var _ execute.TransactionExecutor = &bootstrap{}
var _ execute.PrincipalValidator = &bootstrap{}

func (bootstrap) Type() protocol.TransactionType {
	return protocol.TransactionTypeSyntheticDepositTokens
}

func (b *bootstrap) AllowMissingPrincipal(*protocol.Transaction) bool {
	return true
}

func (b *bootstrap) Execute(st *execute.StateManager, tx *execute.Delivery) (protocol.TransactionResult, error) {
	return b.Validate(st, tx)
}

func (b *bootstrap) Validate(st *execute.StateManager, tx *execute.Delivery) (protocol.TransactionResult, error) {
	b.networkAuthority = protocol.DnUrl().JoinPath(protocol.Operators)
	b.partition = config.NetworkUrl{URL: protocol.PartitionUrl(b.PartitionId)}
	b.localAuthority = b.partition.Operators()

	// Verify that the BVN ID will make a valid partition URL
	if err := protocol.IsValidAdiUrl(protocol.PartitionUrl(b.PartitionId), true); err != nil {
		panic(fmt.Errorf("%q is not a valid partition ID: %v", b.PartitionId, err))
	}

	// Setup globals and create network variable accounts
	if b.GenesisGlobals == nil {
		b.globals = new(core.GlobalValues)
	} else {
		b.globals = b.GenesisGlobals
	}

	// set the initial price to 1/5 fct price * 1/4 market cap dilution = 1/20 fct price
	// for this exercise, we'll assume that 1 FCT = $1, so initial ACME price is $0.05
	if b.globals.Oracle == nil {
		b.globals.Oracle = new(protocol.AcmeOracle)
		b.globals.Oracle.Price = uint64(protocol.InitialAcmeOracleValue)
	}

	// Set the initial threshold to 2/3 & MajorBlockSchedule
	if b.globals.Globals == nil {
		b.globals.Globals = new(protocol.NetworkGlobals)
	}
	if b.globals.Globals.OperatorAcceptThreshold.Numerator == 0 {
		b.globals.Globals.OperatorAcceptThreshold.Set(2, 3)
	}
	if b.globals.Globals.MajorBlockSchedule == "" {
		b.globals.Globals.MajorBlockSchedule = protocol.DefaultMajorBlockSchedule
	}

	if b.globals.Routing == nil {
		b.globals.Routing = b.routingTable
	}

	err := b.globals.Store(b.partition, func(accountUrl *url.URL, target interface{}) error {
		da := new(protocol.DataAccount)
		da.Url = accountUrl
		da.AddAuthority(b.networkAuthority)
		return encoding.SetPtr(da, target)
	}, func(account protocol.Account) error {
		b.WriteRecords(account)
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(errors.StatusUnknownError, err)
	}

	// Create accounts
	b.createIdentity()
	b.createMainLedger()
	b.createSyntheticLedger()
	b.createAnchorPool()
	b.createOperatorBook()
	b.createEvidenceChain()
	b.maybeCreateAcme()
	b.maybeCreateFaucet()

	err = b.createVoteScratchChain()
	if err != nil {
		return nil, errors.Wrap(errors.StatusUnknownError, err)
	}

	err = b.maybeCreateFactomAccounts()
	if err != nil {
		return nil, errors.Wrap(errors.StatusUnknownError, err)
	}

	// Persist accounts
	err = st.Create(b.records...)
	if err != nil {
		return nil, errors.Format(errors.StatusUnknownError, "store records: %w", err)
	}

	// Update the directory index
	err = st.AddDirectoryEntry(b.partition.Identity(), b.urls...)
	if err != nil {
		return nil, errors.Wrap(errors.StatusUnknownError, err)
	}

	// Write data entries
	for _, wd := range b.dataRecords {
		body := new(protocol.SystemWriteData)
		body.Entry = wd.Entry
		txn := new(protocol.Transaction)
		txn.Header.Principal = wd.Account
		txn.Body = body
		st.State.ProcessAdditionalTransaction(tx.NewInternal(txn))
	}

	return nil, nil
}

func (b *bootstrap) createIdentity() {
	// Create the ADI
	adi := new(protocol.ADI)
	adi.Url = b.partition.Identity()
	adi.AddAuthority(b.localAuthority)
	b.WriteRecords(adi)
}

func (b *bootstrap) createMainLedger() {
	// Create the main ledger
	ledger := new(protocol.SystemLedger)
	ledger.Url = b.partition.Ledger()
	ledger.Index = protocol.GenesisBlock
	b.WriteRecords(ledger)
}

func (b *bootstrap) createSyntheticLedger() {
	// Create the synth ledger
	synthLedger := new(protocol.SyntheticLedger)
	synthLedger.Url = b.partition.Synthetic()
	b.WriteRecords(synthLedger)
}

func (b *bootstrap) createAnchorPool() {
	// Create the anchor pool
	anchorLedger := new(protocol.AnchorLedger)
	anchorLedger.Url = b.partition.AnchorPool()

	if b.NetworkType == config.Directory {
		// Initialize the last major block time to prevent a major block from
		// being created immediately once the network boots
		anchorLedger.MajorBlockTime = b.GenesisTime
	}

	b.WriteRecords(anchorLedger)
}

func (b *bootstrap) createVoteScratchChain() error {
	//create a vote scratch chain
	wd := new(protocol.WriteData)
	lci := types.LastCommitInfo{}
	data, err := json.Marshal(&lci)
	if err != nil {
		return errors.Format(errors.StatusInternalError, "marshal last commit info: %w", err)
	}
	wd.Entry = &protocol.AccumulateDataEntry{Data: [][]byte{data}}
	wd.Scratch = true

	da := new(protocol.DataAccount)
	da.Url = b.partition.URL.JoinPath(protocol.Votes)
	da.AddAuthority(b.localAuthority)
	b.writeDataRecord(da, da.Url, DataRecord{da.Url, wd.Entry})
	return nil
}

func (b *bootstrap) createEvidenceChain() {
	//create an evidence chain
	da := new(protocol.DataAccount)
	da.Url = b.partition.JoinPath(protocol.Evidence)
	da.AddAuthority(b.localAuthority)
	b.WriteRecords(da)
	b.urls = append(b.urls, da.Url)
}

func (b *bootstrap) maybeCreateAcme() {
	if !b.shouldCreate(protocol.AcmeUrl()) {
		return
	}

	acme := new(protocol.TokenIssuer)
	acme.AddAuthority(b.networkAuthority)
	acme.Url = protocol.AcmeUrl()
	acme.Precision = 8
	acme.Symbol = "ACME"

	if protocol.IsTestNet {
		// On the TestNet, set the issued amount to the faucet balance
		acme.Issued.SetString(protocol.AcmeFaucetBalance, 10)
	} else {
		// On the MainNet, set the supply limit
		acme.SupplyLimit = big.NewInt(protocol.AcmeSupplyLimit * protocol.AcmePrecision)
	}

	b.WriteRecords(acme)
}

func (b *bootstrap) maybeCreateFaucet() {
	if !b.shouldCreate(protocol.FaucetUrl) {
		return
	}

	liteId := new(protocol.LiteIdentity)
	liteId.Url = protocol.FaucetUrl.RootIdentity()

	liteToken := new(protocol.LiteTokenAccount)
	liteToken.Url = protocol.FaucetUrl
	liteToken.TokenUrl = protocol.AcmeUrl()
	liteToken.Balance.SetString(protocol.AcmeFaucetBalance, 10)
	b.WriteRecords(liteId, liteToken)
}

func (b *bootstrap) maybeCreateFactomAccounts() error {
	if b.FactomAddresses == nil {
		return nil
	}

	rd, err := b.FactomAddresses()
	if err != nil {
		return errors.Wrap(errors.StatusUnknownError, err)
	}

	factomAddresses, err := LoadFactomAddressesAndBalances(rd)
	if err != nil {
		return errors.Wrap(errors.StatusUnknownError, err)
	}

	for _, fa := range factomAddresses {
		if !b.shouldCreate(fa.Address) {
			continue
		}

		lid := new(protocol.LiteIdentity)
		lid.Url = fa.Address.RootIdentity()

		lite := new(protocol.LiteTokenAccount)
		lite.Url = fa.Address
		lite.TokenUrl = protocol.AcmeUrl()
		lite.Balance = *big.NewInt(5 * fa.Balance)

		b.WriteRecords(lid, lite)
	}
	return nil
}

func (b *bootstrap) shouldCreate(url *url.URL) bool {
	partition, err := b.router.RouteAccount(url)
	if err != nil {
		panic(err) // An account should never be unroutable
	}

	return strings.EqualFold(b.PartitionId, partition)
}

func (b *bootstrap) createOperatorBook() {
	book := new(protocol.KeyBook)
	book.Url = b.partition.Operators()
	book.AddAuthority(b.networkAuthority)
	book.PageCount = 1

	page := new(protocol.KeyPage)
	page.Url = book.Url.JoinPath("1")
	page.Version = 1

	for _, operator := range b.OperatorKeys {
		spec := new(protocol.KeySpec)
		kh := sha256.Sum256(operator)
		spec.PublicKeyHash = kh[:]
		page.AddKeySpec(spec)
	}

	page.AcceptThreshold = b.globals.Globals.OperatorAcceptThreshold.Threshold(len(page.Keys))
	b.WriteRecords(book, page)
}

func (b *bootstrap) WriteRecords(record ...protocol.Account) {
	b.records = append(b.records, record...)
	for _, rec := range record {
		b.urls = append(b.urls, rec.GetUrl())
	}
}

func (b *bootstrap) writeDataRecord(account *protocol.DataAccount, url *url.URL, dataRecord DataRecord) {
	b.records = append(b.records, account)
	b.urls = append(b.urls, url)
	b.dataRecords = append(b.dataRecords, dataRecord)
}
