package vulpo

import (
	"regexp"
	"testing"
)

const testDBFForRegex = "mkfdbflib/data/info.dbf"

func TestVulpo_RegexSearch_NoDatabase(t *testing.T) {
	v := &Vulpo{}

	result, err := v.RegexSearch("any_field", ".*", nil)
	if err == nil {
		t.Error("Expected error when database not open")
	}
	if result != nil {
		t.Error("Expected nil result when database not open")
	}
}

func TestVulpo_RegexSearch_InvalidField(t *testing.T) {
	v := &Vulpo{}
	err := v.Open(testDBFForRegex)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer func() {
		err := v.Close()
		if err != nil {
			t.Logf("Warning: Close returned error: %v", err)
		}
	}()

	result, err := v.RegexSearch("NONEXISTENT_FIELD", ".*", nil)
	if err == nil {
		t.Error("Expected error for non-existent field")
	}
	if result != nil {
		t.Error("Expected nil result for non-existent field")
	}
}

func TestVulpo_RegexSearch_NonCharacterField(t *testing.T) {
	v := &Vulpo{}
	err := v.Open(testDBFForRegex)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer func() {
		err := v.Close()
		if err != nil {
			t.Logf("Warning: Close returned error: %v", err)
		}
	}()

	// Find a numeric field to test with
	fieldDefs := v.FieldDefs()
	var numericFieldName string
	for i := 0; i < fieldDefs.Count(); i++ {
		field := fieldDefs.ByIndex(i)
		if field.Type() == FTNumeric || field.Type() == FTInteger {
			numericFieldName = field.Name()
			break
		}
	}

	if numericFieldName == "" {
		t.Skip("No numeric fields found in test file")
	}

	result, err := v.RegexSearch(numericFieldName, ".*", nil)
	if err == nil {
		t.Error("Expected error for non-character field")
	}
	if result != nil {
		t.Error("Expected nil result for non-character field")
	}
}

func TestVulpo_RegexSearch_InvalidPattern(t *testing.T) {
	v := &Vulpo{}
	err := v.Open(testDBFForRegex)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer func() {
		err := v.Close()
		if err != nil {
			t.Logf("Warning: Close returned error: %v", err)
		}
	}()

	// Find a character field
	charFieldName := findCharacterField(v)
	if charFieldName == "" {
		t.Skip("No character fields found in test file")
	}

	// Test with invalid regex pattern
	result, err := v.RegexSearch(charFieldName, "[", nil)
	if err == nil {
		t.Error("Expected error for invalid regex pattern")
	}
	if result != nil {
		t.Error("Expected nil result for invalid regex pattern")
	}
}

