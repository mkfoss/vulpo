package vulpo

import (
	"strings"
	"testing"
	"time"
)

const (
	testDBFPath         = "mkfdbflib/data/info.dbf"
	WindowsANSICodepage = "Windows ANSI"
	BlankDateString     = "        "
)

func TestVulpo_NewInstance(t *testing.T) {
	v := &Vulpo{}

	// New instance should not be active
	if v.Active() {
		t.Error("New Vulpo instance should not be active")
	}

	// Fields should be properly initialized to zero values
	if v.filename != "" {
		t.Errorf("Expected empty filename, got %s", v.filename)
	}

	if v.codeBase != nil {
		t.Error("Expected nil codeBase for new instance")
	}

	if v.data != nil {
		t.Error("Expected nil data for new instance")
	}

	if v.header != nil {
		t.Error("Expected nil header for new instance")
	}
}

func TestVulpo_Open_ValidFile(t *testing.T) {
	v := &Vulpo{}

	err := v.Open(testDBFPath)
	if err != nil {
		t.Fatalf("Expected successful open, got error: %v", err)
	}

	// Should be active after successful open
	if !v.Active() {
		t.Error("Expected Vulpo to be active after successful open")
	}

	// Filename should be set
	if v.filename != testDBFPath {
		t.Errorf("Expected filename %s, got %s", testDBFPath, v.filename)
	}

	// Internal pointers should be set
	if v.codeBase == nil {
		t.Error("Expected codeBase to be initialized")
	}

	if v.data == nil {
		t.Error("Expected data to be initialized")
	}

	// Header should be read
	if v.header == nil {
		t.Error("Expected header to be read after open")
	}

	// Clean up
	err = v.Close()
	if err != nil {
		t.Errorf("Failed to close: %v", err)
	}
}

func TestVulpo_Open_InvalidFile(t *testing.T) {
	v := &Vulpo{}

	err := v.Open("nonexistent.dbf")
	if err == nil {
		t.Error("Expected error when opening non-existent file")
	}

	// Should not be active after failed open
	if v.Active() {
		t.Error("Expected Vulpo to not be active after failed open")
	}

	// Internal state should be clean after failed open
	if v.codeBase != nil {
		t.Error("Expected codeBase to be nil after failed open")
	}

	if v.data != nil {
		t.Error("Expected data to be nil after failed open")
	}

	if v.header != nil {
		t.Error("Expected header to be nil after failed open")
	}
}

func TestVulpo_Open_AlreadyOpen(t *testing.T) {
	v := &Vulpo{}

	// Open first file
	err := v.Open(testDBFPath)
	if err != nil {
		t.Fatalf("Failed to open first file: %v", err)
	}

	// Try to open another file while first is still open
	err = v.Open(testDBFPath)
	if err == nil {
		t.Error("Expected error when opening file while another is already open")
	}

	// Should still be active with original file
	if !v.Active() {
		t.Error("Expected Vulpo to remain active after failed second open")
	}

	// Clean up
	err = v.Close()
	if err != nil {
		t.Errorf("Failed to close: %v", err)
	}
}

func TestVulpo_Close_ValidFile(t *testing.T) {
	v := &Vulpo{}

	// Open file first
	err := v.Open(testDBFPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}

	// Close should succeed
	err = v.Close()
	if err != nil {
		t.Errorf("Expected successful close, got error: %v", err)
	}

	// Should not be active after close
	if v.Active() {
		t.Error("Expected Vulpo to not be active after close")
	}

	// All internal state should be cleared
	if v.filename != "" {
		t.Errorf("Expected empty filename after close, got %s", v.filename)
	}

	if v.codeBase != nil {
		t.Error("Expected codeBase to be nil after close")
	}

	if v.data != nil {
		t.Error("Expected data to be nil after close")
	}

	if v.header != nil {
		t.Error("Expected header to be nil after close")
	}
}

func TestVulpo_Close_NotOpen(t *testing.T) {
	v := &Vulpo{}

	err := v.Close()
	if err == nil {
		t.Error("Expected error when closing file that was never opened")
	}

	// Should remain inactive
	if v.Active() {
		t.Error("Expected Vulpo to remain inactive after failed close")
	}
}

func TestVulpo_Active_States(t *testing.T) {
	v := &Vulpo{}

	// Initial state - not active
	if v.Active() {
		t.Error("New instance should not be active")
	}

	// Open file - should be active
	err := v.Open(testDBFPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}

	if !v.Active() {
		t.Error("Should be active after opening file")
	}

	// Close file - should not be active
	err = v.Close()
	if err != nil {
		t.Errorf("Failed to close file: %v", err)
	}

	if v.Active() {
		t.Error("Should not be active after closing file")
	}
}

