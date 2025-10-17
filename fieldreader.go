package vulpo

import "time"

// FieldReader defines the interface that all field types must implement.
// It provides both native value access, type conversion capabilities, and
// field definition access methods.
//
// Note: In Vulpo v2.0, the Field interface extends FieldReader to provide
// the unified field access API. Users should prefer the Field interface
// accessed via FieldByName() or Field() methods for new code.
//
// This interface includes all methods needed for both field definition
// access and value reading from the current record.
type FieldReader interface {
	// Value returns the field's native value in its appropriate Go type
	Value() (interface{}, error)

	// AsString returns the field value as a string
	AsString() (string, error)

	// AsInt returns the field value as an integer
	AsInt() (int, error)

	// AsFloat returns the field value as a float64
	AsFloat() (float64, error)

	// AsBool returns the field value as a boolean
	AsBool() (bool, error)

	// AsTime returns the field value as a time.Time
	AsTime() (time.Time, error)

	// IsNull returns true if the field contains a null value
	IsNull() (bool, error)

	// Field definition access methods
	Name() string
	Type() FieldType
	Size() uint8
	Decimals() uint8
	IsSystem() bool
	IsNullable() bool
	IsBinary() bool

	// FieldDef returns the field definition (for backward compatibility)
	FieldDef() *FieldDef
}

// baseField provides common functionality for all field types
type baseField struct {
	def  *FieldDef
	data *Vulpo
}

// FieldDef returns the field definition (for backward compatibility)
func (bf *baseField) FieldDef() *FieldDef {
	return bf.def
}

// Field interface implementation - definition methods
func (bf *baseField) Name() string {
	return bf.def.Name()
}

func (bf *baseField) Type() FieldType {
	return bf.def.Type()
}

func (bf *baseField) Size() uint8 {
	return bf.def.Size()
}

func (bf *baseField) Decimals() uint8 {
	return bf.def.Decimals()
}

func (bf *baseField) IsSystem() bool {
	return bf.def.IsSystem()
}

func (bf *baseField) IsNullable() bool {
	return bf.def.IsNullable()
}

func (bf *baseField) IsBinary() bool {
	return bf.def.IsBinary()
}

// checkActive verifies the database is active and positioned at a valid record
func (bf *baseField) checkActive() error {
	if bf.data == nil || !bf.data.Active() {
		return NewError("database not open")
	}

	if bf.data.BOF() {
		return NewError("positioned at beginning of file (BOF)")
	}

	if bf.data.EOF() {
		return NewError("positioned at end of file (EOF)")
	}

	return nil
}

// NewConversionError creates a standardized conversion error
func NewConversionError(fromType, toType string) error {
	return NewErrorf("cannot convert %s to %s", fromType, toType)
}
