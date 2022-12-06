// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package database

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
	"gitlab.com/accumulatenetwork/accumulate/pkg/types/merkle"
	"gitlab.com/accumulatenetwork/accumulate/pkg/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

type BlockStateSynthTxnEntry struct {
	fieldsSet   []bool
	Account     *url.URL `json:"account,omitempty" form:"account" query:"account" validate:"required"`
	Transaction []byte   `json:"transaction,omitempty" form:"transaction" query:"transaction" validate:"required"`
	ChainEntry  uint64   `json:"chainEntry,omitempty" form:"chainEntry" query:"chainEntry" validate:"required"`
	extraData   []byte
}

type ReceiptList struct {
	fieldsSet []bool
	// MerkleState MerkleState at the beginning of the list.
	MerkleState      *MerkleState    `json:"merkleState,omitempty" form:"merkleState" query:"merkleState" validate:"required"`
	Elements         [][]byte        `json:"elements,omitempty" form:"elements" query:"elements" validate:"required"`
	Receipt          *merkle.Receipt `json:"receipt,omitempty" form:"receipt" query:"receipt" validate:"required"`
	ContinuedReceipt *merkle.Receipt `json:"continuedReceipt,omitempty" form:"continuedReceipt" query:"continuedReceipt" validate:"required"`
	extraData        []byte
}

type SigOrTxn struct {
	fieldsSet   []bool
	Transaction *protocol.Transaction `json:"transaction,omitempty" form:"transaction" query:"transaction" validate:"required"`
	Signature   protocol.Signature    `json:"signature,omitempty" form:"signature" query:"signature" validate:"required"`
	Txid        *url.TxID             `json:"txid,omitempty" form:"txid" query:"txid" validate:"required"`
	extraData   []byte
}

type SigSetEntry struct {
	fieldsSet        []bool
	Type             protocol.SignatureType `json:"type,omitempty" form:"type" query:"type" validate:"required"`
	KeyEntryIndex    uint64                 `json:"keyEntryIndex,omitempty" form:"keyEntryIndex" query:"keyEntryIndex" validate:"required"`
	SignatureHash    [32]byte               `json:"signatureHash,omitempty" form:"signatureHash" query:"signatureHash" validate:"required"`
	ValidatorKeyHash *[32]byte              `json:"validatorKeyHash,omitempty" form:"validatorKeyHash" query:"validatorKeyHash" validate:"required"`
	extraData        []byte
}

type TransactionChainEntry struct {
	fieldsSet []bool
	Account   *url.URL `json:"account,omitempty" form:"account" query:"account" validate:"required"`
	// Chain is the name of the chain.
	Chain string `json:"chain,omitempty" form:"chain" query:"chain" validate:"required"`
	// ChainIndex is the index of the entry in the chain's index chain.
	ChainIndex uint64 `json:"chainIndex,omitempty" form:"chainIndex" query:"chainIndex" validate:"required"`
	// AnchorIndex is the index of the entry in the anchor chain's index chain.
	AnchorIndex uint64 `json:"anchorIndex,omitempty" form:"anchorIndex" query:"anchorIndex" validate:"required"`
	extraData   []byte
}

type sigSetData struct {
	fieldsSet []bool
	Version   uint64        `json:"version,omitempty" form:"version" query:"version" validate:"required"`
	Entries   []SigSetEntry `json:"entries,omitempty" form:"entries" query:"entries" validate:"required"`
	extraData []byte
}

func (v *BlockStateSynthTxnEntry) Copy() *BlockStateSynthTxnEntry {
	u := new(BlockStateSynthTxnEntry)

	if v.Account != nil {
		u.Account = v.Account
	}
	u.Transaction = encoding.BytesCopy(v.Transaction)
	u.ChainEntry = v.ChainEntry

	return u
}

func (v *BlockStateSynthTxnEntry) CopyAsInterface() interface{} { return v.Copy() }

func (v *ReceiptList) Copy() *ReceiptList {
	u := new(ReceiptList)

	if v.MerkleState != nil {
		u.MerkleState = (v.MerkleState).Copy()
	}
	u.Elements = make([][]byte, len(v.Elements))
	for i, v := range v.Elements {
		u.Elements[i] = encoding.BytesCopy(v)
	}
	if v.Receipt != nil {
		u.Receipt = (v.Receipt).Copy()
	}
	if v.ContinuedReceipt != nil {
		u.ContinuedReceipt = (v.ContinuedReceipt).Copy()
	}

	return u
}

