package api

// GENERATED BY go run ./tools/cmd/gen-types. DO NOT EDIT.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"gitlab.com/accumulatenetwork/accumulate/internal/core/encoding"
)

// NewRecord creates a new Record for the specified RecordType.
func NewRecord(typ RecordType) (Record, error) {
	switch typ {
	case RecordTypeAccount:
		return new(AccountRecord), nil
	default:
		return nil, fmt.Errorf("unknown record %v", typ)
	}
}

//EqualRecord is used to compare the values of the union
func EqualRecord(a, b Record) bool {
	if a == b {
		return true
	}
	switch a := a.(type) {
	case *AccountRecord:
		b, ok := b.(*AccountRecord)
		return ok && a.Equal(b)
	default:
		return false
	}
}

// UnmarshalRecordType unmarshals the RecordType from the start of a Record.
func UnmarshalRecordType(r io.Reader) (RecordType, error) {
	var typ RecordType
	err := encoding.UnmarshalEnumType(r, &typ)
	return typ, err
}

// UnmarshalRecord unmarshals a Record.
func UnmarshalRecord(data []byte) (Record, error) {
	typ, err := UnmarshalRecordType(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	v, err := NewRecord(typ)
	if err != nil {
		return nil, err
	}

	err = v.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// UnmarshalRecordFrom unmarshals a Record.
func UnmarshalRecordFrom(rd io.ReadSeeker) (Record, error) {
	// Get the reader's current position
	pos, err := rd.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}

	// Read the type code
	typ, err := UnmarshalRecordType(rd)
	if err != nil {
		return nil, err
	}

	// Reset the reader's position
	_, err = rd.Seek(pos, io.SeekStart)
	if err != nil {
		return nil, err
	}

	// Create a new transaction result
	v, err := NewRecord(RecordType(typ))
	if err != nil {
		return nil, err
	}

	// Unmarshal the result
	err = v.UnmarshalBinaryFrom(rd)
	if err != nil {
		return nil, err
	}

	return v, nil
}

// UnmarshalRecordJson unmarshals a Record.
func UnmarshalRecordJSON(data []byte) (Record, error) {
	var typ *struct{ Type RecordType }
	err := json.Unmarshal(data, &typ)
	if err != nil {
		return nil, err
	}

	if typ == nil {
		return nil, nil
	}

	acnt, err := NewRecord(typ.Type)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, acnt)
	if err != nil {
		return nil, err
	}

	return acnt, nil
}
