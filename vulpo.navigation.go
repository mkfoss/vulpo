package vulpo

/*
#include "d4all.h"
*/
import "C"

// Goto positions the cursor to the specified physical record number.
//
// Parameters:
//   - recordidx: 1-indexed physical record number to navigate to
//
// Returns:
//   - error: nil on success, error if navigation fails or record number invalid
//
// This method moves directly to the specified physical record number, bypassing
// any active index ordering. Record numbers are 1-indexed and correspond to
// the physical position in the DBF file.
//
// Example:
//
//	err := v.Goto(5)  // Go to 5th physical record
//	if err != nil {
//		log.Printf("Failed to navigate: %v", err)
//	}
func (v *Vulpo) Goto(recordidx int) error {
	if !v.Active() {
		return NewError("database not open")
	}

	if recordidx <= 0 {
		return NewErrorf("invalid record index: %d (must be > 0)", recordidx)
	}

	result := C.d4go(v.data, C.long(recordidx))
	if result != 0 {
		return NewErrorf("failed to go to record %d: error code %d", recordidx, int(result))
	}

	return nil
}

// Next moves the cursor to the next record in the current navigation order.
//
// Returns:
//   - error: nil on success, error if navigation fails
//
// Navigation order depends on the currently selected index tag. If no tag is
// selected, moves to the next physical record. If an index is active, moves
// to the next record in index order.
//
// Example:
//
//	err := v.Next()
//	if err != nil && !v.EOF() {
//		log.Printf("Navigation failed: %v", err)
//	}
func (v *Vulpo) Next() error {
	if !v.Active() {
		return NewError("database not open")
	}

	result := C.d4skip(v.data, 1)
	if result != 0 {
		return NewErrorf("failed to move to next record: error code %d", int(result))
	}

	return nil
}

// Previous moves the cursor to the previous record in the current navigation order.
//
// Returns:
//   - error: nil on success, error if navigation fails
//
// Navigation order depends on the currently selected index tag. If no tag is
// selected, moves to the previous physical record. If an index is active, moves
// to the previous record in index order.
//
// Example:
//
//	err := v.Previous()
//	if err != nil && !v.BOF() {
//		log.Printf("Navigation failed: %v", err)
//	}
func (v *Vulpo) Previous() error {
	if !v.Active() {
		return NewError("database not open")
	}

	result := C.d4skip(v.data, -1)
	if result != 0 {
		return NewErrorf("failed to move to previous record: error code %d", int(result))
	}

	return nil
}

// Skip moves the cursor by the specified number of records in the current navigation order.
//
// Parameters:
//   - num: Number of records to skip (positive=forward, negative=backward, 0=no movement)
//
// Returns:
//   - error: nil on success, error if navigation fails
//
// Navigation order depends on the currently selected index tag. Positive values
// move forward, negative values move backward. This is equivalent to calling
// Next() or Previous() multiple times but more efficient.
//
// Example:
//
//	err := v.Skip(10)   // Move forward 10 records
//	err = v.Skip(-5)    // Move backward 5 records
func (v *Vulpo) Skip(num int) error {
	if !v.Active() {
		return NewError("database not open")
	}

	result := C.d4skip(v.data, C.long(num))
	if result != 0 {
		return NewErrorf("failed to skip %d records: error code %d", num, int(result))
	}

	return nil
}

// First moves the cursor to the first record in the database.
// Returns an error if the database is not active or navigation fails.
func (v *Vulpo) First() error {
	if !v.Active() {
		return NewError("database not open")
	}

	result := C.d4top(v.data)
	if result != 0 {
		return NewErrorf("failed to go to first record: error code %d", int(result))
	}

	return nil
}

// Last moves the cursor to the last record in the database.
// Returns an error if the database is not active or navigation fails.
func (v *Vulpo) Last() error {
	if !v.Active() {
		return NewError("database not open")
	}

	result := C.d4bottom(v.data)
	if result != 0 {
		return NewErrorf("failed to go to last record: error code %d", int(result))
	}

	return nil
}

// Position returns the current record number (1-indexed).
// Returns -1 if the database is not active or if at EOF/BOF.
func (v *Vulpo) Position() int {
	if !v.Active() {
		return -1
	}

	// Check if at EOF or BOF
	if C.d4eof(v.data) != 0 || C.d4bof(v.data) != 0 {
		return -1
	}

	recordNum := C.d4recNo(v.data)
	return int(recordNum)
}

// BOF returns true if the cursor is at the beginning of file.
// Returns false if the database is not active.
func (v *Vulpo) BOF() bool {
	if !v.Active() {
		return false
	}
	return C.d4bof(v.data) != 0
}

// IsBof returns true if the cursor is at the beginning of file.
// Returns false if the database is not active.
// This is an alias for BOF() to provide consistent naming with IsEof().
func (v *Vulpo) IsBof() bool {
	return v.BOF()
}

// EOF returns true if the cursor is at the end of file.
// Returns false if the database is not active.
func (v *Vulpo) EOF() bool {
	if !v.Active() {
		return false
	}
	return C.d4eof(v.data) != 0
}

// IsEof returns true if the cursor is at the end of file.
// Returns false if the database is not active.
// This is an alias for EOF() to provide consistent naming with IsBof().
func (v *Vulpo) IsEOF() bool {
	return v.EOF()
}
