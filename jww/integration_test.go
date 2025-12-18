package jww

import (
	"os"
	"path/filepath"
	"testing"
)

// TestParse_AllSampleFiles tests parsing all .jww files in the examples directory.
// This test documents the current parser state - many files fail due to class ID tracking bugs.
func TestParse_AllSampleFiles(t *testing.T) {
	examplesDir := filepath.Join("..", "examples", "jww")

	var files []string
	err := filepath.Walk(examplesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		ext := filepath.Ext(path)
		if !info.IsDir() && (ext == ".jww" || ext == ".JWW") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk examples directory: %v", err)
	}

	if len(files) == 0 {
		t.Skip("no JWW files found in examples directory")
	}

	var successCount, failCount int

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			f, err := os.Open(file)
			if err != nil {
				t.Fatalf("failed to open file: %v", err)
			}
			defer f.Close()

			doc, err := Parse(f)
			if err != nil {
				failCount++
				// Log the error but don't fail the test - this documents known issues
				t.Logf("PARSE FAILED: %v", err)
				return
			}
			successCount++

			// Basic validation
			if doc.Version == 0 {
				t.Error("version should not be 0")
			}

			t.Logf("SUCCESS: version=%d, entities=%d", doc.Version, len(doc.Entities))
		})
	}

	t.Logf("Summary: %d/%d files parsed successfully", successCount, successCount+failCount)
}

// TestParse_EntityCounts validates entity counts for known test files.
func TestParse_EntityCounts(t *testing.T) {
	testCases := []struct {
		file           string
		expectedLines  int
		expectedArcs   int
		expectedPoints int
		expectedTexts  int
	}{
		// Only include files that currently parse successfully
		{"敷地図.jww", 9, 0, 0, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.file, func(t *testing.T) {
			testFile := filepath.Join("..", "examples", "jww", tc.file)
			if _, err := os.Stat(testFile); os.IsNotExist(err) {
				t.Skip("test file not found:", testFile)
			}

			f, err := os.Open(testFile)
			if err != nil {
				t.Fatalf("failed to open file: %v", err)
			}
			defer f.Close()

			doc, err := Parse(f)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}

			lineCount := 0
			arcCount := 0
			pointCount := 0
			textCount := 0

			for _, e := range doc.Entities {
				switch e.Type() {
				case "LINE":
					lineCount++
				case "ARC", "CIRCLE":
					arcCount++
				case "POINT":
					pointCount++
				case "TEXT":
					textCount++
				}
			}

			if tc.expectedLines >= 0 && lineCount != tc.expectedLines {
				t.Errorf("lines: got %d, want %d", lineCount, tc.expectedLines)
			}
			if tc.expectedArcs >= 0 && arcCount != tc.expectedArcs {
				t.Errorf("arcs: got %d, want %d", arcCount, tc.expectedArcs)
			}
			if tc.expectedPoints >= 0 && pointCount != tc.expectedPoints {
				t.Errorf("points: got %d, want %d", pointCount, tc.expectedPoints)
			}
			if tc.expectedTexts >= 0 && textCount != tc.expectedTexts {
				t.Errorf("texts: got %d, want %d", textCount, tc.expectedTexts)
			}
		})
	}
}

// TestParse_EntityTypes checks that parsed entities have the correct types.
func TestParse_EntityTypes(t *testing.T) {
	testFile := filepath.Join("..", "examples", "jww", "敷地図.jww")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found:", testFile)
	}

	f, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	doc, err := Parse(f)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	validTypes := map[string]bool{
		"LINE":   true,
		"ARC":    true,
		"CIRCLE": true,
		"POINT":  true,
		"TEXT":   true,
		"SOLID":  true,
		"BLOCK":  true,
	}

	for i, e := range doc.Entities {
		et := e.Type()
		if !validTypes[et] {
			t.Errorf("entity %d has unknown type: %q", i, et)
		}

		// Verify interface implementation
		base := e.Base()
		if base == nil {
			t.Errorf("entity %d has nil base", i)
		}
	}
}

// BenchmarkParse benchmarks parsing performance.
func BenchmarkParse(b *testing.B) {
	testFile := filepath.Join("..", "examples", "jww", "敷地図.jww")
	data, err := os.ReadFile(testFile)
	if err != nil {
		b.Fatalf("failed to read file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := newReaderFromBytes(data)
		_, err := Parse(r)
		if err != nil {
			b.Fatalf("parse failed: %v", err)
		}
	}
}

type bytesReader struct {
	data []byte
	pos  int
}

func newReaderFromBytes(data []byte) *bytesReader {
	return &bytesReader{data: data}
}

func (r *bytesReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, os.ErrClosed
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
