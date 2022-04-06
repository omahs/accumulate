package database

// GENERATED BY go run ./tools/cmd/gen-types. DO NOT EDIT.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"gitlab.com/accumulatenetwork/accumulate/internal/encoding"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

type SigOrTxn struct {
	fieldsSet   []bool
	Transaction *protocol.Transaction `json:"transaction,omitempty" form:"transaction" query:"transaction" validate:"required"`
	Signature   protocol.Signature    `json:"signature,omitempty" form:"signature" query:"signature" validate:"required"`
	Hash        [32]byte              `json:"hash,omitempty" form:"hash" query:"hash" validate:"required"`
}

type exampleFullAccountState struct {
	fieldsSet []bool
	State     protocol.Account `json:"state,omitempty" form:"state" query:"state" validate:"required"`
	Chains    []*merkleState   `json:"chains,omitempty" form:"chains" query:"chains" validate:"required"`
}

type merkleState struct {
	fieldsSet []bool
	Count     uint64     `json:"count,omitempty" form:"count" query:"count" validate:"required"`
	Pending   [][32]byte `json:"pending,omitempty" form:"pending" query:"pending" validate:"required"`
}

type sigSetData struct {
	fieldsSet []bool
	Version   uint64          `json:"version,omitempty" form:"version" query:"version" validate:"required"`
	Entries   []sigSetKeyData `json:"entries,omitempty" form:"entries" query:"entries" validate:"required"`
}

type sigSetKeyData struct {
	fieldsSet []bool
	System    bool     `json:"system,omitempty" form:"system" query:"system" validate:"required"`
	KeyHash   [32]byte `json:"keyHash,omitempty" form:"keyHash" query:"keyHash" validate:"required"`
	EntryHash [32]byte `json:"entryHash,omitempty" form:"entryHash" query:"entryHash" validate:"required"`
}

type txSyntheticTxns struct {
	fieldsSet []bool
	Txids     [][32]byte `json:"txids,omitempty" form:"txids" query:"txids" validate:"required"`
}

func (v *SigOrTxn) Copy() *SigOrTxn {
	u := new(SigOrTxn)

	if v.Transaction != nil {
		u.Transaction = (v.Transaction).Copy()
	}
	u.Signature = v.Signature
	u.Hash = v.Hash

	return u
}

func (v *SigOrTxn) CopyAsInterface() interface{} { return v.Copy() }

func (v *exampleFullAccountState) Copy() *exampleFullAccountState {
	u := new(exampleFullAccountState)

	u.State = v.State
	u.Chains = make([]*merkleState, len(v.Chains))
	for i, v := range v.Chains {
		if v != nil {
			u.Chains[i] = (v).Copy()
		}
	}

	return u
}

func (v *exampleFullAccountState) CopyAsInterface() interface{} { return v.Copy() }

func (v *merkleState) Copy() *merkleState {
	u := new(merkleState)

	u.Count = v.Count
	u.Pending = make([][32]byte, len(v.Pending))
	for i, v := range v.Pending {
		u.Pending[i] = v
	}

	return u
}

func (v *merkleState) CopyAsInterface() interface{} { return v.Copy() }

func (v *sigSetData) Copy() *sigSetData {
	u := new(sigSetData)

	u.Version = v.Version
	u.Entries = make([]sigSetKeyData, len(v.Entries))
	for i, v := range v.Entries {
		u.Entries[i] = *(&v).Copy()
	}

	return u
}

func (v *sigSetData) CopyAsInterface() interface{} { return v.Copy() }

func (v *sigSetKeyData) Copy() *sigSetKeyData {
	u := new(sigSetKeyData)

	u.System = v.System
	u.KeyHash = v.KeyHash
	u.EntryHash = v.EntryHash

	return u
}

func (v *sigSetKeyData) CopyAsInterface() interface{} { return v.Copy() }

func (v *txSyntheticTxns) Copy() *txSyntheticTxns {
	u := new(txSyntheticTxns)

	u.Txids = make([][32]byte, len(v.Txids))
	for i, v := range v.Txids {
		u.Txids[i] = v
	}

	return u
}