func TestVulpo_RegexSearch_BasicFunctionality(t *testing.T) {
	v := &Vulpo{}
	err := v.Open(testDBFForRegex)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer func() {
		err := v.Close()
		if err != nil {
			t.Logf("Warning: Close returned error: %v", err)
		}
	}()

	// Find a character field
	charFieldName := findCharacterField(v)
	if charFieldName == "" {
		t.Skip("No character fields found in test file")
	}

	// Test basic regex search that should match everything
	result, err := v.RegexSearch(charFieldName, ".*", nil)
	if err != nil {
		t.Fatalf("Basic regex search failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	if result.Pattern != ".*" {
		t.Errorf("Expected pattern '.*', got '%s'", result.Pattern)
	}

	if result.TotalMatched != len(result.Matches) {
		t.Errorf("TotalMatched (%d) doesn't match len(Matches) (%d)", result.TotalMatched, len(result.Matches))
	}

	if result.TotalScanned == 0 {
		t.Error("Expected TotalScanned > 0")
	}

	// Verify match structure
	for i, match := range result.Matches {
		if match.RecordNumber <= 0 {
			t.Errorf("Match %d has invalid record number: %d", i, match.RecordNumber)
		}

		if match.FieldValue == "" {
			t.Errorf("Match %d has empty field value", i)
		}

		if len(match.Matches) == 0 {
			t.Errorf("Match %d has no regex matches", i)
		}

		if match.FieldReader == nil {
			t.Errorf("Match %d has nil field reader", i)
		}
	}
}

func TestVulpo_RegexSearch_WithOptions(t *testing.T) {
	v := &Vulpo{}
	err := v.Open(testDBFForRegex)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer func() {
		err := v.Close()
		if err != nil {
			t.Logf("Warning: Close returned error: %v", err)
		}
	}()

	charFieldName := findCharacterField(v)
	if charFieldName == "" {
		t.Skip("No character fields found in test file")
	}

	// Test with max results limit
	options := &RegexSearchOptions{
		MaxResults: 2,
	}

	result, err := v.RegexSearch(charFieldName, ".*", options)
	if err != nil {
		t.Fatalf("Regex search with options failed: %v", err)
	}

	if len(result.Matches) > 2 {
		t.Errorf("Expected max 2 results, got %d", len(result.Matches))
	}

	// Test case insensitive
	options = &RegexSearchOptions{
		CaseInsensitive: true,
	}

	result, err = v.RegexSearch(charFieldName, "test", options)
	if err != nil {
		t.Fatalf("Case insensitive regex search failed: %v", err)
	}

	// Should work even if no matches found
	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Test with UseIndex disabled
	options = &RegexSearchOptions{
		UseIndex: false,
	}

	result, err = v.RegexSearch(charFieldName, ".*", options)
	if err != nil {
		t.Fatalf("Regex search with UseIndex=false failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}
}

func TestVulpo_RegexCount(t *testing.T) {
	v := &Vulpo{}
	err := v.Open(testDBFForRegex)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer func() {
		err := v.Close()
		if err != nil {
			t.Logf("Warning: Close returned error: %v", err)
		}
	}()

	charFieldName := findCharacterField(v)
	if charFieldName == "" {
		t.Skip("No character fields found in test file")
	}

	// Test count functionality
	count, err := v.RegexCount(charFieldName, ".*", nil)
	if err != nil {
		t.Fatalf("RegexCount failed: %v", err)
	}

	if count < 0 {
		t.Errorf("Expected non-negative count, got %d", count)
	}

	// Verify count matches full search
	result, err := v.RegexSearch(charFieldName, ".*", nil)
	if err != nil {
		t.Fatalf("RegexSearch for verification failed: %v", err)
	}

	if count != result.TotalMatched {
		t.Errorf("RegexCount (%d) doesn't match RegexSearch TotalMatched (%d)", count, result.TotalMatched)
	}
}

func TestVulpo_RegexExists(t *testing.T) {
	v := &Vulpo{}
	err := v.Open(testDBFForRegex)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer func() {
		err := v.Close()
		if err != nil {
			t.Logf("Warning: Close returned error: %v", err)
		}
	}()

	charFieldName := findCharacterField(v)
	if charFieldName == "" {
		t.Skip("No character fields found in test file")
	}

	// Test exists functionality with pattern that should match something
	exists, err := v.RegexExists(charFieldName, ".*", nil)
	if err != nil {
		t.Fatalf("RegexExists failed: %v", err)
	}

	// Should return true if there are any non-empty character fields
	// (but could be false if all fields are empty)
	_ = exists // Just verify no error

	// Test with pattern that definitely shouldn't match
	exists, err = v.RegexExists(charFieldName, "DEFINITELY_NOT_IN_ANY_RECORD_EVER_123456789", nil)
	if err != nil {
		t.Fatalf("RegexExists failed: %v", err)
	}

	if exists {
		t.Error("RegexExists should return false for non-matching pattern")
	}
}

func TestRegexMatch_Methods(t *testing.T) {
	v := &Vulpo{}
	err := v.Open(testDBFForRegex)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer func() {
		err := v.Close()
		if err != nil {
			t.Logf("Warning: Close returned error: %v", err)
		}
	}()

	charFieldName := findCharacterField(v)
	if charFieldName == "" {
		t.Skip("No character fields found in test file")
	}

	// Get some matches
	result, err := v.RegexSearch(charFieldName, ".*", &RegexSearchOptions{MaxResults: 1})
	if err != nil {
		t.Fatalf("RegexSearch failed: %v", err)
	}

	if len(result.Matches) == 0 {
		t.Skip("No matches found for testing match methods")
	}

	match := result.Matches[0]

	// Test String method
	str := match.String()
	if str == "" {
		t.Error("Expected non-empty string representation")
	}

	// Test GetSubmatches method
	compiled := regexp.MustCompile(".*")
	submatches := match.GetSubmatches(compiled)
	if len(submatches) == 0 && match.FieldValue != "" {
		t.Error("Expected submatches for non-empty field value")
	}

	// Test GetRecord method
	originalPos := v.Position()
	err = match.GetRecord(v)
	if err != nil {
		t.Errorf("GetRecord failed: %v", err)
	} else {
		newPos := v.Position()
		if newPos != match.RecordNumber {
			t.Errorf("GetRecord positioned to %d, expected %d", newPos, match.RecordNumber)
		}
	}

	// Restore original position
	if originalPos > 0 {
		_ = v.Goto(originalPos) // Ignore error in test
	}
}

func TestRegexOptimization_PrefixPatterns(t *testing.T) {
	// Test prefix pattern detection
	tests := []struct {
		pattern  string
		isPrefix bool
		prefix   string
	}{
		{"^ABC.*", true, "ABC"},
		{"^ABC", true, "ABC"},
		{"^A.*", true, "A"},
		{".*ABC", false, ""},
		{"ABC.*", false, ""},
		{"^", false, ""},
		{"^A[BC]", true, "A"},
		{"^ABC+", true, "ABC"},
	}

	for _, test := range tests {
		isPrefix := isSimplePrefix(test.pattern)
		if isPrefix != test.isPrefix {
			t.Errorf("isSimplePrefix('%s') = %v, expected %v", test.pattern, isPrefix, test.isPrefix)
		}

		if isPrefix {
			prefix := extractPrefix(test.pattern)
			if prefix != test.prefix {
				t.Errorf("extractPrefix('%s') = '%s', expected '%s'", test.pattern, prefix, test.prefix)
			}
		}
	}
}

func TestVulpo_RegexSearch_PrefixOptimization(t *testing.T) {
	v := &Vulpo{}
	err := v.Open(testDBFForRegex)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer func() {
		err := v.Close()
		if err != nil {
			t.Logf("Warning: Close returned error: %v", err)
		}
	}()

	charFieldName := findCharacterField(v)
	if charFieldName == "" {
		t.Skip("No character fields found in test file")
	}

	// Test prefix optimization (may or may not be used depending on available indexes)
	result, err := v.RegexSearch(charFieldName, "^A.*", &RegexSearchOptions{UseIndex: true})
	if err != nil {
		t.Fatalf("Prefix regex search failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected non-nil result")
	}

	// Verify all matches actually start with 'A'
	for i, match := range result.Matches {
		if len(match.FieldValue) == 0 || match.FieldValue[0] != 'A' {
			t.Errorf("Match %d ('%s') doesn't start with 'A'", i, match.FieldValue)
		}
	}

	// Compare with non-optimized version
	resultNonOpt, err := v.RegexSearch(charFieldName, "^A.*", &RegexSearchOptions{UseIndex: false})
	if err != nil {
		t.Fatalf("Non-optimized prefix regex search failed: %v", err)
	}

	// Results should be the same
	if len(result.Matches) != len(resultNonOpt.Matches) {
		t.Errorf("Optimized and non-optimized results differ: %d vs %d matches",
			len(result.Matches), len(resultNonOpt.Matches))
	}
}

// Helper function to find a character field in the test file
func findCharacterField(v *Vulpo) string {
	fieldDefs := v.FieldDefs()
	if fieldDefs == nil {
		return ""
	}

	for i := 0; i < fieldDefs.Count(); i++ {
		field := fieldDefs.ByIndex(i)
		if field.Type() == FTCharacter {
			return field.Name()
		}
	}

	return ""
}

func TestVulpo_RegexSearch_ErrorConditions(t *testing.T) {
	v := &Vulpo{}

	// Test with nil database
	_, err := v.RegexCount("field", "pattern", nil)
	if err == nil {
		t.Error("Expected error for RegexCount with closed database")
	}

	_, err = v.RegexExists("field", "pattern", nil)
	if err == nil {
		t.Error("Expected error for RegexExists with closed database")
	}
}
