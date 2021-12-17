package chain

import (
	"fmt"

	"github.com/AccumulateNetwork/accumulate/internal/url"
	"github.com/AccumulateNetwork/accumulate/protocol"
	"github.com/AccumulateNetwork/accumulate/types"
	"github.com/AccumulateNetwork/accumulate/types/api/transactions"
	"github.com/AccumulateNetwork/accumulate/types/state"
)

type CreateIdentity struct{}

func (CreateIdentity) Type() types.TxType { return types.TxTypeCreateIdentity }

func (CreateIdentity) Validate(st *StateManager, tx *transactions.GenTransaction) error {
	// *protocol.IdentityCreate, *url.URL, state.Chain
	body := new(protocol.IdentityCreate)
	err := tx.As(body)
	if err != nil {
		return fmt.Errorf("invalid payload: %v", err)
	}

	identityUrl, err := url.Parse(body.Url)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	err = protocol.IsValidAdiUrl(identityUrl)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	switch st.Origin.(type) {
	case *protocol.LiteTokenAccount, *state.AdiState:
		// OK
	default:
		return fmt.Errorf("chain type %d cannot be the origininator of ADIs", st.Origin.Header().Type)
	}

	var pageUrl, bookUrl *url.URL
	if body.KeyPageName == "" {
		pageUrl = identityUrl.JoinPath("sigspec0")
	} else {
		pageUrl = identityUrl.JoinPath(body.KeyPageName)
	}
	if body.KeyBookName == "" {
		bookUrl = identityUrl.JoinPath("ssg0")
	} else {
		bookUrl = identityUrl.JoinPath(body.KeyBookName)
	}

	keySpec := new(protocol.KeySpec)
	keySpec.PublicKey = body.PublicKey

	page := protocol.NewKeyPage()
	page.ChainUrl = types.String(pageUrl.String()) // TODO Allow override
	page.Keys = append(page.Keys, keySpec)
	page.KeyBook = types.Bytes(bookUrl.ResourceChain()).AsBytes32()

	book := protocol.NewKeyBook()
	book.ChainUrl = types.String(bookUrl.String()) // TODO Allow override
	book.Pages = append(book.Pages, types.Bytes(pageUrl.ResourceChain()).AsBytes32())

	identity := state.NewADI(types.String(identityUrl.String()), state.KeyTypeSha256, body.PublicKey)
	identity.KeyBook = types.Bytes(bookUrl.ResourceChain()).AsBytes32()

	st.Create(identity, book, page)
	return nil
}