func (v *ReceiptList) CopyAsInterface() interface{} { return v.Copy() }

func (v *SigOrTxn) Copy() *SigOrTxn {
	u := new(SigOrTxn)

	if v.Transaction != nil {
		u.Transaction = (v.Transaction).Copy()
	}
	if v.Signature != nil {
		u.Signature = protocol.CopySignature(v.Signature)
	}
	if v.Txid != nil {
		u.Txid = v.Txid
	}

	return u
}

func (v *SigOrTxn) CopyAsInterface() interface{} { return v.Copy() }

func (v *SigSetEntry) Copy() *SigSetEntry {
	u := new(SigSetEntry)

	u.Type = v.Type
	u.KeyEntryIndex = v.KeyEntryIndex
	u.SignatureHash = v.SignatureHash
	if v.ValidatorKeyHash != nil {
		u.ValidatorKeyHash = new([32]byte)
		*u.ValidatorKeyHash = *v.ValidatorKeyHash
	}

	return u
}

func (v *SigSetEntry) CopyAsInterface() interface{} { return v.Copy() }

func (v *TransactionChainEntry) Copy() *TransactionChainEntry {
	u := new(TransactionChainEntry)

	if v.Account != nil {
		u.Account = v.Account
	}
	u.Chain = v.Chain
	u.ChainIndex = v.ChainIndex
	u.AnchorIndex = v.AnchorIndex

	return u
}

func (v *TransactionChainEntry) CopyAsInterface() interface{} { return v.Copy() }

func (v *sigSetData) Copy() *sigSetData {
	u := new(sigSetData)

	u.Version = v.Version
	u.Entries = make([]SigSetEntry, len(v.Entries))
	for i, v := range v.Entries {
		u.Entries[i] = *(&v).Copy()
	}

	return u
}

func (v *sigSetData) CopyAsInterface() interface{} { return v.Copy() }

func (v *BlockStateSynthTxnEntry) Equal(u *BlockStateSynthTxnEntry) bool {
	switch {
	case v.Account == u.Account:
		// equal
	case v.Account == nil || u.Account == nil:
		return false
	case !((v.Account).Equal(u.Account)):
		return false
	}
	if !(bytes.Equal(v.Transaction, u.Transaction)) {
		return false
	}
	if !(v.ChainEntry == u.ChainEntry) {
		return false
	}

	return true
}

func (v *ReceiptList) Equal(u *ReceiptList) bool {
	switch {
	case v.MerkleState == u.MerkleState:
		// equal
	case v.MerkleState == nil || u.MerkleState == nil:
		return false
	case !((v.MerkleState).Equal(u.MerkleState)):
		return false
	}
	if len(v.Elements) != len(u.Elements) {
		return false
	}
	for i := range v.Elements {
		if !(bytes.Equal(v.Elements[i], u.Elements[i])) {
			return false
		}
	}
	switch {
	case v.Receipt == u.Receipt:
		// equal
	case v.Receipt == nil || u.Receipt == nil:
		return false
	case !((v.Receipt).Equal(u.Receipt)):
		return false
	}
	switch {
	case v.ContinuedReceipt == u.ContinuedReceipt:
		// equal
	case v.ContinuedReceipt == nil || u.ContinuedReceipt == nil:
		return false
	case !((v.ContinuedReceipt).Equal(u.ContinuedReceipt)):
		return false
	}

	return true
}

func (v *SigOrTxn) Equal(u *SigOrTxn) bool {
	switch {
	case v.Transaction == u.Transaction:
		// equal
	case v.Transaction == nil || u.Transaction == nil:
		return false
	case !((v.Transaction).Equal(u.Transaction)):
		return false
	}
	if !(protocol.EqualSignature(v.Signature, u.Signature)) {
		return false
	}
	switch {
	case v.Txid == u.Txid:
		// equal
	case v.Txid == nil || u.Txid == nil:
		return false
	case !((v.Txid).Equal(u.Txid)):
		return false
	}

	return true
}

func (v *SigSetEntry) Equal(u *SigSetEntry) bool {
	if !(v.Type == u.Type) {
		return false
	}
	if !(v.KeyEntryIndex == u.KeyEntryIndex) {
		return false
	}
	if !(v.SignatureHash == u.SignatureHash) {
		return false
	}
	switch {
	case v.ValidatorKeyHash == u.ValidatorKeyHash:
		// equal
	case v.ValidatorKeyHash == nil || u.ValidatorKeyHash == nil:
		return false
	case !(*v.ValidatorKeyHash == *u.ValidatorKeyHash):
		return false
	}

	return true
}

