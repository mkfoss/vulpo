package vulpo

/*
#include "d4all.h"
*/
import "C"
import (
	"time"
)

// MemoField represents a DBF memo field (type 'M')
// Memo fields store large text data and require special handling
// This implementation provides read-only access to memo contents.
type MemoField struct {
	baseField
	cField *C.FIELD4
}

// newMemoField creates a new MemoField instance
func newMemoField(field *C.FIELD4, data *Vulpo, def *FieldDef) *MemoField {
	return &MemoField{
		baseField: baseField{
			def:  def,
			data: data,
		},
		cField: field,
	}
}

// Value returns the field's memo content as a string
func (f *MemoField) Value() (interface{}, error) {
	return f.AsString()
}

// AsString returns the memo content as a string
func (f *MemoField) AsString() (string, error) {
	if err := f.checkActive(); err != nil {
		return "", err
	}

	// Use f4memoStr() to get memo content
	cStr := C.f4memoStr(f.cField)
	if cStr == nil {
		return "", nil
	}

	return C.GoString(cStr), nil
}

// AsInt cannot convert memo to int
func (f *MemoField) AsInt() (int, error) {
	return 0, NewConversionError("memo", "integer")
}

// AsFloat cannot convert memo to float
func (f *MemoField) AsFloat() (float64, error) {
	return 0, NewConversionError("memo", "float")
}

// AsBool returns true if memo content is non-empty
func (f *MemoField) AsBool() (bool, error) {
	str, err := f.AsString()
	if err != nil {
		return false, err
	}
	return str != "", nil
}

// AsTime cannot convert memo to time
func (f *MemoField) AsTime() (time.Time, error) {
	return time.Time{}, NewConversionError("memo", "time")
}

// IsNull returns true if the field is null
func (f *MemoField) IsNull() (bool, error) {
	if err := f.checkActive(); err != nil {
		return false, err
	}
	return C.f4null(f.cField) != 0, nil
}

// Field interface methods are inherited from baseField
