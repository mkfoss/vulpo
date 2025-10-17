// Package vulpo provides a Go interface to DBF (dBASE) files using the CodeBase library.
// It offers comprehensive functionality for reading, writing, navigating, and manipulating
// DBF databases with support for indexes, field types, deleted records, and expression filtering.
//
// Key features:
//   - Full DBF file support (dBASE III, IV, V, FoxPro, Clipper)
//   - Complete field type support (Character, Numeric, Date, Logical, Memo, etc.)
//   - Unified Field API with automatic field reader creation
//   - Index/tag support with seek operations
//   - Record navigation and positioning
//   - Deleted record handling (soft delete/recall/pack)
//   - dBASE expression filtering
//   - Regex search capabilities
//   - Memory-safe C integration
//
// Basic usage with the new unified Field API:
//
//	v := &Vulpo{}
//	err := v.Open("data.dbf")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer v.Close()
//
//	// Navigate and read records using unified Field interface
//	v.First(0)
//	for !v.EOF() {
//		nameField := v.FieldByName("NAME")  // Get field with both definition and reading
//		name, _ := nameField.AsString()     // Read current record value
//		fmt.Printf("%s (type: %s)\n", name, nameField.Type().String())
//		v.Next()
//	}
//
// Migration from v1.x:
//
// The Field API has been simplified in v2.0. Field readers are now created automatically
// at Open() time, eliminating the need for manual FieldReader() calls:
//
//	// Old API (v1.x - still works but deprecated):
//	// fieldReader := v.FieldReader("NAME")
//	// name, _ := fieldReader.AsString()
//
//	// New API (v2.0 - recommended):
//	nameField := v.FieldByName("NAME")
//	name, _ := nameField.AsString()
//	fieldType := nameField.Type()  // Definition access included
package vulpo

/*
#cgo CFLAGS: -I./mkfdbflib
#cgo LDFLAGS: -L./mkfdbflib -lmkfdbf
#include "d4all.h"
#include <stdlib.h>
*/
import "C"
import (
	"runtime"
	"strings"
	"time"
	"unsafe"
)

// headerRead represents the first 32 bytes of a DBF file header
type headerRead struct {
	MagicByte       uint8
	LastupdateYear  uint8
	LastupdateMonth uint8
	LastupdateDay   uint8
	Recordcount     uint32
	RecordOffset    uint16
	RecordSize      uint16
	ReservedOne     [16]byte
	TableFlags      uint8
	CodePage        uint8
	Reserved2       [2]byte
}

// Vulpo represents a connection to a DBF database file.
// It provides methods for opening, navigating, reading, and manipulating DBF files.
// The zero value is ready to use - call Open() to connect to a database file.
//
// Field Access (v2.0):
// Fields are automatically initialized when opening a database. Use FieldByName() or
// Field() to get field instances that provide both definition and reading capabilities.
// This eliminates the need for separate FieldReader() calls.
//
// All navigation, field access, and data modification operations require an active
// connection (Open() called successfully). The connection is automatically closed
// when the Vulpo instance is garbage collected, but Close() should be called
// explicitly for proper resource management.
type Vulpo struct {
	filename  string
	codeBase  *C.CODE4
	data      *C.DATA4
	header    *Header
	fieldDefs *FieldDefs // kept for internal use during creation
	fields    *Fields    // public field collection with readers
}

// Open establishes a connection to the specified DBF file.
// The filename should include the full path and .dbf extension.
//
// Parameters:
//   - filename: Path to the DBF file to open
//
// Returns:
//   - error: nil on success, error describing the failure otherwise
//
// The method initializes the underlying CodeBase library, opens the database file,
// and reads the header and field definitions. If opening fails, all resources
// are cleaned up automatically.
//
// Example:
//
//	v := &Vulpo{}
//	err := v.Open("/path/to/data.dbf")
//	if err != nil {
//		log.Fatalf("Failed to open database: %v", err)
//	}
//	defer v.Close()
func (v *Vulpo) Open(filename string) error {
	if v.Active() {
		return NewError("database already open")
	}

	// Initialize CODE4 structure
	v.codeBase = (*C.CODE4)(C.malloc(C.sizeof_CODE4))
	if v.codeBase == nil {
		return NewError("failed to allocate CODE4 structure")
	}

	// Initialize the codebase using code4initLow (code4init macro expansion)
	result := C.code4initLow(v.codeBase, nil, 6401, C.long(C.sizeof_CODE4))
	if result != 0 {
		C.free(unsafe.Pointer(v.codeBase))
		v.codeBase = nil
		return NewErrorf("failed to initialize codebase: %d", int(result))
	}

	// Convert Go string to C string
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	// Open the data file
	v.data = C.d4open(v.codeBase, cFilename)
	if v.data == nil {
		// Clean up on failure
		C.code4initUndo(v.codeBase)
		C.free(unsafe.Pointer(v.codeBase))
		v.codeBase = nil
		return NewErrorf("failed to open database file: %s", filename)
	}

	v.filename = filename

	// Set finalizer to ensure cleanup
	runtime.SetFinalizer(v, (*Vulpo).finalize)

	return v.readHeader()
}