func (v *TransactionChainEntry) Equal(u *TransactionChainEntry) bool {
	switch {
	case v.Account == u.Account:
		// equal
	case v.Account == nil || u.Account == nil:
		return false
	case !((v.Account).Equal(u.Account)):
		return false
	}
	if !(v.Chain == u.Chain) {
		return false
	}
	if !(v.ChainIndex == u.ChainIndex) {
		return false
	}
	if !(v.AnchorIndex == u.AnchorIndex) {
		return false
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

var fieldNames_BlockStateSynthTxnEntry = []string{
	1: "Account",
	2: "Transaction",
	3: "ChainEntry",
}

func (v *BlockStateSynthTxnEntry) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(v.Account == nil) {
		writer.WriteUrl(1, v.Account)
	}
	if !(len(v.Transaction) == 0) {
		writer.WriteBytes(2, v.Transaction)
	}
	if !(v.ChainEntry == 0) {
		writer.WriteUint(3, v.ChainEntry)
	}

	_, _, err := writer.Reset(fieldNames_BlockStateSynthTxnEntry)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *BlockStateSynthTxnEntry) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 0 && !v.fieldsSet[0] {
		errs = append(errs, "field Account is missing")
	} else if v.Account == nil {
		errs = append(errs, "field Account is not set")
	}
	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Transaction is missing")
	} else if len(v.Transaction) == 0 {
		errs = append(errs, "field Transaction is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field ChainEntry is missing")
	} else if v.ChainEntry == 0 {
		errs = append(errs, "field ChainEntry is not set")
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

var fieldNames_ReceiptList = []string{
	1: "MerkleState",
	2: "Elements",
	3: "Receipt",
	4: "ContinuedReceipt",
}

func (v *ReceiptList) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(v.MerkleState == nil) {
		writer.WriteValue(1, v.MerkleState.MarshalBinary)
	}
	if !(len(v.Elements) == 0) {
		for _, v := range v.Elements {
			writer.WriteBytes(2, v)
		}
	}
	if !(v.Receipt == nil) {
		writer.WriteValue(3, v.Receipt.MarshalBinary)
	}
	if !(v.ContinuedReceipt == nil) {
		writer.WriteValue(4, v.ContinuedReceipt.MarshalBinary)
	}

	_, _, err := writer.Reset(fieldNames_ReceiptList)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *ReceiptList) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 0 && !v.fieldsSet[0] {
		errs = append(errs, "field MerkleState is missing")
	} else if v.MerkleState == nil {
		errs = append(errs, "field MerkleState is not set")
	}
	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Elements is missing")
	} else if len(v.Elements) == 0 {
		errs = append(errs, "field Elements is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field Receipt is missing")
	} else if v.Receipt == nil {
		errs = append(errs, "field Receipt is not set")
	}
	if len(v.fieldsSet) > 3 && !v.fieldsSet[3] {
		errs = append(errs, "field ContinuedReceipt is missing")
	} else if v.ContinuedReceipt == nil {
		errs = append(errs, "field ContinuedReceipt is not set")
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

var fieldNames_SigOrTxn = []string{
	1: "Transaction",
	2: "Signature",
	3: "Txid",
}

func (v *SigOrTxn) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(v.Transaction == nil) {
		writer.WriteValue(1, v.Transaction.MarshalBinary)
	}
	if !(protocol.EqualSignature(v.Signature, nil)) {
		writer.WriteValue(2, v.Signature.MarshalBinary)
	}
	if !(v.Txid == nil) {
		writer.WriteTxid(3, v.Txid)
	}

	_, _, err := writer.Reset(fieldNames_SigOrTxn)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *SigOrTxn) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 0 && !v.fieldsSet[0] {
		errs = append(errs, "field Transaction is missing")
	} else if v.Transaction == nil {
		errs = append(errs, "field Transaction is not set")
	}
	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Signature is missing")
	} else if protocol.EqualSignature(v.Signature, nil) {
		errs = append(errs, "field Signature is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field Txid is missing")
	} else if v.Txid == nil {
		errs = append(errs, "field Txid is not set")
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

var fieldNames_SigSetEntry = []string{
	1: "Type",
	2: "KeyEntryIndex",
	3: "SignatureHash",
	4: "ValidatorKeyHash",
}

func (v *SigSetEntry) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(v.Type == 0) {
		writer.WriteEnum(1, v.Type)
	}
	if !(v.KeyEntryIndex == 0) {
		writer.WriteUint(2, v.KeyEntryIndex)
	}
	if !(v.SignatureHash == ([32]byte{})) {
		writer.WriteHash(3, &v.SignatureHash)
	}
	if !(v.ValidatorKeyHash == nil) {
		writer.WriteHash(4, v.ValidatorKeyHash)
	}

	_, _, err := writer.Reset(fieldNames_SigSetEntry)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *SigSetEntry) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 0 && !v.fieldsSet[0] {
		errs = append(errs, "field Type is missing")
	} else if v.Type == 0 {
		errs = append(errs, "field Type is not set")
	}
	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field KeyEntryIndex is missing")
	} else if v.KeyEntryIndex == 0 {
		errs = append(errs, "field KeyEntryIndex is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field SignatureHash is missing")
	} else if v.SignatureHash == ([32]byte{}) {
		errs = append(errs, "field SignatureHash is not set")
	}
	if len(v.fieldsSet) > 3 && !v.fieldsSet[3] {
		errs = append(errs, "field ValidatorKeyHash is missing")
	} else if v.ValidatorKeyHash == nil {
		errs = append(errs, "field ValidatorKeyHash is not set")
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

var fieldNames_TransactionChainEntry = []string{
	1: "Account",
	2: "Chain",
	3: "ChainIndex",
	4: "AnchorIndex",
}

func (v *TransactionChainEntry) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(v.Account == nil) {
		writer.WriteUrl(1, v.Account)
	}
	if !(len(v.Chain) == 0) {
		writer.WriteString(2, v.Chain)
	}
	if !(v.ChainIndex == 0) {
		writer.WriteUint(3, v.ChainIndex)
	}
	if !(v.AnchorIndex == 0) {
		writer.WriteUint(4, v.AnchorIndex)
	}

	_, _, err := writer.Reset(fieldNames_TransactionChainEntry)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *TransactionChainEntry) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 0 && !v.fieldsSet[0] {
		errs = append(errs, "field Account is missing")
	} else if v.Account == nil {
		errs = append(errs, "field Account is not set")
	}
	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Chain is missing")
	} else if len(v.Chain) == 0 {
		errs = append(errs, "field Chain is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field ChainIndex is missing")
	} else if v.ChainIndex == 0 {
		errs = append(errs, "field ChainIndex is not set")
	}
	if len(v.fieldsSet) > 3 && !v.fieldsSet[3] {
		errs = append(errs, "field AnchorIndex is missing")
	} else if v.AnchorIndex == 0 {
		errs = append(errs, "field AnchorIndex is not set")
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
			writer.WriteValue(2, v.MarshalBinary)
		}
	}

	_, _, err := writer.Reset(fieldNames_sigSetData)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *sigSetData) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 0 && !v.fieldsSet[0] {
		errs = append(errs, "field Version is missing")
	} else if v.Version == 0 {
		errs = append(errs, "field Version is not set")
	}
	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
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

