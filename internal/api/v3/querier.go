// Copyright 2023 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package api

import (
	"context"
	"strings"

	"github.com/tendermint/tendermint/libs/log"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/database/indexing"
	"gitlab.com/accumulatenetwork/accumulate/internal/logging"
	"gitlab.com/accumulatenetwork/accumulate/internal/node/config"
	"gitlab.com/accumulatenetwork/accumulate/pkg/api/v3"
	"gitlab.com/accumulatenetwork/accumulate/pkg/errors"
	"gitlab.com/accumulatenetwork/accumulate/pkg/types/merkle"
	"gitlab.com/accumulatenetwork/accumulate/pkg/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

const defaultPageSize = 50
const maxPageSize = 100

type Querier struct {
	logger    logging.OptionalLogger
	db        database.Viewer
	partition config.NetworkUrl
}

var _ api.Querier = (*Querier)(nil)

type QuerierParams struct {
	Logger    log.Logger
	Database  database.Viewer
	Partition string
}

func NewQuerier(params QuerierParams) *Querier {
	s := new(Querier)
	s.logger.L = params.Logger
	s.db = params.Database
	s.partition.URL = protocol.PartitionUrl(params.Partition)
	return s
}

func (s *Querier) Type() api.ServiceType { return api.ServiceTypeQuery }

func (s *Querier) Query(ctx context.Context, scope *url.URL, query api.Query) (api.Record, error) {
	// Ensure the query parameters are valid
	if query == nil {
		query = new(api.DefaultQuery)
	}
	err := query.IsValid()
	if err != nil {
		return nil, errors.UnknownError.WithFormat("query is invalid: %w", err)
	}

	switch query := query.(type) {
	case *api.ChainQuery:
		fixRange(query.Range)
	case *api.DataQuery:
		fixRange(query.Range)
	case *api.DirectoryQuery:
		fixRange(query.Range)
	case *api.PendingQuery:
		fixRange(query.Range)
	case *api.BlockQuery:
		fixRange(query.MinorRange)
		fixRange(query.MajorRange)
		fixRange(query.EntryRange)
	}

	// Start a batch. If the ABCI were updated to commit in the middle of a
	// block, this would no longer be safe.
	var r api.Record
	err = s.db.View(func(batch *database.Batch) error {
		r, err = s.query(ctx, batch, scope, query)
		return err
	})
	return r, err
}

func (s *Querier) query(ctx context.Context, batch *database.Batch, scope *url.URL, query api.Query) (api.Record, error) {
	switch query := query.(type) {
	case *api.DefaultQuery:
		if txid, err := scope.AsTxID(); err == nil {
			h := txid.Hash()
			return s.queryTransactionOrSignature(ctx, batch, batch.Transaction(h[:]))
		}

		return s.queryAccount(ctx, batch, batch.Account(scope), query.IncludeReceipt)

	case *api.ChainQuery:
		if query.Name == "" {
			if txid, err := scope.AsTxID(); err == nil {
				h := txid.Hash()
				return s.queryTransactionChains(ctx, batch, batch.Transaction(h[:]), query.IncludeReceipt)
			}
			return s.queryAccountChains(ctx, batch.Account(scope))
		}

		record, err := batch.Account(scope).ChainByName(query.Name)
		if err != nil {
			return nil, errors.InternalError.WithFormat("get chain %s: %w", query.Name, err)
		}

		if query.Index != nil {
			return s.queryChainEntryByIndex(ctx, batch, record, *query.Index, true, query.IncludeReceipt)
		}

		if query.Entry != nil {
			return s.queryChainEntryByValue(ctx, batch, record, query.Entry, true, query.IncludeReceipt)
		}

		if query.Range != nil {
			return s.queryChainEntryRange(ctx, batch, record, query.Range)
		}

		return s.queryChain(ctx, record)

	case *api.DataQuery:
		if query.Index != nil {
			return s.queryDataEntryByIndex(ctx, batch, indexing.Data(batch, scope), *query.Index, true)
		}

		if query.Entry != nil {
			return s.queryDataEntryByHash(ctx, batch, indexing.Data(batch, scope), query.Entry, true)
		}

		if query.Range != nil {
			return s.queryDataEntryRange(ctx, batch, indexing.Data(batch, scope), query.Range)
		}

		return s.queryLastDataEntry(ctx, batch, indexing.Data(batch, scope))

	case *api.DirectoryQuery:
		return s.queryDirectoryRange(ctx, batch, batch.Account(scope), query.Range)

	case *api.PendingQuery:
		return s.queryPendingRange(ctx, batch, batch.Account(scope), query.Range)

	case *api.BlockQuery:
		if query.Minor != nil {
			return s.queryMinorBlock(ctx, batch, *query.Minor, query.EntryRange)
		}

		if query.Major != nil {
			return s.queryMajorBlock(ctx, batch, *query.Major, query.MinorRange, query.OmitEmpty)
		}

		if query.MinorRange != nil {
			return s.queryMinorBlockRange(ctx, batch, query.MinorRange, query.OmitEmpty)
		}

		return s.queryMajorBlockRange(ctx, batch, query.MajorRange, query.OmitEmpty)

	case *api.AnchorSearchQuery:
		return s.searchForAnchor(ctx, batch, batch.Account(scope), query.Anchor, query.IncludeReceipt)

	case *api.PublicKeySearchQuery:
		hash, err := protocol.PublicKeyHash(query.PublicKey, query.Type)
		if err != nil {
			return nil, errors.UnknownError.WithFormat("hash key: %w", err)
		}
		return s.searchForKeyEntry(ctx, batch, scope, func(s protocol.Signer) (int, protocol.KeyEntry, bool) {
			return s.EntryByKeyHash(hash)
		})

	case *api.PublicKeyHashSearchQuery:
		return s.searchForKeyEntry(ctx, batch, scope, func(s protocol.Signer) (int, protocol.KeyEntry, bool) {
			return s.EntryByKeyHash(query.PublicKeyHash)
		})

	case *api.DelegateSearchQuery:
		return s.searchForKeyEntry(ctx, batch, scope, func(s protocol.Signer) (int, protocol.KeyEntry, bool) {
			return s.EntryByDelegate(query.Delegate)
		})

	case *api.MessageHashSearchQuery:
		return s.searchForTransactionHash(ctx, batch, query.Hash)

	default:
		return nil, errors.NotAllowed.WithFormat("unknown query type %v", query.QueryType())
	}
}

func (s *Querier) queryAccount(ctx context.Context, batch *database.Batch, record *database.Account, wantReceipt bool) (*api.AccountRecord, error) {
	r := new(api.AccountRecord)

	state, err := record.Main().Get()
	if err != nil {
		return nil, errors.UnknownError.WithFormat("load state: %w", err)
	}
	r.Account = state

	switch state.Type() {
	case protocol.AccountTypeIdentity, protocol.AccountTypeKeyBook:
		directory, err := record.Directory().Get()
		if err != nil {
			return nil, errors.UnknownError.WithFormat("load directory: %w", err)
		}
		r.Directory, _ = api.MakeRange(directory, 0, defaultPageSize, func(v *url.URL) (*api.UrlRecord, error) {
			return &api.UrlRecord{Value: v}, nil
		})
		r.Directory.Total = uint64(len(directory))
	}

	pending, err := record.Pending().Get()
	if err != nil {
		return nil, errors.UnknownError.WithFormat("load pending: %w", err)
	}
	r.Pending, _ = api.MakeRange(pending, 0, defaultPageSize, func(v *url.TxID) (*api.TxIDRecord, error) {
		return &api.TxIDRecord{Value: v}, nil
	})
	r.Pending.Total = uint64(len(pending))

	if !wantReceipt {
		return r, nil
	}

	block, receipt, err := indexing.ReceiptForAccountState(s.partition, batch, record)
	if err != nil {
		return nil, errors.UnknownError.WithFormat("get state receipt: %w", err)
	}

	r.Receipt = new(api.Receipt)
	r.Receipt.Receipt = *receipt
	r.Receipt.LocalBlock = block.BlockIndex
	if block.BlockTime != nil {
		r.Receipt.LocalBlockTime = *block.BlockTime
	}
	return r, nil
}

func (s *Querier) queryTransactionOrSignature(ctx context.Context, batch *database.Batch, record *database.Transaction) (api.Record, error) {
	return loadTransactionOrSignature(batch, record)
}

func (s *Querier) queryTransaction(ctx context.Context, batch *database.Batch, record *database.Transaction) (*api.TransactionRecord, error) {
	state, err := record.Main().Get()
	if err != nil {
		return nil, errors.UnknownError.WithFormat("load state: %w", err)
	}
	if state.Transaction == nil {
		return nil, errors.Conflict.WithFormat("record is not a transaction")
	}

	return loadTransaction(batch, record, state.Transaction)
}

func (s *Querier) querySignature(ctx context.Context, batch *database.Batch, record *database.Transaction) (*api.SignatureRecord, error) {
	state, err := record.Main().Get()
	if err != nil {
		return nil, errors.UnknownError.WithFormat("load state: %w", err)
	}
	if state.Signature == nil {
		return nil, errors.Conflict.WithFormat("record is not a signature")
	}

	return loadSignature(batch, state.Signature, state.Txid)
}

func (s *Querier) queryAccountChains(ctx context.Context, record *database.Account) (*api.RecordRange[*api.ChainRecord], error) {
	chains, err := record.Chains().Get()
	if err != nil {
		return nil, errors.UnknownError.WithFormat("load chains index: %w", err)
	}

	r := new(api.RecordRange[*api.ChainRecord])
	r.Total = uint64(len(chains))
	r.Records = make([]*api.ChainRecord, len(chains))
	for i, c := range chains {
		cr, err := s.queryChainByName(ctx, record, c.Name)
		if err != nil {
			return nil, errors.UnknownError.WithFormat("chain %s: %w", c.Name, err)
		}

		r.Records[i] = cr
	}
	return r, nil
}

func (s *Querier) queryTransactionChains(ctx context.Context, batch *database.Batch, record *database.Transaction, wantReceipt bool) (*api.RecordRange[*api.ChainEntryRecord[api.Record]], error) {
	entries, err := record.Chains().Get()
	if err != nil {
		return nil, errors.UnknownError.WithFormat("load transaction chains: %w", err)
	}

	return api.MakeRange(entries, 0, 0, func(e *database.TransactionChainEntry) (*api.ChainEntryRecord[api.Record], error) {
		c, err := batch.Account(e.Account).ChainByName(e.Chain)
		if err != nil {
			return nil, errors.UnknownError.WithFormat("load %v %s chain: %w", e.Account, e.Chain, err)
		}

		r, err := s.queryChainEntryByIndex(ctx, batch, c, e.ChainIndex, false, wantReceipt)
		if err != nil {
			return nil, errors.UnknownError.Wrap(err)
		}

		r.Account = e.Account
		return r, nil
	})
}

func (s *Querier) queryChainByName(ctx context.Context, record *database.Account, name string) (*api.ChainRecord, error) {
	chain, err := record.ChainByName(name)
	if err != nil {
		return nil, errors.InternalError.WithFormat("get chain %s: %w", name, err)
	}

	return s.queryChain(ctx, chain)
}

func (s *Querier) queryChain(ctx context.Context, record *database.Chain2) (*api.ChainRecord, error) {
	head, err := record.Head().Get()
	if err != nil {
		return nil, errors.UnknownError.WithFormat("load head: %w", err)
	}

	r := new(api.ChainRecord)
	r.Name = record.Name()
	r.Type = record.Type()
	r.Count = uint64(head.Count)
	r.State = head.Pending
	return r, nil
}

func (s *Querier) queryChainEntryByIndex(ctx context.Context, batch *database.Batch, record *database.Chain2, index uint64, expand, wantReceipt bool) (*api.ChainEntryRecord[api.Record], error) {
	value, err := record.Entry(int64(index))
	if err != nil {
		return nil, errors.UnknownError.WithFormat("get entry: %w", err)
	}
	return s.queryChainEntry(ctx, batch, record, index, value, expand, wantReceipt)
}

func (s *Querier) queryChainEntryByValue(ctx context.Context, batch *database.Batch, record *database.Chain2, value []byte, expand, wantReceipt bool) (*api.ChainEntryRecord[api.Record], error) {
	index, err := record.IndexOf(value)
	if err != nil {
		return nil, errors.UnknownError.WithFormat("get entry index: %w", err)
	}
	return s.queryChainEntry(ctx, batch, record, uint64(index), value, expand, wantReceipt)
}

func (s *Querier) queryChainEntry(ctx context.Context, batch *database.Batch, record *database.Chain2, index uint64, value []byte, expand, wantReceipt bool) (*api.ChainEntryRecord[api.Record], error) {
	r := new(api.ChainEntryRecord[api.Record])
	r.Name = record.Name()
	r.Type = record.Type()
	r.Index = index
	r.Entry = *(*[32]byte)(value)

	ms, err := record.Inner().GetAnyState(int64(index))
	if err != nil {
		return nil, errors.UnknownError.WithFormat("load state: %w", err)
	}
	r.State = ms.Pending

	if expand {
		switch r.Type {
		case merkle.ChainTypeIndex:
			v := new(protocol.IndexEntry)
			if v.UnmarshalBinary(value) == nil {
				r.Value = &api.IndexEntryRecord{Value: v}
			}

		case merkle.ChainTypeTransaction:
			record := batch.Transaction(value)
			var typ string
			var err error
			if strings.EqualFold(r.Name, "signature") {
				typ = "signature"
				r.Value, err = s.querySignature(ctx, batch, record)
			} else {
				typ = "transaction"
				r.Value, err = s.queryTransaction(ctx, batch, record)
			}
			if err != nil {
				return nil, errors.UnknownError.WithFormat("load %s: %w", typ, err)
			}
		}
	}

	if wantReceipt {
		block, rootIndexIndex, receipt, err := indexing.ReceiptForChainIndex(s.partition, batch, record, int64(index))
		if err != nil {
			return nil, errors.UnknownError.WithFormat("get chain receipt: %w", err)
		}

		r.Receipt = new(api.Receipt)
		r.Receipt.Receipt = *receipt
		r.Receipt.LocalBlock = block.BlockIndex
		if block.BlockTime != nil {
			r.Receipt.LocalBlockTime = *block.BlockTime
		}

		// Find the major block
		rxc, err := batch.Account(s.partition.AnchorPool()).MajorBlockChain().Get()
		if err != nil {
			return nil, errors.InternalError.WithFormat("load root index chain: %w", err)
		}
		_, major, err := indexing.SearchIndexChain(rxc, 0, indexing.MatchAfter, indexing.SearchIndexChainByRootIndexIndex(rootIndexIndex))
		switch {
		case err == nil:
			r.Receipt.MajorBlock = major.BlockIndex
		case errors.Is(err, errors.NotFound):
			// Not in a major block yet
		default:
			return nil, errors.InternalError.WithFormat("locate major block for root index entry %d: %w", rootIndexIndex, err)
		}
	}

	return r, nil
}

func (s *Querier) queryDataEntryByIndex(ctx context.Context, batch *database.Batch, record *indexing.DataIndexer, index uint64, expand bool) (*api.ChainEntryRecord[*api.TransactionRecord], error) {
	entryHash, err := record.Entry(index)
	if err != nil {
		return nil, errors.UnknownError.WithFormat("get entry hash: %w", err)
	}

	return s.queryDataEntry(ctx, batch, record, index, entryHash, expand)
}

func (s *Querier) queryLastDataEntry(ctx context.Context, batch *database.Batch, record *indexing.DataIndexer) (*api.ChainEntryRecord[*api.TransactionRecord], error) {
	count, err := record.Count()
	if err != nil {
		return nil, errors.UnknownError.WithFormat("get count: %w", err)
	}
	if count == 0 {
		return nil, errors.NotFound.WithFormat("account has no data entries")
	}
	return s.queryDataEntryByIndex(ctx, batch, record, count-1, true)
}

func (s *Querier) queryDataEntryByHash(ctx context.Context, batch *database.Batch, record *indexing.DataIndexer, entryHash []byte, expand bool) (*api.ChainEntryRecord[*api.TransactionRecord], error) {
	// TODO: Find a way to get the index without scanning the entire set
	return s.queryDataEntry(ctx, batch, record, 0, entryHash, expand)
}

func (s *Querier) queryDataEntry(ctx context.Context, batch *database.Batch, record *indexing.DataIndexer, index uint64, entryHash []byte, expand bool) (*api.ChainEntryRecord[*api.TransactionRecord], error) {
	txnHash, err := record.Transaction(entryHash)
	if err != nil {
		return nil, errors.UnknownError.WithFormat("get transaction hash: %w", err)
	}

	r := new(api.ChainEntryRecord[*api.TransactionRecord])
	r.Name = "data"
	r.Type = protocol.ChainTypeTransaction
	r.Index = index
	r.Entry = *(*[32]byte)(entryHash)

	if expand {
		r.Value, err = s.queryTransaction(ctx, batch, batch.Transaction(txnHash))
		if err != nil {
			return nil, errors.UnknownError.WithFormat("load transaction: %w", err)
		}
	}

	return r, nil
}

func (s *Querier) queryChainEntryRange(ctx context.Context, batch *database.Batch, record *database.Chain2, opts *api.RangeOptions) (*api.RecordRange[*api.ChainEntryRecord[api.Record]], error) {
	head, err := record.Head().Get()
	if err != nil {
		return nil, errors.UnknownError.WithFormat("get head: %w", err)
	}

	r := new(api.RecordRange[*api.ChainEntryRecord[api.Record]])
	r.Total = uint64(head.Count)
	allocRange(r, opts, zeroBased)

	for i := range r.Records {
		var expand bool
		if opts.Expand != nil {
			expand = *opts.Expand
		}
		r.Records[i], err = s.queryChainEntryByIndex(ctx, batch, record, r.Start+uint64(i), expand, false)
		if err != nil {
			return nil, errors.UnknownError.WithFormat("get entry %d: %w", r.Start+uint64(i), err)
		}
	}
	return r, nil
}

func (s *Querier) queryDataEntryRange(ctx context.Context, batch *database.Batch, record *indexing.DataIndexer, opts *api.RangeOptions) (*api.RecordRange[*api.ChainEntryRecord[*api.TransactionRecord]], error) {
	total, err := record.Count()
	if err != nil {
		return nil, errors.UnknownError.WithFormat("get count: %w", err)
	}

	r := new(api.RecordRange[*api.ChainEntryRecord[*api.TransactionRecord]])
	r.Total = total
	allocRange(r, opts, zeroBased)

	expand := true
	if opts.Expand != nil {
		expand = *opts.Expand
	}
	for i := range r.Records {
		r.Records[i], err = s.queryDataEntryByIndex(ctx, batch, record, r.Start+uint64(i), expand)
		if err != nil {
			return nil, errors.UnknownError.WithFormat("get entry %d: %w", r.Start+uint64(i), err)
		}
	}
	return r, nil
}

func (s *Querier) queryDirectoryRange(ctx context.Context, batch *database.Batch, record *database.Account, opts *api.RangeOptions) (*api.RecordRange[api.Record], error) {
	directory, err := record.Directory().Get()
	if err != nil {
		return nil, errors.UnknownError.WithFormat("load directory: %w", err)
	}
	r, err := api.MakeRange(directory, opts.Start, *opts.Count, func(v *url.URL) (api.Record, error) {
		if opts.Expand == nil || !*opts.Expand {
			return &api.UrlRecord{Value: v}, nil
		}

		r, err := s.queryAccount(ctx, batch, batch.Account(v), false)
		if err != nil {
			return nil, errors.UnknownError.WithFormat("expand directory entry: %w", err)
		}
		return r, nil
	})
	return r, errors.UnknownError.Wrap(err)
}

func (s *Querier) queryPendingRange(ctx context.Context, batch *database.Batch, record *database.Account, opts *api.RangeOptions) (*api.RecordRange[api.Record], error) {
	pending, err := record.Pending().Get()
	if err != nil {
		return nil, errors.UnknownError.WithFormat("load pending: %w", err)
	}
	r, err := api.MakeRange(pending, opts.Start, *opts.Count, func(v *url.TxID) (api.Record, error) {
		if opts.Expand == nil || !*opts.Expand {
			return &api.TxIDRecord{Value: v}, nil
		}

		h := v.Hash()
		r, err := s.queryTransaction(ctx, batch, batch.Transaction(h[:]))
		if err != nil {
			return nil, errors.UnknownError.WithFormat("expand pending entry: %w", err)
		}
		return r, nil
	})
	return r, errors.UnknownError.Wrap(err)
}

func (s *Querier) queryMinorBlock(ctx context.Context, batch *database.Batch, minorIndex uint64, entryRange *api.RangeOptions) (*api.MinorBlockRecord, error) {
	if entryRange == nil {
		entryRange = new(api.RangeOptions)
		entryRange.Count = new(uint64)
		*entryRange.Count = defaultPageSize
	}

	var ledger *protocol.BlockLedger
	err := batch.Account(s.partition.BlockLedger(minorIndex)).Main().GetAs(&ledger)
	if err != nil {
		return nil, errors.UnknownError.WithFormat("load block ledger: %w", err)
	}

	r := new(api.MinorBlockRecord)
	r.Index = ledger.Index
	r.Time = &ledger.Time
	r.Entries = new(api.RecordRange[*api.ChainEntryRecord[api.Record]])
	r.Entries.Total = uint64(len(ledger.Entries))

	allocRange(r.Entries, entryRange, zeroBased)
	for i := range r.Entries.Records {
		e := ledger.Entries[r.Entries.Start+uint64(i)]
		r.Entries.Records[i], err = loadBlockEntry(batch, e)
		if err == nil {
			continue
		}

		var e2 *errors.Error
		if !errors.As(err, &e2) || e2.Code.IsServerError() {
			return nil, errors.UnknownError.Wrap(err)
		}
	}
	return r, nil
}

func (s *Querier) queryMinorBlockRange(ctx context.Context, batch *database.Batch, minorRange *api.RangeOptions, omitEmpty bool) (*api.RecordRange[*api.MinorBlockRecord], error) {
	var ledger *protocol.SystemLedger
	err := batch.Account(s.partition.Ledger()).Main().GetAs(&ledger)
	if err != nil {
		return nil, errors.UnknownError.WithFormat("load system ledger: %w", err)
	}

	r := new(api.RecordRange[*api.MinorBlockRecord])
	r.Total = ledger.Index
	allocRange(r, minorRange, oneBased)

	r.Records, err = s.queryMinorBlockRange2(ctx, batch, r.Records, r.Start, ledger.Index, omitEmpty)
	if err != nil {
		return nil, errors.UnknownError.Wrap(err)
	}
	return r, nil
}

func (s *Querier) queryMajorBlock(ctx context.Context, batch *database.Batch, majorIndex uint64, minorRange *api.RangeOptions, omitEmpty bool) (*api.MajorBlockRecord, error) {
	if minorRange == nil {
		minorRange = new(api.RangeOptions)
		minorRange.Count = new(uint64)
		*minorRange.Count = defaultPageSize
	}

	chain, err := batch.Account(s.partition.AnchorPool()).MajorBlockChain().Get()
	if err != nil {
		return nil, errors.UnknownError.WithFormat("load major block chain: %w", err)
	}
	if chain.Height() == 0 {
		return nil, errors.NotFound.WithFormat("no major blocks exist")
	}

	if majorIndex == 0 { // We don't have major block 0, avoid crash
		majorIndex = 1
	}

	entryIndex, entry, err := indexing.SearchIndexChain(chain, majorIndex-1, indexing.MatchExact, indexing.SearchIndexChainByBlock(majorIndex))
	if err != nil {
		return nil, errors.UnknownError.WithFormat("get major block %d: %w", majorIndex, err)
	}

	rootEntry, rootPrev, err := getMajorBlockBounds(s.partition, batch, entry, entryIndex)
	if err != nil {
		return nil, errors.UnknownError.Wrap(err)
	}

	r := new(api.MajorBlockRecord)
	r.Index = majorIndex
	r.Time = *entry.BlockTime
	r.MinorBlocks = new(api.RecordRange[*api.MinorBlockRecord])
	r.MinorBlocks.Total = rootEntry.BlockIndex - rootPrev.BlockIndex
	allocRange(r.MinorBlocks, minorRange, zeroBased)

	r.MinorBlocks.Records, err = s.queryMinorBlockRange2(ctx, batch, r.MinorBlocks.Records, rootPrev.BlockIndex+1, rootEntry.BlockIndex, omitEmpty)
	if err != nil {
		return nil, errors.UnknownError.Wrap(err)
	}
	return r, nil
}

func (s *Querier) queryMajorBlockRange(ctx context.Context, batch *database.Batch, major *api.RangeOptions, omitEmpty bool) (*api.RecordRange[*api.MajorBlockRecord], error) {
	chain, err := batch.Account(s.partition.AnchorPool()).MajorBlockChain().Get()
	if err != nil {
		return nil, errors.UnknownError.WithFormat("load major block chain: %w", err)
	}
	if chain.Height() == 0 {
		return nil, errors.NotFound.WithFormat("no major blocks exist")
	}

	r := new(api.RecordRange[*api.MajorBlockRecord])
	{
		last := new(protocol.IndexEntry)
		err = chain.EntryAs(chain.Height()-1, last)
		if err != nil {
			return nil, errors.UnknownError.WithFormat("load last major block entry: %w", err)
		}
		r.Total = last.BlockIndex
	}
	allocRange(r, major, oneBased)

	if r.Start == 0 { // We don't have major block 0, avoid crash
		r.Start = 1
	}

	index, _, err := indexing.SearchIndexChain(chain, r.Start-1, indexing.MatchAfter, indexing.SearchIndexChainByBlock(r.Start))
	if err != nil {
		return nil, errors.UnknownError.WithFormat("get major block %d: %w", r.Start, err)
	}

	var i int
	nextBlock := r.Start
	for i < len(r.Records) && index < uint64(chain.Height()) {
		entry := new(protocol.IndexEntry)
		err = chain.EntryAs(int64(index), entry)
		if err != nil {
			return nil, errors.UnknownError.WithFormat("get major block chain entry %d: %w", index, err)
		}

		if !omitEmpty {
			// Add nil entries if there are missing blocks
			for nextBlock < entry.BlockIndex {
				i++
				nextBlock++
			}
			if i >= len(r.Records) {
				break
			}
		}

		rootEntry, rootPrev, err := getMajorBlockBounds(s.partition, batch, entry, index)
		if err != nil {
			return nil, errors.UnknownError.Wrap(err)
		}

		block := new(api.MajorBlockRecord)
		block.Index = entry.BlockIndex
		block.Time = *entry.BlockTime
		r.Records[i] = block

		// This query should only provide the _number_ of minor blocks in each
		// major block
		block.MinorBlocks = new(api.RecordRange[*api.MinorBlockRecord])
		block.MinorBlocks.Total = rootEntry.BlockIndex - rootPrev.BlockIndex

		index++
		i++
		nextBlock = entry.BlockIndex + 1
	}
	if i < len(r.Records) {
		r.Records = r.Records[:i]
	}
	return r, nil
}

func (s *Querier) queryMinorBlockRange2(ctx context.Context, batch *database.Batch, blocks []*api.MinorBlockRecord, blockIndex, maxIndex uint64, omitEmpty bool) ([]*api.MinorBlockRecord, error) {
	var i int
	var err error
	for i < len(blocks) && blockIndex <= maxIndex {
		blocks[i], err = s.queryMinorBlock(ctx, batch, blockIndex, nil)
		if err == nil {
			// Got a block
			i++
			blockIndex++
			continue
		}

		if errors.Is(err, errors.NotFound) {
			if !omitEmpty {
				blocks[i] = &api.MinorBlockRecord{Index: blockIndex}
				i++
			}
			blockIndex++
			continue
		}

		var e2 *errors.Error
		if !errors.As(err, &e2) || e2.Code.IsServerError() {
			return nil, errors.UnknownError.Wrap(err)
		}
	}
	if i < len(blocks) {
		blocks = blocks[:i]
	}
	return blocks, nil
}

func (s *Querier) searchForAnchor(ctx context.Context, batch *database.Batch, record *database.Account, hash []byte, wantReceipt bool) (*api.RecordRange[*api.ChainEntryRecord[api.Record]], error) {
	chains, err := record.Chains().Get()
	if err != nil {
		return nil, errors.UnknownError.WithFormat("load chains index: %w", err)
	}

	rr := new(api.RecordRange[*api.ChainEntryRecord[api.Record]])
	for _, c := range chains {
		if c.Type != merkle.ChainTypeAnchor {
			continue
		}

		chain, err := record.ChainByName(c.Name)
		if err != nil {
			return nil, errors.UnknownError.WithFormat("get chain %s: %w", c.Name, err)
		}

		r, err := s.queryChainEntryByValue(ctx, batch, chain, hash, true, wantReceipt)
		if err == nil {
			rr.Total++
			rr.Records = append(rr.Records, r)
		} else if !errors.Is(err, errors.NotFound) {
			return nil, errors.UnknownError.WithFormat("query chain entry: %w", err)
		}
	}
	if rr.Total == 0 {
		return nil, errors.NotFound.WithFormat("anchor %X not found", hash[:4])
	}
	return rr, nil
}

func (s *Querier) searchForKeyEntry(ctx context.Context, batch *database.Batch, scope *url.URL, search func(protocol.Signer) (int, protocol.KeyEntry, bool)) (*api.RecordRange[*api.KeyRecord], error) {
	auth, err := getAccountAuthoritySet(batch, scope)
	if err != nil {
		return nil, errors.UnknownError.WithFormat("load authority set: %w", err)
	}

	// For each local authority
	rs := new(api.RecordRange[*api.KeyRecord])
	for _, entry := range auth.Authorities {
		if !entry.Url.LocalTo(scope) {
			continue // Skip remote authorities
		}

		var authority protocol.Authority
		err = batch.Account(entry.Url).Main().GetAs(&authority)
		if err != nil {
			return nil, errors.UnknownError.WithFormat("load authority %v: %w", entry.Url, err)
		}

		// For each signer
		for _, signerUrl := range authority.GetSigners() {
			var signer protocol.Signer
			err = batch.Account(signerUrl).Main().GetAs(&signer)
			if err != nil {
				return nil, errors.UnknownError.WithFormat("load signer %v: %w", signerUrl, err)
			}

			// Check for a matching entry
			i, e, ok := search(signer)
			if !ok {
				continue
			}

			// Found it!
			rec := new(api.KeyRecord)
			rec.Authority = entry.Url
			rec.Signer = signerUrl
			rec.Version = signer.GetVersion()
			rec.Index = uint64(i)

			if ks, ok := e.(*protocol.KeySpec); ok {
				rec.Entry = ks
			} else {
				rec.Entry = new(protocol.KeySpec)
				rec.Entry.LastUsedOn = e.GetLastUsedOn()
			}
			rs.Records = append(rs.Records, rec)
		}
	}
	rs.Total = uint64(len(rs.Records))
	return rs, nil
}

func (s *Querier) searchForTransactionHash(ctx context.Context, batch *database.Batch, hash [32]byte) (*api.RecordRange[*api.TxIDRecord], error) {
	state, err := batch.Transaction(hash[:]).Main().Get()
	switch {
	case errors.Is(err, errors.NotFound):
		return new(api.RecordRange[*api.TxIDRecord]), nil

	case err != nil:
		return nil, errors.UnknownError.Wrap(err)

	case state.Transaction != nil:
		// TODO Replace with principal or signer as appropriate
		txid := s.partition.WithTxID(*(*[32]byte)(state.Transaction.GetHash()))
		r := new(api.RecordRange[*api.TxIDRecord])
		r.Total = 1
		r.Records = []*api.TxIDRecord{{Value: txid}}
		return r, nil

	case state.Signature != nil:
		// TODO Replace with principal or signer as appropriate
		txid := s.partition.WithTxID(*(*[32]byte)(state.Signature.Hash()))
		r := new(api.RecordRange[*api.TxIDRecord])
		r.Total = 1
		r.Records = []*api.TxIDRecord{{Value: txid}}
		return r, nil

	default:
		return new(api.RecordRange[*api.TxIDRecord]), nil
	}
}
