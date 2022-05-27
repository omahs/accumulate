package chain

import (
	"bytes"
	"fmt"

	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/errors"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

type WriteData struct{}

var _ SignerValidator = (*WriteData)(nil)

func (WriteData) Type() protocol.TransactionType { return protocol.TransactionTypeWriteData }

func isWriteToLiteDataAccount(batch *database.Batch, transaction *protocol.Transaction) (bool, error) {
	body, ok := transaction.Body.(*protocol.WriteData)
	if !ok {
		return false, errors.Format(errors.StatusInternalError, "invalid payload: want %T, got %T", new(protocol.WriteData), transaction.Body)
	}

	chainId, err := protocol.ParseLiteDataAddress(transaction.Header.Principal)
	if err != nil {
		return false, nil // Not a lite data address
	}

	account, err := batch.Account(transaction.Header.Principal).GetState()
	switch {
	case err == nil:
		// Found the account, is it a lite data account?
		_, ok := account.(*protocol.LiteDataAccount)
		return ok, nil

	case errors.Is(err, errors.StatusNotFound):
		// Are we creating a lite data account?
		computedChainId := protocol.ComputeLiteDataAccountId(body.Entry)
		return bytes.HasPrefix(computedChainId, chainId), nil

	default:
		// Unknown error
		return false, errors.Wrap(errors.StatusUnknown, err)
	}
}

// SignerIsAuthorized returns nil if the transaction is writing to a lite data
// account.
func (WriteData) SignerIsAuthorized(_ AuthDelegate, batch *database.Batch, transaction *protocol.Transaction, _ protocol.Signer, _ bool) (fallback bool, err error) {
	lite, err := isWriteToLiteDataAccount(batch, transaction)
	if err != nil {
		return false, errors.Wrap(errors.StatusUnknown, err)
	}

	return !lite, nil
}

// TransactionIsReady returns true if the transaction is writing to a lite data
// account.
func (WriteData) TransactionIsReady(_ AuthDelegate, batch *database.Batch, transaction *protocol.Transaction, status *protocol.TransactionStatus) (ready, fallback bool, err error) {
	lite, err := isWriteToLiteDataAccount(batch, transaction)
	if err != nil {
		return false, false, errors.Wrap(errors.StatusUnknown, err)
	}

	// Writing to a lite data account only requires one signature
	if lite {
		return len(status.Signers) > 0, false, nil
	}

	// Fallback to general authorization
	return false, true, nil
}

func (WriteData) Execute(st *StateManager, tx *Delivery) (protocol.TransactionResult, error) {
	return (WriteData{}).Validate(st, tx)
}

func (WriteData) Validate(st *StateManager, tx *Delivery) (protocol.TransactionResult, error) {
	body, ok := tx.Transaction.Body.(*protocol.WriteData)
	if !ok {
		return nil, fmt.Errorf("invalid payload: want %T, got %T", new(protocol.WriteData), tx.Transaction.Body)
	}

	if _, ok := body.Entry.(*protocol.FactomDataEntry); ok {
		return nil, fmt.Errorf("writing new Factom-formatted data entries is not supported")
	}

	//check will return error if there is too much data or no data for the entry
	_, err := protocol.CheckDataEntrySize(body.Entry)
	if err != nil {
		return nil, err
	}

	_, err = protocol.ParseLiteDataAddress(st.OriginUrl)
	if err == nil {
		return executeWriteLiteDataAccount(st, body.Entry, body.Scratch)
	}

	return executeWriteFullDataAccount(st, body.Entry, body.Scratch)
}

func executeWriteFullDataAccount(st *StateManager, entry protocol.DataEntry, scratch bool) (protocol.TransactionResult, error) {
	if st.Origin == nil {
		return nil, errors.NotFound("%v not found", st.OriginUrl)
	}
	account, ok := st.Origin.(*protocol.DataAccount)
	if !ok {
		return nil, fmt.Errorf("invalid principal: want %v, got %v",
			protocol.AccountTypeDataAccount, st.Origin.Type())
	}

	if scratch && !account.Scratch {
		return nil, fmt.Errorf("cannot write scratch data to a non-scratch account")
	}

	result := new(protocol.WriteDataResult)
	result.EntryHash = *(*[32]byte)(entry.Hash())
	result.AccountID = st.OriginUrl.AccountID()
	result.AccountUrl = st.OriginUrl
	st.UpdateData(st.Origin, result.EntryHash[:], entry)
	return result, nil
}
