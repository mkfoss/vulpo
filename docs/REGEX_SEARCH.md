# Regex Search Functionality

This document describes the regular expression search capabilities added to the vulpo DBF library.

## Overview

The regex search functionality provides powerful pattern matching capabilities for character/string fields in DBF files. While DBF indexes only support exact matches and prefix matching, this implementation adds full regex support with optional performance optimizations.

## Key Features

- **Full Regex Support**: Use standard Go regex patterns to search character fields
- **Index Optimization**: Automatic optimization for simple prefix patterns (e.g., `^ABC.*`)
- **Case-Insensitive Search**: Built-in case-insensitive matching option
- **Result Limiting**: Control the maximum number of results returned
- **Performance Monitoring**: Track how many records were scanned vs matched
- **Convenience Methods**: `RegexCount()` and `RegexExists()` for common use cases

## Types

### RegexSearchOptions

Configure regex search behavior:

```go
type RegexSearchOptions struct {
    CaseInsensitive bool   // Make pattern case-insensitive
    MaxResults      int    // Limit number of results (0 = unlimited)
    UseIndex        bool   // Try to optimize with index when possible
    IndexField      string // Field to use for index optimization
}
```

### RegexMatch

Represents a single match result:

```go
type RegexMatch struct {
    RecordNumber int               // 1-indexed record number
    FieldValue   string            // The field value that matched
    Matches      [][]int           // Byte indices of regexp matches
    FieldReader  FieldReader       // Field reader for accessing the record
}
```

### RegexSearchResult

Contains complete search results:

```go
type RegexSearchResult struct {
    Pattern      string        // The regex pattern used
    Matches      []RegexMatch  // All matching records
    TotalScanned int          // Total records scanned
    TotalMatched int          // Total records that matched
}
```

## Basic Usage

### Simple Pattern Matching

```go
v := &vulpo.Vulpo{}
err := v.Open("customers.dbf")
if err != nil {
    panic(err)
}
defer v.Close()

// Find all customers with names containing "smith" (case-insensitive)
result, err := v.RegexSearch("customer_name", "(?i)smith", nil)
if err != nil {
    panic(err)
}

fmt.Printf("Found %d matches out of %d records scanned\n", 
    result.TotalMatched, result.TotalScanned)

for _, match := range result.Matches {
    fmt.Printf("Record %d: %s\n", match.RecordNumber, match.FieldValue)
}
```

### Advanced Pattern Matching

```go
// Email validation pattern
emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
result, err := v.RegexSearch("email", emailPattern, nil)
if err != nil {
    panic(err)
}

// Phone number patterns
phonePattern := `^\(\d{3}\)\s\d{3}-\d{4}$` // (555) 123-4567 format
result, err = v.RegexSearch("phone", phonePattern, nil)

// Find all records with postal codes starting with specific digits
zipPattern := `^9[0-4]\d{3}$` // California ZIP codes
result, err = v.RegexSearch("zip_code", zipPattern, nil)
```

### Using Search Options

```go
options := &vulpo.RegexSearchOptions{
    CaseInsensitive: true,    // Case-insensitive search
    MaxResults:      100,     // Limit to 100 matches
    UseIndex:        true,    // Try to use index optimization
}

result, err := v.RegexSearch("company_name", "^tech.*corp", options)
if err != nil {
    panic(err)
}
```

## Convenience Methods

### RegexCount - Count Matches Only

```go
// Count how many customers have Gmail addresses
count, err := v.RegexCount("email", "@gmail\\.com$", &vulpo.RegexSearchOptions{
    CaseInsensitive: true,
})
if err != nil {
    panic(err)
}
fmt.Printf("Found %d Gmail users\n", count)
```

### RegexExists - Check for Any Matches

```go
// Check if any records have international phone numbers
hasInternational, err := v.RegexExists("phone", "^\\+", nil)
if err != nil {
    panic(err)
}

if hasInternational {
    fmt.Println("Database contains international phone numbers")
}
```

