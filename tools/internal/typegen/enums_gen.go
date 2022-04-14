package typegen

// GENERATED BY go run ./tools/cmd/gen-enum. DO NOT EDIT.

import (
	"encoding/json"
	"fmt"
	"strings"
)

// MarshalAsBasic marshals the field as a basic type.
const MarshalAsBasic MarshalAs = 0

// MarshalAsNone omits the field from marshalling.
const MarshalAsNone MarshalAs = 1

// MarshalAsEnum marshals the field as an enumeration.
const MarshalAsEnum MarshalAs = 2

// MarshalAsValue marshals the field as a value type.
const MarshalAsValue MarshalAs = 3

// MarshalAsReference marshals the field as a reference type.
const MarshalAsReference MarshalAs = 4

// GetEnumValue returns the value of the Marshal As
func (v MarshalAs) GetEnumValue() uint64 { return uint64(v) }

// SetEnumValue sets the value. SetEnumValue returns false if the value is invalid.
func (v *MarshalAs) SetEnumValue(id uint64) bool {
	u := MarshalAs(id)
	switch u {
	case MarshalAsBasic, MarshalAsNone, MarshalAsEnum, MarshalAsValue, MarshalAsReference:
		*v = u
		return true
	default:
		return false
	}
}

// String returns the name of the Marshal As
func (v MarshalAs) String() string {
	switch v {
	case MarshalAsBasic:
		return "basic"
	case MarshalAsNone:
		return "none"
	case MarshalAsEnum:
		return "enum"
	case MarshalAsValue:
		return "value"
	case MarshalAsReference:
		return "reference"
	default:
		return fmt.Sprintf("MarshalAs:%d", v)
	}
}

// MarshalAsByName returns the named Marshal As.
func MarshalAsByName(name string) (MarshalAs, bool) {
	switch strings.ToLower(name) {
	case "basic":
		return MarshalAsBasic, true
	case "":
		return MarshalAsBasic, true
	case "none":
		return MarshalAsNone, true
	case "enum":
		return MarshalAsEnum, true
	case "value":
		return MarshalAsValue, true
	case "reference":
		return MarshalAsReference, true
	default:
		return 0, false
	}
}

// MarshalJSON marshals the Marshal As to JSON as a string.
func (v MarshalAs) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

// UnmarshalJSON unmarshals the Marshal As from JSON as a string.
func (v *MarshalAs) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}

	var ok bool
	*v, ok = MarshalAsByName(s)
	if !ok || strings.ContainsRune(v.String(), ':') {
		return fmt.Errorf("invalid Marshal As %q", s)
	}
	return nil
}
