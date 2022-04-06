package chain

import (
	"errors"
	"fmt"

	"gitlab.com/accumulatenetwork/accumulate/internal/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	"gitlab.com/accumulatenetwork/accumulate/smt/storage"
)

type SyntheticCreateChain struct{}

func (SyntheticCreateChain) Type() protocol.TransactionType {
	return protocol.TransactionTypeSyntheticCreateChain
}

func (SyntheticCreateChain) Execute(st *StateManager, tx *protocol.Envelope) (protocol.TransactionResult, error) {
	return (SyntheticCreateChain{}).Validate(st, tx)
}

func (SyntheticCreateChain) Validate(st *StateManager, tx *protocol.Envelope) (protocol.TransactionResult, error) {
	body, ok := tx.Transaction.Body.(*protocol.SyntheticCreateChain)
	if !ok {
		return nil, fmt.Errorf("invalid payload: want %T, got %T", new(protocol.SyntheticCreateChain), tx.Transaction.Body)
	}

	if body.Cause == [32]byte{} {
		return nil, fmt.Errorf("cause is missing")
	}

	// Do basic validation and add everything to the state manager
	urls := make([]*url.URL, len(body.Chains))
	for i, cc := range body.Chains {
		record, err := protocol.UnmarshalAccount(cc.Data)
		if err != nil {
			return nil, fmt.Errorf("invalid chain payload: %v", err)
		}

		u := record.GetUrl()
		_, err = st.LoadUrl(u)
		switch {
		case err != nil && !errors.Is(err, storage.ErrNotFound):
			// Unknown error
			return nil, fmt.Errorf("error fetching %q: %v", u, err)

		case cc.IsUpdate && err != nil:
			// Attempted to update but the record does not exist
			return nil, fmt.Errorf("cannot update %q: does not exist", u)

		case !cc.IsUpdate && err == nil:
			// Attempted to create but the record already exists
			return nil, fmt.Errorf("cannot create %q: already exists", u)

		case !cc.IsUpdate:
			// Creating a record, add it to the directory
			err = st.AddDirectoryEntry(u.Identity(), u)
			if err != nil {
				return nil, fmt.Errorf("failed to add a directory entry for %q: %v", u, err)
			}
		}

		urls[i] = u
		st.Update(record)
	}

	// Verify everything is sane
	for _, u := range urls {
		record, err := st.LoadUrl(u)
		if err != nil {
			// This really shouldn't happen, but don't panic
			return nil, fmt.Errorf("internal error: failed to fetch pending record")
		}

		// Check the identity
		switch record.Type() {
		case protocol.AccountTypeIdentity:
		default:
			// Make sure the ADI actually exists
			_, err = st.LoadUrl(u.Identity())
			if errors.Is(err, storage.ErrNotFound) {
				return nil, fmt.Errorf("missing identity for %s", u.String())
			} else if err != nil {
				return nil, fmt.Errorf("error fetching %q: %v", u.String(), err)
			}

		}

		// Check the key book
		fullAccount, ok := record.(protocol.FullAccount)
		if !ok {
			continue
		}

		auth := fullAccount.GetAuth()
		if len(auth.Authorities) == 0 {
			return nil, fmt.Errorf("%q does not specify a key book", u)
		}

		// Make sure the key book actually exists
		for _, auth := range auth.Authorities {
			var book *protocol.KeyBook
			err = st.LoadUrlAs(auth.Url, &book)
			if err != nil {
				return nil, fmt.Errorf("invalid key book %q for %q: %v", auth.Url, u, err)
			}
		}
	}

	return nil, nil
}
