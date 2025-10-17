package vulpo

import (
	"testing"
	"time"
)

func TestDateField_AsString(t *testing.T) {
	v := &Vulpo{}
	err := v.Open("testdata/fieldtests/dates.dbf")
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer v.Close()

	// Position at first record
	err = v.First()
	if err != nil {
		t.Fatalf("Failed to go to first record: %v", err)
	}

	// Find a date field to test
	fieldDefs := v.FieldDefs()
	if fieldDefs == nil {
		t.Fatal("No field definitions found")
	}

	var dateFieldDef *FieldDef
	for i := 0; i < fieldDefs.Count(); i++ {
		field := fieldDefs.ByIndex(i)
		if field.Type() == FTDate {
			dateFieldDef = field
			break
		}
	}

	if dateFieldDef == nil {
		t.Skip("No date field found in test file")
	}

	// Get the FieldReader for the date field
	fieldReader := v.FieldReader(dateFieldDef.Name())
	if fieldReader == nil {
		t.Fatalf("Failed to get FieldReader for date field: %s", dateFieldDef.Name())
	}

	// Verify it's a DateField
	dateField, ok := fieldReader.(*DateField)
	if !ok {
		t.Fatalf("Expected DateField, got %T", fieldReader)
	}

	// Test AsString method
	dateStr, err := dateField.AsString()
	if err != nil {
		t.Errorf("AsString failed: %v", err)
	} else {
		t.Logf("Date string: %q", dateStr)
	}

	// Test AsTime method
	dateTime, err := dateField.AsTime()
	if err != nil {
		t.Errorf("AsTime failed: %v", err)
	} else {
		t.Logf("Date time: %v", dateTime)
	}

	// Test AsBool method
	isNotNull, err := dateField.AsBool()
	if err != nil {
		t.Errorf("AsBool failed: %v", err)
	} else {
		t.Logf("Date is not null: %v", isNotNull)
	}

	// Test AsInt method (Julian day number)
	dateInt, err := dateField.AsInt()
	if err != nil {
		t.Errorf("AsInt failed: %v", err)
	} else {
		t.Logf("Date as int (Julian): %d", dateInt)
	}

	// Test AsFloat method
	dateFloat, err := dateField.AsFloat()
	if err != nil {
		t.Errorf("AsFloat failed: %v", err)
	} else {
		t.Logf("Date as float: %f", dateFloat)
	}

	// Test IsNull method
	isNull, err := dateField.IsNull()
	if err != nil {
		t.Errorf("IsNull failed: %v", err)
	} else {
		t.Logf("Date is null: %v", isNull)
	}

	// Test field metadata
	t.Logf("Field name: %s", dateField.Name())
	t.Logf("Field type: %s", dateField.Type().String())
	t.Logf("Field size: %d", dateField.Size())
	t.Logf("Field decimals: %d", dateField.Decimals())
}

func TestDateField_AsTime_ParseDate(t *testing.T) {
	// Test the date parsing logic directly
	// This would need to be adapted once we have proper integration
	testDates := []struct {
		dateStr  string
		expected time.Time
		hasError bool
	}{
		{"20231201", time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC), false},
		{"19900101", time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC), false},
		{"        ", time.Time{}, false}, // blank date
		{"", time.Time{}, false},         // empty date
		{"invalid1", time.Time{}, true},  // invalid format
		{"20231301", time.Time{}, true},  // invalid month
		{"20231232", time.Time{}, true},  // invalid day
	}

	for _, test := range testDates {
		t.Run(test.dateStr, func(t *testing.T) {
			// We'd need to create a mock DateField to test this properly
			// For now, this is a framework for testing
			t.Logf("Would test parsing of: %q", test.dateStr)
		})
	}
}
