package vulpo

import (
	"fmt"
	"regexp"
)

// RegexSearchOptions configures regex search behavior
type RegexSearchOptions struct {
	CaseInsensitive bool   // Make pattern case-insensitive
	MaxResults      int    // Limit number of results (0 = unlimited)
	UseIndex        bool   // Try to optimize with index when possible
	IndexField      string // Field to use for index optimization
}

// RegexMatch represents a single regex match result
type RegexMatch struct {
	RecordNumber int         // 1-indexed record number
	FieldValue   string      // The field value that matched
	Matches      [][]int     // Byte indices of regexp matches
	FieldReader  FieldReader // Field reader for accessing the record
}

// RegexSearchResult contains all matches from a regex search
type RegexSearchResult struct {
	Pattern      string       // The regex pattern used
	Matches      []RegexMatch // All matching records
	TotalScanned int          // Total records scanned
	TotalMatched int          // Total records that matched
}

// RegexSearch performs a regex search on a string/character field
func (v *Vulpo) RegexSearch(fieldName, pattern string, options *RegexSearchOptions) (*RegexSearchResult, error) {
	if !v.Active() {
		return nil, NewError("database not open")
	}

	// Set default options if nil
	if options == nil {
		options = &RegexSearchOptions{
			CaseInsensitive: false,
			MaxResults:      0,
			UseIndex:        true,
			IndexField:      fieldName,
		}
	}

	// Find the field
	fieldDef := v.FieldByName(fieldName)
	if fieldDef == nil {
		return nil, NewErrorf("field '%s' not found", fieldName)
	}

	// Ensure it's a string/character field
	if fieldDef.Type() != FTCharacter {
		return nil, NewErrorf("field '%s' is not a character field (type: %s)", fieldName, fieldDef.Type().String())
	}

	// Compile the regex pattern
	regexFlags := ""
	if options.CaseInsensitive {
		regexFlags = "(?i)"
	}

	compiledPattern, err := regexp.Compile(regexFlags + pattern)
	if err != nil {
		return nil, NewErrorf("invalid regex pattern '%s': %v", pattern, err)
	}

	result := &RegexSearchResult{
		Pattern: pattern,
		Matches: make([]RegexMatch, 0),
	}

	// Try index optimization if requested and possible
	var optimized bool
	if options.UseIndex {
		optimized = v.tryIndexOptimization(fieldName, pattern, compiledPattern, options, result)
	}

	// Fall back to full table scan if not optimized
	if !optimized {
		err = v.performFullRegexScan(fieldName, compiledPattern, options, result)
		if err != nil {
			return nil, err
		}
	}

	result.TotalMatched = len(result.Matches)
	return result, nil
}

// tryIndexOptimization attempts to optimize regex search using indexes
func (v *Vulpo) tryIndexOptimization(fieldName, pattern string, compiled *regexp.Regexp, options *RegexSearchOptions, result *RegexSearchResult) bool {
	// For simple prefix patterns like "^ABC", we can use index seeks
	if isSimplePrefix(pattern) && !options.CaseInsensitive {
		prefix := extractPrefix(pattern)
		if len(prefix) > 0 {
			return v.performIndexPrefixSearch(fieldName, prefix, compiled, options, result)
		}
	}

	// Could add more optimization patterns here (exact matches, etc.)
	return false
}

// isSimplePrefix checks if pattern is a simple prefix match like "^ABC.*"
func isSimplePrefix(pattern string) bool {
	// Simple heuristic - starts with ^ and has literal characters following
	if len(pattern) < 2 || pattern[0] != '^' {
		return false
	}

	// Check if the next few characters are literal (not regex metacharacters)
	for i, r := range pattern[1:] {
		if i > 10 { // Don't check too far
			break
		}
		switch r {
		case '.', '*', '+', '?', '[', ']', '(', ')', '{', '}', '|', '\\', '$':
			return i > 0 // Return true if we found at least one literal char
		}
	}
	return true
}

// extractPrefix extracts the literal prefix from a pattern like "^ABC.*"
func extractPrefix(pattern string) string {
	if len(pattern) < 2 || pattern[0] != '^' {
		return ""
	}

	prefix := ""
	for _, r := range pattern[1:] {
		switch r {
		case '.', '*', '+', '?', '[', ']', '(', ')', '{', '}', '|', '\\', '$':
			return prefix
		default:
			prefix += string(r)
		}
	}
	return prefix
}

