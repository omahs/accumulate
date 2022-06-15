package database

// GENERATED BY go run ./tools/cmd/gen-record. DO NOT EDIT.

import (
	"gitlab.com/accumulatenetwork/accumulate/internal/database/record"
	"gitlab.com/accumulatenetwork/accumulate/internal/errors"
	"gitlab.com/accumulatenetwork/accumulate/internal/logging"
	"gitlab.com/accumulatenetwork/accumulate/smt/storage"
)

type ChangeSet struct {
	logger   logging.OptionalLogger
	store    record.Store
	done     bool
	writable bool
	id       string
	nextId   uint64
	kvStore  storage.KeyValueTxn

	entry map[storage.Key]*record.Wrapped[[]byte]
}

func (c *ChangeSet) Entry(key storage.Key) *record.Wrapped[[]byte] {
	return getOrCreateMap(&c.entry, record.Key{}.Append("Entry", key), func() *record.Wrapped[[]byte] {
		return record.NewWrapped(c.logger.L, c.store, record.Key{}.Append("Entry", key), "entry %[2]v", false, record.NewWrapper(record.BytesWrapper))
	})
}

func (c *ChangeSet) Resolve(key record.Key) (record.Record, record.Key, error) {
	switch key[0] {
	case "Entry":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for change set")
		}
		pKey, okKey := key[1].(storage.Key)
		if !okKey {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for change set")
		}
		v := c.Entry(pKey)
		return v, key[2:], nil
	default:
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for change set")
	}
}

func (c *ChangeSet) IsDirty() bool {
	if c == nil {
		return false
	}

	for _, v := range c.entry {
		if v.IsDirty() {
			return true
		}
	}

	return false
}

func (c *ChangeSet) baseCommit() error {
	if c == nil {
		return nil
	}

	var err error
	for _, v := range c.entry {
		commitField(&err, v)
	}

	return nil
}

func getOrCreateField[T any](ptr **T, create func() *T) *T {
	if *ptr != nil {
		return *ptr
	}

	*ptr = create()
	return *ptr
}

func getOrCreateMap[T any](ptr *map[storage.Key]*T, key record.Key, create func() *T) *T {
	if *ptr == nil {
		*ptr = map[storage.Key]*T{}
	}

	k := key.Hash()
	if v, ok := (*ptr)[k]; ok {
		return v
	}

	v := create()
	(*ptr)[k] = v
	return v
}

func commitField[T any, PT record.RecordPtr[T]](lastErr *error, field PT) {
	if *lastErr != nil || field == nil {
		return
	}

	*lastErr = field.Commit()
}

func fieldIsDirty[T any, PT record.RecordPtr[T]](field PT) bool {
	return field != nil && field.IsDirty()
}
