package indexing

// GENERATED BY go run ./tools/cmd/gen-types. DO NOT EDIT.

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"gitlab.com/accumulatenetwork/accumulate/internal/encoding"
	"gitlab.com/accumulatenetwork/accumulate/internal/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

type BlockChainUpdatesIndex struct {
	fieldsSet []bool
	Entries   []*ChainUpdate `json:"entries,omitempty" form:"entries" query:"entries" validate:"required"`
	extraData []byte
}

type BlockStateIndex struct {
	fieldsSet         []bool
	ProducedSynthTxns []*BlockStateSynthTxnEntry `json:"producedSynthTxns,omitempty" form:"producedSynthTxns" query:"producedSynthTxns" validate:"required"`
	extraData         []byte
}

type BlockStateSynthTxnEntry struct {
	fieldsSet   []bool
	Transaction []byte `json:"transaction,omitempty" form:"transaction" query:"transaction" validate:"required"`
	ChainEntry  uint64 `json:"chainEntry,omitempty" form:"chainEntry" query:"chainEntry" validate:"required"`
	extraData   []byte
}

type ChainUpdate struct {
	fieldsSet   []bool
	Account     *url.URL           `json:"account,omitempty" form:"account" query:"account" validate:"required"`
	Name        string             `json:"name,omitempty" form:"name" query:"name" validate:"required"`
	Type        protocol.ChainType `json:"type,omitempty" form:"type" query:"type" validate:"required"`
	Index       uint64             `json:"index,omitempty" form:"index" query:"index" validate:"required"`
	SourceIndex uint64             `json:"sourceIndex,omitempty" form:"sourceIndex" query:"sourceIndex" validate:"required"`
	SourceBlock uint64             `json:"sourceBlock,omitempty" form:"sourceBlock" query:"sourceBlock" validate:"required"`
	Entry       []byte             `json:"entry,omitempty" form:"entry" query:"entry" validate:"required"`
	extraData   []byte
}

