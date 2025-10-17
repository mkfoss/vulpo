package vulpo

/*
#include "d4all.h"
*/
import "C"
import (
	"encoding/binary"
	"fmt"
	"time"
	"unsafe"
)

// DateTimeField represents a DBF datetime field (type 'T')
// DateTime fields store both date and time information
type DateTimeField struct {
	baseField
	cField *C.FIELD4
}

// newDateTimeField creates a new DateTimeField instance
func newDateTimeField(field *C.FIELD4, data *Vulpo, def *FieldDef) *DateTimeField {
	return &DateTimeField{
		baseField: baseField{
			def:  def,
			data: data,
		},
		cField: field,
	}
}

// Value returns the field value as time.Time
func (f *DateTimeField) Value() (interface{}, error) {
	return f.AsTime()
}

// AsString returns the datetime as an ISO 8601 formatted string
func (f *DateTimeField) AsString() (string, error) {
	if err := f.checkActive(); err != nil {
		return "", err
	}

	dateTime, err := f.AsTime()
	if err != nil {
		return "", err
	}

	if dateTime.IsZero() {
		return "", nil
	}

	// Return ISO 8601 format
	return dateTime.Format(time.RFC3339), nil
}

// AsInt returns the datetime as Unix timestamp (seconds since epoch)
func (f *DateTimeField) AsInt() (int, error) {
	dateTime, err := f.AsTime()
	if err != nil {
		return 0, err
	}

	if dateTime.IsZero() {
		return 0, nil
	}

	return int(dateTime.Unix()), nil
}

// AsFloat returns the datetime as Unix timestamp with fractional seconds
func (f *DateTimeField) AsFloat() (float64, error) {
	dateTime, err := f.AsTime()
	if err != nil {
		return 0, err
	}

	if dateTime.IsZero() {
		return 0, nil
	}

	// Unix timestamp with nanosecond precision as fractional part
	return float64(dateTime.UnixNano()) / 1e9, nil
}

// AsBool returns true if the datetime is not zero/empty
func (f *DateTimeField) AsBool() (bool, error) {
	dateTime, err := f.AsTime()
	if err != nil {
		return false, err
	}

	return !dateTime.IsZero(), nil
}

// AsTime returns the field value as a time.Time
func (f *DateTimeField) AsTime() (time.Time, error) {
	if err := f.checkActive(); err != nil {
		return time.Time{}, err
	}

	// Get the raw binary data
	ptr := C.f4ptr(f.cField)
	if ptr == nil {
		return time.Time{}, nil // Return zero time for null field
	}

	length := C.f4len(f.cField)
	if length == 8 {
		// DateTime fields are stored as 8 bytes:
		// First 4 bytes: Julian day number (little-endian)
		// Last 4 bytes: milliseconds since midnight (little-endian)
		bytes := C.GoBytes(unsafe.Pointer(ptr), 8)

		// Extract Julian day and milliseconds using proper format
		jdays := binary.LittleEndian.Uint32(bytes[:4])
		jmsec := binary.LittleEndian.Uint32(bytes[4:])
		jsec := jmsec / 1000

		// Check for null/zero datetime
		if jdays == 0 && jmsec == 0 {
			return time.Time{}, nil
		}

		// Convert Julian day to Gregorian date using proper algorithm
		y, m, d := JulianToYMD(int(jdays))

		// Return the proper datetime
		return time.Date(y, time.Month(m), d, 0, 0, int(jsec), 0, time.UTC), nil
	}

	// Try parsing as string instead
	dateTimeStr := C.GoStringN(ptr, C.int(length))

	// Handle empty/blank datetime
	if dateTimeStr == "" || len(dateTimeStr) == 0 {
		return time.Time{}, nil
	}

	// Check for blank field (all spaces or null bytes)
	allBlanks := true
	for _, ch := range dateTimeStr {
		if ch != ' ' && ch != '\x00' {
			allBlanks = false
			break
		}
	}
	if allBlanks {
		return time.Time{}, nil
	}

	// Try to parse various datetime formats commonly used in DBF files
	formats := []string{
		"20060102 15:04:05",   // YYYYMMDD HH:MM:SS
		"20060102T15:04:05",   // YYYYMMDDTHH:MM:SS
		"2006-01-02 15:04:05", // YYYY-MM-DD HH:MM:SS
		"2006-01-02T15:04:05", // YYYY-MM-DDTHH:MM:SS
		"20060102",            // YYYYMMDD (date only)
		time.RFC3339,          // ISO 8601
		time.RFC822,           // RFC 822
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateTimeStr); err == nil {
			return t, nil
		}
	}

	// If we can't parse as string either, return zero time
	return time.Time{}, nil
}

// Raw returns the raw bytes of the datetime field
func (f *DateTimeField) Raw() []byte {
	if err := f.checkActive(); err != nil {
		return nil
	}

	ptr := C.f4ptr(f.cField)
	if ptr == nil {
		return nil
	}

	length := C.f4len(f.cField)
	return C.GoBytes(unsafe.Pointer(ptr), C.int(length))
}

// IsNull returns true if the field is null
func (f *DateTimeField) IsNull() (bool, error) {
	if err := f.checkActive(); err != nil {
		return false, err
	}

	return C.f4null(f.cField) != 0, nil
}

// Field interface methods are inherited from baseField

// String returns a string representation of the datetime field
func (f *DateTimeField) String() string {
	dateTimeStr, err := f.AsString()
	if err != nil {
		return fmt.Sprintf("DateTimeField{name: %s, error: %v}", f.Name(), err)
	}

	if dateTimeStr == "" {
		return fmt.Sprintf("DateTimeField{name: %s, value: <empty>}", f.Name())
	}

	return fmt.Sprintf("DateTimeField{name: %s, value: %s}", f.Name(), dateTimeStr)
}
