// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package messaging

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
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

type AuthoritySignature struct {
	fieldsSet []bool
	Signature protocol.Signature         `json:"signature,omitempty" form:"signature" query:"signature" validate:"required"`
	Proof     *protocol.AnnotatedReceipt `json:"proof,omitempty" form:"proof" query:"proof" validate:"required"`
	extraData []byte
}

type LegacyMessage struct {
	fieldsSet   []bool
	Signatures  []protocol.Signature  `json:"signatures,omitempty" form:"signatures" query:"signatures" validate:"required"`
	Transaction *protocol.Transaction `json:"transaction,omitempty" form:"transaction" query:"transaction" validate:"required"`
	extraData   []byte
}

type SyntheticTransaction struct {
	fieldsSet   []bool
	Transaction *protocol.Transaction      `json:"transaction,omitempty" form:"transaction" query:"transaction" validate:"required"`
	Proof       *protocol.AnnotatedReceipt `json:"proof,omitempty" form:"proof" query:"proof" validate:"required"`
	extraData   []byte
}

type UserSignature struct {
	fieldsSet []bool
	Signature protocol.KeySignature `json:"signature,omitempty" form:"signature" query:"signature" validate:"required"`
	extraData []byte
}

type UserTransaction struct {
	fieldsSet   []bool
	Transaction *protocol.Transaction `json:"transaction,omitempty" form:"transaction" query:"transaction" validate:"required"`
	extraData   []byte
}

type ValidatorSignature struct {
	fieldsSet []bool
	Signature protocol.KeySignature `json:"signature,omitempty" form:"signature" query:"signature" validate:"required"`
	extraData []byte
}

func (*AuthoritySignature) Type() MessageType { return MessageTypeAuthoritySignature }

func (*LegacyMessage) Type() MessageType { return MessageTypeLegacy }

func (*SyntheticTransaction) Type() MessageType { return MessageTypeSyntheticTransaction }

func (*UserSignature) Type() MessageType { return MessageTypeUserSignature }

func (*UserTransaction) Type() MessageType { return MessageTypeUserTransaction }

func (*ValidatorSignature) Type() MessageType { return MessageTypeValidatorSignature }

func (v *AuthoritySignature) Copy() *AuthoritySignature {
	u := new(AuthoritySignature)

	if v.Signature != nil {
		u.Signature = protocol.CopySignature(v.Signature)
	}
	if v.Proof != nil {
		u.Proof = (v.Proof).Copy()
	}

	return u
}

func (v *AuthoritySignature) CopyAsInterface() interface{} { return v.Copy() }

func (v *LegacyMessage) Copy() *LegacyMessage {
	u := new(LegacyMessage)

	u.Signatures = make([]protocol.Signature, len(v.Signatures))
	for i, v := range v.Signatures {
		if v != nil {
			u.Signatures[i] = protocol.CopySignature(v)
		}
	}
	if v.Transaction != nil {
		u.Transaction = (v.Transaction).Copy()
	}

	return u
}

func (v *LegacyMessage) CopyAsInterface() interface{} { return v.Copy() }

func (v *SyntheticTransaction) Copy() *SyntheticTransaction {
	u := new(SyntheticTransaction)

	if v.Transaction != nil {
		u.Transaction = (v.Transaction).Copy()
	}
	if v.Proof != nil {
		u.Proof = (v.Proof).Copy()
	}

	return u
}

func (v *SyntheticTransaction) CopyAsInterface() interface{} { return v.Copy() }

func (v *UserSignature) Copy() *UserSignature {
	u := new(UserSignature)

	if v.Signature != nil {
		u.Signature = protocol.CopyKeySignature(v.Signature)
	}

	return u
}

func (v *UserSignature) CopyAsInterface() interface{} { return v.Copy() }

func (v *UserTransaction) Copy() *UserTransaction {
	u := new(UserTransaction)

	if v.Transaction != nil {
		u.Transaction = (v.Transaction).Copy()
	}

	return u
}

