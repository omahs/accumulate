package simulator

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	"gitlab.com/accumulatenetwork/accumulate/config"
	"gitlab.com/accumulatenetwork/accumulate/internal/api/v2"
	"gitlab.com/accumulatenetwork/accumulate/internal/block"
	. "gitlab.com/accumulatenetwork/accumulate/internal/block"
	"gitlab.com/accumulatenetwork/accumulate/internal/chain"
	"gitlab.com/accumulatenetwork/accumulate/internal/client"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/errors"
	"gitlab.com/accumulatenetwork/accumulate/internal/logging"
	"gitlab.com/accumulatenetwork/accumulate/internal/routing"
	acctesting "gitlab.com/accumulatenetwork/accumulate/internal/testing"
	"gitlab.com/accumulatenetwork/accumulate/internal/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	"gitlab.com/accumulatenetwork/accumulate/types/api/query"
	"golang.org/x/sync/errgroup"
)

type Simulator struct {
	tb
	Logger    log.Logger
	Subnets   []config.Subnet
	Executors map[string]*ExecEntry

	LogLevels string

	routingOverrides map[[32]byte]string
}

func (s *Simulator) newLogger() log.Logger {
	levels := s.LogLevels
	if levels == "" {
		levels = acctesting.DefaultLogLevels
	}

	if !acctesting.LogConsole {
		return logging.NewTestLogger(s, "plain", levels, false)
	}

	w, err := logging.NewConsoleWriter("plain")
	require.NoError(s, err)
	level, writer, err := logging.ParseLogLevel(levels, w)
	require.NoError(s, err)
	logger, err := logging.NewTendermintLogger(zerolog.New(writer), level, false)
	require.NoError(s, err)
	return logger
}

func New(t TB, bvnCount int) *Simulator {
	t.Helper()
	sim := new(Simulator)
	sim.TB = t
	sim.Setup(bvnCount)
	return sim
}

func (s *Simulator) Setup(bvnCount int) {
	s.Helper()

	// Initialize the simulartor and network
	s.routingOverrides = map[[32]byte]string{}
	s.Logger = s.newLogger().With("module", "simulator")
	s.Executors = map[string]*ExecEntry{}
	s.Subnets = make([]config.Subnet, bvnCount+1)
	s.Subnets[0] = config.Subnet{Type: config.Directory, ID: protocol.Directory}
	for i := 0; i < bvnCount; i++ {
		s.Subnets[i+1] = config.Subnet{Type: config.BlockValidator, ID: fmt.Sprintf("BVN%d", i)}
	}

	// Initialize each executor
	for i := range s.Subnets {
		subnet := &s.Subnets[i]
		subnet.Nodes = []config.Node{{Type: config.Validator, Address: subnet.ID}}

		logger := s.newLogger().With("subnet", subnet.ID)
		key := acctesting.GenerateKey(s.Name(), subnet.ID)
		db := database.OpenInMemory(logger)

		network := config.Network{
			Type:          subnet.Type,
			LocalSubnetID: subnet.ID,
			LocalAddress:  subnet.ID,
			Subnets:       s.Subnets,
		}

		exec, err := NewNodeExecutor(ExecutorOptions{
			Logger:  logger,
			Key:     key,
			Network: network,
			Router:  s.Router(),
		}, db)
		require.NoError(s, err)

		jrpc, err := api.NewJrpc(api.Options{
			Logger:        logger,
			Network:       &network,
			Router:        s.Router(),
			TxMaxWaitTime: time.Hour,
		})
		require.NoError(s, err)

		s.Executors[subnet.ID] = &ExecEntry{
			Database: db,
			Executor: exec,
			API:      acctesting.DirectJrpcClient(jrpc),
		}
	}
}

func (s *Simulator) SetRouteFor(account *url.URL, subnet string) {
	// Make sure the account is a root identity
	if !account.RootIdentity().Equal(account) {
		s.Fatalf("Cannot set the route for a non-root: %v", account)
	}

	// Make sure the subnet exists
	s.Subnet(subnet)

	// Add/remove the override
	if subnet == "" {
		delete(s.routingOverrides, account.AccountID32())
	} else {
		s.routingOverrides[account.AccountID32()] = subnet
	}
}

func (s *Simulator) Router() routing.Router {
	return router{s}
}

