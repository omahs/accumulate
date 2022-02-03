package protocol

// GENERATED BY go run ./tools/cmd/gen-enum. DO NOT EDIT.

import (
	"encoding/json"
	"fmt"
	"strings"

	"gitlab.com/accumulatenetwork/accumulate/internal/encoding"
)

// ChainTypeUnknown is used when the chain type is not known.
const ChainTypeUnknown ChainType = 0

// ChainTypeTransaction holds transaction hashes.
const ChainTypeTransaction ChainType = 1

// ChainTypeAnchor holds chain anchors.
const ChainTypeAnchor ChainType = 2

// ChainTypeData holds data entry hashes.
const ChainTypeData ChainType = 3

// KeyPageOperationUnknown is used when the key page operation is not known.
const KeyPageOperationUnknown KeyPageOperation = 0

// KeyPageOperationUpdate replaces a key in the page with a new key.
const KeyPageOperationUpdate KeyPageOperation = 1

// KeyPageOperationRemove removes a key from the page.
const KeyPageOperationRemove KeyPageOperation = 2

// KeyPageOperationAdd adds a key to the page.
const KeyPageOperationAdd KeyPageOperation = 3

// KeyPageOperationSetThreshold sets the signing threshold (the M of "M of N" signatures required).
const KeyPageOperationSetThreshold KeyPageOperation = 4

// ObjectTypeUnknown is used when the object type is not known.
const ObjectTypeUnknown ObjectType = 0

// ObjectTypeAccount represents an account object.
const ObjectTypeAccount ObjectType = 1

// ObjectTypeTransaction represents a transaction object.
const ObjectTypeTransaction ObjectType = 2

// TransactionMaxUser is the highest number reserved for user transactions.
const TransactionMaxUser TransactionMax = 47

// TransactionMaxSynthetic is the highest number reserved for synthetic transactions.
const TransactionMaxSynthetic TransactionMax = 95

// TransactionMaxInternal is the highest number reserved for internal transactions.
const TransactionMaxInternal TransactionMax = 255

// TransactionTypeUnknown represents an unknown transaction type.
const TransactionTypeUnknown TransactionType = 0

// TransactionTypeCreateIdentity creates an ADI, which produces a synthetic chain.
const TransactionTypeCreateIdentity TransactionType = 1

// TransactionTypeCreateTokenAccount creates an ADI token account, which produces a synthetic chain create transaction.
const TransactionTypeCreateTokenAccount TransactionType = 2

// TransactionTypeSendTokens transfers tokens between token accounts, which produces a synthetic deposit tokens transaction.
const TransactionTypeSendTokens TransactionType = 3

// TransactionTypeCreateDataAccount creates an ADI Data Account, which produces a synthetic chain create transaction.
const TransactionTypeCreateDataAccount TransactionType = 4

// TransactionTypeWriteData writes data to an ADI Data Account, which *does not* produce a synthetic transaction.
const TransactionTypeWriteData TransactionType = 5

// TransactionTypeWriteDataTo writes data to a Lite Data Account, which produces a synthetic write data transaction.
const TransactionTypeWriteDataTo TransactionType = 6

// TransactionTypeAcmeFaucet produces a synthetic deposit tokens transaction that deposits ACME tokens into a lite token account.
const TransactionTypeAcmeFaucet TransactionType = 7

// TransactionTypeCreateToken creates a token issuer, which produces a synthetic chain create transaction.
const TransactionTypeCreateToken TransactionType = 8

// TransactionTypeIssueTokens issues tokens to a token account, which produces a synthetic token deposit transaction.
const TransactionTypeIssueTokens TransactionType = 9

// TransactionTypeBurnTokens burns tokens from a token account, which produces a synthetic burn tokens transaction.
const TransactionTypeBurnTokens TransactionType = 10

// TransactionTypeCreateKeyPage creates a key page, which produces a synthetic chain create transaction.
const TransactionTypeCreateKeyPage TransactionType = 12

