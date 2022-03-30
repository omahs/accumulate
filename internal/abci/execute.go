package abci

import (
	"crypto/sha256"
	"errors"

	"github.com/getsentry/sentry-go"
	"github.com/tendermint/tendermint/libs/log"
	"gitlab.com/accumulatenetwork/accumulate/internal/chain"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/internal/logging"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	"gitlab.com/accumulatenetwork/accumulate/types/api/transactions"
)

type executeFunc func(*protocol.Envelope) (protocol.TransactionResult, error)

func executeTransactions(logger log.Logger, execute executeFunc, raw []byte) ([]*protocol.Envelope, []*protocol.TransactionStatus, []byte, *protocol.Error) {
	hash := sha256.Sum256(raw)
	envelopes, err := transactions.UnmarshalAll(raw)
	if err != nil {
		sentry.CaptureException(err)
		logger.Info("Failed to unmarshal", "tx", logging.AsHex(hash), "error", err)
		return nil, nil, nil, &protocol.Error{Code: protocol.ErrorCodeEncodingError, Message: errors.New("Unable to decode transaction(s)")}
	}

	results := make([]*protocol.TransactionStatus, len(envelopes))
	for i, env := range envelopes {
		typ := env.Type()
		txid := env.GetTxHash()
		status := new(protocol.TransactionStatus)

		result, err := execute(env)
		if err != nil {
			sentry.CaptureException(err)
			logger.Info("Transaction failed",
				"type", env.Type(),
				"txid", logging.AsHex(txid),
				"hash", logging.AsHex(hash),
				"error", err,
				"principal", env.Transaction.Header.Principal)
			if err, ok := err.(*protocol.Error); ok {
				status.Code = err.Code.GetEnumValue()
			} else {
				status.Code = protocol.ErrorCodeUnknownError.GetEnumValue()
			}
			status.Message = err.Error()
		} else if !typ.IsInternal() && typ != protocol.TransactionTypeSyntheticAnchor {
			logger.Debug("Transaction succeeded",
				"type", typ,
				"txid", logging.AsHex(txid),
				"hash", logging.AsHex(hash))
		}

		status.Result = result
		results[i] = status
	}

	// If the results can't be marshaled, provide no results but do not fail the
	// batch
	var data []byte
	for _, r := range results {
		d, err := r.MarshalBinary()
		if err != nil {
			sentry.CaptureException(err)
			logger.Error("Unable to encode result", "error", err)
			return envelopes, results, nil, nil
		}
		data = append(data, d...)
	}

	return envelopes, results, data, nil
}

func checkTx(chain *chain.Executor, db *database.Database) executeFunc {
	return func(envelope *protocol.Envelope) (protocol.TransactionResult, error) {
		batch := db.Begin(false)
		defer batch.Discard()

		result, err := chain.ValidateEnvelope(batch, envelope)
		if err != nil {
			return nil, protocol.NewError(protocol.ErrorCodeUnknownError, err)
		}
		if result == nil {
			return new(protocol.EmptyResult), nil
		}
		return result, nil
	}
}

func deliverTx(chain *chain.Executor, block *chain.Block) executeFunc {
	return func(envelope *protocol.Envelope) (protocol.TransactionResult, error) {
		// Process signatures
		batch := block.Batch.Begin(true)
		defer batch.Discard()

		sigState, err := processSignatures(chain, batch, envelope)
		if err != nil {
			return nil, err
		}
		block.State.MergeSignature(sigState)

		err = batch.Commit()
		if err != nil {
			return nil, protocol.Errorf(protocol.ErrorCodeUnknownError, "commit batch: %w", err)
		}

		// Process the transaction
		batch = block.Batch.Begin(true)
		defer batch.Discard()

		status, txnState, err := processTransaction(chain, batch, envelope)
		if err != nil {
			return nil, protocol.Errorf(protocol.ErrorCodeUnknownError, "execute transaction: %w", err)
		}
		block.State.MergeTransaction(txnState)

		// Always commit
		err = batch.Commit()
		if err != nil {
			return nil, protocol.Errorf(protocol.ErrorCodeUnknownError, "commit batch: %w", err)
		}

		if status.Code != 0 {
			return status.Result, protocol.NewError(protocol.ErrorCode(status.Code), errors.New(status.Message))
		}

		return status.Result, nil
	}
}

func processSignatures(exec *chain.Executor, batch *database.Batch, envelope *protocol.Envelope) (*chain.ProcessSignatureState, error) {
	// Load the transaction
	transaction, err := exec.LoadTransaction(batch, envelope)
	if err != nil {
		return nil, protocol.NewError(protocol.ErrorCodeUnknownError, err)
	}

	// Process each signature
	state := new(chain.ProcessSignatureState)
	for _, signature := range envelope.Signatures {
		s, err := exec.ProcessSignature(batch, transaction, signature)
		if err != nil {
			return nil, protocol.NewError(protocol.ErrorCodeUnknownError, err)
		}
		state.Merge(s)
	}

	return state, nil
}

func processTransaction(chain *chain.Executor, batch *database.Batch, envelope *protocol.Envelope) (*protocol.TransactionStatus, *chain.ProcessTransactionState, error) {
	transaction, err := chain.LoadTransaction(batch, envelope)
	if err != nil {
		return nil, nil, protocol.NewError(protocol.ErrorCodeUnknownError, err)
	}

	// Process the transaction
	status, state, err := chain.ProcessTransaction(batch, transaction)
	if err != nil {
		return nil, nil, protocol.NewError(protocol.ErrorCodeUnknownError, err)
	}

	// Process synthetic transactions generated by the validator
	err = chain.ProduceSynthetic(batch, transaction, state.ProducedTxns)
	if err != nil {
		return nil, nil, protocol.NewError(protocol.ErrorCodeUnknownError, err)
	}

	return status, state, nil
}
