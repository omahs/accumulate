// Copyright 2023 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package api

// GENERATED BY go run ./tools/cmd/gen-enum. DO NOT EDIT.

import (
	"encoding/json"
	"fmt"
	"strings"
)

// EventTypeError .
const EventTypeError EventType = 1

// EventTypeBlock .
const EventTypeBlock EventType = 2

// EventTypeGlobals .
const EventTypeGlobals EventType = 3

// QueryTypeDefault .
const QueryTypeDefault QueryType = 0

// QueryTypeChain .
const QueryTypeChain QueryType = 1

// QueryTypeData .
const QueryTypeData QueryType = 2

// QueryTypeDirectory .
const QueryTypeDirectory QueryType = 3

// QueryTypePending .
const QueryTypePending QueryType = 4

// QueryTypeBlock .
const QueryTypeBlock QueryType = 5

// QueryTypeAnchorSearch .
const QueryTypeAnchorSearch QueryType = 16

// QueryTypePublicKeySearch .
const QueryTypePublicKeySearch QueryType = 17

// QueryTypePublicKeyHashSearch .
const QueryTypePublicKeyHashSearch QueryType = 18

// QueryTypeDelegateSearch .
const QueryTypeDelegateSearch QueryType = 19

// QueryTypeTransactionHashSearch .
const QueryTypeTransactionHashSearch QueryType = 20

// RecordTypeAccount .
const RecordTypeAccount RecordType = 1

// RecordTypeChain .
const RecordTypeChain RecordType = 2

// RecordTypeChainEntry .
const RecordTypeChainEntry RecordType = 3

// RecordTypeKey .
const RecordTypeKey RecordType = 4

// RecordTypeTransaction .
const RecordTypeTransaction RecordType = 16

// RecordTypeSignature .
const RecordTypeSignature RecordType = 17

// RecordTypeMinorBlock .
const RecordTypeMinorBlock RecordType = 32

// RecordTypeMajorBlock .
const RecordTypeMajorBlock RecordType = 33

// RecordTypeRange .
const RecordTypeRange RecordType = 128

// RecordTypeUrl .
const RecordTypeUrl RecordType = 129

// RecordTypeTxID .
const RecordTypeTxID RecordType = 130

// RecordTypeIndexEntry .
const RecordTypeIndexEntry RecordType = 131

// ServiceTypeUnknown .
const ServiceTypeUnknown ServiceType = 0

// ServiceTypeNode .
const ServiceTypeNode ServiceType = 1

// ServiceTypeNetwork .
const ServiceTypeNetwork ServiceType = 2

// ServiceTypeMetrics .
const ServiceTypeMetrics ServiceType = 3

// ServiceTypeQuery .
const ServiceTypeQuery ServiceType = 4

// ServiceTypeEvent .
const ServiceTypeEvent ServiceType = 5

// ServiceTypeSubmit .
const ServiceTypeSubmit ServiceType = 6

// ServiceTypeValidate .
const ServiceTypeValidate ServiceType = 7

// ServiceTypeFaucet .
const ServiceTypeFaucet ServiceType = 8

// GetEnumValue returns the value of the Event Type
func (v EventType) GetEnumValue() uint64 { return uint64(v) }

// SetEnumValue sets the value. SetEnumValue returns false if the value is invalid.
func (v *EventType) SetEnumValue(id uint64) bool {
	u := EventType(id)
	switch u {
	case EventTypeError, EventTypeBlock, EventTypeGlobals:
		*v = u
		return true
	default:
		return false
	}
}

// String returns the name of the Event Type.
func (v EventType) String() string {
	switch v {
	case EventTypeError:
		return "error"
	case EventTypeBlock:
		return "block"
	case EventTypeGlobals:
		return "globals"
	default:
		return fmt.Sprintf("EventType:%d", v)
	}
}

