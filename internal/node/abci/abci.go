// Copyright 2023 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

// Package abci implements the Accumulate ABCI applications.
//
// # Transaction Processing
//
// Tendermint processes transactions in the following phases:
//
//   - BeginBlock
//   - [CheckTx]
//   - [DeliverTx]
//   - EndBlock
//   - Commit
package abci

import (
	"gitlab.com/accumulatenetwork/accumulate/internal/core/execute"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/pkg/types/messaging"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

// Version is the version of the ABCI applications.
const Version uint64 = 0x1

// ValidateEnvelopeSet validates a set of messages. ValidateEnvelopeSet returns
// one status per delivery. If a message fails validation, the error is written
// to that message's status.
func ValidateEnvelopeSet(x execute.Executor, batch *database.Batch, messages []messaging.Message) []*protocol.TransactionStatus {
	results := make([]*protocol.TransactionStatus, len(messages))
	for i, message := range messages {
		status, err := x.Validate(batch, message)
		if status == nil {
			status = new(protocol.TransactionStatus)
		}
		results[i] = status

		if err != nil {
			status.Set(err)
		}
	}

	return results
}

// ExecuteEnvelopeSet validates a set of messages. ExecuteEnvelopeSet returns
// one status per delivery. If a message fails validation, the error is written
// to that message's status.
func ExecuteEnvelopeSet(block execute.Block, messages []messaging.Message) []*protocol.TransactionStatus {
	results := make([]*protocol.TransactionStatus, len(messages))
	for i, message := range messages {
		status, err := block.Process(message)
		if status == nil {
			status = new(protocol.TransactionStatus)
		}
		results[i] = status

		if err != nil {
			status.Set(err)
		}
	}

	return results
}
