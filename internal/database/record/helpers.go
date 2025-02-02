// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package record

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"gitlab.com/accumulatenetwork/accumulate/pkg/errors"
	"gitlab.com/accumulatenetwork/accumulate/pkg/types/encoding"
	"gitlab.com/accumulatenetwork/accumulate/pkg/url"
)

var ErrNotEnoughData = errors.BadRequest.With("not enough data")
var ErrOverflow = errors.BadRequest.With("overflow")

// var ErrMalformedBigInt = errors.New("invalid big integer string")

func zero[T any]() (z T)                     { return z }
func copyValue[T any](v T) T                 { return v }
func copyRef[T interface{ Copy() T }](v T) T { return v.Copy() } //nolint

func CompareHash(u, v [32]byte) int  { return bytes.Compare(u[:], v[:]) }
func CompareTxid(u, v *url.TxID) int { return u.Compare(v) }
func CompareUrl(u, v *url.URL) int   { return u.Compare(v) }

func ParseString(s string) (string, error) { return s, nil }

func MapKeyBytes(v []byte) [32]byte {
	if len(v) == 32 {
		// Most (all?) of our byte-keys are actually [32]byte so this should help
		return *(*[32]byte)(v)
	}
	return sha256.Sum256(v)
}

func MapKeyUrl(v *url.URL) [32]byte {
	return v.AccountID32()
}

type ValueMarshaller[T any] func(value T) ([]byte, error)
type ValueUnmarshaller[T any] func(data []byte) (T, error)

type wrapperFuncs[T any] struct {
	copy      func(T) T
	marshal   ValueMarshaller[T]
	unmarshal ValueUnmarshaller[T]
}

// UintWrapper defines un/marshalling functions for uint fields.
var UintWrapper = &wrapperFuncs[uint64]{
	copy:      copyValue[uint64],
	marshal:   oldMarshal(encoding.MarshalUint),
	unmarshal: encoding.UnmarshalUint,
}

// BytesWrapper defines un/marshalling functions for byte slice fields.
var BytesWrapper = &wrapperFuncs[[]byte]{
	copy:      func(v []byte) []byte { u := make([]byte, len(v)); copy(u, v); return u },
	marshal:   oldMarshal(encoding.MarshalBytes),
	unmarshal: encoding.UnmarshalBytes,
}

// HashWrapper defines un/marshalling functions for hash fields.
var HashWrapper = &wrapperFuncs[[32]byte]{
	copy:      copyValue[[32]byte],
	marshal:   oldMarshalPtr(encoding.MarshalHash),
	unmarshal: encoding.UnmarshalHash,
}

// UrlWrapper defines un/marshalling functions for url fields.
var UrlWrapper = &wrapperFuncs[*url.URL]{
	copy:      copyValue[*url.URL], // URLs are immutable so don't copy
	marshal:   marshalAsString[*url.URL],
	unmarshal: unmarshalFromString(url.Parse),
}

// TxidWrapper defines un/marshalling functions for txid fields.
var TxidWrapper = &wrapperFuncs[*url.TxID]{
	copy:      copyValue[*url.TxID], // TxIDs are immutable so don't copy
	marshal:   marshalAsString[*url.TxID],
	unmarshal: unmarshalFromString(url.ParseTxID),
}

// StringWrapper defines un/marshalling functions for string fields.
var StringWrapper = &wrapperFuncs[string]{
	copy:      copyValue[string],
	marshal:   oldMarshal(encoding.MarshalString),
	unmarshal: encoding.UnmarshalString,
}

func oldMarshal[T any](fn func(T) []byte) ValueMarshaller[T] {
	return func(value T) ([]byte, error) {
		return fn(value), nil
	}
}

func oldMarshalPtr[T any](fn func(*T) []byte) ValueMarshaller[T] {
	return func(value T) ([]byte, error) {
		return fn(&value), nil
	}
}

func marshalAsString[T fmt.Stringer](v T) ([]byte, error) {
	return encoding.MarshalString(v.String()), nil
}

func unmarshalFromString[T any](fn func(string) (T, error)) ValueUnmarshaller[T] {
	return func(data []byte) (T, error) {
		s, err := encoding.UnmarshalString(data)
		if err != nil {
			return zero[T](), errors.UnknownError.Wrap(err)
		}
		v, err := fn(s)
		if err != nil {
			return zero[T](), errors.UnknownError.Wrap(err)
		}
		return v, nil
	}
}
