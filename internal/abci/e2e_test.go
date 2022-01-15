package abci_test

import (
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	acctesting "github.com/AccumulateNetwork/accumulate/internal/testing"
	"github.com/AccumulateNetwork/accumulate/internal/testing/e2e"
	"github.com/AccumulateNetwork/accumulate/internal/url"
	"github.com/AccumulateNetwork/accumulate/protocol"
	"github.com/AccumulateNetwork/accumulate/types"
	"github.com/AccumulateNetwork/accumulate/types/api/transactions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	randpkg "golang.org/x/exp/rand"
)

var rand = randpkg.New(randpkg.NewSource(0))

type Tx = transactions.Envelope

func TestEndToEndSuite(t *testing.T) {
	acctesting.SkipCI(t, "flaky")

	suite.Run(t, e2e.NewSuite(func(s *e2e.Suite) e2e.DUT {
		// Recreate the app for each test
		subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
		nodes := RunTestNet(t, subnets, daemons, nil, true)
		n := nodes[subnets[1]][0]

		return &e2eDUT{s, n}
	}))
}

func TestCreateLiteAccount(t *testing.T) {
	subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
	nodes := RunTestNet(t, subnets, daemons, nil, true)
	n := nodes[subnets[1]][0]

	var count = 11
	originAddr, balances := n.testLiteTx(count)
	require.Equal(t, int64(5e4*acctesting.TokenMx-count*1000), n.GetLiteTokenAccount(originAddr).Balance.Int64())
	for addr, bal := range balances {
		require.Equal(t, bal, n.GetLiteTokenAccount(addr).Balance.Int64())
	}
}

func (n *FakeNode) testLiteTx(count int) (string, map[string]int64) {
	_, recipient, gtx, err := acctesting.BuildTestSynthDepositGenTx()
	require.NoError(n.t, err)

	origin := acctesting.NewWalletEntry()
	origin.Nonce = 1
	origin.PrivateKey = recipient
	origin.Addr = acctesting.AcmeLiteAddressStdPriv(recipient).String()

	recipients := make([]*transactions.WalletEntry, 10)
	for i := range recipients {
		recipients[i] = acctesting.NewWalletEntry()
	}

	n.Batch(func(send func(*transactions.Envelope)) {
		send(gtx)
	})

	balance := map[string]int64{}
	n.Batch(func(send func(*Tx)) {
		for i := 0; i < count; i++ {
			recipient := recipients[rand.Intn(len(recipients))]
			balance[recipient.Addr] += 1000

			exch := new(protocol.SendTokens)
			exch.AddRecipient(n.ParseUrl(recipient.Addr), 1000)
			tx, err := transactions.New(origin.Addr, 1, func(hash []byte) (*transactions.ED25519Sig, error) {
				return origin.Sign(hash), nil
			}, exch)
			require.NoError(n.t, err)
			send(tx)
		}
	})

	return origin.Addr, balance
}

func TestFaucet(t *testing.T) {
	subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
	nodes := RunTestNet(t, subnets, daemons, nil, true)
	n := nodes[subnets[1]][0]

	alice := generateKey()
	aliceUrl := acctesting.AcmeLiteAddressTmPriv(alice).String()

	n.Batch(func(send func(*transactions.Envelope)) {
		body := new(protocol.AcmeFaucet)
		body.Url = aliceUrl
		tx, err := transactions.New(protocol.FaucetUrl.String(), 1, func(hash []byte) (*transactions.ED25519Sig, error) {
			return protocol.FaucetWallet.Sign(hash), nil
		}, body)
		require.NoError(t, err)
		send(tx)
	})

	require.Equal(t, int64(10*protocol.AcmePrecision), n.GetLiteTokenAccount(aliceUrl).Balance.Int64())
}

