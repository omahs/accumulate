package database

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"

	"gitlab.com/accumulatenetwork/accumulate/internal/encoding/hash"
	"gitlab.com/accumulatenetwork/accumulate/internal/sortutil"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	"gitlab.com/accumulatenetwork/accumulate/smt/storage"
)

type SignatureSet struct {
	txn      *Transaction
	signer   protocol.Signer
	writable bool
	hashes   *sigSetData
}

// newSigSet creates a new SignatureSet.
func newSigSet(txn *Transaction, signer protocol.Signer, writable bool) (*SignatureSet, error) {
	s := new(SignatureSet)
	s.txn = txn
	s.signer = signer
	s.writable = writable
	s.hashes = new(sigSetData)

	err := txn.batch.getValuePtr(s.key(), s.hashes, &s.hashes, true)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return nil, err
	}

	// Reset if the set is writable and the version is different
	if writable && s.hashes.Version != signer.GetVersion() {
		s.hashes.Reset(signer.GetVersion())
	}
	return s, nil
}

func (s *SignatureSet) key() storage.Key {
	return s.txn.key.Signatures(s.signer.GetUrl())
}

func (s *SignatureSet) Count() int {
	return len(s.hashes.Entries)
}

func (s *SignatureSet) EntryHashes() [][32]byte {
	h := make([][32]byte, len(s.hashes.Entries))
	for i, e := range s.hashes.Entries {
		h[i] = e.EntryHash
	}
	return h
}

func (s *sigSetData) Reset(version uint64) {
	// Retain system signature entries
	system := make([]sigSetKeyData, 0, len(s.Entries))
	for _, e := range s.Entries {
		if e.System {
			system = append(system, e)
		}
	}

	// Remove all other entries and update the version
	s.Version = version
	s.Entries = system
}

func (s *sigSetData) Add(newEntry sigSetKeyData, newSignature protocol.Signature) bool {
	// The signature is a system signature if it's one of the system types or if
	// the signer is a node.
	switch {
	case newSignature.Type().IsSystem():
		newEntry.System = true
	case protocol.IsDnUrl(newSignature.GetSigner()):
		newEntry.System = true
	default:
		_, ok := protocol.ParseBvnUrl(newSignature.GetSigner())
		newEntry.System = ok
	}

	// Check the signer version
	if !newEntry.System && s.Version != newSignature.GetSignerVersion() {
		return false
	}

	// Find based on the key keyHash
	ptr, new := sortutil.BinaryInsert(&s.Entries, func(entry sigSetKeyData) int {
		return bytes.Compare(entry.KeyHash[:], newEntry.KeyHash[:])
	})

	*ptr = newEntry
	return new
}

// Add adds a signature to the signature set. Add does nothing if the signature
// set already includes the signer's public key. The entry hash must refer to a
// signature chain entry.
func (s *SignatureSet) Add(newSignature protocol.Signature) (int, error) {
	if !s.writable {
		return 0, fmt.Errorf("signature set opened as read-only")
	}

	data, err := newSignature.MarshalBinary()
	if err != nil {
		return 0, fmt.Errorf("marshal signature: %w", err)
	}

	var newEntry sigSetKeyData
	newEntry.EntryHash = sha256.Sum256(data)
	newEntry.KeyHash, err = signatureKeyHash(newSignature)
	if err != nil {
		return 0, err
	}

	if !s.hashes.Add(newEntry, newSignature) {
		return len(s.hashes.Entries), nil
	}

	err = s.txn.ensureSigner(s.signer)
	if err != nil {
		return 0, err
	}

	s.txn.batch.putValue(s.key(), s.hashes)
	return len(s.hashes.Entries), nil
}

// signatureKeyHash returns a hash that is used to prevent two signatures from
// the same key.
func signatureKeyHash(sig protocol.Signature) ([32]byte, error) {
	switch sig := sig.(type) {
	case protocol.KeySignature:
		// Normal signatures must come from a unique key
		return sha256.Sum256(sig.GetPublicKey()), nil

	case *protocol.ForwardedSignature:
		return signatureKeyHash(sig.Signature)

	case *protocol.SyntheticSignature:
		// Multiple synthetic signatures doesn't make sense, but if they're
		// unique... ok
		hasher := make(hash.Hasher, 0, 3)
		hasher.AddUrl(sig.SourceNetwork)
		hasher.AddUrl(sig.DestinationNetwork)
		hasher.AddUint(sig.SequenceNumber)
		return *(*[32]byte)(hasher.MerkleHash()), nil

	case *protocol.ReceiptSignature:
		// Multiple receipts doesn't make sense, but if they anchor to a unique
		// root... ok
		return *(*[32]byte)(sig.Result), nil

	case *protocol.SystemSignature:
		// Internal signatures only make any kind of sense if they're coming
		// from the local network, so they should never be different
		return sha256.Sum256([]byte("Accumulate Internal Signature")), nil

	default:
		return [32]byte{}, fmt.Errorf("unrecognized signature type %T", sig)
	}
}
