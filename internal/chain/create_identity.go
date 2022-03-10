package chain

import (
	"fmt"

	"gitlab.com/accumulatenetwork/accumulate/internal/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	"gitlab.com/accumulatenetwork/accumulate/types/api/transactions"
	"gitlab.com/accumulatenetwork/accumulate/types/state"
)

type CreateIdentity struct{}

func (CreateIdentity) Type() protocol.TransactionType { return protocol.TransactionTypeCreateIdentity }

func (ci CreateIdentity) Validate(st *StateManager, tx *transactions.Envelope) (protocol.TransactionResult, error) {
	// *protocol.IdentityCreate, *url.URL, state.Chain
	body, ok := tx.Transaction.Body.(*protocol.CreateIdentity)
	if !ok {
		return nil, fmt.Errorf("invalid payload: want %T, got %T", new(protocol.CreateIdentity), tx.Transaction.Body)
	}

	err := validateAdiUrl(body, st.Origin)
	if err != nil {
		return nil, err
	}

	bookUrl := selectBookUrl(body)
	err = validateKeyBookUrl(bookUrl, body.Url)
	if err != nil {
		return nil, err
	}

	identity := protocol.NewADI()
	identity.Url = body.Url
	identity.KeyBook = bookUrl
	identity.ManagerKeyBook = body.Manager

	accounts := []protocol.Account{identity}
	book := protocol.NewKeyBook()
	bookExists := st.LoadUrlAs(bookUrl, book) == nil
	if !bookExists {
		if len(body.PublicKey) == 0 {
			return nil, fmt.Errorf("missing PublicKey which is required when creating a new KeyBook/KeyPage pair")
		}
		book.Url = bookUrl
		book.PageCount = 1
		accounts = append(accounts, book)

		page := protocol.NewKeyPage()
		page.KeyBook = bookUrl
		page.Url = protocol.FormatKeyPageUrl(bookUrl, 0)
		page.Threshold = 1 // Require one signature from the Key Page
		keySpec := new(protocol.KeySpec)
		keySpec.PublicKey = body.PublicKey
		page.Keys = append(page.Keys, keySpec)
		accounts = append(accounts, page)
	}

	st.Create(accounts...)
	return nil, nil
}

func validateAdiUrl(body *protocol.CreateIdentity, origin state.Chain) error {
	err := protocol.IsValidAdiUrl(body.Url)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	switch v := origin.(type) {
	case *protocol.LiteTokenAccount:
	// OK
	case *protocol.ADI:
		if len(body.Url.Path) > 0 {
			parent, _ := body.Url.Parent()
			if !parent.Equal(v.Url) {
				return fmt.Errorf("a sub ADI %s must be a direct child of its origin ADI %s", body.Url.String(), v.Url.String())
			}
		}
	default:
		return fmt.Errorf("account type %d cannot be the origininator of ADIs", origin.GetType())
	}

	return nil
}

func selectBookUrl(body *protocol.CreateIdentity) *url.URL {
	if body.KeyBookUrl == nil {
		return body.Url.JoinPath(protocol.DefaultKeyBook)
	}
	return body.KeyBookUrl
}

func validateKeyBookUrl(bookUrl *url.URL, adiUrl *url.URL) error {
	err := protocol.IsValidAdiUrl(bookUrl)
	if err != nil {
		return fmt.Errorf("invalid KeyBook URL %s: %v", bookUrl.String(), err)
	}
	parent, err := bookUrl.Parent()
	if err != nil {
		return fmt.Errorf("invalid KeyBook URL: %v", err)
	}
	if !parent.Equal(adiUrl) {
		return fmt.Errorf("KeyBook %s must be a direct child of its ADI %s", bookUrl.String(), adiUrl.String())
	}
	return nil
}

func validateKeyPageUrl(pageUrl *url.URL, bookUrl *url.URL) error {
	kpParentUrl, err := pageUrl.Parent()
	if err != nil {
		return fmt.Errorf("invalid KeyPage URL: %w\nthe KeyPage URL must be adi_path/KeyPage", err)
	}

	bkParentUrl, _ := bookUrl.Parent()
	if !bkParentUrl.Equal(kpParentUrl) {
		return fmt.Errorf("KeyPage %s must be in the same path as its KeyBook %s", pageUrl, bookUrl)
	}

	return nil
}