func (v *UserTransaction) CopyAsInterface() interface{} { return v.Copy() }

func (v *ValidatorSignature) Copy() *ValidatorSignature {
	u := new(ValidatorSignature)

	if v.Signature != nil {
		u.Signature = protocol.CopyKeySignature(v.Signature)
	}

	return u
}

func (v *ValidatorSignature) CopyAsInterface() interface{} { return v.Copy() }

func (v *AuthoritySignature) Equal(u *AuthoritySignature) bool {
	if !(protocol.EqualSignature(v.Signature, u.Signature)) {
		return false
	}
	switch {
	case v.Proof == u.Proof:
		// equal
	case v.Proof == nil || u.Proof == nil:
		return false
	case !((v.Proof).Equal(u.Proof)):
		return false
	}

	return true
}

func (v *LegacyMessage) Equal(u *LegacyMessage) bool {
	if len(v.Signatures) != len(u.Signatures) {
		return false
	}
	for i := range v.Signatures {
		if !(protocol.EqualSignature(v.Signatures[i], u.Signatures[i])) {
			return false
		}
	}
	switch {
	case v.Transaction == u.Transaction:
		// equal
	case v.Transaction == nil || u.Transaction == nil:
		return false
	case !((v.Transaction).Equal(u.Transaction)):
		return false
	}

	return true
}

func (v *SyntheticTransaction) Equal(u *SyntheticTransaction) bool {
	switch {
	case v.Transaction == u.Transaction:
		// equal
	case v.Transaction == nil || u.Transaction == nil:
		return false
	case !((v.Transaction).Equal(u.Transaction)):
		return false
	}
	switch {
	case v.Proof == u.Proof:
		// equal
	case v.Proof == nil || u.Proof == nil:
		return false
	case !((v.Proof).Equal(u.Proof)):
		return false
	}

	return true
}

func (v *UserSignature) Equal(u *UserSignature) bool {
	if !(protocol.EqualKeySignature(v.Signature, u.Signature)) {
		return false
	}

	return true
}

func (v *UserTransaction) Equal(u *UserTransaction) bool {
	switch {
	case v.Transaction == u.Transaction:
		// equal
	case v.Transaction == nil || u.Transaction == nil:
		return false
	case !((v.Transaction).Equal(u.Transaction)):
		return false
	}

	return true
}

func (v *ValidatorSignature) Equal(u *ValidatorSignature) bool {
	if !(protocol.EqualKeySignature(v.Signature, u.Signature)) {
		return false
	}

	return true
}

var fieldNames_AuthoritySignature = []string{
	1: "Type",
	2: "Signature",
	3: "Proof",
}

