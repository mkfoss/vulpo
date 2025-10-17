package vulpo

import (
	"testing"
)

func TestStringField_BasicFunctionality(t *testing.T) {
	v := &Vulpo{}

	// Test with a file that should have character fields
	err := v.Open("testdata/idcharsdate.dbf")
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer v.Close()

	// Find a character field
	fieldDefs := v.FieldDefs()
	if fieldDefs == nil {
		t.Fatal("No field definitions found")
	}

	var charField *FieldDef
	for i := 0; i < fieldDefs.Count(); i++ {
		field := fieldDefs.ByIndex(i)
		if field != nil && field.Type() == FTCharacter {
			charField = field
			break
		}
	}

	if charField == nil {
		t.Skip("No character field found in test file")
	}

	// Get field reader
	fieldReader := v.FieldReader(charField.Name())
	if fieldReader == nil {
		t.Fatalf("Failed to get FieldReader for character field: %s", charField.Name())
	}

	// Verify it's a StringField
	stringField, ok := fieldReader.(*StringField)
	if !ok {
		t.Fatalf("Expected StringField, got %T", fieldReader)
	}

	// Move to first record
	err = v.First()
	if err != nil {
		t.Fatalf("Failed to go to first record: %v", err)
	}

	// Test basic interface methods
	t.Logf("Testing StringField: %s", charField.Name())

	// Test Value()
	value, err := stringField.Value()
	if err != nil {
		t.Errorf("Value() failed: %v", err)
	} else {
		t.Logf("Value(): %v (type: %T)", value, value)
	}

	// Test AsString()
	strVal, err := stringField.AsString()
	if err != nil {
		t.Errorf("AsString() failed: %v", err)
	} else {
		t.Logf("AsString(): %q", strVal)
	}

	// Test IsNull()
	isNull, err := stringField.IsNull()
	if err != nil {
		t.Errorf("IsNull() failed: %v", err)
	} else {
		t.Logf("IsNull(): %v", isNull)
	}

	// Test FieldDef()
	fieldDef := stringField.FieldDef()
	if fieldDef == nil {
		t.Error("FieldDef() returned nil")
	} else {
		t.Logf("FieldDef(): name=%s, type=%s, size=%d",
			fieldDef.Name(), fieldDef.Type().String(), fieldDef.Size())
	}

	// Test metadata methods
	t.Logf("Name(): %s", stringField.Name())
	t.Logf("Type(): %s", stringField.Type().String())
	t.Logf("Size(): %d", stringField.Size())
	t.Logf("Decimals(): %d", stringField.Decimals())
}
