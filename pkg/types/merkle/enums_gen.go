// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package merkle

// GENERATED BY go run ./tools/cmd/gen-enum. DO NOT EDIT.

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ChainTypeUnknown is used when the chain type is not known.
const ChainTypeUnknown ChainType = 0

// ChainTypeTransaction holds transaction hashes.
const ChainTypeTransaction ChainType = 1

// ChainTypeAnchor holds chain anchors.
const ChainTypeAnchor ChainType = 2

// ChainTypeIndex indexes other chains.
const ChainTypeIndex ChainType = 4

// GetEnumValue returns the value of the Chain Type
func (v ChainType) GetEnumValue() uint64 { return uint64(v) }

// SetEnumValue sets the value. SetEnumValue returns false if the value is invalid.
func (v *ChainType) SetEnumValue(id uint64) bool {
	u := ChainType(id)
	switch u {
	case ChainTypeUnknown, ChainTypeTransaction, ChainTypeAnchor, ChainTypeIndex:
		*v = u
		return true
	default:
		return false
	}
}

// String returns the name of the Chain Type.
func (v ChainType) String() string {
	switch v {
	case ChainTypeUnknown:
		return "unknown"
	case ChainTypeTransaction:
		return "transaction"
	case ChainTypeAnchor:
		return "anchor"
	case ChainTypeIndex:
		return "index"
	default:
		return fmt.Sprintf("ChainType:%d", v)
	}
}

// ChainTypeByName returns the named Chain Type.
func ChainTypeByName(name string) (ChainType, bool) {
	switch strings.ToLower(name) {
	case "unknown":
		return ChainTypeUnknown, true
	case "transaction":
		return ChainTypeTransaction, true
	case "anchor":
		return ChainTypeAnchor, true
	case "index":
		return ChainTypeIndex, true
	default:
		return 0, false
	}
}

// MarshalJSON marshals the Chain Type to JSON as a string.
func (v ChainType) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

// UnmarshalJSON unmarshals the Chain Type from JSON as a string.
func (v *ChainType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	var ok bool
	*v, ok = ChainTypeByName(s)
	if !ok || strings.ContainsRune(v.String(), ':') {
		return fmt.Errorf("invalid Chain Type %q", s)
	}
	return nil
}
