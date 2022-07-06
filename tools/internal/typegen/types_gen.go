package typegen

// GENERATED BY go run ./tools/cmd/gen-types. DO NOT EDIT.

import (
	"encoding/json"
	"fmt"

	"gitlab.com/accumulatenetwork/accumulate/internal/encoding"
)

type ChainRecord struct {
	Parent       *EntityRecord `json:"parent,omitempty" form:"parent" query:"parent" validate:"required"`
	OmitAccessor bool          `json:"omitAccessor,omitempty" form:"omitAccessor" query:"omitAccessor" validate:"required"`
	Private      bool          `json:"private,omitempty" form:"private" query:"private" validate:"required"`
	Name         string        `json:"name,omitempty" form:"name" query:"name" validate:"required"`
	Parameters   []*Field      `json:"parameters,omitempty" form:"parameters" query:"parameters" validate:"required"`
	ChainType    string        `json:"chainType,omitempty" form:"chainType" query:"chainType" validate:"required"`
	extraData    []byte
}

type EntityRecord struct {
	Parent        *EntityRecord `json:"parent,omitempty" form:"parent" query:"parent" validate:"required"`
	OmitAccessor  bool          `json:"omitAccessor,omitempty" form:"omitAccessor" query:"omitAccessor" validate:"required"`
	Private       bool          `json:"private,omitempty" form:"private" query:"private" validate:"required"`
	Name          string        `json:"name,omitempty" form:"name" query:"name" validate:"required"`
	Fields        []*Field      `json:"fields,omitempty" form:"fields" query:"fields" validate:"required"`
	CustomCommit  bool          `json:"customCommit,omitempty" form:"customCommit" query:"customCommit" validate:"required"`
	CustomResolve bool          `json:"customResolve,omitempty" form:"customResolve" query:"customResolve" validate:"required"`
	CustomIsDirty bool          `json:"customIsDirty,omitempty" form:"customIsDirty" query:"customIsDirty" validate:"required"`
	Parameters    []*Field      `json:"parameters,omitempty" form:"parameters" query:"parameters" validate:"required"`
	Root          bool          `json:"root,omitempty" form:"root" query:"root" validate:"required"`
	Attributes    []Record      `json:"attributes,omitempty" form:"attributes" query:"attributes" validate:"required"`
	extraData     []byte
}

type IndexRecord struct {
	Parent         *EntityRecord  `json:"parent,omitempty" form:"parent" query:"parent" validate:"required"`
	OmitAccessor   bool           `json:"omitAccessor,omitempty" form:"omitAccessor" query:"omitAccessor" validate:"required"`
	Private        bool           `json:"private,omitempty" form:"private" query:"private" validate:"required"`
	Name           string         `json:"name,omitempty" form:"name" query:"name" validate:"required"`
	Parameters     []*Field       `json:"parameters,omitempty" form:"parameters" query:"parameters" validate:"required"`
	DataType       FieldType      `json:"dataType,omitempty" form:"dataType" query:"dataType" validate:"required"`
	Pointer        bool           `json:"pointer,omitempty" form:"pointer" query:"pointer" validate:"required"`
	EmptyIfMissing bool           `json:"emptyIfMissing,omitempty" form:"emptyIfMissing" query:"emptyIfMissing" validate:"required"`
	Union          bool           `json:"union,omitempty" form:"union" query:"union" validate:"required"`
	Collection     CollectionType `json:"collection,omitempty" form:"collection" query:"collection" validate:"required"`
	extraData      []byte
}

type OtherRecord struct {
	Parent       *EntityRecord `json:"parent,omitempty" form:"parent" query:"parent" validate:"required"`
	OmitAccessor bool          `json:"omitAccessor,omitempty" form:"omitAccessor" query:"omitAccessor" validate:"required"`
	Private      bool          `json:"private,omitempty" form:"private" query:"private" validate:"required"`
	Name         string        `json:"name,omitempty" form:"name" query:"name" validate:"required"`
	DataType     string        `json:"dataType,omitempty" form:"dataType" query:"dataType" validate:"required"`
	Parameters   []*Field      `json:"parameters,omitempty" form:"parameters" query:"parameters" validate:"required"`
	Pointer      bool          `json:"pointer,omitempty" form:"pointer" query:"pointer" validate:"required"`
	HasChains    bool          `json:"hasChains,omitempty" form:"hasChains" query:"hasChains" validate:"required"`
	extraData    []byte
}

