package vulpo

import (
	"testing"
)

const testDBFWithIndexPath = "mkfdbflib/data/info.dbf"

func TestVulpo_TagOperations_NoDatabase(t *testing.T) {
	v := &Vulpo{}

	// Test tag operations when no database is open
	if tag := v.TagByName("any_tag"); tag != nil {
		t.Error("Expected nil tag when database not open")
	}

	if tag := v.DefaultTag(); tag != nil {
		t.Error("Expected nil default tag when database not open")
	}

	if tag := v.SelectedTag(); tag != nil {
		t.Error("Expected nil selected tag when database not open")
	}

	if tags := v.ListTags(); tags != nil {
		t.Error("Expected nil tag list when database not open")
	}

	if names := v.TagNames(); len(names) != 0 {
		t.Errorf("Expected empty tag names when database not open, got %d", len(names))
	}

	if count := v.TagCount(); count != 0 {
		t.Errorf("Expected 0 tag count when database not open, got %d", count)
	}

	if has := v.HasTag("any_tag"); has {
		t.Error("Expected false for HasTag when database not open")
	}
}

//nolint:gocyclo // TODO: split this test into smaller focused test functions
func TestVulpo_TagOperations_WithDatabase(t *testing.T) {
	v := &Vulpo{}
	err := v.Open(testDBFWithIndexPath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer func() {
		err := v.Close()
		if err != nil {
			// Close errors may occur due to CodeBase library state
			// after tag queries - log but don't fail test
			t.Logf("Warning: Close returned error: %v", err)
		}
	}()

	// Test listing tags
	tags := v.ListTags()
	if tags == nil {
		t.Fatal("Expected non-nil tag list")
	}

	if len(tags) == 0 {
		t.Skip("Test file has no indexes - this is expected for some test files")
	}

	// Test tag names
	names := v.TagNames()
	if len(names) != len(tags) {
		t.Errorf("Expected %d tag names, got %d", len(tags), len(names))
	}

	// Test tag count
	count := v.TagCount()
	if count != len(tags) {
		t.Errorf("Expected tag count %d, got %d", len(tags), count)
	}

	// Test individual tag properties
	for i, tag := range tags {
		if tag == nil {
			t.Errorf("Tag at index %d is nil", i)
			continue
		}

		if !tag.IsValid() {
			t.Errorf("Tag at index %d is not valid", i)
		}

		name := tag.Name()
		if name == "" {
			t.Errorf("Tag at index %d has empty name", i)
		}

		// Test HasTag
		if !v.HasTag(name) {
			t.Errorf("HasTag returned false for existing tag %s", name)
		}

		// Test TagByName
		tagByName := v.TagByName(name)
		if tagByName == nil {
			t.Errorf("TagByName returned nil for existing tag %s", name)
		} else if tagByName.Name() != name {
			t.Errorf("TagByName returned wrong tag: expected %s, got %s", name, tagByName.Name())
		}
	}

	// Test non-existent tag - be forgiving of CodeBase errors
	// The CodeBase library may log errors for non-existent tags, but that's expected
	if v.HasTag("NONEXISTENT_TAG") {
		t.Error("HasTag returned true for non-existent tag")
	}

	if tag := v.TagByName("NONEXISTENT_TAG"); tag != nil {
		t.Error("TagByName returned non-nil for non-existent tag")
	}
}

func TestVulpo_TagSelection(t *testing.T) {
	v := &Vulpo{}
	err := v.Open(testDBFWithIndexPath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer func() {
		err := v.Close()
		if err != nil {
			t.Logf("Warning: Close returned error: %v", err)
		}
	}()

	tags := v.ListTags()
	if len(tags) == 0 {
		t.Skip("Test file has no indexes - cannot test tag selection")
	}

	// Test selecting each tag
	for _, tag := range tags {
		err := v.SelectTag(tag)
		if err != nil {
			t.Errorf("Failed to select tag %s: %v", tag.Name(), err)
			continue
		}

		// Check if it's selected
		selected := v.SelectedTag()
		if selected == nil {
			t.Errorf("No tag selected after selecting %s", tag.Name())
		} else if selected.Name() != tag.Name() {
			t.Errorf("Selected tag mismatch: expected %s, got %s", tag.Name(), selected.Name())
		}
	}

	// Test selecting nil (record order)
	err = v.SelectTag(nil)
	if err != nil {
		t.Errorf("Failed to select record order: %v", err)
	}

	// Check that no tag is selected
	selected := v.SelectedTag()
	if selected != nil {
		t.Errorf("Expected no selected tag after selecting record order, got %s", selected.Name())
	}
}

func TestVulpo_Seek_NoDatabase(t *testing.T) {
	v := &Vulpo{}

	// Test seek operations when no database is open
	if result, err := v.Seek("test"); err == nil || result != SeekError {
		t.Error("Expected SeekError when database not open")
	}

	if result, err := v.SeekDouble(123.45); err == nil || result != SeekError {
		t.Error("Expected SeekError when database not open")
	}

	if result, err := v.SeekNext("test"); err == nil || result != SeekError {
		t.Error("Expected SeekError when database not open")
	}

	if result, err := v.SeekNextDouble(123.45); err == nil || result != SeekError {
		t.Error("Expected SeekError when database not open")
	}
}

func TestVulpo_SeekWithTags(t *testing.T) {
	v := &Vulpo{}
	err := v.Open(testDBFWithIndexPath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer func() {
		err := v.Close()
		if err != nil {
			t.Logf("Warning: Close returned error: %v", err)
		}
	}()

	tags := v.ListTags()
	if len(tags) == 0 {
		t.Skip("Test file has no indexes - cannot test seeking with tags")
	}

	// Test seeking with each tag
	for _, tag := range tags {
		// Test SeekWithTag with nil tag (should fail)
		if result, err := v.SeekWithTag(nil, "test"); err == nil || result != SeekError {
			t.Error("Expected SeekError when tag is nil")
		}

		// Create an invalid tag
		invalidTag := &Tag{name: "invalid", tagPtr: nil}
		if result, err := v.SeekWithTag(invalidTag, "test"); err == nil || result != SeekError {
			t.Error("Expected SeekError when tag is invalid")
		}

		// Test with valid tag (may not find record, but should not error)
		result, err := v.SeekWithTag(tag, "TESTVALUE")
		if err != nil {
			t.Errorf("Seek with tag %s failed: %v", tag.Name(), err)
		}

		// Result should be a valid SeekResult
		if result < SeekSuccess || result > SeekError {
			t.Errorf("Invalid seek result for tag %s: %v", tag.Name(), result)
		}

		// Test SeekDoubleWithTag (similar tests)
		result, err = v.SeekDoubleWithTag(tag, 12345.0)
		if err != nil {
			t.Errorf("SeekDouble with tag %s failed: %v", tag.Name(), err)
		}

		if result < SeekSuccess || result > SeekError {
			t.Errorf("Invalid seek double result for tag %s: %v", tag.Name(), result)
		}
	}
}

func TestVulpo_SeekBasic(t *testing.T) {
	v := &Vulpo{}
	err := v.Open(testDBFWithIndexPath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer func() {
		err := v.Close()
		if err != nil {
			t.Logf("Warning: Close returned error: %v", err)
		}
	}()

	// Test basic seek operations (may not find records, but should not error)
	result, err := v.Seek("TESTVALUE")
	if err != nil {
		t.Errorf("Basic Seek failed: %v", err)
	}
	if result < SeekSuccess || result > SeekError {
		t.Errorf("Invalid seek result: %v", result)
	}

	result, err = v.SeekDouble(12345.0)
	if err != nil {
		t.Errorf("Basic SeekDouble failed: %v", err)
	}
	if result < SeekSuccess || result > SeekError {
		t.Errorf("Invalid seek double result: %v", result)
	}

	result, err = v.SeekNext("TESTVALUE")
	if err != nil {
		t.Errorf("Basic SeekNext failed: %v", err)
	}
	if result < SeekSuccess || result > SeekError {
		t.Errorf("Invalid seek next result: %v", result)
	}

	result, err = v.SeekNextDouble(12345.0)
	if err != nil {
		t.Errorf("Basic SeekNextDouble failed: %v", err)
	}
	if result < SeekSuccess || result > SeekError {
		t.Errorf("Invalid seek next double result: %v", result)
	}
}

func TestSeekResult_String(t *testing.T) {
	tests := []struct {
		result   SeekResult
		expected string
	}{
		{SeekSuccess, "Success"},
		{SeekAfter, "After"},
		{SeekEOF, "EOF"},
		{SeekEntry, "Entry"},
		{SeekLocked, "Locked"},
		{SeekUnique, "Unique"},
		{SeekNoTag, "NoTag"},
		{SeekError, "Error"},
		{SeekResult(99), "Unknown(99)"},
	}

	for _, test := range tests {
		if got := test.result.String(); got != test.expected {
			t.Errorf("SeekResult(%d).String() = %s, expected %s", int(test.result), got, test.expected)
		}
	}
}

func TestSeekResult_Methods(t *testing.T) {
	// Test IsFound method
	if !SeekSuccess.IsFound() {
		t.Error("SeekSuccess should be found")
	}
	if SeekAfter.IsFound() {
		t.Error("SeekAfter should not be found")
	}
	if SeekError.IsFound() {
		t.Error("SeekError should not be found")
	}

	// Test IsPositioned method
	if !SeekSuccess.IsPositioned() {
		t.Error("SeekSuccess should be positioned")
	}
	if !SeekAfter.IsPositioned() {
		t.Error("SeekAfter should be positioned")
	}
	if SeekEOF.IsPositioned() {
		t.Error("SeekEOF should not be positioned")
	}
	if SeekError.IsPositioned() {
		t.Error("SeekError should not be positioned")
	}
}

func TestTag_Methods(t *testing.T) {
	v := &Vulpo{}
	err := v.Open(testDBFWithIndexPath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer func() {
		err := v.Close()
		if err != nil {
			t.Logf("Warning: Close returned error: %v", err)
		}
	}()

	tags := v.ListTags()
	if len(tags) == 0 {
		t.Skip("Test file has no indexes - cannot test Tag methods")
	}

	for _, tag := range tags {
		// Test Name method
		name := tag.Name()
		if name == "" {
			t.Error("Tag name should not be empty")
		}

		// Test IsValid method
		if !tag.IsValid() {
			t.Error("Valid tag should return true for IsValid")
		}
	}

	// Test invalid tag
	invalidTag := &Tag{name: "invalid", tagPtr: nil}
	if invalidTag.IsValid() {
		t.Error("Invalid tag should return false for IsValid")
	}
}

func TestVulpo_SeekWithTagRestoresOriginalSelection(t *testing.T) {
	v := &Vulpo{}
	err := v.Open(testDBFWithIndexPath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}
	defer func() {
		err := v.Close()
		if err != nil {
			t.Logf("Warning: Close returned error: %v", err)
		}
	}()

	tags := v.ListTags()
	if len(tags) < 2 {
		t.Skip("Need at least 2 indexes to test tag selection restoration")
	}

	// Select first tag
	err = v.SelectTag(tags[0])
	if err != nil {
		t.Fatalf("Failed to select first tag: %v", err)
	}

	originalTag := v.SelectedTag()
	if originalTag == nil || originalTag.Name() != tags[0].Name() {
		t.Fatal("Original tag not properly selected")
	}

	// Use SeekWithTag with a different tag
	_, err = v.SeekWithTag(tags[1], "TESTVALUE")
	if err != nil {
		t.Fatalf("SeekWithTag failed: %v", err)
	}

	// Check that original tag is restored
	restoredTag := v.SelectedTag()
	if restoredTag == nil {
		t.Error("No tag selected after SeekWithTag")
	} else if restoredTag.Name() != originalTag.Name() {
		t.Errorf("Tag selection not restored: expected %s, got %s",
			originalTag.Name(), restoredTag.Name())
	}
}
