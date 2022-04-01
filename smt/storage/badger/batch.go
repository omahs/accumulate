package badger

import (
	"errors"
	"sync"

	"github.com/dgraph-io/badger"
	"gitlab.com/accumulatenetwork/accumulate/smt/storage"
	"gitlab.com/accumulatenetwork/accumulate/smt/storage/memory"
)

type Batch struct {
	db       *DB
	txn      *badger.Txn
	writable bool
}

var _ storage.KeyValueTxn = (*Batch)(nil)

func (db *DB) Begin(writable bool) storage.KeyValueTxn {
	b := new(Batch)
	b.db = db
	b.txn = db.badgerDB.NewTransaction(writable)
	b.writable = writable
	if db.logger == nil {
		return b
	}
	return &storage.DebugBatch{Batch: b, Logger: db.logger, Writable: writable}
}

func (b *Batch) Begin(writable bool) storage.KeyValueTxn {
	if !b.writable || !writable {
		return memory.NewBatch(b.Get, nil)
	}
	return memory.NewBatch(b.Get, b.PutAll)
}

func (b *Batch) lock() (sync.Locker, error) {
	l, err := b.db.lock(false)
	if err == nil {
		return l, nil
	}

	b.Discard() // Is this a good idea?
	return nil, err
}

func (b *Batch) Put(key storage.Key, value []byte) error {
	if l, err := b.lock(); err != nil {
		return err
	} else {
		defer l.Unlock()
	}

	return b.txn.Set(key[:], value)
}

func (b *Batch) PutAll(values map[storage.Key][]byte) error {
	if l, err := b.lock(); err != nil {
		return err
	} else {
		defer l.Unlock()
	}

	for k, v := range values {
		// The statement below takes a copy of K. This is necessary because K is
		// `var k [32]byte`, a fixed-length array, and arrays in go are
		// pass-by-value. This means that range variable K is overwritten on
		// each loop iteration. Without this statement, `k[:]` creates a slice
		// that points to the range variable, so every call to `txn.Set` gets a
		// slice pointing to the same memory. Since the transaction defers the
		// actual write until `txn.Commit` is called, it saves the slice. And
		// since all of the slices are pointing to the same variable, and that
		// variable is overwritten on each iteration, the slices held by `txn`
		// all point to the same value. When the transaction is committed, every
		// value is written to the last key. Taking a copy solves this because
		// each loop iteration creates a new copy, and `k[:]` references that
		// copy instead of the original. See also:
		// https://github.com/golang/go/wiki/CommonMistakes#using-reference-to-loop-iterator-variable
		k := k
		err := b.txn.Set(k[:], v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Batch) Get(key storage.Key) (v []byte, err error) {
	if l, err := b.db.lock(false); err != nil {
		return nil, err
	} else {
		defer l.Unlock()
	}

	item, err := b.txn.Get(key[:])
	if err != nil {
		// If we didn't find the value, return ErrNotFound
		if errors.Is(err, badger.ErrKeyNotFound) {
			err = storage.ErrNotFound
		}
		return nil, err
	}

	v, err = item.ValueCopy(nil)
	// If we didn't find the value, return ErrNotFound
	if errors.Is(err, badger.ErrKeyNotFound) {
		return nil, storage.ErrNotFound
	}

	return v, nil
}

func (b *Batch) Commit() error {
	if l, err := b.lock(); err != nil {
		return err
	} else {
		defer l.Unlock()
	}

	return b.txn.Commit()
}

func (b *Batch) Discard() {
	b.txn.Discard()
}
