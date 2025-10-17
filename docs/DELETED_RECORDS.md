# Deleted Record Handling in Vulpo

The vulpo library provides comprehensive support for DBF deleted record handling, following standard dBASE/xBase conventions. In DBF files, records are not immediately removed when "deleted" - they are marked for deletion and can be physically removed later with a pack operation.

## Overview

DBF files use a "soft delete" system:
1. **Delete** - Records are marked for deletion (first byte becomes '*')
2. **Recall** - Deletion marks can be removed ("undelete")
3. **Pack** - Physically removes marked records from the file

## Core Functions

### Checking Deletion Status

```go
// Check if current record is marked for deletion
isDeleted := vulpo.Deleted()        // or vulpo.IsDeleted()

// Example usage
vulpo.First(0)
if vulpo.Deleted() {
    fmt.Println("First record is marked for deletion")
}
```

### Marking Records for Deletion

```go
// Mark current record for deletion
err := vulpo.Delete()
if err != nil {
    // Handle error (e.g., no current record, database not open)
}

// The record is now marked but still exists in the file
```

### Recalling (Undeleting) Records

```go
// Remove deletion mark from current record
err := vulpo.Recall()
if err != nil {
    // Handle error
}

// The record is now "undeleted" and active again
```

### Physical Removal (Pack)

```go
// Permanently remove all deleted records from the file
// WARNING: This is destructive and cannot be undone
err := vulpo.Pack()
if err != nil {
    // Handle error (e.g., locking issues, disk space)
}

// After pack, position is undefined - call a positioning function
vulpo.First(0)
```

## Counting and Analysis

### Count Records by Status

```go
// Count deleted records
deletedCount, err := vulpo.CountDeleted()
if err != nil {
    // Handle error
}
fmt.Printf("Found %d deleted records\n", deletedCount)

// Count active (non-deleted) records  
activeCount, err := vulpo.CountActive()
if err != nil {
    // Handle error
}
fmt.Printf("Found %d active records\n", activeCount)
```

### List All Deleted Records

```go
// Get information about all deleted records
deletedRecords, err := vulpo.ListDeletedRecords()
if err != nil {
    // Handle error
}

for _, record := range deletedRecords {
    fmt.Printf("Record %d is deleted\n", record.RecordNumber)
}
```

### Iterate Through Deleted Records

```go
// Process each deleted record with a callback
err := vulpo.ForEachDeletedRecord(func(recordNumber int) error {
    fmt.Printf("Processing deleted record %d\n", recordNumber)
    
    // You can position to the record if needed:
    // vulpo.Goto(recordNumber)
    
    return nil // Continue iteration
})

if err != nil {
    // Handle error
}
```

## Batch Operations

### Recall All Deleted Records

```go
// "Undelete" all records marked for deletion
recalledCount, err := vulpo.RecallAllDeleted()
if err != nil {
    // Handle error
}
fmt.Printf("Recalled %d deleted records\n", recalledCount)
```

## Complete Workflow Example

```go
package main

import (
    "fmt"
    "log"
)

func main() {
    vulpo := &Vulpo{}
    err := vulpo.Open("mydata.dbf")
    if err != nil {
        log.Fatal(err)
    }
    defer vulpo.Close()
    
    // Check initial state
    totalRecords := vulpo.Header().recordcount
    deletedCount, _ := vulpo.CountDeleted()
    activeCount, _ := vulpo.CountActive()
    
    fmt.Printf("Total records: %d\n", totalRecords)
    fmt.Printf("Deleted: %d, Active: %d\n", deletedCount, activeCount)
    
    // Mark some records for deletion
    vulpo.First(0)
    for i := 0; i < 3 && !vulpo.EOF(); i++ {
        if !vulpo.Deleted() { // Don't double-delete
            err := vulpo.Delete()
            if err != nil {
                log.Printf("Failed to delete record: %v", err)
            }
        }
        vulpo.Next()
    }
    
    // Check deletion status
    deletedCount, _ = vulpo.CountDeleted()
    fmt.Printf("Records marked for deletion: %d\n", deletedCount)
    
    // List deleted records
    deletedRecords, err := vulpo.ListDeletedRecords()
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Deleted record numbers:")
    for _, record := range deletedRecords {
        fmt.Printf("  Record %d\n", record.RecordNumber)
    }
    
    // Option 1: Recall (undelete) all deleted records
    recalledCount, err := vulpo.RecallAllDeleted()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Recalled %d records\n", recalledCount)
    
    // Option 2: Or permanently remove deleted records
    // WARNING: This is destructive!
    // err = vulpo.Pack()
    // if err != nil {
    //     log.Fatal(err)
    // }
    // fmt.Println("Deleted records permanently removed")
}
```

## Best Practices

### 1. **Always Check Errors**
```go
err := vulpo.Delete()
if err != nil {
    // Handle the error - don't ignore it
}
```

### 2. **Backup Before Packing**
```go
// Pack is destructive and cannot be undone
// Always backup your data first!
err := vulpo.Pack()
```

### 3. **Position After Pack**
```go
err := vulpo.Pack()
if err != nil {
    return err
}
// Position is undefined after pack - must reposition
vulpo.First(0)
```

### 4. **Use Exclusive Access for Pack**
```go
// For best performance and data integrity,
// ensure exclusive access when packing
// (This depends on your application's file opening strategy)
```

### 5. **Handle EOF/BOF Gracefully**
```go
if vulpo.EOF() || vulpo.BOF() {
    fmt.Println("No current record to delete")
    return
}
err := vulpo.Delete()
```

## Integration with Navigation and Indexing

### Deleted Records and Navigation
- **Navigation functions** (`Next()`, `Previous()`, etc.) **include** deleted records
- Deleted records are **not skipped** during navigation
- Use `vulpo.Deleted()` to check if the current record is deleted

### Position Preservation
- Analysis functions (`CountDeleted()`, `ListDeletedRecords()`, etc.) **preserve** your current position
- They save and restore both position and selected tag
- You can safely call them without affecting your current location

### Index Interaction
- Deleted records remain in indexes until physically removed with `Pack()`
- Seek operations can find deleted records
- Pack automatically rebuilds all indexes after removing records

## Error Conditions

Common error scenarios:
- **Database not open**: All functions check if database is active
- **No current record**: Delete/Recall operations require a valid current record
- **EOF/BOF position**: Cannot delete/recall when at end or beginning of file
- **Lock failures**: Pack operations may fail due to file locking issues
- **Disk space**: Pack operations require sufficient disk space for reorganization

## Performance Considerations

- **Counting operations** scan the entire database
- **Pack operations** are I/O intensive and should be done during maintenance windows
- **Large databases** may take significant time for counting and pack operations
- **Index rebuilding** during pack can be time-consuming for databases with many indexes

## Compatibility

This implementation follows standard dBASE/xBase deleted record conventions and is compatible with:
- dBASE III, IV, V
- FoxPro
- Clipper
- Other xBase variants

The deleted record functionality integrates seamlessly with all other vulpo features including navigation, indexing, field access, and expression filtering.