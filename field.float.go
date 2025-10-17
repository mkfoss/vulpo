package vulpo

/*
#include "d4all.h"
*/
import "C"
import (
	"strconv"
	"time"
)

// FloatField represents a DBF float field (type 'F')
type FloatField struct {
	baseField
	cField *C.FIELD4
}

// newFloatField creates a new FloatField instance
func newFloatField(field *C.FIELD4, data *Vulpo, def *FieldDef) *FloatField {
	return &FloatField{
		baseField: baseField{
			def:  def,
			data: data,
		},
		cField: field,
	}
}

// Value returns the field value as float64
func (f *FloatField) Value() (interface{}, error) {
	return f.AsFloat()
}

// AsString returns the float as a string
func (f *FloatField) AsString() (string, error) {
	if err := f.checkActive(); err != nil {
		return "", err
	}

	floatVal, err := f.AsFloat()
	if err != nil {
		return "", err
	}

	return strconv.FormatFloat(floatVal, 'f', -1, 64), nil
}

// AsInt returns the float as an integer (truncated)
func (f *FloatField) AsInt() (int, error) {
	floatVal, err := f.AsFloat()
	if err != nil {
		return 0, err
	}

	return int(floatVal), nil
}

// AsFloat returns the field value as a float64
func (f *FloatField) AsFloat() (float64, error) {
	if err := f.checkActive(); err != nil {
		return 0, err
	}

	// Use f4double() to get the float value from CodeBase
	floatVal := float64(C.f4double(f.cField))
	return floatVal, nil
}

// AsBool returns true if the float is not zero
func (f *FloatField) AsBool() (bool, error) {
	floatVal, err := f.AsFloat()
	if err != nil {
		return false, err
	}

	return floatVal != 0.0, nil
}

// AsTime cannot convert float to time
func (f *FloatField) AsTime() (time.Time, error) {
	return time.Time{}, NewConversionError("float", "time")
}

// IsNull returns true if the field is null
func (f *FloatField) IsNull() (bool, error) {
	if err := f.checkActive(); err != nil {
		return false, err
	}

	return C.f4null(f.cField) != 0, nil
}

// Field interface methods are inherited from baseField

// String returns a string representation of the float field
func (f *FloatField) String() string {
	floatStr, err := f.AsString()
	if err != nil {
		return "FloatField{name: " + f.Name() + ", error: " + err.Error() + "}"
	}

	return "FloatField{name: " + f.Name() + ", value: " + floatStr + "}"
}