func (v *txSyntheticTxns) CopyAsInterface() interface{} { return v.Copy() }

func (v *SigOrTxn) Equal(u *SigOrTxn) bool {
	switch {
	case v.Transaction == u.Transaction:
		// equal
	case v.Transaction == nil || u.Transaction == nil:
		return false
	case !((v.Transaction).Equal(u.Transaction)):
		return false
	}
	if !(v.Signature == u.Signature) {
		return false
	}
	if !(v.Hash == u.Hash) {
		return false
	}

	return true
}

func (v *exampleFullAccountState) Equal(u *exampleFullAccountState) bool {
	if !(v.State == u.State) {
		return false
	}
	if len(v.Chains) != len(u.Chains) {
		return false
	}
	for i := range v.Chains {
		if !((v.Chains[i]).Equal(u.Chains[i])) {
			return false
		}
	}

	return true
}

func (v *merkleState) Equal(u *merkleState) bool {
	if !(v.Count == u.Count) {
		return false
	}
	if len(v.Pending) != len(u.Pending) {
		return false
	}
	for i := range v.Pending {
		if !(v.Pending[i] == u.Pending[i]) {
			return false
		}
	}

	return true
}

func (v *sigSetData) Equal(u *sigSetData) bool {
	if !(v.Version == u.Version) {
		return false
	}
	if len(v.Entries) != len(u.Entries) {
		return false
	}
	for i := range v.Entries {
		if !((&v.Entries[i]).Equal(&u.Entries[i])) {
			return false
		}
	}

	return true
}

func (v *sigSetKeyData) Equal(u *sigSetKeyData) bool {
	if !(v.System == u.System) {
		return false
	}
	if !(v.KeyHash == u.KeyHash) {
		return false
	}
	if !(v.EntryHash == u.EntryHash) {
		return false
	}

	return true
}

func (v *txSyntheticTxns) Equal(u *txSyntheticTxns) bool {
	if len(v.Txids) != len(u.Txids) {
		return false
	}
	for i := range v.Txids {
		if !(v.Txids[i] == u.Txids[i]) {
			return false
		}
	}

	return true
}

var fieldNames_SigOrTxn = []string{
	1: "Transaction",
	2: "Signature",
	3: "Hash",
}

func (v *SigOrTxn) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(v.Transaction == nil) {
		writer.WriteValue(1, v.Transaction)
	}
	if !(v.Signature == (nil)) {
		writer.WriteValue(2, v.Signature)
	}
	if !(v.Hash == ([32]byte{})) {
		writer.WriteHash(3, &v.Hash)
	}

	_, _, err := writer.Reset(fieldNames_SigOrTxn)
	return buffer.Bytes(), err
}

func (v *SigOrTxn) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Transaction is missing")
	} else if v.Transaction == nil {
		errs = append(errs, "field Transaction is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field Signature is missing")
	} else if v.Signature == (nil) {
		errs = append(errs, "field Signature is not set")
	}
	if len(v.fieldsSet) > 3 && !v.fieldsSet[3] {
		errs = append(errs, "field Hash is missing")
	} else if v.Hash == ([32]byte{}) {
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

var fieldNames_exampleFullAccountState = []string{
	1: "State",
	2: "Chains",
}

func (v *exampleFullAccountState) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(v.State == (nil)) {
		writer.WriteValue(1, v.State)
	}
	if !(len(v.Chains) == 0) {
		for _, v := range v.Chains {
			writer.WriteValue(2, v)
		}
	}

	_, _, err := writer.Reset(fieldNames_exampleFullAccountState)
	return buffer.Bytes(), err
}

func (v *exampleFullAccountState) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field State is missing")
	} else if v.State == (nil) {
		errs = append(errs, "field State is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field Chains is missing")
	} else if len(v.Chains) == 0 {
		errs = append(errs, "field Chains is not set")
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

var fieldNames_merkleState = []string{
	1: "Count",
	2: "Pending",
}

func (v *merkleState) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(v.Count == 0) {
		writer.WriteUint(1, v.Count)
	}
	if !(len(v.Pending) == 0) {
		for _, v := range v.Pending {
			writer.WriteHash(2, &v)
		}
	}

	_, _, err := writer.Reset(fieldNames_merkleState)
	return buffer.Bytes(), err
}

