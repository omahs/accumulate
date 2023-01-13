// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package p2p

// GENERATED BY go run ./tools/cmd/gen-types. DO NOT EDIT.

//lint:file-ignore S1001,S1002,S1008,SA4013 generated code

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"gitlab.com/accumulatenetwork/accumulate/pkg/api/v3"
	"gitlab.com/accumulatenetwork/accumulate/pkg/types/encoding"
	"gitlab.com/accumulatenetwork/accumulate/pkg/types/p2p"
)

type AddrInfo struct {
	fieldsSet []bool
	Info
	Addrs     []p2p.Multiaddr `json:"addrs,omitempty" form:"addrs" query:"addrs" validate:"required"`
	extraData []byte
}

type Info struct {
	fieldsSet []bool
	ID        p2p.PeerID     `json:"id,omitempty" form:"id" query:"id" validate:"required"`
	Services  []*ServiceInfo `json:"services,omitempty" form:"services" query:"services" validate:"required"`
	extraData []byte
}

type ServiceInfo struct {
	fieldsSet []bool
	Address   *api.ServiceAddress `json:"address,omitempty" form:"address" query:"address" validate:"required"`
	extraData []byte
}

type Whoami struct {
	fieldsSet []bool
	Self      *Info       `json:"self,omitempty" form:"self" query:"self" validate:"required"`
	Known     []*AddrInfo `json:"known,omitempty" form:"known" query:"known" validate:"required"`
	extraData []byte
}

func (v *AddrInfo) Copy() *AddrInfo {
	u := new(AddrInfo)

	u.Info = *v.Info.Copy()
	u.Addrs = make([]p2p.Multiaddr, len(v.Addrs))
	for i, v := range v.Addrs {
		if v != nil {
			u.Addrs[i] = p2p.CopyMultiaddr(v)
		}
	}

	return u
}

func (v *AddrInfo) CopyAsInterface() interface{} { return v.Copy() }

func (v *Info) Copy() *Info {
	u := new(Info)

	if v.ID != "" {
		u.ID = p2p.CopyPeerID(v.ID)
	}
	u.Services = make([]*ServiceInfo, len(v.Services))
	for i, v := range v.Services {
		if v != nil {
			u.Services[i] = (v).Copy()
		}
	}

	return u
}

func (v *Info) CopyAsInterface() interface{} { return v.Copy() }

func (v *ServiceInfo) Copy() *ServiceInfo {
	u := new(ServiceInfo)

	if v.Address != nil {
		u.Address = (v.Address).Copy()
	}

	return u
}

func (v *ServiceInfo) CopyAsInterface() interface{} { return v.Copy() }

func (v *Whoami) Copy() *Whoami {
	u := new(Whoami)

	if v.Self != nil {
		u.Self = (v.Self).Copy()
	}
	u.Known = make([]*AddrInfo, len(v.Known))
	for i, v := range v.Known {
		if v != nil {
			u.Known[i] = (v).Copy()
		}
	}

	return u
}

func (v *Whoami) CopyAsInterface() interface{} { return v.Copy() }

func (v *AddrInfo) Equal(u *AddrInfo) bool {
	if !v.Info.Equal(&u.Info) {
		return false
	}
	if len(v.Addrs) != len(u.Addrs) {
		return false
	}
	for i := range v.Addrs {
		if !(p2p.EqualMultiaddr(v.Addrs[i], u.Addrs[i])) {
			return false
		}
	}

	return true
}

func (v *Info) Equal(u *Info) bool {
	if !(p2p.EqualPeerID(v.ID, u.ID)) {
		return false
	}
	if len(v.Services) != len(u.Services) {
		return false
	}
	for i := range v.Services {
		if !((v.Services[i]).Equal(u.Services[i])) {
			return false
		}
	}

	return true
}

func (v *ServiceInfo) Equal(u *ServiceInfo) bool {
	switch {
	case v.Address == u.Address:
		// equal
	case v.Address == nil || u.Address == nil:
		return false
	case !((v.Address).Equal(u.Address)):
		return false
	}

	return true
}

func (v *Whoami) Equal(u *Whoami) bool {
	switch {
	case v.Self == u.Self:
		// equal
	case v.Self == nil || u.Self == nil:
		return false
	case !((v.Self).Equal(u.Self)):
		return false
	}
	if len(v.Known) != len(u.Known) {
		return false
	}
	for i := range v.Known {
		if !((v.Known[i]).Equal(u.Known[i])) {
			return false
		}
	}

	return true
}

