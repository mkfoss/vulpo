package vulpo

import (
	"testing"
)

func TestBasicDeletedRecordFunctionality(t *testing.T) {
	// Test the basic deleted record functions with an inactive database
	v := &Vulpo{}

	t.Run("TestDeletedWithInactiveDB", func(t *testing.T) {
		// Test Deleted() with inactive database
		if v.Deleted() {
			t.Error("Expected Deleted() to return false for inactive database")
		}

		if v.IsDeleted() {
			t.Error("Expected IsDeleted() to return false for inactive database")
		}
	})

	t.Run("TestDeleteWithInactiveDB", func(t *testing.T) {
		// Test Delete() with inactive database
		err := v.Delete()
		if err == nil {
			t.Error("Expected error for Delete() with inactive database")
		}
	})

	t.Run("TestRecallWithInactiveDB", func(t *testing.T) {
		// Test Recall() with inactive database
		err := v.Recall()
		if err == nil {
			t.Error("Expected error for Recall() with inactive database")
		}
	})

	t.Run("TestPackWithInactiveDB", func(t *testing.T) {
		// Test Pack() with inactive database
		err := v.Pack()
		if err == nil {
			t.Error("Expected error for Pack() with inactive database")
		}
	})

	t.Run("TestCountDeletedWithInactiveDB", func(t *testing.T) {
		// Test CountDeleted() with inactive database
		count, err := v.CountDeleted()
		if err == nil {
			t.Error("Expected error for CountDeleted() with inactive database")
		}
		if count != 0 {
			t.Errorf("Expected count of 0 for inactive database, got %d", count)
		}
	})

	t.Run("TestCountActiveWithInactiveDB", func(t *testing.T) {
		// Test CountActive() with inactive database
		count, err := v.CountActive()
		if err == nil {
			t.Error("Expected error for CountActive() with inactive database")
		}
		if count != 0 {
			t.Errorf("Expected count of 0 for inactive database, got %d", count)
		}
	})

	t.Run("TestListDeletedRecordsWithInactiveDB", func(t *testing.T) {
		// Test ListDeletedRecords() with inactive database
		records, err := v.ListDeletedRecords()
		if err == nil {
			t.Error("Expected error for ListDeletedRecords() with inactive database")
		}
		if records != nil {
			t.Error("Expected nil records for inactive database")
		}
	})

	t.Run("TestForEachDeletedRecordWithInactiveDB", func(t *testing.T) {
		// Test ForEachDeletedRecord() with inactive database
		err := v.ForEachDeletedRecord(func(recordNumber int) error {
			return nil
		})
		if err == nil {
			t.Error("Expected error for ForEachDeletedRecord() with inactive database")
		}
	})

	t.Run("TestRecallAllDeletedWithInactiveDB", func(t *testing.T) {
		// Test RecallAllDeleted() with inactive database
		count, err := v.RecallAllDeleted()
		if err == nil {
			t.Error("Expected error for RecallAllDeleted() with inactive database")
		}
		if count != 0 {
			t.Errorf("Expected count of 0 for inactive database, got %d", count)
		}
	})
}

func TestDeletedRecordInfoStruct(t *testing.T) {
	// Test the DeletedRecordInfo struct
	info := DeletedRecordInfo{
		RecordNumber: 42,
		IsDeleted:    true,
	}

	if info.RecordNumber != 42 {
		t.Errorf("Expected RecordNumber 42, got %d", info.RecordNumber)
	}

	if !info.IsDeleted {
		t.Error("Expected IsDeleted to be true")
	}
}