func (v *BlockStateSynthTxnEntry) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *BlockStateSynthTxnEntry) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x, ok := reader.ReadUrl(1); ok {
		v.Account = x
	}
	if x, ok := reader.ReadBytes(2); ok {
		v.Transaction = x
	}
	if x, ok := reader.ReadUint(3); ok {
		v.ChainEntry = x
	}

	seen, err := reader.Reset(fieldNames_BlockStateSynthTxnEntry)
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

func (v *ReceiptList) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *ReceiptList) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x := new(MerkleState); reader.ReadValue(1, x.UnmarshalBinaryFrom) {
		v.MerkleState = x
	}
	for {
		if x, ok := reader.ReadBytes(2); ok {
			v.Elements = append(v.Elements, x)
		} else {
			break
		}
	}
	if x := new(merkle.Receipt); reader.ReadValue(3, x.UnmarshalBinaryFrom) {
		v.Receipt = x
	}
	if x := new(merkle.Receipt); reader.ReadValue(4, x.UnmarshalBinaryFrom) {
		v.ContinuedReceipt = x
	}

	seen, err := reader.Reset(fieldNames_ReceiptList)
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

func (v *SigOrTxn) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *SigOrTxn) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x := new(protocol.Transaction); reader.ReadValue(1, x.UnmarshalBinaryFrom) {
		v.Transaction = x
	}
	reader.ReadValue(2, func(r io.Reader) error {
		x, err := protocol.UnmarshalSignatureFrom(r)
		if err == nil {
			v.Signature = x
		}
		return err
	})
	if x, ok := reader.ReadTxid(3); ok {
		v.Txid = x
	}

	seen, err := reader.Reset(fieldNames_SigOrTxn)
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

