package genesis

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/big"
	"path"
	"time"

	"github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	"gitlab.com/accumulatenetwork/accumulate/config"
	"gitlab.com/accumulatenetwork/accumulate/internal/block"
	"gitlab.com/accumulatenetwork/accumulate/internal/chain"
	"gitlab.com/accumulatenetwork/accumulate/internal/core"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/encoding"
	"gitlab.com/accumulatenetwork/accumulate/internal/errors"
	"gitlab.com/accumulatenetwork/accumulate/internal/routing"
	"gitlab.com/accumulatenetwork/accumulate/internal/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	"gitlab.com/accumulatenetwork/accumulate/smt/storage"
	"gitlab.com/accumulatenetwork/accumulate/smt/storage/memory"
)

type NetworkValidatorMap map[string][]tmtypes.GenesisValidator

type InitOpts struct {
	Network             config.Network
	AllConfigs          []*config.Config
	Validators          []tmtypes.GenesisValidator
	NetworkValidatorMap NetworkValidatorMap
	GenesisTime         time.Time
	Logger              log.Logger
	FactomAddressesFile string
	Keys                [][]byte
}

func Init(kvdb storage.KeyValueStore, opts InitOpts) (Bootstrap, error) {
	b := &bootstrap{
		InitOpts:    opts,
		kvdb:        kvdb,
		db:          database.New(kvdb, opts.Logger.With("module", "database")),
		dataRecords: make([]DataRecord, 0),
		records:     make([]protocol.Account, 0),
	}

	// Add validator keys to NetworkValidatorMap when not there
	if b.InitOpts.NetworkValidatorMap == nil {
		panic("NetworkValidatorMap is not present")
	}
	if _, ok := b.InitOpts.NetworkValidatorMap[b.InitOpts.Network.LocalSubnetID]; !ok {
		b.InitOpts.NetworkValidatorMap[b.InitOpts.Network.LocalSubnetID] = b.InitOpts.Validators
	}

	// Build the routing table
	var err error
	b.router, b.routingTable, err = routing.NewSimpleRouter(&opts.Network, nil)
	if err != nil {
		return nil, err
	}

	b.genesisExec, err = block.NewGenesisExecutor(b.db, opts.Logger, opts.Network, b.router)
	if err != nil {
		return nil, err
	}

	return b, nil
}

type Bootstrap interface {
	Bootstrap() error
	GetDBState() ([]byte, error)
}

type bootstrap struct {
	InitOpts
	kvdb         storage.KeyValueStore
	db           *database.Database
	block        *block.Block
	nodeUrl      *url.URL
	authorityUrl *url.URL
	urls         []*url.URL
	records      []protocol.Account
	dataRecords  []DataRecord
	genesisExec  *block.Executor
	router       routing.Router
	routingTable *protocol.RoutingTable
	globals      *core.GlobalValues
}

func (b *bootstrap) Bootstrap() error {
	b.block = new(block.Block)
	b.block.Index = protocol.GenesisBlock
	b.block.Time = b.InitOpts.GenesisTime
	b.block.Batch = b.db.Begin(true)
	defer b.block.Batch.Discard()

	err := b.genesisExec.Genesis(b.block, b)
	if err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}

	err = b.block.Batch.Commit()
	if err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}

	batch := b.db.Begin(false)
	defer batch.Discard()
	err = b.writeGenesisFile(batch.BptRoot())
	if err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}
	return nil
}

func (b *bootstrap) GetDBState() ([]byte, error) {
	memDb, ok := b.kvdb.(*memory.DB)

	var state []byte
	var err error
	if ok {
		state, err = memDb.MarshalJSON()
		if err != nil {
			return nil, nil
		}
	}

	return state, err
}

type DataRecord struct {
	Account *url.URL
	Entry   protocol.DataEntry
}

var _ chain.TransactionExecutor = &bootstrap{}
var _ chain.PrincipalValidator = &bootstrap{}

func (bootstrap) Type() protocol.TransactionType {
	return protocol.TransactionTypeSyntheticDepositTokens
}

func (b *bootstrap) AllowMissingPrincipal(*protocol.Transaction) (allow, fallback bool) {
	return true, false
}

func (b *bootstrap) Execute(st *chain.StateManager, tx *chain.Delivery) (protocol.TransactionResult, error) {
	return b.Validate(st, tx)
}

