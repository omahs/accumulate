// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package e2e

import (
	"context"
	"crypto/sha256"
	"math/big"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/accumulatenetwork/accumulate/config"
	"gitlab.com/accumulatenetwork/accumulate/internal/api/v2"
	"gitlab.com/accumulatenetwork/accumulate/internal/chain"
	"gitlab.com/accumulatenetwork/accumulate/internal/core"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/indexing"
	acctesting "gitlab.com/accumulatenetwork/accumulate/internal/testing"
	"gitlab.com/accumulatenetwork/accumulate/pkg/build"
	"gitlab.com/accumulatenetwork/accumulate/pkg/client/signing"
	"gitlab.com/accumulatenetwork/accumulate/pkg/url"
	. "gitlab.com/accumulatenetwork/accumulate/protocol"
	. "gitlab.com/accumulatenetwork/accumulate/test/helpers"
	"gitlab.com/accumulatenetwork/accumulate/test/simulator"
)

func TestBadOperatorPageUpdate(t *testing.T) {
	// Tests AC-3238

	// Initialize
	sim := NewSim(t,
		simulator.MemoryDatabase,
		simulator.SimpleNetwork(t.Name(), 1, 3),
		simulator.Genesis(GenesisTime),
	)

	before := GetAccount[*KeyPage](t, sim.Database(Directory), DnUrl().JoinPath(Operators, "1"))

	// Execute
	st := sim.SubmitSuccessfully(MustBuild(t,
		build.Transaction().For(DnUrl(), Operators, "1").
			UpdateKeyPage().Add().Entry().Hash([32]byte{1}).FinishEntry().FinishOperation().
			SignWith(DnUrl(), Operators, "1").Version(1).Timestamp(1).Signer(sim.SignWithNode(Directory, 0)).
			SignWith(DnUrl(), Operators, "1").Version(1).Timestamp(2).Signer(sim.SignWithNode(Directory, 1)).
			SignWith(DnUrl(), Operators, "1").Version(1).Timestamp(3).Signer(sim.SignWithNode(Directory, 2))))

	sim.StepUntil(
		Txn(st.TxID).Fails())

	// Verify the page did not change
	after := GetAccount[*KeyPage](t, sim.Database(Directory), DnUrl().JoinPath(Operators, "1"))
	require.Equal(t, before.AcceptThreshold, after.AcceptThreshold)
	require.Equal(t, len(before.Keys), len(after.Keys))
}

func TestBadOracleUpdate(t *testing.T) {
	// Tests AC-3238

	// Initialize
	sim := NewSim(t,
		simulator.MemoryDatabase,
		simulator.SimpleNetwork(t.Name(), 1, 3),
		simulator.Genesis(GenesisTime),
	)

	before := GetAccount[*DataAccount](t, sim.Database(Directory), DnUrl().JoinPath(Oracle))
	v := new(AcmeOracle)
	require.NoError(t, v.UnmarshalBinary(before.Entry.GetData()[0]))

	// Execute
	st := sim.SubmitSuccessfully(MustBuild(t,
		build.Transaction().For(DnUrl(), Oracle).
			WriteData([]byte("foo")).ToState().
			SignWith(DnUrl(), Operators, "1").Version(1).Timestamp(1).Signer(sim.SignWithNode(Directory, 0)).
			SignWith(DnUrl(), Operators, "1").Version(1).Timestamp(2).Signer(sim.SignWithNode(Directory, 1)).
			SignWith(DnUrl(), Operators, "1").Version(1).Timestamp(3).Signer(sim.SignWithNode(Directory, 2))))

	sim.StepUntil(
		Txn(st.TxID).Fails())

	// Verify the entry did not change
	after := GetAccount[*DataAccount](t, sim.Database(Directory), DnUrl().JoinPath(Oracle))
	v = new(AcmeOracle)
	require.NoError(t, v.UnmarshalBinary(after.Entry.GetData()[0]))
	require.True(t, before.Equal(after))
}

