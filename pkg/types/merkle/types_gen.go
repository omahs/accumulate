// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package merkle

// GENERATED BY go run ./tools/cmd/gen-types. DO NOT EDIT.

//lint:file-ignore S1001,S1002,S1008,SA4013 generated code

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"gitlab.com/accumulatenetwork/accumulate/pkg/types/encoding"
)

type Receipt struct {
	fieldsSet []bool
	// Start is the entry for which we want a proof.
	Start      []byte `json:"start,omitempty" form:"start" query:"start" validate:"required"`
	StartIndex int64  `json:"startIndex,omitempty" form:"startIndex" query:"startIndex" validate:"required"`
	// End is the entry at the index where the anchor was created.
	End      []byte `json:"end,omitempty" form:"end" query:"end" validate:"required"`
	EndIndex int64  `json:"endIndex,omitempty" form:"endIndex" query:"endIndex" validate:"required"`
	// Anchor is the root expected once all nodes are applied.
	Anchor []byte `json:"anchor,omitempty" form:"anchor" query:"anchor" validate:"required"`
	// Entries is the list of hashes to apply to create an anchor.
	Entries   []*ReceiptEntry `json:"entries,omitempty" form:"entries" query:"entries" validate:"required"`
	extraData []byte
}

type ReceiptEntry struct {
	fieldsSet []bool
	Right     bool   `json:"right,omitempty" form:"right" query:"right" validate:"required"`
	Hash      []byte `json:"hash,omitempty" form:"hash" query:"hash" validate:"required"`
	extraData []byte
}

type State struct {

	// Count is the count of hashes added to the tree.
	Count int64 `json:"count,omitempty" form:"count" query:"count" validate:"required"`
	// Pending is the hashes that represent the left edge of the tree.
	Pending [][]byte `json:"pending,omitempty" form:"pending" query:"pending" validate:"required"`
	// HashList is the hashes added to the tree.
	HashList [][]byte `json:"hashList,omitempty" form:"hashList" query:"hashList" validate:"required"`
}

func (v *Receipt) Copy() *Receipt {
	u := new(Receipt)

	u.Start = encoding.BytesCopy(v.Start)
	u.StartIndex = v.StartIndex
	u.End = encoding.BytesCopy(v.End)
	u.EndIndex = v.EndIndex
	u.Anchor = encoding.BytesCopy(v.Anchor)
	u.Entries = make([]*ReceiptEntry, len(v.Entries))
	for i, v := range v.Entries {
		if v != nil {
			u.Entries[i] = (v).Copy()
		}
	}

	return u
}

func (v *Receipt) CopyAsInterface() interface{} { return v.Copy() }

func (v *ReceiptEntry) Copy() *ReceiptEntry {
	u := new(ReceiptEntry)

	u.Right = v.Right
	u.Hash = encoding.BytesCopy(v.Hash)

	return u
}

func (v *ReceiptEntry) CopyAsInterface() interface{} { return v.Copy() }

func (v *State) Copy() *State {
	u := new(State)

	u.Count = v.Count
	u.Pending = make([][]byte, len(v.Pending))
	for i, v := range v.Pending {
		u.Pending[i] = encoding.BytesCopy(v)
	}
	u.HashList = make([][]byte, len(v.HashList))
	for i, v := range v.HashList {
		u.HashList[i] = encoding.BytesCopy(v)
	}

	return u
}

func (v *State) CopyAsInterface() interface{} { return v.Copy() }

func (v *Receipt) Equal(u *Receipt) bool {
	if !(bytes.Equal(v.Start, u.Start)) {
		return false
	}
	if !(v.StartIndex == u.StartIndex) {
		return false
	}
	if !(bytes.Equal(v.End, u.End)) {
		return false
	}
	if !(v.EndIndex == u.EndIndex) {
		return false
	}
	if !(bytes.Equal(v.Anchor, u.Anchor)) {
		return false
	}
	if len(v.Entries) != len(u.Entries) {
		return false
	}
	for i := range v.Entries {
		if !((v.Entries[i]).Equal(u.Entries[i])) {
			return false
		}
	}

	return true
}

func (v *ReceiptEntry) Equal(u *ReceiptEntry) bool {
	if !(v.Right == u.Right) {
		return false
	}
	if !(bytes.Equal(v.Hash, u.Hash)) {
		return false
	}

	return true
}

var fieldNames_Receipt = []string{
	1: "Start",
	2: "StartIndex",
	3: "End",
	4: "EndIndex",
	5: "Anchor",
	6: "Entries",
}

