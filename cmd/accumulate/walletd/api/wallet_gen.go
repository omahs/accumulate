package api

// GENERATED BY go run ./tools/cmd/gen-types. DO NOT EDIT.

//lint:file-ignore S1001,S1002,S1008,SA4013 generated code

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"gitlab.com/accumulatenetwork/accumulate/internal/encoding"
	"gitlab.com/accumulatenetwork/accumulate/pkg/url"
	"gitlab.com/accumulatenetwork/accumulate/protocol"
)

type Adi struct {
	fieldsSet []bool
	Url       url.URL `json:"url,omitempty" form:"url" query:"url" validate:"required"`
	Pages     []Page  `json:"pages,omitempty" form:"pages" query:"pages"`
	extraData []byte
}

type DerivationCount struct {
	Type  protocol.SignatureType `json:"type,omitempty" form:"type" query:"type" validate:"required"`
	Count uint64                 `json:"count,omitempty" form:"count" query:"count" validate:"required"`
}

type Key struct {
	fieldsSet  []bool
	PrivateKey []byte  `json:"privateKey,omitempty" form:"privateKey" query:"privateKey" validate:"required"`
	PublicKey  []byte  `json:"publicKey,omitempty" form:"publicKey" query:"publicKey" validate:"required"`
	KeyInfo    KeyInfo `json:"keyInfo,omitempty" form:"keyInfo" query:"keyInfo" validate:"required"`
	extraData  []byte
}

type KeyInfo struct {
	fieldsSet  []bool
	Type       protocol.SignatureType `json:"type,omitempty" form:"type" query:"type" validate:"required"`
	Derivation string                 `json:"derivation,omitempty" form:"derivation" query:"derivation"`
	WalletID   *url.URL               `json:"walletID,omitempty" form:"walletID" query:"walletID"`
	extraData  []byte
}

type KeyName struct {
	fieldsSet []bool
	Name      string `json:"name,omitempty" form:"name" query:"name" validate:"required"`
	PublicKey []byte `json:"publicKey,omitempty" form:"publicKey" query:"publicKey" validate:"required"`
	extraData []byte
}

type LiteLabel struct {
	LiteName string `json:"liteName,omitempty" form:"liteName" query:"liteName" validate:"required"`
	KeyName  string `json:"keyName,omitempty" form:"keyName" query:"keyName" validate:"required"`
}

type Page struct {
	fieldsSet []bool
	Url       url.URL  `json:"url,omitempty" form:"url" query:"url" validate:"required"`
	KeyNames  []string `json:"keyNames,omitempty" form:"keyNames" query:"keyNames" validate:"required"`
	extraData []byte
}

type SeedInfo struct {
	Mnemonic    string            `json:"mnemonic,omitempty" form:"mnemonic" query:"mnemonic" validate:"required"`
	Seed        []byte            `json:"seed,omitempty" form:"seed" query:"seed" validate:"required"`
	Derivations []DerivationCount `json:"derivations,omitempty" form:"derivations" query:"derivations"`
}

type Version struct {
	Major    uint64 `json:"major,omitempty" form:"major" query:"major" validate:"required"`
	Minor    uint64 `json:"minor,omitempty" form:"minor" query:"minor" validate:"required"`
	Revision uint64 `json:"revision,omitempty" form:"revision" query:"revision" validate:"required"`
	Commit   uint64 `json:"commit,omitempty" form:"commit" query:"commit" validate:"required"`
}

type Wallet struct {
	Version    Version     `json:"version,omitempty" form:"version" query:"version" validate:"required"`
	SeedInfo   SeedInfo    `json:"seedInfo,omitempty" form:"seedInfo" query:"seedInfo" validate:"required"`
	Keys       []Key       `json:"keys,omitempty" form:"keys" query:"keys" validate:"required"`
	KeyNames   []KeyName   `json:"keyNames,omitempty" form:"keyNames" query:"keyNames" validate:"required"`
	LiteLabels []LiteLabel `json:"liteLabels,omitempty" form:"liteLabels" query:"liteLabels" validate:"required"`
	Adis       []Adi       `json:"adis,omitempty" form:"adis" query:"adis"`
}