func TestDirectlyQueryReceiptSignature(t *testing.T) {
	// Tests AC-3254

	alice := url.MustParse("alice")
	bob := url.MustParse("bob")
	aliceKey := acctesting.GenerateKey(alice)
	bobKey := acctesting.GenerateKey(bob)

	// Initialize
	sim := NewSim(t,
		simulator.MemoryDatabase,
		simulator.SimpleNetwork(t.Name(), 3, 3),
		simulator.Genesis(GenesisTime),
	)

	MakeIdentity(t, sim.DatabaseFor(alice), alice, aliceKey[32:])
	CreditCredits(t, sim.DatabaseFor(alice), alice.JoinPath("book", "1"), 1e9)
	MakeAccount(t, sim.DatabaseFor(alice), &TokenAccount{Url: alice.JoinPath("tokens"), TokenUrl: AcmeUrl()})
	CreditTokens(t, sim.DatabaseFor(alice), alice.JoinPath("tokens"), big.NewInt(1e12))
	MakeIdentity(t, sim.DatabaseFor(bob), bob, bobKey[32:])
	MakeAccount(t, sim.DatabaseFor(bob), &TokenAccount{Url: bob.JoinPath("tokens"), TokenUrl: AcmeUrl()})

	// Execute
	st := sim.SubmitSuccessfully(MustBuild(t,
		build.Transaction().For(alice, "tokens").
			SendTokens(123, 0).To(bob, "tokens").
			SignWith(alice, "book", "1").Version(1).Timestamp(1).PrivateKey(aliceKey)))

	sim.StepUntil(
		Txn(st.TxID).Succeeds(),
		Txn(st.TxID).Produced().Succeeds())

	// Verify the receipt signature can be queried directly
	var synthHash *url.TxID
	View(t, sim.DatabaseFor(alice), func(batch *database.Batch) {
		h := st.TxID.Hash()
		p, err := batch.Transaction(h[:]).Produced().Get()
		require.NoError(t, err)
		require.Len(t, p, 1)
		synthHash = p[0]
	})

	var receiptHash *[32]byte
	View(t, sim.DatabaseFor(bob), func(batch *database.Batch) {
		h := synthHash.Hash()
		sigs, err := batch.Transaction(h[:]).ReadSignatures(DnUrl().JoinPath(Network))
		require.NoError(t, err)
		for _, entry := range sigs.Entries() {
			entry := entry
			s, err := batch.Transaction(entry.SignatureHash[:]).Main().Get()
			require.NoError(t, err)
			_, ok := s.Signature.(*ReceiptSignature)
			if ok {
				receiptHash = &entry.SignatureHash
			}
		}
		require.NotNil(t, receiptHash)
	})

	req := new(api.GeneralQuery)
	req.Url = bob.WithTxID(*receiptHash).AsUrl()
	resp := new(api.TransactionQueryResponse)
	part, err := sim.Router().RouteAccount(bob)
	require.NoError(t, err)
	err = sim.Router().RequestAPIv2(context.Background(), part, "query", req, resp)
	require.NoError(t, err)
}