// TransactionTypeCreateKeyBook creates a key book, which produces a synthetic chain create transaction.
const TransactionTypeCreateKeyBook TransactionType = 13

// TransactionTypeAddCredits converts ACME tokens to credits, which produces a synthetic deposit credits transaction.
const TransactionTypeAddCredits TransactionType = 14

// TransactionTypeUpdateKeyPage adds, removes, or updates keys in a key page, which *does not* produce a synthetic transaction.
const TransactionTypeUpdateKeyPage TransactionType = 15

// TransactionTypeSignPending is used to sign a pending transaction.
const TransactionTypeSignPending TransactionType = 48

// TransactionTypeSyntheticCreateChain creates or updates chains.
const TransactionTypeSyntheticCreateChain TransactionType = 49

// TransactionTypeSyntheticWriteData writes data to a data account.
const TransactionTypeSyntheticWriteData TransactionType = 50

// TransactionTypeSyntheticDepositTokens deposits tokens into token accounts.
const TransactionTypeSyntheticDepositTokens TransactionType = 51

// TransactionTypeSyntheticAnchor anchors one network to another.
const TransactionTypeSyntheticAnchor TransactionType = 52

// TransactionTypeSyntheticDepositCredits deposits credits into a credit holder.
const TransactionTypeSyntheticDepositCredits TransactionType = 53

// TransactionTypeSyntheticBurnTokens returns tokens to a token issuer's pool of issuable tokens.
const TransactionTypeSyntheticBurnTokens TransactionType = 54

// TransactionTypeSyntheticMirror mirrors records from one network to another.
const TransactionTypeSyntheticMirror TransactionType = 56

// TransactionTypeSegWitDataEntry is a surrogate transaction segregated witness for a WriteData transaction.
const TransactionTypeSegWitDataEntry TransactionType = 57

// TransactionTypeInternalGenesis initializes system chains.
const TransactionTypeInternalGenesis TransactionType = 96

// TransactionTypeInternalSendTransactions reserved for internal send.
const TransactionTypeInternalSendTransactions TransactionType = 97

// TransactionTypeInternalTransactionsSigned notifies the executor of synthetic transactions that have been signed.
const TransactionTypeInternalTransactionsSigned TransactionType = 98

// TransactionTypeInternalTransactionsSent notifies the executor of synthetic transactions that have been sent.
const TransactionTypeInternalTransactionsSent TransactionType = 99

// ID returns the ID of the Chain Type
func (v ChainType) ID() uint64 { return uint64(v) }

// String returns the name of the Chain Type
func (v ChainType) String() string {
	switch v {
	case ChainTypeUnknown:
		return "unknown"
	case ChainTypeTransaction:
		return "transaction"
	case ChainTypeAnchor:
		return "anchor"
	case ChainTypeData:
		return "data"
	default:
		return fmt.Sprintf("ChainType:%d", v)
	}
}