func (v *SigSetEntry) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *SigSetEntry) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x := new(protocol.SignatureType); reader.ReadEnum(1, x) {
		v.Type = *x
	}
	if x, ok := reader.ReadUint(2); ok {
		v.KeyEntryIndex = x
	}
	if x, ok := reader.ReadHash(3); ok {
		v.SignatureHash = *x
	}
	if x, ok := reader.ReadHash(4); ok {
		v.ValidatorKeyHash = x
	}

	seen, err := reader.Reset(fieldNames_SigSetEntry)
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

func (v *TransactionChainEntry) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *TransactionChainEntry) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x, ok := reader.ReadUrl(1); ok {
		v.Account = x
	}
	if x, ok := reader.ReadString(2); ok {
		v.Chain = x
	}
	if x, ok := reader.ReadUint(3); ok {
		v.ChainIndex = x
	}
	if x, ok := reader.ReadUint(4); ok {
		v.AnchorIndex = x
	}

	seen, err := reader.Reset(fieldNames_TransactionChainEntry)
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

func (v *sigSetData) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *sigSetData) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x, ok := reader.ReadUint(1); ok {
		v.Version = x
	}
	for {
		if x := new(SigSetEntry); reader.ReadValue(2, x.UnmarshalBinaryFrom) {
			v.Entries = append(v.Entries, *x)
		} else {
			break
		}
	}

	seen, err := reader.Reset(fieldNames_sigSetData)
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

func (v *BlockStateSynthTxnEntry) MarshalJSON() ([]byte, error) {
	u := struct {
		Account     *url.URL `json:"account,omitempty"`
		Transaction *string  `json:"transaction,omitempty"`
		ChainEntry  uint64   `json:"chainEntry,omitempty"`
	}{}
	u.Account = v.Account
	u.Transaction = encoding.BytesToJSON(v.Transaction)
	u.ChainEntry = v.ChainEntry
	return json.Marshal(&u)
}

func (v *ReceiptList) MarshalJSON() ([]byte, error) {
	u := struct {
		MerkleState      *MerkleState               `json:"merkleState,omitempty"`
		Elements         encoding.JsonList[*string] `json:"elements,omitempty"`
		Receipt          *merkle.Receipt            `json:"receipt,omitempty"`
		ContinuedReceipt *merkle.Receipt            `json:"continuedReceipt,omitempty"`
	}{}
	u.MerkleState = v.MerkleState
	u.Elements = make(encoding.JsonList[*string], len(v.Elements))
	for i, x := range v.Elements {
		u.Elements[i] = encoding.BytesToJSON(x)
	}
	u.Receipt = v.Receipt
	u.ContinuedReceipt = v.ContinuedReceipt
	return json.Marshal(&u)
}

func (v *SigOrTxn) MarshalJSON() ([]byte, error) {
	u := struct {
		Transaction *protocol.Transaction                          `json:"transaction,omitempty"`
		Signature   encoding.JsonUnmarshalWith[protocol.Signature] `json:"signature,omitempty"`
		Txid        *url.TxID                                      `json:"txid,omitempty"`
	}{}
	u.Transaction = v.Transaction
	u.Signature = encoding.JsonUnmarshalWith[protocol.Signature]{Value: v.Signature, Func: protocol.UnmarshalSignatureJSON}
	u.Txid = v.Txid
	return json.Marshal(&u)
}

func (v *SigSetEntry) MarshalJSON() ([]byte, error) {
	u := struct {
		Type             protocol.SignatureType `json:"type,omitempty"`
		KeyEntryIndex    uint64                 `json:"keyEntryIndex,omitempty"`
		SignatureHash    string                 `json:"signatureHash,omitempty"`
		ValidatorKeyHash string                 `json:"validatorKeyHash,omitempty"`
	}{}
	u.Type = v.Type
	u.KeyEntryIndex = v.KeyEntryIndex
	u.SignatureHash = encoding.ChainToJSON(v.SignatureHash)
	if v.ValidatorKeyHash != nil {
		u.ValidatorKeyHash = encoding.ChainToJSON(*v.ValidatorKeyHash)
	}
	return json.Marshal(&u)
}

