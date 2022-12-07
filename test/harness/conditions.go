// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package harness

import (
	"context"

	"github.com/stretchr/testify/require"
	"gitlab.com/accumulatenetwork/accumulate/pkg/api/v3"
	"gitlab.com/accumulatenetwork/accumulate/pkg/errors"
	"gitlab.com/accumulatenetwork/accumulate/pkg/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

// A Condition is a function that is used by Harness.StepUntil to wait until
// some condition is met.
type Condition func(*Harness) bool

// Txn is used to define a condition on a transaction.
func Txn(id *url.TxID) *condTxn { return &condTxn{id} }

// IsDelivered waits until the transaction has been delivered (executed, whether
// success or failure).
func (c condTxn) IsDelivered() Condition { return c.status(isDelivered) }

// IsPending waits until the transaction is pending (received but not executed).
// IsPending will fail if the transaction has been recorded with any status
// other than pending.
func (c condTxn) IsPending() Condition { return c.status(isPending) }

// Succeeds waits until the transaction has been delivered and succeeds if the
// transaction succeeded (and fails otherwise).
func (c condTxn) Succeeds() Condition { return c.status(succeeds) }

// Fails waits until the transaction has been delivered and succeeds if the
// transaction failed (and fails otherwise).
func (c condTxn) Fails() Condition { return c.status(fails) }

// Produced is used to define a condition on the transactions produced by a
// transaction.
func (c condTxn) Produced() condProduced { return condProduced(c) }

// IsDelivered waits until the produced transaction(s) have been delivered
// (executed, whether success or failure).
func (c condProduced) IsDelivered() Condition { return c.status(isDelivered) }

// IsPending waits until the produced transaction(s) are pending (received but
// not executed). IsPending will fail if the transaction(s) have been recorded
// with any status other than pending.
func (c condProduced) IsPending() Condition { return c.status(isPending) }

// Succeeds waits until the produced transaction(s) have been delivered and
// succeeds if the transaction(s) succeeded (and fails otherwise).
func (c condProduced) Succeeds() Condition { return c.status(succeeds) }

// Fails waits until the produced transaction(s) have been delivered and
// succeeds if the transaction(s) failed (and fails otherwise).
func (c condProduced) Fails() Condition { return c.status(fails) }

type condTxn struct{ id *url.TxID }
type condProduced struct{ id *url.TxID }

func (c condTxn) status(predicate func(h *Harness, status *protocol.TransactionStatus) bool) Condition {
	return func(h *Harness) bool {
		// Query the transaction
		h.TB.Helper()
		r, err := h.Query().QueryTransaction(context.Background(), c.id, nil)
		switch {
		case err == nil:
			// Evaluate the predicate
			r.Status.TxID = c.id
			return predicate(h, r.Status)

		case errors.Is(err, errors.NotFound):
			// Wait
			return false

		default:
			// Unknown error
			require.NoError(h.TB, err)
			panic("not reached")
		}
	}
}

func (c condProduced) status(predicate func(sim *Harness, status *protocol.TransactionStatus) bool) Condition {
	var produced []*api.TxIDRecord
	return func(h *Harness) bool {
		h.TB.Helper()

		// Wait for the transaction to resolve
		if produced == nil {
			r, err := h.Query().QueryTransaction(context.Background(), c.id, nil)
			switch {
			case err == nil:
				// If the transaction is pending, wait
				if !r.Status.Delivered() {
					return false
				}

				// Record the produced transactions
				produced = r.Produced.Records

			case errors.Is(err, errors.NotFound):
				// Wait
				return false

			default:
				require.NoError(h.TB, err)
				panic("not reached")
			}
		}

		// Expect produced transactions
		if len(produced) == 0 {
			h.TB.Fatalf("%v did not produce transactions", c.id)
		}

		// Wait for the produced transactions to be received
		for _, r := range produced {
			h.TB.Helper()
			r, err := h.Query().QueryTransaction(context.Background(), r.Value, nil)
			switch {
			case err == nil:
				// Evaluate the predicate
				r.Status.TxID = r.TxID
				if !predicate(h, r.Status) {
					return false
				}

			case errors.Is(err, errors.NotFound):
				// Wait
				return false

			default:
				require.NoError(h.TB, err)
				panic("not reached")
			}
		}

		// All predicates passed
		return true
	}
}

func isDelivered(h *Harness, status *protocol.TransactionStatus) bool {
	h.TB.Helper()
	return status.Delivered()
}

func isPending(h *Harness, status *protocol.TransactionStatus) bool {
	h.TB.Helper()

	// Wait for a non-zero status
	if status.Code == 0 {
		return false
	}

	// Must be pending
	if status.Code != errors.Pending {
		h.TB.Fatal("Expected transaction to be pending")
	}
	return true
}

func succeeds(h *Harness, status *protocol.TransactionStatus) bool {
	h.TB.Helper()

	// Wait for delivery
	if !status.Delivered() {
		return false
	}

	// Must be success
	if status.Failed() {
		h.TB.Fatal("Expected transaction to succeeed")
	}
	return true
}

func fails(h *Harness, status *protocol.TransactionStatus) bool {
	h.TB.Helper()

	// Wait for delivery
	if !status.Delivered() {
		return false
	}

	// Must be failure
	if !status.Failed() {
		h.TB.Fatal("Expected transaction to fail")
	}
	return true
}