// ChainTypeByName returns the named Chain Type.
func ChainTypeByName(name string) (ChainType, bool) {
	switch name {
	case "unknown":
		return ChainTypeUnknown, true
	case "transaction":
		return ChainTypeTransaction, true
	case "anchor":
		return ChainTypeAnchor, true
	case "data":
		return ChainTypeData, true
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

// BinarySize returns the number of bytes required to binary marshal the Chain Type.
func (v ChainType) BinarySize() int {
	return encoding.UvarintBinarySize(v.ID())
}

// MarshalBinary marshals the Chain Type to bytes as a unsigned varint.
func (v ChainType) MarshalBinary() ([]byte, error) {
	return encoding.UvarintMarshalBinary(v.ID()), nil
}

// UnmarshalBinary unmarshals the Chain Type from bytes as a unsigned varint.
func (v *ChainType) UnmarshalBinary(data []byte) error {
	u, err := encoding.UvarintUnmarshalBinary(data)
	if err != nil {
		return err
	}

	*v = ChainType(u)
	return nil
}

// ID returns the ID of the Key PageOpe ration
func (v KeyPageOperation) ID() uint64 { return uint64(v) }

// String returns the name of the Key PageOpe ration
func (v KeyPageOperation) String() string {
	switch v {
	case KeyPageOperationUnknown:
		return "unknown"
	case KeyPageOperationUpdate:
		return "update"
	case KeyPageOperationRemove:
		return "remove"
	case KeyPageOperationAdd:
		return "add"
	case KeyPageOperationSetThreshold:
		return "setThreshold"
	default:
		return fmt.Sprintf("KeyPageOperation:%d", v)
	}
}

// KeyPageOperationByName returns the named Key PageOpe ration.
func KeyPageOperationByName(name string) (KeyPageOperation, bool) {
	switch name {
	case "unknown":
		return KeyPageOperationUnknown, true
	case "update":
		return KeyPageOperationUpdate, true
	case "remove":
		return KeyPageOperationRemove, true
	case "add":
		return KeyPageOperationAdd, true
	case "setThreshold":
		return KeyPageOperationSetThreshold, true
	default:
		return 0, false
	}
}

// MarshalJSON marshals the Key PageOpe ration to JSON as a string.
func (v KeyPageOperation) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

// UnmarshalJSON unmarshals the Key PageOpe ration from JSON as a string.
func (v *KeyPageOperation) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	var ok bool
	*v, ok = KeyPageOperationByName(s)
	if !ok || strings.ContainsRune(v.String(), ':') {
		return fmt.Errorf("invalid Key PageOpe ration %q", s)
	}
	return nil
}

// BinarySize returns the number of bytes required to binary marshal the Key PageOpe ration.
func (v KeyPageOperation) BinarySize() int {
	return encoding.UvarintBinarySize(v.ID())
}

// MarshalBinary marshals the Key PageOpe ration to bytes as a unsigned varint.
func (v KeyPageOperation) MarshalBinary() ([]byte, error) {
	return encoding.UvarintMarshalBinary(v.ID()), nil
}

// UnmarshalBinary unmarshals the Key PageOpe ration from bytes as a unsigned varint.
func (v *KeyPageOperation) UnmarshalBinary(data []byte) error {
	u, err := encoding.UvarintUnmarshalBinary(data)
	if err != nil {
		return err
	}

	*v = KeyPageOperation(u)
	return nil
}

// ID returns the ID of the Object Type
func (v ObjectType) ID() uint64 { return uint64(v) }

// String returns the name of the Object Type
func (v ObjectType) String() string {
	switch v {
	case ObjectTypeUnknown:
		return "unknown"
	case ObjectTypeAccount:
		return "account"
	case ObjectTypeTransaction:
		return "transaction"
	default:
		return fmt.Sprintf("ObjectType:%d", v)
	}
}

// ObjectTypeByName returns the named Object Type.
func ObjectTypeByName(name string) (ObjectType, bool) {
	switch name {
	case "unknown":
		return ObjectTypeUnknown, true
	case "account":
		return ObjectTypeAccount, true
	case "transaction":
		return ObjectTypeTransaction, true
	default:
		return 0, false
	}
}

// MarshalJSON marshals the Object Type to JSON as a string.
func (v ObjectType) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

// UnmarshalJSON unmarshals the Object Type from JSON as a string.
func (v *ObjectType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	var ok bool
	*v, ok = ObjectTypeByName(s)
	if !ok || strings.ContainsRune(v.String(), ':') {
		return fmt.Errorf("invalid Object Type %q", s)
	}
	return nil
}

// BinarySize returns the number of bytes required to binary marshal the Object Type.
func (v ObjectType) BinarySize() int {
	return encoding.UvarintBinarySize(v.ID())
}

// MarshalBinary marshals the Object Type to bytes as a unsigned varint.
func (v ObjectType) MarshalBinary() ([]byte, error) {
	return encoding.UvarintMarshalBinary(v.ID()), nil
}

// UnmarshalBinary unmarshals the Object Type from bytes as a unsigned varint.
func (v *ObjectType) UnmarshalBinary(data []byte) error {
	u, err := encoding.UvarintUnmarshalBinary(data)
	if err != nil {
		return err
	}

	*v = ObjectType(u)
	return nil
}

// ID returns the ID of the Transaction Max
func (v TransactionMax) ID() uint64 { return uint64(v) }

// String returns the name of the Transaction Max
func (v TransactionMax) String() string {
	switch v {
	case TransactionMaxUser:
		return "user"
	case TransactionMaxSynthetic:
		return "synthetic"
	case TransactionMaxInternal:
		return "internal"
	default:
		return fmt.Sprintf("TransactionMax:%d", v)
	}
}

// TransactionMaxByName returns the named Transaction Max.
func TransactionMaxByName(name string) (TransactionMax, bool) {
	switch name {
	case "user":
		return TransactionMaxUser, true
	case "synthetic":
		return TransactionMaxSynthetic, true
	case "internal":
		return TransactionMaxInternal, true
	default:
		return 0, false
	}
}

// MarshalJSON marshals the Transaction Max to JSON as a string.
func (v TransactionMax) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

// UnmarshalJSON unmarshals the Transaction Max from JSON as a string.
func (v *TransactionMax) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	var ok bool
	*v, ok = TransactionMaxByName(s)
	if !ok || strings.ContainsRune(v.String(), ':') {
		return fmt.Errorf("invalid Transaction Max %q", s)
	}
	return nil
}

