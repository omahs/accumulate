// Copyright 2023 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package abci

//lint:file-ignore ST1001 Don't care

import (
	"crypto/sha256"

	"github.com/tendermint/tendermint/libs/log"
	"gitlab.com/accumulatenetwork/accumulate/internal/core/execute"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/logging"
	"gitlab.com/accumulatenetwork/accumulate/pkg/errors"
	"gitlab.com/accumulatenetwork/accumulate/pkg/types/messaging"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

type executeFunc func([]messaging.Message) []*protocol.TransactionStatus

func executeTransactions(logger log.Logger, execute executeFunc, raw []byte) ([]messaging.Message, []*protocol.TransactionStatus, []byte, error) {
	hash := sha256.Sum256(raw)
	envelope := new(protocol.Envelope)
	err := envelope.UnmarshalBinary(raw)
	if err != nil {
		logger.Info("Failed to unmarshal", "tx", logging.AsHex(hash), "error", err)
		return nil, nil, nil, errors.UnknownError.WithFormat("decoding envelopes: %w", err)
	}

	deliveries, err := messaging.NormalizeLegacy(envelope)
	if err != nil {
		logger.Info("Failed to normalize envelope", "tx", logging.AsHex(hash), "error", err)
		return nil, nil, nil, errors.UnknownError.Wrap(err)
	}

	results := execute(deliveries)

	// If the results can't be marshaled, provide no results but do not fail the
	// batch
	rset, err := (&protocol.TransactionResultSet{Results: results}).MarshalBinary()
	if err != nil {
		logger.Error("Unable to encode result", "error", err)
		return deliveries, results, nil, nil
	}

	return deliveries, results, rset, nil
}

func checkTx(exec execute.Executor, batch *database.Batch) executeFunc {
	return func(deliveries []messaging.Message) []*protocol.TransactionStatus {
		return ValidateEnvelopeSet(exec, batch, deliveries)
	}
}

func deliverTx(exec execute.Executor, block execute.Block) executeFunc {
	return func(deliveries []messaging.Message) []*protocol.TransactionStatus {
		return ExecuteEnvelopeSet(block, deliveries)
	}
}