var fieldNames_AddrInfo = []string{
	1: "Info",
	2: "Addrs",
}

func (v *AddrInfo) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	writer.WriteValue(1, v.Info.MarshalBinary)
	if !(len(v.Addrs) == 0) {
		for _, v := range v.Addrs {
			writer.WriteValue(2, v.MarshalBinary)
		}
	}

	_, _, err := writer.Reset(fieldNames_AddrInfo)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *AddrInfo) IsValid() error {
	var errs []string

	if err := v.Info.IsValid(); err != nil {
		errs = append(errs, err.Error())
	}
	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Addrs is missing")
	} else if len(v.Addrs) == 0 {
		errs = append(errs, "field Addrs is not set")
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

var fieldNames_Info = []string{
	1: "ID",
	2: "Services",
}

func (v *Info) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(v.ID == ("")) {
		writer.WriteValue(1, v.ID.MarshalBinary)
	}
	if !(len(v.Services) == 0) {
		for _, v := range v.Services {
			writer.WriteValue(2, v.MarshalBinary)
		}
	}

	_, _, err := writer.Reset(fieldNames_Info)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *Info) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 0 && !v.fieldsSet[0] {
		errs = append(errs, "field ID is missing")
	} else if v.ID == ("") {
		errs = append(errs, "field ID is not set")
	}
	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Services is missing")
	} else if len(v.Services) == 0 {
		errs = append(errs, "field Services is not set")
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

var fieldNames_ServiceInfo = []string{
	1: "Address",
}

func (v *ServiceInfo) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(v.Address == nil) {
		writer.WriteValue(1, v.Address.MarshalBinary)
	}

	_, _, err := writer.Reset(fieldNames_ServiceInfo)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *ServiceInfo) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 0 && !v.fieldsSet[0] {
		errs = append(errs, "field Address is missing")
	} else if v.Address == nil {
		errs = append(errs, "field Address is not set")
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

var fieldNames_Whoami = []string{
	1: "Self",
	2: "Known",
}

func (v *Whoami) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(v.Self == nil) {
		writer.WriteValue(1, v.Self.MarshalBinary)
	}
	if !(len(v.Known) == 0) {
		for _, v := range v.Known {
			writer.WriteValue(2, v.MarshalBinary)
		}
	}

	_, _, err := writer.Reset(fieldNames_Whoami)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *Whoami) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 0 && !v.fieldsSet[0] {
		errs = append(errs, "field Self is missing")
	} else if v.Self == nil {
		errs = append(errs, "field Self is not set")
	}
	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Known is missing")
	} else if len(v.Known) == 0 {
		errs = append(errs, "field Known is not set")
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

func (v *AddrInfo) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *AddrInfo) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	reader.ReadValue(1, v.Info.UnmarshalBinaryFrom)
	for {
		ok := reader.ReadValue(2, func(r io.Reader) error {
			x, err := p2p.UnmarshalMultiaddrFrom(r)
			if err == nil {
				v.Addrs = append(v.Addrs, x)
			}
			return err
		})
		if !ok {
			break
		}
	}

	seen, err := reader.Reset(fieldNames_AddrInfo)
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

func (v *Info) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *Info) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	reader.ReadValue(1, func(r io.Reader) error {
		x, err := p2p.UnmarshalPeerIDFrom(r)
		if err == nil {
			v.ID = x
		}
		return err
	})
	for {
		if x := new(ServiceInfo); reader.ReadValue(2, x.UnmarshalBinaryFrom) {
			v.Services = append(v.Services, x)
		} else {
			break
		}
	}

	seen, err := reader.Reset(fieldNames_Info)
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

func (v *ServiceInfo) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *ServiceInfo) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x := new(api.ServiceAddress); reader.ReadValue(1, x.UnmarshalBinaryFrom) {
		v.Address = x
	}

	seen, err := reader.Reset(fieldNames_ServiceInfo)
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

