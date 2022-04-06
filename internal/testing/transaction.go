package testing

import (
	"fmt"
	"time"

	"gitlab.com/accumulatenetwork/accumulate/internal/url"
	"gitlab.com/accumulatenetwork/accumulate/pkg/client/signing"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

type TransactionBuilder struct {
	*protocol.Envelope
	signer signing.Builder
}

func NewTransaction() TransactionBuilder {
	var tb TransactionBuilder
	tb.Envelope = new(protocol.Envelope)
	return tb
}

func (tb *TransactionBuilder) ensureTxn() {
	if tb.Transaction != nil {
		return
	}
	// if tb.TxHash != nil {
	// 	panic("can't have both a transaction hash and body")
	// }
	tb.Transaction = new(protocol.Transaction)
}

func (tb TransactionBuilder) WithHeader(hdr *protocol.TransactionHeader) TransactionBuilder {
	tb.ensureTxn()
	tb.Transaction.Header = *hdr
	return tb
}

func (tb TransactionBuilder) WithPrincipal(origin *url.URL) TransactionBuilder {
	tb.ensureTxn()
	tb.Transaction.Header.Principal = origin
	return tb
}

func (tb TransactionBuilder) WithSigner(signer *url.URL, height uint64) TransactionBuilder {
	tb.signer.SetUrl(signer)
	tb.signer.SetVersion(height)
	return tb
}

func (tb TransactionBuilder) WithTimestamp(nonce uint64) TransactionBuilder {
	tb.signer.SetTimestamp(nonce)
	return tb
}

func (tb TransactionBuilder) WithNonceVar(nonce *uint64) TransactionBuilder {
	tb.signer.SetTimestampWithVar(nonce)
	return tb
}

func (tb TransactionBuilder) WithCurrentTimestamp() TransactionBuilder {
	tb.signer.SetTimestamp(uint64(time.Now().UTC().UnixNano()))
	return tb
}

func (tb TransactionBuilder) WithBody(body protocol.TransactionBody) TransactionBuilder {
	tb.ensureTxn()
	tb.Transaction.Body = body
	return tb
}

func (tb TransactionBuilder) WithTxnHash(hash []byte) TransactionBuilder {
	// if tb.Transaction != nil {
	// 	panic("can't have both a transaction hash and body")
	// }
	tb.TxHash = hash
	return tb
}

func (tb TransactionBuilder) Sign(typ protocol.SignatureType, privateKey []byte) *protocol.Envelope {
	switch {
	case tb.TxHash != nil:
		// OK
	case tb.Transaction == nil:
		panic("cannot sign a transaction without the transaction body or transaction hash")
	case tb.Transaction.Header.Initiator == ([32]byte{}):
		panic("cannot sign a transaction before setting the initiator")
	}

	tb.signer.SetPrivateKey(privateKey)
	sig, err := tb.signer.Sign(tb.GetTxHash())
	if err != nil {
		panic(err)
	}

	tb.Signatures = append(tb.Signatures, sig)
	return tb.Envelope
}

func (tb TransactionBuilder) Initiate(typ protocol.SignatureType, privateKey []byte) *protocol.Envelope {
	if tb.TxHash != nil {
		panic("cannot initiate transaction: have hash instead of body")
	}
	if tb.Transaction.Header.Initiator != ([32]byte{}) {
		panic("cannot initiate transaction: already initiated")
	}

	tb.signer.Type = typ
	tb.signer.SetPrivateKey(privateKey)
	sig, err := tb.signer.Initiate(tb.Transaction)
	if err != nil {
		panic(err)
	}

	tb.Signatures = append(tb.Signatures, sig)
	return tb.Envelope
}

func (tb TransactionBuilder) InitiateSynthetic(destSubnetUrl *url.URL) TransactionBuilder {
	if tb.TxHash != nil {
		panic("cannot initiate transaction: have hash instead of body")
	}
	if tb.Transaction.Header.Initiator != ([32]byte{}) {
		panic("cannot initiate transaction: already initiated")
	}
	if tb.signer.Url == nil {
		panic("missing signer")
	}
	if tb.signer.Version == 0 {
		panic("missing version")
	}

	initSig := new(protocol.SyntheticSignature)
	initSig.SourceNetwork = tb.signer.Url
	initSig.DestinationNetwork = destSubnetUrl
	initSig.SequenceNumber = tb.signer.Version

	initHash, err := initSig.InitiatorHash()
	if err != nil {
		// This should never happen
		panic(fmt.Errorf("failed to calculate the synthetic signature initiator hash: %v", err))
	}

	tb.Transaction.Header.Initiator = *(*[32]byte)(initHash)
	tb.Signatures = append(tb.Signatures, initSig)
	return tb
}

func (tb TransactionBuilder) Faucet() *protocol.Envelope {
	sig, err := new(signing.Builder).UseFaucet().Initiate(tb.Transaction)
	if err != nil {
		panic(err)
	}

	tb.Signatures = append(tb.Signatures, sig)
	return tb.Envelope
}