func (v *merkleState) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Count is missing")
	} else if v.Count == 0 {
		errs = append(errs, "field Count is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field Pending is missing")
	} else if len(v.Pending) == 0 {
		errs = append(errs, "field Pending is not set")
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

var fieldNames_sigSetData = []string{
	1: "Version",
	2: "Entries",
}

func (v *sigSetData) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(v.Version == 0) {
		writer.WriteUint(1, v.Version)
	}
	if !(len(v.Entries) == 0) {
		for _, v := range v.Entries {
			writer.WriteValue(2, &v)
		}
	}

	_, _, err := writer.Reset(fieldNames_sigSetData)
	return buffer.Bytes(), err
}

func (v *sigSetData) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Version is missing")
	} else if v.Version == 0 {
		errs = append(errs, "field Version is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
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

var fieldNames_sigSetKeyData = []string{
	1: "System",
	2: "KeyHash",
	3: "EntryHash",
}

func (v *sigSetKeyData) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(!v.System) {
		writer.WriteBool(1, v.System)
	}
	if !(v.KeyHash == ([32]byte{})) {
		writer.WriteHash(2, &v.KeyHash)
	}
	if !(v.EntryHash == ([32]byte{})) {
		writer.WriteHash(3, &v.EntryHash)
	}

	_, _, err := writer.Reset(fieldNames_sigSetKeyData)
	return buffer.Bytes(), err
}

func (v *sigSetKeyData) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field System is missing")
	} else if !v.System {
		errs = append(errs, "field System is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field KeyHash is missing")
	} else if v.KeyHash == ([32]byte{}) {
		errs = append(errs, "field KeyHash is not set")
	}
	if len(v.fieldsSet) > 3 && !v.fieldsSet[3] {
		errs = append(errs, "field EntryHash is missing")
	} else if v.EntryHash == ([32]byte{}) {
		errs = append(errs, "field EntryHash is not set")
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

var fieldNames_txSyntheticTxns = []string{
	1: "Txids",
}

func (v *txSyntheticTxns) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(len(v.Txids) == 0) {
		for _, v := range v.Txids {
			writer.WriteHash(1, &v)
		}
	}

	_, _, err := writer.Reset(fieldNames_txSyntheticTxns)
	return buffer.Bytes(), err
}

func (v *txSyntheticTxns) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Txids is missing")
	} else if len(v.Txids) == 0 {
		errs = append(errs, "field Txids is not set")
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

func (v *SigOrTxn) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *SigOrTxn) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x := new(protocol.Transaction); reader.ReadValue(1, x.UnmarshalBinary) {
		v.Transaction = x
	}
	reader.ReadValue(2, func(b []byte) error {
		x, err := protocol.UnmarshalSignature(b)
		if err == nil {
			v.Signature = x
		}
		return err
	})
	if x, ok := reader.ReadHash(3); ok {
		v.Hash = *x
	}

	seen, err := reader.Reset(fieldNames_SigOrTxn)
	v.fieldsSet = seen
	return err
}

func (v *exampleFullAccountState) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *exampleFullAccountState) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	reader.ReadValue(1, func(b []byte) error {
		x, err := protocol.UnmarshalAccount(b)
		if err == nil {
			v.State = x
		}
		return err
	})
	for {
		if x := new(merkleState); reader.ReadValue(2, x.UnmarshalBinary) {
			v.Chains = append(v.Chains, x)
		} else {
			break
		}
	}

	seen, err := reader.Reset(fieldNames_exampleFullAccountState)
	v.fieldsSet = seen
	return err
}

func (v *merkleState) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *merkleState) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x, ok := reader.ReadUint(1); ok {
		v.Count = x
	}
	for {
		if x, ok := reader.ReadHash(2); ok {
			v.Pending = append(v.Pending, *x)
		} else {
			break
		}
	}

	seen, err := reader.Reset(fieldNames_merkleState)
	v.fieldsSet = seen
	return err
}

