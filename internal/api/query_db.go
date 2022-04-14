package api

import (
	"strings"

	"gitlab.com/accumulatenetwork/accumulate/config"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/errors"
	"gitlab.com/accumulatenetwork/accumulate/internal/indexing"
	"gitlab.com/accumulatenetwork/accumulate/internal/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

type DatabaseQueryModule struct {
	Network *config.Network
	DB      *database.Database
}

var _ QueryModule = (*DatabaseQueryModule)(nil)

func (q *DatabaseQueryModule) QueryState(account *url.URL, fragment []string, opts QueryStateOptions) (Record, error) {
	batch := q.DB.Begin(false)
	defer batch.Discard()

	if len(fragment) > 0 {
		return nil, errors.Format(errors.StatusBadRequest, "unsupported fragment query %q", strings.Join(fragment, "/"))
	}

	return q.queryAccount(batch, account, opts)
}

func (q *DatabaseQueryModule) QuerySet(account *url.URL, fragment []string, opts QuerySetOptions) (Record, error) {
	return nil, errors.Format(errors.StatusBadRequest, "unsupported fragment query %q", strings.Join(fragment, "/"))
}

func (q *DatabaseQueryModule) Search(scope *url.URL, query string, opts SearchOptions) (Record, error) {
	if opts.Kind == "" {
		return nil, errors.Format(errors.StatusBadRequest, "missing option `kind`")
	}

	return nil, errors.Format(errors.StatusBadRequest, "unsupported search kind %q", opts.Kind)
}

func (q *DatabaseQueryModule) queryAccount(batch *database.Batch, accountUrl *url.URL, opts QueryStateOptions) (Record, error) {
	account := batch.Account(accountUrl)
	rec := new(AccountRecord)
	var err error
	rec.Account, err = account.GetState()
	if err != nil {
		return nil, errors.Format(errors.StatusUnknown, "get account %v main state: %w", accountUrl, err)
	}

	if opts.Expand {
		obj, err := account.GetObject()
		if err != nil {
			return nil, errors.Format(errors.StatusUnknown, "get account %v state: %w", accountUrl, err)
		}

		for _, c := range obj.Chains {
			chain, err := account.ReadChain(c.Name)
			if err != nil {
				return nil, errors.Format(errors.StatusUnknown, "read account %v chain %s: %w", accountUrl, c.Name, err)
			}

			state := new(ChainState)
			state.Name = c.Name
			state.Type = c.Type
			state.Height = uint64(chain.Height())
			for _, hash := range chain.CurrentState().Pending {
				state.Roots = append(state.Roots, hash)
			}
			rec.Chains = append(rec.Chains, state)
		}
	}

	if opts.Prove {
		receipt := new(Receipt)
		rec.Receipt = receipt
		block, mr, err := indexing.ReceiptForAccountState(q.Network, batch, account)
		if err != nil {
			receipt.Error = errors.Wrap(errors.StatusUnknown, err)
		} else {
			receipt.LocalBlock = block
			receipt.Receipt = *protocol.ReceiptFromManaged(mr)
		}
	}

	return rec, nil
}
