// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package errors

// GENERATED BY go run ./tools/cmd/gen-types. DO NOT EDIT.

//lint:file-ignore S1001,S1002,S1008,SA4013 generated code

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"gitlab.com/accumulatenetwork/accumulate/pkg/types/encoding"
)

type CallSite struct {
	fieldsSet []bool
	FuncName  string `json:"funcName,omitempty" form:"funcName" query:"funcName" validate:"required"`
	File      string `json:"file,omitempty" form:"file" query:"file" validate:"required"`
	Line      int64  `json:"line,omitempty" form:"line" query:"line" validate:"required"`
	extraData []byte
}

type Error struct {
	fieldsSet []bool
	Message   string      `json:"message,omitempty" form:"message" query:"message" validate:"required"`
	Code      Status      `json:"code,omitempty" form:"code" query:"code" validate:"required"`
	Cause     *Error      `json:"cause,omitempty" form:"cause" query:"cause" validate:"required"`
	CallStack []*CallSite `json:"callStack,omitempty" form:"callStack" query:"callStack" validate:"required"`
	extraData []byte
}

func (v *CallSite) Copy() *CallSite {
	u := new(CallSite)

	u.FuncName = v.FuncName
	u.File = v.File
	u.Line = v.Line

	return u
}

func (v *CallSite) CopyAsInterface() interface{} { return v.Copy() }

func (v *Error) Copy() *Error {
	u := new(Error)

	u.Message = v.Message
	u.Code = v.Code
	if v.Cause != nil {
		u.Cause = (v.Cause).Copy()
	}
	u.CallStack = make([]*CallSite, len(v.CallStack))
	for i, v := range v.CallStack {
		if v != nil {
			u.CallStack[i] = (v).Copy()
		}
	}

	return u
}

func (v *Error) CopyAsInterface() interface{} { return v.Copy() }

func (v *CallSite) Equal(u *CallSite) bool {
	if !(v.FuncName == u.FuncName) {
		return false
	}
	if !(v.File == u.File) {
		return false
	}
	if !(v.Line == u.Line) {
		return false
	}

	return true
}

func (v *Error) Equal(u *Error) bool {
	if !(v.Message == u.Message) {
		return false
	}
	if !(v.Code == u.Code) {
		return false
	}
	switch {
	case v.Cause == u.Cause:
		// equal
	case v.Cause == nil || u.Cause == nil:
		return false
	case !((v.Cause).Equal(u.Cause)):
		return false
	}
	if len(v.CallStack) != len(u.CallStack) {
		return false
	}
	for i := range v.CallStack {
		if !((v.CallStack[i]).Equal(u.CallStack[i])) {
			return false
		}
	}

	return true
}

var fieldNames_CallSite = []string{
	1: "FuncName",
	2: "File",
	3: "Line",
}

func (v *CallSite) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(len(v.FuncName) == 0) {
		writer.WriteString(1, v.FuncName)
	}
	if !(len(v.File) == 0) {
		writer.WriteString(2, v.File)
	}
	if !(v.Line == 0) {
		writer.WriteInt(3, v.Line)
	}

	_, _, err := writer.Reset(fieldNames_CallSite)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *CallSite) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 0 && !v.fieldsSet[0] {
		errs = append(errs, "field FuncName is missing")
	} else if len(v.FuncName) == 0 {
		errs = append(errs, "field FuncName is not set")
	}
	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field File is missing")
	} else if len(v.File) == 0 {
		errs = append(errs, "field File is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field Line is missing")
	} else if v.Line == 0 {
		errs = append(errs, "field Line is not set")
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

var fieldNames_Error = []string{
	1: "Message",
	2: "Code",
	3: "Cause",
	4: "CallStack",
}

func (v *Error) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(len(v.Message) == 0) {
		writer.WriteString(1, v.Message)
	}
	if !(v.Code == 0) {
		writer.WriteEnum(2, v.Code)
	}
	if !(v.Cause == nil) {
		writer.WriteValue(3, v.Cause.MarshalBinary)
	}
	if !(len(v.CallStack) == 0) {
		for _, v := range v.CallStack {
			writer.WriteValue(4, v.MarshalBinary)
		}
	}

	_, _, err := writer.Reset(fieldNames_Error)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *Error) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 0 && !v.fieldsSet[0] {
		errs = append(errs, "field Message is missing")
	} else if len(v.Message) == 0 {
		errs = append(errs, "field Message is not set")
	}
	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Code is missing")
	} else if v.Code == 0 {
		errs = append(errs, "field Code is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field Cause is missing")
	} else if v.Cause == nil {
		errs = append(errs, "field Cause is not set")
	}
	if len(v.fieldsSet) > 3 && !v.fieldsSet[3] {
		errs = append(errs, "field CallStack is missing")
	} else if len(v.CallStack) == 0 {
		errs = append(errs, "field CallStack is not set")
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

func (v *CallSite) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *CallSite) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x, ok := reader.ReadString(1); ok {
		v.FuncName = x
	}
	if x, ok := reader.ReadString(2); ok {
		v.File = x
	}
	if x, ok := reader.ReadInt(3); ok {
		v.Line = x
	}

	seen, err := reader.Reset(fieldNames_CallSite)
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

func (v *Error) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *Error) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x, ok := reader.ReadString(1); ok {
		v.Message = x
	}
	if x := new(Status); reader.ReadEnum(2, x) {
		v.Code = *x
	}
	if x := new(Error); reader.ReadValue(3, x.UnmarshalBinary) {
		v.Cause = x
	}
	for {
		if x := new(CallSite); reader.ReadValue(4, x.UnmarshalBinary) {
			v.CallStack = append(v.CallStack, x)
		} else {
			break
		}
	}

	seen, err := reader.Reset(fieldNames_Error)
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

func (v *Error) MarshalJSON() ([]byte, error) {
	u := struct {
		Message   string                       `json:"message,omitempty"`
		Code      Status                       `json:"code,omitempty"`
		CodeID    uint64                       `json:"codeID,omitempty"`
		Cause     *Error                       `json:"cause,omitempty"`
		CallStack encoding.JsonList[*CallSite] `json:"callStack,omitempty"`
	}{}
	u.Message = v.Message
	u.Code = v.Code
	u.CodeID = v.CodeID()
	u.Cause = v.Cause
	u.CallStack = v.CallStack
	return json.Marshal(&u)
}

func (v *Error) UnmarshalJSON(data []byte) error {
	u := struct {
		Message   string                       `json:"message,omitempty"`
		Code      Status                       `json:"code,omitempty"`
		CodeID    uint64                       `json:"codeID,omitempty"`
		Cause     *Error                       `json:"cause,omitempty"`
		CallStack encoding.JsonList[*CallSite] `json:"callStack,omitempty"`
	}{}
	u.Message = v.Message
	u.Code = v.Code
	u.CodeID = v.CodeID()
	u.Cause = v.Cause
	u.CallStack = v.CallStack
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Message = u.Message
	v.Code = u.Code
	v.Cause = u.Cause
	v.CallStack = u.CallStack
	return nil
}