func TestSignatureChainAnchoring(t *testing.T) {
	// Tests #3149

	alice := AccountUrl("alice")
	aliceKey := acctesting.GenerateKey(alice)

	// Start with executor version 0
	values := new(core.GlobalValues)

	// Initialize
	sim := NewSim(t,
		simulator.MemoryDatabase,
		simulator.SimpleNetwork(t.Name(), 1, 1),
		simulator.GenesisWith(GenesisTime, values),
	)

	MakeIdentity(t, sim.DatabaseFor(alice), alice, aliceKey[32:])
	CreditCredits(t, sim.DatabaseFor(alice), alice.JoinPath("book", "1"), 1e9)

	// Execute
	st := sim.SubmitSuccessfully(MustBuild(t,
		build.Transaction().For(alice).
			CreateDataAccount(alice, "foo").
			SignWith(alice, "book", "1").Version(1).Timestamp(1).PrivateKey(aliceKey)))

	sim.StepUntil(
		Txn(st.TxID).Succeeds())

	includesChain := func(block *BlockLedger, account *url.URL, name string) bool {
		for _, entry := range block.Entries {
			if entry.Account.Equal(account) && strings.EqualFold(name, entry.Chain) {
				return true
			}
		}
		return false
	}

	// Verify that the buggy behavior is retained
	alicePage := alice.JoinPath("book", "1")
	View(t, sim.DatabaseFor(alice), func(batch *database.Batch) {
		var ledger *SystemLedger
		require.NoError(t, batch.Account(PartitionUrl("BVN0").JoinPath(Ledger)).Main().GetAs(&ledger))

		var block *BlockLedger
		require.NoError(t, batch.Account(PartitionUrl("BVN0").JoinPath(Ledger, strconv.FormatUint(ledger.Index, 10))).Main().GetAs(&block))

		require.False(t, includesChain(block, alicePage, "signature"), "%v#chain/signature was anchored", alicePage)
		c, err := batch.Account(alicePage).SignatureChain().Index().Get()
		require.NoError(t, err)
		require.Zero(t, c.Height(), "%v#chain/signature was indexed", alicePage)
		_, _, err = indexing.ReceiptForChainIndex(&config.Describe{PartitionId: "BVN0"}, batch, batch.Account(alicePage).SignatureChain(), 0)
		require.EqualError(t, err, "cannot create receipt for entry 0 of signature chain: index chain is empty")
	})

	// Activate the new behavior
	st = sim.SubmitSuccessfully(MustBuild(t,
		build.Transaction().For(DnUrl()).
			ActivateProtocolVersion(ExecutorVersionV1SignatureAnchoring).
			SignWith(DnUrl(), Operators, "1").Version(1).Timestamp(1).Signer(sim.SignWithNode(Directory, 0))))

	sim.StepUntil(
		Txn(st.TxID).Succeeds())

	// Give the anchor a few blocks to propagate
	sim.StepN(10)

	// Execute
	st = sim.SubmitSuccessfully(MustBuild(t,
		build.Transaction().For(alice).
			CreateDataAccount(alice, "bar").
			SignWith(alice, "book", "1").Version(1).Timestamp(2).PrivateKey(aliceKey)))

	sim.StepUntil(
		Txn(st.TxID).Succeeds())

	// Verify the new behavior
	View(t, sim.DatabaseFor(alice), func(batch *database.Batch) {
		var ledger *SystemLedger
		require.NoError(t, batch.Account(PartitionUrl("BVN0").JoinPath(Ledger)).Main().GetAs(&ledger))

		var block *BlockLedger
		require.NoError(t, batch.Account(PartitionUrl("BVN0").JoinPath(Ledger, strconv.FormatUint(ledger.Index, 10))).Main().GetAs(&block))

		require.True(t, includesChain(block, alicePage, "signature"), "%v#chain/signature was not anchored", alicePage)
		c, err := batch.Account(alicePage).SignatureChain().Index().Get()
		require.NoError(t, err)
		require.NotZero(t, c.Height(), "%v#chain/signature was not indexed", alicePage)
		_, _, err = indexing.ReceiptForChainIndex(&config.Describe{PartitionId: "BVN0"}, batch, batch.Account(alicePage).SignatureChain(), 0)
		require.NoError(t, err)
	})
}