// EventTypeByName returns the named Event Type.
func EventTypeByName(name string) (EventType, bool) {
	switch strings.ToLower(name) {
	case "error":
		return EventTypeError, true
	case "block":
		return EventTypeBlock, true
	case "globals":
		return EventTypeGlobals, true
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

// GetEnumValue returns the value of the Query Type
func (v QueryType) GetEnumValue() uint64 { return uint64(v) }

// SetEnumValue sets the value. SetEnumValue returns false if the value is invalid.
func (v *QueryType) SetEnumValue(id uint64) bool {
	u := QueryType(id)
	switch u {
	case QueryTypeDefault, QueryTypeChain, QueryTypeData, QueryTypeDirectory, QueryTypePending, QueryTypeBlock, QueryTypeAnchorSearch, QueryTypePublicKeySearch, QueryTypePublicKeyHashSearch, QueryTypeDelegateSearch, QueryTypeTransactionHashSearch:
		*v = u
		return true
	default:
		return false
	}
}

// String returns the name of the Query Type.
func (v QueryType) String() string {
	switch v {
	case QueryTypeDefault:
		return "default"
	case QueryTypeChain:
		return "chain"
	case QueryTypeData:
		return "data"
	case QueryTypeDirectory:
		return "directory"
	case QueryTypePending:
		return "pending"
	case QueryTypeBlock:
		return "block"
	case QueryTypeAnchorSearch:
		return "anchorSearch"
	case QueryTypePublicKeySearch:
		return "publicKeySearch"
	case QueryTypePublicKeyHashSearch:
		return "publicKeyHashSearch"
	case QueryTypeDelegateSearch:
		return "delegateSearch"
	case QueryTypeTransactionHashSearch:
		return "transactionHashSearch"
	default:
		return fmt.Sprintf("QueryType:%d", v)
	}
}

// QueryTypeByName returns the named Query Type.
func QueryTypeByName(name string) (QueryType, bool) {
	switch strings.ToLower(name) {
	case "default":
		return QueryTypeDefault, true
	case "chain":
		return QueryTypeChain, true
	case "data":
		return QueryTypeData, true
	case "directory":
		return QueryTypeDirectory, true
	case "pending":
		return QueryTypePending, true
	case "block":
		return QueryTypeBlock, true
	case "anchorsearch":
		return QueryTypeAnchorSearch, true
	case "publickeysearch":
		return QueryTypePublicKeySearch, true
	case "publickeyhashsearch":
		return QueryTypePublicKeyHashSearch, true
	case "delegatesearch":
		return QueryTypeDelegateSearch, true
	case "transactionhashsearch":
		return QueryTypeTransactionHashSearch, true
	default:
		return 0, false
	}
}

// MarshalJSON marshals the Query Type to JSON as a string.
func (v QueryType) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

// UnmarshalJSON unmarshals the Query Type from JSON as a string.
func (v *QueryType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	var ok bool
	*v, ok = QueryTypeByName(s)
	if !ok || strings.ContainsRune(v.String(), ':') {
		return fmt.Errorf("invalid Query Type %q", s)
	}
	return nil
}

// GetEnumValue returns the value of the Record Type
func (v RecordType) GetEnumValue() uint64 { return uint64(v) }

// SetEnumValue sets the value. SetEnumValue returns false if the value is invalid.
func (v *RecordType) SetEnumValue(id uint64) bool {
	u := RecordType(id)
	switch u {
	case RecordTypeAccount, RecordTypeChain, RecordTypeChainEntry, RecordTypeKey, RecordTypeTransaction, RecordTypeSignature, RecordTypeMinorBlock, RecordTypeMajorBlock, RecordTypeRange, RecordTypeUrl, RecordTypeTxID, RecordTypeIndexEntry:
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
	case RecordTypeKey:
		return "key"
	case RecordTypeTransaction:
		return "transaction"
	case RecordTypeSignature:
		return "signature"
	case RecordTypeMinorBlock:
		return "minorBlock"
	case RecordTypeMajorBlock:
		return "majorBlock"
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
	case "key":
		return RecordTypeKey, true
	case "transaction":
		return RecordTypeTransaction, true
	case "signature":
		return RecordTypeSignature, true
	case "minorblock":
		return RecordTypeMinorBlock, true
	case "majorblock":
		return RecordTypeMajorBlock, true
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

// GetEnumValue returns the value of the Service Type
func (v ServiceType) GetEnumValue() uint64 { return uint64(v) }

// SetEnumValue sets the value. SetEnumValue returns false if the value is invalid.
func (v *ServiceType) SetEnumValue(id uint64) bool {
	u := ServiceType(id)
	switch u {
	case ServiceTypeUnknown, ServiceTypeNode, ServiceTypeNetwork, ServiceTypeMetrics, ServiceTypeQuery, ServiceTypeEvent, ServiceTypeSubmit, ServiceTypeValidate, ServiceTypeFaucet:
		*v = u
		return true
	default:
		return false
	}
}

// String returns the name of the Service Type.
func (v ServiceType) String() string {
	switch v {
	case ServiceTypeUnknown:
		return "unknown"
	case ServiceTypeNode:
		return "node"
	case ServiceTypeNetwork:
		return "network"
	case ServiceTypeMetrics:
		return "metrics"
	case ServiceTypeQuery:
		return "query"
	case ServiceTypeEvent:
		return "event"
	case ServiceTypeSubmit:
		return "submit"
	case ServiceTypeValidate:
		return "validate"
	case ServiceTypeFaucet:
		return "faucet"
	default:
		return fmt.Sprintf("ServiceType:%d", v)
	}
}

// ServiceTypeByName returns the named Service Type.
func ServiceTypeByName(name string) (ServiceType, bool) {
	switch strings.ToLower(name) {
	case "unknown":
		return ServiceTypeUnknown, true
	case "node":
		return ServiceTypeNode, true
	case "network":
		return ServiceTypeNetwork, true
	case "metrics":
		return ServiceTypeMetrics, true
	case "query":
		return ServiceTypeQuery, true
	case "event":
		return ServiceTypeEvent, true
	case "submit":
		return ServiceTypeSubmit, true
	case "validate":
		return ServiceTypeValidate, true
	case "faucet":
		return ServiceTypeFaucet, true
	default:
		return 0, false
	}
}

// MarshalJSON marshals the Service Type to JSON as a string.
func (v ServiceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

// UnmarshalJSON unmarshals the Service Type from JSON as a string.
func (v *ServiceType) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	var ok bool
	*v, ok = ServiceTypeByName(s)
	if !ok || strings.ContainsRune(v.String(), ':') {
		return fmt.Errorf("invalid Service Type %q", s)
	}
	return nil
}