// BinarySize returns the number of bytes required to binary marshal the Transaction Max.
func (v TransactionMax) BinarySize() int {
	return encoding.UvarintBinarySize(v.ID())
}

// MarshalBinary marshals the Transaction Max to bytes as a unsigned varint.
func (v TransactionMax) MarshalBinary() ([]byte, error) {
	return encoding.UvarintMarshalBinary(v.ID()), nil
}

// UnmarshalBinary unmarshals the Transaction Max from bytes as a unsigned varint.
func (v *TransactionMax) UnmarshalBinary(data []byte) error {
	u, err := encoding.UvarintUnmarshalBinary(data)
	if err != nil {
		return err
	}

	*v = TransactionMax(u)
	return nil
}

// ID returns the ID of the Transaction Type
func (v TransactionType) ID() uint64 { return uint64(v) }

// String returns the name of the Transaction Type
func (v TransactionType) String() string {
	switch v {
	case TransactionTypeUnknown:
		return "unknown"
	case TransactionTypeCreateIdentity:
		return "createIdentity"
	case TransactionTypeCreateTokenAccount:
		return "createTokenAccount"
	case TransactionTypeSendTokens:
		return "sendTokens"
	case TransactionTypeCreateDataAccount:
		return "createDataAccount"
	case TransactionTypeWriteData:
		return "writeData"
	case TransactionTypeWriteDataTo:
		return "writeDataTo"
	case TransactionTypeAcmeFaucet:
		return "acmeFaucet"
	case TransactionTypeCreateToken:
		return "createToken"
	case TransactionTypeIssueTokens:
		return "issueTokens"
	case TransactionTypeBurnTokens:
		return "burnTokens"
	case TransactionTypeCreateKeyPage:
		return "createKeyPage"
	case TransactionTypeCreateKeyBook:
		return "createKeyBook"
	case TransactionTypeAddCredits:
		return "addCredits"
	case TransactionTypeUpdateKeyPage:
		return "updateKeyPage"
	case TransactionTypeSignPending:
		return "signPending"
	case TransactionTypeSyntheticCreateChain:
		return "syntheticCreateChain"
	case TransactionTypeSyntheticWriteData:
		return "syntheticWriteData"
	case TransactionTypeSyntheticDepositTokens:
		return "syntheticDepositTokens"
	case TransactionTypeSyntheticAnchor:
		return "syntheticAnchor"
	case TransactionTypeSyntheticDepositCredits:
		return "syntheticDepositCredits"
	case TransactionTypeSyntheticBurnTokens:
		return "syntheticBurnTokens"
	case TransactionTypeSyntheticMirror:
		return "syntheticMirror"
	case TransactionTypeSegWitDataEntry:
		return "segWitDataEntry"
	case TransactionTypeInternalGenesis:
		return "internalGenesis"
	case TransactionTypeInternalSendTransactions:
		return "internalSendTransactions"
	case TransactionTypeInternalTransactionsSigned:
		return "internalTransactionsSigned"
	case TransactionTypeInternalTransactionsSent:
		return "internalTransactionsSent"
	default:
		return fmt.Sprintf("TransactionType:%d", v)
	}
}