type StateRecord struct {
	Parent         *EntityRecord  `json:"parent,omitempty" form:"parent" query:"parent" validate:"required"`
	OmitAccessor   bool           `json:"omitAccessor,omitempty" form:"omitAccessor" query:"omitAccessor" validate:"required"`
	Private        bool           `json:"private,omitempty" form:"private" query:"private" validate:"required"`
	Name           string         `json:"name,omitempty" form:"name" query:"name" validate:"required"`
	Parameters     []*Field       `json:"parameters,omitempty" form:"parameters" query:"parameters" validate:"required"`
	DataType       FieldType      `json:"dataType,omitempty" form:"dataType" query:"dataType" validate:"required"`
	Pointer        bool           `json:"pointer,omitempty" form:"pointer" query:"pointer" validate:"required"`
	EmptyIfMissing bool           `json:"emptyIfMissing,omitempty" form:"emptyIfMissing" query:"emptyIfMissing" validate:"required"`
	Union          bool           `json:"union,omitempty" form:"union" query:"union" validate:"required"`
	Collection     CollectionType `json:"collection,omitempty" form:"collection" query:"collection" validate:"required"`
	extraData      []byte
}

func (*ChainRecord) Type() RecordType { return RecordTypeChain }

func (*EntityRecord) Type() RecordType { return RecordTypeEntity }

func (*IndexRecord) Type() RecordType { return RecordTypeIndex }

func (*OtherRecord) Type() RecordType { return RecordTypeOther }

func (*StateRecord) Type() RecordType { return RecordTypeState }

func (v *ChainRecord) MarshalJSON() ([]byte, error) {
	u := struct {
		Type         RecordType                `json:"type"`
		Parent       *EntityRecord             `json:"parent,omitempty"`
		OmitAccessor bool                      `json:"omitAccessor,omitempty"`
		Private      bool                      `json:"private,omitempty"`
		Name         string                    `json:"name,omitempty"`
		Parameters   encoding.JsonList[*Field] `json:"parameters,omitempty"`
		ChainType    string                    `json:"chainType,omitempty"`
	}{}
	u.Type = v.Type()
	u.Parent = v.Parent
	u.OmitAccessor = v.OmitAccessor
	u.Private = v.Private
	u.Name = v.Name
	u.Parameters = v.Parameters
	u.ChainType = v.ChainType
	return json.Marshal(&u)
}

func (v *EntityRecord) MarshalJSON() ([]byte, error) {
	u := struct {
		Type          RecordType                             `json:"type"`
		Parent        *EntityRecord                          `json:"parent,omitempty"`
		OmitAccessor  bool                                   `json:"omitAccessor,omitempty"`
		Private       bool                                   `json:"private,omitempty"`
		Name          string                                 `json:"name,omitempty"`
		Fields        encoding.JsonList[*Field]              `json:"fields,omitempty"`
		CustomCommit  bool                                   `json:"customCommit,omitempty"`
		CustomResolve bool                                   `json:"customResolve,omitempty"`
		CustomIsDirty bool                                   `json:"customIsDirty,omitempty"`
		Parameters    encoding.JsonList[*Field]              `json:"parameters,omitempty"`
		Root          bool                                   `json:"root,omitempty"`
		Attributes    encoding.JsonUnmarshalListWith[Record] `json:"attributes,omitempty"`
	}{}
	u.Type = v.Type()
	u.Parent = v.Parent
	u.OmitAccessor = v.OmitAccessor
	u.Private = v.Private
	u.Name = v.Name
	u.Fields = v.Fields
	u.CustomCommit = v.CustomCommit
	u.CustomResolve = v.CustomResolve
	u.CustomIsDirty = v.CustomIsDirty
	u.Parameters = v.Parameters
	u.Root = v.Root
	u.Attributes = encoding.JsonUnmarshalListWith[Record]{Value: v.Attributes, Func: UnmarshalRecordJSON}
	return json.Marshal(&u)
}

