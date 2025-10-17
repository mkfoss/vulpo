package vulpo

import (
	"testing"
	"time"
)

const blankDateStr = "        "

func TestDateField_MultipleRecords(t *testing.T) {
	v := &Vulpo{}
	err := v.Open("testdata/fieldtests/dates.dbf")
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer v.Close()

	// Get the date field reader
	dateField := v.FieldReader("dates")
	if dateField == nil {
		t.Fatal("Failed to get date field reader")
	}

	// Navigate through records and test different date values
	err = v.First()
	if err != nil {
		t.Fatalf("Failed to go to first record: %v", err)
	}

	recordCount := 0
	for !v.EOF() {
		recordCount++

		dateStr, err := dateField.AsString()
		if err != nil {
			t.Errorf("Record %d: AsString failed: %v", recordCount, err)
		}

		dateTime, err := dateField.AsTime()
		if err != nil {
			t.Errorf("Record %d: AsTime failed: %v", recordCount, err)
		}

		isNull, err := dateField.IsNull()
		if err != nil {
			t.Errorf("Record %d: IsNull failed: %v", recordCount, err)
		}

		t.Logf("Record %d: dateStr=%q, dateTime=%v, isNull=%v",
			recordCount, dateStr, dateTime, isNull)

		// Validate consistency between AsString and AsTime
		if dateStr != "" && dateStr != blankDateStr {
			// Should have a valid time
			if dateTime.IsZero() {
				t.Errorf("Record %d: Non-empty date string %q but zero time", recordCount, dateStr)
			}
		} else {
			// Should have zero time for empty dates
			if !dateTime.IsZero() {
				t.Errorf("Record %d: Empty date string %q but non-zero time %v", recordCount, dateStr, dateTime)
			}
		}

		// Move to next record
		err = v.Next()
		if err != nil {
			// Check if we've reached EOF (error code 3 is normal EOF)
			if v.EOF() {
				t.Logf("Reached EOF after %d records", recordCount)
				break
			}
			t.Errorf("Record %d: Failed to move to next: %v", recordCount, err)
			break
		}
	}

	t.Logf("Tested %d records", recordCount)
}

func TestDateField_ValueMethod(t *testing.T) {
	v := &Vulpo{}
	err := v.Open("testdata/fieldtests/dates.dbf")
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer v.Close()

	err = v.First()
	if err != nil {
		t.Fatalf("Failed to go to first record: %v", err)
	}

	dateField := v.FieldReader("dates")
	if dateField == nil {
		t.Fatal("Failed to get date field reader")
	}

	// Test Value() method (should return time.Time)
	value, err := dateField.Value()
	if err != nil {
		t.Fatalf("Value() failed: %v", err)
	}

	timeValue, ok := value.(time.Time)
	if !ok {
		t.Fatalf("Value() returned %T, expected time.Time", value)
	}

	// Compare with AsTime()
	asTime, err := dateField.AsTime()
	if err != nil {
		t.Fatalf("AsTime() failed: %v", err)
	}

	if !timeValue.Equal(asTime) {
		t.Errorf("Value() returned %v, AsTime() returned %v", timeValue, asTime)
	}

	t.Logf("Value() correctly returned: %v", timeValue)
}

func TestDateField_EdgeCaseDates(t *testing.T) {
	// This test would ideally use a DBF file with various edge case dates
	// For now, we test the parsing logic with some known edge cases

	testCases := []struct {
		name     string
		dateStr  string
		expected time.Time
		hasError bool
	}{
		{
			name:     "Valid date",
			dateStr:  "19320205",
			expected: time.Date(1932, 2, 5, 0, 0, 0, 0, time.UTC),
			hasError: false,
		},
		{
			name:     "Leap year date",
			dateStr:  "20000229",
			expected: time.Date(2000, 2, 29, 0, 0, 0, 0, time.UTC),
			hasError: false,
		},
		{
			name:     "Year 1900 (not a leap year)",
			dateStr:  "19000228",
			expected: time.Date(1900, 2, 28, 0, 0, 0, 0, time.UTC),
			hasError: false,
		},
		{
			name:     "Blank date (spaces)",
			dateStr:  blankDateStr,
			expected: time.Time{},
			hasError: false,
		},
		{
			name:     "Empty date",
			dateStr:  "",
			expected: time.Time{},
			hasError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We'd need to create a test file with these dates or use a more direct test
			// For now, just log what we would test
			t.Logf("Would test: %s -> %v (error expected: %v)", tc.dateStr, tc.expected, tc.hasError)
		})
	}
}

func TestDateField_FieldDefIntegration(t *testing.T) {
	v := &Vulpo{}
	err := v.Open("testdata/fieldtests/dates.dbf")
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer v.Close()

	dateField := v.FieldReader("dates")
	if dateField == nil {
		t.Fatal("Failed to get date field reader")
	}

	// Test FieldDef() method
	fieldDef := dateField.FieldDef()
	if fieldDef == nil {
		t.Fatal("FieldDef() returned nil")
	}

	// Verify field definition properties
	if fieldDef.Name() != "dates" {
		t.Errorf("Expected field name 'dates', got '%s'", fieldDef.Name())
	}

	if fieldDef.Type() != FTDate {
		t.Errorf("Expected field type FTDate, got %v", fieldDef.Type())
	}

	if fieldDef.Size() != 8 {
		t.Errorf("Expected field size 8, got %d", fieldDef.Size())
	}

	if fieldDef.Decimals() != 0 {
		t.Errorf("Expected field decimals 0, got %d", fieldDef.Decimals())
	}

	t.Logf("FieldDef integration test passed")
}
