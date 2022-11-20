package main

import (
	_ "embed"
	"fmt"
	"text/template"
)

//go:embed types.ts.tmpl
var tsSrc string

func init() {
	Templates.Register(tsSrc, "typescript", tsFuncs, "TypeScript")
}

var tsFuncs = template.FuncMap{
	"resolveType": TsResolveType,
	"objectify":   TsObjectify,
}

func TsResolveType(field *Field) string {
	typ := field.Type.TypescriptType()
	if field.Repeatable {
		typ = typ + "[]"
	}
	return typ
}

func TsObjectify(field *Field, varName string) string {
	switch field.Type.Code {
	case BigInt, Url, TxID:
		return fmt.Sprintf("%s.toString()", varName)
	case Bytes, Hash:
		return fmt.Sprintf("Buffer.from(%s).toString('hex')", varName)
	}

	switch field.MarshalAs {
	case Reference, Union, Value:
		return fmt.Sprintf("%s.asObject()", varName)
	case Enum:
		return fmt.Sprintf("%s.toString()", varName)
	}

	return varName
}