// Close closes the database connection and releases all associated resources.
// This method should be called when done with the database, typically using defer.
//
// Returns:
//   - error: nil on success, error if closing fails
//
// After Close() is called, the Vulpo instance can be reused by calling Open()
// with a new filename. All field readers and other references become invalid.
//
// Example:
//
//	v := &Vulpo{}
//	err := v.Open("data.dbf")
//	if err != nil {
//		return err
//	}
//	defer v.Close()  // Always close when done
func (v *Vulpo) Close() error {
	if !v.Active() {
		return NewError("database not open")
	}

	// Remove finalizer since we've cleaned up manually
	runtime.SetFinalizer(v, nil)

	return v.reset()
}

// Active reports whether the database connection is active and ready for use.
//
// Returns:
//   - bool: true if Open() was called successfully and Close() has not been called
//
// This method is used internally by most operations to validate the connection state.
// All navigation, field access, and modification operations require an active connection.
//
// Example:
//
//	v := &Vulpo{}
//	if !v.Active() {
//		fmt.Println("Database not open")
//	}
func (v *Vulpo) Active() bool {
	return v.data != nil
}

// finalize is called by the garbage collector to ensure cleanup
func (v *Vulpo) finalize() {
	if v.Active() {
		// Best effort cleanup - ignore errors since we can't return them
		_ = v.Close()
	}
}

func (v *Vulpo) reset() error {
	// Close the data file
	if v.data != nil {
		result := C.d4close(v.data)
		v.data = nil
		if result != 0 {
			return NewErrorf("failed to close database: %d", int(result))
		}
	}

	// Cleanup the codebase
	if v.codeBase != nil {
		C.code4initUndo(v.codeBase)
		C.free(unsafe.Pointer(v.codeBase))
		v.codeBase = nil
	}

	// Clear all state
	v.filename = ""
	v.header = nil
	v.fieldDefs = nil

	// Clean up field readers
	if v.fields != nil {
		v.fields.reset()
		v.fields = nil
	}

	return nil
}

// Header returns the database file header information.
//
// Returns:
//   - Header: Contains record count, last update date, codepage, index flags, etc.
//
// Returns a zero-value Header if no database is open. The header provides
// metadata about the database structure and content.
//
// Example:
//
//	header := v.Header()
//	fmt.Printf("Records: %d, Last Updated: %s\n",
//		header.RecordCount(), header.LastUpdated().Format("2006-01-02"))
func (v *Vulpo) Header() Header {
	if v.header == nil {
		// Return zero-value header when no file is open
		return Header{}
	}
	return *v.header
}

// Fields returns the field collection for all fields in the database.
//
// Returns:
//   - *Fields: Container with all field instances, nil if no database is open
//
// The Fields collection provides access to both field metadata and value reading
// capabilities. Field readers are created automatically when opening the database.
//
// Example:
//
//	fields := v.Fields()
//	if fields != nil {
//		fmt.Printf("Database has %d fields\n", fields.Count())
//		nameField := fields.ByName("NAME")
//		if nameField != nil {
//			value, _ := nameField.AsString()
//			fmt.Printf("Name: %s\n", value)
//		}
//	}
func (v *Vulpo) Fields() *Fields {
	return v.fields
}

// FieldDefs returns the field definitions container for all fields in the database.
// This method is kept for backward compatibility - prefer using Fields() for new code.
//
// Returns:
//   - *FieldDefs: Container with all field definitions, nil if no database is open
func (v *Vulpo) FieldDefs() *FieldDefs {
	return v.fieldDefs
}

// FieldCount returns the total number of fields in the database.
//
// Returns:
//   - int: Number of fields in the database, 0 if no database is open
//
// This is useful for iterating through all fields or validating field indexes.
//
// Example:
//
//	for i := 0; i < v.FieldCount(); i++ {
//		field := v.Field(i)
//		fmt.Printf("Field %d: %s (%s)\n", i, field.Name(), field.Type())
//	}
func (v *Vulpo) FieldCount() int {
	if v.fields == nil {
		return 0
	}
	return v.fields.Count()
}

// Field returns the field instance at the specified zero-based index.
//
// Parameters:
//   - index: Zero-based field index (0 to FieldCount()-1)
//
// Returns:
//   - Field: Field instance with both definition and reading capabilities, or nil if invalid
//
// Returns nil if the index is out of bounds or no database is open.
// The returned field can be used to access both metadata and current record values.
//
// The Field instance provides unified access to both field definition and
// current record data, eliminating the need for separate FieldReader calls.
//
// Example:
//
//	// Iterate through all fields
//	for i := 0; i < v.FieldCount(); i++ {
//		field := v.Field(i)
//		if field != nil {
//			fmt.Printf("Field %d: %s (%s)\n", i, field.Name(), field.Type().String())
//			value, _ := field.AsString()
//			fmt.Printf("  Current value: %s\n", value)
//		}
//	}
func (v *Vulpo) Field(index int) Field {
	if v.fields == nil {
		return nil
	}
	return v.fields.ByIndex(index)
}

