# dBASE Expression Filtering in Vulpo

The vulpo library now supports dBASE expression-based filtering, allowing you to use the full power of dBASE expressions to query and filter records in DBF files.

## Overview

dBASE expressions are a powerful query language that supports:
- Field references (e.g., `NAME`, `AGE`, `BIRTH_DATE`)
- Mathematical operations (`+`, `-`, `*`, `/`)
- Logical operations (`.AND.`, `.OR.`, `.NOT.`)
- Comparison operators (`=`, `>`, `<`, `>=`, `<=`, `!=`)
- String functions (`SUBSTR()`, `UPPER()`, `LEFT()`, `RIGHT()`, `TRIM()`, etc.)
- Date functions (`YEAR()`, `MONTH()`, `DAY()`, `CTOD()`, `DTOS()`)
- Numeric functions (`STR()`, `VAL()`)
- Conditional functions (`IIF()`)
- Record functions (`RECNO()`, `RECCOUNT()`, `DELETED()`)

## API Functions

### Creating Expression Filters

```go
// Create a compiled expression filter
filter, err := vulpo.NewExprFilter("NAME = 'John Smith'")
if err != nil {
    // Handle parsing error
}
defer filter.Free() // Always free resources

// Evaluate the expression for the current record
matches, err := filter.Evaluate()
if err != nil {
    // Handle evaluation error
}

// Get different result types
stringResult, _ := filter.EvaluateAsString()
numericResult, _ := filter.EvaluateAsDouble()
```

### Searching Records

```go
// Search for records matching an expression
options := &ExprSearchOptions{
    MaxResults: 100,    // Limit results (0 = unlimited)
    UseIndex:   true,   // Try to use indexes for optimization
}

result, err := vulpo.SearchByExpression("AGE > 30 .AND. ACTIVE", options)
if err != nil {
    // Handle error
}

// Process results
for _, match := range result.Matches {
    recordNum := match.RecordNumber
    fieldReaders := match.FieldReaders
    
    // Access field values
    if nameReader, exists := fieldReaders["NAME"]; exists {
        name, _ := nameReader.AsString()
        fmt.Printf("Record %d: %s\n", recordNum, name)
    }
}
```

### Counting Matches

```go
// Count records matching an expression
count, err := vulpo.CountByExpression("SALARY > 50000")
if err != nil {
    // Handle error
}
fmt.Printf("Found %d high-salary records\n", count)
```

### Iterating Through Matches

```go
// Iterate through matching records with a callback
err := vulpo.ForEachExpressionMatch("YEAR(BIRTH_DATE) = 1990", func(fieldReaders map[string]FieldReader) error {
    if nameReader, exists := fieldReaders["NAME"]; exists {
        name, _ := nameReader.AsString()
        fmt.Printf("Born in 1990: %s\n", name)
    }
    return nil // Continue iteration
})
```

## Expression Examples

### Basic Field Matching
```
NAME = 'John Smith'
AGE = 25
ACTIVE        // Boolean field check
```

### Numeric Comparisons
```
AGE > 30
SALARY >= 50000.00
BALANCE < 0
```

### String Operations
```
SUBSTR(NAME, 1, 3) = 'ABC'      // First 3 characters
UPPER(LEFT(NAME, 1)) = 'J'      // First letter uppercase
TRIM(DESCRIPTION) != ''         // Non-empty after trimming
```

### Date Operations
```
YEAR(BIRTH_DATE) = 1990         // Born in 1990
MONTH(HIRE_DATE) = 12           // Hired in December
BIRTH_DATE > CTOD('01/01/1980') // Born after 1980
```

### Complex Expressions
```
AGE >= 18 .AND. AGE <= 65 .AND. ACTIVE
(SALARY > 50000 .OR. BONUS > 10000) .AND. .NOT. DELETED()
IIF(AGE >= 65, 'Senior', IIF(AGE >= 18, 'Adult', 'Minor'))
```

### Record-Level Functions
```
RECNO() = 1              // First record
RECCOUNT() > 1000        // Large table
.NOT. DELETED()          // Non-deleted records
```

## Best Practices

1. **Always Free Filters**: Expression filters allocate C memory that must be freed:
   ```go
   filter, err := vulpo.NewExprFilter("expression")
   if err != nil {
       return err
   }
   defer filter.Free() // Critical!
   ```

2. **Handle Parse Errors**: Invalid expressions will return parsing errors:
   ```go
   filter, err := vulpo.NewExprFilter("INVALID_SYNTAX((")
   if err != nil {
       // Expression syntax error
       return err
   }
   ```

3. **Use Appropriate Result Types**: Choose the right evaluation method:
   - `Evaluate()` for boolean/logical expressions
   - `EvaluateAsString()` for string expressions
   - `EvaluateAsDouble()` for numeric expressions

4. **Optimize with Indexes**: When possible, structure expressions to take advantage of existing indexes.

5. **Position Restoration**: Search functions automatically restore the original record position after completion.

## Limitations

- Expression parsing depends on the underlying CodeBase library capabilities
- Some advanced dBASE functions may not be available depending on the library version
- Index optimization is limited and depends on available indexes
- Error messages from the C library may be limited

## Error Handling

The expression system provides several types of errors:
- **Parse Errors**: Invalid expression syntax
- **Evaluation Errors**: Runtime errors during expression evaluation  
- **Database Errors**: Database not open or navigation failures
- **Field Errors**: Unknown field names or type mismatches

Always check for errors when creating filters and evaluating expressions.