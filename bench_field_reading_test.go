package vulpo

import (
	"testing"
)

// Large DBF files for benchmarking
const (
	detailDBF   = "/data/seandata/atcdbf/detail.DBF"
	billlistDBF = "/data/seandata/atcdbf/BILLLIST.dbf"
)

func BenchmarkFieldReading_Detail_Sequential(b *testing.B) {
	v := &Vulpo{}
	err := v.Open(detailDBF)
	if err != nil {
		b.Fatalf("Failed to open %s: %v", detailDBF, err)
	}
	defer v.Close()

	// Get field information
	fieldDefs := v.FieldDefs()
	if fieldDefs == nil || fieldDefs.Count() == 0 {
		b.Fatal("No fields found")
	}

	// Create field readers for all fields
	var fieldReaders []FieldReader
	for i := 0; i < fieldDefs.Count(); i++ {
		fieldDef := fieldDefs.ByIndex(i)
		if fieldDef != nil {
			reader := v.FieldReader(fieldDef.Name())
			if reader != nil {
				fieldReaders = append(fieldReaders, reader)
			}
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Navigate to first record
		err := v.First()
		if err != nil {
			b.Fatalf("Failed to go to first record: %v", err)
		}

		recordCount := 0
		for !v.EOF() && recordCount < 1000 { // Limit to 1000 records per iteration
			// Read all fields in the current record
			for _, reader := range fieldReaders {
				_, _ = reader.AsString() // Read field value
			}

			recordCount++
			if recordCount >= 1000 {
				break
			}

			err = v.Next()
			if err != nil {
				break
			}
		}
	}
}

func BenchmarkFieldReading_Billlist_Sequential(b *testing.B) {
	v := &Vulpo{}
	err := v.Open(billlistDBF)
	if err != nil {
		b.Fatalf("Failed to open %s: %v", billlistDBF, err)
	}
	defer v.Close()

	// Get field information
	fieldDefs := v.FieldDefs()
	if fieldDefs == nil || fieldDefs.Count() == 0 {
		b.Fatal("No fields found")
	}

	// Create field readers for all fields
	var fieldReaders []FieldReader
	for i := 0; i < fieldDefs.Count(); i++ {
		fieldDef := fieldDefs.ByIndex(i)
		if fieldDef != nil {
			reader := v.FieldReader(fieldDef.Name())
			if reader != nil {
				fieldReaders = append(fieldReaders, reader)
			}
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Navigate to first record
		err := v.First()
		if err != nil {
			b.Fatalf("Failed to go to first record: %v", err)
		}

		recordCount := 0
		for !v.EOF() && recordCount < 1000 { // Limit to 1000 records per iteration
			// Read all fields in the current record
			for _, reader := range fieldReaders {
				_, _ = reader.AsString() // Read field value
			}

			recordCount++
			if recordCount >= 1000 {
				break
			}

			err = v.Next()
			if err != nil {
				break
			}
		}
	}
}

func BenchmarkFieldReading_Detail_RandomAccess(b *testing.B) {
	v := &Vulpo{}
	err := v.Open(detailDBF)
	if err != nil {
		b.Fatalf("Failed to open %s: %v", detailDBF, err)
	}
	defer v.Close()

	// Get total record count
	err = v.First()
	if err != nil {
		b.Fatalf("Failed to go to first record: %v", err)
	}

	totalRecords := 0
	for !v.EOF() && totalRecords < 10000 { // Count up to 10k records
		totalRecords++
		err = v.Next()
		if err != nil {
			break
		}
	}

	if totalRecords == 0 {
		b.Skip("No records found")
	}

	// Get first field for testing
	fieldDefs := v.FieldDefs()
	if fieldDefs == nil || fieldDefs.Count() == 0 {
		b.Fatal("No fields found")
	}

	firstField := v.FieldReader(fieldDefs.ByIndex(0).Name())
	if firstField == nil {
		b.Fatal("Failed to get first field reader")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Random access to records
		recordNum := (i % totalRecords) + 1
		err := v.Goto(recordNum)
		if err != nil {
			b.Errorf("Failed to goto record %d: %v", recordNum, err)
			continue
		}

		// Read field value
		_, _ = firstField.AsString()
	}
}

//nolint:gocyclo // TODO: refactor this benchmark to reduce complexity
func BenchmarkFieldReading_TypedAccess(b *testing.B) {
	v := &Vulpo{}
	err := v.Open(detailDBF)
	if err != nil {
		b.Fatalf("Failed to open %s: %v", detailDBF, err)
	}
	defer v.Close()

	// Get field readers by type
	fieldDefs := v.FieldDefs()
	if fieldDefs == nil || fieldDefs.Count() == 0 {
		b.Fatal("No fields found")
	}

	var stringFields, numericFields, dateFields []FieldReader

	for i := 0; i < fieldDefs.Count(); i++ {
		fieldDef := fieldDefs.ByIndex(i)
		if fieldDef == nil {
			continue
		}

		reader := v.FieldReader(fieldDef.Name())
		if reader == nil {
			continue
		}

		switch fieldDef.Type() {
		case FTCharacter:
			stringFields = append(stringFields, reader)
		case FTNumeric, FTInteger, FTFloat, FTDouble, FTCurrency:
			numericFields = append(numericFields, reader)
		case FTDate, FTDateTime:
			dateFields = append(dateFields, reader)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Navigate to first record
		err := v.First()
		if err != nil {
			b.Fatalf("Failed to go to first record: %v", err)
		}

		recordCount := 0
		for !v.EOF() && recordCount < 500 { // Limit to 500 records per iteration
			// Read string fields as strings
			for _, reader := range stringFields {
				_, _ = reader.AsString()
			}

			// Read numeric fields as floats
			for _, reader := range numericFields {
				_, _ = reader.AsFloat()
			}

			// Read date fields as time
			for _, reader := range dateFields {
				_, _ = reader.AsTime()
			}

			recordCount++
			if recordCount >= 500 {
				break
			}

			err = v.Next()
			if err != nil {
				break
			}
		}
	}
}

func BenchmarkFieldReading_SingleField(b *testing.B) {
	v := &Vulpo{}
	err := v.Open(detailDBF)
	if err != nil {
		b.Fatalf("Failed to open %s: %v", detailDBF, err)
	}
	defer v.Close()

	// Get first field
	fieldDefs := v.FieldDefs()
	if fieldDefs == nil || fieldDefs.Count() == 0 {
		b.Fatal("No fields found")
	}

	firstField := v.FieldReader(fieldDefs.ByIndex(0).Name())
	if firstField == nil {
		b.Fatal("Failed to get first field reader")
	}

	// Navigate to first record
	err = v.First()
	if err != nil {
		b.Fatalf("Failed to go to first record: %v", err)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Read single field value repeatedly
		_, _ = firstField.AsString()
	}
}

func BenchmarkFieldReading_Conversion_Types(b *testing.B) {
	v := &Vulpo{}
	err := v.Open(detailDBF)
	if err != nil {
		b.Fatalf("Failed to open %s: %v", detailDBF, err)
	}
	defer v.Close()

	// Get first field
	fieldDefs := v.FieldDefs()
	if fieldDefs == nil || fieldDefs.Count() == 0 {
		b.Fatal("No fields found")
	}

	firstField := v.FieldReader(fieldDefs.ByIndex(0).Name())
	if firstField == nil {
		b.Fatal("Failed to get first field reader")
	}

	// Navigate to first record
	err = v.First()
	if err != nil {
		b.Fatalf("Failed to go to first record: %v", err)
	}

	b.Run("AsString", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = firstField.AsString()
		}
	})

	b.Run("AsInt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = firstField.AsInt()
		}
	})

	b.Run("AsFloat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = firstField.AsFloat()
		}
	})

	b.Run("AsBool", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = firstField.AsBool()
		}
	})

	b.Run("Value", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = firstField.Value()
		}
	})
}

