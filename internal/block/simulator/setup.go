package simulator

import (
	"crypto/sha256"

	"github.com/stretchr/testify/require"
	"gitlab.com/accumulatenetwork/accumulate/internal/chain"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

func writeAccountState(t TB, batch *database.Batch, account protocol.Account) {
	record := batch.Account(account.GetUrl())
	require.NoError(tb{t}, record.PutState(account))

	txid := sha256.Sum256([]byte("fake txid"))
	mainChain, err := record.Chain(protocol.MainChain, protocol.ChainTypeTransaction)
	require.NoError(tb{t}, err)
	require.NoError(tb{t}, mainChain.AddEntry(txid[:], true))

	identity, ok := account.GetUrl().Parent()
	if ok {
		require.NoError(tb{t}, chain.AddDirectoryEntry(func(account *url.URL, key ...interface{}) chain.Value {
			return batch.Account(account).Index(key...)
		}, identity, account.GetUrl()))
	}
}

func (s *Simulator) CreateAccount(account protocol.Account) {
	_ = s.SubnetFor(account.GetUrl()).Database.Update(func(batch *database.Batch) error {
		full, ok := account.(protocol.FullAccount)
		if !ok {
			writeAccountState(s, batch, account)
			return nil
		}

		auth := full.GetAuth()
		if len(auth.Authorities) > 0 {
			writeAccountState(s, batch, account)
			return nil
		}

		identityUrl, ok := account.GetUrl().Parent()
		require.True(tb{s}, ok, "Attempted to create an account with no auth")

		var identity *protocol.ADI
		require.NoError(tb{s}, batch.Account(identityUrl).GetStateAs(&identity))

		*auth = identity.AccountAuth
		writeAccountState(s, batch, account)
		return nil
	})
}

func (s *Simulator) CreateIdentity(identityUrl *url.URL, pubKey ...[]byte) {
	_ = s.SubnetFor(identityUrl).Database.Update(func(batch *database.Batch) error {
		identity := new(protocol.ADI)
		identity.Url = identityUrl
		identity.AddAuthority(identityUrl.JoinPath("book"))

		book := new(protocol.KeyBook)
		book.Url = identityUrl.JoinPath("book")
		book.AddAuthority(identityUrl.JoinPath("book"))
		book.PageCount = 1

		page := new(protocol.KeyPage)
		page.Url = protocol.FormatKeyPageUrl(identityUrl.JoinPath("book"), 0)
		page.AcceptThreshold = 1
		page.Version = 1

		for _, pubKey := range pubKey {
			keyHash := sha256.Sum256(pubKey)
			key := new(protocol.KeySpec)
			key.PublicKeyHash = keyHash[:]
			page.Keys = append(page.Keys, key)
		}

		writeAccountState(s, batch, identity)
		writeAccountState(s, batch, book)
		writeAccountState(s, batch, page)
		return nil
	})
}

func (s *Simulator) CreateKeyBook(bookUrl *url.URL, pubKey ...[]byte) {
	_ = s.SubnetFor(bookUrl).Database.Update(func(batch *database.Batch) error {
		book := new(protocol.KeyBook)
		book.Url = bookUrl
		book.AddAuthority(bookUrl)
		book.PageCount = 1

		page := new(protocol.KeyPage)
		page.Url = protocol.FormatKeyPageUrl(bookUrl, 0)
		page.AcceptThreshold = 1
		page.Version = 1

		for _, pubKey := range pubKey {
			keyHash := sha256.Sum256(pubKey)
			key := new(protocol.KeySpec)
			key.PublicKeyHash = keyHash[:]
			page.Keys = append(page.Keys, key)
		}

		writeAccountState(s, batch, book)
		writeAccountState(s, batch, page)
		return nil
	})
}

func (s *Simulator) CreateKeyPage(bookUrl *url.URL, pubKey ...[]byte) {
	_ = s.SubnetFor(bookUrl).Database.Update(func(batch *database.Batch) error {
		var book *protocol.KeyBook
		require.NoError(tb{s}, batch.Account(bookUrl).GetStateAs(&book))
		pageUrl := protocol.FormatKeyPageUrl(bookUrl, book.PageCount)
		book.PageCount++

		page := new(protocol.KeyPage)
		page.Url = pageUrl
		page.AcceptThreshold = 1
		page.Version = 1

		for _, pubKey := range pubKey {
			keyHash := sha256.Sum256(pubKey)
			key := new(protocol.KeySpec)
			key.PublicKeyHash = keyHash[:]
			page.Keys = append(page.Keys, key)
		}

		writeAccountState(s, batch, book)
		writeAccountState(s, batch, page)
		return nil
	})
}

func (s *Simulator) UpdateAccount(accountUrl *url.URL, fn func(account protocol.Account)) {
	_ = s.SubnetFor(accountUrl).Database.Update(func(batch *database.Batch) error {
		account, err := batch.Account(accountUrl).GetState()
		require.NoError(tb{s}, err)
		fn(account)
		writeAccountState(s, batch, account)
		return nil
	})
}
