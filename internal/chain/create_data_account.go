package chain

import (
	"fmt"

	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

type CreateDataAccount struct{}

func (CreateDataAccount) Type() protocol.TransactionType {
	return protocol.TransactionTypeCreateDataAccount
}

func (CreateDataAccount) Execute(st *StateManager, tx *protocol.Envelope) (protocol.TransactionResult, error) {
	return (CreateDataAccount{}).Validate(st, tx)
}

func (CreateDataAccount) Validate(st *StateManager, tx *protocol.Envelope) (protocol.TransactionResult, error) {
	body, ok := tx.Transaction.Body.(*protocol.CreateDataAccount)
	if !ok {
		return nil, fmt.Errorf("invalid payload: want %T, got %T", new(protocol.CreateDataAccount), tx.Transaction.Body)
	}

	//only the ADI can create the data account associated with the ADI
	if !body.Url.Identity().Equal(st.OriginUrl) {
		return nil, fmt.Errorf("%q cannot be the origininator of %q", st.OriginUrl, body.Url)
	}

	//create the data account
	account := protocol.NewDataAccount()
	account.Url = body.Url
	account.Scratch = body.Scratch
	account.ManagerKeyBook = body.ManagerKeyBookUrl

	err := st.setKeyBook(account, body.KeyBookUrl)
	if err != nil {
		return nil, err
	}

	st.Create(account)
	return nil, nil
}
