// Copyright 2023 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package block

import (
	"bytes"

	"gitlab.com/accumulatenetwork/accumulate/internal/core/execute/v2/chain"
	"gitlab.com/accumulatenetwork/accumulate/internal/core/execute/v2/internal"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/logging"
	"gitlab.com/accumulatenetwork/accumulate/pkg/errors"
	"gitlab.com/accumulatenetwork/accumulate/pkg/types/messaging"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

// bundle is a bundle of messages to be processed.
type bundle struct {
	*BlockV2

	// messages is the bundle of messages.
	messages []messaging.Message

	// additional is an additional bundle of messages that should be processed
	// after this one.
	additional []messaging.Message

	// transactionsToProcess is a list of transactions that have been signed and
	// thus should be processed. It will be removed once authority signatures
	// are implemented. MUST BE ORDERED.
	transactionsToProcess hashSet

	// internal tracks which messages were produced internally (e.g. network
	// account updates).
	internal set[[32]byte]

	// forwarded tracks which messages were forwarded within a forwarding
	// message.
	forwarded set[[32]byte]

	// state tracks transaction state objects.
	state orderedMap[[32]byte, *chain.ProcessTransactionState]
}

// Process processes a message bundle.
func (b *BlockV2) Process(messages []messaging.Message) ([]*protocol.TransactionStatus, error) {
	var statuses []*protocol.TransactionStatus

	// Make sure every transaction is signed
	err := b.checkForUnsignedTransactions(messages)
	if err != nil {
		return nil, errors.UnknownError.Wrap(err)
	}

	// Do not check for unsigned transactions when processing additional
	// messages
	additional := false
additional:

	// Do this now for the sake of comparing logs
	for _, msg := range messages {
		if fwd, ok := msg.(*internal.ForwardedMessage); ok {
			msg = fwd.Message
		}

		msg, ok := msg.(*messaging.UserTransaction)
		if !ok {
			continue
		}

		fn := b.Executor.logger.Debug
		kv := []interface{}{
			"block", b.Block.Index,
			"type", msg.Transaction.Body.Type(),
			"txn-hash", logging.AsHex(msg.Transaction.GetHash()).Slice(0, 4),
			"principal", msg.Transaction.Header.Principal,
		}
		switch msg.Transaction.Body.Type() {
		case protocol.TransactionTypeDirectoryAnchor,
			protocol.TransactionTypeBlockValidatorAnchor:
			fn = b.Executor.logger.Info
			kv = append(kv, "module", "anchoring")
		}
		if additional {
			fn("Executing additional", kv...)
		} else {
			fn("Executing transaction", kv...)
		}
	}

	d := new(bundle)
	d.BlockV2 = b
	d.messages = messages
	d.state = orderedMap[[32]byte, *chain.ProcessTransactionState]{cmp: func(u, v [32]byte) int { return bytes.Compare(u[:], v[:]) }}
	d.transactionsToProcess = hashSet{}
	d.internal = set[[32]byte]{}
	d.forwarded = set[[32]byte]{}

	// Process each message
	remote := set[[32]byte]{}
	for _, msg := range messages {
		// Unpack forwarded messages and mark them as forwarded
		if fwd, ok := msg.(*internal.ForwardedMessage); ok {
			d.forwarded.Add(msg.ID().Hash())
			msg = fwd.Message
		}

		var st *protocol.TransactionStatus
		var err error
		switch msg := msg.(type) {
		case *messaging.UserTransaction:
			remote.Add(msg.ID().Hash())
			st, err = d.processUserTransactionMessage(msg)
		case *messaging.UserSignature:
			st, err = d.processUserSignatureMessage(msg)
		case *internal.NetworkUpdate:
			st, err = d.processNetworkUpdateMessage(msg)
		case *internal.SyntheticMessage:
			d.transactionsToProcess.Add(msg.TxID.Hash())
			continue
		default:
			st = protocol.NewErrorStatus(msg.ID(), errors.BadRequest.WithFormat("unsupported message type %v", msg.Type()))
		}
		if err != nil {
			return nil, errors.UnknownError.Wrap(err)
		}

		// Some messages may not produce a status at this stage
		if st != nil {
			statuses = append(statuses, st)
		}
	}

	// Process transactions (MUST BE ORDERED)
	for _, hash := range d.transactionsToProcess {
		remote.Remove(hash)
		st, err := d.executeTransaction(hash, additional)
		if err != nil {
			return nil, errors.UnknownError.Wrap(err)
		}
		if st != nil {
			statuses = append(statuses, st)
		}
	}

	// Record remote transactions (remove?)
	for hash := range remote {
		hash := hash // See docs/developer/rangevarref.md
		record := b.Block.Batch.Transaction(hash[:])
		st, err := record.Status().Get()
		if err != nil {
			return nil, errors.UnknownError.WithFormat("load transaction status: %w", err)
		}
		if st.Code != 0 {
			continue
		}

		st.Code = errors.Remote
		err = record.Status().Put(st)
		if err != nil {
			return nil, errors.UnknownError.WithFormat("store transaction status: %w", err)
		}
	}

	// Produce remote signatures
	err = d.ProcessRemoteSignatures()
	if err != nil {
		return nil, errors.UnknownError.Wrap(err)
	}

	// Process synthetic transactions generated by the validator (MUST BE ORDERED)
	{
		batch := b.Block.Batch.Begin(true)
		defer batch.Discard()

		err := d.state.For(func(hash [32]byte, state *chain.ProcessTransactionState) error {
			// Load the transaction. Earlier checks should guarantee this never fails.
			txn, err := batch.Transaction(hash[:]).Main().Get()
			switch {
			case err != nil:
				return errors.InternalError.WithFormat("load transaction: %w", err)
			case txn.Transaction == nil:
				return errors.InternalError.WithFormat("%x is not a transaction", hash)
			}

			err = b.Executor.ProduceSynthetic(batch, txn.Transaction, state.ProducedTxns)
			return errors.UnknownError.Wrap(err)
		})
		if err != nil {
			return nil, errors.UnknownError.Wrap(err)
		}

		err = batch.Commit()
		if err != nil {
			return nil, errors.UnknownError.WithFormat("commit batch: %w", err)
		}
	}

	// Update the block state (MUST BE ORDERED)
	_ = d.state.For(func(_ [32]byte, state *chain.ProcessTransactionState) error {
		b.Block.State.MergeTransaction(state)
		return nil
	})

	// Process additional transactions. It would be simpler to do this
	// recursively, but it's possible that could cause a stack overflow.
	if len(d.additional) > 0 {
		additional = true
		messages = d.additional
		goto additional
	}

	return statuses, nil
}

// checkForUnsignedTransactions returns an error if the message bundle includes
// any unsigned transactions.
func (b *BlockV2) checkForUnsignedTransactions(messages []messaging.Message) error {
	unsigned := set[[32]byte]{}
	for _, msg := range messages {
		if msg, ok := msg.(*messaging.UserTransaction); ok {
			unsigned[msg.ID().Hash()] = struct{}{}
		}
	}
	for _, msg := range messages {
		if msg, ok := msg.(*messaging.UserSignature); ok {
			delete(unsigned, msg.TransactionHash)
		}
	}
	if len(unsigned) > 0 {
		return errors.BadRequest.With("message bundle includes an unsigned transaction")
	}
	return nil
}

// processUserTransactionMessage records the transaction but does not execute
// it. Transactions are executed in response to _authority signature_ messages,
// not user transaction messages.
func (b *bundle) processUserTransactionMessage(txn *messaging.UserTransaction) (*protocol.TransactionStatus, error) {
	// Ensure the transaction is signed
	var signed bool
	for _, other := range b.messages {
		if fwd, ok := other.(*internal.ForwardedMessage); ok {
			other = fwd.Message
		}
		sig, ok := other.(*messaging.UserSignature)
		if ok && sig.TransactionHash == txn.ID().Hash() {
			signed = true
			break
		}
	}
	if !signed {
		return protocol.NewErrorStatus(txn.ID(), errors.BadRequest.WithFormat("%v is not signed", txn.ID())), nil
	}

	batch := b.Block.Batch.Begin(true)
	defer batch.Discard()

	isRemote := txn.Transaction.Body.Type() == protocol.TransactionTypeRemote
	record := batch.Transaction(txn.Transaction.GetHash())
	s, err := record.Main().Get()
	switch {
	case errors.Is(err, errors.NotFound) && !isRemote:
		// Store the transaction

	case err != nil:
		// Unknown error or remote transaction with no local copy
		return nil, errors.UnknownError.WithFormat("load transaction: %w", err)

	case s.Transaction != nil:
		// It's not a transaction
		return protocol.NewErrorStatus(txn.ID(), errors.BadRequest.With("not a transaction")), nil

	case isRemote || s.Transaction.Equal(txn.Transaction):
		// Transaction has already been recorded
		return nil, nil

	default:
		// This should be impossible
		return nil, errors.InternalError.WithFormat("submitted transaction does not match the locally stored transaction")
	}

	// TODO Can we remove this or do it a better way?
	if txn.Transaction.Body.Type() == protocol.TransactionTypeSystemWriteData {
		return protocol.NewErrorStatus(txn.ID(), errors.BadRequest.WithFormat("a %v transaction cannot be submitted directly", protocol.TransactionTypeSystemWriteData)), nil
	}

	// If we reach this point, Validate should have verified that there is a
	// signer that can be charged for this recording
	err = record.Main().Put(&database.SigOrTxn{Transaction: txn.Transaction})
	if err != nil {
		return nil, errors.UnknownError.WithFormat("store transaction: %w", err)
	}

	// Record when the transaction is received
	status, err := record.Status().Get()
	if err != nil {
		return nil, errors.UnknownError.Wrap(err)
	}
	if status.Received == 0 {
		status.Received = b.Block.Index
		err = record.Status().Put(status)
		if err != nil {
			return nil, errors.UnknownError.Wrap(err)
		}
	}

	err = batch.Commit()
	if err != nil {
		return nil, errors.UnknownError.Wrap(err)
	}

	// The transaction has not yet been processed so don't add its status
	return nil, nil
}

// processUserSignatureMessage processes a signature. See
// [bundle.executeSignature].
func (b *bundle) processUserSignatureMessage(sig *messaging.UserSignature) (*protocol.TransactionStatus, error) {
	batch := b.Block.Batch.Begin(true)
	defer batch.Discard()

	// Load the transaction
	txn, err := batch.Transaction(sig.TransactionHash[:]).Main().Get()
	switch {
	case err != nil:
		return nil, errors.UnknownError.WithFormat("load transaction: %w", err)
	case txn.Transaction == nil:
		return protocol.NewErrorStatus(sig.ID(), errors.BadRequest.WithFormat("%x is not a transaction", sig.TransactionHash)), nil
	}

	status, err := b.executeSignature(batch, sig.Signature, txn.Transaction)
	if err != nil {
		return nil, errors.UnknownError.Wrap(err)
	}

	err = batch.Commit()
	if err != nil {
		return nil, errors.UnknownError.Wrap(err)
	}
	return status, nil
}

// processNetworkUpdateMessage constructs a transaction and signature for the
// network update, stores the transaction, and executes the signature (which
// queues the transaction for processing).
func (b *bundle) processNetworkUpdateMessage(msg *internal.NetworkUpdate) (*protocol.TransactionStatus, error) {
	sig := new(protocol.InternalSignature)
	sig.Cause = msg.Cause
	txn := new(protocol.Transaction)
	txn.Header.Principal = msg.Account
	txn.Header.Initiator = *(*[32]byte)(sig.Metadata().Hash())
	txn.Body = msg.Body
	sig.TransactionHash = *(*[32]byte)(txn.GetHash())

	b.internal.Add(txn.ID().Hash())
	b.internal.Add(*(*[32]byte)(sig.Hash()))

	batch := b.Block.Batch.Begin(true)
	defer batch.Discard()

	// Store the transaction
	err := batch.Transaction(sig.TransactionHash[:]).Main().Put(&database.SigOrTxn{Transaction: txn})
	if err != nil {
		return nil, errors.UnknownError.WithFormat("store transaction: %w", err)
	}

	// Process the signature
	_, err = b.executeSignature(batch, sig, txn)
	if err != nil {
		return nil, errors.UnknownError.Wrap(err)
	}

	// Don't return a status
	return nil, nil
}

// executeSignature executes the signature, queuing the transaction for
// processing when appropriate.
func (b *bundle) executeSignature(batch *database.Batch, signature protocol.Signature, transaction *protocol.Transaction) (*protocol.TransactionStatus, error) {
	// Process the transaction if it is synthetic or system, or the signature is
	// internal, or the signature is local to the principal
	if !transaction.Body.Type().IsUser() ||
		signature.Type() == protocol.SignatureTypeInternal ||
		signature.RoutingLocation().LocalTo(transaction.Header.Principal) {
		b.transactionsToProcess.Add(transaction.ID().Hash())
	}

	status := new(protocol.TransactionStatus)
	status.Received = b.Block.Index

	// TODO: add an ID method to signatures
	sigHash := *(*[32]byte)(signature.Hash())
	switch signature := signature.(type) {
	case protocol.KeySignature:
		status.TxID = signature.GetSigner().WithTxID(sigHash)
	case *protocol.ReceiptSignature, *protocol.InternalSignature:
		status.TxID = transaction.Header.Principal.WithTxID(sigHash)
	default:
		status.TxID = signature.RoutingLocation().WithTxID(sigHash)
	}

	s, err := b.Executor.ProcessSignature(batch, &chain.Delivery{
		Transaction: transaction,
		Internal:    b.internal.Has(sigHash),
		Forwarded:   b.forwarded.Has(sigHash),
	}, signature)
	b.Block.State.MergeSignature(s)
	if err == nil {
		status.Code = errors.Delivered
	} else {
		status.Set(err)
	}

	// Always record the signature and status
	if sig, ok := signature.(*protocol.RemoteSignature); ok {
		signature = sig.Signature
	}
	err = batch.Transaction(signature.Hash()).Main().Put(&database.SigOrTxn{Signature: signature, Txid: transaction.ID()})
	if err != nil {
		return nil, errors.UnknownError.WithFormat("store signature: %w", err)
	}
	err = batch.Transaction(signature.Hash()).Status().Put(status)
	if err != nil {
		return nil, errors.UnknownError.WithFormat("store signature status: %w", err)
	}
	return status, nil
}

// executeTransaction executes a transaction.
func (b *bundle) executeTransaction(hash [32]byte, additional bool) (*protocol.TransactionStatus, error) {
	batch := b.Block.Batch.Begin(true)
	defer batch.Discard()

	// Load the transaction. Earlier checks should guarantee this never fails.
	record := batch.Transaction(hash[:])
	txn, err := record.Main().Get()
	switch {
	case err != nil:
		return nil, errors.InternalError.WithFormat("load transaction: %w", err)
	case txn.Transaction == nil:
		return nil, errors.InternalError.WithFormat("%x is not a transaction", hash)
	}

	status, err := record.Status().Get()
	if err != nil {
		return nil, errors.UnknownError.Wrap(err)
	}

	delivery := &chain.Delivery{Transaction: txn.Transaction, Internal: b.internal.Has(hash)}
	err = delivery.LoadSyntheticMetadata(batch, txn.Transaction.Body.Type(), status)
	if err != nil {
		return nil, errors.UnknownError.Wrap(err)
	}

	status, state, err := b.Executor.ProcessTransaction(batch, delivery)
	if err != nil {
		return nil, err
	}

	err = batch.Commit()
	if err != nil {
		return nil, errors.UnknownError.WithFormat("commit batch: %w", err)
	}

	kv := []interface{}{
		"block", b.Block.Index,
		"type", txn.Transaction.Body.Type(),
		"code", status.Code,
		"txn-hash", logging.AsHex(txn.Transaction.GetHash()).Slice(0, 4),
		"principal", txn.Transaction.Header.Principal,
	}
	if status.Error != nil {
		kv = append(kv, "error", status.Error)
		if additional {
			b.Executor.logger.Info("Additional transaction failed", kv...)
		} else {
			b.Executor.logger.Info("Transaction failed", kv...)
		}
	} else if status.Pending() {
		if additional {
			b.Executor.logger.Debug("Additional transaction pending", kv...)
		} else {
			b.Executor.logger.Debug("Transaction pending", kv...)
		}
	} else {
		fn := b.Executor.logger.Debug
		switch txn.Transaction.Body.Type() {
		case protocol.TransactionTypeDirectoryAnchor,
			protocol.TransactionTypeBlockValidatorAnchor:
			fn = b.Executor.logger.Info
			kv = append(kv, "module", "anchoring")
		}
		if additional {
			fn("Additional transaction succeeded", kv...)
		} else {
			fn("Transaction succeeded", kv...)
		}
	}

	b.additional = append(b.additional, state.AdditionalMessages...)
	b.state.Set(hash, state)
	return status, nil
}
