package api

// GENERATED BY go run ./tools/cmd/gen-enum. DO NOT EDIT.

import (
	"encoding/json"
	"fmt"
	"strings"
)

// EventTypeBlock .
const EventTypeBlock EventType = 1

// RecordTypeAccount .
const RecordTypeAccount RecordType = 1

// RecordTypeChain .
const RecordTypeChain RecordType = 2

// RecordTypeChainEntry .
const RecordTypeChainEntry RecordType = 3

// RecordTypeTransaction .
const RecordTypeTransaction RecordType = 16

// RecordTypeSignature .
const RecordTypeSignature RecordType = 17

// RecordTypeRange .
const RecordTypeRange RecordType = 128

// RecordTypeUrl .
const RecordTypeUrl RecordType = 129

// RecordTypeTxID .
const RecordTypeTxID RecordType = 130

// RecordTypeIndexEntry .
const RecordTypeIndexEntry RecordType = 131

// GetEnumValue returns the value of the Event Type
func (v EventType) GetEnumValue() uint64 { return uint64(v) }

// SetEnumValue sets the value. SetEnumValue returns false if the value is invalid.
func (v *EventType) SetEnumValue(id uint64) bool {
	u := EventType(id)
	switch u {
	case EventTypeBlock:
		*v = u
		return true
	default:
		return false
	}
}

// String returns the name of the Event Type.
func (v EventType) String() string {
	switch v {
	case EventTypeBlock:
		return "block"
	default:
		return fmt.Sprintf("EventType:%d", v)
	}
}

// EventTypeByName returns the named Event Type.
func EventTypeByName(name string) (EventType, bool) {
	switch strings.ToLower(name) {
	case "block":
		return EventTypeBlock, true
	default:
		return 0, false
	}
}

// MarshalJSON marshals the Event Type to JSON as a string.
func (v EventType) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

// UnmarshalJSON unmarshals the Event Type from JSON as a string.
func (v *EventType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	var ok bool
	*v, ok = EventTypeByName(s)
	if !ok || strings.ContainsRune(v.String(), ':') {
		return fmt.Errorf("invalid Event Type %q", s)
	}
	return nil
}

// GetEnumValue returns the value of the Record Type
func (v RecordType) GetEnumValue() uint64 { return uint64(v) }

// SetEnumValue sets the value. SetEnumValue returns false if the value is invalid.
func (v *RecordType) SetEnumValue(id uint64) bool {
	u := RecordType(id)
	switch u {
	case RecordTypeAccount, RecordTypeChain, RecordTypeChainEntry, RecordTypeTransaction, RecordTypeSignature, RecordTypeRange, RecordTypeUrl, RecordTypeTxID, RecordTypeIndexEntry:
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
	case RecordTypeChain:
		return "chain"
	case RecordTypeChainEntry:
		return "chainEntry"
	case RecordTypeTransaction:
		return "transaction"
	case RecordTypeSignature:
		return "signature"
	case RecordTypeRange:
		return "range"
	case RecordTypeUrl:
		return "url"
	case RecordTypeTxID:
		return "txID"
	case RecordTypeIndexEntry:
		return "indexEntry"
	default:
		return fmt.Sprintf("RecordType:%d", v)
	}
}

// RecordTypeByName returns the named Record Type.
func RecordTypeByName(name string) (RecordType, bool) {
	switch strings.ToLower(name) {
	case "account":
		return RecordTypeAccount, true
	case "chain":
		return RecordTypeChain, true
	case "chainentry":
		return RecordTypeChainEntry, true
	case "transaction":
		return RecordTypeTransaction, true
	case "signature":
		return RecordTypeSignature, true
	case "range":
		return RecordTypeRange, true
	case "url":
		return RecordTypeUrl, true
	case "txid":
		return RecordTypeTxID, true
	case "indexentry":
		return RecordTypeIndexEntry, true
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
