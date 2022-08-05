package core

// GENERATED BY go run ./tools/cmd/gen-types. DO NOT EDIT.

//lint:file-ignore S1001,S1002,S1008,SA4013 generated code

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"gitlab.com/accumulatenetwork/accumulate/internal/encoding"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

type GlobalValues struct {
	fieldsSet   []bool
	Oracle      *protocol.AcmeOracle         `json:"oracle,omitempty" form:"oracle" query:"oracle" validate:"required"`
	Globals     *protocol.NetworkGlobals     `json:"globals,omitempty" form:"globals" query:"globals" validate:"required"`
	Network     *protocol.NetworkDefinition  `json:"network,omitempty" form:"network" query:"network" validate:"required"`
	Routing     *protocol.RoutingTable       `json:"routing,omitempty" form:"routing" query:"routing" validate:"required"`
	AddressBook *protocol.AddressBookEntries `json:"addressBook,omitempty" form:"addressBook" query:"addressBook" validate:"required"`
	extraData   []byte
}

func (v *GlobalValues) Copy() *GlobalValues {
	u := new(GlobalValues)

	if v.Oracle != nil {
		u.Oracle = (v.Oracle).Copy()
	}
	if v.Globals != nil {
		u.Globals = (v.Globals).Copy()
	}
	if v.Network != nil {
		u.Network = (v.Network).Copy()
	}
	if v.Routing != nil {
		u.Routing = (v.Routing).Copy()
	}
	if v.AddressBook != nil {
		u.AddressBook = (v.AddressBook).Copy()
	}

	return u
}

func (v *GlobalValues) CopyAsInterface() interface{} { return v.Copy() }

func (v *GlobalValues) Equal(u *GlobalValues) bool {
	switch {
	case v.Oracle == u.Oracle:
		// equal
	case v.Oracle == nil || u.Oracle == nil:
		return false
	case !((v.Oracle).Equal(u.Oracle)):
		return false
	}
	switch {
	case v.Globals == u.Globals:
		// equal
	case v.Globals == nil || u.Globals == nil:
		return false
	case !((v.Globals).Equal(u.Globals)):
		return false
	}
	switch {
	case v.Network == u.Network:
		// equal
	case v.Network == nil || u.Network == nil:
		return false
	case !((v.Network).Equal(u.Network)):
		return false
	}
	switch {
	case v.Routing == u.Routing:
		// equal
	case v.Routing == nil || u.Routing == nil:
		return false
	case !((v.Routing).Equal(u.Routing)):
		return false
	}
	switch {
	case v.AddressBook == u.AddressBook:
		// equal
	case v.AddressBook == nil || u.AddressBook == nil:
		return false
	case !((v.AddressBook).Equal(u.AddressBook)):
		return false
	}

	return true
}

var fieldNames_GlobalValues = []string{
	1: "Oracle",
	2: "Globals",
	3: "Network",
	4: "Routing",
	5: "AddressBook",
}

func (v *GlobalValues) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(v.Oracle == nil) {
		writer.WriteValue(1, v.Oracle.MarshalBinary)
	}
	if !(v.Globals == nil) {
		writer.WriteValue(2, v.Globals.MarshalBinary)
	}
	if !(v.Network == nil) {
		writer.WriteValue(3, v.Network.MarshalBinary)
	}
	if !(v.Routing == nil) {
		writer.WriteValue(4, v.Routing.MarshalBinary)
	}
	if !(v.AddressBook == nil) {
		writer.WriteValue(5, v.AddressBook.MarshalBinary)
	}

	_, _, err := writer.Reset(fieldNames_GlobalValues)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *GlobalValues) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Oracle is missing")
	} else if v.Oracle == nil {
		errs = append(errs, "field Oracle is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field Globals is missing")
	} else if v.Globals == nil {
		errs = append(errs, "field Globals is not set")
	}
	if len(v.fieldsSet) > 3 && !v.fieldsSet[3] {
		errs = append(errs, "field Network is missing")
	} else if v.Network == nil {
		errs = append(errs, "field Network is not set")
	}
	if len(v.fieldsSet) > 4 && !v.fieldsSet[4] {
		errs = append(errs, "field Routing is missing")
	} else if v.Routing == nil {
		errs = append(errs, "field Routing is not set")
	}
	if len(v.fieldsSet) > 5 && !v.fieldsSet[5] {
		errs = append(errs, "field AddressBook is missing")
	} else if v.AddressBook == nil {
		errs = append(errs, "field AddressBook is not set")
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

func (v *GlobalValues) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *GlobalValues) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x := new(protocol.AcmeOracle); reader.ReadValue(1, x.UnmarshalBinary) {
		v.Oracle = x
	}
	if x := new(protocol.NetworkGlobals); reader.ReadValue(2, x.UnmarshalBinary) {
		v.Globals = x
	}
	if x := new(protocol.NetworkDefinition); reader.ReadValue(3, x.UnmarshalBinary) {
		v.Network = x
	}
	if x := new(protocol.RoutingTable); reader.ReadValue(4, x.UnmarshalBinary) {
		v.Routing = x
	}
	if x := new(protocol.AddressBookEntries); reader.ReadValue(5, x.UnmarshalBinary) {
		v.AddressBook = x
	}

	seen, err := reader.Reset(fieldNames_GlobalValues)
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