func TestAnchorChain(t *testing.T) {
	subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
	nodes := RunTestNet(t, subnets, daemons, nil, true)
	n := nodes[subnets[1]][0]

	liteAccount := generateKey()
	batch := n.db.Begin()
	require.NoError(n.t, acctesting.CreateLiteTokenAccount(batch, liteAccount, 5e4))
	require.NoError(t, batch.Commit())

	n.Batch(func(send func(*Tx)) {
		adi := new(protocol.CreateIdentity)
		adi.Url = "RoadRunner"
		adi.KeyBookName = "book"
		adi.KeyPageName = "page"

		sponsorUrl := acctesting.AcmeLiteAddressTmPriv(liteAccount).String()
		tx, err := transactions.New(sponsorUrl, 1, edSigner(liteAccount, 1), adi)
		require.NoError(t, err)

		send(tx)
	})

	// Sanity check
	require.Equal(t, types.String("acc://RoadRunner"), n.GetADI("RoadRunner").ChainUrl)

	// Get the anchor chain manager
	batch = n.db.Begin()
	defer batch.Discard()
	ledger := batch.Account(n.network.NodeUrl(protocol.Ledger))

	// Check each anchor
	ledgerState := protocol.NewInternalLedger()
	require.NoError(t, ledger.GetStateAs(ledgerState))
	rootChain, err := ledger.ReadChain(protocol.MinorRootChain)
	require.NoError(t, err)
	first := rootChain.Height() - int64(len(ledgerState.Updates))
	var accounts []string
	for i, meta := range ledgerState.Updates {
		accounts = append(accounts, fmt.Sprintf("%s#chain/%s", meta.Account, meta.Name))

		root, err := rootChain.Entry(first + int64(i))
		require.NoError(t, err)

		if meta.Name == "bpt" {
			assert.Equal(t, root, batch.RootHash(), "wrong anchor for BPT")
			continue
		}

		mgr, err := batch.Account(meta.Account).ReadChain(meta.Name)
		require.NoError(t, err)

		assert.Equal(t, root, mgr.Anchor(), "wrong anchor for %s#chain/%s", meta.Account, meta.Name)
	}

	// Verify that the ADI accounts are included
	assert.Subset(t, accounts, []string{
		"acc://RoadRunner#chain/main",
		"acc://RoadRunner#chain/pending",
		"acc://RoadRunner/book#chain/main",
		"acc://RoadRunner/page#chain/main",
	})
}

func TestCreateADI(t *testing.T) {
	subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
	nodes := RunTestNet(t, subnets, daemons, nil, true)
	n := nodes[subnets[1]][0]

	liteAccount := generateKey()
	newAdi := generateKey()
	keyHash := sha256.Sum256(newAdi.PubKey().Address())
	batch := n.db.Begin()
	require.NoError(n.t, acctesting.CreateLiteTokenAccount(batch, liteAccount, 5e4))
	require.NoError(t, batch.Commit())

	wallet := new(transactions.WalletEntry)
	wallet.Nonce = 1
	wallet.PrivateKey = liteAccount.Bytes()
	wallet.Addr = acctesting.AcmeLiteAddressTmPriv(liteAccount).String()

	n.Batch(func(send func(*Tx)) {
		adi := new(protocol.CreateIdentity)
		adi.Url = "RoadRunner"
		adi.PublicKey = keyHash[:]
		adi.KeyBookName = "foo-book"
		adi.KeyPageName = "bar-page"

		sponsorUrl := acctesting.AcmeLiteAddressTmPriv(liteAccount).String()
		tx, err := transactions.New(sponsorUrl, 1, func(hash []byte) (*transactions.ED25519Sig, error) {
			return wallet.Sign(hash), nil
		}, adi)
		require.NoError(t, err)

		send(tx)
	})

	r := n.GetADI("RoadRunner")
	require.Equal(t, types.String("acc://RoadRunner"), r.ChainUrl)

	kg := n.GetKeyBook("RoadRunner/foo-book")
	require.Len(t, kg.Pages, 1)

	ks := n.GetKeyPage("RoadRunner/bar-page")
	require.Len(t, ks.Keys, 1)
	require.Equal(t, keyHash[:], ks.Keys[0].PublicKey)
}

func TestCreateLiteDataAccount(t *testing.T) {

	//this test exercises WriteDataTo and SyntheticWriteData validators

	firstEntry := protocol.DataEntry{}

	firstEntry.ExtIds = append(firstEntry.ExtIds, []byte("Factom PRO"))
	firstEntry.ExtIds = append(firstEntry.ExtIds, []byte("Tutorial"))

	//create a lite data account aka factom chainId
	chainId := protocol.ComputeLiteDataAccountId(&firstEntry)

	lde := protocol.LiteDataEntry{}
	lde.DataEntry = new(protocol.DataEntry)
	copy(lde.AccountId[:], chainId)
	lde.Data = []byte("This is useful content of the entry. You can save text, hash, JSON or raw ASCII data here.")
	for i := 0; i < 3; i++ {
		lde.ExtIds = append(lde.ExtIds, []byte(fmt.Sprintf("Tag #%d of entry", i+1)))
	}
	liteDataAddress, err := protocol.LiteDataAddress(chainId)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("Create ADI then write to Lite Data Account", func(t *testing.T) {
		subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
		nodes := RunTestNet(t, subnets, daemons, nil, true)
		n := nodes[subnets[1]][0]

		adiKey := generateKey()
		batch := n.db.Begin()
		require.NoError(t, acctesting.CreateADI(batch, adiKey, "FooBar"))
		require.NoError(t, batch.Commit())
		n.Batch(func(send func(*transactions.Envelope)) {
			wdt := new(protocol.WriteDataTo)
			wdt.Recipient = liteDataAddress.String()
			wdt.Entry = firstEntry
			tx, err := transactions.New("FooBar", 1, edSigner(adiKey, 1), wdt)
			require.NoError(t, err)
			send(tx)
		})

		partialChainId, err := protocol.ParseLiteDataAddress(liteDataAddress)
		if err != nil {
			t.Fatal(err)
		}
		r := n.GetLiteDataAccount(liteDataAddress.String())
		require.Equal(t, types.AccountTypeLiteDataAccount, r.Type)
		require.Equal(t, types.String(liteDataAddress.String()), r.ChainUrl)
		require.Equal(t, append(partialChainId, r.Tail...), chainId)
	})
}

