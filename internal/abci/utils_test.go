package abci_test

import (
	"crypto/ed25519"

	tmed25519 "github.com/tendermint/tendermint/crypto/ed25519"
	acctesting "gitlab.com/accumulatenetwork/accumulate/internal/testing"
	"gitlab.com/accumulatenetwork/accumulate/internal/url"
)

var globalNonce uint64

func generateKey() tmed25519.PrivKey {
	_, key, _ := ed25519.GenerateKey(rand)
	return tmed25519.PrivKey(key)
}

func newTxn(origin string) acctesting.TransactionBuilder {
	return acctesting.NewTransaction().
		WithPrincipalStr(origin).
		WithSigner(url.MustParse(origin), 1).
		WithNonceVar(&globalNonce)
}
