// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package managed

import (
	"gitlab.com/accumulatenetwork/accumulate/pkg/types/encoding"
	"gitlab.com/accumulatenetwork/accumulate/smt/storage"
)

//go:generate go run ../../tools/cmd/gen-types --package managed types.yml

type DbValue[T encoding.BinaryValue] interface {
	Get() (T, error)
	Put(T) error
}

type DbInt interface {
	Get() (int64, error)
	Put(int64) error
}

type DbHash interface {
	Get() (Hash, error)
	Put(Hash) error
}

type DbManager interface {
	Int(key storage.Key) DbInt
	Hash(key storage.Key) DbHash
	State(key storage.Key) DbValue[*MerkleState]
}
