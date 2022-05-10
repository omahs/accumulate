package block

import (
	"encoding/json"
	"fmt"
	"time"

	"gitlab.com/accumulatenetwork/accumulate/config"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/indexing"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

// EndBlock implements ./Chain
func (m *Executor) EndBlock(block *Block) error {
	// Load the ledger
	ledgerUrl := m.Network.NodeUrl(protocol.Ledger)
	ledger := block.Batch.Account(ledgerUrl)
	var ledgerState *protocol.InternalLedger
	err := ledger.GetStateAs(&ledgerState)
	if err != nil {
		return err
	}

	// Set active oracle from pending
	ledgerState.ActiveOracle = ledgerState.PendingOracle

	m.logInfo("Committing",
		"height", block.Index,
		"delivered", block.State.Delivered,
		"signed", block.State.SynthSigned,
		"sent", block.State.SynthSent,
		"updated", len(block.State.ChainUpdates.Entries),
		"submitted", len(block.State.ProducedTxns))
	t := time.Now()

	// 	err = m.doEndBlock(block, ledgerState)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	// Write the updated ledger
	// 	err = ledger.PutState(ledgerState)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	m.logInfo("Committed", "height", block.Index, "duration", time.Since(t))
	// 	return nil
	// }

	// func (m *Executor) doEndBlock(block *Block, ledgerState *protocol.InternalLedger) error {

	// Load the main chain of the minor root
	rootChain, err := ledger.Chain(protocol.MinorRootChain, protocol.ChainTypeAnchor)
	if err != nil {
		return err
	}

	// Pending transaction-chain index entries
	type txChainIndexEntry struct {
		indexing.TransactionChainEntry
		Txid []byte
	}
	txChainEntries := make([]*txChainIndexEntry, 0, len(block.State.ChainUpdates.Entries))

	// Process chain updates
	accountSeen := map[string]bool{}
	for _, u := range block.State.ChainUpdates.Entries {
		accountSeen[u.Account.String()] = true

		// Do not create root chain or BPT entries for the ledger
		if ledgerUrl.Equal(u.Account) {
			continue
		}

		// Anchor and index the chain
		m.logDebug("Updated a chain", "url", fmt.Sprintf("%s#chain/%s", u.Account, u.Name))
		account := block.Batch.Account(u.Account)
		indexIndex, didIndex, err := addChainAnchor(rootChain, account, u.Account, u.Name, u.Type)
		if err != nil {
			return err
		}

		// Add a pending transaction-chain index update
		if didIndex && u.Type == protocol.ChainTypeTransaction {
			e := new(txChainIndexEntry)
			e.Txid = u.Entry
			e.Account = u.Account
			e.Chain = u.Name
			e.ChainIndex = indexIndex
			txChainEntries = append(txChainEntries, e)
		}
	}

	// Create a BlockChainUpdates Index
	err = indexing.BlockChainUpdates(block.Batch, &m.Network, uint64(block.Index)).Set(block.State.ChainUpdates.Entries)
	if err != nil {
		return err
	}

	// If dn/oracle was updated, update the ledger's oracle value, but only if
	// we're on the DN - mirroring can cause dn/oracle to be updated on the BVN
	if accountSeen[protocol.PriceOracleAuthority] && m.Network.LocalSubnetID == protocol.Directory {
		// If things go south here, don't return and error, instead, just log one
		newOracleValue, err := m.updateOraclePrice(block)
		if err != nil {
			m.logError(fmt.Sprintf("%v", err))
		} else {
			ledgerState.PendingOracle = newOracleValue
		}
	}

	// Add the synthetic transaction chain to the root chain
	var synthIndexIndex uint64
	var synthAnchorIndex uint64
	if len(block.State.ProducedTxns) > 0 {
		synthAnchorIndex = uint64(rootChain.Height())
		synthIndexIndex, err = m.anchorSynthChain(block, rootChain)
		if err != nil {
			return err
		}
	}

	// Write the updated ledger
	err = ledger.PutState(ledgerState)
	if err != nil {
		return err
	}

	// Index the root chain
	rootIndexIndex, err := addIndexChainEntry(ledger, protocol.MinorRootIndexChain, &protocol.IndexEntry{
		Source:     uint64(rootChain.Height() - 1),
		BlockIndex: uint64(block.Index),
		BlockTime:  &block.Time,
	})
	if err != nil {
		return err
	}

	// Update the transaction-chain index
	for _, e := range txChainEntries {
		e.AnchorIndex = rootIndexIndex
		err = indexing.TransactionChain(block.Batch, e.Txid).Add(&e.TransactionChainEntry)
		if err != nil {
			return err
		}
	}

	// Add transaction-chain index entries for synthetic transactions
	blockState, err := indexing.BlockState(block.Batch, ledgerUrl).Get()
	if err != nil {
		return err
	}

	for _, e := range blockState.ProducedSynthTxns {
		err = indexing.TransactionChain(block.Batch, e.Transaction).Add(&indexing.TransactionChainEntry{
			Account:     m.Network.Synthetic(),
			Chain:       protocol.MainChain,
			ChainIndex:  synthIndexIndex,
			AnchorIndex: rootIndexIndex,
		})
		if err != nil {
			return err
		}
	}

	// Build synthetic receipts on Directory nodes
	if m.Network.Type == config.Directory && len(block.State.ProducedTxns) > 0 {
		err = m.createLocalDNReceipt(block, rootChain, synthAnchorIndex)
		if err != nil {
			return err
		}
	}

	err = block.Batch.CommitBpt()
	if err != nil {
		return err
	}

	m.logInfo("Committed", "height", block.Index, "duration", time.Since(t))
	return nil
}

// updateOraclePrice reads the oracle from the oracle account and updates the
// value on the ledger.
func (m *Executor) updateOraclePrice(block *Block) (uint64, error) {
	data, err := block.Batch.Account(protocol.PriceOracle()).Data()
	if err != nil {
		return 0, fmt.Errorf("cannot retrieve oracle data entry: %v", err)
	}
	_, e, err := data.GetLatest()
	if err != nil {
		return 0, fmt.Errorf("cannot retrieve latest oracle data entry: data batch at height %d: %v", data.Height(), err)
	}

	o := protocol.AcmeOracle{}
	if e.Data == nil {
		return 0, fmt.Errorf("no data in oracle data account")
	}
	err = json.Unmarshal(e.Data[0], &o)
	if err != nil {
		return 0, fmt.Errorf("cannot unmarshal oracle data entry %x", e.Data)
	}

	if o.Price == 0 {
		return 0, fmt.Errorf("invalid oracle price, must be > 0")
	}

	return o.Price, nil
}

func (m *Executor) createLocalDNReceipt(block *Block, rootChain *database.Chain, synthAnchorIndex uint64) error {
	rootReceipt, err := rootChain.Receipt(int64(synthAnchorIndex), rootChain.Height()-1)
	if err != nil {
		return err
	}

	synthChain, err := block.Batch.Account(m.Network.Synthetic()).ReadChain(protocol.MainChain)
	if err != nil {
		return fmt.Errorf("unable to load synthetic transaction chain: %w", err)
	}

	height := synthChain.Height()
	offset := height - int64(len(block.State.ProducedTxns))
	for i, txn := range block.State.ProducedTxns {
		if txn.Type() == protocol.TransactionTypeDirectoryAnchor || txn.Type() == protocol.TransactionTypePartitionAnchor || txn.Type() == protocol.TransactionTypeMirrorSystemRecords {
			// Do not generate a receipt for the anchor
			continue
		}

		synthReceipt, err := synthChain.Receipt(offset+int64(i), height-1)
		if err != nil {
			return err
		}

		receipt, err := synthReceipt.Combine(rootReceipt)
		if err != nil {
			return err
		}

		// This should be the second signature (SyntheticSignature should be first)
		sig := new(protocol.ReceiptSignature)
		sig.SourceNetwork = m.Network.NodeUrl()
		sig.TransactionHash = *(*[32]byte)(txn.GetHash())
		sig.Receipt = *protocol.ReceiptFromManaged(receipt)
		_, err = block.Batch.Transaction(txn.GetHash()).AddSignature(sig)
		if err != nil {
			return err
		}
	}
	return nil
}

// anchorSynthChain anchors the synthetic transaction chain.
func (m *Executor) anchorSynthChain(block *Block, rootChain *database.Chain) (indexIndex uint64, err error) {
	url := m.Network.Synthetic()
	indexIndex, _, err = addChainAnchor(rootChain, block.Batch.Account(url), url, protocol.MainChain, protocol.ChainTypeTransaction)
	if err != nil {
		return 0, err
	}

	block.State.ChainUpdates.DidUpdateChain(indexing.ChainUpdate{
		Name:    protocol.MainChain,
		Type:    protocol.ChainTypeTransaction,
		Account: url,
	})

	return indexIndex, nil
}
