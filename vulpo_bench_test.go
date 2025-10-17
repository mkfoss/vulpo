package vulpo

import (
	"testing"
)

const benchDBFPath = "mkfdbflib/data/info.dbf"

// BenchmarkVulpo_OpenClose measures the performance of opening and closing DBF files
func BenchmarkVulpo_OpenClose(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := &Vulpo{}

		err := v.Open(benchDBFPath)
		if err != nil {
			b.Fatalf("Failed to open file: %v", err)
		}

		err = v.Close()
		if err != nil {
			b.Fatalf("Failed to close file: %v", err)
		}
	}
}

// BenchmarkVulpo_HeaderAccess measures the performance of accessing header information
func BenchmarkVulpo_HeaderAccess(b *testing.B) {
	v := &Vulpo{}

	err := v.Open(benchDBFPath)
	if err != nil {
		b.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		header := v.Header()

		// Access all header fields to ensure they're computed
		_ = header.RecordCount()
		_ = header.LastUpdated()
		_ = header.HasIndex()
		_ = header.HasFpt()
		_ = header.Codepage()
	}
}

// BenchmarkVulpo_Active measures the performance of the Active method
func BenchmarkVulpo_Active(b *testing.B) {
	v := &Vulpo{}

	err := v.Open(benchDBFPath)
	if err != nil {
		b.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		_ = v.Close()
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = v.Active()
	}
}

// BenchmarkCodepage_Methods measures the performance of codepage operations
func BenchmarkCodepage_Methods(b *testing.B) {
	cp := Codepage(0x03)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = cp.Name()
		_ = cp.String()
		_ = cp.VfpCodepageID()
		_ = cp.MsCodepageID()
		_ = cp.Supported()
	}
}
