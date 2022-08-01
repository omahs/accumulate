package e2e

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/AccumulateNetwork/jsonrpc2/v15"
	"github.com/stretchr/testify/require"
	"gitlab.com/accumulatenetwork/accumulate/internal/api/v2"
	"gitlab.com/accumulatenetwork/accumulate/internal/block/simulator"
	acctesting "gitlab.com/accumulatenetwork/accumulate/internal/testing"
	"gitlab.com/accumulatenetwork/accumulate/internal/url"
	. "gitlab.com/accumulatenetwork/accumulate/protocol"
)

func TestOverwriteCreditBalance(t *testing.T) {
	const x = 0.05
	const y = 10000
	big.NewInt(x * y)

	// Tests AC-1555
	var timestamp uint64

	// Initialize
	sim := simulator.New(t, 3)
	sim.InitFromGenesis()

	// Setup accounts
	const initialBalance = 100
	alice := AccountUrl("alice")
	aliceKey, liteKey := acctesting.GenerateKey(alice), acctesting.GenerateKey("lite")
	liteUrl := acctesting.AcmeLiteAddressStdPriv(liteKey)
	sim.CreateAccount(&LiteIdentity{Url: liteUrl.RootIdentity(), CreditBalance: 1e9})
	sim.CreateAccount(&LiteTokenAccount{Url: liteUrl, TokenUrl: AcmeUrl(), Balance: *big.NewInt(1e9 * AcmePrecision)})
	sim.CreateIdentity(alice, aliceKey[32:])
	updateAccount(sim, alice.JoinPath("book", "1"), func(page *KeyPage) { page.CreditBalance = initialBalance * CreditPrecision })

	// Add credits
	const additionalBalance = 99
	const oracle = InitialAcmeOracleValue
	acme := big.NewInt(AcmePrecision)
	acme.Mul(acme, big.NewInt(additionalBalance))
	acme.Div(acme, big.NewInt(CreditsPerDollar))
	acme.Mul(acme, big.NewInt(AcmeOraclePrecision))
	acme.Div(acme, big.NewInt(oracle))
	sim.WaitForTransactions(delivered, sim.MustSubmitAndExecuteBlock(
		acctesting.NewTransaction().
			WithPrincipal(liteUrl).
			WithSigner(liteUrl, 1).
			WithTimestampVar(&timestamp).
			WithBody(&AddCredits{
				Recipient: alice.JoinPath("book", "1"),
				Amount:    *acme,
				Oracle:    oracle,
			}).
			Initiate(SignatureTypeED25519, liteKey).
			Build(),
	)...)

	// The balance should be added
	page := simulator.GetAccount[*KeyPage](sim, alice.JoinPath("book", "1"))
	require.Equal(t, int((initialBalance+additionalBalance)*CreditPrecision), int(page.CreditBalance))
}

func TestQueryKeyIndexWithRemoteAuthority(t *testing.T) {
	// var timestamp uint64

	// Initialize
	sim := simulator.New(t, 3)
	sim.InitFromGenesis()

	// Setup
	alice, bob := AccountUrl("alice"), AccountUrl("bob")
	aliceKey, bobKey := acctesting.GenerateKey(alice), acctesting.GenerateKey(bob)
	sim.SetRouteFor(alice, "BVN0")
	sim.SetRouteFor(bob, "BVN1")
	sim.CreateIdentity(alice, aliceKey[32:])
	sim.CreateIdentity(bob, bobKey[32:])

	sim.CreateAccount(&TokenAccount{
		Url:         alice.JoinPath("managed-tokens"),
		AccountAuth: AccountAuth{Authorities: []AuthorityEntry{{Url: bob.JoinPath("book")}}},
		TokenUrl:    AcmeUrl(),
	})

	// Query key
	req := new(api.KeyPageIndexQuery)
	req.Url = alice.JoinPath("managed-tokens")
	req.Key = aliceKey[32:]
	_, err := sim.PartitionFor(req.Url).API.QueryKeyPageIndex(context.Background(), req)
	require.Error(t, err)
	require.IsType(t, jsonrpc2.Error{}, err)
	require.Equal(t, err.(jsonrpc2.Error).Data, fmt.Sprintf("no authority of %s holds %X", req.Url, req.Key))
}