## Performance Optimization

### Index Optimization for Prefix Patterns

The library automatically optimizes simple prefix patterns using existing indexes:

```go
// These patterns can be optimized if an index exists on the field:
result, err := v.RegexSearch("last_name", "^SMITH.*", nil)    // Optimized
result, err := v.RegexSearch("company", "^TECH", nil)        // Optimized
result, err := v.RegexSearch("product", "^A[A-Z].*", nil)    // Optimized

// These patterns require full table scans:
result, err := v.RegexSearch("name", ".*SMITH.*", nil)       // Full scan
result, err := v.RegexSearch("email", "@gmail\\.com$", nil)  // Full scan
```

### Controlling Index Usage

```go
// Force full table scan (disable index optimization)
options := &vulpo.RegexSearchOptions{UseIndex: false}
result, err := v.RegexSearch("name", "^SMITH.*", options)

// Enable index optimization (default)
options = &vulpo.RegexSearchOptions{UseIndex: true}
result, err = v.RegexSearch("name", "^SMITH.*", options)
```

### Performance Monitoring

```go
result, err := v.RegexSearch("description", "urgent|priority|asap", nil)
if err != nil {
    panic(err)
}

fmt.Printf("Efficiency: Found %d matches by scanning %d records (%.1f%%)\n",
    result.TotalMatched, result.TotalScanned,
    float64(result.TotalMatched)/float64(result.TotalScanned)*100)
```

## Working with Results

### Processing Matches

```go
result, err := v.RegexSearch("notes", "bug|error|issue", &vulpo.RegexSearchOptions{
    CaseInsensitive: true,
})
if err != nil {
    panic(err)
}

for _, match := range result.Matches {
    // Get detailed submatch information
    regex := regexp.MustCompile("(?i)bug|error|issue")
    submatches := match.GetSubmatches(regex)
    
    fmt.Printf("Record %d found: %v in '%s'\n", 
        match.RecordNumber, submatches, match.FieldValue)
    
    // Navigate to the matching record for further processing
    err := match.GetRecord(v)
    if err != nil {
        continue
    }
    
    // Now you can read other fields from this record
    // using v.FieldReader("other_field_name")
}
```

### Extracting Match Details

```go
result, err := v.RegexSearch("product_code", "([A-Z]{2})-([0-9]{4})", nil)
if err != nil {
    panic(err)
}

regex := regexp.MustCompile("([A-Z]{2})-([0-9]{4})")
for _, match := range result.Matches {
    // Get full matches and subgroups
    fullMatches := regex.FindAllStringSubmatch(match.FieldValue, -1)
    for _, fullMatch := range fullMatches {
        if len(fullMatch) >= 3 {
            category := fullMatch[1]  // First capture group
            number := fullMatch[2]    // Second capture group
            fmt.Printf("Product: Category=%s, Number=%s\n", category, number)
        }
    }
}
```

## Common Patterns

### Email Addresses
```go
emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
```

### Phone Numbers
```go
// US Phone: (555) 123-4567
usPhonePattern := `^\(\d{3}\)\s\d{3}-\d{4}$`

// International: +1-555-123-4567
intlPhonePattern := `^\+\d{1,3}-\d{3}-\d{3}-\d{4}$`
```

### Postal Codes
```go
// US ZIP: 12345 or 12345-6789
zipPattern := `^\d{5}(-\d{4})?$`

// Canadian: A1A 1A1
canadaPattern := `^[A-Za-z]\d[A-Za-z]\s\d[A-Za-z]\d$`
```

### Product Codes
```go
// Format: ABC-1234
productPattern := `^[A-Z]{3}-\d{4}$`

// Format: Category-SubCat-Number
complexPattern := `^([A-Z]+)-([A-Z]+)-(\d+)$`
```

### Date Formats (in text fields)
```go
// ISO Date: 2023-12-01
isoDatePattern := `^\d{4}-\d{2}-\d{2}$`

// US Date: 12/01/2023
usDatePattern := `^\d{1,2}/\d{1,2}/\d{4}$`
```