func (v *AuthoritySignature) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	writer.WriteEnum(1, v.Type())
	if !(protocol.EqualSignature(v.Signature, nil)) {
		writer.WriteValue(2, v.Signature.MarshalBinary)
	}
	if !(v.Proof == nil) {
		writer.WriteValue(3, v.Proof.MarshalBinary)
	}

	_, _, err := writer.Reset(fieldNames_AuthoritySignature)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *AuthoritySignature) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 0 && !v.fieldsSet[0] {
		errs = append(errs, "field Type is missing")
	}
	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Signature is missing")
	} else if protocol.EqualSignature(v.Signature, nil) {
		errs = append(errs, "field Signature is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field Proof is missing")
	} else if v.Proof == nil {
		errs = append(errs, "field Proof is not set")
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

var fieldNames_LegacyMessage = []string{
	1: "Type",
	2: "Signatures",
	3: "Transaction",
}

func (v *LegacyMessage) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	writer.WriteEnum(1, v.Type())
	if !(len(v.Signatures) == 0) {
		for _, v := range v.Signatures {
			writer.WriteValue(2, v.MarshalBinary)
		}
	}
	if !(v.Transaction == nil) {
		writer.WriteValue(3, v.Transaction.MarshalBinary)
	}

	_, _, err := writer.Reset(fieldNames_LegacyMessage)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *LegacyMessage) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 0 && !v.fieldsSet[0] {
		errs = append(errs, "field Type is missing")
	}
	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Signatures is missing")
	} else if len(v.Signatures) == 0 {
		errs = append(errs, "field Signatures is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field Transaction is missing")
	} else if v.Transaction == nil {
		errs = append(errs, "field Transaction is not set")
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

var fieldNames_SyntheticTransaction = []string{
	1: "Type",
	2: "Transaction",
	3: "Proof",
}

func (v *SyntheticTransaction) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	writer.WriteEnum(1, v.Type())
	if !(v.Transaction == nil) {
		writer.WriteValue(2, v.Transaction.MarshalBinary)
	}
	if !(v.Proof == nil) {
		writer.WriteValue(3, v.Proof.MarshalBinary)
	}

	_, _, err := writer.Reset(fieldNames_SyntheticTransaction)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *SyntheticTransaction) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 0 && !v.fieldsSet[0] {
		errs = append(errs, "field Type is missing")
	}
	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Transaction is missing")
	} else if v.Transaction == nil {
		errs = append(errs, "field Transaction is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field Proof is missing")
	} else if v.Proof == nil {
		errs = append(errs, "field Proof is not set")
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

var fieldNames_UserSignature = []string{
	1: "Type",
	2: "Signature",
}

func (v *UserSignature) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	writer.WriteEnum(1, v.Type())
	if !(protocol.EqualKeySignature(v.Signature, nil)) {
		writer.WriteValue(2, v.Signature.MarshalBinary)
	}

	_, _, err := writer.Reset(fieldNames_UserSignature)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *UserSignature) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 0 && !v.fieldsSet[0] {
		errs = append(errs, "field Type is missing")
	}
	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Signature is missing")
	} else if protocol.EqualKeySignature(v.Signature, nil) {
		errs = append(errs, "field Signature is not set")
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

var fieldNames_UserTransaction = []string{
	1: "Type",
	2: "Transaction",
}

func (v *UserTransaction) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	writer.WriteEnum(1, v.Type())
	if !(v.Transaction == nil) {
		writer.WriteValue(2, v.Transaction.MarshalBinary)
	}

	_, _, err := writer.Reset(fieldNames_UserTransaction)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *UserTransaction) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 0 && !v.fieldsSet[0] {
		errs = append(errs, "field Type is missing")
	}
	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Transaction is missing")
	} else if v.Transaction == nil {
		errs = append(errs, "field Transaction is not set")
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

var fieldNames_ValidatorSignature = []string{
	1: "Type",
	2: "Signature",
}

func (v *ValidatorSignature) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	writer.WriteEnum(1, v.Type())
	if !(protocol.EqualKeySignature(v.Signature, nil)) {
		writer.WriteValue(2, v.Signature.MarshalBinary)
	}

	_, _, err := writer.Reset(fieldNames_ValidatorSignature)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *ValidatorSignature) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 0 && !v.fieldsSet[0] {
		errs = append(errs, "field Type is missing")
	}
	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Signature is missing")
	} else if protocol.EqualKeySignature(v.Signature, nil) {
		errs = append(errs, "field Signature is not set")
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

func (v *AuthoritySignature) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *AuthoritySignature) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	var vType MessageType
	if x := new(MessageType); reader.ReadEnum(1, x) {
		vType = *x
	}
	if !(v.Type() == vType) {
		return fmt.Errorf("field Type: not equal: want %v, got %v", v.Type(), vType)
	}

	return v.UnmarshalFieldsFrom(reader)
}

