package abci

import (
	"crypto/sha256"

	"github.com/getsentry/sentry-go"
	"github.com/tendermint/tendermint/libs/log"
	. "gitlab.com/accumulatenetwork/accumulate/internal/block"
	"gitlab.com/accumulatenetwork/accumulate/internal/chain"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/errors"
	"gitlab.com/accumulatenetwork/accumulate/internal/logging"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

type executeFunc func(*chain.Delivery) (*protocol.TransactionStatus, error)

func executeTransactions(logger log.Logger, execute executeFunc, raw []byte) ([]*chain.Delivery, []*protocol.TransactionStatus, []byte, *protocol.Error) {
	hash := sha256.Sum256(raw)
	envelope := new(protocol.Envelope)
	err := envelope.UnmarshalBinary(raw)
	if err != nil {
		sentry.CaptureException(err)
		logger.Info("Failed to unmarshal", "tx", logging.AsHex(hash), "error", err)
		return nil, nil, nil, &protocol.Error{Code: protocol.ErrorCodeEncodingError, Message: errors.New(errors.StatusBadRequest, "Unable to decode transaction(s)")}
	}

	deliveries, err := chain.NormalizeEnvelope(envelope)
	if err != nil {
		sentry.CaptureException(err)
		logger.Info("Failed to normalize envelope", "tx", logging.AsHex(hash), "error", err)
		return nil, nil, nil, protocol.NewError(protocol.ErrorCodeUnknownError, err)
	}

	results := make([]*protocol.TransactionStatus, len(deliveries))
	for i, env := range deliveries {
		status, err := execute(env)
		if err == nil {
			results[i] = status
			continue
		}

		if status == nil {
			status = new(protocol.TransactionStatus)
		}

		sentry.CaptureException(err)
		status.Message = err.Error()

		var err1 *protocol.Error
		if status.Code == 0 && errors.As(err, &err1) {
			status.Code = err1.Code.GetEnumValue()
		}

		var err2 *errors.Error
		if status.Error == nil && errors.As(err, &err2) {
			status.Error = err2
		}

		if status.Code == 0 {
			status.Code = protocol.ErrorCodeUnknownError.GetEnumValue()
		}

		results[i] = status
	}

	// If the results can't be marshaled, provide no results but do not fail the
	// batch
	rset, err := (&protocol.TransactionResultSet{Results: results}).MarshalBinary()
	if err != nil {
		sentry.CaptureException(err)
		logger.Error("Unable to encode result", "error", err)
		return deliveries, results, nil, nil
	}

	return deliveries, results, rset, nil
}

func checkTx(exec *Executor, db *database.Database) executeFunc {
	return func(envelope *chain.Delivery) (*protocol.TransactionStatus, error) {
		if exec.CheckTxBatch == nil { // For cases where we haven't started/ended a block yet
			exec.CheckTxBatch = db.Begin(true)
		}

		result, err := exec.ValidateEnvelope(envelope)
		if err != nil {
			return nil, protocol.NewError(protocol.ErrorCodeUnknownError, err)
		}
		if result == nil {
			result = new(protocol.EmptyResult)
		}
		return &protocol.TransactionStatus{Result: result}, nil
	}
}

func deliverTx(exec *Executor, block *Block) executeFunc {
	return func(envelope *chain.Delivery) (*protocol.TransactionStatus, error) {
		status, err := exec.ExecuteEnvelope(block, envelope)
		if err != nil {
			return nil, err
		}

		return status, nil
	}
}
