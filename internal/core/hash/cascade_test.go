// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package hash_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/accumulatenetwork/accumulate/internal/core/hash"
	"gitlab.com/accumulatenetwork/accumulate/test/testdata"
	acctesting "gitlab.com/accumulatenetwork/accumulate/test/testing"
	"gopkg.in/yaml.v3"
)

func TestMerkleCascade(t *testing.T) {
	var cases []*acctesting.MerkleTestCase
	require.NoError(t, yaml.Unmarshal([]byte(testdata.Merkle), &cases))

	for _, c := range cases {
		var entries [][]byte
		for _, e := range c.Entries {
			entries = append(entries, e)
		}
		var result [][]byte
		for _, e := range c.Cascade {
			result = append(result, e)
		}
		t.Run(fmt.Sprintf("%X", c.Root[:4]), func(t *testing.T) {
			cascade := hash.MerkleCascade(nil, entries, -1)
			require.Equal(t, result, cascade)
		})
	}
}