type PendingTransactionsIndex struct {
	fieldsSet    []bool
	Transactions [][32]byte `json:"transactions,omitempty" form:"transactions" query:"transactions" validate:"required"`
	extraData    []byte
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

type TransactionChainIndex struct {
	fieldsSet []bool
	Entries   []*TransactionChainEntry `json:"entries,omitempty" form:"entries" query:"entries" validate:"required"`
	extraData []byte
}

func (v *BlockChainUpdatesIndex) Copy() *BlockChainUpdatesIndex {
	u := new(BlockChainUpdatesIndex)

	u.Entries = make([]*ChainUpdate, len(v.Entries))
	for i, v := range v.Entries {
		if v != nil {
			u.Entries[i] = (v).Copy()
		}
	}

	return u
}

func (v *BlockChainUpdatesIndex) CopyAsInterface() interface{} { return v.Copy() }

func (v *BlockStateIndex) Copy() *BlockStateIndex {
	u := new(BlockStateIndex)

	u.ProducedSynthTxns = make([]*BlockStateSynthTxnEntry, len(v.ProducedSynthTxns))
	for i, v := range v.ProducedSynthTxns {
		if v != nil {
			u.ProducedSynthTxns[i] = (v).Copy()
		}
	}

	return u
}

func (v *BlockStateIndex) CopyAsInterface() interface{} { return v.Copy() }

func (v *BlockStateSynthTxnEntry) Copy() *BlockStateSynthTxnEntry {
	u := new(BlockStateSynthTxnEntry)

	u.Transaction = encoding.BytesCopy(v.Transaction)
	u.ChainEntry = v.ChainEntry

	return u
}

func (v *BlockStateSynthTxnEntry) CopyAsInterface() interface{} { return v.Copy() }

func (v *ChainUpdate) Copy() *ChainUpdate {
	u := new(ChainUpdate)

	if v.Account != nil {
		u.Account = (v.Account).Copy()
	}
	u.Name = v.Name
	u.Type = v.Type
	u.Index = v.Index
	u.SourceIndex = v.SourceIndex
	u.SourceBlock = v.SourceBlock
	u.Entry = encoding.BytesCopy(v.Entry)

	return u
}

func (v *ChainUpdate) CopyAsInterface() interface{} { return v.Copy() }

func (v *PendingTransactionsIndex) Copy() *PendingTransactionsIndex {
	u := new(PendingTransactionsIndex)

	u.Transactions = make([][32]byte, len(v.Transactions))
	for i, v := range v.Transactions {
		u.Transactions[i] = v
	}

	return u
}

func (v *PendingTransactionsIndex) CopyAsInterface() interface{} { return v.Copy() }

func (v *TransactionChainEntry) Copy() *TransactionChainEntry {
	u := new(TransactionChainEntry)

	if v.Account != nil {
		u.Account = (v.Account).Copy()
	}
	u.Chain = v.Chain
	u.ChainIndex = v.ChainIndex
	u.AnchorIndex = v.AnchorIndex

	return u
}

func (v *TransactionChainEntry) CopyAsInterface() interface{} { return v.Copy() }

func (v *TransactionChainIndex) Copy() *TransactionChainIndex {
	u := new(TransactionChainIndex)

	u.Entries = make([]*TransactionChainEntry, len(v.Entries))
	for i, v := range v.Entries {
		if v != nil {
			u.Entries[i] = (v).Copy()
		}
	}

	return u
}

func (v *TransactionChainIndex) CopyAsInterface() interface{} { return v.Copy() }

func (v *BlockChainUpdatesIndex) Equal(u *BlockChainUpdatesIndex) bool {
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

func (v *BlockStateIndex) Equal(u *BlockStateIndex) bool {
	if len(v.ProducedSynthTxns) != len(u.ProducedSynthTxns) {
		return false
	}
	for i := range v.ProducedSynthTxns {
		if !((v.ProducedSynthTxns[i]).Equal(u.ProducedSynthTxns[i])) {
			return false
		}
	}

	return true
}

func (v *BlockStateSynthTxnEntry) Equal(u *BlockStateSynthTxnEntry) bool {
	if !(bytes.Equal(v.Transaction, u.Transaction)) {
		return false
	}
	if !(v.ChainEntry == u.ChainEntry) {
		return false
	}

	return true
}

func (v *ChainUpdate) Equal(u *ChainUpdate) bool {
	switch {
	case v.Account == u.Account:
		// equal
	case v.Account == nil || u.Account == nil:
		return false
	case !((v.Account).Equal(u.Account)):
		return false
	}
	if !(v.Name == u.Name) {
		return false
	}
	if !(v.Type == u.Type) {
		return false
	}
	if !(v.Index == u.Index) {
		return false
	}
	if !(v.SourceIndex == u.SourceIndex) {
		return false
	}
	if !(v.SourceBlock == u.SourceBlock) {
		return false
	}
	if !(bytes.Equal(v.Entry, u.Entry)) {
		return false
	}

	return true
}

func (v *PendingTransactionsIndex) Equal(u *PendingTransactionsIndex) bool {
	if len(v.Transactions) != len(u.Transactions) {
		return false
	}
	for i := range v.Transactions {
		if !(v.Transactions[i] == u.Transactions[i]) {
			return false
		}
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

func (v *TransactionChainIndex) Equal(u *TransactionChainIndex) bool {
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

var fieldNames_BlockChainUpdatesIndex = []string{

	1: "Entries",

	2: "extraData",
}

func (v *BlockChainUpdatesIndex) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(len(v.Entries) == 0) {
		for _, v := range v.Entries {
			writer.WriteValue(1, v)
		}
	}
	_, _, err := writer.Reset(fieldNames_BlockChainUpdatesIndex)
	if err != nil {
		return nil, err
	}

	buffer.Write(v.extraData)

	return buffer.Bytes(), err
}

func (v *BlockChainUpdatesIndex) IsValid() error {
	var errs []string

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

var fieldNames_BlockStateIndex = []string{

	1: "ProducedSynthTxns",

	2: "extraData",
}

func (v *BlockStateIndex) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(len(v.ProducedSynthTxns) == 0) {
		for _, v := range v.ProducedSynthTxns {
			writer.WriteValue(1, v)
		}
	}
	_, _, err := writer.Reset(fieldNames_BlockStateIndex)
	if err != nil {
		return nil, err
	}

	buffer.Write(v.extraData)

	return buffer.Bytes(), err
}

func (v *BlockStateIndex) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field ProducedSynthTxns is missing")
	} else if len(v.ProducedSynthTxns) == 0 {
		errs = append(errs, "field ProducedSynthTxns is not set")
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

var fieldNames_BlockStateSynthTxnEntry = []string{

	1: "Transaction",

	2: "ChainEntry",

	3: "extraData",
}

func (v *BlockStateSynthTxnEntry) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(len(v.Transaction) == 0) {
		writer.WriteBytes(1, v.Transaction)
	}
	if !(v.ChainEntry == 0) {
		writer.WriteUint(2, v.ChainEntry)
	}
	_, _, err := writer.Reset(fieldNames_BlockStateSynthTxnEntry)
	if err != nil {
		return nil, err
	}

	buffer.Write(v.extraData)

	return buffer.Bytes(), err
}

func (v *BlockStateSynthTxnEntry) IsValid() error {
	var errs []string

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

var fieldNames_ChainUpdate = []string{

	1: "Account",

	2: "Name",

	3: "Type",

	4: "Index",

	5: "SourceIndex",

	6: "SourceBlock",

	7: "Entry",

	8: "extraData",
}

func (v *ChainUpdate) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(v.Account == nil) {
		writer.WriteUrl(1, v.Account)
	}
	if !(len(v.Name) == 0) {
		writer.WriteString(2, v.Name)
	}
	if !(v.Type == 0) {
		writer.WriteEnum(3, v.Type)
	}
	if !(v.Index == 0) {
		writer.WriteUint(4, v.Index)
	}
	if !(v.SourceIndex == 0) {
		writer.WriteUint(5, v.SourceIndex)
	}
	if !(v.SourceBlock == 0) {
		writer.WriteUint(6, v.SourceBlock)
	}
	if !(len(v.Entry) == 0) {
		writer.WriteBytes(7, v.Entry)
	}
	_, _, err := writer.Reset(fieldNames_ChainUpdate)
	if err != nil {
		return nil, err
	}

	buffer.Write(v.extraData)

	return buffer.Bytes(), err
}

func (v *ChainUpdate) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Account is missing")
	} else if v.Account == nil {
		errs = append(errs, "field Account is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field Name is missing")
	} else if len(v.Name) == 0 {
		errs = append(errs, "field Name is not set")
	}
	if len(v.fieldsSet) > 3 && !v.fieldsSet[3] {
		errs = append(errs, "field Type is missing")
	} else if v.Type == 0 {
		errs = append(errs, "field Type is not set")
	}
	if len(v.fieldsSet) > 4 && !v.fieldsSet[4] {
		errs = append(errs, "field Index is missing")
	} else if v.Index == 0 {
		errs = append(errs, "field Index is not set")
	}
	if len(v.fieldsSet) > 5 && !v.fieldsSet[5] {
		errs = append(errs, "field SourceIndex is missing")
	} else if v.SourceIndex == 0 {
		errs = append(errs, "field SourceIndex is not set")
	}
	if len(v.fieldsSet) > 6 && !v.fieldsSet[6] {
		errs = append(errs, "field SourceBlock is missing")
	} else if v.SourceBlock == 0 {
		errs = append(errs, "field SourceBlock is not set")
	}
	if len(v.fieldsSet) > 7 && !v.fieldsSet[7] {
		errs = append(errs, "field Entry is missing")
	} else if len(v.Entry) == 0 {
		errs = append(errs, "field Entry is not set")
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

var fieldNames_PendingTransactionsIndex = []string{

	1: "Transactions",

	2: "extraData",
}

func (v *PendingTransactionsIndex) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(len(v.Transactions) == 0) {
		for _, v := range v.Transactions {
			writer.WriteHash(1, &v)
		}
	}
	_, _, err := writer.Reset(fieldNames_PendingTransactionsIndex)
	if err != nil {
		return nil, err
	}

	buffer.Write(v.extraData)

	return buffer.Bytes(), err
}

func (v *PendingTransactionsIndex) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Transactions is missing")
	} else if len(v.Transactions) == 0 {
		errs = append(errs, "field Transactions is not set")
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

	5: "extraData",
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
		return nil, err
	}

	buffer.Write(v.extraData)

	return buffer.Bytes(), err
}

