package chain

import (
	"fmt"

	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

type InternalSendTransactions struct{}

func (InternalSendTransactions) Type() protocol.TransactionType {
	return protocol.TransactionTypeInternalSendTransactions
}

func (InternalSendTransactions) Execute(st *StateManager, tx *Delivery) (protocol.TransactionResult, error) {
	return (InternalSendTransactions{}).Validate(st, tx)
}

func (InternalSendTransactions) Validate(st *StateManager, tx *Delivery) (protocol.TransactionResult, error) {
	body, ok := tx.Transaction.Body.(*protocol.InternalSendTransactions)
	if !ok {
		return nil, fmt.Errorf("invalid payload: want %T, got %T", new(protocol.InternalSendTransactions), tx.Transaction.Body)
	}

	for _, tx := range body.Transactions {
		st.Submit(tx.Recipient, tx.Payload)
		st.logger.Debug("Submitting transaction",
			"principal", tx.Recipient,
			"type", tx.Payload.Type(),
			"module", "governor")
	}

	return nil, nil
}