// performIndexPrefixSearch uses index seeking to optimize prefix searches
func (v *Vulpo) performIndexPrefixSearch(fieldName, prefix string, compiled *regexp.Regexp, options *RegexSearchOptions, result *RegexSearchResult) bool {
	// Find a tag for this field
	tag := v.findTagForField(fieldName)
	if tag == nil {
		return false // No suitable index found
	}

	// Save original position and tag selection
	originalPosition := v.Position()
	originalTag := v.SelectedTag()

	// Select the field's tag
	err := v.SelectTag(tag)
	if err != nil {
		return false
	}

	defer func() {
		// Restore original state
		_ = v.SelectTag(originalTag) // Ignore error in defer
		if originalPosition > 0 {
			_ = v.Goto(originalPosition) // Ignore error in defer
		}
	}()

	// Seek to the prefix
	seekResult, err := v.Seek(prefix)
	if err != nil {
		return false
	}

	if !seekResult.IsPositioned() {
		return true // No matches, but optimization worked
	}

	// Scan records starting from the seek position
	for !v.EOF() && (options.MaxResults == 0 || len(result.Matches) < options.MaxResults) {
		result.TotalScanned++

		// Get field reader for current record
		fieldReader, err := v.getFieldReader(fieldName)
		if err != nil {
			break
		}

		fieldValue, _ := fieldReader.AsString()

		// Check if we're still in the prefix range
		if len(fieldValue) < len(prefix) || fieldValue[:len(prefix)] != prefix {
			break // We've moved beyond the prefix range
		}

		// Apply regex to the full value
		if matches := compiled.FindAllStringIndex(fieldValue, -1); len(matches) > 0 {
			match := RegexMatch{
				RecordNumber: v.Position(),
				FieldValue:   fieldValue,
				Matches:      matches,
				FieldReader:  fieldReader,
			}
			result.Matches = append(result.Matches, match)
		}

		// Move to next record
		err = v.Next()
		if err != nil {
			break
		}
	}

	return true
}

// findTagForField attempts to find an index tag for the given field
func (v *Vulpo) findTagForField(fieldName string) *Tag {
	tags := v.ListTags()

	// First, look for tags that exactly match the field name
	for _, tag := range tags {
		if tag.Name() == fieldName || tag.Name() == fieldName+"_IDX" {
			return tag
		}
	}

	// Then look for tags that contain the field name
	for _, tag := range tags {
		tagName := tag.Name()
		if len(tagName) > len(fieldName) &&
			(tagName[:len(fieldName)] == fieldName || tagName[len(tagName)-len(fieldName):] == fieldName) {
			return tag
		}
	}

	return nil
}

// performFullRegexScan performs a full table scan with regex matching
func (v *Vulpo) performFullRegexScan(fieldName string, compiled *regexp.Regexp, options *RegexSearchOptions, result *RegexSearchResult) error {
	// Save original position
	originalPosition := v.Position()

	defer func() {
		// Restore original position
		if originalPosition > 0 {
			_ = v.Goto(originalPosition) // Ignore error in defer
		}
	}()

	// Go to first record
	err := v.First()
	if err != nil {
		return err
	}

	// Scan all records
	for !v.EOF() && (options.MaxResults == 0 || len(result.Matches) < options.MaxResults) {
		result.TotalScanned++

		// Get field reader for current record
		fieldReader, err := v.getFieldReader(fieldName)
		if err != nil {
			return err
		}

		fieldValue, _ := fieldReader.AsString()

		// Apply regex
		if matches := compiled.FindAllStringIndex(fieldValue, -1); len(matches) > 0 {
			match := RegexMatch{
				RecordNumber: v.Position(),
				FieldValue:   fieldValue,
				Matches:      matches,
				FieldReader:  fieldReader,
			}
			result.Matches = append(result.Matches, match)
		}

		// Move to next record
		err = v.Next()
		if err != nil {
			break
		}
	}

	return nil
}

// getFieldReader creates a field reader for the specified field at the current record
func (v *Vulpo) getFieldReader(fieldName string) (FieldReader, error) {
	fieldDef := v.FieldByName(fieldName)
	if fieldDef == nil {
		return nil, NewErrorf("field '%s' not found", fieldName)
	}

	// Use the existing FieldReader method
	fieldReader := v.FieldReader(fieldName)
	if fieldReader == nil {
		return nil, NewErrorf("failed to create field reader for '%s'", fieldName)
	}
	return fieldReader, nil
}

// RegexCount performs a regex search and returns only the count of matches
func (v *Vulpo) RegexCount(fieldName, pattern string, options *RegexSearchOptions) (int, error) {
	if options == nil {
		options = &RegexSearchOptions{}
	}

	// Override max results to unlimited for counting
	countOptions := *options
	countOptions.MaxResults = 0

	result, err := v.RegexSearch(fieldName, pattern, &countOptions)
	if err != nil {
		return 0, err
	}

	return result.TotalMatched, nil
}

// RegexExists checks if any record matches the regex pattern
func (v *Vulpo) RegexExists(fieldName, pattern string, options *RegexSearchOptions) (bool, error) {
	if options == nil {
		options = &RegexSearchOptions{}
	}

	// Override max results to 1 for existence check
	existsOptions := *options
	existsOptions.MaxResults = 1

	result, err := v.RegexSearch(fieldName, pattern, &existsOptions)
	if err != nil {
		return false, err
	}

	return result.TotalMatched > 0, nil
}

// Helper methods for RegexMatch

// GetSubmatches returns the actual submatches from the regex
func (rm *RegexMatch) GetSubmatches(compiled *regexp.Regexp) []string {
	return compiled.FindAllString(rm.FieldValue, -1)
}

// GetRecord positions the database to this match's record
func (rm *RegexMatch) GetRecord(v *Vulpo) error {
	return v.Goto(rm.RecordNumber)
}

// String provides a string representation of the match
func (rm *RegexMatch) String() string {
	return fmt.Sprintf("Record %d: '%s'", rm.RecordNumber, rm.FieldValue)
}