func (v *TransactionChainEntry) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Account is missing")
	} else if v.Account == nil {
		errs = append(errs, "field Account is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field Chain is missing")
	} else if len(v.Chain) == 0 {
		errs = append(errs, "field Chain is not set")
	}
	if len(v.fieldsSet) > 3 && !v.fieldsSet[3] {
		errs = append(errs, "field ChainIndex is missing")
	} else if v.ChainIndex == 0 {
		errs = append(errs, "field ChainIndex is not set")
	}
	if len(v.fieldsSet) > 4 && !v.fieldsSet[4] {
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

var fieldNames_TransactionChainIndex = []string{

	1: "Entries",

	2: "extraData",
}

func (v *TransactionChainIndex) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(len(v.Entries) == 0) {
		for _, v := range v.Entries {
			writer.WriteValue(1, v)
		}
	}
	_, _, err := writer.Reset(fieldNames_TransactionChainIndex)
	if err != nil {
		return nil, err
	}

	buffer.Write(v.extraData)

	return buffer.Bytes(), err
}

func (v *TransactionChainIndex) IsValid() error {
	var errs []string

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

func (v *BlockChainUpdatesIndex) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *BlockChainUpdatesIndex) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	for {
		if x := new(ChainUpdate); reader.ReadValue(1, x.UnmarshalBinary) {
			v.Entries = append(v.Entries, x)
		} else {
			break
		}
	}

	seen, err := reader.Reset(fieldNames_BlockChainUpdatesIndex)
	if err != nil {
		return err
	}
	v.fieldsSet = seen
	v.extraData, err = reader.ReadAll()
	return err
}