func (v *IndexRecord) MarshalJSON() ([]byte, error) {
	u := struct {
		Type           RecordType                `json:"type"`
		Parent         *EntityRecord             `json:"parent,omitempty"`
		OmitAccessor   bool                      `json:"omitAccessor,omitempty"`
		Private        bool                      `json:"private,omitempty"`
		Name           string                    `json:"name,omitempty"`
		Parameters     encoding.JsonList[*Field] `json:"parameters,omitempty"`
		DataType       FieldType                 `json:"dataType,omitempty"`
		Pointer        bool                      `json:"pointer,omitempty"`
		EmptyIfMissing bool                      `json:"emptyIfMissing,omitempty"`
		Union          bool                      `json:"union,omitempty"`
		Collection     CollectionType            `json:"collection,omitempty"`
	}{}
	u.Type = v.Type()
	u.Parent = v.Parent
	u.OmitAccessor = v.OmitAccessor
	u.Private = v.Private
	u.Name = v.Name
	u.Parameters = v.Parameters
	u.DataType = v.DataType
	u.Pointer = v.Pointer
	u.EmptyIfMissing = v.EmptyIfMissing
	u.Union = v.Union
	u.Collection = v.Collection
	return json.Marshal(&u)
}

func (v *OtherRecord) MarshalJSON() ([]byte, error) {
	u := struct {
		Type         RecordType                `json:"type"`
		Parent       *EntityRecord             `json:"parent,omitempty"`
		OmitAccessor bool                      `json:"omitAccessor,omitempty"`
		Private      bool                      `json:"private,omitempty"`
		Name         string                    `json:"name,omitempty"`
		DataType     string                    `json:"dataType,omitempty"`
		Parameters   encoding.JsonList[*Field] `json:"parameters,omitempty"`
		Pointer      bool                      `json:"pointer,omitempty"`
		HasChains    bool                      `json:"hasChains,omitempty"`
	}{}
	u.Type = v.Type()
	u.Parent = v.Parent
	u.OmitAccessor = v.OmitAccessor
	u.Private = v.Private
	u.Name = v.Name
	u.DataType = v.DataType
	u.Parameters = v.Parameters
	u.Pointer = v.Pointer
	u.HasChains = v.HasChains
	return json.Marshal(&u)
}

func (v *StateRecord) MarshalJSON() ([]byte, error) {
	u := struct {
		Type           RecordType                `json:"type"`
		Parent         *EntityRecord             `json:"parent,omitempty"`
		OmitAccessor   bool                      `json:"omitAccessor,omitempty"`
		Private        bool                      `json:"private,omitempty"`
		Name           string                    `json:"name,omitempty"`
		Parameters     encoding.JsonList[*Field] `json:"parameters,omitempty"`
		DataType       FieldType                 `json:"dataType,omitempty"`
		Pointer        bool                      `json:"pointer,omitempty"`
		EmptyIfMissing bool                      `json:"emptyIfMissing,omitempty"`
		Union          bool                      `json:"union,omitempty"`
		Collection     CollectionType            `json:"collection,omitempty"`
	}{}
	u.Type = v.Type()
	u.Parent = v.Parent
	u.OmitAccessor = v.OmitAccessor
	u.Private = v.Private
	u.Name = v.Name
	u.Parameters = v.Parameters
	u.DataType = v.DataType
	u.Pointer = v.Pointer
	u.EmptyIfMissing = v.EmptyIfMissing
	u.Union = v.Union
	u.Collection = v.Collection
	return json.Marshal(&u)
}