func TestCreateAdiDataAccount(t *testing.T) {

	t.Run("Data Account w/ Default Key Book and no Manager Key Book", func(t *testing.T) {
		subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
		nodes := RunTestNet(t, subnets, daemons, nil, true)
		n := nodes[subnets[1]][0]

		adiKey := generateKey()
		batch := n.db.Begin()
		require.NoError(t, acctesting.CreateADI(batch, adiKey, "FooBar"))
		require.NoError(t, batch.Commit())

		n.Batch(func(send func(*transactions.Envelope)) {
			tac := new(protocol.CreateDataAccount)
			tac.Url = "FooBar/oof"
			tx, err := transactions.New("FooBar", 1, edSigner(adiKey, 1), tac)
			require.NoError(t, err)
			send(tx)
		})

		r := n.GetDataAccount("FooBar/oof")
		require.Equal(t, types.AccountTypeDataAccount, r.Type)
		require.Equal(t, types.String("acc://FooBar/oof"), r.ChainUrl)

		require.Contains(t, n.GetDirectory("FooBar"), n.ParseUrl("FooBar/oof").String())
	})

	t.Run("Data Account w/ Custom Key Book and Manager Key Book Url", func(t *testing.T) {
		subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
		nodes := RunTestNet(t, subnets, daemons, nil, true)
		n := nodes[subnets[1]][0]

		adiKey, pageKey := generateKey(), generateKey()
		batch := n.db.Begin()
		require.NoError(t, acctesting.CreateADI(batch, adiKey, "FooBar"))
		require.NoError(t, acctesting.CreateKeyPage(batch, "acc://FooBar/foo/page1", pageKey.PubKey().Bytes()))
		require.NoError(t, acctesting.CreateKeyBook(batch, "acc://FooBar/foo/book1", "acc://FooBar/foo/page1"))
		require.NoError(t, acctesting.CreateKeyPage(batch, "acc://FooBar/mgr/page1", pageKey.PubKey().Bytes()))
		require.NoError(t, acctesting.CreateKeyBook(batch, "acc://FooBar/mgr/book1", "acc://FooBar/mgr/page1"))
		require.NoError(t, batch.Commit())

		n.Batch(func(send func(*transactions.Envelope)) {
			cda := new(protocol.CreateDataAccount)
			cda.Url = "FooBar/oof"
			cda.KeyBookUrl = "acc://FooBar/foo/book1"
			cda.ManagerKeyBookUrl = "acc://FooBar/mgr/book1"
			tx, err := transactions.New("FooBar", 1, edSigner(adiKey, 1), cda)
			require.NoError(t, err)
			send(tx)
		})

		u := n.ParseUrl("acc://FooBar/foo/book1")

		r := n.GetDataAccount("FooBar/oof")
		require.Equal(t, types.AccountTypeDataAccount, r.Type)
		require.Equal(t, types.String("acc://FooBar/oof"), r.ChainUrl)
		require.Equal(t, types.String("acc://FooBar/mgr/book1"), r.ManagerKeyBook)
		require.Equal(t, types.String(u.String()), r.KeyBook)

	})

	t.Run("Data Account data entry", func(t *testing.T) {
		subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
		nodes := RunTestNet(t, subnets, daemons, nil, true)
		n := nodes[subnets[1]][0]

		adiKey := generateKey()
		batch := n.db.Begin()
		require.NoError(t, acctesting.CreateADI(batch, adiKey, "FooBar"))
		require.NoError(t, batch.Commit())

		n.Batch(func(send func(*transactions.Envelope)) {
			tac := new(protocol.CreateDataAccount)
			tac.Url = "FooBar/oof"
			tx, err := transactions.New("FooBar", 1, edSigner(adiKey, 1), tac)
			require.NoError(t, err)
			send(tx)
		})

		r := n.GetDataAccount("FooBar/oof")
		require.Equal(t, types.AccountTypeDataAccount, r.Type)
		require.Equal(t, types.String("acc://FooBar/oof"), r.ChainUrl)
		require.Contains(t, n.GetDirectory("FooBar"), n.ParseUrl("FooBar/oof").String())

		wd := new(protocol.WriteData)
		n.Batch(func(send func(*transactions.Envelope)) {
			for i := 0; i < 10; i++ {
				wd.Entry.ExtIds = append(wd.Entry.ExtIds, []byte(fmt.Sprintf("test id %d", i)))
			}

			wd.Entry.Data = []byte("thequickbrownfoxjumpsoverthelazydog")

			tx, err := transactions.New("FooBar/oof", 1, edSigner(adiKey, 2), wd)
			require.NoError(t, err)
			send(tx)
		})

		// Without the sleep, this test fails on Windows and macOS
		time.Sleep(3 * time.Second)

		// Test getting the data by URL
		r2 := n.GetChainDataByUrl("FooBar/oof")
		if r2 == nil {
			t.Fatalf("error getting chain data by URL")
		}

		if r2.Data == nil {
			t.Fatalf("no data returned")
		}

		rde := protocol.ResponseDataEntry{}

		err := rde.UnmarshalJSON(*r2.Data)
		if err != nil {
			t.Fatal(err)
		}

		if !rde.Entry.Equal(&wd.Entry) {
			t.Fatalf("data query does not match what was entered")
		}

		//now test query by entry hash.
		r3 := n.GetChainDataByEntryHash("FooBar/oof", wd.Entry.Hash())

		if r3.Data == nil {
			t.Fatalf("no data returned")
		}

		rde2 := protocol.ResponseDataEntry{}

		err = rde2.UnmarshalJSON(*r3.Data)
		if err != nil {
			t.Fatal(err)
		}

		if !rde.Entry.Equal(&rde2.Entry) {
			t.Fatalf("data query does not match what was entered")
		}

		//now test query by entry set
		r4 := n.GetChainDataSet("FooBar/oof", 0, 1, true)

		if r4.Data == nil {
			t.Fatalf("no data returned")
		}

		if len(r4.Data) != 1 {
			t.Fatalf("insufficent data return from set query")
		}
		rde3 := protocol.ResponseDataEntry{}
		err = rde3.UnmarshalJSON(*r4.Data[0].Data)
		if err != nil {
			t.Fatal(err)
		}

		if !rde.Entry.Equal(&rde3.Entry) {
			t.Fatalf("data query does not match what was entered")
		}

	})
}