func (v *Adi) Copy() *Adi {
	u := new(Adi)

	u.Url = *&v.Url
	u.Pages = make([]Page, len(v.Pages))
	for i, v := range v.Pages {
		u.Pages[i] = *(&v).Copy()
	}

	return u
}

func (v *Adi) CopyAsInterface() interface{} { return v.Copy() }

func (v *DerivationCount) Copy() *DerivationCount {
	u := new(DerivationCount)

	u.Type = v.Type
	u.Count = v.Count

	return u
}

func (v *DerivationCount) CopyAsInterface() interface{} { return v.Copy() }

func (v *Key) Copy() *Key {
	u := new(Key)

	u.PrivateKey = encoding.BytesCopy(v.PrivateKey)
	u.PublicKey = encoding.BytesCopy(v.PublicKey)
	u.KeyInfo = *(&v.KeyInfo).Copy()

	return u
}

func (v *Key) CopyAsInterface() interface{} { return v.Copy() }

func (v *KeyInfo) Copy() *KeyInfo {
	u := new(KeyInfo)

	u.Type = v.Type
	u.Derivation = v.Derivation
	if v.WalletID != nil {
		u.WalletID = v.WalletID
	}

	return u
}

func (v *KeyInfo) CopyAsInterface() interface{} { return v.Copy() }

func (v *KeyName) Copy() *KeyName {
	u := new(KeyName)

	u.Name = v.Name
	u.PublicKey = encoding.BytesCopy(v.PublicKey)

	return u
}

func (v *KeyName) CopyAsInterface() interface{} { return v.Copy() }

func (v *LiteLabel) Copy() *LiteLabel {
	u := new(LiteLabel)

	u.LiteName = v.LiteName
	u.KeyName = v.KeyName

	return u
}

func (v *LiteLabel) CopyAsInterface() interface{} { return v.Copy() }

func (v *Page) Copy() *Page {
	u := new(Page)

	u.Url = *&v.Url
	u.KeyNames = make([]string, len(v.KeyNames))
	for i, v := range v.KeyNames {
		u.KeyNames[i] = v
	}

	return u
}

func (v *Page) CopyAsInterface() interface{} { return v.Copy() }

func (v *SeedInfo) Copy() *SeedInfo {
	u := new(SeedInfo)

	u.Mnemonic = v.Mnemonic
	u.Seed = encoding.BytesCopy(v.Seed)
	u.Derivations = make([]DerivationCount, len(v.Derivations))
	for i, v := range v.Derivations {
		u.Derivations[i] = *(&v).Copy()
	}

	return u
}

func (v *SeedInfo) CopyAsInterface() interface{} { return v.Copy() }

func (v *Version) Copy() *Version {
	u := new(Version)

	u.Major = v.Major
	u.Minor = v.Minor
	u.Revision = v.Revision
	u.Commit = v.Commit

	return u
}

func (v *Version) CopyAsInterface() interface{} { return v.Copy() }

func (v *Wallet) Copy() *Wallet {
	u := new(Wallet)

	u.Version = *(&v.Version).Copy()
	u.SeedInfo = *(&v.SeedInfo).Copy()
	u.Keys = make([]Key, len(v.Keys))
	for i, v := range v.Keys {
		u.Keys[i] = *(&v).Copy()
	}
	u.KeyNames = make([]KeyName, len(v.KeyNames))
	for i, v := range v.KeyNames {
		u.KeyNames[i] = *(&v).Copy()
	}
	u.LiteLabels = make([]LiteLabel, len(v.LiteLabels))
	for i, v := range v.LiteLabels {
		u.LiteLabels[i] = *(&v).Copy()
	}
	u.Adis = make([]Adi, len(v.Adis))
	for i, v := range v.Adis {
		u.Adis[i] = *(&v).Copy()
	}

	return u
}

func (v *Wallet) CopyAsInterface() interface{} { return v.Copy() }

func (v *Adi) Equal(u *Adi) bool {
	if !((&v.Url).Equal(&u.Url)) {
		return false
	}
	if len(v.Pages) != len(u.Pages) {
		return false
	}
	for i := range v.Pages {
		if !((&v.Pages[i]).Equal(&u.Pages[i])) {
			return false
		}
	}

	return true
}

