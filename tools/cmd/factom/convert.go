package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/FactomProject/factomd/common/entryBlock"
	"github.com/spf13/cobra"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/errors"
	"gitlab.com/accumulatenetwork/accumulate/internal/indexing"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	"gitlab.com/accumulatenetwork/accumulate/smt/storage/badger"
	"gitlab.com/accumulatenetwork/accumulate/tools/internal/factom"
)

var cmdConvert = &cobra.Command{
	Use:   "convert [output database] [output snapshot] [input object files*]",
	Short: "convert a Factom object dump to Accumulate",
	Args:  cobra.MinimumNArgs(3),
	Run:   convert,
}

func init() {
	cmd.AddCommand(cmdConvert)
}

func convert(_ *cobra.Command, args []string) {
	db, err := database.OpenBadger(args[0], nil)
	checkf(err, "output database")
	defer db.Close()

	output, err := os.Create(args[1])
	checkf(err, "output file")
	defer output.Close()

	for _, filename := range args[2:] {
		input, err := ioutil.ReadFile(filename)
		checkf(err, "read %s", filename)

		entries := map[[32]byte][]*entryBlock.Entry{}
		err = factom.ReadObjectFile(input, func(_ *factom.Header, object interface{}) {
			entry, ok := object.(*entryBlock.Entry)
			if !ok {
				return
			}

			id := entry.ChainID.Fixed()
			entries[id] = append(entries[id], entry)
		})
		checkf(err, "process object file")

		for chainId, entries := range entries {
			chainId := chainId // See docs/developer/rangevarref.md
			address, err := protocol.LiteDataAddress(chainId[:])
			checkf(err, "create LDA URL")

			// Commit each account separately so we don't exceed Badger's limits
			batch := db.Begin(true)
			account := batch.Account(address)

			var lda *protocol.LiteDataAccount
			err = account.Main().GetAs(&lda)
			switch {
			case err == nil:
				// Record exists
			case errors.Is(err, errors.StatusNotFound):
				// Create the record
				lda = new(protocol.LiteDataAccount)
				lda.Url = address
				err = account.Main().Put(lda)
				checkf(err, "store record")
			default:
				checkf(err, "load record")
			}

			for _, entry := range entries {
				entry := factom.ConvertEntry(entry).Wrap()
				entryHash, err := protocol.ComputeFactomEntryHashForAccount(chainId[:], entry.GetData())
				checkf(err, "calculate entry hash")

				txn := new(protocol.Transaction)
				txn.Header.Principal = address
				txn.Body = &protocol.WriteData{Entry: entry}

				// Each transaction needs to be unique so add a timestamp
				// TODO: Derive this from Factom?
				txn.Header.Memo = fmt.Sprintf("Imported on %v", time.Now())

				result := new(protocol.WriteDataResult)
				result.AccountID = chainId[:]
				result.AccountUrl = address
				result.EntryHash = *(*[32]byte)(entryHash)

				status := new(protocol.TransactionStatus)
				status.TxID = txn.ID()
				status.Code = errors.StatusDelivered
				status.Result = result

				txnrec := batch.Transaction(txn.GetHash())
				_, err = txnrec.Main().Get()
				switch {
				case err == nil:
					fatalf("Somehow we created a duplicate transaction")
				case errors.Is(err, errors.StatusNotFound):
					// Ok
				default:
					checkf(err, "check for duplicate transaction")
				}

				err = indexing.Data(batch, address).Put(entryHash, txn.GetHash())
				checkf(err, "add data index")

				err = txnrec.Main().Put(&database.SigOrTxn{Transaction: txn})
				checkf(err, "store transaction")

				err = txnrec.Status().Put(status)
				checkf(err, "store status")

				mainChain, err := account.MainChain().Get()
				checkf(err, "load main chain")
				err = mainChain.AddEntry(txn.GetHash(), false)
				checkf(err, "store main chain entry")
			}

			err = batch.Commit()
			checkf(err, "commit")
		}

		err = db.GC(0.5)
		if err != nil && !errors.Is(err, badger.ErrNoRewrite) {
			checkf(err, "compact database")
		}
	}

	check(db.View(func(batch *database.Batch) error {
		return batch.SaveFactomSnapshot(output)
	}))
}