func TestCreateAdiTokenAccount(t *testing.T) {
	t.Run("Default Key Book", func(t *testing.T) {
		subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
		nodes := RunTestNet(t, subnets, daemons, nil, true)
		n := nodes[subnets[1]][0]

		adiKey := generateKey()
		batch := n.db.Begin()
		require.NoError(t, acctesting.CreateADI(batch, adiKey, "FooBar"))
		require.NoError(t, batch.Commit())

		n.Batch(func(send func(*transactions.Envelope)) {
			tac := new(protocol.CreateTokenAccount)
			tac.Url = "FooBar/Baz"
			tac.TokenUrl = protocol.AcmeUrl().String()
			tx, err := transactions.New("FooBar", 1, edSigner(adiKey, 1), tac)
			require.NoError(t, err)
			send(tx)
		})

		r := n.GetTokenAccount("FooBar/Baz")
		require.Equal(t, types.AccountTypeTokenAccount, r.Type)
		require.Equal(t, types.String("acc://FooBar/Baz"), r.ChainUrl)
		require.Equal(t, protocol.AcmeUrl().String(), r.TokenUrl)

		require.Equal(t, []string{
			n.ParseUrl("FooBar").String(),
			n.ParseUrl("FooBar/book0").String(),
			n.ParseUrl("FooBar/page0").String(),
			n.ParseUrl("FooBar/Baz").String(),
		}, n.GetDirectory("FooBar"))
	})

	t.Run("Custom Key Book", func(t *testing.T) {
		subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
		nodes := RunTestNet(t, subnets, daemons, nil, true)
		n := nodes[subnets[1]][0]

		adiKey, pageKey := generateKey(), generateKey()
		batch := n.db.Begin()
		require.NoError(t, acctesting.CreateADI(batch, adiKey, "FooBar"))
		require.NoError(t, acctesting.CreateKeyPage(batch, "foo/page1", pageKey.PubKey().Bytes()))
		require.NoError(t, acctesting.CreateKeyBook(batch, "foo/book1", "foo/page1"))
		require.NoError(t, batch.Commit())

		n.Batch(func(send func(*transactions.Envelope)) {
			tac := new(protocol.CreateTokenAccount)
			tac.Url = "FooBar/Baz"
			tac.TokenUrl = protocol.AcmeUrl().String()
			tac.KeyBookUrl = "foo/book1"
			tx, err := transactions.New("FooBar", 1, edSigner(adiKey, 1), tac)
			require.NoError(t, err)
			send(tx)
		})

		u := n.ParseUrl("foo/book1")

		r := n.GetTokenAccount("FooBar/Baz")
		require.Equal(t, types.AccountTypeTokenAccount, r.Type)
		require.Equal(t, types.String("acc://FooBar/Baz"), r.ChainUrl)
		require.Equal(t, protocol.AcmeUrl().String(), r.TokenUrl)
		require.Equal(t, types.String(u.String()), r.KeyBook)
	})
}