func (v *DerivationCount) Equal(u *DerivationCount) bool {
	if !(v.Type == u.Type) {
		return false
	}
	if !(v.Count == u.Count) {
		return false
	}

	return true
}

func (v *Key) Equal(u *Key) bool {
	if !(bytes.Equal(v.PrivateKey, u.PrivateKey)) {
		return false
	}
	if !(bytes.Equal(v.PublicKey, u.PublicKey)) {
		return false
	}
	if !((&v.KeyInfo).Equal(&u.KeyInfo)) {
		return false
	}

	return true
}

func (v *KeyInfo) Equal(u *KeyInfo) bool {
	if !(v.Type == u.Type) {
		return false
	}
	if !(v.Derivation == u.Derivation) {
		return false
	}
	switch {
	case v.WalletID == u.WalletID:
		// equal
	case v.WalletID == nil || u.WalletID == nil:
		return false
	case !((v.WalletID).Equal(u.WalletID)):
		return false
	}

	return true
}

func (v *KeyName) Equal(u *KeyName) bool {
	if !(v.Name == u.Name) {
		return false
	}
	if !(bytes.Equal(v.PublicKey, u.PublicKey)) {
		return false
	}

	return true
}

func (v *LiteLabel) Equal(u *LiteLabel) bool {
	if !(v.LiteName == u.LiteName) {
		return false
	}
	if !(v.KeyName == u.KeyName) {
		return false
	}

	return true
}

func (v *Page) Equal(u *Page) bool {
	if !((&v.Url).Equal(&u.Url)) {
		return false
	}
	if len(v.KeyNames) != len(u.KeyNames) {
		return false
	}
	for i := range v.KeyNames {
		if !(v.KeyNames[i] == u.KeyNames[i]) {
			return false
		}
	}

	return true
}

func (v *SeedInfo) Equal(u *SeedInfo) bool {
	if !(v.Mnemonic == u.Mnemonic) {
		return false
	}
	if !(bytes.Equal(v.Seed, u.Seed)) {
		return false
	}
	if len(v.Derivations) != len(u.Derivations) {
		return false
	}
	for i := range v.Derivations {
		if !((&v.Derivations[i]).Equal(&u.Derivations[i])) {
			return false
		}
	}

	return true
}

func (v *Version) Equal(u *Version) bool {
	if !(v.Major == u.Major) {
		return false
	}
	if !(v.Minor == u.Minor) {
		return false
	}
	if !(v.Revision == u.Revision) {
		return false
	}
	if !(v.Commit == u.Commit) {
		return false
	}

	return true
}

func (v *Wallet) Equal(u *Wallet) bool {
	if !((&v.Version).Equal(&u.Version)) {
		return false
	}
	if !((&v.SeedInfo).Equal(&u.SeedInfo)) {
		return false
	}
	if len(v.Keys) != len(u.Keys) {
		return false
	}
	for i := range v.Keys {
		if !((&v.Keys[i]).Equal(&u.Keys[i])) {
			return false
		}
	}
	if len(v.KeyNames) != len(u.KeyNames) {
		return false
	}
	for i := range v.KeyNames {
		if !((&v.KeyNames[i]).Equal(&u.KeyNames[i])) {
			return false
		}
	}
	if len(v.LiteLabels) != len(u.LiteLabels) {
		return false
	}
	for i := range v.LiteLabels {
		if !((&v.LiteLabels[i]).Equal(&u.LiteLabels[i])) {
			return false
		}
	}
	if len(v.Adis) != len(u.Adis) {
		return false
	}
	for i := range v.Adis {
		if !((&v.Adis[i]).Equal(&u.Adis[i])) {
			return false
		}
	}

	return true
}

var fieldNames_Adi = []string{
	1: "Url",
	2: "Pages",
}

