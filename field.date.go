package vulpo

/*
#include "mkfdbflib/d4all.h"
#include <stdlib.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

const blankDateValue = "        "

// DateField represents a DBF date field (type 'D')
type DateField struct {
	baseField
	cField *C.FIELD4
}

// newDateField creates a new DateField instance
func newDateField(field *C.FIELD4, data *Vulpo, def *FieldDef) *DateField {
	return &DateField{
		baseField: baseField{
			def:  def,
			data: data,
		},
		cField: field,
	}
}

// Value returns the field value as time.Time, or error if conversion fails
func (f *DateField) Value() (interface{}, error) {
	return f.AsTime()
}

// AsString returns the date as a string in YYYYMMDD format
func (f *DateField) AsString() (string, error) {
	if err := f.checkActive(); err != nil {
		return "", err
	}

	ptr := C.f4ptr(f.cField)
	if ptr == nil {
		return "", errors.New("failed to get field pointer")
	}

	// DBF date fields are stored as 8-character strings in YYYYMMDD format
	length := C.f4len(f.cField)
	if length != 8 {
		return "", fmt.Errorf("invalid date field length: %d", length)
	}

	dateStr := C.GoStringN(ptr, 8)

	// Check if the date is blank (all spaces)
	if dateStr == blankDateValue || dateStr == "" {
		return "", nil
	}

	return dateStr, nil
}

// AsInt returns the date as a long integer (Julian day number)
func (f *DateField) AsInt() (int, error) {
	if err := f.checkActive(); err != nil {
		return 0, err
	}

	ptr := C.f4ptr(f.cField)
	if ptr == nil {
		return 0, errors.New("failed to get field pointer")
	}

	// Check if the field is blank
	dateStr := C.GoStringN(ptr, 8)
	if dateStr == blankDateValue || dateStr == "" {
		return 0, nil
	}

	// Use CodeBase date4long function to convert to long integer
	longVal := C.date4long(ptr)
	if longVal == 0 {
		// Could be a valid date (January 1, 1900) or an error
		// We need to check if the input was valid
		if dateStr != "19000101" {
			return 0, fmt.Errorf("invalid date format: %s", dateStr)
		}
	}

	return int(longVal), nil
}

// AsFloat returns the date as a float64 (Julian day number)
func (f *DateField) AsFloat() (float64, error) {
	intVal, err := f.AsInt()
	if err != nil {
		return 0, err
	}
	return float64(intVal), nil
}

// AsBool returns true if the date is not blank/empty
func (f *DateField) AsBool() (bool, error) {
	dateStr, err := f.AsString()
	if err != nil {
		return false, err
	}

	return dateStr != "" && dateStr != blankDateValue, nil
}

// AsTime returns the date as a time.Time value
func (f *DateField) AsTime() (time.Time, error) {
	dateStr, err := f.AsString()
	if err != nil {
		return time.Time{}, err
	}

	// Handle blank dates
	if dateStr == "" || dateStr == blankDateValue {
		return time.Time{}, nil
	}

	// Parse YYYYMMDD format
	if len(dateStr) != 8 {
		return time.Time{}, fmt.Errorf("invalid date format: expected YYYYMMDD, got %s", dateStr)
	}

	year, err := strconv.Atoi(dateStr[0:4])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid year in date %s: %v", dateStr, err)
	}

	month, err := strconv.Atoi(dateStr[4:6])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid month in date %s: %v", dateStr, err)
	}

	day, err := strconv.Atoi(dateStr[6:8])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid day in date %s: %v", dateStr, err)
	}

	// Validate ranges
	if month < 1 || month > 12 {
		return time.Time{}, fmt.Errorf("invalid month %d in date %s", month, dateStr)
	}
	if day < 1 || day > 31 {
		return time.Time{}, fmt.Errorf("invalid day %d in date %s", day, dateStr)
	}

	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC), nil
}

// IsNull returns true if the date field is blank
func (f *DateField) IsNull() (bool, error) {
	if err := f.checkActive(); err != nil {
		return false, err
	}

	return C.f4null(f.cField) != 0, nil
}

// Field interface methods are inherited from baseField

// String returns a string representation of the date field
func (f *DateField) String() string {
	dateStr, err := f.AsString()
	if err != nil {
		return fmt.Sprintf("DateField{name: %s, error: %v}", f.Name(), err)
	}

	if dateStr == "" {
		return fmt.Sprintf("DateField{name: %s, value: <blank>}", f.Name())
	}

	// Try to format as a more readable date
	if t, err := f.AsTime(); err == nil && !t.IsZero() {
		return fmt.Sprintf("DateField{name: %s, value: %s}", f.Name(), t.Format("2006-01-02"))
	}

	return fmt.Sprintf("DateField{name: %s, value: %s}", f.Name(), dateStr)
}
