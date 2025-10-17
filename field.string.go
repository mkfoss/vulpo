package vulpo

/*
#include "d4all.h"
*/
import "C"
import (
	"strconv"
	"strings"
	"time"
)

// StringField handles character/string fields
type StringField struct {
	baseField
	cField *C.FIELD4
}

// newStringField creates a new StringField instance
func newStringField(field *C.FIELD4, data *Vulpo, def *FieldDef) *StringField {
	return &StringField{
		baseField: baseField{
			def:  def,
			data: data,
		},
		cField: field,
	}
}

// Value returns the field's string value
func (sf *StringField) Value() (interface{}, error) {
	if err := sf.checkActive(); err != nil {
		return nil, err
	}

	// Get string value using f4str()
	cStr := C.f4str(sf.cField)
	if cStr == nil {
		return "", nil
	}

	goStr := C.GoString(cStr)
	return strings.TrimSpace(goStr), nil
}

// AsString returns the field value as a string
func (sf *StringField) AsString() (string, error) {
	val, err := sf.Value()
	if err != nil {
		return "", err
	}
	return val.(string), nil
}

// AsInt attempts to convert the string to an integer
func (sf *StringField) AsInt() (int, error) {
	strVal, err := sf.AsString()
	if err != nil {
		return 0, err
	}

	strVal = strings.TrimSpace(strVal)
	if strVal == "" {
		return 0, nil
	}

	val, err := strconv.Atoi(strVal)
	if err != nil {
		return 0, NewConversionError("character", "integer")
	}
	return val, nil
}

// AsFloat attempts to convert the string to a float
func (sf *StringField) AsFloat() (float64, error) {
	strVal, err := sf.AsString()
	if err != nil {
		return 0, err
	}

	strVal = strings.TrimSpace(strVal)
	if strVal == "" {
		return 0, nil
	}

	val, err := strconv.ParseFloat(strVal, 64)
	if err != nil {
		return 0, NewConversionError("character", "float")
	}
	return val, nil
}

// AsBool attempts to convert the string to a boolean
func (sf *StringField) AsBool() (bool, error) {
	strVal, err := sf.AsString()
	if err != nil {
		return false, err
	}

	strVal = strings.ToUpper(strings.TrimSpace(strVal))
	switch strVal {
	case "T", "TRUE", "Y", "YES", "1":
		return true, nil
	case "F", "FALSE", "N", "NO", "0", "":
		return false, nil
	default:
		return false, NewConversionError("character", "boolean")
	}
}

// AsTime attempts to parse the string as a date/time
func (sf *StringField) AsTime() (time.Time, error) {
	strVal, err := sf.AsString()
	if err != nil {
		return time.Time{}, err
	}

	strVal = strings.TrimSpace(strVal)
	if strVal == "" {
		return time.Time{}, nil
	}

	// Try common date formats
	formats := []string{
		"2006-01-02",
		"01/02/2006",
		"02/01/2006",
		"2006/01/02",
		"20060102",
		time.RFC3339,
		time.RFC822,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, strVal); err == nil {
			return t, nil
		}
	}

	return time.Time{}, NewConversionError("character", "time")
}

// IsNull returns true if the field is null
func (sf *StringField) IsNull() (bool, error) {
	if err := sf.checkActive(); err != nil {
		return false, err
	}

	return C.f4null(sf.cField) != 0, nil
}

// Field interface methods are inherited from baseField

// String returns a string representation of the string field
func (sf *StringField) String() string {
	strVal, err := sf.AsString()
	if err != nil {
		return "StringField{name: " + sf.Name() + ", error: " + err.Error() + "}"
	}

	return "StringField{name: " + sf.Name() + ", value: \"" + strVal + "\"}"
}
