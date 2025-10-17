# Index Search Functionality

This document describes the index search capabilities added to the vulpo DBF library.

## Overview

The index search functionality allows you to perform fast, indexed lookups in DBF files that have associated index files (CDX, NDX, etc.). This is much more efficient than sequential scanning for specific records.

## Key Types

### Tag

The `Tag` type represents an index tag (a single index within an index file):

```go
type Tag struct {
    name   string
    tagPtr *C.TAG4
}

// Methods
func (t *Tag) Name() string      // Returns the tag name
func (t *Tag) IsValid() bool     // Returns true if the tag is valid
```

### SeekResult

The `SeekResult` type represents the outcome of a search operation:

```go
type SeekResult int

const (
    SeekSuccess  SeekResult = iota // Found exact match
    SeekAfter                      // Not found, positioned after where it would be
    SeekEOF                        // Not found, positioned at EOF
    SeekEntry                      // Record didn't exist (CODE4.errGo is false)
    SeekLocked                     // Lock failed
    SeekUnique                     // Duplicate key in unique index
    SeekNoTag                      // No tag available
    SeekError                      // Other error
)

// Methods
func (sr SeekResult) String() string      // String representation
func (sr SeekResult) IsFound() bool       // True if exact match found
func (sr SeekResult) IsPositioned() bool  // True if cursor positioned somewhere
```

## Tag Management

### Listing Available Tags

```go
v := &vulpo.Vulpo{}
err := v.Open("myfile.dbf")
if err != nil {
    panic(err)
}
defer v.Close()

// Get all available tags
tags := v.ListTags()
for _, tag := range tags {
    fmt.Printf("Tag: %s\n", tag.Name())
}

// Get just the tag names
names := v.TagNames()
fmt.Printf("Available indexes: %v\n", names)

// Check if a specific tag exists
if v.HasTag("NAME_INDEX") {
    fmt.Println("NAME_INDEX exists")
}

// Get tag count
count := v.TagCount()
fmt.Printf("Total indexes: %d\n", count)
```

### Finding Specific Tags

```go
// Find a tag by name
nameTag := v.TagByName("NAME_INDEX")
if nameTag != nil {
    fmt.Printf("Found tag: %s\n", nameTag.Name())
}

// Get the default tag
defaultTag := v.DefaultTag()
if defaultTag != nil {
    fmt.Printf("Default tag: %s\n", defaultTag.Name())
}

// Get the currently selected tag
selectedTag := v.SelectedTag()
if selectedTag != nil {
    fmt.Printf("Currently selected: %s\n", selectedTag.Name())
}
```

### Selecting Tags

```go
// Select a specific tag for subsequent operations
nameTag := v.TagByName("NAME_INDEX")
if nameTag != nil {
    err := v.SelectTag(nameTag)
    if err != nil {
        panic(err)
    }
}

// Select record number ordering (no index)
err := v.SelectTag(nil)
if err != nil {
    panic(err)
}
```

## Searching

### Basic Search Operations

```go
// Search using the currently selected tag
result, err := v.Seek("SMITH")
if err != nil {
    panic(err)
}

switch result {
case vulpo.SeekSuccess:
    fmt.Println("Found exact match!")
case vulpo.SeekAfter:
    fmt.Println("Not found, positioned after")
case vulpo.SeekEOF:
    fmt.Println("Not found, at end of file")
default:
    fmt.Printf("Search result: %s\n", result.String())
}

// Search for numeric values (more efficient for numeric indexes)
result, err = v.SeekDouble(12345.67)
if err != nil {
    panic(err)
}

// Continue searching for next matching record
result, err = v.SeekNext("SMITH")
if err != nil {
    panic(err)
}

result, err = v.SeekNextDouble(12345.67)
if err != nil {
    panic(err)
}
```

### Search With Specific Tag

The `SeekWithTag` and `SeekDoubleWithTag` methods automatically select a tag, perform the search, and restore the original tag selection:

