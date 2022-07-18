package managed

// GENERATED BY go run ./tools/cmd/gen-model. DO NOT EDIT.

//lint:file-ignore S1008,U1000 generated code

import (
	"gitlab.com/accumulatenetwork/accumulate/internal/database/record"
	"gitlab.com/accumulatenetwork/accumulate/internal/errors"
	"gitlab.com/accumulatenetwork/accumulate/internal/logging"
	"gitlab.com/accumulatenetwork/accumulate/smt/storage"
)

type Chain struct {
	logger    logging.OptionalLogger
	store     record.Store
	key       record.Key
	label     string
	typ       ChainType
	name      string
	markPower int64
	markFreq  int64
	markMask  int64

	head         *record.Value[*MerkleState]
	states       map[storage.Key]*record.Value[*MerkleState]
	elementIndex map[storage.Key]*record.Value[uint64]
	element      map[storage.Key]*record.Value[[]byte]
}

func (c *Chain) Head() *record.Value[*MerkleState] {
	return getOrCreateField(&c.head, func() *record.Value[*MerkleState] {
		return record.NewValue(c.logger.L, c.store, c.key.Append("Head"), c.label+" head", true, record.Struct[MerkleState]())
	})
}

func (c *Chain) States(index uint64) *record.Value[*MerkleState] {
	return getOrCreateMap(&c.states, c.key.Append("States", index), func() *record.Value[*MerkleState] {
		return record.NewValue(c.logger.L, c.store, c.key.Append("States", index), c.label+" states %[3]v", false, record.Struct[MerkleState]())
	})
}

func (c *Chain) ElementIndex(hash []byte) *record.Value[uint64] {
	return getOrCreateMap(&c.elementIndex, c.key.Append("ElementIndex", hash), func() *record.Value[uint64] {
		return record.NewValue(c.logger.L, c.store, c.key.Append("ElementIndex", hash), c.label+" element index %[3]x", false, record.Wrapped(record.UintWrapper))
	})
}

func (c *Chain) Element(index uint64) *record.Value[[]byte] {
	return getOrCreateMap(&c.element, c.key.Append("Element", index), func() *record.Value[[]byte] {
		return record.NewValue(c.logger.L, c.store, c.key.Append("Element", index), c.label+" element %[3]v", false, record.Wrapped(record.BytesWrapper))
	})
}

func (c *Chain) Resolve(key record.Key) (record.Record, record.Key, error) {
	switch key[0] {
	case "Head":
		return c.Head(), key[1:], nil
	case "States":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for chain")
		}
		index, okIndex := key[1].(uint64)
		if !okIndex {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for chain")
		}
		v := c.States(index)
		return v, key[2:], nil
	case "ElementIndex":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for chain")
		}
		hash, okHash := key[1].([]byte)
		if !okHash {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for chain")
		}
		v := c.ElementIndex(hash)
		return v, key[2:], nil
	case "Element":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for chain")
		}
		index, okIndex := key[1].(uint64)
		if !okIndex {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for chain")
		}
		v := c.Element(index)
		return v, key[2:], nil
	default:
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for chain")
	}
}

func (c *Chain) IsDirty() bool {
	if c == nil {
		return false
	}

	if fieldIsDirty(c.head) {
		return true
	}
	for _, v := range c.states {
		if v.IsDirty() {
			return true
		}
	}
	for _, v := range c.elementIndex {
		if v.IsDirty() {
			return true
		}
	}
	for _, v := range c.element {
		if v.IsDirty() {
			return true
		}
	}

	return false
}

func (c *Chain) Commit() error {
	if c == nil {
		return nil
	}

	var err error
	commitField(&err, c.head)
	for _, v := range c.states {
		commitField(&err, v)
	}
	for _, v := range c.elementIndex {
		commitField(&err, v)
	}
	for _, v := range c.element {
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

func getOrCreateMap[T any](ptr *map[storage.Key]T, key record.Key, create func() T) T {
	if *ptr == nil {
		*ptr = map[storage.Key]T{}
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
