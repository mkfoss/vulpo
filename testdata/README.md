# Test Data for Kitsune FoxPro Library

This directory contains Visual FoxPro test files copied from the original redfox project. These files provide comprehensive test coverage for various FoxPro features and edge cases.

## File Descriptions

### Basic Test Files

- **`empty.dbf`** - Empty DBF file for testing basic file structure
- **`simpletwofld.dbf`** - Simple two-field table for basic functionality testing  
- **`fieldy.dbf`** - Basic field testing
- **`charcasesinflds.dbf`** - Tests character case handling in field names
- **`decimalfld.dbf`** - Tests decimal field handling
- **`threenames.dbf`** - Three-field table with names
- **`threenames-null.dbf`** - Same as above but with null values
- **`unknownvarbinary.dbf`** - Tests handling of unknown/varbinary field types

### Composite Data Files

- **`idcharsdate.dbf`** - Tests ID, character, and date fields together
- **`idfirstlast18.dbf`** - Name table with 18 records (ID, first name, last name)
- **`intcharsnumeric.dbf`** - Tests integer, character, and numeric field combinations
- **`logictime.dbf`** - Tests logical (boolean) and time field types

### Memo File Tests

- **`basicmemo.dbf` + `basicmemo.fpt`** - Basic memo field testing
- **`memopic.dbf` + `memopic.fpt`** - Memo fields with picture/binary data

### Field-Specific Tests (`fieldtests/`)

- **`bools.dbf`** - Boolean/Logical field type tests
- **`currencies.dbf`** - Currency field type tests  
- **`dates.dbf`** - Date field type tests
- **`datetimes.dbf`** - DateTime field type tests
- **`deleteds.dbf`** - Tests deleted record handling
- **`integers.dbf`** - Integer field type tests
- **`memos.dbf` + `memos.fpt`** - Memo field tests
- **`numerics.dbf`** - Numeric field type tests

## FoxPro Field Types Covered

The test data covers these Visual FoxPro field types:

- **C** - Character/String fields
- **N** - Numeric fields  
- **D** - Date fields
- **L** - Logical/Boolean fields
- **I** - Integer fields
- **Y** - Currency fields
- **T** - DateTime fields
- **M** - Memo fields (with corresponding .fpt files)

## Usage in Tests

These files can be used to create comprehensive tests for:

1. **File format parsing** - Ensure headers and field definitions are read correctly
2. **Data type handling** - Verify each field type is parsed and converted properly
3. **Navigation** - Test record traversal, positioning, and boundary conditions
4. **Edge cases** - Null values, deleted records, empty files
5. **Memo files** - Binary data storage and retrieval from .fpt files

## File Format Notes

- **`.dbf`** files contain the main database structure and data
- **`.fpt`** files contain memo field data (binary blobs, long text)
- All files are in Visual FoxPro format (magic byte 0x30)
- Files use little-endian byte ordering
- Character encoding varies by codepage (stored in header)