package block

import (
	"bytes"
	"errors"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/indexing"
	"gitlab.com/accumulatenetwork/accumulate/internal/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	"gitlab.com/accumulatenetwork/accumulate/smt/storage"
	"gitlab.com/accumulatenetwork/accumulate/types/api/query"
)

func (m *Executor) queryMinorBlocksFromDN(batch *database.Batch, req *query.RequestMinorBlocksFromDN) (*query.ResponseMinorBlocks, *protocol.Error) {
	ledgerAcc := batch.Account(m.Network.NodeUrl(protocol.Ledger))
	var ledger *protocol.InternalLedger
	err := ledgerAcc.GetStateAs(&ledger)
	if err != nil {
		return nil, &protocol.Error{Code: protocol.ErrorCodeUnMarshallingError, Message: err}
	}

	idxChain, err := ledgerAcc.ReadChain(protocol.MinorRootIndexChain)
	if err != nil {
		return nil, &protocol.Error{Code: protocol.ErrorCodeQueryChainUpdatesError, Message: err}
	}

	startIndex, _, err := indexing.SearchIndexChain(idxChain, uint64(idxChain.Height())-1, indexing.MatchAfter, indexing.SearchIndexChainByBlock(req.Start))
	if err != nil {
		return nil, &protocol.Error{Code: protocol.ErrorCodeQueryEntriesError, Message: err}
	}

	entryIdx := startIndex

	resp := query.ResponseMinorBlocks{TotalBlocks: uint64(ledger.Index)}
	curEntry := new(protocol.IndexEntry)
	resultCnt := uint64(0)

	var rmtQuerier RemoteMinorBlockQuerier
	if req.TxFetchMode == query.TxFetchModeExpand {
		rmtQuerier = NewRemoteMinorBlockQuerier(m.Router)
	}

resultLoop:
	for resultCnt < req.Limit {
		err = idxChain.EntryAs(int64(entryIdx), curEntry)
		switch {
		case err == nil:
		case errors.Is(err, storage.ErrNotFound):
			break resultLoop
		default:
			return nil, &protocol.Error{Code: protocol.ErrorCodeUnMarshallingError, Message: err}
		}

		minorEntry := new(query.ResponseMinorEntry)
		for {
			if req.BlockFilterMode == query.BlockFilterModeExcludeNone {
				minorEntry.BlockIndex = req.Start + resultCnt

				// Create new entry, when BlockFilterModeExcludeNone append empty entry when blocks were missing
				if minorEntry.BlockIndex < curEntry.BlockIndex || curEntry.BlockIndex == 0 {
					resp.Entries = append(resp.Entries, minorEntry)
					resultCnt++
					minorEntry = new(query.ResponseMinorEntry)
					continue
				}
			} else {
				minorEntry.BlockIndex = curEntry.BlockIndex
			}
			break
		}
		minorEntry.BlockTime = curEntry.BlockTime

		if req.TxFetchMode < query.TxFetchModeOmit {
			chainUpdatesIndex, err := indexing.BlockChainUpdates(batch, &m.Network, curEntry.BlockIndex).Get()
			if err != nil {
				return nil, &protocol.Error{Code: protocol.ErrorCodeChainIdError, Message: err}
			}

			minorEntry.TxCount = uint64(0)
			systemTxCount := uint64(0)
			var lastTxid []byte
			for _, updIdx := range chainUpdatesIndex.Entries {
				if bytes.Equal(updIdx.Entry, lastTxid) { // There are like 4 ChainUpdates for each tx, we don't need duplicates
					continue
				}

				if req.TxFetchMode <= query.TxFetchModeIds {
					minorEntry.TxIds = append(minorEntry.TxIds, updIdx.Entry)
				}
				if req.TxFetchMode == query.TxFetchModeExpand {
					qr, err := m.queryByTxId(batch, updIdx.Entry, false, false)
					if err == nil {
						minorEntry.TxCount++
						txt := qr.Envelope.Transaction[0].Body.Type()
						if txt == protocol.TransactionTypePartitionAnchor {
							body, ok := qr.Envelope.Transaction[0].Body.(*protocol.PartitionAnchor)
							if !ok {
								return nil, &protocol.Error{Code: protocol.ErrorCodeQueryEntriesError, Message: err}
							}

							rmtQuerier.SubmitQuery(body.Source, body.Block, minorEntry)
						} else if txt.IsSystem() {
							systemTxCount++
						} else if req.TxFetchMode == query.TxFetchModeExpand {
							minorEntry.Transactions = append(minorEntry.Transactions, qr)
						}
					}
				} else {
					minorEntry.TxCount++
				}
				lastTxid = updIdx.Entry
			}
			if minorEntry.TxCount <= systemTxCount && req.BlockFilterMode == query.BlockFilterModeExcludeEmpty {
				entryIdx++
				continue
			}
		}
		resp.Entries = append(resp.Entries, minorEntry)
		entryIdx++
		resultCnt++
	}
	if rmtQuerier != nil {
		rmtQuerier.Flush()
	}
	return &resp, nil
}

