// Copyright 2023 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package block

import "gitlab.com/accumulatenetwork/accumulate/protocol"

type SignatureContext struct {
	*MessageContext
	signature   protocol.Signature
	transaction *protocol.Transaction
}

func (s *SignatureContext) Type() protocol.SignatureType { return s.signature.Type() }