func (v *AuthoritySignature) UnmarshalFieldsFrom(reader *encoding.Reader) error {
	reader.ReadValue(2, func(r io.Reader) error {
		x, err := protocol.UnmarshalSignatureFrom(r)
		if err == nil {
			v.Signature = x
		}
		return err
	})
	if x := new(protocol.AnnotatedReceipt); reader.ReadValue(3, x.UnmarshalBinaryFrom) {
		v.Proof = x
	}

	seen, err := reader.Reset(fieldNames_AuthoritySignature)
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

func (v *LegacyMessage) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *LegacyMessage) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	var vType MessageType
	if x := new(MessageType); reader.ReadEnum(1, x) {
		vType = *x
	}
	if !(v.Type() == vType) {
		return fmt.Errorf("field Type: not equal: want %v, got %v", v.Type(), vType)
	}

	return v.UnmarshalFieldsFrom(reader)
}

func (v *LegacyMessage) UnmarshalFieldsFrom(reader *encoding.Reader) error {
	for {
		ok := reader.ReadValue(2, func(r io.Reader) error {
			x, err := protocol.UnmarshalSignatureFrom(r)
			if err == nil {
				v.Signatures = append(v.Signatures, x)
			}
			return err
		})
		if !ok {
			break
		}
	}
	if x := new(protocol.Transaction); reader.ReadValue(3, x.UnmarshalBinaryFrom) {
		v.Transaction = x
	}

	seen, err := reader.Reset(fieldNames_LegacyMessage)
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

func (v *SyntheticTransaction) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *SyntheticTransaction) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	var vType MessageType
	if x := new(MessageType); reader.ReadEnum(1, x) {
		vType = *x
	}
	if !(v.Type() == vType) {
		return fmt.Errorf("field Type: not equal: want %v, got %v", v.Type(), vType)
	}

	return v.UnmarshalFieldsFrom(reader)
}

func (v *SyntheticTransaction) UnmarshalFieldsFrom(reader *encoding.Reader) error {
	if x := new(protocol.Transaction); reader.ReadValue(2, x.UnmarshalBinaryFrom) {
		v.Transaction = x
	}
	if x := new(protocol.AnnotatedReceipt); reader.ReadValue(3, x.UnmarshalBinaryFrom) {
		v.Proof = x
	}

	seen, err := reader.Reset(fieldNames_SyntheticTransaction)
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

func (v *UserSignature) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *UserSignature) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	var vType MessageType
	if x := new(MessageType); reader.ReadEnum(1, x) {
		vType = *x
	}
	if !(v.Type() == vType) {
		return fmt.Errorf("field Type: not equal: want %v, got %v", v.Type(), vType)
	}

	return v.UnmarshalFieldsFrom(reader)
}

func (v *UserSignature) UnmarshalFieldsFrom(reader *encoding.Reader) error {
	reader.ReadValue(2, func(r io.Reader) error {
		x, err := protocol.UnmarshalKeySignatureFrom(r)
		if err == nil {
			v.Signature = x
		}
		return err
	})

	seen, err := reader.Reset(fieldNames_UserSignature)
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

func (v *UserTransaction) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *UserTransaction) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	var vType MessageType
	if x := new(MessageType); reader.ReadEnum(1, x) {
		vType = *x
	}
	if !(v.Type() == vType) {
		return fmt.Errorf("field Type: not equal: want %v, got %v", v.Type(), vType)
	}

	return v.UnmarshalFieldsFrom(reader)
}

func (v *UserTransaction) UnmarshalFieldsFrom(reader *encoding.Reader) error {
	if x := new(protocol.Transaction); reader.ReadValue(2, x.UnmarshalBinaryFrom) {
		v.Transaction = x
	}

	seen, err := reader.Reset(fieldNames_UserTransaction)
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

func (v *ValidatorSignature) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *ValidatorSignature) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	var vType MessageType
	if x := new(MessageType); reader.ReadEnum(1, x) {
		vType = *x
	}
	if !(v.Type() == vType) {
		return fmt.Errorf("field Type: not equal: want %v, got %v", v.Type(), vType)
	}

	return v.UnmarshalFieldsFrom(reader)
}

