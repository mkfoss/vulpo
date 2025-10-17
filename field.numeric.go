package vulpo

/*
#include "d4all.h"
*/
import "C"
import (
	"strconv"
	"time"
)

// NumericField handles numeric fields (stored as double in DBF)
type NumericField struct {
	baseField
	cField *C.FIELD4
}

// Value returns the field's numeric value as float64
func (nf *NumericField) Value() (interface{}, error) {
	if err := nf.checkActive(); err != nil {
		return nil, err
	}

	// Get double value using f4double()
	val := float64(C.f4double(nf.cField))
	return val, nil
}

// AsString returns the field value as a string
func (nf *NumericField) AsString() (string, error) {
	val, err := nf.Value()
	if err != nil {
		return "", err
	}

	floatVal := val.(float64)

	// Check if the field has decimal places defined
	decimals := int(nf.def.Decimals())
	if decimals > 0 {
		return strconv.FormatFloat(floatVal, 'f', decimals, 64), nil
	}

	// If no decimals and value is a whole number, format as integer
	if floatVal == float64(int64(floatVal)) {
		return strconv.FormatInt(int64(floatVal), 10), nil
	}

	return strconv.FormatFloat(floatVal, 'g', -1, 64), nil
}

// AsInt returns the field value as an integer (truncated)
func (nf *NumericField) AsInt() (int, error) {
	val, err := nf.Value()
	if err != nil {
		return 0, err
	}
	return int(val.(float64)), nil
}

// AsFloat returns the field value as a float
func (nf *NumericField) AsFloat() (float64, error) {
	val, err := nf.Value()
	if err != nil {
		return 0, err
	}
	return val.(float64), nil
}

// AsBool returns the field value as a boolean (0 = false, non-zero = true)
func (nf *NumericField) AsBool() (bool, error) {
	val, err := nf.Value()
	if err != nil {
		return false, err
	}
	return val.(float64) != 0, nil
}

// AsTime cannot convert numeric to time
func (nf *NumericField) AsTime() (time.Time, error) {
	return time.Time{}, NewConversionError("numeric", "time")
}

// IsNull returns true if the field is null
func (nf *NumericField) IsNull() (bool, error) {
	if err := nf.checkActive(); err != nil {
		return false, err
	}

	return C.f4null(nf.cField) != 0, nil
}
