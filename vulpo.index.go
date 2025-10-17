package vulpo

/*
#include "d4all.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// Tag represents an index tag in the DBF file
type Tag struct {
	name   string
	tagPtr *C.TAG4
}

// Name returns the name of the tag/index
func (t *Tag) Name() string {
	return t.name
}

// IsValid returns true if the tag pointer is valid
func (t *Tag) IsValid() bool {
	return t.tagPtr != nil
}

// TagByName finds and returns a tag by name.
// Returns nil if the tag is not found or database is not open.
func (v *Vulpo) TagByName(tagName string) *Tag {
	if !v.Active() {
		return nil
	}

	cTagName := C.CString(tagName)
	defer C.free(unsafe.Pointer(cTagName))

	tagPtr := C.d4tag(v.data, cTagName)
	if tagPtr == nil {
		return nil
	}

	return &Tag{
		name:   tagName,
		tagPtr: tagPtr,
	}
}

// DefaultTag returns the default tag for the data file.
// Returns nil if no default tag exists or database is not open.
func (v *Vulpo) DefaultTag() *Tag {
	if !v.Active() {
		return nil
	}

	tagPtr := C.d4tagDefault(v.data)
	if tagPtr == nil {
		return nil
	}

	// Get the tag name using t4alias function
	tagName := C.GoString(C.t4alias(tagPtr))

	return &Tag{
		name:   tagName,
		tagPtr: tagPtr,
	}
}

// SelectedTag returns the currently selected tag.
// Returns nil if no tag is selected or database is not open.
func (v *Vulpo) SelectedTag() *Tag {
	if !v.Active() {
		return nil
	}

	tagPtr := C.d4tagSelected(v.data)
	if tagPtr == nil {
		return nil
	}

	// Get the tag name using t4alias function
	tagName := C.GoString(C.t4alias(tagPtr))

	return &Tag{
		name:   tagName,
		tagPtr: tagPtr,
	}
}

// SelectTag selects a tag to be used for positioning operations.
// Pass nil to select record number ordering (no index).
// Returns an error if the database is not open.
func (v *Vulpo) SelectTag(tag *Tag) error {
	if !v.Active() {
		return NewError("database not open")
	}

	if tag == nil {
		C.d4tagSelect(v.data, nil)
		return nil
	}

	if !tag.IsValid() {
		return NewError("invalid tag")
	}

	C.d4tagSelect(v.data, tag.tagPtr)
	return nil
}

// SeekResult represents the result of a seek operation
type SeekResult int

const (
	SeekSuccess SeekResult = iota // Found exact match
	SeekAfter                     // Not found, positioned after where it would be
	SeekEOF                       // Not found, positioned at EOF
	SeekEntry                     // Record didn't exist (CODE4.errGo is false)
	SeekLocked                    // Lock failed
	SeekUnique                    // Duplicate key in unique index
	SeekNoTag                     // No tag available
	SeekError                     // Other error
)

// String returns a string representation of the SeekResult
func (sr SeekResult) String() string {
	switch sr {
	case SeekSuccess:
		return "Success"
	case SeekAfter:
		return "After"
	case SeekEOF:
		return "EOF"
	case SeekEntry:
		return "Entry"
	case SeekLocked:
		return "Locked"
	case SeekUnique:
		return "Unique"
	case SeekNoTag:
		return "NoTag"
	case SeekError:
		return "Error"
	default:
		return fmt.Sprintf("Unknown(%d)", int(sr))
	}
}

// convertSeekResult converts CodeBase seek result to SeekResult
func convertSeekResult(result C.int) SeekResult {
	switch result {
	case 0: // r4success
		return SeekSuccess
	case 1: // r4after
		return SeekAfter
	case 2: // r4eof
		return SeekEOF
	case 3: // r4entry
		return SeekEntry
	case 4: // r4locked
		return SeekLocked
	case 5: // r4unique
		return SeekUnique
	case -1: // r4noTag (typically)
		return SeekNoTag
	default:
		if result < 0 {
			return SeekError
		}
		return SeekResult(result)
	}
}

// Seek searches for a record using the selected tag.
// The searchValue should be formatted appropriately for the tag type:
// - Date: "CCYYMMDD" (e.g., "20231201")
// - DateTime: "CCYYMMDDhh:mm:ss:ttt"
// - Numeric/Float/Double/Integer/Currency: "123.45"
// - Character: any string (partial matches allowed)
// Returns SeekResult indicating the outcome of the search.
func (v *Vulpo) Seek(searchValue string) (SeekResult, error) {
	if !v.Active() {
		return SeekError, NewError("database not open")
	}

	cSearchValue := C.CString(searchValue)
	defer C.free(unsafe.Pointer(cSearchValue))

	result := C.d4seek(v.data, cSearchValue)
	return convertSeekResult(result), nil
}

// SeekDouble searches for a record using a double value.
// This is more efficient than Seek for numeric searches.
func (v *Vulpo) SeekDouble(searchValue float64) (SeekResult, error) {
	if !v.Active() {
		return SeekError, NewError("database not open")
	}

	result := C.d4seekDouble(v.data, C.double(searchValue))
	return convertSeekResult(result), nil
}

// SeekNext searches for the next record matching the search value.
// This continues a search started with Seek().
func (v *Vulpo) SeekNext(searchValue string) (SeekResult, error) {
	if !v.Active() {
		return SeekError, NewError("database not open")
	}

	cSearchValue := C.CString(searchValue)
	defer C.free(unsafe.Pointer(cSearchValue))

	result := C.d4seekNext(v.data, cSearchValue)
	return convertSeekResult(result), nil
}

// SeekNextDouble searches for the next record matching a double value.
// This continues a search started with SeekDouble().
func (v *Vulpo) SeekNextDouble(searchValue float64) (SeekResult, error) {
	if !v.Active() {
		return SeekError, NewError("database not open")
	}

	result := C.d4seekNextDouble(v.data, C.double(searchValue))
	return convertSeekResult(result), nil
}

// SeekWithTag is a convenience method that selects a tag and performs a seek in one operation.
// The original selected tag is restored after the search.
func (v *Vulpo) SeekWithTag(tag *Tag, searchValue string) (SeekResult, error) {
	if !v.Active() {
		return SeekError, NewError("database not open")
	}

	if tag == nil {
		return SeekError, NewError("tag cannot be nil")
	}

	if !tag.IsValid() {
		return SeekError, NewError("invalid tag")
	}

	// Save current tag selection
	originalTag := v.SelectedTag()

	// Select the desired tag
	err := v.SelectTag(tag)
	if err != nil {
		return SeekError, err
	}

	// Perform the search
	result, err := v.Seek(searchValue)

	// Restore original tag selection
	_ = v.SelectTag(originalTag) // Ignore restore errors

	return result, err
}

// SeekDoubleWithTag is a convenience method that selects a tag and performs a double seek in one operation.
// The original selected tag is restored after the search.
func (v *Vulpo) SeekDoubleWithTag(tag *Tag, searchValue float64) (SeekResult, error) {
	if !v.Active() {
		return SeekError, NewError("database not open")
	}

	if tag == nil {
		return SeekError, NewError("tag cannot be nil")
	}

	if !tag.IsValid() {
		return SeekError, NewError("invalid tag")
	}

	// Save current tag selection
	originalTag := v.SelectedTag()

	// Select the desired tag
	err := v.SelectTag(tag)
	if err != nil {
		return SeekError, err
	}

	// Perform the search
	result, err := v.SeekDouble(searchValue)

	// Restore original tag selection
	_ = v.SelectTag(originalTag) // Ignore restore errors

	return result, err
}

// IsSeekFound returns true if the seek result indicates a successful find
func (sr SeekResult) IsFound() bool {
	return sr == SeekSuccess
}

// IsSeekPositioned returns true if the seek positioned the cursor somewhere
func (sr SeekResult) IsPositioned() bool {
	return sr == SeekSuccess || sr == SeekAfter
}

// ListTags returns a list of all available tags for the data file.
// Returns an empty slice if no tags are available or database is not open.
func (v *Vulpo) ListTags() []*Tag {
	if !v.Active() {
		return nil
	}

	var tags []*Tag

	// Start with first tag (passing NULL to d4tagNext gets the first tag)
	tagPtr := C.d4tagNext(v.data, nil)
	for tagPtr != nil {
		// Get the tag name using t4alias function
		tagName := C.GoString(C.t4alias(tagPtr))

		tag := &Tag{
			name:   tagName,
			tagPtr: tagPtr,
		}
		tags = append(tags, tag)

		// Get next tag
		tagPtr = C.d4tagNext(v.data, tagPtr)
	}

	return tags
}

// TagNames returns a list of tag names for the data file.
// This is a convenience function that returns just the names.
func (v *Vulpo) TagNames() []string {
	tags := v.ListTags()
	names := make([]string, len(tags))

	for i, tag := range tags {
		names[i] = tag.Name()
	}

	return names
}

// HasTag checks if a tag with the given name exists.
func (v *Vulpo) HasTag(tagName string) bool {
	tag := v.TagByName(tagName)
	return tag != nil
}

// TagCount returns the number of tags available for the data file.
func (v *Vulpo) TagCount() int {
	tags := v.ListTags()
	return len(tags)
}