func (v *ValidatorSignature) UnmarshalFieldsFrom(reader *encoding.Reader) error {
	reader.ReadValue(2, func(r io.Reader) error {
		x, err := protocol.UnmarshalKeySignatureFrom(r)
		if err == nil {
			v.Signature = x
		}
		return err
	})

	seen, err := reader.Reset(fieldNames_ValidatorSignature)
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

func (v *AuthoritySignature) MarshalJSON() ([]byte, error) {
	u := struct {
		Type      MessageType                                    `json:"type"`
		Signature encoding.JsonUnmarshalWith[protocol.Signature] `json:"signature,omitempty"`
		Proof     *protocol.AnnotatedReceipt                     `json:"proof,omitempty"`
	}{}
	u.Type = v.Type()
	u.Signature = encoding.JsonUnmarshalWith[protocol.Signature]{Value: v.Signature, Func: protocol.UnmarshalSignatureJSON}
	u.Proof = v.Proof
	return json.Marshal(&u)
}

func (v *LegacyMessage) MarshalJSON() ([]byte, error) {
	u := struct {
		Type        MessageType                                        `json:"type"`
		Signatures  encoding.JsonUnmarshalListWith[protocol.Signature] `json:"signatures,omitempty"`
		Transaction *protocol.Transaction                              `json:"transaction,omitempty"`
	}{}
	u.Type = v.Type()
	u.Signatures = encoding.JsonUnmarshalListWith[protocol.Signature]{Value: v.Signatures, Func: protocol.UnmarshalSignatureJSON}
	u.Transaction = v.Transaction
	return json.Marshal(&u)
}

func (v *SyntheticTransaction) MarshalJSON() ([]byte, error) {
	u := struct {
		Type        MessageType                `json:"type"`
		Transaction *protocol.Transaction      `json:"transaction,omitempty"`
		Proof       *protocol.AnnotatedReceipt `json:"proof,omitempty"`
	}{}
	u.Type = v.Type()
	u.Transaction = v.Transaction
	u.Proof = v.Proof
	return json.Marshal(&u)
}

func (v *UserSignature) MarshalJSON() ([]byte, error) {
	u := struct {
		Type      MessageType                                       `json:"type"`
		Signature encoding.JsonUnmarshalWith[protocol.KeySignature] `json:"signature,omitempty"`
	}{}
	u.Type = v.Type()
	u.Signature = encoding.JsonUnmarshalWith[protocol.KeySignature]{Value: v.Signature, Func: protocol.UnmarshalKeySignatureJSON}
	return json.Marshal(&u)
}

func (v *UserTransaction) MarshalJSON() ([]byte, error) {
	u := struct {
		Type        MessageType           `json:"type"`
		Transaction *protocol.Transaction `json:"transaction,omitempty"`
	}{}
	u.Type = v.Type()
	u.Transaction = v.Transaction
	return json.Marshal(&u)
}

func (v *ValidatorSignature) MarshalJSON() ([]byte, error) {
	u := struct {
		Type      MessageType                                       `json:"type"`
		Signature encoding.JsonUnmarshalWith[protocol.KeySignature] `json:"signature,omitempty"`
	}{}
	u.Type = v.Type()
	u.Signature = encoding.JsonUnmarshalWith[protocol.KeySignature]{Value: v.Signature, Func: protocol.UnmarshalKeySignatureJSON}
	return json.Marshal(&u)
}

func (v *AuthoritySignature) UnmarshalJSON(data []byte) error {
	u := struct {
		Type      MessageType                                    `json:"type"`
		Signature encoding.JsonUnmarshalWith[protocol.Signature] `json:"signature,omitempty"`
		Proof     *protocol.AnnotatedReceipt                     `json:"proof,omitempty"`
	}{}
	u.Type = v.Type()
	u.Signature = encoding.JsonUnmarshalWith[protocol.Signature]{Value: v.Signature, Func: protocol.UnmarshalSignatureJSON}
	u.Proof = v.Proof
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	if !(v.Type() == u.Type) {
		return fmt.Errorf("field Type: not equal: want %v, got %v", v.Type(), u.Type)
	}
	v.Signature = u.Signature.Value

	v.Proof = u.Proof
	return nil
}

func (v *LegacyMessage) UnmarshalJSON(data []byte) error {
	u := struct {
		Type        MessageType                                        `json:"type"`
		Signatures  encoding.JsonUnmarshalListWith[protocol.Signature] `json:"signatures,omitempty"`
		Transaction *protocol.Transaction                              `json:"transaction,omitempty"`
	}{}
	u.Type = v.Type()
	u.Signatures = encoding.JsonUnmarshalListWith[protocol.Signature]{Value: v.Signatures, Func: protocol.UnmarshalSignatureJSON}
	u.Transaction = v.Transaction
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	if !(v.Type() == u.Type) {
		return fmt.Errorf("field Type: not equal: want %v, got %v", v.Type(), u.Type)
	}
	v.Signatures = make([]protocol.Signature, len(u.Signatures.Value))
	for i, x := range u.Signatures.Value {
		v.Signatures[i] = x
	}
	v.Transaction = u.Transaction
	return nil
}

func (v *SyntheticTransaction) UnmarshalJSON(data []byte) error {
	u := struct {
		Type        MessageType                `json:"type"`
		Transaction *protocol.Transaction      `json:"transaction,omitempty"`
		Proof       *protocol.AnnotatedReceipt `json:"proof,omitempty"`
	}{}
	u.Type = v.Type()
	u.Transaction = v.Transaction
	u.Proof = v.Proof
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	if !(v.Type() == u.Type) {
		return fmt.Errorf("field Type: not equal: want %v, got %v", v.Type(), u.Type)
	}
	v.Transaction = u.Transaction
	v.Proof = u.Proof
	return nil
}

func (v *UserSignature) UnmarshalJSON(data []byte) error {
	u := struct {
		Type      MessageType                                       `json:"type"`
		Signature encoding.JsonUnmarshalWith[protocol.KeySignature] `json:"signature,omitempty"`
	}{}
	u.Type = v.Type()
	u.Signature = encoding.JsonUnmarshalWith[protocol.KeySignature]{Value: v.Signature, Func: protocol.UnmarshalKeySignatureJSON}
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	if !(v.Type() == u.Type) {
		return fmt.Errorf("field Type: not equal: want %v, got %v", v.Type(), u.Type)
	}
	v.Signature = u.Signature.Value

	return nil
}

func (v *UserTransaction) UnmarshalJSON(data []byte) error {
	u := struct {
		Type        MessageType           `json:"type"`
		Transaction *protocol.Transaction `json:"transaction,omitempty"`
	}{}
	u.Type = v.Type()
	u.Transaction = v.Transaction
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	if !(v.Type() == u.Type) {
		return fmt.Errorf("field Type: not equal: want %v, got %v", v.Type(), u.Type)
	}
	v.Transaction = u.Transaction
	return nil
}

func (v *ValidatorSignature) UnmarshalJSON(data []byte) error {
	u := struct {
		Type      MessageType                                       `json:"type"`
		Signature encoding.JsonUnmarshalWith[protocol.KeySignature] `json:"signature,omitempty"`
	}{}
	u.Type = v.Type()
	u.Signature = encoding.JsonUnmarshalWith[protocol.KeySignature]{Value: v.Signature, Func: protocol.UnmarshalKeySignatureJSON}
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	if !(v.Type() == u.Type) {
		return fmt.Errorf("field Type: not equal: want %v, got %v", v.Type(), u.Type)
	}
	v.Signature = u.Signature.Value

	return nil
}