func (v *BlockStateIndex) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *BlockStateIndex) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	for {
		if x := new(BlockStateSynthTxnEntry); reader.ReadValue(1, x.UnmarshalBinary) {
			v.ProducedSynthTxns = append(v.ProducedSynthTxns, x)
		} else {
			break
		}
	}

	seen, err := reader.Reset(fieldNames_BlockStateIndex)
	if err != nil {
		return err
	}
	v.fieldsSet = seen
	v.extraData, err = reader.ReadAll()
	return err
}

func (v *BlockStateSynthTxnEntry) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *BlockStateSynthTxnEntry) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x, ok := reader.ReadBytes(1); ok {
		v.Transaction = x
	}
	if x, ok := reader.ReadUint(2); ok {
		v.ChainEntry = x
	}

	seen, err := reader.Reset(fieldNames_BlockStateSynthTxnEntry)
	if err != nil {
		return err
	}
	v.fieldsSet = seen
	v.extraData, err = reader.ReadAll()
	return err
}

func (v *ChainUpdate) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *ChainUpdate) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x, ok := reader.ReadUrl(1); ok {
		v.Account = x
	}
	if x, ok := reader.ReadString(2); ok {
		v.Name = x
	}
	if x := new(protocol.ChainType); reader.ReadEnum(3, x) {
		v.Type = *x
	}
	if x, ok := reader.ReadUint(4); ok {
		v.Index = x
	}
	if x, ok := reader.ReadUint(5); ok {
		v.SourceIndex = x
	}
	if x, ok := reader.ReadUint(6); ok {
		v.SourceBlock = x
	}
	if x, ok := reader.ReadBytes(7); ok {
		v.Entry = x
	}

	seen, err := reader.Reset(fieldNames_ChainUpdate)
	if err != nil {
		return err
	}
	v.fieldsSet = seen
	v.extraData, err = reader.ReadAll()
	return err
}

