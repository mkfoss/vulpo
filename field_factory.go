package vulpo

/*
#include "d4all.h"
*/
import "C"

// FieldReader creates a FieldReader instance for the specified field name.
// DEPRECATED: Use FieldByName() instead, which returns a Field interface that includes both
// definition and reading capabilities. This method is kept for backward compatibility.
//
// Returns nil if the field is not found or the database is not open.
func (v *Vulpo) FieldReader(name string) FieldReader {
	if !v.Active() {
		return nil
	}

	// Simply return the field from the new Fields collection
	// Since Field interface extends FieldReader, this works
	return v.FieldByName(name)
}

// FieldReaderByIndex creates a FieldReader instance for the specified field index (0-based).
// DEPRECATED: Use Field(index) instead, which returns a Field interface that includes both
// definition and reading capabilities. This method is kept for backward compatibility.
//
// Returns nil if the index is out of bounds or the database is not open.
func (v *Vulpo) FieldReaderByIndex(index int) FieldReader {
	if !v.Active() {
		return nil
	}

	// Simply return the field from the new Fields collection
	// Since Field interface extends FieldReader, this works
	return v.Field(index)
}

// createFieldReader creates the appropriate FieldReader implementation based on field type
func (v *Vulpo) createFieldReader(cField *C.FIELD4, fieldDef *FieldDef) FieldReader {
	switch fieldDef.Type() {
	case FTCharacter:
		return newStringField(cField, v, fieldDef)
	case FTInteger:
		return &IntegerField{
			baseField: baseField{
				def:  fieldDef,
				data: v,
			},
			cField: cField,
		}
	case FTNumeric:
		return &NumericField{
			baseField: baseField{
				def:  fieldDef,
				data: v,
			},
			cField: cField,
		}
	case FTLogical:
		return &LogicalField{
			baseField: baseField{
				def:  fieldDef,
				data: v,
			},
			cField: cField,
		}
	case FTDate:
		return newDateField(cField, v, fieldDef)
	case FTDateTime:
		return newDateTimeField(cField, v, fieldDef)
	case FTCurrency:
		return newCurrencyField(cField, v, fieldDef)
	case FTFloat:
		return newFloatField(cField, v, fieldDef)
	case FTDouble:
		return newDoubleField(cField, v, fieldDef)
	case FTMemo:
		return newMemoField(cField, v, fieldDef)
	default:
		// For unsupported field types, return a StringField as fallback
		return newStringField(cField, v, fieldDef)
	}
}