func (v *Adi) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(v.Url == (url.URL{})) {
		writer.WriteUrl(1, &v.Url)
	}
	if !(len(v.Pages) == 0) {
		for _, v := range v.Pages {
			writer.WriteValue(2, v.MarshalBinary)
		}
	}

	_, _, err := writer.Reset(fieldNames_Adi)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *Adi) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Url is missing")
	} else if v.Url == (url.URL{}) {
		errs = append(errs, "field Url is not set")
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

var fieldNames_Key = []string{
	1: "PrivateKey",
	2: "PublicKey",
	3: "KeyInfo",
}

func (v *Key) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(len(v.PrivateKey) == 0) {
		writer.WriteBytes(1, v.PrivateKey)
	}
	if !(len(v.PublicKey) == 0) {
		writer.WriteBytes(2, v.PublicKey)
	}
	if !((v.KeyInfo).Equal(new(KeyInfo))) {
		writer.WriteValue(3, v.KeyInfo.MarshalBinary)
	}

	_, _, err := writer.Reset(fieldNames_Key)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *Key) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field PrivateKey is missing")
	} else if len(v.PrivateKey) == 0 {
		errs = append(errs, "field PrivateKey is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field PublicKey is missing")
	} else if len(v.PublicKey) == 0 {
		errs = append(errs, "field PublicKey is not set")
	}
	if len(v.fieldsSet) > 3 && !v.fieldsSet[3] {
		errs = append(errs, "field KeyInfo is missing")
	} else if (v.KeyInfo).Equal(new(KeyInfo)) {
		errs = append(errs, "field KeyInfo is not set")
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

var fieldNames_KeyInfo = []string{
	1: "Type",
	2: "Derivation",
	3: "WalletID",
}

func (v *KeyInfo) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(v.Type == 0) {
		writer.WriteEnum(1, v.Type)
	}
	if !(len(v.Derivation) == 0) {
		writer.WriteString(2, v.Derivation)
	}
	if !(v.WalletID == nil) {
		writer.WriteUrl(3, v.WalletID)
	}

	_, _, err := writer.Reset(fieldNames_KeyInfo)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *KeyInfo) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Type is missing")
	} else if v.Type == 0 {
		errs = append(errs, "field Type is not set")
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

var fieldNames_KeyName = []string{
	1: "Name",
	2: "PublicKey",
}

func (v *KeyName) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(len(v.Name) == 0) {
		writer.WriteString(1, v.Name)
	}
	if !(len(v.PublicKey) == 0) {
		writer.WriteBytes(2, v.PublicKey)
	}

	_, _, err := writer.Reset(fieldNames_KeyName)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *KeyName) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Name is missing")
	} else if len(v.Name) == 0 {
		errs = append(errs, "field Name is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field PublicKey is missing")
	} else if len(v.PublicKey) == 0 {
		errs = append(errs, "field PublicKey is not set")
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

var fieldNames_Page = []string{
	1: "Url",
	2: "KeyNames",
}

func (v *Page) MarshalBinary() ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := encoding.NewWriter(buffer)

	if !(v.Url == (url.URL{})) {
		writer.WriteUrl(1, &v.Url)
	}
	if !(len(v.KeyNames) == 0) {
		for _, v := range v.KeyNames {
			writer.WriteString(2, v)
		}
	}

	_, _, err := writer.Reset(fieldNames_Page)
	if err != nil {
		return nil, encoding.Error{E: err}
	}
	buffer.Write(v.extraData)
	return buffer.Bytes(), nil
}

func (v *Page) IsValid() error {
	var errs []string

	if len(v.fieldsSet) > 1 && !v.fieldsSet[1] {
		errs = append(errs, "field Url is missing")
	} else if v.Url == (url.URL{}) {
		errs = append(errs, "field Url is not set")
	}
	if len(v.fieldsSet) > 2 && !v.fieldsSet[2] {
		errs = append(errs, "field KeyNames is missing")
	} else if len(v.KeyNames) == 0 {
		errs = append(errs, "field KeyNames is not set")
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

func (v *Adi) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *Adi) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x, ok := reader.ReadUrl(1); ok {
		v.Url = *x
	}
	for {
		if x := new(Page); reader.ReadValue(2, x.UnmarshalBinary) {
			v.Pages = append(v.Pages, *x)
		} else {
			break
		}
	}

	seen, err := reader.Reset(fieldNames_Adi)
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

func (v *Key) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *Key) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x, ok := reader.ReadBytes(1); ok {
		v.PrivateKey = x
	}
	if x, ok := reader.ReadBytes(2); ok {
		v.PublicKey = x
	}
	if x := new(KeyInfo); reader.ReadValue(3, x.UnmarshalBinary) {
		v.KeyInfo = *x
	}

	seen, err := reader.Reset(fieldNames_Key)
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