func (v *ChainRecord) UnmarshalJSON(data []byte) error {
	u := struct {
		Type         RecordType                `json:"type"`
		Parent       *EntityRecord             `json:"parent,omitempty"`
		OmitAccessor bool                      `json:"omitAccessor,omitempty"`
		Private      bool                      `json:"private,omitempty"`
		Name         string                    `json:"name,omitempty"`
		Parameters   encoding.JsonList[*Field] `json:"parameters,omitempty"`
		ChainType    string                    `json:"chainType,omitempty"`
	}{}
	u.Type = v.Type()
	u.Parent = v.Parent
	u.OmitAccessor = v.OmitAccessor
	u.Private = v.Private
	u.Name = v.Name
	u.Parameters = v.Parameters
	u.ChainType = v.ChainType
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	if !(v.Type() == u.Type) {
		return fmt.Errorf("field Type: not equal: want %v, got %v", v.Type(), u.Type)
	}
	v.Parent = u.Parent
	v.OmitAccessor = u.OmitAccessor
	v.Private = u.Private
	v.Name = u.Name
	v.Parameters = u.Parameters
	v.ChainType = u.ChainType
	return nil
}

func (v *EntityRecord) UnmarshalJSON(data []byte) error {
	u := struct {
		Type          RecordType                             `json:"type"`
		Parent        *EntityRecord                          `json:"parent,omitempty"`
		OmitAccessor  bool                                   `json:"omitAccessor,omitempty"`
		Private       bool                                   `json:"private,omitempty"`
		Name          string                                 `json:"name,omitempty"`
		Fields        encoding.JsonList[*Field]              `json:"fields,omitempty"`
		CustomCommit  bool                                   `json:"customCommit,omitempty"`
		CustomResolve bool                                   `json:"customResolve,omitempty"`
		CustomIsDirty bool                                   `json:"customIsDirty,omitempty"`
		Parameters    encoding.JsonList[*Field]              `json:"parameters,omitempty"`
		Root          bool                                   `json:"root,omitempty"`
		Attributes    encoding.JsonUnmarshalListWith[Record] `json:"attributes,omitempty"`
	}{}
	u.Type = v.Type()
	u.Parent = v.Parent
	u.OmitAccessor = v.OmitAccessor
	u.Private = v.Private
	u.Name = v.Name
	u.Fields = v.Fields
	u.CustomCommit = v.CustomCommit
	u.CustomResolve = v.CustomResolve
	u.CustomIsDirty = v.CustomIsDirty
	u.Parameters = v.Parameters
	u.Root = v.Root
	u.Attributes = encoding.JsonUnmarshalListWith[Record]{Value: v.Attributes, Func: UnmarshalRecordJSON}
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	if !(v.Type() == u.Type) {
		return fmt.Errorf("field Type: not equal: want %v, got %v", v.Type(), u.Type)
	}
	v.Parent = u.Parent
	v.OmitAccessor = u.OmitAccessor
	v.Private = u.Private
	v.Name = u.Name
	v.Fields = u.Fields
	v.CustomCommit = u.CustomCommit
	v.CustomResolve = u.CustomResolve
	v.CustomIsDirty = u.CustomIsDirty
	v.Parameters = u.Parameters
	v.Root = u.Root
	v.Attributes = make([]Record, len(u.Attributes.Value))
	for i, x := range u.Attributes.Value {
		v.Attributes[i] = x
	}
	return nil
}

func (v *IndexRecord) UnmarshalJSON(data []byte) error {
	u := struct {
		Type           RecordType                `json:"type"`
		Parent         *EntityRecord             `json:"parent,omitempty"`
		OmitAccessor   bool                      `json:"omitAccessor,omitempty"`
		Private        bool                      `json:"private,omitempty"`
		Name           string                    `json:"name,omitempty"`
		Parameters     encoding.JsonList[*Field] `json:"parameters,omitempty"`
		DataType       FieldType                 `json:"dataType,omitempty"`
		Pointer        bool                      `json:"pointer,omitempty"`
		EmptyIfMissing bool                      `json:"emptyIfMissing,omitempty"`
		Union          bool                      `json:"union,omitempty"`
		Collection     CollectionType            `json:"collection,omitempty"`
	}{}
	u.Type = v.Type()
	u.Parent = v.Parent
	u.OmitAccessor = v.OmitAccessor
	u.Private = v.Private
	u.Name = v.Name
	u.Parameters = v.Parameters
	u.DataType = v.DataType
	u.Pointer = v.Pointer
	u.EmptyIfMissing = v.EmptyIfMissing
	u.Union = v.Union
	u.Collection = v.Collection
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	if !(v.Type() == u.Type) {
		return fmt.Errorf("field Type: not equal: want %v, got %v", v.Type(), u.Type)
	}
	v.Parent = u.Parent
	v.OmitAccessor = u.OmitAccessor
	v.Private = u.Private
	v.Name = u.Name
	v.Parameters = u.Parameters
	v.DataType = u.DataType
	v.Pointer = u.Pointer
	v.EmptyIfMissing = u.EmptyIfMissing
	v.Union = u.Union
	v.Collection = u.Collection
	return nil
}

