package vulpo

/*
#include "d4all.h"
*/
import "C"
import (
	"strconv"
	"time"
)

// DoubleField represents a DBF double field (type 'X')
// Double fields provide double precision floating point numbers
type DoubleField struct {
	baseField
	cField *C.FIELD4
}

// newDoubleField creates a new DoubleField instance
func newDoubleField(field *C.FIELD4, data *Vulpo, def *FieldDef) *DoubleField {
	return &DoubleField{
		baseField: baseField{
			def:  def,
			data: data,
		},
		cField: field,
	}
}

// Value returns the field value as float64
func (f *DoubleField) Value() (interface{}, error) {
	return f.AsFloat()
}

// AsString returns the double as a string
func (f *DoubleField) AsString() (string, error) {
	if err := f.checkActive(); err != nil {
		return "", err
	}

	doubleVal, err := f.AsFloat()
	if err != nil {
		return "", err
	}

	return strconv.FormatFloat(doubleVal, 'g', -1, 64), nil
}

// AsInt returns the double as an integer (truncated)
func (f *DoubleField) AsInt() (int, error) {
	doubleVal, err := f.AsFloat()
	if err != nil {
		return 0, err
	}

	return int(doubleVal), nil
}

// AsFloat returns the field value as a float64
func (f *DoubleField) AsFloat() (float64, error) {
	if err := f.checkActive(); err != nil {
		return 0, err
	}

	// Use f4double() to get the double value from CodeBase
	doubleVal := float64(C.f4double(f.cField))
	return doubleVal, nil
}

// AsBool returns true if the double is not zero
func (f *DoubleField) AsBool() (bool, error) {
	doubleVal, err := f.AsFloat()
	if err != nil {
		return false, err
	}

	return doubleVal != 0.0, nil
}

// AsTime cannot convert double to time
func (f *DoubleField) AsTime() (time.Time, error) {
	return time.Time{}, NewConversionError("double", "time")
}

// IsNull returns true if the field is null
func (f *DoubleField) IsNull() (bool, error) {
	if err := f.checkActive(); err != nil {
		return false, err
	}

	return C.f4null(f.cField) != 0, nil
}

// Field interface methods are inherited from baseField

// String returns a string representation of the double field
func (f *DoubleField) String() string {
	doubleStr, err := f.AsString()
	if err != nil {
		return "DoubleField{name: " + f.Name() + ", error: " + err.Error() + "}"
	}

	return "DoubleField{name: " + f.Name() + ", value: " + doubleStr + "}"
}