func TestLiteAccountTx(t *testing.T) {
	subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
	nodes := RunTestNet(t, subnets, daemons, nil, true)
	n := nodes[subnets[1]][0]

	alice, bob, charlie := generateKey(), generateKey(), generateKey()
	batch := n.db.Begin()
	require.NoError(n.t, acctesting.CreateLiteTokenAccount(batch, alice, 5e4))
	require.NoError(n.t, acctesting.CreateLiteTokenAccount(batch, bob, 0))
	require.NoError(n.t, acctesting.CreateLiteTokenAccount(batch, charlie, 0))
	require.NoError(t, batch.Commit())

	aliceUrl := acctesting.AcmeLiteAddressTmPriv(alice).String()
	bobUrl := acctesting.AcmeLiteAddressTmPriv(bob).String()
	charlieUrl := acctesting.AcmeLiteAddressTmPriv(charlie).String()

	n.Batch(func(send func(*transactions.Envelope)) {
		exch := new(protocol.SendTokens)
		exch.AddRecipient(acctesting.MustParseUrl(bobUrl), 1000)
		exch.AddRecipient(acctesting.MustParseUrl(charlieUrl), 2000)

		tx, err := transactions.New(aliceUrl, 2, edSigner(alice, 1), exch)
		require.NoError(t, err)
		send(tx)
	})

	require.Equal(t, int64(5e4*acctesting.TokenMx-3000), n.GetLiteTokenAccount(aliceUrl).Balance.Int64())
	require.Equal(t, int64(1000), n.GetLiteTokenAccount(bobUrl).Balance.Int64())
	require.Equal(t, int64(2000), n.GetLiteTokenAccount(charlieUrl).Balance.Int64())
}

func TestAdiAccountTx(t *testing.T) {
	subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
	nodes := RunTestNet(t, subnets, daemons, nil, true)
	n := nodes[subnets[1]][0]

	fooKey, barKey := generateKey(), generateKey()
	batch := n.db.Begin()
	require.NoError(t, acctesting.CreateADI(batch, fooKey, "foo"))
	require.NoError(t, acctesting.CreateTokenAccount(batch, "foo/tokens", protocol.AcmeUrl().String(), 1, false))
	require.NoError(t, acctesting.CreateADI(batch, barKey, "bar"))
	require.NoError(t, acctesting.CreateTokenAccount(batch, "bar/tokens", protocol.AcmeUrl().String(), 0, false))
	require.NoError(t, batch.Commit())

	n.Batch(func(send func(*transactions.Envelope)) {
		exch := new(protocol.SendTokens)
		exch.AddRecipient(n.ParseUrl("bar/tokens"), 68)

		tx, err := transactions.New("foo/tokens", 1, edSigner(fooKey, 1), exch)
		require.NoError(t, err)
		send(tx)
	})

	require.Equal(t, int64(acctesting.TokenMx-68), n.GetTokenAccount("foo/tokens").Balance.Int64())
	require.Equal(t, int64(68), n.GetTokenAccount("bar/tokens").Balance.Int64())
}

