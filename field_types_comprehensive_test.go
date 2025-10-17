package vulpo

import (
	"testing"
)

//nolint:gocyclo // TODO: refactor this comprehensive test to reduce complexity
func TestAllFieldTypes_Comprehensive(t *testing.T) {
	// Test different field types with their corresponding DBF files
	testCases := []struct {
		file        string
		fieldType   FieldType
		fieldName   string
		description string
	}{
		{"testdata/fieldtests/dates.dbf", FTDate, "dates", "Date field"},
		{"testdata/fieldtests/currencies.dbf", FTCurrency, "", "Currency field"}, // Will find first currency field
		{"testdata/fieldtests/integers.dbf", FTInteger, "", "Integer field"},
		{"testdata/fieldtests/numerics.dbf", FTNumeric, "", "Numeric field"},
		{"testdata/fieldtests/bools.dbf", FTLogical, "", "Logical field"},
		{"testdata/fieldtests/memos.dbf", FTMemo, "", "Memo field"},
		{"testdata/fieldtests/datetimes.dbf", FTDateTime, "", "DateTime field"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			v := &Vulpo{}
			err := v.Open(tc.file)
			if err != nil {
				t.Fatalf("Failed to open %s: %v", tc.file, err)
			}
			defer v.Close()

			// Find a field of the expected type
			fieldDefs := v.FieldDefs()
			if fieldDefs == nil {
				t.Fatal("No field definitions found")
			}

			var targetFieldDef *FieldDef
			if tc.fieldName != "" {
				// Use specific field name
				targetFieldDef = fieldDefs.ByName(tc.fieldName)
			} else {
				// Find first field of the expected type
				for i := 0; i < fieldDefs.Count(); i++ {
					field := fieldDefs.ByIndex(i)
					if field != nil && field.Type() == tc.fieldType {
						targetFieldDef = field
						break
					}
				}
			}

			if targetFieldDef == nil {
				t.Skipf("No %s field found in %s", tc.description, tc.file)
			}

			t.Logf("Testing field: %s (type: %s)", targetFieldDef.Name(), targetFieldDef.Type().String())

			// Get field reader
			fieldReader := v.FieldReader(targetFieldDef.Name())
			if fieldReader == nil {
				t.Fatalf("Failed to get FieldReader for %s", targetFieldDef.Name())
			}

			// Navigate to first record
			err = v.First()
			if err != nil {
				t.Fatalf("Failed to go to first record: %v", err)
			}

			// Test basic FieldReader interface methods
			t.Run("BasicInterface", func(t *testing.T) {
				// Test Value()
				value, err := fieldReader.Value()
				if err != nil {
					t.Errorf("Value() failed: %v", err)
				} else {
					t.Logf("Value(): %v (type: %T)", value, value)
				}

				// Test AsString()
				strVal, err := fieldReader.AsString()
				if err != nil {
					t.Errorf("AsString() failed: %v", err)
				} else {
					t.Logf("AsString(): %q", strVal)
				}

				// Test AsBool()
				boolVal, err := fieldReader.AsBool()
				if err != nil {
					t.Errorf("AsBool() failed: %v", err)
				} else {
					t.Logf("AsBool(): %v", boolVal)
				}

				// Test IsNull()
				isNull, err := fieldReader.IsNull()
				if err != nil {
					t.Errorf("IsNull() failed: %v", err)
				} else {
					t.Logf("IsNull(): %v", isNull)
				}

				// Test FieldDef()
				fieldDef := fieldReader.FieldDef()
				if fieldDef == nil {
					t.Error("FieldDef() returned nil")
				} else {
					t.Logf("FieldDef(): name=%s, type=%s, size=%d, decimals=%d",
						fieldDef.Name(), fieldDef.Type().String(), fieldDef.Size(), fieldDef.Decimals())
				}
			})

			// Test type-specific conversions (where applicable)
			t.Run("TypeSpecificConversions", func(t *testing.T) {
				switch tc.fieldType {
				case FTInteger, FTNumeric, FTFloat, FTDouble, FTCurrency:
					// Test numeric conversions
					intVal, err := fieldReader.AsInt()
					if err != nil {
						t.Logf("AsInt() failed (expected for some types): %v", err)
					} else {
						t.Logf("AsInt(): %d", intVal)
					}

					floatVal, err := fieldReader.AsFloat()
					if err != nil {
						t.Logf("AsFloat() failed (expected for some types): %v", err)
					} else {
						t.Logf("AsFloat(): %f", floatVal)
					}

				case FTDate, FTDateTime:
					// Test time conversions
					timeVal, err := fieldReader.AsTime()
					if err != nil {
						t.Errorf("AsTime() failed for date/datetime field: %v", err)
					} else {
						t.Logf("AsTime(): %v", timeVal)
					}

				case FTMemo, FTCharacter:
					// Test string-based operations
					strVal, err := fieldReader.AsString()
					if err != nil {
						t.Errorf("AsString() failed for text field: %v", err)
					} else {
						t.Logf("AsString() length: %d", len(strVal))
					}
				}
			})
		})
	}
}

