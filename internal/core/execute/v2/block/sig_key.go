// Copyright 2023 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package block

import (
	"gitlab.com/accumulatenetwork/accumulate/internal/core/execute/v2/chain"
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/pkg/errors"
	"gitlab.com/accumulatenetwork/accumulate/pkg/types/messaging"
	"gitlab.com/accumulatenetwork/accumulate/pkg/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

func init() {
	registerSimpleExec[KeySignature](&signatureExecutors,
		protocol.SignatureTypeLegacyED25519,
		protocol.SignatureTypeED25519,
		protocol.SignatureTypeRCD1,
		protocol.SignatureTypeBTC,
		protocol.SignatureTypeBTCLegacy,
		protocol.SignatureTypeETH,

		// TODO Remove?
		protocol.SignatureTypeDelegated,

		// TODO Remove
		protocol.SignatureTypeRemote,
		protocol.SignatureTypeSet,
	)
}

// KeySignature processes key signatures.
type KeySignature struct{}

func (KeySignature) Process(batch *database.Batch, ctx *SignatureContext) (*protocol.TransactionStatus, error) {
	status := new(protocol.TransactionStatus)
	status.TxID = ctx.message.ID()
	status.Received = ctx.Block.Index

	_, err := ctx.Executor.processSignature2(batch, &chain.Delivery{
		Transaction: ctx.transaction,
		Forwarded:   ctx.forwarded.Has(ctx.message.ID().Hash()),
	}, ctx.signature)
	ctx.Block.State.MergeSignature(&ProcessSignatureState{})
	if err == nil {
		status.Code = errors.Delivered
	} else {
		status.Set(err)
	}
	if status.Failed() {
		return status, nil
	}

	// Collect delegators and the inner signature
	var delegators []*url.URL
	sig := ctx.signature
	for {
		del, ok := sig.(*protocol.DelegatedSignature)
		if !ok {
			break
		}
		delegators = append(delegators, del.Delegator)
		sig = del.Signature
	}
	for i, n := 0, len(delegators); i < n/2; i++ {
		j := n - 1 - i
		delegators[i], delegators[j] = delegators[j], delegators[i]
	}

	// If the signer's authority is satisfied, send the next authority signature
	txst, err := batch.Transaction(ctx.transaction.GetHash()).Status().Get()
	if err != nil {
		return nil, errors.UnknownError.WithFormat("load status: %w", err)
	}

	signerAuth := sig.GetSigner()
	if key, _ := protocol.ParseLiteIdentity(signerAuth); key != nil {
		// Ok
	} else if key, _, _ := protocol.ParseLiteTokenAddress(signerAuth); key != nil {
		signerAuth = signerAuth.RootIdentity()
	} else {
		signerAuth = signerAuth.Identity()
	}

	ok, err := ctx.Executor.AuthorityIsSatisfied(batch, ctx.transaction, txst, signerAuth)
	if err != nil {
		return nil, errors.UnknownError.Wrap(err)
	}
	if ok {
		// TODO Add payment if initiator
		auth := &protocol.AuthoritySignature{
			Authority: signerAuth,
			Vote:      protocol.VoteTypeAccept,
			TxID:      ctx.transaction.ID(),
			Initiator: protocol.SignatureDidInitiate(ctx.signature, ctx.transaction.Header.Initiator[:], nil),
			Delegator: delegators,
		}

		// TODO Deduplicate
		ctx.didProduce(&messaging.UserSignature{
			Signature: auth,
			TxID:      ctx.transaction.ID(),
		})
	}

	return status, nil
}