## Complete Example

```go
package main

import (
    "fmt"
    "regexp"
    "github.com/mkfoss/vulpo"
)

func main() {
    v := &vulpo.Vulpo{}
    err := v.Open("customer_data.dbf")
    if err != nil {
        panic(err)
    }
    defer v.Close()
    
    // Find customers with suspicious email patterns
    suspiciousEmails := []string{
        `\.tk$`,        // .tk domains
        `\.ml$`,        // .ml domains  
        `\d{4,}@`,      // Numbers at start of email
        `^[a-z]\@`,     // Single character before @
    }
    
    fmt.Println("Checking for suspicious email patterns...")
    
    for i, pattern := range suspiciousEmails {
        count, err := v.RegexCount("email", pattern, &vulpo.RegexSearchOptions{
            CaseInsensitive: true,
        })
        if err != nil {
            fmt.Printf("Error checking pattern %d: %v\n", i+1, err)
            continue
        }
        
        if count > 0 {
            fmt.Printf("Pattern %d: Found %d suspicious emails\n", i+1, count)
            
            // Get details for first few matches
            result, err := v.RegexSearch("email", pattern, &vulpo.RegexSearchOptions{
                CaseInsensitive: true,
                MaxResults:      5,
            })
            if err != nil {
                continue
            }
            
            for _, match := range result.Matches {
                fmt.Printf("  Record %d: %s\n", match.RecordNumber, match.FieldValue)
            }
        }
    }
    
    // Find all phone numbers and categorize them
    phoneResult, err := v.RegexSearch("phone", ".+", nil) // Any non-empty phone
    if err != nil {
        panic(err)
    }
    
    usPattern := regexp.MustCompile(`^\(\d{3}\)\s\d{3}-\d{4}$`)
    intlPattern := regexp.MustCompile(`^\+`)
    
    usCount := 0
    intlCount := 0
    otherCount := 0
    
    for _, match := range phoneResult.Matches {
        if usPattern.MatchString(match.FieldValue) {
            usCount++
        } else if intlPattern.MatchString(match.FieldValue) {
            intlCount++
        } else {
            otherCount++
        }
    }
    
    fmt.Printf("\nPhone number analysis:\n")
    fmt.Printf("  US Format: %d\n", usCount)
    fmt.Printf("  International: %d\n", intlCount) 
    fmt.Printf("  Other/Invalid: %d\n", otherCount)
}
```

## Error Handling

```go
result, err := v.RegexSearch("field_name", "pattern", nil)
if err != nil {
    // Common errors:
    // - Database not open
    // - Field not found
    // - Field is not a character type
    // - Invalid regex pattern
    fmt.Printf("Regex search failed: %v\n", err)
    return
}

// Check if any matches were found
if result.TotalMatched == 0 {
    fmt.Println("No matches found")
} else {
    fmt.Printf("Found %d matches\n", result.TotalMatched)
}
```

## Performance Tips

1. **Use Index Optimization**: Simple prefix patterns (`^ABC.*`) are much faster when indexes exist
2. **Limit Results**: Use `MaxResults` for large datasets when you only need a few examples  
3. **Use RegexExists**: When you only need to know if any matches exist
4. **Use RegexCount**: When you only need the count, not the actual results
5. **Optimize Patterns**: More specific patterns reduce false positives and improve performance
6. **Profile Your Queries**: Monitor `TotalScanned` vs `TotalMatched` to understand efficiency

## Limitations

1. **Character Fields Only**: Regex search only works on character/string fields
2. **No Native Index Support**: Complex patterns require full table scans
3. **Memory Usage**: Large result sets consume memory proportional to the number of matches
4. **Pattern Complexity**: Very complex regex patterns can be slow on large datasets

The regex search functionality provides a powerful way to find complex patterns in your DBF data while maintaining good performance through intelligent optimizations.