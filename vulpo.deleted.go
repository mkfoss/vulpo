package vulpo

/*
#cgo CFLAGS: -I./mkfdbflib
#cgo LDFLAGS: -L./mkfdbflib -lmkfdbf
#include "d4all.h"
#include <stdlib.h>
*/
import "C"

// Deleted returns true if the current record is marked for deletion.
// Returns false if the database is not active or if at EOF/BOF.
func (v *Vulpo) Deleted() bool {
	if !v.Active() {
		return false
	}

	// Check if at EOF or BOF
	if v.EOF() || v.BOF() {
		return false
	}

	return C.d4deleted(v.data) != 0
}

// IsDeleted is an alias for Deleted() for consistency with other Is* methods.
func (v *Vulpo) IsDeleted() bool {
	return v.Deleted()
}

// Delete marks the current record for deletion.
// The record is not physically removed until Pack() is called.
// Returns an error if the database is not active or if at EOF/BOF.
func (v *Vulpo) Delete() error {
	if !v.Active() {
		return NewError("database not open")
	}

	// Check if at EOF or BOF
	if v.EOF() || v.BOF() {
		return NewError("no current record to delete")
	}

	C.d4delete(v.data)
	return nil
}

// Recall removes the deletion mark from the current record.
// This "undeletes" a record that was previously marked for deletion.
// Returns an error if the database is not active or if at EOF/BOF.
func (v *Vulpo) Recall() error {
	if !v.Active() {
		return NewError("database not open")
	}

	// Check if at EOF or BOF
	if v.EOF() || v.BOF() {
		return NewError("no current record to recall")
	}

	C.d4recall(v.data)
	return nil
}

// Pack physically removes all records marked for deletion from the database.
// This operation:
// - Permanently removes deleted records from the file
// - Automatically reindexes all open index files
// - Makes the record buffer and position undefined (call a positioning function after)
// - Should be done exclusively (no other users) for best performance
//
// WARNING: This is a destructive operation. Take appropriate backups first.
// Returns an error if the operation fails.
func (v *Vulpo) Pack() error {
	if !v.Active() {
		return NewError("database not open")
	}

	result := C.d4pack(v.data)
	if result != 0 {
		return NewErrorf("failed to pack database: error code %d", int(result))
	}

	return nil
}

// CountDeleted counts the total number of records marked for deletion.
// This scans the entire database and preserves the current position.
func (v *Vulpo) CountDeleted() (int, error) {
	if !v.Active() {
		return 0, NewError("database not open")
	}

	// Save original position and tag selection
	originalPosition := v.Position()
	originalTag := v.SelectedTag()

	defer func() {
		// Restore original state
		_ = v.SelectTag(originalTag)
		if originalPosition > 0 {
			_ = v.Goto(originalPosition)
		}
	}()

	// Use record ordering (no index) for counting
	err := v.SelectTag(nil)
	if err != nil {
		return 0, err
	}

	count := 0

	// Go to first record
	err = v.First()
	if err != nil {
		return 0, NewErrorf("failed to go to first record: %v", err)
	}

	// Scan all records
	for !v.EOF() {
		if v.Deleted() {
			count++
		}

		// Move to next record
		err = v.Next()
		if err != nil {
			break // End of file or error
		}
	}

	return count, nil
}

// CountActive counts the number of non-deleted (active) records.
// This scans the entire database and preserves the current position.
func (v *Vulpo) CountActive() (int, error) {
	if !v.Active() {
		return 0, NewError("database not open")
	}

	// Save original position and tag selection
	originalPosition := v.Position()
	originalTag := v.SelectedTag()

	defer func() {
		// Restore original state
		_ = v.SelectTag(originalTag)
		if originalPosition > 0 {
			_ = v.Goto(originalPosition)
		}
	}()

	// Use record ordering (no index) for counting
	err := v.SelectTag(nil)
	if err != nil {
		return 0, err
	}

	count := 0

	// Go to first record
	err = v.First()
	if err != nil {
		return 0, NewErrorf("failed to go to first record: %v", err)
	}

	// Scan all records
	for !v.EOF() {
		if !v.Deleted() {
			count++
		}

		// Move to next record
		err = v.Next()
		if err != nil {
			break // End of file or error
		}
	}

	return count, nil
}

