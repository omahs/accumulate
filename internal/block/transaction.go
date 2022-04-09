package block

import (
	"errors"
	"fmt"

	"gitlab.com/accumulatenetwork/accumulate/internal/chain"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/indexing"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	"gitlab.com/accumulatenetwork/accumulate/smt/storage"
)

// ProcessTransaction processes a transaction. It will not return an error if
// the transaction fails - in that case the status code will be non zero. It
// only returns an error in cases like a database failure.
func (x *Executor) ProcessTransaction(batch *database.Batch, transaction *protocol.Transaction) (*protocol.TransactionStatus, *chain.ProcessTransactionState, error) {
	// Load the status
	status, err := batch.Transaction(transaction.GetHash()).GetStatus()
	if err != nil {
		return nil, nil, err
	}
	if status.Initiator == nil {
		// This should never happen
		return nil, nil, fmt.Errorf("transaction initiator is missing")
	}

	// Load the principal
	principal, err := batch.Account(transaction.Header.Principal).GetState()
	switch {
	case err == nil:
		// Ok
	case !errors.Is(err, storage.ErrNotFound):
		err = fmt.Errorf("load principal: %w", err)
		return recordFailedTransaction(batch, transaction, err)
	case !transactionAllowsMissingPrincipal(transaction):
		err = fmt.Errorf("load principal: %w", err)
		return recordFailedTransaction(batch, transaction, err)
	}

	// Check if the transaction is ready to be executed
	ready, err := TransactionIsReady(batch, transaction, status)
	if err != nil {
		return recordFailedTransaction(batch, transaction, err)
	}
	if !ready {
		return recordPendingTransaction(batch, transaction)
	}

	if transaction.Body.Type().IsSynthetic() {
		// Verify that the synthetic transaction has all the right signatures
		err = processSyntheticTransaction(&x.Network, batch, transaction, status)
		if err != nil {
			return recordFailedTransaction(batch, transaction, err)
		}
	}

	// Set up the state manager
	st, err := chain.LoadStateManager(batch.Begin(true), x.Network.NodeUrl(), principal, transaction, status, x.logger.With("operation", "ProcessTransaction"))
	if err != nil {
		return recordFailedTransaction(batch, transaction, err)
	}
	defer st.Discard()

	// Execute the transaction
	executor, ok := x.executors[transaction.Body.Type()]
	if !ok {
		// An invalid transaction should not make it to this point
		err = protocol.Errorf(protocol.ErrorCodeInternal, "missing executor for %v", transaction.Body.Type())
		return recordFailedTransaction(batch, transaction, err)
	}

	result, err := executor.Execute(st, &chain.Delivery{Transaction: transaction})
	if err != nil {
		return recordFailedTransaction(batch, transaction, err)
	}

	// Commit changes, queue state creates for synthetic transactions
	state, err := st.Commit()
	if err != nil {
		err = fmt.Errorf("commit: %w", err)
		return recordFailedTransaction(batch, transaction, err)
	}

	return recordSuccessfulTransaction(batch, state, transaction, result)
}

func transactionAllowsMissingPrincipal(transaction *protocol.Transaction) bool {
	switch transaction.Body.Type() {
	case protocol.TransactionTypeSyntheticCreateChain,
		protocol.TransactionTypeSyntheticDepositTokens,
		protocol.TransactionTypeSyntheticWriteData:
		// These transactions allow for a missing origin
		return true
	default:
		return false
	}
}

func TransactionIsReady(batch *database.Batch, transaction *protocol.Transaction, status *protocol.TransactionStatus) (bool, error) {
	// TODO When we implement 'pending' synthetic transactions, this is where
	// the logic will go

	if !transaction.Body.Type().IsUser() {
		return true, nil
	}

	// UpdateKey transactions are always M=1 and always require a signature from
	// the initiator
	txnObj := batch.Transaction(transaction.GetHash())
	if transaction.Body.Type() == protocol.TransactionTypeUpdateKey {
		if status.Initiator == nil {
			return false, fmt.Errorf("missing initiator")
		}

		initSigs, err := txnObj.ReadSignatures(status.Initiator)
		if err != nil {
			return false, fmt.Errorf("load initiator signatures: %w", err)
		}

		if initSigs.Count() == 0 {
			return false, fmt.Errorf("missing initiator signature")
		}

		return true, nil
	}

	// Load the principal
	principal, err := batch.Account(transaction.Header.Principal).GetState()
	if err != nil {
		return false, fmt.Errorf("load principal: %w", err)
	}

	// Get the principal's account auth
	auth, err := getAccountAuth(batch, principal)
	if err != nil {
		return false, fmt.Errorf("unable to load authority of %v: %w", transaction.Header.Principal, err)
	}

	// For each authority
outer:
	for _, entry := range auth.Authorities {
		// Do not check signers for disabled authorities
		if entry.Disabled {
			continue
		}

		// Check if any signer has reached its threshold
		for _, signer := range status.FindSigners(entry.Url) {
			// Load the signature set
			signatures, err := txnObj.ReadSignaturesForSigner(signer)
			if err != nil {
				return false, fmt.Errorf("load signatures set %v: %w", signer.GetUrl(), err)
			}

			// Check if the threshold has been reached
			if uint64(signatures.Count()) >= signer.GetSignatureThreshold() {
				continue outer
			}
		}

		return false, nil
	}

	// If every authority is disabled, at least one signature is required
	return len(status.Signers) > 0, nil
}

