package build

import "gitlab.com/accumulatenetwork/accumulate/protocol"

type EnvelopeBuilder struct {
	parser
	transaction *protocol.Transaction
	signatures  []protocol.Signature
}

func (b EnvelopeBuilder) SignWith(signer any, path ...string) SignatureBuilder {
	return SignatureBuilder{env: b}.WithSigner(signer, path...)
}

func (b EnvelopeBuilder) Build() (*protocol.Envelope, error) {
	if !b.ok() {
		return nil, b.err()
	}

	env := new(protocol.Envelope)
	env.Transaction = []*protocol.Transaction{b.transaction}
	env.Signatures = b.signatures
	return env, nil
}