// TransactionTypeByName returns the named Transaction Type.
func TransactionTypeByName(name string) (TransactionType, bool) {
	switch name {
	case "unknown":
		return TransactionTypeUnknown, true
	case "createIdentity":
		return TransactionTypeCreateIdentity, true
	case "createTokenAccount":
		return TransactionTypeCreateTokenAccount, true
	case "sendTokens":
		return TransactionTypeSendTokens, true
	case "createDataAccount":
		return TransactionTypeCreateDataAccount, true
	case "writeData":
		return TransactionTypeWriteData, true
	case "writeDataTo":
		return TransactionTypeWriteDataTo, true
	case "acmeFaucet":
		return TransactionTypeAcmeFaucet, true
	case "createToken":
		return TransactionTypeCreateToken, true
	case "issueTokens":
		return TransactionTypeIssueTokens, true
	case "burnTokens":
		return TransactionTypeBurnTokens, true
	case "createKeyPage":
		return TransactionTypeCreateKeyPage, true
	case "createKeyBook":
		return TransactionTypeCreateKeyBook, true
	case "addCredits":
		return TransactionTypeAddCredits, true
	case "updateKeyPage":
		return TransactionTypeUpdateKeyPage, true
	case "signPending":
		return TransactionTypeSignPending, true
	case "syntheticCreateChain":
		return TransactionTypeSyntheticCreateChain, true
	case "syntheticWriteData":
		return TransactionTypeSyntheticWriteData, true
	case "syntheticDepositTokens":
		return TransactionTypeSyntheticDepositTokens, true
	case "syntheticAnchor":
		return TransactionTypeSyntheticAnchor, true
	case "syntheticDepositCredits":
		return TransactionTypeSyntheticDepositCredits, true
	case "syntheticBurnTokens":
		return TransactionTypeSyntheticBurnTokens, true
	case "syntheticMirror":
		return TransactionTypeSyntheticMirror, true
	case "segWitDataEntry":
		return TransactionTypeSegWitDataEntry, true
	case "internalGenesis":
		return TransactionTypeInternalGenesis, true
	case "internalSendTransactions":
		return TransactionTypeInternalSendTransactions, true
	case "internalTransactionsSigned":
		return TransactionTypeInternalTransactionsSigned, true
	case "internalTransactionsSent":
		return TransactionTypeInternalTransactionsSent, true
	default:
		return 0, false
	}
}

// MarshalJSON marshals the Transaction Type to JSON as a string.
func (v TransactionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

// UnmarshalJSON unmarshals the Transaction Type from JSON as a string.
func (v *TransactionType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	var ok bool
	*v, ok = TransactionTypeByName(s)
	if !ok || strings.ContainsRune(v.String(), ':') {
		return fmt.Errorf("invalid Transaction Type %q", s)
	}
	return nil
}

// BinarySize returns the number of bytes required to binary marshal the Transaction Type.
func (v TransactionType) BinarySize() int {
	return encoding.UvarintBinarySize(v.ID())
}

// MarshalBinary marshals the Transaction Type to bytes as a unsigned varint.
func (v TransactionType) MarshalBinary() ([]byte, error) {
	return encoding.UvarintMarshalBinary(v.ID()), nil
}

// UnmarshalBinary unmarshals the Transaction Type from bytes as a unsigned varint.
func (v *TransactionType) UnmarshalBinary(data []byte) error {
	u, err := encoding.UvarintUnmarshalBinary(data)
	if err != nil {
		return err
	}

	*v = TransactionType(u)
	return nil
}
