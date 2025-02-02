// Copyright 2023 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package database

import (
	"fmt"
	"sync/atomic"

	"gitlab.com/accumulatenetwork/accumulate/internal/database/record"
	"gitlab.com/accumulatenetwork/accumulate/internal/database/smt/storage"
	"gitlab.com/accumulatenetwork/accumulate/pkg/errors"
	"gitlab.com/accumulatenetwork/accumulate/pkg/url"
)

type Viewer interface {
	View(func(batch *Batch) error) error
}

type Updater interface {
	Viewer
	Update(func(batch *Batch) error) error
}

// A Beginner can be a Database or a Batch
type Beginner interface {
	Updater
	Begin(bool) *Batch
}

var _ Beginner = (*Database)(nil)
var _ Beginner = (*Batch)(nil)

// Begin starts a new batch.
func (d *Database) Begin(writable bool) *Batch {
	id := atomic.AddInt64(&d.nextBatchId, 1)

	b := new(Batch)
	b.id = fmt.Sprint(id)
	b.writable = writable
	b.logger.L = d.logger
	b.kvstore = d.store.Begin(writable)
	b.store = record.KvStore{Store: b.kvstore}
	b.bptEntries = map[storage.Key][32]byte{}
	return b
}

func (b *Batch) Begin(writable bool) *Batch {
	if writable && !b.writable {
		b.logger.Info("Attempted to create a writable batch from a read-only batch")
	}

	b.nextChildId++

	c := new(Batch)
	c.id = fmt.Sprintf("%s.%d", b.id, b.nextChildId)
	c.writable = b.writable && writable
	c.parent = b
	c.logger = b.logger
	c.store = b
	c.kvstore = b.kvstore.Begin(c.writable)
	c.bptEntries = map[storage.Key][32]byte{}
	return c
}

// DeleteAccountState_TESTONLY is intended for testing purposes only. It deletes an
// account from the database.
func (b *Batch) DeleteAccountState_TESTONLY(url *url.URL) error {
	a := record.Key{"Account", url, "Main"}
	return b.kvstore.Put(a.Hash(), nil)
}

// View runs the function with a read-only transaction.
func (d *Database) View(fn func(batch *Batch) error) error {
	batch := d.Begin(false)
	defer batch.Discard()
	return fn(batch)
}

// Update runs the function with a writable transaction and commits if the
// function succeeds.
func (d *Database) Update(fn func(batch *Batch) error) error {
	batch := d.Begin(true)
	defer batch.Discard()
	err := fn(batch)
	if err != nil {
		return err
	}
	return batch.Commit()
}

// View runs the function with a read-only transaction.
func (b *Batch) View(fn func(batch *Batch) error) error {
	batch := b.Begin(false)
	defer batch.Discard()
	return fn(batch)
}

// Update runs the function with a writable transaction and commits if the
// function succeeds.
func (b *Batch) Update(fn func(batch *Batch) error) error {
	batch := b.Begin(true)
	defer batch.Discard()
	err := fn(batch)
	if err != nil {
		return err
	}
	return batch.Commit()
}

// Commit commits pending writes to the key-value store or the parent batch.
// Attempting to use the Batch after calling Commit or Discard will result in a
// panic.
func (b *Batch) Commit() error {
	if b.done {
		panic(fmt.Sprintf("batch %s: attempted to use a committed or discarded batch", b.id))
	}
	defer func() { b.done = true }()

	err := b.baseCommit()
	if err != nil {
		return errors.UnknownError.Wrap(err)
	}

	if b.parent != nil {
		for k, v := range b.bptEntries {
			b.parent.bptEntries[k] = v
		}
		if db, ok := b.kvstore.(*storage.DebugBatch); ok {
			db.PretendWrite()
		}
	} else {
		err := b.commitBpt()
		if err != nil {
			return errors.UnknownError.Wrap(err)
		}
	}

	return b.kvstore.Commit()
}

// Discard discards pending writes. Attempting to use the Batch after calling
// Discard will result in a panic.
func (b *Batch) Discard() {
	if !b.done && b.writable {
		b.logger.Debug("Discarding a writable batch")
	}
	b.done = true
	b.kvstore.Discard()
}

// Transaction returns an Transaction for the given hash.
func (b *Batch) Transaction(id []byte) *Transaction {
	return b.getTransaction(*(*[32]byte)(id))
}

func (b *Batch) getAccountUrl(key record.Key) (*url.URL, error) {
	v, err := record.NewValue(
		b.logger.L,
		b.store,
		// This must match the key used for the account's Url state
		key.Append("Url"),
		fmt.Sprintf("account %v URL", key),
		false,
		record.Wrapped(record.UrlWrapper),
	).Get()
	if err != nil {
		return nil, errors.UnknownError.Wrap(err)
	}
	return v, nil
}

// AccountByID returns an Account for the given ID.
//
// This is still needed in one place, so the deprecation warning is disabled in
// order to pass static analysis.
//
// Deprecated: Use Account.
func (b *Batch) AccountByID(id []byte) (*Account, error) {
	u, err := b.getAccountUrl(record.Key{"Account", id})
	if err != nil {
		return nil, errors.UnknownError.Wrap(err)
	}
	return b.Account(u), nil
}

// GetValue implements record.Store.
func (b *Batch) GetValue(key record.Key, value record.ValueWriter) error {
	if b.done {
		panic(fmt.Sprintf("batch %s: attempted to use a committed or discarded batch", b.id))
	}

	v, err := resolveValue[record.ValueReader](b, key)
	if err != nil {
		return errors.UnknownError.Wrap(err)
	}

	err = value.LoadValue(v, false)
	return errors.UnknownError.Wrap(err)
}

// PutValue implements record.Store.
func (b *Batch) PutValue(key record.Key, value record.ValueReader) error {
	if b.done {
		panic(fmt.Sprintf("batch %s: attempted to use a committed or discarded batch", b.id))
	}

	v, err := resolveValue[record.ValueWriter](b, key)
	if err != nil {
		return errors.UnknownError.Wrap(err)
	}

	err = v.LoadValue(value, true)
	return errors.UnknownError.Wrap(err)
}

// resolveValue resolves the value for the given key.
func resolveValue[T any](c *Batch, key record.Key) (T, error) {
	var r record.Record = c
	var err error
	for len(key) > 0 {
		r, key, err = r.Resolve(key)
		if err != nil {
			return zero[T](), errors.UnknownError.Wrap(err)
		}
	}

	if s, _, err := r.Resolve(nil); err == nil {
		r = s
	}

	v, ok := r.(T)
	if !ok {
		return zero[T](), errors.InternalError.WithFormat("bad key: %T is not value", r)
	}

	return v, nil
}

// UpdatedAccounts returns every account updated in this database batch.
func (b *Batch) UpdatedAccounts() []*Account {
	accounts := make([]*Account, 0, len(b.account))
	for _, a := range b.account {
		if a.IsDirty() {
			accounts = append(accounts, a)
		}
	}
	return accounts
}
