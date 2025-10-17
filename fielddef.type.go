package vulpo

import "strings"

type FieldType int

var dbfFieldTypes = "CNLDITYMBFGPQVWX"

const (
	FTUnknown   FieldType = iota
	FTCharacter           // C - Character/String
	FTNumeric             // N - Numeric
	FTLogical             // L - Logical/Boolean
	FTDate                // D - Date
	FTInteger             // I - Integer (32-bit)
	FTDateTime            // T - DateTime
	FTCurrency            // Y - Currency
	FTMemo                // M - Memo
	FTBlob                // B - Binary/Blob (deprecated)
	FTFloat               // F - Float
	FTGeneral             // G - General (OLE object)
	FTPicture             // P - Picture (OLE object)
	FTVarBinary           // Q - VarBinary
	FTVarchar             // V - Varchar
	FTTimestamp           // W - Timestamp (not standard)
	FTDouble              // X - Double (not standard)
)

func FromString(s string) FieldType {
	if len(s) != 1 {
		return FTUnknown
	}

	idx := strings.Index(dbfFieldTypes, strings.ToUpper(s))
	if idx == -1 {
		return FTUnknown
	}

	return FieldType(idx + 1)
}

func (ft FieldType) String() string {
	if ft >= 1 && int(ft) <= len(dbfFieldTypes) {
		return string(dbfFieldTypes[ft-1])
	}
	return "unknown"
}

//nolint:gocyclo // TODO: consider using a map lookup instead of switch for field type names
func (ft FieldType) Name() string {
	switch ft {
	case FTCharacter:
		return "character"
	case FTNumeric:
		return "numeric"
	case FTLogical:
		return "logical"
	case FTDate:
		return "date"
	case FTInteger:
		return "integer"
	case FTDateTime:
		return "datetime"
	case FTCurrency:
		return "currency"
	case FTMemo:
		return "memo"
	case FTBlob:
		return "blob"
	case FTFloat:
		return "float"
	case FTGeneral:
		return "general"
	case FTPicture:
		return "picture"
	case FTVarBinary:
		return "varbinary"
	case FTVarchar:
		return "varchar"
	case FTTimestamp:
		return "timestamp"
	case FTDouble:
		return "double"
	default:
		return "unknown"
	}
}