// This test demonstrates how the deleted record functionality would work with a real database
//
//nolint:gocyclo // TODO: break this test into smaller sub-tests to reduce complexity
func TestDeletedRecordWorkflow(t *testing.T) {
	// This test is a conceptual demonstration of the workflow
	// In practice, it would need a real DBF file

	t.Skip("Skipping integration test - requires test data file")

	v := &Vulpo{}
	err := v.Open("testdata/sample.dbf") // This would need a real test file
	if err != nil {
		t.Skip("No test data file available")
	}
	defer v.Close()

	t.Run("TestDeleteRecordWorkflow", func(t *testing.T) {
		// Go to first record
		err := v.First()
		if err != nil {
			t.Fatalf("Failed to go to first record: %v", err)
		}

		// Check if it's deleted initially
		initiallyDeleted := v.Deleted()

		// Mark for deletion
		err = v.Delete()
		if err != nil {
			t.Fatalf("Failed to delete record: %v", err)
		}

		// Verify it's now marked as deleted
		if !v.Deleted() {
			t.Error("Record should be marked as deleted")
		}

		// Count deleted records
		deletedCount, err := v.CountDeleted()
		if err != nil {
			t.Fatalf("Failed to count deleted records: %v", err)
		}

		expectedDeletedCount := 1
		if initiallyDeleted {
			expectedDeletedCount = 1 // It was already deleted
		}

		if deletedCount < expectedDeletedCount {
			t.Errorf("Expected at least %d deleted record, got %d", expectedDeletedCount, deletedCount)
		}

		// Recall the record (undelete)
		err = v.Recall()
		if err != nil {
			t.Fatalf("Failed to recall record: %v", err)
		}

		// Verify it's no longer deleted
		if v.Deleted() {
			t.Error("Record should not be marked as deleted after recall")
		}

		// Count active records
		activeCount, err := v.CountActive()
		if err != nil {
			t.Fatalf("Failed to count active records: %v", err)
		}

		if activeCount < 1 {
			t.Error("Expected at least 1 active record")
		}
	})

	t.Run("TestListDeletedRecords", func(t *testing.T) {
		// Mark some records for deletion
		_ = v.First()  // Ignore error in test
		_ = v.Delete() // Ignore error in test

		// List deleted records
		deletedRecords, err := v.ListDeletedRecords()
		if err != nil {
			t.Fatalf("Failed to list deleted records: %v", err)
		}

		if len(deletedRecords) == 0 {
			t.Error("Expected at least one deleted record")
		}

		for _, record := range deletedRecords {
			if !record.IsDeleted {
				t.Error("Expected all records in list to be deleted")
			}
			if record.RecordNumber <= 0 {
				t.Error("Expected positive record number")
			}
		}
	})

	t.Run("TestForEachDeletedRecord", func(t *testing.T) {
		var deletedRecordNumbers []int

		err := v.ForEachDeletedRecord(func(recordNumber int) error {
			deletedRecordNumbers = append(deletedRecordNumbers, recordNumber)
			return nil
		})

		if err != nil {
			t.Fatalf("ForEachDeletedRecord failed: %v", err)
		}

		// Verify we found some deleted records
		if len(deletedRecordNumbers) == 0 {
			t.Log("No deleted records found (this might be expected)")
		}

		// Verify all record numbers are positive
		for _, recordNum := range deletedRecordNumbers {
			if recordNum <= 0 {
				t.Errorf("Expected positive record number, got %d", recordNum)
			}
		}
	})

	t.Run("TestRecallAllDeleted", func(t *testing.T) {
		// Mark a record for deletion first
		_ = v.First()  // Ignore error in test
		_ = v.Delete() // Ignore error in test

		// Recall all deleted records
		recalledCount, err := v.RecallAllDeleted()
		if err != nil {
			t.Fatalf("Failed to recall all deleted records: %v", err)
		}

		if recalledCount < 1 {
			t.Error("Expected at least one record to be recalled")
		}

		// Verify no records are deleted now
		deletedCount, err := v.CountDeleted()
		if err != nil {
			t.Fatalf("Failed to count deleted records after recall: %v", err)
		}

		if deletedCount != 0 {
			t.Errorf("Expected 0 deleted records after recall all, got %d", deletedCount)
		}
	})
}

func TestDeletedRecordEdgeCases(t *testing.T) {
	// Test edge cases and error conditions
	t.Skip("Skipping integration test - requires test data file")

	v := &Vulpo{}
	err := v.Open("testdata/sample.dbf")
	if err != nil {
		t.Skip("No test data file available")
	}
	defer v.Close()

	t.Run("TestDeleteAtEOF", func(t *testing.T) {
		// Position at EOF
		_ = v.Last() // Ignore error in test
		_ = v.Next() // This should put us at EOF

		if !v.EOF() {
			t.Skip("Not at EOF, skipping test")
		}

		// Try to delete at EOF - should fail
		err := v.Delete()
		if err == nil {
			t.Error("Expected error when trying to delete at EOF")
		}

		// Try to recall at EOF - should fail
		err = v.Recall()
		if err == nil {
			t.Error("Expected error when trying to recall at EOF")
		}

		// Deleted() should return false at EOF
		if v.Deleted() {
			t.Error("Expected Deleted() to return false at EOF")
		}
	})

	t.Run("TestDeleteAtBOF", func(t *testing.T) {
		// Position at BOF
		_ = v.First()    // Ignore error in test
		_ = v.Previous() // This should put us at BOF

		if !v.BOF() {
			t.Skip("Not at BOF, skipping test")
		}

		// Try to delete at BOF - should fail
		err := v.Delete()
		if err == nil {
			t.Error("Expected error when trying to delete at BOF")
		}

		// Try to recall at BOF - should fail
		err = v.Recall()
		if err == nil {
			t.Error("Expected error when trying to recall at BOF")
		}

		// Deleted() should return false at BOF
		if v.Deleted() {
			t.Error("Expected Deleted() to return false at BOF")
		}
	})
}