func TestSendCreditsFromAdiAccountToMultiSig(t *testing.T) {
	subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
	nodes := RunTestNet(t, subnets, daemons, nil, true)
	n := nodes[subnets[1]][0]

	fooKey := generateKey()
	batch := n.db.Begin()
	require.NoError(t, acctesting.CreateADI(batch, fooKey, "foo"))
	require.NoError(t, acctesting.CreateTokenAccount(batch, "foo/tokens", protocol.AcmeUrl().String(), 1e2, false))
	require.NoError(t, batch.Commit())

	n.Batch(func(send func(*transactions.Envelope)) {
		ac := new(protocol.AddCredits)
		ac.Amount = 55
		ac.Recipient = "foo/page0"

		tx, err := transactions.New("foo/tokens", 1, edSigner(fooKey, 1), ac)
		require.NoError(t, err)
		send(tx)
	})

	ks := n.GetKeyPage("foo/page0")
	acct := n.GetTokenAccount("foo/tokens")
	require.Equal(t, int64(55), ks.CreditBalance.Int64())
	require.Equal(t, int64(protocol.AcmePrecision*1e2-protocol.AcmePrecision/protocol.CreditsPerFiatUnit*55), acct.Balance.Int64())
}

func TestCreateKeyPage(t *testing.T) {
	subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
	nodes := RunTestNet(t, subnets, daemons, nil, true)
	n := nodes[subnets[1]][0]

	fooKey, testKey := generateKey(), generateKey()
	batch := n.db.Begin()
	require.NoError(t, acctesting.CreateADI(batch, fooKey, "foo"))
	require.NoError(t, batch.Commit())

	n.Batch(func(send func(*transactions.Envelope)) {
		cms := new(protocol.CreateKeyPage)
		cms.Url = "foo/keyset1"
		cms.Keys = append(cms.Keys, &protocol.KeySpecParams{
			PublicKey: testKey.PubKey().Bytes(),
		})

		tx, err := transactions.New("foo", 1, edSigner(fooKey, 1), cms)
		require.NoError(t, err)
		send(tx)
	})

	spec := n.GetKeyPage("foo/keyset1")
	require.Len(t, spec.Keys, 1)
	key := spec.Keys[0]
	require.Equal(t, types.String(""), spec.KeyBook)
	require.Equal(t, uint64(0), key.Nonce)
	require.Equal(t, testKey.PubKey().Bytes(), key.PublicKey)
}

func TestCreateKeyBook(t *testing.T) {
	subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
	nodes := RunTestNet(t, subnets, daemons, nil, true)
	n := nodes[subnets[1]][0]

	fooKey, testKey := generateKey(), generateKey()
	batch := n.db.Begin()
	require.NoError(t, acctesting.CreateADI(batch, fooKey, "foo"))
	require.NoError(t, acctesting.CreateKeyPage(batch, "foo/page1", testKey.PubKey().Bytes()))
	require.NoError(t, batch.Commit())

	specUrl := n.ParseUrl("foo/page1")

	groupUrl := n.ParseUrl("foo/book1")

	n.Batch(func(send func(*transactions.Envelope)) {
		csg := new(protocol.CreateKeyBook)
		csg.Url = "foo/book1"
		csg.Pages = append(csg.Pages, specUrl.String())

		tx, err := transactions.New("foo", 1, edSigner(fooKey, 1), csg)
		require.NoError(t, err)
		send(tx)
	})

	group := n.GetKeyBook("foo/book1")
	require.Len(t, group.Pages, 1)
	require.Equal(t, specUrl.String(), group.Pages[0])

	spec := n.GetKeyPage("foo/page1")
	require.Equal(t, spec.KeyBook, types.String(groupUrl.String()))
}

func TestAddKeyPage(t *testing.T) {
	subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
	nodes := RunTestNet(t, subnets, daemons, nil, true)
	n := nodes[subnets[1]][0]

	fooKey, testKey1, testKey2 := generateKey(), generateKey(), generateKey()

	u := n.ParseUrl("foo/book1")

	batch := n.db.Begin()
	require.NoError(t, acctesting.CreateADI(batch, fooKey, "foo"))
	require.NoError(t, acctesting.CreateKeyPage(batch, "foo/page1", testKey1.PubKey().Bytes()))
	require.NoError(t, acctesting.CreateKeyBook(batch, "foo/book1", "foo/page1"))
	require.NoError(t, batch.Commit())

	// Sanity check
	require.Equal(t, types.String(u.String()), n.GetKeyPage("foo/page1").KeyBook)

	n.Batch(func(send func(*transactions.Envelope)) {
		cms := new(protocol.CreateKeyPage)
		cms.Url = "foo/page2"
		cms.Keys = append(cms.Keys, &protocol.KeySpecParams{
			PublicKey: testKey2.PubKey().Bytes(),
		})

		tx, err := transactions.New("foo/book1", 2, edSigner(testKey1, 1), cms)
		require.NoError(t, err)
		send(tx)
	})

	spec := n.GetKeyPage("foo/page2")
	require.Len(t, spec.Keys, 1)
	key := spec.Keys[0]
	require.Equal(t, types.String(u.String()), spec.KeyBook)
	require.Equal(t, uint64(0), key.Nonce)
	require.Equal(t, testKey2.PubKey().Bytes(), key.PublicKey)
}