func TestUpdateKeyWithDelegate(t *testing.T) {
	alice := AccountUrl("alice")
	bob := AccountUrl("bob")
	aliceKey := acctesting.GenerateKey(alice)
	bobKey := acctesting.GenerateKey(bob)

	// Initialize
	sim := NewSim(t,
		simulator.MemoryDatabase,
		simulator.SimpleNetwork(t.Name(), 3, 3),
		simulator.Genesis(GenesisTime),
	)

	// On different BVNs
	sim.SetRoute(alice, "BVN0")
	sim.SetRoute(bob, "BVN1")

	MakeIdentity(t, sim.DatabaseFor(alice), alice, aliceKey[32:])
	CreditCredits(t, sim.DatabaseFor(alice), alice.JoinPath("book", "1"), 1e9)
	MakeIdentity(t, sim.DatabaseFor(bob), bob, bobKey[32:])
	CreditCredits(t, sim.DatabaseFor(bob), bob.JoinPath("book", "1"), 1e9)

	UpdateAccount(t, sim.DatabaseFor(bob), bob.JoinPath("book", "1"), func(p *KeyPage) { p.Keys = nil; p.AddKeySpec(&KeySpec{Delegate: alice.JoinPath("book")}) })

	st := sim.SubmitSuccessfully(MustBuild(t,
		build.Transaction().For(bob, "book", "1").
			UpdateKey(aliceKey, SignatureTypeED25519).
			SignWith(alice, "book", "1").Version(1).Timestamp(1).PrivateKey(aliceKey).Delegator(bob, "book", "1")))

	sim.StepUntil(
		Txn(st.TxID).Succeeds())

	hash := sha256.Sum256(aliceKey[32:])
	p := GetAccount[*KeyPage](t, sim.DatabaseFor(bob), bob.JoinPath("book", "1"))
	require.Len(t, p.Keys, 1)
	require.Equal(t, hash[:], p.Keys[0].PublicKeyHash)
	require.True(t, p.Keys[0].Delegate.Equal(alice.JoinPath("book")))
}

