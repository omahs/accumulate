// Copyright 2022 The Accumulate Authors
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://opensource.org/licenses/MIT.

package typegen

func (f FieldType) TypescriptType() string {
	switch f.Code {
	case TypeCodeUnknown:
		return f.Name
	case TypeCodeBytes:
		return "Uint8Array"
	case TypeCodeRawJson:
		return "unknown"
	case TypeCodeUrl:
		return "url.URL"
	case TypeCodeTxid:
		return "url.TxID"
	case TypeCodeBigInt:
		return "BN"
	case TypeCodeUint:
		return "number"
	case TypeCodeInt:
		return "number"
	case TypeCodeHash:
		return "Uint8Array"
	case TypeCodeDuration:
		return "number"
	case TypeCodeTime:
		return "Date"
	case TypeCodeBool:
		return "boolean"
	default:
		return f.Code.String()
	}
}