func (v *Receipt) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(len(v.Start) == 0) {
		writer.WriteBytes(1, v.Start)
	}
	if !(v.StartIndex == 0) {
		writer.WriteInt(2, v.StartIndex)
	}
	if !(len(v.End) == 0) {
		writer.WriteBytes(3, v.End)
	}
	if !(v.EndIndex == 0) {
		writer.WriteInt(4, v.EndIndex)
	}
	if !(len(v.Anchor) == 0) {
		writer.WriteBytes(5, v.Anchor)
	}
	if !(len(v.Entries) == 0) {
		for _, v := range v.Entries {
			writer.WriteValue(6, v.MarshalBinary)
		}
	}

	_, _, err := writer.Reset(fieldNames_Receipt)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *Receipt) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 0 && !v.fieldsSet[0] {
		errs = append(errs, "field Start is missing")
	} else if len(v.Start) == 0 {
		errs = append(errs, "field Start is not set")
	}
	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field StartIndex is missing")
	} else if v.StartIndex == 0 {
		errs = append(errs, "field StartIndex is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field End is missing")
	} else if len(v.End) == 0 {
		errs = append(errs, "field End is not set")
	}
	if len(v.fieldsSet) > 3 && !v.fieldsSet[3] {
		errs = append(errs, "field EndIndex is missing")
	} else if v.EndIndex == 0 {
		errs = append(errs, "field EndIndex is not set")
	}
	if len(v.fieldsSet) > 4 && !v.fieldsSet[4] {
		errs = append(errs, "field Anchor is missing")
	} else if len(v.Anchor) == 0 {
		errs = append(errs, "field Anchor is not set")
	}
	if len(v.fieldsSet) > 5 && !v.fieldsSet[5] {
		errs = append(errs, "field Entries is missing")
	} else if len(v.Entries) == 0 {
		errs = append(errs, "field Entries is not set")
	}

	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errors.New(errs[0])
	default:
		return errors.New(strings.Join(errs, "; "))
	}
}

var fieldNames_ReceiptEntry = []string{
	1: "Right",
	2: "Hash",
}

func (v *ReceiptEntry) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(!v.Right) {
		writer.WriteBool(1, v.Right)
	}
	if !(len(v.Hash) == 0) {
		writer.WriteBytes(2, v.Hash)
	}

	_, _, err := writer.Reset(fieldNames_ReceiptEntry)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *ReceiptEntry) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 0 && !v.fieldsSet[0] {
		errs = append(errs, "field Right is missing")
	} else if !v.Right {
		errs = append(errs, "field Right is not set")
	}
	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Hash is missing")
	} else if len(v.Hash) == 0 {
		errs = append(errs, "field Hash is not set")
	}

	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errors.New(errs[0])
	default:
		return errors.New(strings.Join(errs, "; "))
	}
}

func (v *Receipt) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *Receipt) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x, ok := reader.ReadBytes(1); ok {
		v.Start = x
	}
	if x, ok := reader.ReadInt(2); ok {
		v.StartIndex = x
	}
	if x, ok := reader.ReadBytes(3); ok {
		v.End = x
	}
	if x, ok := reader.ReadInt(4); ok {
		v.EndIndex = x
	}
	if x, ok := reader.ReadBytes(5); ok {
		v.Anchor = x
	}
	for {
		if x := new(ReceiptEntry); reader.ReadValue(6, x.UnmarshalBinaryFrom) {
			v.Entries = append(v.Entries, x)
		} else {
			break
		}
	}

	seen, err := reader.Reset(fieldNames_Receipt)
	if err != nil {
		return encoding.Error{E: err}
	}
	v.fieldsSet = seen
	v.extraData, err = reader.ReadAll()
	if err != nil {
		return encoding.Error{E: err}
	}
	return nil
}

func (v *ReceiptEntry) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *ReceiptEntry) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x, ok := reader.ReadBool(1); ok {
		v.Right = x
	}
	if x, ok := reader.ReadBytes(2); ok {
		v.Hash = x
	}

	seen, err := reader.Reset(fieldNames_ReceiptEntry)
	if err != nil {
		return encoding.Error{E: err}
	}
	v.fieldsSet = seen
	v.extraData, err = reader.ReadAll()
	if err != nil {
		return encoding.Error{E: err}
	}
	return nil
}

func (v *Receipt) MarshalJSON() ([]byte, error) {
	u := struct {
		Start      *string                          `json:"start,omitempty"`
		StartIndex int64                            `json:"startIndex,omitempty"`
		End        *string                          `json:"end,omitempty"`
		EndIndex   int64                            `json:"endIndex,omitempty"`
		Anchor     *string                          `json:"anchor,omitempty"`
		Entries    encoding.JsonList[*ReceiptEntry] `json:"entries,omitempty"`
	}{}
	if !(len(v.Start) == 0) {
		u.Start = encoding.BytesToJSON(v.Start)
	}
	if !(v.StartIndex == 0) {
		u.StartIndex = v.StartIndex
	}
	if !(len(v.End) == 0) {
		u.End = encoding.BytesToJSON(v.End)
	}
	if !(v.EndIndex == 0) {
		u.EndIndex = v.EndIndex
	}
	if !(len(v.Anchor) == 0) {
		u.Anchor = encoding.BytesToJSON(v.Anchor)
	}
	if !(len(v.Entries) == 0) {
		u.Entries = v.Entries
	}
	return json.Marshal(&u)
}

