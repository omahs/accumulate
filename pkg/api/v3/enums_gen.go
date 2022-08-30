package api

// GENERATED BY go run ./tools/cmd/gen-enum. DO NOT EDIT.

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RecordTypeAccount .
const RecordTypeAccount RecordType = 1

// RecordTypeTransaction .
const RecordTypeTransaction RecordType = 2

// RecordTypeChain .
const RecordTypeChain RecordType = 3

// GetEnumValue returns the value of the Record Type
func (v RecordType) GetEnumValue() uint64 { return uint64(v) }

// SetEnumValue sets the value. SetEnumValue returns false if the value is invalid.
func (v *RecordType) SetEnumValue(id uint64) bool {
	u := RecordType(id)
	switch u {
	case RecordTypeAccount, RecordTypeTransaction, RecordTypeChain:
		*v = u
		return true
	default:
		return false
	}
}

// String returns the name of the Record Type.
func (v RecordType) String() string {
	switch v {
	case RecordTypeAccount:
		return "account"
	case RecordTypeTransaction:
		return "transaction"
	case RecordTypeChain:
		return "chain"
	default:
		return fmt.Sprintf("RecordType:%d", v)
	}
}

// RecordTypeByName returns the named Record Type.
func RecordTypeByName(name string) (RecordType, bool) {
	switch strings.ToLower(name) {
	case "account":
		return RecordTypeAccount, true
	case "transaction":
		return RecordTypeTransaction, true
	case "chain":
		return RecordTypeChain, true
	default:
		return 0, false
	}
}

// MarshalJSON marshals the Record Type to JSON as a string.
func (v RecordType) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

// UnmarshalJSON unmarshals the Record Type from JSON as a string.
func (v *RecordType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	var ok bool
	*v, ok = RecordTypeByName(s)
	if !ok || strings.ContainsRune(v.String(), ':') {
		return fmt.Errorf("invalid Record Type %q", s)
	}
	return nil
}