func (b *bootstrap) Validate(st *chain.StateManager, tx *chain.Delivery) (protocol.TransactionResult, error) {
	b.nodeUrl = b.InitOpts.Network.NodeUrl()
	b.authorityUrl = b.nodeUrl.JoinPath(protocol.ValidatorBook)
	b.globals = new(core.GlobalValues)

	// set the initial price to 1/5 fct price * 1/4 market cap dilution = 1/20 fct price
	// for this exercise, we'll assume that 1 FCT = $1, so initial ACME price is $0.05
	b.globals.Oracle = new(protocol.AcmeOracle)
	b.globals.Oracle.Price = uint64(protocol.InitialAcmeOracleValue)

	b.globals.Globals = new(protocol.NetworkGlobals)
	b.globals.Globals.ValidatorThreshold.Set(2, 3)

	if b.InitOpts.NetworkValidatorMap != nil {
		b.globals.Network = b.buildNetworkDefinition()
	}

	b.createADI()
	b.createValidatorBook()

	b.createMainLedger()
	b.createSyntheticLedger()
	b.createAnchorPool()

	err := b.createVoteScratchChain()
	if err != nil {
		return nil, err
	}

	b.createEvidenceChain()

	// Write the global data accounts
	err = b.globals.Store(&b.Network, func(accountUrl *url.URL, target interface{}) error {
		da := new(protocol.DataAccount)
		da.Url = accountUrl
		da.AddAuthority(b.authorityUrl) // TODO Lock BVN accounts so they can't be updated directly
		return encoding.SetPtr(da, target)
	}, func(account protocol.Account) error {
		b.WriteRecords(account)
		return nil
	})
	if err != nil {
		return nil, err
	}

	err = b.createRouting()
	if err != nil {
		return nil, err
	}

	switch b.InitOpts.Network.Type {
	case config.Directory:
		err = b.initDN()
	case config.BlockValidator:
		err = b.initBVN()
	}
	if err != nil {
		return nil, err
	}

	err = st.Create(b.records...)
	if err != nil {
		return nil, fmt.Errorf("failed to create records: %w", err)
	}

	for _, wd := range b.dataRecords {
		body := new(protocol.SystemWriteData)
		body.Entry = wd.Entry
		txn := new(protocol.Transaction)
		txn.Header.Principal = wd.Account
		txn.Body = body
		st.State.ProcessAdditionalTransaction(tx.NewInternal(txn))
	}

	return nil, st.AddDirectoryEntry(b.nodeUrl, b.urls...)
}

func (b *bootstrap) createADI() {
	// Create the ADI
	adi := new(protocol.ADI)
	adi.Url = b.nodeUrl
	adi.AddAuthority(b.authorityUrl)
	b.WriteRecords(adi)
}

func (b *bootstrap) createValidatorBook() {
	uBook := b.authorityUrl
	book := new(protocol.KeyBook)
	book.Url = uBook
	book.BookType = protocol.BookTypeValidator
	book.AddAuthority(uBook)
	book.PageCount = 1

	page := createOperatorPage(uBook, 0, b.InitOpts.NetworkValidatorMap, true)
	b.WriteRecords(book, page)
}

func (b *bootstrap) createMainLedger() {
	// Create the main ledger
	ledger := new(protocol.SystemLedger)
	ledger.Url = b.nodeUrl.JoinPath(protocol.Ledger)
	ledger.Index = protocol.GenesisBlock
	b.WriteRecords(ledger)
}

func (b *bootstrap) createSyntheticLedger() {
	// Create the synth ledger
	synthLedger := new(protocol.SyntheticLedger)
	synthLedger.Url = b.nodeUrl.JoinPath(protocol.Synthetic)
	b.WriteRecords(synthLedger)
}

