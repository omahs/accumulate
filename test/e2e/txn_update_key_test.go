// Copyright 2023 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package e2e

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/accumulatenetwork/accumulate/pkg/url"
	. "gitlab.com/accumulatenetwork/accumulate/protocol"
	. "gitlab.com/accumulatenetwork/accumulate/test/harness"
	simulator "gitlab.com/accumulatenetwork/accumulate/test/simulator/compat"
	acctesting "gitlab.com/accumulatenetwork/accumulate/test/testing"
)

func TestUpdateKey(t *testing.T) {
	alice := url.MustParse("alice")
	aliceKey := acctesting.GenerateKey(alice, "main")
	otherKey := acctesting.GenerateKey(alice, "other")

	// Initialize
	var timestamp uint64
	sim := simulator.New(t, 3)
	sim.InitFromGenesis()

	sim.CreateIdentity(alice, aliceKey[32:])
	updateAccount(sim, alice.JoinPath("book", "1"), func(page *KeyPage) {
		page.CreditBalance = 1e9
	})

	// Update the key
	sim.WaitForTransactions(delivered, sim.MustSubmitAndExecuteBlock(
		acctesting.NewTransaction().
			WithPrincipal(alice.JoinPath("book", "1")).
			WithSigner(alice.JoinPath("book", "1"), 1).
			WithTimestampVar(&timestamp).
			UseSimpleHash(). // Test AC-2953
			WithBody(&UpdateKey{NewKeyHash: hash(otherKey[32:])}).
			Initiate(SignatureTypeED25519, aliceKey).
			Build(),
	)...)

	// Verify the key changed
	page := simulator.GetAccount[*KeyPage](sim, alice.JoinPath("book", "1"))
	require.Len(t, page.Keys, 1)
	require.Equal(t, hash(otherKey[32:]), page.Keys[0].PublicKeyHash)
}

func TestUpdateKey_HasDelegate(t *testing.T) {
	alice := url.MustParse("alice")
	aliceKey := acctesting.GenerateKey(alice, "main")
	otherKey := acctesting.GenerateKey(alice, "other")

	// Initialize
	var timestamp uint64
	sim := simulator.New(t, 3)
	sim.InitFromGenesis()

	sim.CreateIdentity(alice, aliceKey[32:])
	updateAccount(sim, alice.JoinPath("book", "1"), func(page *KeyPage) {
		page.CreditBalance = 1e9
		page.Keys[0].Delegate = AccountUrl("foo")
	})

	// Update the key
	sim.WaitForTransactions(delivered, sim.MustSubmitAndExecuteBlock(
		acctesting.NewTransaction().
			WithPrincipal(alice.JoinPath("book", "1")).
			WithSigner(alice.JoinPath("book", "1"), 1).
			WithTimestampVar(&timestamp).
			WithBody(&UpdateKey{NewKeyHash: hash(otherKey[32:])}).
			Initiate(SignatureTypeED25519, aliceKey).
			Build(),
	)...)

	// Verify the delegate is unchanged
	page := simulator.GetAccount[*KeyPage](sim, alice.JoinPath("book", "1"))
	require.Len(t, page.Keys, 1)
	require.NotNil(t, page.Keys[0].Delegate)
	require.Equal(t, "foo.acme", page.Keys[0].Delegate.ShortString())
}

func TestUpdateKey_MultiLevel(t *testing.T) {
	alice := url.MustParse("alice")
	aliceKey := acctesting.GenerateKey(alice, "main")
	otherKey := acctesting.GenerateKey(alice, "other")
	newKey := acctesting.GenerateKey(alice, "new")

	// Initialize
	var timestamp uint64
	sim := simulator.New(t, 3)
	sim.InitFromGenesis()

	sim.CreateIdentity(alice, aliceKey[32:])
	sim.CreateKeyBook(alice.JoinPath("book2"), make([]byte, 32))
	sim.CreateKeyBook(alice.JoinPath("book3"), otherKey[32:])
	updateAccount(sim, alice.JoinPath("book", "1"), func(page *KeyPage) { page.Keys[0].Delegate = alice.JoinPath("book2") })
	updateAccount(sim, alice.JoinPath("book2", "1"), func(page *KeyPage) { page.Keys[0].Delegate = alice.JoinPath("book3") })
	updateAccount(sim, alice.JoinPath("book3", "1"), func(page *KeyPage) { page.CreditBalance = 1e9 })

	// Update the key
	st := sim.H.SubmitSuccessfully(
		acctesting.NewTransaction().
			WithPrincipal(alice.JoinPath("book", "1")).
			WithSigner(alice.JoinPath("book3", "1"), 1).
			WithDelegator(alice.JoinPath("book2", "1")).
			WithDelegator(alice.JoinPath("book", "1")).
			WithTimestampVar(&timestamp).
			WithBody(&UpdateKey{NewKeyHash: hash(newKey[32:])}).
			Initiate(SignatureTypeED25519, otherKey).
			Build())
	sim.H.StepUntil(Txn(st.TxID).Fails())
	st = sim.H.QueryTransaction(st.TxID, nil).Status
	require.NotNil(t, st.Error)
	require.EqualError(t, st.Error, "cannot UpdateKey with a multi-level delegated signature")
}