func (v *PendingTransactionsIndex) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *PendingTransactionsIndex) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	for {
		if x, ok := reader.ReadHash(1); ok {
			v.Transactions = append(v.Transactions, *x)
		} else {
			break
		}
	}

	seen, err := reader.Reset(fieldNames_PendingTransactionsIndex)
	if err != nil {
		return err
	}
	v.fieldsSet = seen
	v.extraData, err = reader.ReadAll()
	return err
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
		return err
	}
	v.fieldsSet = seen
	v.extraData, err = reader.ReadAll()
	return err
}

func (v *TransactionChainIndex) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *TransactionChainIndex) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	for {
		if x := new(TransactionChainEntry); reader.ReadValue(1, x.UnmarshalBinary) {
			v.Entries = append(v.Entries, x)
		} else {
			break
		}
	}

	seen, err := reader.Reset(fieldNames_TransactionChainIndex)
	if err != nil {
		return err
	}
	v.fieldsSet = seen
	v.extraData, err = reader.ReadAll()
	return err
}

func (v *BlockChainUpdatesIndex) MarshalJSON() ([]byte, error) {
	u := struct {
		Entries encoding.JsonList[*ChainUpdate] `json:"entries,omitempty"`
	}{}
	u.Entries = v.Entries
	return json.Marshal(&u)
}

func (v *BlockStateIndex) MarshalJSON() ([]byte, error) {
	u := struct {
		ProducedSynthTxns encoding.JsonList[*BlockStateSynthTxnEntry] `json:"producedSynthTxns,omitempty"`
	}{}
	u.ProducedSynthTxns = v.ProducedSynthTxns
	return json.Marshal(&u)
}

func (v *BlockStateSynthTxnEntry) MarshalJSON() ([]byte, error) {
	u := struct {
		Transaction *string `json:"transaction,omitempty"`
		ChainEntry  uint64  `json:"chainEntry,omitempty"`
	}{}
	u.Transaction = encoding.BytesToJSON(v.Transaction)
	u.ChainEntry = v.ChainEntry
	return json.Marshal(&u)
}

func (v *ChainUpdate) MarshalJSON() ([]byte, error) {
	u := struct {
		Account     *url.URL           `json:"account,omitempty"`
		Name        string             `json:"name,omitempty"`
		Type        protocol.ChainType `json:"type,omitempty"`
		Index       uint64             `json:"index,omitempty"`
		SourceIndex uint64             `json:"sourceIndex,omitempty"`
		SourceBlock uint64             `json:"sourceBlock,omitempty"`
		Entry       *string            `json:"entry,omitempty"`
	}{}
	u.Account = v.Account
	u.Name = v.Name
	u.Type = v.Type
	u.Index = v.Index
	u.SourceIndex = v.SourceIndex
	u.SourceBlock = v.SourceBlock
	u.Entry = encoding.BytesToJSON(v.Entry)
	return json.Marshal(&u)
}

func (v *PendingTransactionsIndex) MarshalJSON() ([]byte, error) {
	u := struct {
		Transactions encoding.JsonList[string] `json:"transactions,omitempty"`
	}{}
	u.Transactions = make(encoding.JsonList[string], len(v.Transactions))
	for i, x := range v.Transactions {
		u.Transactions[i] = encoding.ChainToJSON(x)
	}
	return json.Marshal(&u)
}