func TestAddCreditsToLiteIdentityOnOtherBVN(t *testing.T) {
	// Tests AC-1859
	var timestamp uint64

	// Initialize
	sim := simulator.New(t, 3)
	sim.InitFromGenesis()

	// Setup
	sendKey, recvKey := acctesting.GenerateKey("sender"), acctesting.GenerateKey("receiver")
	sender, receiver := acctesting.AcmeLiteAddressStdPriv(sendKey), acctesting.AcmeLiteAddressStdPriv(recvKey)
	sim.SetRouteFor(sender.RootIdentity(), "BVN0")
	sim.SetRouteFor(receiver.RootIdentity(), "BVN1")
	sim.CreateAccount(&LiteIdentity{Url: sender.RootIdentity(), CreditBalance: 1e9})
	sim.CreateAccount(&LiteTokenAccount{Url: sender, Balance: *big.NewInt(1e12), TokenUrl: AcmeUrl()})

	// Add credits
	const creditAmount = 99
	const oracle = InitialAcmeOracleValue
	acme := big.NewInt(AcmePrecision)
	acme.Mul(acme, big.NewInt(creditAmount))
	acme.Div(acme, big.NewInt(CreditsPerDollar))
	acme.Mul(acme, big.NewInt(AcmeOraclePrecision))
	acme.Div(acme, big.NewInt(oracle))
	sim.WaitForTransactions(delivered, sim.MustSubmitAndExecuteBlock(
		acctesting.NewTransaction().
			WithPrincipal(sender).
			WithSigner(sender, 1).
			WithTimestampVar(&timestamp).
			WithBody(&AddCredits{
				Recipient: receiver,
				Amount:    *acme,
				Oracle:    oracle,
			}).
			Initiate(SignatureTypeED25519, sendKey).
			Build(),
	)...)

	// Verify
	recvId := simulator.GetAccount[*LiteIdentity](sim, receiver.RootIdentity())
	require.Equal(t, int(creditAmount*CreditPrecision), int(recvId.CreditBalance))
}

func TestSynthTxnWithMissingPrincipal(t *testing.T) {
	// Tests AC-1704
	var timestamp uint64

	// Initialize
	sim := simulator.New(t, 3)
	sim.InitFromGenesis()

	// Setup a lite token account for a token type that does not exist
	liteKey := acctesting.GenerateKey("Lite")
	lite := sim.CreateLiteTokenAccount(liteKey, url.MustParse("fake.acme/tokens"), 1e9, 1)

	// Burn credits
	txn := sim.MustSubmitAndExecuteBlock(
		acctesting.NewTransaction().
			WithPrincipal(lite).
			WithSigner(lite, 1).
			WithTimestampVar(&timestamp).
			WithBody(&BurnTokens{
				Amount: *big.NewInt(1),
			}).
			Initiate(SignatureTypeED25519, liteKey).
			Build(),
	)
	_, _, synth := sim.WaitForTransaction(delivered, txn[0].Transaction[0].GetHash(), 50)

	// The synthetic transaction should be received and marked as pending
	require.Len(t, synth, 1)
	hash := synth[0].Hash()
	_, status, _ := sim.WaitForTransaction(received, hash[:], 50)
	require.True(t, status.Pending(), "The transaction was delivered prematurely")

	// The synthetic transaction must fail, but only after the anchor is received
	_, status, _ = sim.WaitForTransaction(delivered, hash[:], 50)
	require.NotZero(t, status.Code, "The transaction did not fail")
}