func (b *bootstrap) createAnchorPool() {
	// Create the anchor pool
	anchorLedger := new(protocol.AnchorLedger)
	anchorLedger.Url = b.nodeUrl.JoinPath(protocol.AnchorPool)

	if b.Network.Type == config.Directory {
		// Initialize the last major block time to prevent a major block from
		// being created immediately once the network boots
		anchorLedger.MajorBlockTime = b.InitOpts.GenesisTime
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

	da := new(protocol.DataAccount)
	da.Scratch = true
	da.Url = b.nodeUrl.JoinPath(protocol.Votes)
	da.AddAuthority(b.authorityUrl)
	b.writeDataRecord(da, da.Url, DataRecord{da.Url, wd.Entry})
	return nil
}

func (b *bootstrap) createEvidenceChain() {
	//create an evidence scratch chain
	da := new(protocol.DataAccount)
	da.Scratch = true
	da.Url = b.nodeUrl.JoinPath(protocol.Evidence)
	da.AddAuthority(b.authorityUrl)
	b.WriteRecords(da)
	b.urls = append(b.urls, da.Url)
}

func (b *bootstrap) createRouting() error {
	// Create an account for the routing table
	account := new(protocol.DataAccount)
	account.Url = b.nodeUrl.JoinPath(protocol.Routing)
	account.AddAuthority(b.authorityUrl)

	data, err := b.routingTable.MarshalBinary()
	if err != nil {
		return errors.Format(errors.StatusInternalError, "marshal routing table: %w", err)
	}

	account.Entry = &protocol.AccumulateDataEntry{Data: [][]byte{data}}
	b.WriteRecords(account)
	return nil
}

func (b *bootstrap) initDN() error {
	b.createDNOperatorBook()

	acme := new(protocol.TokenIssuer)
	acme.AddAuthority(b.authorityUrl)
	acme.Url = protocol.AcmeUrl()
	acme.Precision = 8
	acme.Symbol = "ACME"
	b.WriteRecords(acme)

	if protocol.IsTestNet {
		// On the TestNet, set the issued amount to the faucet balance
		acme.Issued.SetString(protocol.AcmeFaucetBalance, 10)
	} else {
		// On the MainNet, set the supply limit
		acme.SupplyLimit = big.NewInt(protocol.AcmeSupplyLimit * protocol.AcmePrecision)
	}
	return nil
}

func (b *bootstrap) initBVN() error {
	// Verify that the BVN ID will make a valid subnet URL
	network := b.InitOpts.Network
	if err := protocol.IsValidAdiUrl(protocol.SubnetUrl(network.LocalSubnetID), true); err != nil {
		panic(fmt.Errorf("%q is not a valid subnet ID: %v", network.LocalSubnetID, err))
	}

	b.createBVNOperatorBook(b.nodeUrl, b.InitOpts.NetworkValidatorMap)

	subnet, err := b.router.RouteAccount(protocol.FaucetUrl)
	if err == nil && subnet == network.LocalSubnetID {
		liteId := new(protocol.LiteIdentity)
		liteId.Url = protocol.FaucetUrl.RootIdentity()

		liteToken := new(protocol.LiteTokenAccount)
		liteToken.Url = protocol.FaucetUrl
		liteToken.TokenUrl = protocol.AcmeUrl()
		liteToken.Balance.SetString(protocol.AcmeFaucetBalance, 10)
		b.WriteRecords(liteId, liteToken)
	}
	if b.InitOpts.FactomAddressesFile != "" {
		factomAddresses, err := LoadFactomAddressesAndBalances(b.InitOpts.FactomAddressesFile)
		if err != nil {
			return errors.Wrap(errors.StatusUnknown, err)
		}
		for _, factomAddress := range factomAddresses {
			subnet, err := b.router.RouteAccount(factomAddress.Address)
			if err == nil && subnet == network.LocalSubnetID {
				lite := new(protocol.LiteTokenAccount)
				lite.Url = factomAddress.Address
				lite.TokenUrl = protocol.AcmeUrl()
				lite.Balance = *big.NewInt(5 * factomAddress.Balance)
				b.WriteRecords(lite)
			}
		}
	}
	return nil
}

func (b *bootstrap) createDNOperatorBook() {
	book := new(protocol.KeyBook)
	book.Url = b.nodeUrl.JoinPath(protocol.OperatorBook)
	book.AddAuthority(book.Url)
	book.PageCount = 1

	page := createOperatorPage(book.Url, 0, b.InitOpts.NetworkValidatorMap, false)
	b.WriteRecords(book, page)
}

func (b *bootstrap) createBVNOperatorBook(nodeUrl *url.URL, operators NetworkValidatorMap) {
	book := new(protocol.KeyBook)
	book.Url = nodeUrl.JoinPath(protocol.OperatorBook)
	book.AddAuthority(book.Url)
	book.PageCount = 2

	page1 := new(protocol.KeyPage)
	page1.Url = protocol.FormatKeyPageUrl(book.Url, 0)
	page1.AcceptThreshold = 1
	page1.Version = 1
	page1.Keys = make([]*protocol.KeySpec, 1)
	spec := new(protocol.KeySpec)
	spec.Delegate = protocol.DnUrl().JoinPath(protocol.OperatorBook)
	page1.Keys[0] = spec

	page2 := createOperatorPage(book.Url, 1, operators, false)
	blacklistTxsForPage(page2, protocol.TransactionTypeUpdateKeyPage, protocol.TransactionTypeUpdateAccountAuth)

	b.WriteRecords(book, page1, page2)
}

func createOperatorPage(uBook *url.URL, pageIndex uint64, operators NetworkValidatorMap, validatorsOnly bool) *protocol.KeyPage {
	page := new(protocol.KeyPage)
	page.Url = protocol.FormatKeyPageUrl(uBook, pageIndex)
	page.Version = 1

	if validatorsOnly {
		subnet, ok := protocol.ParseSubnetUrl(uBook)
		if !ok {
			panic("book URL does not belong to a subnet")
		}

		operators, ok := operators[subnet]
		if !ok {
			panic("missing operators for subnet")
		}

		for _, operator := range operators {
			/* TODO
			Determine which operators are also validators and which not. Followers should be omitted,
			but DNs which also don't have voting power not.	(DNs need to sign Oracle updates)
			*/
			spec := new(protocol.KeySpec)
			kh := sha256.Sum256(operator.PubKey.Bytes())
			spec.PublicKeyHash = kh[:]
			page.AddKeySpec(spec)
		}

	} else {
		for _, operators := range operators {
			for _, operator := range operators {
				spec := new(protocol.KeySpec)
				kh := sha256.Sum256(operator.PubKey.Bytes())
				spec.PublicKeyHash = kh[:]
				page.AddKeySpec(spec)
			}
		}
	}

	page.AcceptThreshold = protocol.GetValidatorsMOfN(len(page.Keys), protocol.FallbackValidatorThreshold)
	return page
}

func blacklistTxsForPage(page *protocol.KeyPage, txTypes ...protocol.TransactionType) {
	page.TransactionBlacklist = new(protocol.AllowedTransactions)
	for _, txType := range txTypes {
		bit, ok := txType.AllowedTransactionBit()
		if !ok {
			panic(fmt.Errorf("failed to blacklist %v", txType))
		}
		page.TransactionBlacklist.Set(bit)
	}
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

func (b *bootstrap) writeGenesisFile(appHash []byte) error {
	state, err := b.GetDBState()
	if err != nil {
		return errors.Wrap(errors.StatusUnknown, err)
	}

	genDoc := &tmtypes.GenesisDoc{
		ChainID:         b.InitOpts.Network.LocalSubnetID,
		GenesisTime:     b.InitOpts.GenesisTime,
		InitialHeight:   protocol.GenesisBlock + 1,
		Validators:      b.InitOpts.Validators,
		ConsensusParams: tmtypes.DefaultConsensusParams(),
		AppState:        state,
		AppHash:         appHash,
	}

	for _, config := range b.InitOpts.AllConfigs {
		if err := genDoc.SaveAs(path.Join(config.RootDir, config.Genesis)); err != nil {
			return fmt.Errorf("failed to save gen doc: %v", err)
		}
	}
	return nil
}

func (b *bootstrap) buildNetworkDefinition() *protocol.NetworkDefinition {
	netDef := new(protocol.NetworkDefinition)

	for _, subnet := range b.InitOpts.Network.Subnets {

		// Add the validator hashes from the subnet's genesis doc
		var vkHashes [][32]byte
		for _, validator := range b.InitOpts.NetworkValidatorMap[subnet.ID] {
			pkh := sha256.Sum256(validator.PubKey.Bytes())
			vkHashes = append(vkHashes, pkh)
		}

		subnetDef := protocol.SubnetDefinition{
			SubnetID:           subnet.ID,
			ValidatorKeyHashes: vkHashes,
		}
		netDef.Subnets = append(netDef.Subnets, subnetDef)
	}
	return netDef
}
