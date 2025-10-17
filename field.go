package vulpo

import "strings"

// Field defines the unified interface for accessing both field definition information
// and field value reading capabilities. This interface extends FieldReader to provide
// a complete field access solution.
//
// Key Benefits of the Field Interface:
//   - Unified API: Single interface for both metadata and value access
//   - Automatic Creation: Field instances are created automatically at Open() time
//   - No Manual Management: No need to create or manage FieldReader instances
//   - Performance: Field readers are cached and reused for optimal performance
//   - Type Safety: Full type safety with comprehensive conversion methods
//
// Field instances are created automatically when opening a database and remain
// valid until the database is closed. Use FieldByName() or Field() to access
// field instances.
//
// Usage Example:
//
//	v := &Vulpo{}
//	v.Open("data.dbf")
//	defer v.Close()
//
//	v.First(0)
//	nameField := v.FieldByName("CUSTOMER_NAME")
//	name, _ := nameField.AsString()           // Read current record value
//	fieldType := nameField.Type()            // Access field definition
//	fieldSize := nameField.Size()            // Field metadata
//	isNull, _ := nameField.IsNull()          // Check for null values
type Field interface {
	FieldReader
}

// Fields provides access to the database field collection with both
// index-based and name-based lookup capabilities.
//
// The Fields collection is automatically populated when opening a database
// and contains pre-initialized Field instances for optimal performance.
// All field readers are created once during Open() and cached for reuse.
//
// Key Features:
//   - Automatic initialization at Open() time
//   - Case-insensitive field name lookup
//   - Zero-based indexing for field access
//   - Thread-safe for read operations
//   - Automatic cleanup at Close() time
//
// Fields remain valid until the database is closed.
type Fields struct {
	fields  []Field
	indices map[string]int // name -> index mapping (case-insensitive)
}

// Count returns the total number of fields in the database.
func (f *Fields) Count() int {
	return len(f.fields)
}

// ByIndex returns the field at the specified zero-based index.
// Returns nil if the index is out of bounds.
func (f *Fields) ByIndex(index int) Field {
	if index < 0 || index >= len(f.fields) {
		return nil
	}
	return f.fields[index]
}

// ByName returns the field with the specified name.
// The lookup is case-insensitive. Returns nil if the field is not found.
func (f *Fields) ByName(name string) Field {
	if f.indices == nil {
		return nil
	}

	index, exists := f.indices[strings.ToLower(name)]
	if !exists {
		return nil
	}

	return f.ByIndex(index)
}

// reset clears all fields and indices, preparing for reuse or cleanup
func (f *Fields) reset() {
	f.fields = nil
	f.indices = nil
}