func TestAddKey(t *testing.T) {
	subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
	nodes := RunTestNet(t, subnets, daemons, nil, true)
	n := nodes[subnets[1]][0]

	fooKey, testKey := generateKey(), generateKey()

	batch := n.db.Begin()
	require.NoError(t, acctesting.CreateADI(batch, fooKey, "foo"))
	require.NoError(t, acctesting.CreateKeyPage(batch, "foo/page1", testKey.PubKey().Bytes()))
	require.NoError(t, acctesting.CreateKeyBook(batch, "foo/book1", "foo/page1"))
	require.NoError(t, batch.Commit())

	newKey := generateKey()
	n.Batch(func(send func(*transactions.Envelope)) {
		body := new(protocol.UpdateKeyPage)
		body.Operation = protocol.KeyPageOperationAdd
		body.NewKey = newKey.PubKey().Bytes()

		tx, err := transactions.New("foo/page1", 2, edSigner(testKey, 1), body)
		require.NoError(t, err)
		send(tx)
	})

	spec := n.GetKeyPage("foo/page1")
	require.Len(t, spec.Keys, 2)
	require.Equal(t, newKey.PubKey().Bytes(), spec.Keys[1].PublicKey)
}

func TestUpdateKey(t *testing.T) {
	subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
	nodes := RunTestNet(t, subnets, daemons, nil, true)
	n := nodes[subnets[1]][0]

	fooKey, testKey := generateKey(), generateKey()

	batch := n.db.Begin()
	require.NoError(t, acctesting.CreateADI(batch, fooKey, "foo"))
	require.NoError(t, acctesting.CreateKeyPage(batch, "foo/page1", testKey.PubKey().Bytes()))
	require.NoError(t, acctesting.CreateKeyBook(batch, "foo/book1", "foo/page1"))
	require.NoError(t, batch.Commit())

	newKey := generateKey()
	n.Batch(func(send func(*transactions.Envelope)) {
		body := new(protocol.UpdateKeyPage)
		body.Operation = protocol.KeyPageOperationUpdate
		body.Key = testKey.PubKey().Bytes()
		body.NewKey = newKey.PubKey().Bytes()

		tx, err := transactions.New("foo/page1", 2, edSigner(testKey, 1), body)
		require.NoError(t, err)
		send(tx)
	})

	spec := n.GetKeyPage("foo/page1")
	require.Len(t, spec.Keys, 1)
	require.Equal(t, newKey.PubKey().Bytes(), spec.Keys[0].PublicKey)
}

func TestRemoveKey(t *testing.T) {
	subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
	nodes := RunTestNet(t, subnets, daemons, nil, true)
	n := nodes[subnets[1]][0]

	fooKey, testKey1, testKey2 := generateKey(), generateKey(), generateKey()

	batch := n.db.Begin()
	require.NoError(t, acctesting.CreateADI(batch, fooKey, "foo"))
	require.NoError(t, acctesting.CreateKeyPage(batch, "foo/page1", testKey1.PubKey().Bytes(), testKey2.PubKey().Bytes()))
	require.NoError(t, acctesting.CreateKeyBook(batch, "foo/book1", "foo/page1"))
	require.NoError(t, batch.Commit())

	n.Batch(func(send func(*transactions.Envelope)) {
		body := new(protocol.UpdateKeyPage)
		body.Operation = protocol.KeyPageOperationRemove
		body.Key = testKey1.PubKey().Bytes()

		tx, err := transactions.New("foo/page1", 2, edSigner(testKey2, 1), body)
		require.NoError(t, err)
		send(tx)
	})

	spec := n.GetKeyPage("foo/page1")
	require.Len(t, spec.Keys, 1)
	require.Equal(t, testKey2.PubKey().Bytes(), spec.Keys[0].PublicKey)
}