// FieldByName returns the field instance with the specified name.
//
// Parameters:
//   - name: Field name to lookup (case-insensitive)
//
// Returns:
//   - Field: Field instance with both definition and reading capabilities, or nil if not found
//
// The lookup is case-insensitive and returns nil if the field is not found
// or no database is open. The returned field can be used to access both
// metadata and current record values.
//
// The Field instance includes both definition methods (Type, Size, etc.) and
// reading methods (AsString, AsInt, etc.), eliminating the need for separate
// FieldReader calls.
//
// Example:
//
//	field := v.FieldByName("CUSTOMER_NAME")
//	if field != nil {
//		// Access field definition
//		fmt.Printf("Field type: %s, Size: %d\n", field.Type().String(), field.Size())
//
//		// Read current record value
//		value, _ := field.AsString()
//		fmt.Printf("Current value: %s\n", value)
//
//		// Check for null
//		isNull, _ := field.IsNull()
//		fmt.Printf("Is null: %t\n", isNull)
//	}
func (v *Vulpo) FieldByName(name string) Field {
	if v.fields == nil {
		return nil
	}
	return v.fields.ByName(name)
}

func (v *Vulpo) readHeader() error {
	if v.data == nil {
		return NewError("no database open")
	}

	// Access the DATA4FILE structure through the DATA4
	dataFile := v.data.dataFile
	if dataFile == nil {
		return NewError("invalid database file structure")
	}

	// Read the first 32 bytes of the DBF file header directly
	headerBytes := make([]byte, 32)
	result := C.file4read(&dataFile.file, 0, unsafe.Pointer(&headerBytes[0]), 32)
	if result != 32 {
		return NewError("failed to read DBF header")
	}

	// Parse header using the struct layout
	headerRead := (*headerRead)(unsafe.Pointer(&headerBytes[0]))

	// Create new header with data from file
	header := &Header{}

	// Record count from file header (little-endian)
	header.recordcount = uint(headerRead.Recordcount)

	// Date from file header
	year := int(headerRead.LastupdateYear)
	if year < 80 {
		year += 2000 // Y2K handling: 00-79 = 2000-2079
	} else {
		year += 1900 // 80-99 = 1980-1999
	}
	month := int(headerRead.LastupdateMonth)
	day := int(headerRead.LastupdateDay)

	if month >= 1 && month <= 12 && day >= 1 && day <= 31 {
		header.lastUpdated = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	} else {
		// Invalid date, use zero time
		header.lastUpdated = time.Time{}
	}

	// Read actual codepage from file header
	header.codepage = Codepage(headerRead.CodePage)

	// For FoxPro files, detect CDX index from table flags
	// TableFlags bit 0 = CDX index exists
	header.hasIndex = (headerRead.TableFlags & 0x01) != 0

	// FoxPro memo files use FPT extension, detected via table flags bit 1
	header.hasFpt = (headerRead.TableFlags & 0x02) != 0

	// Validate against codebase values for consistency
	if uint32(C.d4recCountDo(v.data)) != headerRead.Recordcount {
		return NewError("header record count mismatch with codebase")
	}

	v.header = header

	// Read field definitions
	err := v.readFieldDefs()
	if err != nil {
		return err
	}

	return nil
}

func (v *Vulpo) readFieldDefs() error {
	if v.data == nil {
		return NewError("no database open")
	}

	// Get field count from codebase
	fieldCount := int(C.d4numFields(v.data))
	if fieldCount <= 0 {
		return NewError("no fields found in database")
	}

	// Create FieldDefs structure (internal use)
	fieldDefs := &FieldDefs{
		fields:   make([]*FieldDef, 0, fieldCount),
		indicies: make(map[string]int),
	}

	// Create Fields structure (public API)
	fields := &Fields{
		fields:  make([]Field, 0, fieldCount),
		indices: make(map[string]int),
	}

	// Read each field definition and create readers
	for i := 0; i < fieldCount; i++ {
		// Get field pointer from codebase (1-indexed)
		cField := C.d4fieldJ(v.data, C.int(i+1))
		if cField == nil {
			return NewErrorf("failed to get field %d", i+1)
		}

		// Extract field information
		fieldDef := &FieldDef{
			fieldname: C.GoString(&cField.name[0]),
			fieldtype: FromString(string(rune(cField._type))),
			size:      uint8(cField.len),
			decimals:  uint8(cField.dec),
			nullable:  cField.null != 0,
			binary:    cField.binary != 0,
			system:    false, // Basic implementation - can be enhanced
		}

		// Add to FieldDefs (internal)
		fieldDefs.fields = append(fieldDefs.fields, fieldDef)
		fieldDefs.indicies[strings.ToLower(fieldDef.fieldname)] = i

		// Create field reader and add to Fields (public)
		fieldReader := v.createFieldReader(cField, fieldDef)
		fields.fields = append(fields.fields, fieldReader)
		fields.indices[strings.ToLower(fieldDef.fieldname)] = i
	}

	v.fieldDefs = fieldDefs
	v.fields = fields
	return nil
}