func (v *TransactionChainIndex) MarshalJSON() ([]byte, error) {
	u := struct {
		Entries encoding.JsonList[*TransactionChainEntry] `json:"entries,omitempty"`
	}{}
	u.Entries = v.Entries
	return json.Marshal(&u)
}

func (v *BlockChainUpdatesIndex) UnmarshalJSON(data []byte) error {
	u := struct {
		Entries encoding.JsonList[*ChainUpdate] `json:"entries,omitempty"`
	}{}
	u.Entries = v.Entries
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Entries = u.Entries
	return nil
}

func (v *BlockStateIndex) UnmarshalJSON(data []byte) error {
	u := struct {
		ProducedSynthTxns encoding.JsonList[*BlockStateSynthTxnEntry] `json:"producedSynthTxns,omitempty"`
	}{}
	u.ProducedSynthTxns = v.ProducedSynthTxns
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.ProducedSynthTxns = u.ProducedSynthTxns
	return nil
}

func (v *BlockStateSynthTxnEntry) UnmarshalJSON(data []byte) error {
	u := struct {
		Transaction *string `json:"transaction,omitempty"`
		ChainEntry  uint64  `json:"chainEntry,omitempty"`
	}{}
	u.Transaction = encoding.BytesToJSON(v.Transaction)
	u.ChainEntry = v.ChainEntry
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	if x, err := encoding.BytesFromJSON(u.Transaction); err != nil {
		return fmt.Errorf("error decoding Transaction: %w", err)
	} else {
		v.Transaction = x
	}
	v.ChainEntry = u.ChainEntry
	return nil
}

func (v *ChainUpdate) UnmarshalJSON(data []byte) error {
	u := struct {
		Account     *url.URL           `json:"account,omitempty"`
		Name        string             `json:"name,omitempty"`
		Type        protocol.ChainType `json:"type,omitempty"`
		Index       uint64             `json:"index,omitempty"`
		SourceIndex uint64             `json:"sourceIndex,omitempty"`
		SourceBlock uint64             `json:"sourceBlock,omitempty"`
		Entry       *string            `json:"entry,omitempty"`
	}{}
	u.Account = v.Account
	u.Name = v.Name
	u.Type = v.Type
	u.Index = v.Index
	u.SourceIndex = v.SourceIndex
	u.SourceBlock = v.SourceBlock
	u.Entry = encoding.BytesToJSON(v.Entry)
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Account = u.Account
	v.Name = u.Name
	v.Type = u.Type
	v.Index = u.Index
	v.SourceIndex = u.SourceIndex
	v.SourceBlock = u.SourceBlock
	if x, err := encoding.BytesFromJSON(u.Entry); err != nil {
		return fmt.Errorf("error decoding Entry: %w", err)
	} else {
		v.Entry = x
	}
	return nil
}

func (v *PendingTransactionsIndex) UnmarshalJSON(data []byte) error {
	u := struct {
		Transactions encoding.JsonList[string] `json:"transactions,omitempty"`
	}{}
	u.Transactions = make(encoding.JsonList[string], len(v.Transactions))
	for i, x := range v.Transactions {
		u.Transactions[i] = encoding.ChainToJSON(x)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Transactions = make([][32]byte, len(u.Transactions))
	for i, x := range u.Transactions {
		if x, err := encoding.ChainFromJSON(x); err != nil {
			return fmt.Errorf("error decoding Transactions: %w", err)
		} else {
			v.Transactions[i] = x
		}
	}
	return nil
}

func (v *TransactionChainIndex) UnmarshalJSON(data []byte) error {
	u := struct {
		Entries encoding.JsonList[*TransactionChainEntry] `json:"entries,omitempty"`
	}{}
	u.Entries = v.Entries
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Entries = u.Entries
	return nil
}
