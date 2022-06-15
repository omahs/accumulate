package record_test

// GENERATED BY go run ./tools/cmd/gen-model. DO NOT EDIT.

import (
	"gitlab.com/accumulatenetwork/accumulate/internal/database/record"
	"gitlab.com/accumulatenetwork/accumulate/internal/errors"
	"gitlab.com/accumulatenetwork/accumulate/internal/logging"
	"gitlab.com/accumulatenetwork/accumulate/internal/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
	"gitlab.com/accumulatenetwork/accumulate/smt/managed"
	"gitlab.com/accumulatenetwork/accumulate/smt/storage"
)

type ChangeSet struct {
	logger logging.OptionalLogger
	store  record.Store

	entity    map[storage.Key]*Entity
	changeLog *record.Counted[string]
}

func (c *ChangeSet) Entity(name string) *Entity {
	return getOrCreateMap(&c.entity, record.Key{}.Append("Entity", name), func() *Entity {
		v := new(Entity)
		v.logger = c.logger
		v.store = c.store
		v.key = record.Key{}.Append("Entity", name)
		v.parent = c
		return v
	})
}

func (c *ChangeSet) ChangeLog() *record.Counted[string] {
	return getOrCreateField(&c.changeLog, func() *record.Counted[string] {

		return record.NewCounted(c.logger.L, c.store, record.Key{}.Append("ChangeLog"), "change log", record.NewCountableWrapped(record.StringWrapper))
	})
}

func (c *ChangeSet) Resolve(key record.Key) (record.Record, record.Key, error) {
	switch key[0] {
	case "Entity":
		if len(key) < 2 {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for change set")
		}
		name, okName := key[1].(string)
		if !okName {
			return nil, nil, errors.New(errors.StatusInternalError, "bad key for change set")
		}
		v := c.Entity(name)
		return v, key[2:], nil
	case "ChangeLog":
		return c.ChangeLog(), key[1:], nil
	default:
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for change set")
	}
}

func (c *ChangeSet) IsDirty() bool {
	if c == nil {
		return false
	}

	for _, v := range c.entity {
		if v.IsDirty() {
			return true
		}
	}
	if fieldIsDirty(c.changeLog) {
		return true
	}

	return false
}

func (c *ChangeSet) dirtyChains() []*managed.Chain {
	if c == nil {
		return nil
	}

	var chains []*managed.Chain

	for _, v := range c.entity {
		chains = append(chains, v.dirtyChains()...)
	}

	return chains
}

func (c *ChangeSet) Commit() error {
	if c == nil {
		return nil
	}

	var err error
	for _, v := range c.entity {
		commitField(&err, v)
	}
	commitField(&err, c.changeLog)

	return nil
}

type Entity struct {
	logger logging.OptionalLogger
	store  record.Store
	key    record.Key
	parent *ChangeSet

	union *record.Wrapped[protocol.Account]
	set   *record.Set[*url.TxID]
	chain *managed.Chain
}

func (c *Entity) Union() *record.Wrapped[protocol.Account] {
	return getOrCreateField(&c.union, func() *record.Wrapped[protocol.Account] {
		return record.NewWrapped(c.logger.L, c.store, c.key.Append("Union"), "entity %[2]v union", false, record.NewWrapper(record.UnionWrapper(protocol.UnmarshalAccount)))
	})
}

func (c *Entity) Set() *record.Set[*url.TxID] {
	return getOrCreateField(&c.set, func() *record.Set[*url.TxID] {
		return record.NewSet(c.logger.L, c.store, c.key.Append("Set"), "entity %[2]v set", record.NewWrapperSlice(record.TxidWrapper), record.CompareTxid)
	})
}

func (c *Entity) Chain() *managed.Chain {
	return getOrCreateField(&c.chain, func() *managed.Chain {
		return managed.NewChain(c.logger.L, c.store, c.key.Append("Chain"), markPower, "entity %[2]v chain")
	})
}

func (c *Entity) Resolve(key record.Key) (record.Record, record.Key, error) {
	switch key[0] {
	case "Union":
		return c.Union(), key[1:], nil
	case "Set":
		return c.Set(), key[1:], nil
	case "Chain":
		return c.Chain(), key[1:], nil
	default:
		return nil, nil, errors.New(errors.StatusInternalError, "bad key for entity")
	}
}

func (c *Entity) IsDirty() bool {
	if c == nil {
		return false
	}

	if fieldIsDirty(c.union) {
		return true
	}
	if fieldIsDirty(c.set) {
		return true
	}
	if fieldIsDirty(c.chain) {
		return true
	}

	return false
}

func (c *Entity) resolveChain(name string) (chain *managed.Chain, ok bool) {
	if name == "" {
		return c.Chain(), true
	}
	return
}

func (c *Entity) dirtyChains() []*managed.Chain {
	if c == nil {
		return nil
	}

	var chains []*managed.Chain

	if fieldIsDirty(c.chain) {
		chains = append(chains, c.chain)
	}

	return chains
}

func (c *Entity) baseCommit() error {
	if c == nil {
		return nil
	}

	var err error
	commitField(&err, c.union)
	commitField(&err, c.set)
	commitField(&err, c.chain)

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
