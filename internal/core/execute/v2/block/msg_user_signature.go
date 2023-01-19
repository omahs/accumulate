// Copyright 2023 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package block

import (
	"gitlab.com/accumulatenetwork/accumulate/pkg/errors"
	"gitlab.com/accumulatenetwork/accumulate/pkg/types/messaging"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

func init() {
	messageExecutors = append(messageExecutors, func(ExecutorOptions) MessageExecutor { return UserSignature{} })
}

// UserSignature processes a signature. See [bundle.executeSignature].
type UserSignature struct{}

func (UserSignature) Type() messaging.MessageType { return messaging.MessageTypeUserSignature }

func (UserSignature) Process(b *bundle, msg messaging.Message) (*protocol.TransactionStatus, error) {
	sig, ok := msg.(*messaging.UserSignature)
	if !ok {
		return nil, errors.InternalError.WithFormat("invalid message type: expected %v, got %v", messaging.MessageTypeUserSignature, msg.Type())
	}

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