```go
nameTag := v.TagByName("NAME_INDEX")
if nameTag != nil {
    // Temporarily use NAME_INDEX for this search
    result, err := v.SeekWithTag(nameTag, "SMITH")
    if err != nil {
        panic(err)
    }
    
    if result.IsFound() {
        fmt.Println("Found Smith in name index!")
    }
    // Original tag selection is automatically restored
}

// Same for numeric searches
ageTag := v.TagByName("AGE_INDEX")
if ageTag != nil {
    result, err := v.SeekDoubleWithTag(ageTag, 25.0)
    if err != nil {
        panic(err)
    }
    
    if result.IsPositioned() {
        fmt.Println("Found or positioned near age 25")
    }
}
```

## Search Value Formats

Different tag types require different search value formats:

### Character Tags
```go
// Partial matches are allowed
result, _ := v.Seek("SMI")        // Finds first key starting with "SMI"
result, _ := v.Seek("SMITH")      // Exact match for "SMITH"
```

### Date Tags
```go
// Format: "CCYYMMDD"
result, _ := v.Seek("20231201")   // December 1, 2023
```

### DateTime Tags
```go
// Format: "CCYYMMDDhh:mm:ss:ttt"
result, _ := v.Seek("20231201123045000")  // Dec 1, 2023 12:30:45.000
```

### Numeric/Float/Currency Tags
```go
// String format
result, _ := v.Seek("123.45")
// Or use SeekDouble for better performance
result, _ := v.SeekDouble(123.45)
```

## Complete Example

```go
package main

import (
    "fmt"
    "github.com/mkfoss/vulpo"
)

func main() {
    v := &vulpo.Vulpo{}
    err := v.Open("customers.dbf")
    if err != nil {
        panic(err)
    }
    defer v.Close()
    
    // List available indexes
    fmt.Println("Available indexes:")
    for _, name := range v.TagNames() {
        fmt.Printf("  - %s\n", name)
    }
    
    // Search by customer name
    nameTag := v.TagByName("CUST_NAME")
    if nameTag == nil {
        fmt.Println("No CUST_NAME index available")
        return
    }
    
    // Search for all customers starting with "SMITH"
    result, err := v.SeekWithTag(nameTag, "SMITH")
    if err != nil {
        panic(err)
    }
    
    if result.IsPositioned() {
        fmt.Println("Found customer(s) starting with SMITH:")
        
        // Process current record and continue searching
        for result.IsPositioned() && !v.EOF() {
            // Read customer data here
            // ... (use field readers to get customer data)
            
            // Look for next SMITH
            result, err = v.SeekNext("SMITH")
            if err != nil {
                panic(err)
            }
        }
    } else {
        fmt.Println("No customers found starting with SMITH")
    }
    
    // Search by customer ID (numeric)
    idTag := v.TagByName("CUST_ID")
    if idTag != nil {
        result, err := v.SeekDoubleWithTag(idTag, 12345)
        if err != nil {
            panic(err)
        }
        
        if result.IsFound() {
            fmt.Println("Found customer ID 12345")
            // Process this customer
        } else {
            fmt.Println("Customer ID 12345 not found")
        }
    }
}
```

## Error Handling

The CodeBase library may log errors to stderr for certain operations (like searching for non-existent tags). These are typically informational and don't affect the Go-level error handling. The Go functions will return appropriate error values when operations fail.

Common scenarios:
- Searching with no selected tag returns `SeekNoTag`
- Searching for non-existent values returns `SeekAfter` or `SeekEOF`
- Invalid tag operations return Go errors
- CodeBase library state errors may cause close operations to return errors (these can usually be ignored)

## Performance Notes

1. **Use SeekDouble for numeric searches** - More efficient than string conversion
2. **Use SeekWithTag for one-off searches** - Avoids manual tag selection management  
3. **Use SeekNext to continue searches** - More efficient than repeated Seek calls
4. **Select appropriate tags** - Choose the most selective index for your query

## Limitations

1. Only works with DBF files that have associated index files (CDX, NDX, etc.)
2. Search value formats must match the tag type exactly
3. Some CodeBase library errors may be logged to stderr but don't affect Go error handling
4. Index operations may leave the CodeBase library in a state that causes close errors (these are usually harmless)