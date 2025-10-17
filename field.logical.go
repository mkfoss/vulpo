package vulpo

/*
#include "d4all.h"
*/
import "C"
import (
	"time"
)

// LogicalField handles logical/boolean fields
type LogicalField struct {
	baseField
	cField *C.FIELD4
}

// Value returns the field's boolean value
func (lf *LogicalField) Value() (interface{}, error) {
	if err := lf.checkActive(); err != nil {
		return nil, err
	}

	// Get boolean value using f4true() (returns 1 for true, 0 for false)
	val := C.f4true(lf.cField) != 0
	return val, nil
}

// AsString returns the field value as a string ("T"/"F")
func (lf *LogicalField) AsString() (string, error) {
	val, err := lf.Value()
	if err != nil {
		return "", err
	}

	if val.(bool) {
		return "T", nil
	}
	return "F", nil
}

// AsInt returns the field value as an integer (1 for true, 0 for false)
func (lf *LogicalField) AsInt() (int, error) {
	val, err := lf.Value()
	if err != nil {
		return 0, err
	}

	if val.(bool) {
		return 1, nil
	}
	return 0, nil
}

// AsFloat returns the field value as a float (1.0 for true, 0.0 for false)
func (lf *LogicalField) AsFloat() (float64, error) {
	val, err := lf.Value()
	if err != nil {
		return 0, err
	}

	if val.(bool) {
		return 1.0, nil
	}
	return 0.0, nil
}

// AsBool returns the field value as a boolean
func (lf *LogicalField) AsBool() (bool, error) {
	val, err := lf.Value()
	if err != nil {
		return false, err
	}
	return val.(bool), nil
}

// AsTime cannot convert logical to time
func (lf *LogicalField) AsTime() (time.Time, error) {
	return time.Time{}, NewConversionError("logical", "time")
}

// IsNull returns true if the field is null
func (lf *LogicalField) IsNull() (bool, error) {
	if err := lf.checkActive(); err != nil {
		return false, err
	}

	return C.f4null(lf.cField) != 0, nil
}