func (s *Simulator) Subnet(id string) *ExecEntry {
	e, ok := s.Executors[id]
	require.Truef(s, ok, "Unknown subnet %q", id)
	return e
}

func (s *Simulator) SubnetFor(url *url.URL) *ExecEntry {
	s.Helper()

	subnet, err := s.Router().RouteAccount(url)
	require.NoError(s, err)
	return s.Subnet(subnet)
}

func (s *Simulator) Query(url *url.URL, req query.Request, prove bool) interface{} {
	s.Helper()

	x := s.SubnetFor(url)
	return Query(s, x.Database, x.Executor, req, prove)
}

func (s *Simulator) InitFromGenesis() {
	s.Helper()

	for _, subnet := range s.Subnets {
		x := s.Subnet(subnet.ID)
		InitFromGenesis(s, x.Database, x.Executor)
	}
}

func (s *Simulator) InitFromSnapshot(filename func(string) string) {
	s.Helper()

	for _, subnet := range s.Subnets {
		x := s.Subnet(subnet.ID)
		InitFromSnapshot(s, x.Database, x.Executor, filename(subnet.ID))
	}
}

// ExecuteBlock executes a block after submitting envelopes. If a status channel
// is provided, statuses will be sent to the channel as transactions are
// executed. Once the block is complete, the status channel will be closed (if
// provided).
func (s *Simulator) ExecuteBlock(statusChan chan<- *protocol.TransactionStatus) {
	s.Helper()

	if statusChan != nil {
		defer close(statusChan)
	}

	errg := new(errgroup.Group)
	for _, subnet := range s.Subnets {
		x := s.Subnet(subnet.ID)
		submitted := x.TakeSubmitted()
		errg.Go(func() error {
			status, err := ExecuteBlock(s, x.Database, x.Executor, nil, submitted...)
			if err != nil {
				return err
			}
			if statusChan == nil {
				return nil
			}

			for _, status := range status {
				statusChan <- status
			}
			return nil
		})
	}

	// Wait for all subnets to complete
	err := errg.Wait()
	require.NoError(tb{s}, err)
}

// ExecuteBlocks executes a number of blocks. This is useful for things like
// waiting for a block to be anchored.
func (s *Simulator) ExecuteBlocks(n int) {
	for ; n > 0; n-- {
		s.ExecuteBlock(nil)
	}
}

func (s *Simulator) Submit(envelopes ...*protocol.Envelope) ([]*protocol.Envelope, error) {
	s.Helper()

	for _, envelope := range envelopes {
		// Route
		subnet, err := s.Router().Route(envelope)
		require.NoError(s, err)
		x := s.Subnet(subnet)

		// Normalize - use a copy to avoid weird issues caused by modifying values
		deliveries, err := chain.NormalizeEnvelope(envelope.Copy())
		if err != nil {
			return nil, err
		}

		// Check
		for _, delivery := range deliveries {
			_, err = CheckTx(s, x.Database, x.Executor, delivery)
			if err != nil {
				return nil, err
			}
		}

		// Enqueue
		x.Submit(envelope)
	}

	return envelopes, nil
}

// MustSubmitAndExecuteBlock executes a block with the envelopes and fails the test if
// any envelope fails.
func (s *Simulator) MustSubmitAndExecuteBlock(envelopes ...*protocol.Envelope) []*protocol.Envelope {
	s.Helper()

	status, err := s.SubmitAndExecuteBlock(envelopes...)
	require.NoError(tb{s}, err)

	var didFail bool
	for _, status := range status {
		if status.Code == 0 {
			continue
		}

		assert.Zero(s, status.Code, status.Message)
		didFail = true
	}
	if didFail {
		s.FailNow()
	}
	return envelopes
}

// SubmitAndExecuteBlock executes a block with the envelopes.
func (s *Simulator) SubmitAndExecuteBlock(envelopes ...*protocol.Envelope) ([]*protocol.TransactionStatus, error) {
	s.Helper()

	_, err := s.Submit(envelopes...)
	if err != nil {
		return nil, errors.Wrap(errors.StatusUnknown, err)
	}

	ids := map[[32]byte]bool{}
	for _, env := range envelopes {
		for _, d := range NormalizeEnvelope(s, env) {
			ids[*(*[32]byte)(d.Transaction.GetHash())] = true
		}
	}

	ch1 := make(chan *protocol.TransactionStatus)
	ch2 := make(chan *protocol.TransactionStatus)
	go func() {
		s.ExecuteBlock(ch1)
		s.ExecuteBlock(ch2)
	}()

	status := make([]*protocol.TransactionStatus, 0, len(envelopes))
	for s := range ch1 {
		if ids[s.For] {
			status = append(status, s)
		}
	}
	for s := range ch2 {
		if ids[s.For] {
			status = append(status, s)
		}
	}

	return status, nil
}

