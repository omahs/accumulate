// Copyright 2023 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package block

import (
	"gitlab.com/accumulatenetwork/accumulate/internal/database"
	"gitlab.com/accumulatenetwork/accumulate/pkg/errors"
	"gitlab.com/accumulatenetwork/accumulate/pkg/types/messaging"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

// MessageContext is the context in which a message is processed.
type MessageContext struct {
	// bundle is the message bundle being processed.
	*bundle

	// parent is the parent message context, or nil.
	parent *MessageContext

	// message is the Message being processed.
	message messaging.Message

	// produced is other messages produced while processing the message.
	produced []messaging.Message
}

func (m *MessageContext) Type() messaging.MessageType { return m.message.Type() }

func (m *MessageContext) childWith(msg messaging.Message) *MessageContext {
	n := new(MessageContext)
	n.bundle = m.bundle
	n.message = msg
	n.parent = m
	return n
}

func (m *MessageContext) sigWith(sig protocol.Signature, txn *protocol.Transaction) *SignatureContext {
	s := new(SignatureContext)
	s.MessageContext = m
	s.signature = sig
	s.transaction = txn
	return s
}

func (m *MessageContext) parentIsSynthetic() bool {
	return m.parent != nil && m.parent.message.Type() == messaging.MessageTypeSynthetic
}

func (m *MessageContext) didProduce(msg messaging.Message) {
	m.produced = append(m.produced, msg)
}

func (m *MessageContext) callMessageExecutor(batch *database.Batch, msg messaging.Message) (*protocol.TransactionStatus, error) {
	c := m.childWith(msg)
	st, err := m.bundle.callMessageExecutor(batch, c)
	if err != nil || st.Failed() {
		return st, errors.UnknownError.Wrap(err)
	}
	m.produced = append(m.produced, c.produced...)
	return st, nil
}