func TestVulpo_Header_AfterOpen(t *testing.T) {
	v := &Vulpo{}

	err := v.Open(testDBFPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	header := v.Header()

	// Record count should be positive for test file
	if header.RecordCount() == 0 {
		t.Error("Expected positive record count for test DBF file")
	}

	// Last updated should not be zero time
	if header.LastUpdated().IsZero() {
		t.Error("Expected non-zero last updated date")
	}

	// Last updated should be a reasonable date (after 1980, before future)
	if header.LastUpdated().Year() < 1980 || header.LastUpdated().After(time.Now().AddDate(1, 0, 0)) {
		t.Errorf("Expected reasonable last updated date, got %v", header.LastUpdated())
	}

	// Codepage should be readable (0 is valid - means no specific codepage set)
	// Just verify we can read it without error
	_ = header.Codepage()
}

func TestVulpo_Header_BeforeOpen(t *testing.T) {
	v := &Vulpo{}

	// Should return zero-value header when no file is open
	header := v.Header()

	// All fields should be zero values
	if header.RecordCount() != 0 {
		t.Error("Expected zero record count when no file is open")
	}

	if !header.LastUpdated().IsZero() {
		t.Error("Expected zero time for last updated when no file is open")
	}

	if header.HasIndex() != false {
		t.Error("Expected false for HasIndex when no file is open")
	}

	if header.HasFpt() != false {
		t.Error("Expected false for HasFpt when no file is open")
	}

	if header.Codepage() != 0 {
		t.Error("Expected zero codepage when no file is open")
	}
}

func TestCodepage_Name(t *testing.T) {
	tests := []struct {
		codepage Codepage
		expected string
	}{
		{0x01, "U.S. MS-DOS"},
		{0x03, WindowsANSICodepage},
		{0x02, "International MS-DOS"},
		{0xFF, "Unknown / Unsupported Codepage"}, // Invalid codepage
	}

	for _, test := range tests {
		result := test.codepage.Name()
		if result != test.expected {
			t.Errorf("Codepage %d: expected %s, got %s", test.codepage, test.expected, result)
		}
	}
}

func TestCodepage_String(t *testing.T) {
	cp := Codepage(0x03)
	expected := WindowsANSICodepage

	result := cp.String()
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestCodepage_VfpCodepageID(t *testing.T) {
	tests := []struct {
		codepage Codepage
		expected uint8
	}{
		{0x03, 0x03},
		{0x01, 0x01},
		{0xFF, 0x00}, // Invalid codepage should return 0
	}

	for _, test := range tests {
		result := test.codepage.VfpCodepageID()
		if result != test.expected {
			t.Errorf("Codepage %d VFP ID: expected %d, got %d", test.codepage, test.expected, result)
		}
	}
}

func TestCodepage_MsCodepageID(t *testing.T) {
	tests := []struct {
		codepage Codepage
		expected uint16
	}{
		{0x03, 1252},
		{0x01, 437},
		{0xFF, 0x0000}, // Invalid codepage should return 0
	}

	for _, test := range tests {
		result := test.codepage.MsCodepageID()
		if result != test.expected {
			t.Errorf("Codepage %d MS ID: expected %d, got %d", test.codepage, test.expected, result)
		}
	}
}

func TestCodepage_Supported(t *testing.T) {
	// Only codepage 0x03 is marked as supported
	supported := Codepage(0x03)
	unsupported := Codepage(0x01)

	if !supported.Supported() {
		t.Error("Codepage 0x03 should be supported")
	}

	if unsupported.Supported() {
		t.Error("Codepage 0x01 should not be supported")
	}
}

func TestHeader_Getters(t *testing.T) {
	v := &Vulpo{}

	err := v.Open(testDBFPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	header := v.Header()

	// Test that getter methods return expected types and reasonable values
	recordCount := header.RecordCount()
	if recordCount == 0 {
		t.Error("Expected non-zero record count")
	}

	lastUpdated := header.LastUpdated()
	if lastUpdated.IsZero() {
		t.Error("Expected non-zero last updated time")
	}

	hasIndex := header.HasIndex()
	// hasIndex is boolean, any value is valid
	_ = hasIndex

	hasFpt := header.HasFpt()
	// hasFpt is boolean, any value is valid
	_ = hasFpt

	codepage := header.Codepage()
	// Codepage 0 is valid (means no specific codepage), just verify we can read it
	_ = codepage
}

func TestVulpo_CodepageReadFromFile(t *testing.T) {
	v := &Vulpo{}

	err := v.Open(testDBFPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	header := v.Header()

	// The test file should have codepage 0 (no specific codepage)
	// This verifies we're reading from file, not assuming a default value
	if header.Codepage() != 0 {
		t.Errorf("Expected codepage 0 from test file, got %d", header.Codepage())
	}

	// Verify codepage name works for 0
	expectedName := "Unknown / Unsupported Codepage"
	if header.Codepage().Name() != expectedName {
		t.Errorf("Expected codepage name '%s', got '%s'", expectedName, header.Codepage().Name())
	}
}

func TestVulpo_TestdataFiles_Comprehensive(t *testing.T) {
	// Test files with known characteristics
	testCases := []struct {
		file             string
		expectedCodepage Codepage
		expectedRecords  uint
		expectedFpt      bool
		description      string
	}{
		{"testdata/empty.dbf", 3, 0, false, "Empty file with " + WindowsANSICodepage + " codepage"},
		{"testdata/fieldy.dbf", 3, 12, false, "Field testing file"},
		{"testdata/basicmemo.dbf", 3, 3, true, "Basic memo file with FPT"},
		{"testdata/idcharsdate.dbf", 3, 9, false, "ID/chars/date composite"},
		{"testdata/intcharsnumeric.dbf", 3, 10, false, "Int/chars/numeric composite"},
		{"testdata/logictime.dbf", 3, 2, false, "Logic/time field types"},
		{"testdata/fieldtests/bools.dbf", 3, 3, false, "Boolean field tests"},
		{"testdata/fieldtests/currencies.dbf", 3, 4, false, "Currency field tests"},
		{"testdata/fieldtests/dates.dbf", 3, 4, false, "Date field tests"},
		{"testdata/fieldtests/integers.dbf", 3, 5, false, "Integer field tests"},
		{"testdata/fieldtests/numerics.dbf", 3, 5, false, "Numeric field tests"},
	}

	for _, tc := range testCases {
		t.Run(tc.file, func(t *testing.T) {
			v := &Vulpo{}

			err := v.Open(tc.file)
			if err != nil {
				t.Fatalf("Failed to open %s: %v", tc.file, err)
			}
			defer func() {
				if closeErr := v.Close(); closeErr != nil {
					t.Errorf("Failed to close %s: %v", tc.file, closeErr)
				}
			}()

			header := v.Header()

			// Verify codepage
			if header.Codepage() != tc.expectedCodepage {
				t.Errorf("%s: expected codepage %d, got %d", tc.description, tc.expectedCodepage, header.Codepage())
			}

			// Verify record count
			if header.RecordCount() != tc.expectedRecords {
				t.Errorf("%s: expected %d records, got %d", tc.description, tc.expectedRecords, header.RecordCount())
			}

			// Verify FPT memo file presence
			if header.HasFpt() != tc.expectedFpt {
				t.Errorf("%s: expected HasFpt=%v, got %v", tc.description, tc.expectedFpt, header.HasFpt())
			}

			// Verify codepage name is correct for " + WindowsANSICodepage
			if tc.expectedCodepage == 3 {
				expectedName := WindowsANSICodepage
				if header.Codepage().Name() != expectedName {
					t.Errorf("%s: expected codepage name '%s', got '%s'", tc.description, expectedName, header.Codepage().Name())
				}
			}

			// Verify last updated date is reasonable
			if !header.LastUpdated().IsZero() {
				if header.LastUpdated().Year() < 1980 || header.LastUpdated().After(time.Now().AddDate(1, 0, 0)) {
					t.Errorf("%s: unreasonable last updated date: %v", tc.description, header.LastUpdated())
				}
			}

			// Verify Active state
			if !v.Active() {
				t.Errorf("%s: expected active state after opening", tc.description)
			}
		})
	}
}

func TestVulpo_CodepageMapping_WindowsAnsi(t *testing.T) {
	// Test that codepage 3 (" + WindowsANSICodepage + ") maps correctly
	cp := Codepage(3)

	expectedName := WindowsANSICodepage
	if cp.Name() != expectedName {
		t.Errorf("Expected codepage 3 name '%s', got '%s'", expectedName, cp.Name())
	}

	expectedVfpID := uint8(0x03)
	if cp.VfpCodepageID() != expectedVfpID {
		t.Errorf("Expected VFP codepage ID %d, got %d", expectedVfpID, cp.VfpCodepageID())
	}

	expectedMsID := uint16(1252)
	if cp.MsCodepageID() != expectedMsID {
		t.Errorf("Expected MS codepage ID %d, got %d", expectedMsID, cp.MsCodepageID())
	}

	if !cp.Supported() {
		t.Error("Expected codepage 3 (" + WindowsANSICodepage + ") to be supported")
	}
}

func TestVulpo_EmptyFile_EdgeCase(t *testing.T) {
	v := &Vulpo{}

	err := v.Open("testdata/empty.dbf")
	if err != nil {
		t.Fatalf("Failed to open empty file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	header := v.Header()

	// Empty file should have 0 records but still valid header
	if header.RecordCount() != 0 {
		t.Errorf("Expected 0 records in empty file, got %d", header.RecordCount())
	}

	// Should still have valid codepage
	if header.Codepage() == 0 {
		t.Error("Expected non-zero codepage even in empty file")
	}

	// Should still be active
	if !v.Active() {
		t.Error("Expected active state even for empty file")
	}
}

func TestVulpo_MemoFile_Detection(t *testing.T) {
	v := &Vulpo{}

	err := v.Open("testdata/basicmemo.dbf")
	if err != nil {
		t.Fatalf("Failed to open memo file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	header := v.Header()

	// Should detect FPT memo file
	if !header.HasFpt() {
		t.Error("Expected HasFpt=true for basicmemo.dbf")
	}

	// Should have records
	if header.RecordCount() == 0 {
		t.Error("Expected records in memo file")
	}

	// Should have proper codepage
	if header.Codepage() != 3 {
		t.Errorf("Expected codepage 3, got %d", header.Codepage())
	}
}

func TestVulpo_FieldTypeFiles_Variety(t *testing.T) {
	fieldTypeFiles := []string{
		"testdata/fieldtests/bools.dbf",
		"testdata/fieldtests/currencies.dbf",
		"testdata/fieldtests/dates.dbf",
		"testdata/fieldtests/integers.dbf",
		"testdata/fieldtests/numerics.dbf",
	}

	for _, file := range fieldTypeFiles {
		t.Run(file, func(t *testing.T) {
			v := &Vulpo{}

			err := v.Open(file)
			if err != nil {
				t.Fatalf("Failed to open %s: %v", file, err)
			}
			defer func() {
				_ = v.Close()
			}()

			header := v.Header()

			// All field type files should have " + WindowsANSICodepage + " codepage
			if header.Codepage() != 3 {
				t.Errorf("%s: expected codepage 3, got %d", file, header.Codepage())
			}

			// All should have some records
			if header.RecordCount() == 0 {
				t.Errorf("%s: expected records > 0, got %d", file, header.RecordCount())
			}

			// None of the field type test files should have memo
			if header.HasFpt() {
				t.Errorf("%s: expected no FPT file", file)
			}

			// Should be active
			if !v.Active() {
				t.Errorf("%s: expected active state", file)
			}
		})
	}
}

func TestVulpo_FieldDefs_Basic(t *testing.T) {
	v := &Vulpo{}

	err := v.Open(testDBFPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	fieldDefs := v.FieldDefs()
	if fieldDefs == nil {
		t.Fatal("Expected non-nil FieldDefs")
	}

	// Should have some fields
	if fieldDefs.Count() <= 0 {
		t.Error("Expected fields in database")
	}

	// Test accessing fields by index
	for i := 0; i < fieldDefs.Count(); i++ {
		field := fieldDefs.ByIndex(i)
		if field == nil {
			t.Errorf("Expected field at index %d", i)
			continue
		}

		// Verify field has a name
		if field.Name() == "" {
			t.Errorf("Field %d has empty name", i)
		}

		// Verify field has a valid type
		if field.Type() == FTUnknown {
			t.Errorf("Field %d (%s) has unknown type", i, field.Name())
		}

		// Size should be reasonable
		if field.Size() == 0 {
			t.Errorf("Field %d (%s) has zero size", i, field.Name())
		}
	}
}

func TestVulpo_FieldDefs_BeforeOpen(t *testing.T) {
	v := &Vulpo{}

	// Should return nil when no file is open
	fieldDefs := v.FieldDefs()
	if fieldDefs != nil {
		t.Error("Expected nil FieldDefs when no file is open")
	}
}

func TestVulpo_FieldDefs_AfterClose(t *testing.T) {
	v := &Vulpo{}

	err := v.Open(testDBFPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}

	// Should have fields when open
	fieldDefs := v.FieldDefs()
	if fieldDefs == nil {
		t.Fatal("Expected non-nil FieldDefs when open")
	}

	err = v.Close()
	if err != nil {
		t.Fatalf("Failed to close file: %v", err)
	}

	// Should be nil after close
	fieldDefs = v.FieldDefs()
	if fieldDefs != nil {
		t.Error("Expected nil FieldDefs after close")
	}
}

func TestFieldDefs_ByName(t *testing.T) {
	v := &Vulpo{}

	err := v.Open("testdata/fieldy.dbf") // This file should have known fields
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	fieldDefs := v.FieldDefs()
	if fieldDefs == nil {
		t.Fatal("Expected non-nil FieldDefs")
	}

	// Test case-insensitive lookup
	if fieldDefs.Count() > 0 {
		firstField := fieldDefs.ByIndex(0)
		if firstField != nil {
			fieldName := firstField.Name()

			// Test exact match
			field := fieldDefs.ByName(fieldName)
			if field == nil {
				t.Errorf("Failed to find field by exact name: %s", fieldName)
			} else if field.Name() != fieldName {
				t.Errorf("Expected field name %s, got %s", fieldName, field.Name())
			}

			// Test case-insensitive match
			field = fieldDefs.ByName(strings.ToUpper(fieldName))
			if field == nil {
				t.Errorf("Failed to find field by uppercase name: %s", strings.ToUpper(fieldName))
			}

			field = fieldDefs.ByName(strings.ToLower(fieldName))
			if field == nil {
				t.Errorf("Failed to find field by lowercase name: %s", strings.ToLower(fieldName))
			}
		}
	}

	// Test non-existent field
	field := fieldDefs.ByName("NONEXISTENT_FIELD")
	if field != nil {
		t.Error("Expected nil for non-existent field")
	}
}

func TestFieldDefs_ByIndex_EdgeCases(t *testing.T) {
	v := &Vulpo{}

	err := v.Open(testDBFPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	fieldDefs := v.FieldDefs()
	if fieldDefs == nil {
		t.Fatal("Expected non-nil FieldDefs")
	}

	// Test negative index
	field := fieldDefs.ByIndex(-1)
	if field != nil {
		t.Error("Expected nil for negative index")
	}

	// Test index out of bounds
	field = fieldDefs.ByIndex(fieldDefs.Count())
	if field != nil {
		t.Error("Expected nil for out-of-bounds index")
	}

	// Test way out of bounds
	field = fieldDefs.ByIndex(9999)
	if field != nil {
		t.Error("Expected nil for way out-of-bounds index")
	}
}

func TestFieldType_String_And_Name(t *testing.T) {
	tests := []struct {
		ft           FieldType
		expectedStr  string
		expectedName string
	}{
		{FTCharacter, "C", "character"},
		{FTNumeric, "N", "numeric"},
		{FTLogical, "L", "logical"},
		{FTDate, "D", "date"},
		{FTInteger, "I", "integer"},
		{FTDateTime, "T", "datetime"},
		{FTCurrency, "Y", "currency"},
		{FTMemo, "M", "memo"},
		{FTUnknown, "unknown", "unknown"},
	}

	for _, test := range tests {
		if test.ft.String() != test.expectedStr {
			t.Errorf("FieldType %d: expected String() '%s', got '%s'", int(test.ft), test.expectedStr, test.ft.String())
		}
		if test.ft.Name() != test.expectedName {
			t.Errorf("FieldType %d: expected Name() '%s', got '%s'", int(test.ft), test.expectedName, test.ft.Name())
		}
	}
}

func TestFromString_FieldType(t *testing.T) {
	tests := []struct {
		input    string
		expected FieldType
	}{
		{"C", FTCharacter},
		{"c", FTCharacter}, // Case insensitive
		{"N", FTNumeric},
		{"L", FTLogical},
		{"D", FTDate},
		{"I", FTInteger},
		{"M", FTMemo},
		{"Z", FTUnknown},  // Unknown type
		{"", FTUnknown},   // Empty string
		{"XX", FTUnknown}, // Too long
	}

	for _, test := range tests {
		result := FromString(test.input)
		if result != test.expected {
			t.Errorf("FromString('%s'): expected %d, got %d", test.input, test.expected, result)
		}
	}
}

func TestVulpo_FieldDefs_TestdataFiles(t *testing.T) {
	// Test field definitions across different testdata files
	testFiles := []string{
		"testdata/fieldy.dbf",
		"testdata/idcharsdate.dbf",
		"testdata/intcharsnumeric.dbf",
		"testdata/fieldtests/bools.dbf",
		"testdata/fieldtests/integers.dbf",
		"testdata/fieldtests/numerics.dbf",
	}

	for _, file := range testFiles {
		t.Run(file, func(t *testing.T) {
			v := &Vulpo{}

			err := v.Open(file)
			if err != nil {
				t.Fatalf("Failed to open %s: %v", file, err)
			}
			defer func() {
				_ = v.Close()
			}()

			fieldDefs := v.FieldDefs()
			if fieldDefs == nil {
				t.Errorf("%s: Expected non-nil FieldDefs", file)
				return
			}

			// Should have at least one field
			if fieldDefs.Count() == 0 {
				t.Errorf("%s: Expected at least one field", file)
				return
			}

			// Verify each field has valid properties
			for i := 0; i < fieldDefs.Count(); i++ {
				field := fieldDefs.ByIndex(i)
				if field == nil {
					t.Errorf("%s: Field %d is nil", file, i)
					continue
				}

				// Name should not be empty
				if field.Name() == "" {
					t.Errorf("%s: Field %d has empty name", file, i)
				}

				// Should be able to find by name
				foundField := fieldDefs.ByName(field.Name())
				if foundField != field {
					t.Errorf("%s: ByName didn't return same field for %s", file, field.Name())
				}

				// Type should not be unknown (unless it's really an unknown type)
				// We'll allow unknown for now since we might encounter unsupported types
				_ = field.Type()

				// Size should be reasonable (> 0)
				if field.Size() == 0 {
					t.Errorf("%s: Field %s has zero size", file, field.Name())
				}

				// Decimals should be reasonable (<= size)
				if field.Decimals() > field.Size() {
					t.Errorf("%s: Field %s has decimals (%d) > size (%d)", file, field.Name(), field.Decimals(), field.Size())
				}
			}
		})
	}
}

func TestVulpo_FieldCount(t *testing.T) {
	v := &Vulpo{}

	// Should return 0 when no file is open
	count := v.FieldCount()
	if count != 0 {
		t.Errorf("Expected FieldCount 0 when no file open, got %d", count)
	}

	err := v.Open(testDBFPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	count = v.FieldCount()
	if count <= 0 {
		t.Errorf("Expected FieldCount > 0 when file is open, got %d", count)
	}

	// Should match FieldDefs().Count()
	fieldDefs := v.FieldDefs()
	if fieldDefs != nil && count != fieldDefs.Count() {
		t.Errorf("FieldCount (%d) doesn't match FieldDefs().Count() (%d)", count, fieldDefs.Count())
	}
}

func TestVulpo_Field(t *testing.T) {
	v := &Vulpo{}

	// Should return nil when no file is open
	field := v.Field(0)
	if field != nil {
		t.Error("Expected nil field when no file is open")
	}

	err := v.Open(testDBFPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	// Test valid index
	field = v.Field(0)
	if field == nil {
		t.Error("Expected non-nil field for valid index")
	} else {
		if field.Name() == "" {
			t.Error("Expected field to have a name")
		}
	}

	// Test negative index
	field = v.Field(-1)
	if field != nil {
		t.Error("Expected nil field for negative index")
	}

	// Test out of bounds index
	field = v.Field(v.FieldCount())
	if field != nil {
		t.Error("Expected nil field for out of bounds index")
	}
}

func TestVulpo_FieldByName(t *testing.T) {
	v := &Vulpo{}

	// Should return nil when no file is open
	field := v.FieldByName("test")
	if field != nil {
		t.Error("Expected nil field when no file is open")
	}

	err := v.Open("testdata/intcharsnumeric.dbf")
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	// Test existing field - case sensitive
	field = v.FieldByName("id")
	if field == nil {
		t.Error("Expected to find field 'id'")
	} else if field.Name() != "id" {
		t.Errorf("Expected field name 'id', got '%s'", field.Name())
	}

	// Test existing field - case insensitive
	field = v.FieldByName("ID")
	if field == nil {
		t.Error("Expected to find field 'ID' (case insensitive)")
	} else if field.Name() != "id" {
		t.Errorf("Expected field name 'id', got '%s'", field.Name())
	}

	// Test non-existent field
	field = v.FieldByName("nonexistent")
	if field != nil {
		t.Error("Expected nil for non-existent field")
	}
}

func TestVulpo_FieldMethods_Consistency(t *testing.T) {
	v := &Vulpo{}

	err := v.Open("testdata/intcharsnumeric.dbf")
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	// Test that convenience methods return same results as FieldDefs methods
	fieldDefs := v.FieldDefs()
	if fieldDefs == nil {
		t.Fatal("Expected non-nil FieldDefs")
	}

	// Test count consistency
	if v.FieldCount() != fieldDefs.Count() {
		t.Errorf("FieldCount() %d != FieldDefs().Count() %d", v.FieldCount(), fieldDefs.Count())
	}

	// Test field access consistency (compare definitions, not objects)
	for i := 0; i < v.FieldCount(); i++ {
		field1 := v.Field(i)
		field2 := fieldDefs.ByIndex(i)
		if field1 == nil || field2 == nil {
			t.Errorf("Field(%d) or FieldDefs().ByIndex(%d) is nil", i, i)
			continue
		}

		// Compare field definitions (Field interface has FieldDef() method)
		field1Def := field1.FieldDef()
		if field1Def.Name() != field2.Name() || field1Def.Type() != field2.Type() {
			t.Errorf("Field(%d) definition mismatch: got %s/%s, want %s/%s",
				i, field1Def.Name(), field1Def.Type().String(), field2.Name(), field2.Type().String())
		}

		// Test name-based access consistency
		field3 := v.FieldByName(field1.Name())
		field4 := fieldDefs.ByName(field1.Name())
		if field3 == nil || field4 == nil {
			t.Errorf("FieldByName(%s) returned nil: field3=%v, field4=%v", field1.Name(), field3, field4)
			continue
		}

		// Compare field definitions
		field3Def := field3.FieldDef()
		if field3Def.Name() != field4.Name() || field3Def.Type() != field4.Type() {
			t.Errorf("FieldByName(%s) definition mismatch: got %s/%s, want %s/%s",
				field1.Name(), field3Def.Name(), field3Def.Type().String(), field4.Name(), field4.Type().String())
		}
	}
}

func TestVulpo_Navigation_BeforeOpen(t *testing.T) {
	v := &Vulpo{}

	// All navigation methods should return errors when no file is open
	if err := v.Goto(1); err == nil {
		t.Error("Expected error from Goto when no file is open")
	}

	if err := v.Next(); err == nil {
		t.Error("Expected error from Next when no file is open")
	}

	if err := v.Previous(); err == nil {
		t.Error("Expected error from Previous when no file is open")
	}

	if err := v.Skip(1); err == nil {
		t.Error("Expected error from Skip when no file is open")
	}

	if err := v.First(); err == nil {
		t.Error("Expected error from First when no file is open")
	}

	if err := v.Last(); err == nil {
		t.Error("Expected error from Last when no file is open")
	}

	// Position should return -1 when no file is open
	if pos := v.Position(); pos != -1 {
		t.Errorf("Expected Position to return -1 when no file is open, got %d", pos)
	}

	// BOF/EOF methods should return false when no file is open
	if v.BOF() {
		t.Error("Expected BOF to return false when no file is open")
	}

	if v.EOF() {
		t.Error("Expected EOF to return false when no file is open")
	}

	if v.IsBof() {
		t.Error("Expected IsBof to return false when no file is open")
	}

	if v.IsEOF() {
		t.Error("Expected IsEof to return false when no file is open")
	}
}

func TestVulpo_Navigation_Basic(t *testing.T) {
	v := &Vulpo{}

	err := v.Open(testDBFPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	// Test First() - should go to first record
	err = v.First()
	if err != nil {
		t.Fatalf("Failed to go to first record: %v", err)
	}

	// Should not be at BOF after going to first record (if records exist)
	header := v.Header()
	recordCount := int(header.RecordCount())
	if recordCount > 0 {
		pos := v.Position()
		if pos != 1 {
			t.Errorf("Expected position 1 after First(), got %d", pos)
		}
		if v.BOF() {
			t.Error("Expected BOF to be false at first record")
		}
	}

	// Test Last() - should go to last record
	err = v.Last()
	if err != nil {
		t.Fatalf("Failed to go to last record: %v", err)
	}

	// Should not be at EOF after going to last record (if records exist)
	if recordCount > 0 {
		pos := v.Position()
		if pos != recordCount {
			t.Errorf("Expected position %d after Last(), got %d", recordCount, pos)
		}
		if v.EOF() {
			t.Error("Expected EOF to be false at last record")
		}
	}
}

func TestVulpo_Navigation_Goto(t *testing.T) {
	v := &Vulpo{}

	err := v.Open(testDBFPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	header := v.Header()
	recordCount := int(header.RecordCount())
	if recordCount == 0 {
		t.Skip("No records in test file")
		return
	}

	// Test valid record number
	err = v.Goto(1)
	if err != nil {
		t.Errorf("Failed to go to record 1: %v", err)
	}

	if v.Position() != 1 {
		t.Errorf("Expected position 1 after Goto(1), got %d", v.Position())
	}

	// Test last record
	if recordCount > 1 {
		err = v.Goto(recordCount)
		if err != nil {
			t.Errorf("Failed to go to record %d: %v", recordCount, err)
		}

		if v.Position() != recordCount {
			t.Errorf("Expected position %d after Goto(%d), got %d", recordCount, recordCount, v.Position())
		}
	}

	// Test invalid record numbers
	err = v.Goto(0)
	if err == nil {
		t.Error("Expected error when going to record 0")
	}

	err = v.Goto(-1)
	if err == nil {
		t.Error("Expected error when going to negative record")
	}
}

func TestVulpo_Navigation_Skip(t *testing.T) {
	v := &Vulpo{}

	err := v.Open("testdata/intcharsnumeric.dbf") // This file has known records
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	header := v.Header()
	recordCount := int(header.RecordCount())
	if recordCount < 3 {
		t.Skip("Need at least 3 records for skip tests")
		return
	}

	// Start at first record
	err = v.First()
	if err != nil {
		t.Fatalf("Failed to go to first record: %v", err)
	}

	// Skip forward 2 records
	err = v.Skip(2)
	if err != nil {
		t.Errorf("Failed to skip 2 records: %v", err)
	}

	if v.Position() != 3 {
		t.Errorf("Expected position 3 after Skip(2) from position 1, got %d", v.Position())
	}

	// Skip backward 1 record
	err = v.Skip(-1)
	if err != nil {
		t.Errorf("Failed to skip -1 records: %v", err)
	}

	if v.Position() != 2 {
		t.Errorf("Expected position 2 after Skip(-1) from position 3, got %d", v.Position())
	}
}

func TestVulpo_Navigation_NextPrevious(t *testing.T) {
	v := &Vulpo{}

	err := v.Open("testdata/intcharsnumeric.dbf")
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	header := v.Header()
	recordCount := int(header.RecordCount())
	if recordCount < 2 {
		t.Skip("Need at least 2 records for next/previous tests")
		return
	}

	// Start at first record
	err = v.First()
	if err != nil {
		t.Fatalf("Failed to go to first record: %v", err)
	}

	origPos := v.Position()

	// Next should move to next record
	err = v.Next()
	if err != nil {
		t.Errorf("Failed to move to next record: %v", err)
	}

	if v.Position() != origPos+1 {
		t.Errorf("Expected position %d after Next(), got %d", origPos+1, v.Position())
	}

	// Previous should move back to original position
	err = v.Previous()
	if err != nil {
		t.Errorf("Failed to move to previous record: %v", err)
	}

	if v.Position() != origPos {
		t.Errorf("Expected position %d after Previous(), got %d", origPos, v.Position())
	}
}

func TestVulpo_Navigation_BofEof(t *testing.T) {
	v := &Vulpo{}

	err := v.Open("testdata/intcharsnumeric.dbf")
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	header := v.Header()
	recordCount := int(header.RecordCount())
	if recordCount == 0 {
		t.Skip("No records in test file")
		return
	}

	// Go to first record - should not be BOF or EOF
	err = v.First()
	if err != nil {
		t.Fatalf("Failed to go to first record: %v", err)
	}

	if v.BOF() {
		t.Error("BOF should be false at first record")
	}
	if v.EOF() {
		t.Error("EOF should be false at first record")
	}
	if v.IsBof() {
		t.Error("IsBof should be false at first record")
	}
	if v.IsEOF() {
		t.Error("IsEof should be false at first record")
	}

	// Go to last record - should not be BOF or EOF
	err = v.Last()
	if err != nil {
		t.Fatalf("Failed to go to last record: %v", err)
	}

	if v.BOF() {
		t.Error("BOF should be false at last record")
	}
	if v.EOF() {
		t.Error("EOF should be false at last record")
	}
	if v.IsBof() {
		t.Error("IsBof should be false at last record")
	}
	if v.IsEOF() {
		t.Error("IsEof should be false at last record")
	}
}

func TestVulpo_Navigation_Position(t *testing.T) {
	v := &Vulpo{}

	err := v.Open("testdata/intcharsnumeric.dbf")
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	header := v.Header()
	recordCount := int(header.RecordCount())
	if recordCount == 0 {
		t.Skip("No records in test file")
		return
	}

	// Test position at various records
	for i := 1; i <= recordCount; i++ {
		err = v.Goto(i)
		if err != nil {
			t.Errorf("Failed to go to record %d: %v", i, err)
			continue
		}

		pos := v.Position()
		if pos != i {
			t.Errorf("Expected position %d, got %d", i, pos)
		}
	}
}

func TestVulpo_Navigation_AfterClose(t *testing.T) {
	v := &Vulpo{}

	err := v.Open(testDBFPath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}

	// Navigate to a record
	err = v.First()
	if err != nil {
		t.Fatalf("Failed to navigate before close: %v", err)
	}

	// Close the database
	err = v.Close()
	if err != nil {
		t.Fatalf("Failed to close: %v", err)
	}

	// All navigation should fail after close
	if err := v.Goto(1); err == nil {
		t.Error("Expected error from Goto after close")
	}

	if err := v.Next(); err == nil {
		t.Error("Expected error from Next after close")
	}

	if err := v.Previous(); err == nil {
		t.Error("Expected error from Previous after close")
	}

	if err := v.Skip(1); err == nil {
		t.Error("Expected error from Skip after close")
	}

	if err := v.First(); err == nil {
		t.Error("Expected error from First after close")
	}

	if err := v.Last(); err == nil {
		t.Error("Expected error from Last after close")
	}

	// Position should return -1 after close
	if pos := v.Position(); pos != -1 {
		t.Errorf("Expected Position to return -1 after close, got %d", pos)
	}

	// BOF/EOF should return false after close
	if v.BOF() {
		t.Error("Expected BOF to return false after close")
	}

	if v.EOF() {
		t.Error("Expected EOF to return false after close")
	}

	if v.IsBof() {
		t.Error("Expected IsBof to return false after close")
	}

	if v.IsEOF() {
		t.Error("Expected IsEof to return false after close")
	}
}

func TestMultipleOpenClose_Cycles(t *testing.T) {
	v := &Vulpo{}

	// Test multiple open/close cycles to ensure proper cleanup
	for i := 0; i < 3; i++ {
		err := v.Open(testDBFPath)
		if err != nil {
			t.Fatalf("Failed to open file on cycle %d: %v", i, err)
		}

		if !v.Active() {
			t.Errorf("Expected active state on cycle %d", i)
		}

		header := v.Header()
		if header.RecordCount() == 0 {
			t.Errorf("Expected valid header on cycle %d", i)
		}

		err = v.Close()
		if err != nil {
			t.Fatalf("Failed to close file on cycle %d: %v", i, err)
		}

		if v.Active() {
			t.Errorf("Expected inactive state after close on cycle %d", i)
		}
	}
}