func (v *Whoami) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *Whoami) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x := new(Info); reader.ReadValue(1, x.UnmarshalBinaryFrom) {
		v.Self = x
	}
	for {
		if x := new(AddrInfo); reader.ReadValue(2, x.UnmarshalBinaryFrom) {
			v.Known = append(v.Known, x)
		} else {
			break
		}
	}

	seen, err := reader.Reset(fieldNames_Whoami)
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

func (v *AddrInfo) MarshalJSON() ([]byte, error) {
	u := struct {
		ID       *encoding.JsonUnmarshalWith[p2p.PeerID]        `json:"id,omitempty"`
		Services encoding.JsonList[*ServiceInfo]                `json:"services,omitempty"`
		Addrs    *encoding.JsonUnmarshalListWith[p2p.Multiaddr] `json:"addrs,omitempty"`
	}{}
	if !(v.Info.ID == ("")) {

		u.ID = &encoding.JsonUnmarshalWith[p2p.PeerID]{Value: v.Info.ID, Func: p2p.UnmarshalPeerIDJSON}
	}
	if !(len(v.Info.Services) == 0) {

		u.Services = v.Info.Services
	}
	if !(len(v.Addrs) == 0) {
		u.Addrs = &encoding.JsonUnmarshalListWith[p2p.Multiaddr]{Value: v.Addrs, Func: p2p.UnmarshalMultiaddrJSON}
	}
	return json.Marshal(&u)
}

func (v *Info) MarshalJSON() ([]byte, error) {
	u := struct {
		ID       *encoding.JsonUnmarshalWith[p2p.PeerID] `json:"id,omitempty"`
		Services encoding.JsonList[*ServiceInfo]         `json:"services,omitempty"`
	}{}
	if !(v.ID == ("")) {
		u.ID = &encoding.JsonUnmarshalWith[p2p.PeerID]{Value: v.ID, Func: p2p.UnmarshalPeerIDJSON}
	}
	if !(len(v.Services) == 0) {
		u.Services = v.Services
	}
	return json.Marshal(&u)
}

func (v *Whoami) MarshalJSON() ([]byte, error) {
	u := struct {
		Self  *Info                        `json:"self,omitempty"`
		Known encoding.JsonList[*AddrInfo] `json:"known,omitempty"`
	}{}
	if !(v.Self == nil) {
		u.Self = v.Self
	}
	if !(len(v.Known) == 0) {
		u.Known = v.Known
	}
	return json.Marshal(&u)
}

func (v *AddrInfo) UnmarshalJSON(data []byte) error {
	u := struct {
		ID       *encoding.JsonUnmarshalWith[p2p.PeerID]        `json:"id,omitempty"`
		Services encoding.JsonList[*ServiceInfo]                `json:"services,omitempty"`
		Addrs    *encoding.JsonUnmarshalListWith[p2p.Multiaddr] `json:"addrs,omitempty"`
	}{}
	u.ID = &encoding.JsonUnmarshalWith[p2p.PeerID]{Value: v.Info.ID, Func: p2p.UnmarshalPeerIDJSON}
	u.Services = v.Info.Services
	u.Addrs = &encoding.JsonUnmarshalListWith[p2p.Multiaddr]{Value: v.Addrs, Func: p2p.UnmarshalMultiaddrJSON}
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Info.ID = u.ID.Value

	v.Info.Services = u.Services
	v.Addrs = make([]p2p.Multiaddr, len(u.Addrs.Value))
	for i, x := range u.Addrs.Value {
		v.Addrs[i] = x
	}
	return nil
}

func (v *Info) UnmarshalJSON(data []byte) error {
	u := struct {
		ID       *encoding.JsonUnmarshalWith[p2p.PeerID] `json:"id,omitempty"`
		Services encoding.JsonList[*ServiceInfo]         `json:"services,omitempty"`
	}{}
	u.ID = &encoding.JsonUnmarshalWith[p2p.PeerID]{Value: v.ID, Func: p2p.UnmarshalPeerIDJSON}
	u.Services = v.Services
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.ID = u.ID.Value

	v.Services = u.Services
	return nil
}

func (v *Whoami) UnmarshalJSON(data []byte) error {
	u := struct {
		Self  *Info                        `json:"self,omitempty"`
		Known encoding.JsonList[*AddrInfo] `json:"known,omitempty"`
	}{}
	u.Self = v.Self
	u.Known = v.Known
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Self = u.Self
	v.Known = u.Known
	return nil
}