func (v *sigSetData) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *sigSetData) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x, ok := reader.ReadUint(1); ok {
		v.Version = x
	}
	for {
		if x := new(sigSetKeyData); reader.ReadValue(2, x.UnmarshalBinary) {
			v.Entries = append(v.Entries, *x)
		} else {
			break
		}
	}

	seen, err := reader.Reset(fieldNames_sigSetData)
	v.fieldsSet = seen
	return err
}

func (v *sigSetKeyData) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *sigSetKeyData) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x, ok := reader.ReadBool(1); ok {
		v.System = x
	}
	if x, ok := reader.ReadHash(2); ok {
		v.KeyHash = *x
	}
	if x, ok := reader.ReadHash(3); ok {
		v.EntryHash = *x
	}

	seen, err := reader.Reset(fieldNames_sigSetKeyData)
	v.fieldsSet = seen
	return err
}

func (v *txSyntheticTxns) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *txSyntheticTxns) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	for {
		if x, ok := reader.ReadHash(1); ok {
			v.Txids = append(v.Txids, *x)
		} else {
			break
		}
	}

	seen, err := reader.Reset(fieldNames_txSyntheticTxns)
	v.fieldsSet = seen
	return err
}

func (v *SigOrTxn) MarshalJSON() ([]byte, error) {
	u := struct {
		Transaction *protocol.Transaction                          `json:"transaction,omitempty"`
		Signature   encoding.JsonUnmarshalWith[protocol.Signature] `json:"signature,omitempty"`
		Hash        string                                         `json:"hash,omitempty"`
	}{}
	u.Transaction = v.Transaction
	u.Signature = encoding.JsonUnmarshalWith[protocol.Signature]{Value: v.Signature, Func: protocol.UnmarshalSignatureJSON}
	u.Hash = encoding.ChainToJSON(v.Hash)
	return json.Marshal(&u)
}

func (v *exampleFullAccountState) MarshalJSON() ([]byte, error) {
	u := struct {
		State  encoding.JsonUnmarshalWith[protocol.Account] `json:"state,omitempty"`
		Chains encoding.JsonList[*merkleState]              `json:"chains,omitempty"`
	}{}
	u.State = encoding.JsonUnmarshalWith[protocol.Account]{Value: v.State, Func: protocol.UnmarshalAccountJSON}
	u.Chains = v.Chains
	return json.Marshal(&u)
}

func (v *merkleState) MarshalJSON() ([]byte, error) {
	u := struct {
		Count   uint64                    `json:"count,omitempty"`
		Pending encoding.JsonList[string] `json:"pending,omitempty"`
	}{}
	u.Count = v.Count
	u.Pending = make(encoding.JsonList[string], len(v.Pending))
	for i, x := range v.Pending {
		u.Pending[i] = encoding.ChainToJSON(x)
	}
	return json.Marshal(&u)
}

func (v *sigSetData) MarshalJSON() ([]byte, error) {
	u := struct {
		Version uint64                           `json:"version,omitempty"`
		Entries encoding.JsonList[sigSetKeyData] `json:"entries,omitempty"`
	}{}
	u.Version = v.Version
	u.Entries = v.Entries
	return json.Marshal(&u)
}

func (v *sigSetKeyData) MarshalJSON() ([]byte, error) {
	u := struct {
		System    bool   `json:"system,omitempty"`
		KeyHash   string `json:"keyHash,omitempty"`
		EntryHash string `json:"entryHash,omitempty"`
	}{}
	u.System = v.System
	u.KeyHash = encoding.ChainToJSON(v.KeyHash)
	u.EntryHash = encoding.ChainToJSON(v.EntryHash)
	return json.Marshal(&u)
}

func (v *txSyntheticTxns) MarshalJSON() ([]byte, error) {
	u := struct {
		Txids encoding.JsonList[string] `json:"txids,omitempty"`
	}{}
	u.Txids = make(encoding.JsonList[string], len(v.Txids))
	for i, x := range v.Txids {
		u.Txids[i] = encoding.ChainToJSON(x)
	}
	return json.Marshal(&u)
}

