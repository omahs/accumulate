// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package hash

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gitlab.com/accumulatenetwork/accumulate/internal/database/smt/storage"
)

func getHash(v ...interface{}) [32]byte {
	return storage.MakeKey(v...)
}

func TestHasher_Receipt(t *testing.T) {
	e0 := getHash(0)
	e1 := getHash(1)
	e2 := getHash(2)
	e3 := getHash(3)
	e4 := getHash(4)
	e5 := getHash(5)
	e6 := getHash(6)
	e01 := sha256.Sum256(append(e0[:], e1[:]...))
	e23 := sha256.Sum256(append(e2[:], e3[:]...))
	e45 := sha256.Sum256(append(e4[:], e5[:]...))
	e0123 := sha256.Sum256(append(e01[:], e23[:]...))

	fmt.Printf("%-7s %x %v\n", "e01", e01, e01[:3])
	fmt.Printf("%-7s %x %v\n", "e23", e23, e23[:3])
	fmt.Printf("%-7s %x %v\n", "e45", e45, e45[:3])
	fmt.Printf("%-7s %x %v\n", "e0123", e0123, e0123[:3])

	fmt.Print("\n\n")
	fmt.Printf("%x  %x  %x  %x  %x  %x  %x\n", e0[:3], e1[:3], e2[:3], e3[:3], e4[:3], e5[:3], e6[:3])
	fmt.Printf("        %x          %x          %x\n", e01[:3], e23[:3], e45[:3])
	fmt.Printf("                        %x\n", e0123[:3])

	hasher := make(Hasher, 0, 7)
	hasher.AddHash(&e0)
	hasher.AddHash(&e1)
	hasher.AddHash(&e2)
	hasher.AddHash(&e3)
	hasher.AddHash(&e4)
	hasher.AddHash(&e5)
	hasher.AddHash(&e6)

	for i := 0; i < 7; i++ {
		for j := i; j < 7; j++ {
			r := hasher.Receipt(i, j)
			// fmt.Println(r.String())
			require.Equal(t, hasher[i], r.Start)
			assert.Truef(t, r.Validate(), "Receipt fails for %d to %d", i, j)
			fmt.Printf("Build receipt from %d to %d\n", i, j)
		}
	}
}