func (v *OtherRecord) UnmarshalJSON(data []byte) error {
	u := struct {
		Type         RecordType                `json:"type"`
		Parent       *EntityRecord             `json:"parent,omitempty"`
		OmitAccessor bool                      `json:"omitAccessor,omitempty"`
		Private      bool                      `json:"private,omitempty"`
		Name         string                    `json:"name,omitempty"`
		DataType     string                    `json:"dataType,omitempty"`
		Parameters   encoding.JsonList[*Field] `json:"parameters,omitempty"`
		Pointer      bool                      `json:"pointer,omitempty"`
		HasChains    bool                      `json:"hasChains,omitempty"`
	}{}
	u.Type = v.Type()
	u.Parent = v.Parent
	u.OmitAccessor = v.OmitAccessor
	u.Private = v.Private
	u.Name = v.Name
	u.DataType = v.DataType
	u.Parameters = v.Parameters
	u.Pointer = v.Pointer
	u.HasChains = v.HasChains
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	if !(v.Type() == u.Type) {
		return fmt.Errorf("field Type: not equal: want %v, got %v", v.Type(), u.Type)
	}
	v.Parent = u.Parent
	v.OmitAccessor = u.OmitAccessor
	v.Private = u.Private
	v.Name = u.Name
	v.DataType = u.DataType
	v.Parameters = u.Parameters
	v.Pointer = u.Pointer
	v.HasChains = u.HasChains
	return nil
}

func (v *StateRecord) UnmarshalJSON(data []byte) error {
	u := struct {
		Type           RecordType                `json:"type"`
		Parent         *EntityRecord             `json:"parent,omitempty"`
		OmitAccessor   bool                      `json:"omitAccessor,omitempty"`
		Private        bool                      `json:"private,omitempty"`
		Name           string                    `json:"name,omitempty"`
		Parameters     encoding.JsonList[*Field] `json:"parameters,omitempty"`
		DataType       FieldType                 `json:"dataType,omitempty"`
		Pointer        bool                      `json:"pointer,omitempty"`
		EmptyIfMissing bool                      `json:"emptyIfMissing,omitempty"`
		Union          bool                      `json:"union,omitempty"`
		Collection     CollectionType            `json:"collection,omitempty"`
	}{}
	u.Type = v.Type()
	u.Parent = v.Parent
	u.OmitAccessor = v.OmitAccessor
	u.Private = v.Private
	u.Name = v.Name
	u.Parameters = v.Parameters
	u.DataType = v.DataType
	u.Pointer = v.Pointer
	u.EmptyIfMissing = v.EmptyIfMissing
	u.Union = v.Union
	u.Collection = v.Collection
	if err := json.Unmarshal(data, &u); err != nil {
		return err
	}
	if !(v.Type() == u.Type) {
		return fmt.Errorf("field Type: not equal: want %v, got %v", v.Type(), u.Type)
	}
	v.Parent = u.Parent
	v.OmitAccessor = u.OmitAccessor
	v.Private = u.Private
	v.Name = u.Name
	v.Parameters = u.Parameters
	v.DataType = u.DataType
	v.Pointer = u.Pointer
	v.EmptyIfMissing = u.EmptyIfMissing
	v.Union = u.Union
	v.Collection = u.Collection
	return nil
}