func TestSignatorHeight(t *testing.T) {
	subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
	nodes := RunTestNet(t, subnets, daemons, nil, true)
	n := nodes[subnets[1]][0]

	liteKey, fooKey := generateKey(), generateKey()

	liteUrl, err := protocol.LiteTokenAddress(liteKey.PubKey().Bytes(), "ACME")
	require.NoError(t, err)
	tokenUrl, err := url.Parse("foo/tokens")
	require.NoError(t, err)
	keyPageUrl, err := url.Parse("foo/page0")
	require.NoError(t, err)

	batch := n.db.Begin()
	require.NoError(t, acctesting.CreateLiteTokenAccount(batch, liteKey, 1))
	require.NoError(t, batch.Commit())

	getHeight := func(u *url.URL) uint64 {
		batch := n.db.Begin()
		defer batch.Discard()
		chain, err := batch.Account(u).ReadChain(protocol.MainChain)
		require.NoError(t, err)
		return uint64(chain.Height())
	}

	liteHeight := getHeight(liteUrl)

	n.Batch(func(send func(*transactions.Envelope)) {
		adi := new(protocol.CreateIdentity)
		adi.Url = "foo"
		adi.PublicKey = fooKey.PubKey().Bytes()
		adi.KeyBookName = "book"
		adi.KeyPageName = "page0"

		tx, err := transactions.New(liteUrl.String(), 1, edSigner(liteKey, 1), adi)
		require.NoError(t, err)
		send(tx)
	})

	require.Equal(t, liteHeight, getHeight(liteUrl), "Lite account height changed")

	keyPageHeight := getHeight(keyPageUrl)

	n.Batch(func(send func(*transactions.Envelope)) {
		tac := new(protocol.CreateTokenAccount)
		tac.Url = tokenUrl.String()
		tac.TokenUrl = protocol.AcmeUrl().String()
		tx, err := transactions.New("foo", 1, edSigner(fooKey, 1), tac)
		require.NoError(t, err)
		send(tx)
	})

	require.Equal(t, keyPageHeight, getHeight(keyPageUrl), "Key page height changed")
}

func TestCreateToken(t *testing.T) {
	subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
	nodes := RunTestNet(t, subnets, daemons, nil, true)
	n := nodes[subnets[1]][0]

	fooKey := generateKey()
	batch := n.db.Begin()
	require.NoError(t, acctesting.CreateADI(batch, fooKey, "foo"))
	require.NoError(t, batch.Commit())

	n.Batch(func(send func(*transactions.Envelope)) {
		body := new(protocol.CreateToken)
		body.Url = "foo/tokens"
		body.Symbol = "FOO"
		body.Precision = 10

		tx, err := transactions.New("foo", 1, edSigner(fooKey, 1), body)
		require.NoError(t, err)
		send(tx)
	})

	n.GetTokenIssuer("foo/tokens")
}

func TestIssueTokens(t *testing.T) {
	subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
	nodes := RunTestNet(t, subnets, daemons, nil, true)
	n := nodes[subnets[1]][0]

	fooKey, liteKey := generateKey(), generateKey()
	batch := n.db.Begin()
	require.NoError(t, acctesting.CreateADI(batch, fooKey, "foo"))
	require.NoError(t, acctesting.CreateTokenIssuer(batch, "foo/tokens", "FOO", 10))
	require.NoError(t, batch.Commit())

	liteAddr, err := protocol.LiteTokenAddress(liteKey[32:], "foo/tokens")
	require.NoError(t, err)

	n.Batch(func(send func(*transactions.Envelope)) {
		body := new(protocol.IssueTokens)
		body.Recipient = liteAddr.String()
		body.Amount.SetUint64(123)

		tx, err := transactions.New("foo/tokens", 1, edSigner(fooKey, 1), body)
		require.NoError(t, err)
		send(tx)
	})

	account := n.GetLiteTokenAccount(liteAddr.String())
	require.Equal(t, "acc://foo/tokens", account.TokenUrl)
	require.Equal(t, int64(123), account.Balance.Int64())
}

func TestInvalidDeposit(t *testing.T) {
	// The lite address ends with `foo/tokens` but the token is `foo2/tokens` so
	// the synthetic transaction will fail. This test verifies that the
	// transaction fails, but more importantly it verifies that
	// `Executor.Commit()` does *not* break if DeliverTx fails with a
	// non-existent origin. This is motivated by a bug that has been fixed. This
	// bug could have been triggered by a failing SyntheticCreateChains,
	// SyntheticDepositTokens, or SyntheticDepositCredits.

	subnets, daemons := acctesting.CreateTestNet(t, 1, 1, 0)
	nodes := RunTestNet(t, subnets, daemons, nil, true)
	n := nodes[subnets[1]][0]

	liteKey := generateKey()
	liteAddr, err := protocol.LiteTokenAddress(liteKey[32:], "foo/tokens")
	require.NoError(t, err)

	id := n.Batch(func(send func(*transactions.Envelope)) {
		body := new(protocol.SyntheticDepositTokens)
		body.Token = "foo2/tokens"
		body.Amount.SetUint64(123)

		tx, err := transactions.New(liteAddr.String(), 1, edSigner(n.key.Bytes(), 1), body)
		require.NoError(t, err)
		send(tx)
	})[0]

	tx := n.GetTx(id[:])
	require.NotZero(t, tx.Status.Code)
}