func TestRemoteAuthorityInitiator(t *testing.T) {
	alice := AccountUrl("alice")
	bob := AccountUrl("bob")
	charlie := AccountUrl("charlie")
	aliceKey1 := acctesting.GenerateKey(alice, 1)
	aliceKey2 := acctesting.GenerateKey(alice, 2)
	bobKey := acctesting.GenerateKey(bob)
	charlieKey := acctesting.GenerateKey(charlie)

	setup := func(t *testing.T, v ExecutorVersion) (*Sim, *chain.Delivery) {
		// Initialize with V1+sig
		sim := NewSim(t,
			simulator.MemoryDatabase,
			simulator.SimpleNetwork(t.Name(), 3, 1),
			simulator.GenesisWith(GenesisTime, &core.GlobalValues{ExecutorVersion: v}),
		)

		// The account (charlie) and authority (bob) are on one partition and the
		// delegate (alice) is on another
		sim.SetRoute(alice, "BVN0")
		sim.SetRoute(bob, "BVN1")
		sim.SetRoute(charlie, "BVN1")

		MakeIdentity(t, sim.DatabaseFor(alice), alice, aliceKey1[32:])
		UpdateAccount(t, sim.DatabaseFor(alice), alice.JoinPath("book", "1"), func(p *KeyPage) {
			p.CreditBalance = 1e9
			hash := sha256.Sum256(aliceKey2[32:])
			p.AddKeySpec(&KeySpec{PublicKeyHash: hash[:]})
		})
		MakeIdentity(t, sim.DatabaseFor(bob), bob, bobKey[32:])
		UpdateAccount(t, sim.DatabaseFor(bob), bob.JoinPath("book", "1"), func(p *KeyPage) {
			p.CreditBalance = 1e9
			p.AddKeySpec(&KeySpec{Delegate: alice.JoinPath("book")})
			p.AcceptThreshold = 2
		})
		MakeIdentity(t, sim.DatabaseFor(charlie), charlie, charlieKey[32:])
		CreditCredits(t, sim.DatabaseFor(charlie), charlie.JoinPath("book", "1"), 1e9)
		MakeAccount(t, sim.DatabaseFor(charlie), &TokenAccount{Url: charlie.JoinPath("tokens"), TokenUrl: AcmeUrl(), Balance: *big.NewInt(1), AccountAuth: AccountAuth{Authorities: []AuthorityEntry{{Url: bob.JoinPath("book")}}}})

		// Initiate but do not submit the transaction with alice
		delivery := MustBuild(t,
			build.Transaction().For(charlie, "tokens").
				SendTokens(1, 0).To("foo").
				SignWith(alice, "book", "1").Version(1).Timestamp(1).PrivateKey(aliceKey1).Delegator(bob, "book", "1"))

		return sim, delivery
	}

	outOfOrder := func(sim *Sim, delivery *chain.Delivery) *TransactionStatus {
		// Sign and submit the transaction with bob
		st := sim.SubmitSuccessfully(MustBuild(t,
			build.SignatureForTransaction(delivery.Transaction).
				Url(bob, "book", "1").Version(1).Timestamp(1).PrivateKey(bobKey)))

		sim.StepUntil(
			Txn(st.TxID).Received())

		// Submit alice's signature
		sim.SubmitSuccessfully(delivery)

		return st
	}

	extraSig := func(sim *Sim, delivery *chain.Delivery) *TransactionStatus {
		// Submit alice's signature
		st := sim.SubmitSuccessfully(delivery)

		sim.StepUntil(
			Txn(st.TxID).Received())

		// Submit with alice's other key
		sim.SubmitSuccessfully(MustBuild(t,
			build.SignatureForTransaction(delivery.Transaction).
				Url(alice, "book", "1").Version(1).Timestamp(1).PrivateKey(aliceKey2).Delegator(bob, "book", "1")))

		sim.StepN(50)

		// Sign and submit the transaction with bob
		sim.SubmitSuccessfully(MustBuild(t,
			build.SignatureForTransaction(delivery.Transaction).
				Url(bob, "book", "1").Version(1).Timestamp(1).PrivateKey(bobKey)))

		return st
	}

	captureFwd := func(sim *Sim) func() *url.TxID {
		var sigId *url.TxID
		sim.SetSubmitHook("BVN1", func(d *chain.Delivery) (dropTx bool, keepHook bool) {
			fwd, ok := d.Transaction.Body.(*SyntheticForwardTransaction)
			if !ok || len(fwd.Signatures) != 1 || !fwd.Signatures[0].Destination.Equal(charlie.JoinPath("tokens")) {
				return false, true
			}
			sig := fwd.Signatures[0]
			sigId = sig.Destination.WithTxID(*(*[32]byte)(sig.Signature.Hash()))
			return false, false
		})
		return func() *url.TxID {
			for sigId == nil {
				sim.Step()
			}
			return sigId
		}
	}

	fwdFails := func(sim *Sim, sigId *url.TxID, errstr string) {
		sim.StepUntil(
			Txn(sigId).Fails())
		st := sim.GetStatus(sigId)
		require.NotNil(sim.T, st.Error)
		require.EqualError(sim.T, st.Error, errstr)
	}

	// Broken in V1
	t.Run("V1", func(t *testing.T) {
		t.Run("Out of order", func(t *testing.T) {
			sim, delivery := setup(t, ExecutorVersionV1)
			waitForFwd := captureFwd(sim)

			outOfOrder(sim, delivery)

			// Fails
			fwdFails(sim, waitForFwd(), "initiator is already set and does not match the signature")
		})

		t.Run("Extra signature", func(t *testing.T) {
			sim, delivery := setup(t, ExecutorVersionV1)
			waitForFwd := captureFwd(sim)

			extraSig(sim, delivery)

			// Fails
			fwdFails(sim, waitForFwd(), "initiator is already set and does not match the signature")
		})
	})

	// Fixed in V1+sig
	t.Run("V1+sig", func(t *testing.T) {
		t.Run("Out of order", func(t *testing.T) {
			sim, delivery := setup(t, ExecutorVersionV1SignatureAnchoring)

			st := outOfOrder(sim, delivery)

			// Succeeds
			sim.StepUntil(
				Txn(st.TxID).Succeeds())
		})

		t.Run("Extra signature", func(t *testing.T) {
			sim, delivery := setup(t, ExecutorVersionV1SignatureAnchoring)

			st := extraSig(sim, delivery)

			// Succeeds
			sim.StepUntil(
				Txn(st.TxID).Succeeds())
		})
	})
}