//nolint:gocyclo // TODO: refactor this test function to reduce complexity
func TestFieldFactory_TypeDetection(t *testing.T) {
	// Test that the field factory correctly identifies and creates the right field types
	testFiles := []string{
		"testdata/fieldtests/dates.dbf",
		"testdata/fieldtests/currencies.dbf",
		"testdata/fieldtests/integers.dbf",
		"testdata/fieldtests/numerics.dbf",
		"testdata/fieldtests/bools.dbf",
		"testdata/fieldtests/memos.dbf",
		"testdata/fieldtests/datetimes.dbf",
	}

	for _, file := range testFiles {
		t.Run(file, func(t *testing.T) {
			v := &Vulpo{}
			err := v.Open(file)
			if err != nil {
				t.Fatalf("Failed to open %s: %v", file, err)
			}
			defer v.Close()

			fieldDefs := v.FieldDefs()
			if fieldDefs == nil {
				t.Fatal("No field definitions found")
			}

			for i := 0; i < fieldDefs.Count(); i++ {
				fieldDef := fieldDefs.ByIndex(i)
				if fieldDef == nil {
					continue
				}

				fieldReader := v.FieldReader(fieldDef.Name())
				if fieldReader == nil {
					t.Errorf("Failed to create FieldReader for field %s (type: %s)",
						fieldDef.Name(), fieldDef.Type().String())
					continue
				}

				// Verify the field reader type matches the expected type based on FieldDef
				expectedType := fieldDef.Type()
				switch expectedType {
				case FTDate:
					if _, ok := fieldReader.(*DateField); !ok {
						t.Errorf("Expected DateField for field %s, got %T", fieldDef.Name(), fieldReader)
					}
				case FTDateTime:
					if _, ok := fieldReader.(*DateTimeField); !ok {
						t.Errorf("Expected DateTimeField for field %s, got %T", fieldDef.Name(), fieldReader)
					}
				case FTCurrency:
					if _, ok := fieldReader.(*CurrencyField); !ok {
						t.Errorf("Expected CurrencyField for field %s, got %T", fieldDef.Name(), fieldReader)
					}
				case FTFloat:
					if _, ok := fieldReader.(*FloatField); !ok {
						t.Errorf("Expected FloatField for field %s, got %T", fieldDef.Name(), fieldReader)
					}
				case FTDouble:
					if _, ok := fieldReader.(*DoubleField); !ok {
						t.Errorf("Expected DoubleField for field %s, got %T", fieldDef.Name(), fieldReader)
					}
				case FTMemo:
					if _, ok := fieldReader.(*MemoField); !ok {
						t.Errorf("Expected MemoField for field %s, got %T", fieldDef.Name(), fieldReader)
					}
				case FTInteger:
					if _, ok := fieldReader.(*IntegerField); !ok {
						t.Errorf("Expected IntegerField for field %s, got %T", fieldDef.Name(), fieldReader)
					}
				case FTNumeric:
					if _, ok := fieldReader.(*NumericField); !ok {
						t.Errorf("Expected NumericField for field %s, got %T", fieldDef.Name(), fieldReader)
					}
				case FTLogical:
					if _, ok := fieldReader.(*LogicalField); !ok {
						t.Errorf("Expected LogicalField for field %s, got %T", fieldDef.Name(), fieldReader)
					}
				case FTCharacter:
					if _, ok := fieldReader.(*StringField); !ok {
						t.Errorf("Expected StringField for field %s, got %T", fieldDef.Name(), fieldReader)
					}
				}

				t.Logf("Field %s (type %s): correctly created %T",
					fieldDef.Name(), expectedType.String(), fieldReader)
			}
		})
	}
}
