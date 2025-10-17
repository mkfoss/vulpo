# Vulpo - Go DBF (dBASE) Library

Vulpo is a comprehensive Go library for reading, writing, and manipulating DBF (dBASE) database files. Built on the proven CodeBase library, it provides a modern, type-safe interface to DBF files with full support for indexes, field types, navigation, and advanced features like expression filtering and deleted record handling.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Database Operations](#database-operations)
- [Field Access](#field-access)
- [Navigation](#navigation)
- [Index Operations](#index-operations)
- [Searching](#searching)
- [Deleted Record Handling](#deleted-record-handling)
- [Advanced Features](#advanced-features)
- [API Reference](#api-reference)
- [Examples](#examples)
- [Compatibility](#compatibility)

## Features

### Core Features
- **Full DBF Support**: Compatible with dBASE III, IV, V, FoxPro, Clipper, and other xBase variants
- **Complete Field Types**: Character, Numeric, Date, Logical, Memo, Currency, DateTime, Float, Double, Integer
- **Memory Safe**: Proper C library integration with automatic resource management
- **Thread Safe**: Safe for concurrent read operations (write operations require external synchronization)

### Navigation & Indexing
- **Flexible Navigation**: First, Last, Next, Previous, Skip, Goto operations
- **Index Support**: Full index/tag operations with seek capabilities
- **Position Tracking**: Physical and logical record positioning
- **Multiple Index Types**: CDX, NDX, and other xBase index formats

### Advanced Features
- **Expression Filtering**: Native dBASE expression evaluation for complex queries
- **Regex Search**: Pattern-based searching with index optimization
- **Deleted Record Handling**: Soft delete, recall, pack operations following dBASE conventions
- **Header Information**: Access to file metadata, record counts, update dates
- **Error Handling**: Comprehensive error reporting with context

## Installation

```bash
go get github.com/mkfoss/vulpo
```

### Version Information

- **v2.0+**: Unified Field API with automatic field reader creation (current)
- **v1.x**: Manual field reader creation with separate FieldDef access (legacy)

### Prerequisites
- Go 1.18 or later
- Linux, macOS, or Windows
- CGO enabled (for C library integration)

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/mkfoss/vulpo"
)

func main() {
    // Open a DBF file
    v := &vulpo.Vulpo{}
    err := v.Open("customers.dbf")
    if err != nil {
        log.Fatal(err)
    }
    defer v.Close()
    
    // Get database information
    header := v.Header()
    fmt.Printf("Records: %d, Last Updated: %s\n", 
        header.RecordCount(), header.LastUpdated().Format("2006-01-02"))
    
    // List all fields
    fmt.Println("Fields:")
    for i := 0; i < v.FieldCount(); i++ {
        field := v.Field(i)
        fmt.Printf("  %s (%s) - width:%d\n", 
            field.Name(), field.Type().String(), field.Size())
    }
    
    // Navigate through records
    v.First(0)
    for !v.EOF() {
        // Read field values using the new Field API
        nameField := v.FieldByName("NAME")
        name, _ := nameField.AsString()
        
        ageField := v.FieldByName("AGE")
        age, _ := ageField.AsInt()
        
        fmt.Printf("Record %d: %s, Age: %d\n", v.Position(), name, age)
        
        v.Next()
    }
}
```

## Database Operations

### Opening and Closing

```go
// Open a database
v := &vulpo.Vulpo{}
err := v.Open("data.dbf")
if err != nil {
    return fmt.Errorf("failed to open database: %w", err)
}
defer v.Close()  // Always close when done

// Check if database is open
if v.Active() {
    fmt.Println("Database is open and ready")
}
```

### Header Information

```go
header := v.Header()
fmt.Printf("Database Information:\n")
fmt.Printf("  Total Records: %d\n", header.RecordCount())
fmt.Printf("  Last Updated: %s\n", header.LastUpdated().Format("2006-01-02"))
fmt.Printf("  Has Index: %t\n", header.HasIndex())
fmt.Printf("  Codepage: %d\n", int(header.Codepage()))
```

## Field Access (Unified API v2.0)

### Overview

Vulpo v2.0 introduces a unified Field API that simplifies field access by automatically creating field readers when opening a database. This eliminates the need for manual field reader creation and provides a cleaner, more intuitive interface.

### Key Improvements

- **Automatic Initialization**: Field readers are created automatically at `Open()` time
- **Unified Interface**: Single `Field` interface provides both definition and reading capabilities
- **Better Performance**: Field readers are cached and reused, avoiding repeated creation
- **Simplified API**: No need to manage separate `FieldDef` and `FieldReader` objects
- **Backward Compatible**: Old APIs still work but are deprecated

### Quick Comparison

**Old API (v1.x - still works but deprecated):**
```go
// Separate objects for definition and reading
fieldDef := v.FieldByName("NAME")         // Get definition
fieldReader := v.FieldReader("NAME")     // Create reader
value, _ := fieldReader.AsString()       // Read value
fieldType := fieldDef.Type()             // Get type from definition
```

**New API (v2.0 - recommended):**
```go
// Single object for both definition and reading
field := v.FieldByName("NAME")           // Get field with both capabilities
value, _ := field.AsString()             // Read value
fieldType := field.Type()                // Get type from same object
```

### Field Information

```go
// Get field count
fieldCount := v.FieldCount()
fmt.Printf("Database has %d fields\n", fieldCount)

// Access fields by index
for i := 0; i < v.FieldCount(); i++ {
    field := v.Field(i)
    fmt.Printf("Field %d: %s (%s) - Size:%d, Decimals:%d\n",
        i, field.Name(), field.Type().String(), field.Size(), field.Decimals())
}

// Access field by name (case-insensitive)
field := v.FieldByName("CUSTOMER_NAME")
if field != nil {
    fmt.Printf("Found field: %s\n", field.Name())
}
```

### Reading Field Values

```go
// Position at a record first
v.First(0)

// Access fields directly (no need to create separate readers)
nameField := v.FieldByName("NAME")
ageField := v.FieldByName("AGE")
salaryField := v.FieldByName("SALARY")
activeField := v.FieldByName("ACTIVE")
birthDateField := v.FieldByName("BIRTH_DATE")

// Read values in different types
name, err := nameField.AsString()
age, err := ageField.AsInt()
salary, err := salaryField.AsFloat()
active, err := activeField.AsBool()
birthDate, err := birthDateField.AsTime()

fmt.Printf("Name: %s, Age: %d, Salary: %.2f, Active: %t, Born: %s\n",
    name, age, salary, active, birthDate.Format("2006-01-02"))

// Check for null values
if nameField.IsNull() {
    fmt.Println("Name field is null")
}

// Access field metadata
fmt.Printf("Name field: type=%s, size=%d\n", nameField.Type().String(), nameField.Size())
```

### Field Type Support

Vulpo supports all standard DBF field types:

- **Character (C)**: Text fields, fixed width
- **Numeric (N)**: Numbers with optional decimals  
- **Date (D)**: Dates in CCYYMMDD format
- **Logical (L)**: Boolean values (T/F, Y/N)
- **Memo (M)**: Large text fields stored separately
- **Currency (Y)**: Money values with 4 decimal places
- **DateTime (T)**: Date and time values
- **Float (F)**: Floating-point numbers
- **Double (B)**: Double-precision numbers  
- **Integer (I)**: 32-bit integers

## Navigation

### Basic Navigation

```go
// Move to specific positions
err := v.First(0)    // Go to first record
err = v.Last(0)      // Go to last record
err = v.Goto(15)     // Go to record 15 (1-indexed)

// Sequential navigation
err = v.Next()       // Next record
err = v.Previous()   // Previous record
err = v.Skip(10)     // Skip 10 records forward
err = v.Skip(-5)     // Skip 5 records backward

// Check position and boundaries
pos := v.Position()  // Current record number (1-indexed)
if v.EOF() {
    fmt.Println("At end of file")
}
if v.BOF() {
    fmt.Println("At beginning of file") 
}
```

### Navigation with Indexes

Navigation order follows the currently selected index:

```go
// List available indexes
tags := v.ListTags()
fmt.Printf("Available indexes: %d\n", len(tags))
for _, tag := range tags {
    fmt.Printf("  %s\n", tag.Name())
}

// Select an index for navigation
nameTag := v.TagByName("NAME_IDX")
if nameTag != nil {
    v.SelectTag(nameTag)
    
    // Now navigation follows name alphabetical order
    v.First(0)  // First alphabetically
    v.Next()    // Next alphabetically
}

// Return to physical record order
v.SelectTag(nil)
```

## Index Operations

### Working with Indexes

```go
// Check what index is currently selected
selectedTag := v.SelectedTag()
if selectedTag != nil {
    fmt.Printf("Using index: %s\n", selectedTag.Name())
} else {
    fmt.Println("Using physical record order")
}

// Get the default index
defaultTag := v.DefaultTag()
if defaultTag != nil {
    fmt.Printf("Default index: %s\n", defaultTag.Name())
}
```

### Seeking Records

```go
// Select an index first
nameTag := v.TagByName("NAME_IDX")
v.SelectTag(nameTag)

// Seek for specific values
result, err := v.Seek("Smith")
if err != nil {
    log.Printf("Seek failed: %v", err)
} else {
    switch result {
    case vulpo.SeekSuccess:
        fmt.Println("Found exact match")
    case vulpo.SeekAfter:
        fmt.Println("Positioned after where record would be")
    case vulpo.SeekEOF:
        fmt.Println("Value would be after last record")
    }
}

// Continue seeking for more matches
for {
    result, err := v.SeekNext("Smith")
    if result != vulpo.SeekSuccess {
        break
    }
    
    // Process matching record
    nameReader := v.FieldReader("NAME")
    name, _ := nameReader.AsString()
    fmt.Printf("Found: %s\n", name)
}

// Seek with numeric values
salaryTag := v.TagByName("SALARY_IDX")
v.SelectTag(salaryTag)
result, err = v.SeekDouble(50000.0)  // More efficient for numbers
```

## Searching

### Expression-Based Searching

Vulpo supports native dBASE expressions for powerful filtering:

```go
// Simple field comparisons
results, err := v.SearchByExpression("AGE > 30", nil)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d records with age > 30\n", len(results.Matches))
for _, match := range results.Matches {
    nameReader := match.FieldReaders["NAME"]
    ageReader := match.FieldReaders["AGE"]
    
    name, _ := nameReader.AsString()
    age, _ := ageReader.AsInt()
    
    fmt.Printf("Record %d: %s, Age: %d\n", match.RecordNumber, name, age)
}

// Complex expressions with functions
results, err = v.SearchByExpression(
    "YEAR(BIRTH_DATE) = 1990 .AND. SUBSTR(NAME, 1, 1) = 'J'", 
    &vulpo.ExprSearchOptions{MaxResults: 10})

// String functions
results, err = v.SearchByExpression("UPPER(LEFT(NAME, 3)) = 'SMI'", nil)

// Date functions  
results, err = v.SearchByExpression("MONTH(HIRE_DATE) = 12", nil)

// Logical operations
results, err = v.SearchByExpression(
    "(SALARY > 50000 .OR. BONUS > 10000) .AND. ACTIVE", nil)
```

### Expression Functions

Supported dBASE expression functions include:

**String Functions:**
- `SUBSTR(string, start, length)` - Extract substring
- `LEFT(string, count)` - Left characters  
- `RIGHT(string, count)` - Right characters
- `UPPER(string)` - Convert to uppercase
- `TRIM(string)` - Remove trailing spaces
- `LTRIM(string)` - Remove leading spaces

**Date Functions:**
- `YEAR(date)` - Extract year
- `MONTH(date)` - Extract month
- `DAY(date)` - Extract day
- `CTOD(string)` - Convert string to date
- `DTOS(date)` - Convert date to string

**Numeric Functions:**
- `STR(number, length, decimals)` - Convert number to string
- `VAL(string)` - Convert string to number

**Conditional:**
- `IIF(condition, true_value, false_value)` - Conditional expression

**Record Functions:**
- `RECNO()` - Current record number
- `RECCOUNT()` - Total record count
- `DELETED()` - Check if record is deleted

### Counting and Iteration

```go
// Count matching records
count, err := v.CountByExpression("ACTIVE .AND. SALARY > 40000")
fmt.Printf("Found %d active high-salary employees\n", count)

// Iterate through matches
err = v.ForEachExpressionMatch("YEAR(BIRTH_DATE) = 1985", 
    func(fieldReaders map[string]vulpo.FieldReader) error {
        nameReader := fieldReaders["NAME"]
        name, _ := nameReader.AsString()
        fmt.Printf("Born in 1985: %s\n", name)
        return nil  // Continue iteration
    })
```

### Regex Searching

For pattern-based searching on character fields:

```go
// Basic regex search
results, err := v.RegexSearch("NAME", "^Smith", &vulpo.RegexSearchOptions{
    CaseInsensitive: true,
    MaxResults: 50,
})

if err != nil {
    log.Fatal(err)
}

for _, match := range results.Matches {
    fmt.Printf("Record %d: %s (matches: %v)\n", 
        match.RecordNumber, match.FieldValue, match.Matches)
}

// Count regex matches
count, err := v.RegexCount("EMAIL", "@gmail\\.com$", nil)
fmt.Printf("Found %d Gmail addresses\n", count)

// Check if any records match
exists, err := v.RegexExists("PHONE", "\\(555\\)", nil)
if exists {
    fmt.Println("Found records with 555 area code")
}
```

## Deleted Record Handling

DBF files use a "soft delete" system where records are marked for deletion but remain in the file until physically removed.

### Working with Deleted Records

```go
// Check if current record is deleted
v.First(0)
if v.Deleted() {
    fmt.Println("Current record is marked for deletion")
}

// Mark record for deletion
err := v.Delete()
if err != nil {
    log.Printf("Failed to delete record: %v", err)
}

// Undelete (recall) a record
err = v.Recall()
if err != nil {
    log.Printf("Failed to recall record: %v", err)
}

// Physical removal (permanent)
err = v.Pack()  // WARNING: This is destructive!
if err != nil {
    log.Printf("Failed to pack database: %v", err)
}
// After pack, must reposition
v.First(0)
```

### Deleted Record Analysis

```go
// Count deleted records
deletedCount, err := v.CountDeleted()
activeCount, err := v.CountActive()

fmt.Printf("Records: %d deleted, %d active\n", deletedCount, activeCount)

// List all deleted records
deletedRecords, err := v.ListDeletedRecords()
fmt.Printf("Deleted record numbers: ")
for _, record := range deletedRecords {
    fmt.Printf("%d ", record.RecordNumber)
}
fmt.Println()

// Process each deleted record
err = v.ForEachDeletedRecord(func(recordNumber int) error {
    fmt.Printf("Processing deleted record %d\n", recordNumber)
    return nil
})

// Batch recall all deleted records
recalledCount, err := v.RecallAllDeleted()
fmt.Printf("Recalled %d records\n", recalledCount)
```

## Advanced Features

### Expression Filters

For reusable expression filtering:

```go
// Create a compiled expression filter
filter, err := v.NewExprFilter("AGE >= 18 .AND. ACTIVE")
if err != nil {
    log.Fatal(err)
}
defer filter.Free()  // Always free resources

// Use filter on different records
v.First(0)
for !v.EOF() {
    matches, err := filter.Evaluate()
    if err != nil {
        log.Fatal(err)
    }
    
    if matches {
        fmt.Printf("Record %d matches criteria\n", v.Position())
    }
    
    v.Next()
}

// Get different result types from expressions
stringResult, err := filter.EvaluateAsString()
numericResult, err := filter.EvaluateAsDouble()
```

### Field Readers

Field readers provide type-safe access to field values:

```go
v.First(0)

// Get field reader
reader := v.FieldReader("BIRTH_DATE")
if reader == nil {
    log.Fatal("Field not found")
}

// Multiple conversion options
dateValue, err := reader.AsTime()       // as time.Time
stringValue, err := reader.AsString()   // as string
intValue, err := reader.AsInt()         // as int (if convertible)

// Check for null values
if isNull, _ := reader.IsNull(); isNull {
    fmt.Println("Field is null")
}

// Get field definition
fieldDef := reader.FieldDef()
fmt.Printf("Field: %s, Type: %s, Size: %d\n", 
    fieldDef.Name(), fieldDef.Type().String(), fieldDef.Size())
```

## API Reference

### Core Types

#### Vulpo
Main database connection type.

```go
type Vulpo struct {
    // private fields
}
```

**Methods:**
- `Open(filename string) error` - Open database file
- `Close() error` - Close database and free resources
- `Active() bool` - Check if database is open

#### Header
Database header information.

```go
type Header struct {
    // Contains metadata about the DBF file
}
```

**Methods:**
- `RecordCount() uint` - Total number of records
- `LastUpdated() time.Time` - Last modification date
- `HasIndex() bool` - Whether database has associated index files
- `Codepage() Codepage` - Character encoding

#### FieldDef
Field definition containing metadata.

```go
type FieldDef struct {
    // Field metadata
}
```

**Methods:**
- `Name() string` - Field name
- `Type() FieldType` - Field type (Character, Numeric, etc.)
- `Size() uint8` - Field width in bytes
- `Decimals() uint8` - Number of decimal places (for numeric fields)

### Navigation Methods

- `Goto(recordNumber int) error` - Move to specific physical record
- `First(num int) error` - Move to first record in current order
- `Last(num int) error` - Move to last record in current order
- `Next() error` - Move to next record
- `Previous() error` - Move to previous record
- `Skip(count int) error` - Skip multiple records
- `Position() int` - Get current record number (1-indexed)
- `EOF() bool` - Check if at end of file
- `BOF() bool` - Check if at beginning of file

### Field Access Methods (v2.0 Unified API)

- `FieldCount() int` - Number of fields in database
- `Field(index int) Field` - Get field instance by index (includes both definition and reading)
- `FieldByName(name string) Field` - Get field instance by name (includes both definition and reading)
- `Fields() *Fields` - Get the complete field collection

### Field Instance Methods

Each Field instance provides unified access to both definition and reading capabilities:

**Field Definition Methods:**
- `Name() string` - Field name
- `Type() FieldType` - Field type (Character, Numeric, Date, etc.)
- `Size() uint8` - Field size in bytes
- `Decimals() uint8` - Number of decimal places (for numeric fields)
- `IsSystem() bool` - Whether this is a system field
- `IsNullable() bool` - Whether field can contain null values
- `IsBinary() bool` - Whether field contains binary data

**Field Reading Methods (operate on current record):**
- `Value() (interface{}, error)` - Get field value in its native type
- `AsString() (string, error)` - Get field value as string
- `AsInt() (int, error)` - Get field value as integer
- `AsFloat() (float64, error)` - Get field value as float64
- `AsBool() (bool, error)` - Get field value as boolean
- `AsTime() (time.Time, error)` - Get field value as time.Time
- `IsNull() (bool, error)` - Check if field value is null

### Deprecated Methods (v1.x compatibility)

- `FieldReader(name string) FieldReader` - Create field reader for current record (deprecated: use `FieldByName()`)
- `FieldReaderByIndex(index int) FieldReader` - Create field reader by index (deprecated: use `Field()`)
- `FieldDefs() *FieldDefs` - Get field definitions collection (deprecated: use `Fields()`)

### Index Methods

- `ListTags() []*Tag` - Get all available index tags
- `TagByName(name string) *Tag` - Find tag by name
- `SelectedTag() *Tag` - Get currently selected tag
- `SelectTag(tag *Tag) error` - Select tag for navigation
- `Seek(value string) (SeekResult, error)` - Seek for value in current index
- `SeekNext(value string) (SeekResult, error)` - Find next matching value

### Deleted Record Methods

- `Deleted() bool` - Check if current record is deleted
- `Delete() error` - Mark current record for deletion
- `Recall() error` - Undelete current record  
- `Pack() error` - Physically remove deleted records
- `CountDeleted() (int, error)` - Count deleted records
- `CountActive() (int, error)` - Count active records
- `ListDeletedRecords() ([]DeletedRecordInfo, error)` - List deleted record info
- `RecallAllDeleted() (int, error)` - Undelete all records

### Expression Methods

- `NewExprFilter(expression string) (*ExprFilter, error)` - Create expression filter
- `SearchByExpression(expression string, options *ExprSearchOptions) (*ExprSearchResult, error)` - Search with expression
- `CountByExpression(expression string) (int, error)` - Count matching records
- `ForEachExpressionMatch(expression string, callback func(map[string]FieldReader) error) error` - Iterate matches

### Regex Methods

- `RegexSearch(fieldName, pattern string, options *RegexSearchOptions) (*RegexSearchResult, error)` - Regex search
- `RegexCount(fieldName, pattern string, options *RegexSearchOptions) (int, error)` - Count regex matches
- `RegexExists(fieldName, pattern string, options *RegexSearchOptions) (bool, error)` - Check regex existence

## Examples

### Complete Database Processing with Unified Field API

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/mkfoss/vulpo"
)

func main() {
    // Open database - field readers are created automatically
    v := &vulpo.Vulpo{}
    err := v.Open("employees.dbf")
    if err != nil {
        log.Fatal(err)
    }
    defer v.Close()
    
    // Display database info
    header := v.Header()
    fmt.Printf("Employee Database\n")
    fmt.Printf("Total Records: %d\n", header.RecordCount())
    fmt.Printf("Last Updated: %s\n", header.LastUpdated().Format("2006-01-02"))
    
    // Show field structure using unified Field API
    fmt.Printf("\nField Structure:\n")
    for i := 0; i < v.FieldCount(); i++ {
        field := v.Field(i)  // Gets field with both definition and reading capabilities
        fmt.Printf("  %-15s %-10s %3d,%d\n", 
            field.Name(), field.Type().String(), field.Size(), field.Decimals())
    }
    
    // Find high-salary employees
    fmt.Printf("\nHigh Salary Employees (>$75,000):\n")
    results, err := v.SearchByExpression("SALARY > 75000 .AND. .NOT. DELETED()", 
        &vulpo.ExprSearchOptions{MaxResults: 10})
    if err != nil {
        log.Fatal(err)
    }
    
    for _, match := range results.Matches {
        nameField := match.FieldReaders["NAME"]
        salaryField := match.FieldReaders["SALARY"]
        deptField := match.FieldReaders["DEPARTMENT"]
        
        name, _ := nameField.AsString()
        salary, _ := salaryField.AsFloat()
        dept, _ := deptField.AsString()
        
        fmt.Printf("  %-20s $%8.2f  %s\n", name, salary, dept)
    }
    
    // Department summary using expressions
    fmt.Printf("\nDepartment Summary:\n")
    departments := []string{"SALES", "ENGINEERING", "MARKETING", "HR"}
    for _, dept := range departments {
        expression := fmt.Sprintf("TRIM(UPPER(DEPARTMENT)) = '%s' .AND. .NOT. DELETED()", dept)
        count, err := v.CountByExpression(expression)
        if err == nil {
            fmt.Printf("  %-15s: %d employees\n", dept, count)
        }
    }
    
    // Show employees hired in current year
    currentYear := time.Now().Year()
    fmt.Printf("\nEmployees Hired in %d:\n", currentYear)
    expression := fmt.Sprintf("YEAR(HIRE_DATE) = %d .AND. .NOT. DELETED()", currentYear)
    
    err = v.ForEachExpressionMatch(expression, func(fields map[string]vulpo.FieldReader) error {
        nameField := fields["NAME"]     // These are actually Field instances
        hireDateField := fields["HIRE_DATE"]
        
        name, _ := nameField.AsString()
        hireDate, _ := hireDateField.AsTime()
        
        fmt.Printf("  %-20s  %s\n", name, hireDate.Format("2006-01-02"))
        return nil
    })
    
    // Clean up deleted records if any
    deletedCount, _ := v.CountDeleted()
    if deletedCount > 0 {
        fmt.Printf("\nFound %d deleted records\n", deletedCount)
        
        // Show deleted record numbers
        deletedRecords, _ := v.ListDeletedRecords()
        fmt.Printf("Deleted record numbers: ")
        for _, record := range deletedRecords {
            fmt.Printf("%d ", record.RecordNumber)
        }
        fmt.Println()
        
        // Optionally pack database (remove deleted records permanently)
        // WARNING: This is destructive - make backups first!
        // err = v.Pack()
        // if err != nil {
        //     log.Printf("Pack failed: %v", err)
        // } else {
        //     fmt.Println("Database packed successfully")
        // }
    }
    
    fmt.Printf("\nDatabase processing complete\n")
}
```

### Working with Indexes

```go
func demonstrateIndexes(v *vulpo.Vulpo) {
    // List all available indexes
    tags := v.ListTags()
    fmt.Printf("Available indexes: %d\n", len(tags))
    for _, tag := range tags {
        fmt.Printf("  %s\n", tag.Name())
    }
    
    // Use name index for alphabetical processing
    nameTag := v.TagByName("NAME_IDX")
    if nameTag != nil {
        fmt.Println("\nProcessing records in alphabetical order:")
        v.SelectTag(nameTag)
        
        v.First(0)
        count := 0
        for !v.EOF() && count < 5 {
            nameReader := v.FieldReader("NAME")
            name, _ := nameReader.AsString()
            fmt.Printf("  %s\n", name)
            
            v.Next()
            count++
        }
        
        // Seek for specific names
        fmt.Println("\nSeeking for names starting with 'Smith':")
        result, err := v.Seek("Smith")
        if err == nil && result.IsPositioned() {
            for {
                nameReader := v.FieldReader("NAME")
                name, _ := nameReader.AsString()
                
                // Check if still matches prefix
                if len(name) < 5 || name[:5] != "Smith" {
                    break
                }
                
                fmt.Printf("  Found: %s\n", name)
                
                // Look for next match
                result, err := v.SeekNext("Smith")
                if result != vulpo.SeekSuccess || err != nil {
                    break
                }
            }
        }
    }
    
    // Return to physical record order
    v.SelectTag(nil)
    fmt.Println("\nReturned to physical record order")
}
```

## Compatibility

### DBF File Support
- **dBASE III**: Complete support
- **dBASE IV**: Complete support including memo fields
- **dBASE V**: Full compatibility
- **FoxPro**: 2.x and Visual FoxPro support
- **Clipper**: Full compatibility
- **Other xBase**: Most variants supported

### Index File Support
- **CDX**: Compound index (FoxPro/Visual FoxPro)
- **NDX**: Single index (dBASE III)
- **NTX**: Clipper index format
- **MDX**: Multiple index (dBASE IV)

### Field Type Compatibility
All standard DBF field types are supported with proper type conversion and validation.

### Platform Support
- **Linux**: Full support (primary platform)
- **macOS**: Full support
- **Windows**: Full support with CGO
- **FreeBSD/OpenBSD**: Should work (untested)

### Thread Safety
- **Read Operations**: Thread-safe for multiple concurrent readers
- **Write Operations**: Require external synchronization
- **Database Handles**: Not thread-safe - use one Vulpo instance per goroutine

## Error Handling

Vulpo provides comprehensive error reporting:

```go
// Always check for errors
err := v.Open("nonexistent.dbf")
if err != nil {
    log.Printf("Open failed: %v", err)
}

// Navigation errors
err = v.Goto(999999)
if err != nil {
    log.Printf("Invalid record number: %v", err)
}

// Expression errors
_, err = v.SearchByExpression("INVALID_SYNTAX((", nil)
if err != nil {
    log.Printf("Expression parse error: %v", err)
}
```

## Migration from Previous Versions

### Field API Changes

Starting with v2.0, Vulpo has simplified field access by automatically creating field readers when opening a database. This eliminates the need to manually create field readers for each field.

**Old API (deprecated but still works):**
```go
// Old way - manual field reader creation
fieldReader := v.FieldReader("NAME")
value, err := fieldReader.AsString()

// Getting field definition separately
fieldDef := v.FieldByName("NAME")
fmt.Printf("Field type: %s\n", fieldDef.Type().String())
```

**New API (recommended):**
```go
// New way - unified field access
field := v.FieldByName("NAME")
value, err := field.AsString()

// Field definition is included in the same object
fmt.Printf("Field type: %s\n", field.Type().String())
```

### Key Changes:

1. **Automatic Field Reader Creation**: Field readers are created automatically at `Open()` time
2. **Unified Interface**: `FieldByName()` and `Field()` now return `Field` interface that includes both definition and reading capabilities
3. **No Manual Reader Creation**: No need to call `FieldReader()` or `FieldReaderByIndex()` 
4. **Backward Compatibility**: Old `FieldReader()` methods are deprecated but still work
5. **Performance**: Better performance since field readers are created once at open time

### Migration Steps:

1. Replace `v.FieldReader(name)` calls with `v.FieldByName(name)`
2. Replace `v.FieldReaderByIndex(idx)` calls with `v.Field(idx)`
3. Access field metadata directly from the field object instead of getting separate `FieldDef`
4. Remove any code that manually creates or caches field readers

## Best Practices

1. **Always Close**: Use `defer v.Close()` immediately after successful `Open()`
2. **Check Active**: Verify `v.Active()` before database operations in long-running code
3. **Handle EOF/BOF**: Always check `v.EOF()` and `v.BOF()` during navigation
4. **Index Awareness**: Remember that selected indexes affect navigation order
5. **Expression Performance**: Use indexes when possible for expression searches
6. **Memory Management**: Free expression filters with `defer filter.Free()`
7. **Backup Before Pack**: Pack operations are destructive - backup first
8. **Error Handling**: Always check and handle errors appropriately

## Contributing

Contributions are welcome! Please see the contribution guidelines for details on:
- Code style and formatting
- Test requirements
- Documentation standards
- Pull request process

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For questions, issues, or feature requests, please:
1. Check the documentation and examples
2. Search existing GitHub issues
3. Create a new issue with detailed information
4. Include a minimal reproducible example when reporting bugs

---

*Vulpo - Professional DBF file handling for Go applications*