func TestSignerOverwritten(t *testing.T) {
	alice := AccountUrl("alice")
	bob := AccountUrl("bob")
	aliceKey := acctesting.GenerateKey(alice)
	bobKey := acctesting.GenerateKey(bob)

	// Initialize with V1+sig
	sim := NewSim(t,
		simulator.MemoryDatabase,
		simulator.SimpleNetwork(t.Name(), 3, 1),
		simulator.GenesisWith(GenesisTime, &core.GlobalValues{ExecutorVersion: ExecutorVersionV1SignatureAnchoring}),
	)

	MakeIdentity(t, sim.DatabaseFor(alice), alice, aliceKey[32:])
	UpdateAccount(t, sim.DatabaseFor(alice), alice.JoinPath("book", "1"), func(p *KeyPage) {
		p.CreditBalance = 1e9
		p.AddKeySpec(&KeySpec{Delegate: alice.JoinPath("book")})
		p.AcceptThreshold = 2
	})
	MakeIdentity(t, sim.DatabaseFor(bob), bob, bobKey[32:])
	CreditCredits(t, sim.DatabaseFor(bob), bob.JoinPath("book", "1"), 1e9)
	MakeAccount(t, sim.DatabaseFor(bob), &TokenAccount{Url: bob.JoinPath("tokens"), TokenUrl: AcmeUrl(), Balance: *big.NewInt(1), AccountAuth: AccountAuth{Authorities: []AuthorityEntry{{Url: alice.JoinPath("book")}}}})

	st := sim.SubmitSuccessfully(MustBuild(t,
		build.Transaction().For(bob, "tokens").
			SendTokens(1, 0).To("foo").
			SignWith(alice, "book", "1").Version(1).Timestamp(1).PrivateKey(aliceKey)))

	h := st.TxID.Hash()
	st = new(TransactionStatus)
	for i := 0; st.Code == 0 && i < 50; i++ {
		sim.Step()

		View(t, sim.DatabaseFor(alice), func(batch *database.Batch) {
			var err error
			st, err = batch.Transaction(h[:]).Status().Get()
			require.NoError(t, err)
		})
	}
	if st.Code == 0 {
		t.Fatal("Condition not met after 50 blocks")
	}

	// Ensure Alice is added to the signers
	require.Len(t, st.Signers, 1)
	require.Equal(t, "alice.acme/book/1", st.Signers[0].GetUrl().ShortString())
}

func TestMissingPrincipal(t *testing.T) {
	liteKey := acctesting.GenerateKey()
	liteUrl := acctesting.AcmeLiteAddressStdPriv(liteKey)

	// Initialize with V1+sig
	sim := NewSim(t,
		simulator.MemoryDatabase,
		simulator.SimpleNetwork(t.Name(), 3, 1),
		simulator.Genesis(GenesisTime),
	)

	MakeLiteTokenAccount(t, sim.DatabaseFor(liteUrl), liteKey[32:], AcmeUrl())

	txn := new(Transaction)
	txn.Body = new(SendTokens)
	sig, err := new(signing.Builder).
		SetUrl(liteUrl).
		SetVersion(1).
		SetTimestamp(1).
		SetPrivateKey(liteKey).
		Initiate(txn)
	require.NoError(t, err)

	st := sim.Submit(&chain.Delivery{Transaction: txn, Signatures: []Signature{sig}})
	require.NotNil(t, st.Error)
	require.EqualError(t, st.Error, "missing principal")
}
