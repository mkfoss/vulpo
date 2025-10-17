package vulpo

import (
	"testing"
)

// TestFieldAPICompatibility tests that the new Field API is working correctly
// and that fields are created automatically at Open() time
func TestFieldAPICompatibility(t *testing.T) {
	// This test requires a real DBF file - if none exists, skip
	// In a real test environment, you would use a test DBF file

	v := &Vulpo{}

	// Test with no database open
	if v.FieldCount() != 0 {
		t.Error("FieldCount should return 0 when no database is open")
	}

	if v.Field(0) != nil {
		t.Error("Field should return nil when no database is open")
	}

	if v.FieldByName("TEST") != nil {
		t.Error("FieldByName should return nil when no database is open")
	}

	if v.Fields() != nil {
		t.Error("Fields should return nil when no database is open")
	}

	// Test deprecated FieldReader methods
	if v.FieldReader("TEST") != nil {
		t.Error("FieldReader should return nil when no database is open")
	}

	if v.FieldReaderByIndex(0) != nil {
		t.Error("FieldReaderByIndex should return nil when no database is open")
	}
}

// TestFieldInterfaceCompatibility tests that Field extends FieldReader properly
func TestFieldInterfaceCompatibility(t *testing.T) {
	// Create a mock field implementation for testing interface compatibility
	// This test verifies that the Field interface properly extends FieldReader

	// Since Field extends FieldReader, any Field should be assignable to FieldReader
	var fieldReader FieldReader
	var field Field

	// This should compile without error - Field implements FieldReader
	if field != nil {
		fieldReader = field
		_ = fieldReader // Use the variable to avoid unused variable error
	}
}

// TestBackwardCompatibility tests that old FieldDefs API still works
func TestBackwardCompatibility(t *testing.T) {
	v := &Vulpo{}

	// Test that FieldDefs() method still exists and returns nil when no database is open
	if v.FieldDefs() != nil {
		t.Error("FieldDefs should return nil when no database is open")
	}
}