func (m *Executor) queryMinorBlocksByUrl(batch *database.Batch, req *query.RequestMinorBlocksByUrl) (*query.ResponseMinorBlocks, *protocol.Error) {
	ledgerAcc := batch.Account(m.Network.NodeUrl(protocol.Ledger))
	var ledger *protocol.InternalLedger
	err := ledgerAcc.GetStateAs(&ledger)
	if err != nil {
		return nil, &protocol.Error{Code: protocol.ErrorCodeUnMarshallingError, Message: err}
	}

	idxChain, err := ledgerAcc.ReadChain(protocol.MinorRootIndexChain)
	if err != nil {
		return nil, &protocol.Error{Code: protocol.ErrorCodeQueryChainUpdatesError, Message: err}
	}

	rmb := &query.ResponseMinorBlocks{TotalBlocks: ledger.Index}

	for _, rg := range req.Ranges {
		err := m.queryMinorBlocksByUrlRange(batch, rmb, rg, idxChain, req.BlockFilterMode, req.TxFetchMode)
		if err != nil {
			return nil, err
		}
	}
	return rmb, nil
}

func (m *Executor) queryMinorBlocksByUrlRange(batch *database.Batch, resp *query.ResponseMinorBlocks, qryRange query.Range,
	idxChain *database.Chain, blockFilterMode query.BlockFilterMode, txFetchMode query.TxFetchMode) *protocol.Error {

	startIndex, _, err := indexing.SearchIndexChain(idxChain, uint64(idxChain.Height())-1, indexing.MatchAfter, indexing.SearchIndexChainByBlock(qryRange.Start))
	if err != nil {
		return &protocol.Error{Code: protocol.ErrorCodeQueryEntriesError, Message: err}
	}

	entryIdx := startIndex
	curEntry := new(protocol.IndexEntry)
	resultCnt := uint64(0)

resultLoop:
	for resultCnt < qryRange.Count {
		err = idxChain.EntryAs(int64(entryIdx), curEntry)
		switch {
		case err == nil:
		case errors.Is(err, storage.ErrNotFound):
			break resultLoop
		default:
			return &protocol.Error{Code: protocol.ErrorCodeUnMarshallingError, Message: err}
		}

		minorEntry := new(query.ResponseMinorEntry)
		for {
			if blockFilterMode == query.BlockFilterModeExcludeNone {
				minorEntry.BlockIndex = qryRange.Start + resultCnt

				// Create new entry, when BlockFilterModeExcludeNone append empty entry when blocks were missing
				if minorEntry.BlockIndex < curEntry.BlockIndex || curEntry.BlockIndex == 0 {
					resp.Entries = append(resp.Entries, minorEntry)
					resultCnt++
					minorEntry = new(query.ResponseMinorEntry)
					continue
				}
			} else {
				minorEntry.BlockIndex = curEntry.BlockIndex
			}
			break
		}
		minorEntry.BlockTime = curEntry.BlockTime

		if txFetchMode < query.TxFetchModeOmit {
			chainUpdatesIndex, err := indexing.BlockChainUpdates(batch, &m.Network, curEntry.BlockIndex).Get()
			if err != nil {
				return nil
			}

			minorEntry.TxCount = uint64(0)
			systemTxCount := uint64(0)
			var lastTxid []byte
			for _, updIdx := range chainUpdatesIndex.Entries {
				if bytes.Equal(updIdx.Entry, lastTxid) { // There are like 4 ChainUpdates for each tx, we don't need duplicates
					continue
				}

				if txFetchMode <= query.TxFetchModeIds {
					minorEntry.TxIds = append(minorEntry.TxIds, updIdx.Entry)
				}
				if txFetchMode == query.TxFetchModeExpand {
					qr, err := m.queryByTxId(batch, updIdx.Entry, false, false)
					if err == nil {
						minorEntry.TxCount++
						txt := qr.Envelope.Transaction[0].Body.Type()
						if txt.IsSystem() {
							systemTxCount++
						} else if txFetchMode == query.TxFetchModeExpand {
							minorEntry.Transactions = append(minorEntry.Transactions, qr)
						}
					}
				} else {
					minorEntry.TxCount++
				}
				lastTxid = updIdx.Entry
			}
			if minorEntry.TxCount <= systemTxCount && blockFilterMode == query.BlockFilterModeExcludeEmpty {
				entryIdx++
				continue
			}
		}
		resp.Entries = append(resp.Entries, minorEntry)
		entryIdx++
		resultCnt++
	}
	return nil
}

func (m *Executor) expandChainEntries(batch *database.Batch, entries []string) ([]protocol.Account, error) {
	expEntries := make([]protocol.Account, len(entries))
	for i, entry := range entries {
		index := i
		u, err := url.Parse(entry)
		if err != nil {
			return nil, err
		}
		r, err := batch.Account(u).GetState()
		if err != nil {
			return nil, err
		}
		expEntries[index] = r
	}
	return expEntries, nil
}
