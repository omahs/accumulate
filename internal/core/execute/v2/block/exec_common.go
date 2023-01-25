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

type ExecutorFor[T any, V interface{ Type() T }] interface {
	Process(*database.Batch, V) (*protocol.TransactionStatus, error)
}

type MessageExecutor = ExecutorFor[messaging.MessageType, *MessageContext]
type SignatureExecutor = ExecutorFor[protocol.SignatureType, *SignatureContext]

func newExecutorMap[T comparable, V interface{ Type() T }](opts ExecutorOptions, list []func(ExecutorOptions) (T, ExecutorFor[T, V])) map[T]ExecutorFor[T, V] {
	m := map[T]ExecutorFor[T, V]{}
	for _, fn := range list {
		typ, x := fn(opts)
		if _, ok := m[typ]; ok {
			panic(errors.InternalError.WithFormat("duplicate executor for %v", typ))
		}
		m[typ] = x
	}
	return m
}

func registerSimpleExec[X ExecutorFor[T, V], T any, V interface{ Type() T }](list *[]func(ExecutorOptions) (T, ExecutorFor[T, V]), typ ...T) {
	for _, typ := range typ {
		typ := typ // See docs/developer/rangevarref.md
		*list = append(*list, func(ExecutorOptions) (T, ExecutorFor[T, V]) {
			var x X
			return typ, x
		})
	}
}

// MessageContext is the context in which a message is executed.
type MessageContext struct {
	*bundle
	message messaging.Message
	parent  *MessageContext
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

type SignatureContext struct {
	*MessageContext
	signature   protocol.Signature
	transaction *protocol.Transaction
}

func (s *SignatureContext) Type() protocol.SignatureType { return s.signature.Type() }