func (s *Simulator) findTxn(status func(*protocol.TransactionStatus) bool, hash []byte) *ExecEntry {
	s.Helper()

	for _, subnet := range s.Subnets {
		x := s.Subnet(subnet.ID)

		batch := x.Database.Begin(false)
		defer batch.Discard()
		obj, err := batch.Transaction(hash).GetStatus()
		require.NoError(s, err)
		if status(obj) {
			return x
		}
	}

	return nil
}

func (s *Simulator) WaitForTransactions(status func(*protocol.TransactionStatus) bool, envelopes ...*protocol.Envelope) ([]*protocol.TransactionStatus, []*protocol.Transaction) {
	s.Helper()

	var statuses []*protocol.TransactionStatus
	var transactions []*protocol.Transaction
	for _, envelope := range envelopes {
		for _, delivery := range NormalizeEnvelope(s, envelope) {
			st, txn := s.WaitForTransactionFlow(status, delivery.Transaction.GetHash())
			statuses = append(statuses, st...)
			transactions = append(transactions, txn...)
		}
	}
	return statuses, transactions
}

func (s *Simulator) WaitForTransactionFlow(statusCheck func(*protocol.TransactionStatus) bool, txnHash []byte) ([]*protocol.TransactionStatus, []*protocol.Transaction) {
	s.Helper()

	var x *ExecEntry
	for i := 0; i < 50; i++ {
		x = s.findTxn(statusCheck, txnHash)
		if x != nil {
			break
		}

		s.ExecuteBlock(nil)
	}
	if x == nil {
		s.Errorf("Transaction %X has not been delivered after 50 blocks", txnHash[:4])
		s.FailNow()
		panic("unreachable")
	}

	batch := x.Database.Begin(false)
	synth, err1 := batch.Transaction(txnHash).GetSyntheticTxns()
	state, err2 := batch.Transaction(txnHash).GetState()
	status, err3 := batch.Transaction(txnHash).GetStatus()
	batch.Discard()
	require.NoError(s, err1)
	require.NoError(s, err2)
	require.NoError(s, err3)

	status.For = *(*[32]byte)(txnHash)
	statuses := []*protocol.TransactionStatus{status}
	transactions := []*protocol.Transaction{state.Transaction}
	for _, id := range synth.Hashes {
		// Wait for synthetic transactions to be delivered
		st, txn := s.WaitForTransactionFlow(func(status *protocol.TransactionStatus) bool {
			return status.Delivered
		}, id[:]) //nolint:rangevarref
		statuses = append(statuses, st...)
		transactions = append(transactions, txn...)
	}

	return statuses, transactions
}

type ExecEntry struct {
	mu                      sync.Mutex
	nextBlock, currentBlock []*protocol.Envelope

	Database *database.Database
	Executor *block.Executor
	API      *client.Client

	// SubmitHook can be used to control how envelopes are submitted to the
	// subnet. It is not safe to change SubmitHook concurrently with calls to
	// Submit.
	SubmitHook func([]*protocol.Envelope) []*protocol.Envelope
}

// Submit adds the envelopes to the next block's queue.
//
// By adding transactions to the next block and swaping queues when a block is
// executed, we roughly simulate the process Tendermint uses to build blocks.
func (s *ExecEntry) Submit(envelopes ...*protocol.Envelope) {
	// Capturing the field in a variable is more concurrency safe than using the
	// field directly
	if h := s.SubmitHook; h != nil {
		envelopes = h(envelopes)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextBlock = append(s.nextBlock, envelopes...)
}

// TakeSubmitted returns the envelopes for the current block.
func (s *ExecEntry) TakeSubmitted() []*protocol.Envelope {
	s.mu.Lock()
	defer s.mu.Unlock()
	submitted := s.currentBlock
	s.currentBlock = s.nextBlock
	s.nextBlock = nil
	return submitted
}