func TestFaucetMultiNetwork(t *testing.T) {
	// Initialize
	sim := simulator.New(t, 3)
	sim.InitFromGenesis()

	// Setup
	liteKey := acctesting.GenerateKey("Lite")
	lite := sim.CreateLiteTokenAccount(liteKey, AcmeUrl(), 1e9, 1e12)

	// Set the lite account routing to a different BVN from the faucet
	faucetBvn := sim.PartitionFor(FaucetUrl)
	for _, partition := range sim.Partitions[1:] {
		if faucetBvn.Partition.Id != partition.Id {
			sim.SetRouteFor(lite.RootIdentity(), partition.Id)
			break
		}
	}

	// Execute
	resp, err := sim.Executors[Directory].API.Faucet(context.Background(), &AcmeFaucet{Url: lite})
	require.NoError(t, err)
	sim.WaitForTransactionFlow(delivered, resp.TransactionHash)

	// Verify
	lta := simulator.GetAccount[*LiteTokenAccount](sim, lite)
	require.NotZero(t, lta.Balance.Int64())
}

func TestSigningDeliveredTxnDoesNothing(t *testing.T) {
	var timestamp uint64

	// Initialize
	sim := simulator.New(t, 3)
	sim.InitFromGenesis()

	// Setup
	aliceKey, bobKey := acctesting.GenerateKey("Alice"), acctesting.GenerateKey("Bob")
	alice := sim.CreateLiteTokenAccount(aliceKey, AcmeUrl(), 1e9, 2)
	bob := acctesting.AcmeLiteAddressStdPriv(bobKey)

	// Execute
	_, txns := sim.WaitForTransactions(delivered, sim.MustSubmitAndExecuteBlock(
		acctesting.NewTransaction().
			WithPrincipal(alice).
			WithTimestampVar(&timestamp).
			WithSigner(alice, 1).
			WithBody(&SendTokens{
				To: []*TokenRecipient{{
					Url:    bob,
					Amount: *big.NewInt(1),
				}},
			}).
			Initiate(SignatureTypeED25519, aliceKey).
			Build(),
	)...)

	// Verify
	require.Equal(t, 1, int(simulator.GetAccount[*LiteTokenAccount](sim, alice).Balance.Int64()))
	require.Equal(t, 1, int(simulator.GetAccount[*LiteTokenAccount](sim, bob).Balance.Int64()))

	// Send another signature
	_, err := sim.SubmitAndExecuteBlock(
		acctesting.NewTransaction().
			WithPrincipal(alice).
			WithSigner(alice, 1).
			WithTxnHash(txns[0].GetHash()).
			Sign(SignatureTypeED25519, aliceKey).
			Build(),
	)

	// It should fail
	require.Error(t, err)

	// Verify no double-spend
	require.Equal(t, 1, int(simulator.GetAccount[*LiteTokenAccount](sim, alice).Balance.Int64()))
	require.Equal(t, 1, int(simulator.GetAccount[*LiteTokenAccount](sim, bob).Balance.Int64()))
}

func TestSynthTxnToDn(t *testing.T) {
	// Tests AC-2231
	var timestamp uint64

	// Initialize
	sim := simulator.New(t, 3)
	sim.InitFromGenesis()

	// Setup accounts
	alice := AccountUrl("alice")
	aliceKey, liteKey := acctesting.GenerateKey(alice), acctesting.GenerateKey("lite")
	liteUrl := acctesting.AcmeLiteAddressStdPriv(liteKey)
	sim.CreateAccount(&LiteIdentity{Url: liteUrl.RootIdentity(), CreditBalance: 1e9})
	sim.CreateAccount(&LiteTokenAccount{Url: liteUrl, TokenUrl: AcmeUrl(), Balance: *big.NewInt(1e9 * AcmePrecision)})
	sim.SetRouteFor(alice, "Directory")
	sim.CreateIdentity(alice, aliceKey[32:])
	sim.CreateAccount(&TokenAccount{Url: alice.JoinPath("tokens"), TokenUrl: AcmeUrl()})

	// Send some tokens
	sim.WaitForTransactions(delivered, sim.MustSubmitAndExecuteBlock(
		acctesting.NewTransaction().
			WithPrincipal(liteUrl).
			WithSigner(liteUrl, 1).
			WithTimestampVar(&timestamp).
			WithBody(&SendTokens{To: []*TokenRecipient{{
				Url:    alice,
				Amount: *big.NewInt(1),
			}}}).
			Initiate(SignatureTypeED25519, liteKey).
			Build(),
	)...)
}
