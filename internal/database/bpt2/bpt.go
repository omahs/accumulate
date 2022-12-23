package bpt2

import (
	"sort"

	"github.com/tendermint/tendermint/libs/log"
	"gitlab.com/accumulatenetwork/accumulate/internal/database/record"
	"gitlab.com/accumulatenetwork/accumulate/pkg/errors"
)

func New(logger log.Logger, store record.Store, key record.Key, power uint64, label string) *BPT {
	b := new(BPT)
	b.logger.Set(logger)
	b.store = store
	b.key = key
	b.label = label
	b.power = 1 << power
	b.mask = b.power - 1
	return b
}

func (b *BPT) Insert(key, hash [32]byte) error {
	_, err := b.getRoot().insert(&Value{Key: key, Hash: hash})
	return err
}

func (b *BPT) getRoot() *Node {
	k, _ := nodeKeyAt(0, [32]byte{})
	if n, ok := b.node[bptnodeKey{k}]; ok {
		return n
	}
	return newRootNode(b)
}

func (b *BPT) GetHash() ([]byte, error) {
	err := b.getRoot().load()
	if err != nil {
		return nil, errors.UnknownError.Wrap(err)
	}
	return b.getRoot().GetHash(), nil
}

func (b *BPT) Resolve(key record.Key) (record.Record, record.Key, error) {
	if len(key) == 0 {
		return nil, nil, errors.InternalError.With("bad key for bpt")
	}

	if key[0] == "Root" {
		return b.getState(), key[1:], nil
	}

	nodeKey, ok := key[0].([32]byte)
	if !ok {
		return nil, nil, errors.InternalError.With("bad key for bpt")
	}

	height, k, ok := parseNodeKey(nodeKey)
	if !ok {
		return nil, nil, errors.InternalError.With("bad key for bpt")
	}
	n, err := b.getRoot().getAt(height, k)
	if err != nil {
		return nil, nil, errors.UnknownError.Wrap(err)
	}
	return n, key[1:], nil
}

func (b *BPT) Commit() error {
	// Update hashes
	b.getRoot().GetHash()

	// Until all nodes are clean
	for {
		// Get a list of all dirty nodes
		var dirty []*Node
		for _, n := range b.node {
			if n.IsDirty() {
				dirty = append(dirty, n)
			}
		}
		if len(dirty) == 0 {
			break
		}

		// Order by height
		sort.Slice(dirty, func(i, j int) bool {
			return dirty[i].Height > dirty[j].Height
		})

		// Commit only nodes at the highest level
		h := dirty[0].Height
		for _, n := range dirty {
			if n.Height < h {
				break
			}
			err := n.Commit()
			if err != nil {
				return errors.UnknownError.Wrap(err)
			}
		}
	}

	// Set the root hash (so we don't have to load Root)
	s, err := b.getState().Get()
	if err != nil {
		return errors.UnknownError.WithFormat("load state: %w", err)
	}
	s.RootHash = *(*[32]byte)(b.getRoot().GetHash())
	s.Power = b.power
	s.Mask = b.mask
	err = b.getState().Put(s)
	if err != nil {
		return errors.UnknownError.WithFormat("store state: %w", err)
	}

	return b.baseCommit()
}