func (v *sigSetData) MarshalJSON() ([]byte, error) {
	u := struct {
		Version uint64                         `json:"version,omitempty"`
		Entries encoding.JsonList[SigSetEntry] `json:"entries,omitempty"`
	}{}
	u.Version = v.Version
	u.Entries = v.Entries
	return json.Marshal(&u)
}

func (v *BlockStateSynthTxnEntry) UnmarshalJSON(data []byte) error {
	u := struct {
		Account     *url.URL `json:"account,omitempty"`
		Transaction *string  `json:"transaction,omitempty"`
		ChainEntry  uint64   `json:"chainEntry,omitempty"`
	}{}
	u.Account = v.Account
	u.Transaction = encoding.BytesToJSON(v.Transaction)
	u.ChainEntry = v.ChainEntry
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Account = u.Account
	if x, err := encoding.BytesFromJSON(u.Transaction); err != nil {
		return fmt.Errorf("error decoding Transaction: %w", err)
	} else {
		v.Transaction = x
	}
	v.ChainEntry = u.ChainEntry
	return nil
}

func (v *ReceiptList) UnmarshalJSON(data []byte) error {
	u := struct {
		MerkleState      *MerkleState               `json:"merkleState,omitempty"`
		Elements         encoding.JsonList[*string] `json:"elements,omitempty"`
		Receipt          *merkle.Receipt            `json:"receipt,omitempty"`
		ContinuedReceipt *merkle.Receipt            `json:"continuedReceipt,omitempty"`
	}{}
	u.MerkleState = v.MerkleState
	u.Elements = make(encoding.JsonList[*string], len(v.Elements))
	for i, x := range v.Elements {
		u.Elements[i] = encoding.BytesToJSON(x)
	}
	u.Receipt = v.Receipt
	u.ContinuedReceipt = v.ContinuedReceipt
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.MerkleState = u.MerkleState
	v.Elements = make([][]byte, len(u.Elements))
	for i, x := range u.Elements {
		if x, err := encoding.BytesFromJSON(x); err != nil {
			return fmt.Errorf("error decoding Elements: %w", err)
		} else {
			v.Elements[i] = x
		}
	}
	v.Receipt = u.Receipt
	v.ContinuedReceipt = u.ContinuedReceipt
	return nil
}

func (v *SigOrTxn) UnmarshalJSON(data []byte) error {
	u := struct {
		Transaction *protocol.Transaction                          `json:"transaction,omitempty"`
		Signature   encoding.JsonUnmarshalWith[protocol.Signature] `json:"signature,omitempty"`
		Txid        *url.TxID                                      `json:"txid,omitempty"`
	}{}
	u.Transaction = v.Transaction
	u.Signature = encoding.JsonUnmarshalWith[protocol.Signature]{Value: v.Signature, Func: protocol.UnmarshalSignatureJSON}
	u.Txid = v.Txid
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Transaction = u.Transaction
	v.Signature = u.Signature.Value

	v.Txid = u.Txid
	return nil
}

func (v *SigSetEntry) UnmarshalJSON(data []byte) error {
	u := struct {
		Type             protocol.SignatureType `json:"type,omitempty"`
		KeyEntryIndex    uint64                 `json:"keyEntryIndex,omitempty"`
		SignatureHash    string                 `json:"signatureHash,omitempty"`
		ValidatorKeyHash string                 `json:"validatorKeyHash,omitempty"`
	}{}
	u.Type = v.Type
	u.KeyEntryIndex = v.KeyEntryIndex
	u.SignatureHash = encoding.ChainToJSON(v.SignatureHash)
	if v.ValidatorKeyHash != nil {
		u.ValidatorKeyHash = encoding.ChainToJSON(*v.ValidatorKeyHash)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Type = u.Type
	v.KeyEntryIndex = u.KeyEntryIndex
	if x, err := encoding.ChainFromJSON(u.SignatureHash); err != nil {
		return fmt.Errorf("error decoding SignatureHash: %w", err)
	} else {
		v.SignatureHash = x
	}
	if x, err := encoding.ChainFromJSON(u.ValidatorKeyHash); err != nil {
		return fmt.Errorf("error decoding ValidatorKeyHash: %w", err)
	} else {
		v.ValidatorKeyHash = &x
	}
	return nil
}

func (v *sigSetData) UnmarshalJSON(data []byte) error {
	u := struct {
		Version uint64                         `json:"version,omitempty"`
		Entries encoding.JsonList[SigSetEntry] `json:"entries,omitempty"`
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