// Benchmark field reading with navigation combinations
//
//nolint:gocyclo // TODO: simplify this benchmark function to reduce complexity
func BenchmarkFieldReading_Navigation_Patterns(b *testing.B) {
	v := &Vulpo{}
	err := v.Open(detailDBF)
	if err != nil {
		b.Fatalf("Failed to open %s: %v", detailDBF, err)
	}
	defer v.Close()

	// Get first field
	fieldDefs := v.FieldDefs()
	if fieldDefs == nil || fieldDefs.Count() == 0 {
		b.Fatal("No fields found")
	}

	firstField := v.FieldReader(fieldDefs.ByIndex(0).Name())
	if firstField == nil {
		b.Fatal("Failed to get first field reader")
	}

	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := v.First()
			if err != nil {
				b.Fatalf("Failed to go to first record: %v", err)
			}

			recordCount := 0
			for !v.EOF() && recordCount < 100 {
				_, _ = firstField.AsString()
				recordCount++
				if recordCount >= 100 {
					break
				}

				err = v.Next()
				if err != nil {
					break
				}
			}
		}
	})

	b.Run("Skip2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := v.First()
			if err != nil {
				b.Fatalf("Failed to go to first record: %v", err)
			}

			recordCount := 0
			for !v.EOF() && recordCount < 50 {
				_, _ = firstField.AsString()
				recordCount++
				if recordCount >= 50 {
					break
				}

				err = v.Skip(2) // Skip every other record
				if err != nil {
					break
				}
			}
		}
	})

	b.Run("Skip10", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := v.First()
			if err != nil {
				b.Fatalf("Failed to go to first record: %v", err)
			}

			recordCount := 0
			for !v.EOF() && recordCount < 20 {
				_, _ = firstField.AsString()
				recordCount++
				if recordCount >= 20 {
					break
				}

				err = v.Skip(10) // Skip 10 records at a time
				if err != nil {
					break
				}
			}
		}
	})
}