func (v *KeyInfo) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *KeyInfo) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x := new(protocol.SignatureType); reader.ReadEnum(1, x) {
		v.Type = *x
	}
	if x, ok := reader.ReadString(2); ok {
		v.Derivation = x
	}
	if x, ok := reader.ReadUrl(3); ok {
		v.WalletID = x
	}

	seen, err := reader.Reset(fieldNames_KeyInfo)
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

func (v *KeyName) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *KeyName) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x, ok := reader.ReadString(1); ok {
		v.Name = x
	}
	if x, ok := reader.ReadBytes(2); ok {
		v.PublicKey = x
	}

	seen, err := reader.Reset(fieldNames_KeyName)
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

func (v *Page) UnmarshalBinary(data []byte) error {
	return v.UnmarshalBinaryFrom(bytes.NewReader(data))
}

func (v *Page) UnmarshalBinaryFrom(rd io.Reader) error {
	reader := encoding.NewReader(rd)

	if x, ok := reader.ReadUrl(1); ok {
		v.Url = *x
	}
	for {
		if x, ok := reader.ReadString(2); ok {
			v.KeyNames = append(v.KeyNames, x)
		} else {
			break
		}
	}

	seen, err := reader.Reset(fieldNames_Page)
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

func (v *Adi) MarshalJSON() ([]byte, error) {
	u := struct {
		Url   url.URL                 `json:"url,omitempty"`
		Pages encoding.JsonList[Page] `json:"pages,omitempty"`
	}{}
	u.Url = v.Url
	u.Pages = v.Pages
	return json.Marshal(&u)
}

func (v *Key) MarshalJSON() ([]byte, error) {
	u := struct {
		PrivateKey *string `json:"privateKey,omitempty"`
		PublicKey  *string `json:"publicKey,omitempty"`
		KeyInfo    KeyInfo `json:"keyInfo,omitempty"`
	}{}
	u.PrivateKey = encoding.BytesToJSON(v.PrivateKey)
	u.PublicKey = encoding.BytesToJSON(v.PublicKey)
	u.KeyInfo = v.KeyInfo
	return json.Marshal(&u)
}

func (v *KeyName) MarshalJSON() ([]byte, error) {
	u := struct {
		Name      string  `json:"name,omitempty"`
		PublicKey *string `json:"publicKey,omitempty"`
	}{}
	u.Name = v.Name
	u.PublicKey = encoding.BytesToJSON(v.PublicKey)
	return json.Marshal(&u)
}

func (v *Page) MarshalJSON() ([]byte, error) {
	u := struct {
		Url      url.URL                   `json:"url,omitempty"`
		KeyNames encoding.JsonList[string] `json:"keyNames,omitempty"`
	}{}
	u.Url = v.Url
	u.KeyNames = v.KeyNames
	return json.Marshal(&u)
}

func (v *SeedInfo) MarshalJSON() ([]byte, error) {
	u := struct {
		Mnemonic    string                             `json:"mnemonic,omitempty"`
		Seed        *string                            `json:"seed,omitempty"`
		Derivations encoding.JsonList[DerivationCount] `json:"derivations,omitempty"`
	}{}
	u.Mnemonic = v.Mnemonic
	u.Seed = encoding.BytesToJSON(v.Seed)
	u.Derivations = v.Derivations
	return json.Marshal(&u)
}

func (v *Wallet) MarshalJSON() ([]byte, error) {
	u := struct {
		Version    Version                      `json:"version,omitempty"`
		SeedInfo   SeedInfo                     `json:"seedInfo,omitempty"`
		Keys       encoding.JsonList[Key]       `json:"keys,omitempty"`
		KeyNames   encoding.JsonList[KeyName]   `json:"keyNames,omitempty"`
		LiteLabels encoding.JsonList[LiteLabel] `json:"liteLabels,omitempty"`
		Adis       encoding.JsonList[Adi]       `json:"adis,omitempty"`
	}{}
	u.Version = v.Version
	u.SeedInfo = v.SeedInfo
	u.Keys = v.Keys
	u.KeyNames = v.KeyNames
	u.LiteLabels = v.LiteLabels
	u.Adis = v.Adis
	return json.Marshal(&u)
}

func (v *Adi) UnmarshalJSON(data []byte) error {
	u := struct {
		Url   url.URL                 `json:"url,omitempty"`
		Pages encoding.JsonList[Page] `json:"pages,omitempty"`
	}{}
	u.Url = v.Url
	u.Pages = v.Pages
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Url = u.Url
	v.Pages = u.Pages
	return nil
}

func (v *Key) UnmarshalJSON(data []byte) error {
	u := struct {
		PrivateKey *string `json:"privateKey,omitempty"`
		PublicKey  *string `json:"publicKey,omitempty"`
		KeyInfo    KeyInfo `json:"keyInfo,omitempty"`
	}{}
	u.PrivateKey = encoding.BytesToJSON(v.PrivateKey)
	u.PublicKey = encoding.BytesToJSON(v.PublicKey)
	u.KeyInfo = v.KeyInfo
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	if x, err := encoding.BytesFromJSON(u.PrivateKey); err != nil {
		return fmt.Errorf("error decoding PrivateKey: %w", err)
	} else {
		v.PrivateKey = x
	}
	if x, err := encoding.BytesFromJSON(u.PublicKey); err != nil {
		return fmt.Errorf("error decoding PublicKey: %w", err)
	} else {
		v.PublicKey = x
	}
	v.KeyInfo = u.KeyInfo
	return nil
}

func (v *KeyName) UnmarshalJSON(data []byte) error {
	u := struct {
		Name      string  `json:"name,omitempty"`
		PublicKey *string `json:"publicKey,omitempty"`
	}{}
	u.Name = v.Name
	u.PublicKey = encoding.BytesToJSON(v.PublicKey)
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Name = u.Name
	if x, err := encoding.BytesFromJSON(u.PublicKey); err != nil {
		return fmt.Errorf("error decoding PublicKey: %w", err)
	} else {
		v.PublicKey = x
	}
	return nil
}

func (v *Page) UnmarshalJSON(data []byte) error {
	u := struct {
		Url      url.URL                   `json:"url,omitempty"`
		KeyNames encoding.JsonList[string] `json:"keyNames,omitempty"`
	}{}
	u.Url = v.Url
	u.KeyNames = v.KeyNames
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Url = u.Url
	v.KeyNames = u.KeyNames
	return nil
}

func (v *SeedInfo) UnmarshalJSON(data []byte) error {
	u := struct {
		Mnemonic    string                             `json:"mnemonic,omitempty"`
		Seed        *string                            `json:"seed,omitempty"`
		Derivations encoding.JsonList[DerivationCount] `json:"derivations,omitempty"`
	}{}
	u.Mnemonic = v.Mnemonic
	u.Seed = encoding.BytesToJSON(v.Seed)
	u.Derivations = v.Derivations
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Mnemonic = u.Mnemonic
	if x, err := encoding.BytesFromJSON(u.Seed); err != nil {
		return fmt.Errorf("error decoding Seed: %w", err)
	} else {
		v.Seed = x
	}
	v.Derivations = u.Derivations
	return nil
}

func (v *Wallet) UnmarshalJSON(data []byte) error {
	u := struct {
		Version    Version                      `json:"version,omitempty"`
		SeedInfo   SeedInfo                     `json:"seedInfo,omitempty"`
		Keys       encoding.JsonList[Key]       `json:"keys,omitempty"`
		KeyNames   encoding.JsonList[KeyName]   `json:"keyNames,omitempty"`
		LiteLabels encoding.JsonList[LiteLabel] `json:"liteLabels,omitempty"`
		Adis       encoding.JsonList[Adi]       `json:"adis,omitempty"`
	}{}
	u.Version = v.Version
	u.SeedInfo = v.SeedInfo
	u.Keys = v.Keys
	u.KeyNames = v.KeyNames
	u.LiteLabels = v.LiteLabels
	u.Adis = v.Adis
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	v.Version = u.Version
	v.SeedInfo = u.SeedInfo
	v.Keys = u.Keys
	v.KeyNames = u.KeyNames
	v.LiteLabels = u.LiteLabels
	v.Adis = u.Adis
	return nil
}