func (v *SigOrTxn) UnmarshalJSON(data []byte) error {
	u := struct {
		Transaction *protocol.Transaction                          `json:"transaction,omitempty"`
		Signature   encoding.JsonUnmarshalWith[protocol.Signature] `json:"signature,omitempty"`
		Hash        string                                         `json:"hash,omitempty"`
	}{}
	u.Transaction = v.Transaction
	u.Signature = encoding.JsonUnmarshalWith[protocol.Signature]{Value: v.Signature, Func: protocol.UnmarshalSignatureJSON}
	u.Hash = encoding.ChainToJSON(v.Hash)
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Transaction = u.Transaction
	v.Signature = u.Signature.Value

	if x, err := encoding.ChainFromJSON(u.Hash); err != nil {
		return fmt.Errorf("error decoding Hash: %w", err)
	} else {
		v.Hash = x
	}
	return nil
}

func (v *exampleFullAccountState) UnmarshalJSON(data []byte) error {
	u := struct {
		State  encoding.JsonUnmarshalWith[protocol.Account] `json:"state,omitempty"`
		Chains encoding.JsonList[*merkleState]              `json:"chains,omitempty"`
	}{}
	u.State = encoding.JsonUnmarshalWith[protocol.Account]{Value: v.State, Func: protocol.UnmarshalAccountJSON}
	u.Chains = v.Chains
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.State = u.State.Value

	v.Chains = u.Chains
	return nil
}

func (v *merkleState) UnmarshalJSON(data []byte) error {
	u := struct {
		Count   uint64                    `json:"count,omitempty"`
		Pending encoding.JsonList[string] `json:"pending,omitempty"`
	}{}
	u.Count = v.Count
	u.Pending = make(encoding.JsonList[string], len(v.Pending))
	for i, x := range v.Pending {
		u.Pending[i] = encoding.ChainToJSON(x)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Count = u.Count
	v.Pending = make([][32]byte, len(u.Pending))
	for i, x := range u.Pending {
		if x, err := encoding.ChainFromJSON(x); err != nil {
			return fmt.Errorf("error decoding Pending: %w", err)
		} else {
			v.Pending[i] = x
		}
	}
	return nil
}

func (v *sigSetData) UnmarshalJSON(data []byte) error {
	u := struct {
		Version uint64                           `json:"version,omitempty"`
		Entries encoding.JsonList[sigSetKeyData] `json:"entries,omitempty"`
	}{}
	u.Version = v.Version
	u.Entries = v.Entries
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Version = u.Version
	v.Entries = u.Entries
	return nil
}

func (v *sigSetKeyData) UnmarshalJSON(data []byte) error {
	u := struct {
		System    bool   `json:"system,omitempty"`
		KeyHash   string `json:"keyHash,omitempty"`
		EntryHash string `json:"entryHash,omitempty"`
	}{}
	u.System = v.System
	u.KeyHash = encoding.ChainToJSON(v.KeyHash)
	u.EntryHash = encoding.ChainToJSON(v.EntryHash)
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.System = u.System
	if x, err := encoding.ChainFromJSON(u.KeyHash); err != nil {
		return fmt.Errorf("error decoding KeyHash: %w", err)
	} else {
		v.KeyHash = x
	}
	if x, err := encoding.ChainFromJSON(u.EntryHash); err != nil {
		return fmt.Errorf("error decoding EntryHash: %w", err)
	} else {
		v.EntryHash = x
	}
	return nil
}

func (v *txSyntheticTxns) UnmarshalJSON(data []byte) error {
	u := struct {
		Txids encoding.JsonList[string] `json:"txids,omitempty"`
	}{}
	u.Txids = make(encoding.JsonList[string], len(v.Txids))
	for i, x := range v.Txids {
		u.Txids[i] = encoding.ChainToJSON(x)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Txids = make([][32]byte, len(u.Txids))
	for i, x := range u.Txids {
		if x, err := encoding.ChainFromJSON(x); err != nil {
			return fmt.Errorf("error decoding Txids: %w", err)
		} else {
			v.Txids[i] = x
		}
	}
	return nil
}
