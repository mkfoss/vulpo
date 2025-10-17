package vulpo

/*
#include "d4all.h"
*/
import "C"
import (
	"fmt"
	"time"
)

// CurrencyField represents a DBF currency field (type 'Y')
// Currency fields are 8-byte fixed-point values with 4 decimal places
type CurrencyField struct {
	baseField
	cField *C.FIELD4
}

// newCurrencyField creates a new CurrencyField instance
func newCurrencyField(field *C.FIELD4, data *Vulpo, def *FieldDef) *CurrencyField {
	return &CurrencyField{
		baseField: baseField{
			def:  def,
			data: data,
		},
		cField: field,
	}
}

// Value returns the field value as float64 (monetary amount)
func (f *CurrencyField) Value() (interface{}, error) {
	return f.AsFloat()
}

// AsString returns the currency as a formatted string
func (f *CurrencyField) AsString() (string, error) {
	if err := f.checkActive(); err != nil {
		return "", err
	}

	currencyVal, err := f.AsFloat()
	if err != nil {
		return "", err
	}

	// Format currency with 4 decimal places (standard for currency fields)
	return fmt.Sprintf("%.4f", currencyVal), nil
}

// AsInt returns the currency as an integer (truncated, losing fractional part)
func (f *CurrencyField) AsInt() (int, error) {
	currencyVal, err := f.AsFloat()
	if err != nil {
		return 0, err
	}

	return int(currencyVal), nil
}

// AsFloat returns the field value as a float64
func (f *CurrencyField) AsFloat() (float64, error) {
	if err := f.checkActive(); err != nil {
		return 0, err
	}

	// Currency fields in DBF are stored as 8-byte fixed-point values
	// Use f4double() to get the currency value from CodeBase
	currencyVal := float64(C.f4double(f.cField))
	return currencyVal, nil
}

// AsBool returns true if the currency value is not zero
func (f *CurrencyField) AsBool() (bool, error) {
	currencyVal, err := f.AsFloat()
	if err != nil {
		return false, err
	}

	return currencyVal != 0.0, nil
}

// AsTime cannot convert currency to time
func (f *CurrencyField) AsTime() (time.Time, error) {
	return time.Time{}, NewConversionError("currency", "time")
}

// IsNull returns true if the field is null
func (f *CurrencyField) IsNull() (bool, error) {
	if err := f.checkActive(); err != nil {
		return false, err
	}

	return C.f4null(f.cField) != 0, nil
}

// Field interface methods are inherited from baseField

// String returns a string representation of the currency field
func (f *CurrencyField) String() string {
	currencyStr, err := f.AsString()
	if err != nil {
		return "CurrencyField{name: " + f.Name() + ", error: " + err.Error() + "}"
	}

	return "CurrencyField{name: " + f.Name() + ", value: " + currencyStr + "}"
}

// AsCents returns the currency value as integer cents (multiplied by 10000)
// This is useful for precise monetary calculations
func (f *CurrencyField) AsCents() (int64, error) {
	currencyVal, err := f.AsFloat()
	if err != nil {
		return 0, err
	}

	// Convert to cents (multiply by 10000 since currency has 4 decimal places)
	return int64(currencyVal * 10000), nil
}

// FromCents sets a currency value from integer cents
// Note: This would be used for writing, but this is a read-only implementation
func (f *CurrencyField) FromCents(cents int64) float64 {
	return float64(cents) / 10000.0
}