// DeletedRecordInfo contains information about deleted records
type DeletedRecordInfo struct {
	RecordNumber int  // 1-indexed physical record number
	IsDeleted    bool // Always true for records in this structure
}

// ListDeletedRecords returns information about all deleted records.
// This preserves the current position.
func (v *Vulpo) ListDeletedRecords() ([]DeletedRecordInfo, error) {
	if !v.Active() {
		return nil, NewError("database not open")
	}

	// Save original position and tag selection
	originalPosition := v.Position()
	originalTag := v.SelectedTag()

	defer func() {
		// Restore original state
		_ = v.SelectTag(originalTag)
		if originalPosition > 0 {
			_ = v.Goto(originalPosition)
		}
	}()

	// Use record ordering (no index) for scanning
	err := v.SelectTag(nil)
	if err != nil {
		return nil, err
	}

	var deletedRecords []DeletedRecordInfo

	// Go to first record
	err = v.First()
	if err != nil {
		return nil, NewErrorf("failed to go to first record: %v", err)
	}

	// Scan all records
	for !v.EOF() {
		if v.Deleted() {
			deletedRecords = append(deletedRecords, DeletedRecordInfo{
				RecordNumber: v.Position(),
				IsDeleted:    true,
			})
		}

		// Move to next record
		err = v.Next()
		if err != nil {
			break // End of file or error
		}
	}

	return deletedRecords, nil
}

// ForEachDeletedRecord iterates through all deleted records with a callback.
// This preserves the current position.
func (v *Vulpo) ForEachDeletedRecord(callback func(recordNumber int) error) error {
	if !v.Active() {
		return NewError("database not open")
	}

	// Save original position and tag selection
	originalPosition := v.Position()
	originalTag := v.SelectedTag()

	defer func() {
		// Restore original state
		_ = v.SelectTag(originalTag)
		if originalPosition > 0 {
			_ = v.Goto(originalPosition)
		}
	}()

	// Use record ordering (no index) for scanning
	err := v.SelectTag(nil)
	if err != nil {
		return err
	}

	// Go to first record
	err = v.First()
	if err != nil {
		return NewErrorf("failed to go to first record: %v", err)
	}

	// Scan all records
	for !v.EOF() {
		if v.Deleted() {
			// Call the callback with the record number
			if err := callback(v.Position()); err != nil {
				return err
			}
		}

		// Move to next record
		err = v.Next()
		if err != nil {
			break // End of file or error
		}
	}

	return nil
}

// RecallAllDeleted removes the deletion mark from all deleted records.
// This "undeletes" all records that were previously marked for deletion.
// This preserves the current position.
func (v *Vulpo) RecallAllDeleted() (int, error) {
	if !v.Active() {
		return 0, NewError("database not open")
	}

	// Save original position and tag selection
	originalPosition := v.Position()
	originalTag := v.SelectedTag()

	defer func() {
		// Restore original state
		_ = v.SelectTag(originalTag)
		if originalPosition > 0 {
			_ = v.Goto(originalPosition)
		}
	}()

	// Use record ordering (no index) for processing
	err := v.SelectTag(nil)
	if err != nil {
		return 0, err
	}

	count := 0

	// Go to first record
	err = v.First()
	if err != nil {
		return 0, NewErrorf("failed to go to first record: %v", err)
	}

	// Scan all records and recall deleted ones
	for !v.EOF() {
		if v.Deleted() {
			C.d4recall(v.data) // Recall this record
			count++
		}

		// Move to next record
		err = v.Next()
		if err != nil {
			break // End of file or error
		}
	}

	return count, nil
}