func recordTransaction(batch *database.Batch, transaction *protocol.Transaction, updateStatus func(*protocol.TransactionStatus)) (*protocol.TransactionStatus, error) {
	// Store the transaction state (without signatures)
	db := batch.Transaction(transaction.GetHash())
	err := db.PutState(&database.SigOrTxn{Transaction: transaction})
	if err != nil {
		return nil, fmt.Errorf("store transaction: %w", err)
	}

	// Update the status
	status, err := db.GetStatus()
	if err != nil {
		return nil, fmt.Errorf("load transaction status: %w", err)
	}

	updateStatus(status)
	err = db.PutStatus(status)
	if err != nil {
		return nil, fmt.Errorf("store transaction status: %w", err)
	}

	return status, nil
}

func recordPendingTransaction(batch *database.Batch, transaction *protocol.Transaction) (*protocol.TransactionStatus, *chain.ProcessTransactionState, error) {
	// Record the transaction
	status, err := recordTransaction(batch, transaction, func(status *protocol.TransactionStatus) {
		status.Remote = false
		status.Pending = true
	})
	if err != nil {
		return nil, nil, err
	}

	// Add the transaction to the principal's list of pending transactions
	pending := indexing.PendingTransactions(batch, transaction.Header.Principal)
	err = pending.Add(*(*[32]byte)(transaction.GetHash()))
	if err != nil {
		return nil, nil, fmt.Errorf("store pending list: %w", err)
	}

	return status, new(chain.ProcessTransactionState), nil
}

func recordSuccessfulTransaction(batch *database.Batch, state *chain.ProcessTransactionState, transaction *protocol.Transaction, result protocol.TransactionResult) (*protocol.TransactionStatus, *chain.ProcessTransactionState, error) {
	// Record the transaction
	status, err := recordTransaction(batch, transaction, func(status *protocol.TransactionStatus) {
		status.Remote = false
		status.Pending = false
		status.Delivered = true
		status.Code = 0
		if result == nil {
			status.Result = new(protocol.EmptyResult)
		} else {
			status.Result = result
		}
	})
	if err != nil {
		return nil, nil, err
	}

	// Create receipt
	if chain.NeedsReceipt(transaction.Body.Type()) {
		state.DidProduceTxn(chain.CreateSynthReceipt(transaction, status))
	}

	// Don't add internal transactions to chains
	if transaction.Body.Type().IsInternal() {
		return status, state, nil
	}

	// Remove the transaction from the principal's list of pending transactions
	pending := indexing.PendingTransactions(batch, transaction.Header.Principal)
	err = pending.Remove(*(*[32]byte)(transaction.GetHash()))
	if err != nil {
		return nil, nil, fmt.Errorf("store pending list: %w", err)
	}

	// Add the transaction to the principal's main chain
	err = state.ChainUpdates.AddChainEntry(batch, transaction.Header.Principal, protocol.MainChain, protocol.ChainTypeTransaction, transaction.GetHash(), 0, 0)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return nil, nil, fmt.Errorf("add to chain: %v", err)
	}

	return status, state, nil
}

func recordFailedTransaction(batch *database.Batch, transaction *protocol.Transaction, failure error) (*protocol.TransactionStatus, *chain.ProcessTransactionState, error) {
	// Record the transaction
	status, err := recordTransaction(batch, transaction, func(status *protocol.TransactionStatus) {
		failure := protocol.NewError(protocol.ErrorCodeUnknownError, failure)
		status.Remote = false
		status.Delivered = true
		status.Code = failure.Code.GetEnumValue()
		status.Message = failure.Error()
	})
	if err != nil {
		return nil, nil, err
	}

	// Create receipt
	state := new(chain.ProcessTransactionState)
	if chain.NeedsReceipt(transaction.Body.Type()) {
		state.DidProduceTxn(chain.CreateSynthReceipt(transaction, status))
	}

	// Remove the transaction from the principal's list of pending transactions
	pending := indexing.PendingTransactions(batch, transaction.Header.Principal)
	err = pending.Remove(*(*[32]byte)(transaction.GetHash()))
	if err != nil {
		return nil, nil, fmt.Errorf("update pending list: %w", err)
	}

	// Refund the signer
	if status.Initiator == nil || !transaction.Body.Type().IsUser() {
		return status, state, nil
	}

	// TODO Send a refund for a failed remotely initiated transaction
	if !transaction.Header.Principal.LocalTo(status.Initiator) {
		return status, state, nil
	}

	// But only if the paid paid is larger than the max failure paid
	paid, err := protocol.ComputeTransactionFee(transaction)
	if err != nil {
		return nil, nil, fmt.Errorf("compute fee: %w", err)
	}
	if paid <= protocol.FeeFailedMaximum {
		return status, state, nil
	}

	var signer protocol.Signer
	obj := batch.Account(status.Initiator)
	err = obj.GetStateAs(&signer)
	if err != nil {
		return nil, nil, fmt.Errorf("load initial signer: %w", err)
	}

	refund := paid - protocol.FeeFailedMaximum
	signer.CreditCredits(refund.AsUInt64())
	err = obj.PutState(signer)
	if err != nil {
		return nil, nil, fmt.Errorf("store initial signer: %w", err)
	}

	return status, state, nil
}