func (v *ReceiptEntry) MarshalJSON() ([]byte, error) {
	u := struct {
		Right bool    `json:"right,omitempty"`
		Hash  *string `json:"hash,omitempty"`
	}{}
	if !(!v.Right) {
		u.Right = v.Right
	}
	if !(len(v.Hash) == 0) {
		u.Hash = encoding.BytesToJSON(v.Hash)
	}
	return json.Marshal(&u)
}

func (v *State) MarshalJSON() ([]byte, error) {
	u := struct {
		Count    int64                      `json:"count,omitempty"`
		Pending  encoding.JsonList[*string] `json:"pending,omitempty"`
		HashList encoding.JsonList[*string] `json:"hashList,omitempty"`
	}{}
	if !(v.Count == 0) {
		u.Count = v.Count
	}
	if !(len(v.Pending) == 0) {
		u.Pending = make(encoding.JsonList[*string], len(v.Pending))
		for i, x := range v.Pending {
			u.Pending[i] = encoding.BytesToJSON(x)
		}
	}
	if !(len(v.HashList) == 0) {
		u.HashList = make(encoding.JsonList[*string], len(v.HashList))
		for i, x := range v.HashList {
			u.HashList[i] = encoding.BytesToJSON(x)
		}
	}
	return json.Marshal(&u)
}

func (v *Receipt) UnmarshalJSON(data []byte) error {
	u := struct {
		Start      *string                          `json:"start,omitempty"`
		StartIndex int64                            `json:"startIndex,omitempty"`
		End        *string                          `json:"end,omitempty"`
		EndIndex   int64                            `json:"endIndex,omitempty"`
		Anchor     *string                          `json:"anchor,omitempty"`
		Entries    encoding.JsonList[*ReceiptEntry] `json:"entries,omitempty"`
	}{}
	u.Start = encoding.BytesToJSON(v.Start)
	u.StartIndex = v.StartIndex
	u.End = encoding.BytesToJSON(v.End)
	u.EndIndex = v.EndIndex
	u.Anchor = encoding.BytesToJSON(v.Anchor)
	u.Entries = v.Entries
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	if x, err := encoding.BytesFromJSON(u.Start); err != nil {
		return fmt.Errorf("error decoding Start: %w", err)
	} else {
		v.Start = x
	}
	v.StartIndex = u.StartIndex
	if x, err := encoding.BytesFromJSON(u.End); err != nil {
		return fmt.Errorf("error decoding End: %w", err)
	} else {
		v.End = x
	}
	v.EndIndex = u.EndIndex
	if x, err := encoding.BytesFromJSON(u.Anchor); err != nil {
		return fmt.Errorf("error decoding Anchor: %w", err)
	} else {
		v.Anchor = x
	}
	v.Entries = u.Entries
	return nil
}

func (v *ReceiptEntry) UnmarshalJSON(data []byte) error {
	u := struct {
		Right bool    `json:"right,omitempty"`
		Hash  *string `json:"hash,omitempty"`
	}{}
	u.Right = v.Right
	u.Hash = encoding.BytesToJSON(v.Hash)
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Right = u.Right
	if x, err := encoding.BytesFromJSON(u.Hash); err != nil {
		return fmt.Errorf("error decoding Hash: %w", err)
	} else {
		v.Hash = x
	}
	return nil
}

func (v *State) UnmarshalJSON(data []byte) error {
	u := struct {
		Count    int64                      `json:"count,omitempty"`
		Pending  encoding.JsonList[*string] `json:"pending,omitempty"`
		HashList encoding.JsonList[*string] `json:"hashList,omitempty"`
	}{}
	u.Count = v.Count
	u.Pending = make(encoding.JsonList[*string], len(v.Pending))
	for i, x := range v.Pending {
		u.Pending[i] = encoding.BytesToJSON(x)
	}
	u.HashList = make(encoding.JsonList[*string], len(v.HashList))
	for i, x := range v.HashList {
		u.HashList[i] = encoding.BytesToJSON(x)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Count = u.Count
	v.Pending = make([][]byte, len(u.Pending))
	for i, x := range u.Pending {
		if x, err := encoding.BytesFromJSON(x); err != nil {
			return fmt.Errorf("error decoding Pending: %w", err)
		} else {
			v.Pending[i] = x
		}
	}
	v.HashList = make([][]byte, len(u.HashList))
	for i, x := range u.HashList {
		if x, err := encoding.BytesFromJSON(x); err != nil {
			return fmt.Errorf("error decoding HashList: %w", err)
		} else {
			v.HashList[i] = x
		}
	}
	return nil
}
