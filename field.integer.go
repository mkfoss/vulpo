package vulpo

/*
#include "d4all.h"
*/
import "C"
import (
	"strconv"
	"time"
)

// IntegerField handles integer fields
type IntegerField struct {
	baseField
	cField *C.FIELD4
}

// Value returns the field's integer value
func (intf *IntegerField) Value() (interface{}, error) {
	if err := intf.checkActive(); err != nil {
		return nil, err
	}

	// Get integer value using f4int()
	val := int(C.f4int(intf.cField))
	return val, nil
}

// AsString returns the field value as a string
func (intf *IntegerField) AsString() (string, error) {
	val, err := intf.Value()
	if err != nil {
		return "", err
	}
	return strconv.Itoa(val.(int)), nil
}

// AsInt returns the field value as an integer
func (intf *IntegerField) AsInt() (int, error) {
	val, err := intf.Value()
	if err != nil {
		return 0, err
	}
	return val.(int), nil
}

// AsFloat returns the field value as a float
func (intf *IntegerField) AsFloat() (float64, error) {
	val, err := intf.Value()
	if err != nil {
		return 0, err
	}
	return float64(val.(int)), nil
}

// AsBool returns the field value as a boolean (0 = false, non-zero = true)
func (intf *IntegerField) AsBool() (bool, error) {
	val, err := intf.Value()
	if err != nil {
		return false, err
	}
	return val.(int) != 0, nil
}

// AsTime cannot convert integer to time
func (intf *IntegerField) AsTime() (time.Time, error) {
	return time.Time{}, NewConversionError("integer", "time")
}

// IsNull returns true if the field is null
func (intf *IntegerField) IsNull() (bool, error) {
	if err := intf.checkActive(); err != nil {
		return false, err
	}

	return C.f4null(intf.cField) != 0, nil
}
