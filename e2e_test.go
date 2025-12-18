package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/f4ah6o/jww-dxf/dxf"
	"github.com/f4ah6o/jww-dxf/jww"
)

// TestE2E_ConvertSampleFile tests full JWW to DXF conversion pipeline.
func TestE2E_ConvertSampleFile(t *testing.T) {
	testFile := filepath.Join("examples", "jww", "敷地図.jww")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found:", testFile)
	}

	// Parse JWW file
	f, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	jwwDoc, err := jww.Parse(f)
	if err != nil {
		t.Fatalf("JWW parse failed: %v", err)
	}

	// Convert to DXF
	dxfDoc := dxf.ConvertDocument(jwwDoc)
	if dxfDoc == nil {
		t.Fatal("DXF conversion returned nil")
	}

	// Verify conversion results
	t.Logf("Converted: %d layers, %d entities, %d blocks",
		len(dxfDoc.Layers), len(dxfDoc.Entities), len(dxfDoc.Blocks))

	// Should have 256 layers (16 groups x 16 layers)
	if len(dxfDoc.Layers) != 256 {
		t.Errorf("layers: got %d, want 256", len(dxfDoc.Layers))
	}

	// Should have converted entities
	if len(dxfDoc.Entities) == 0 {
		t.Error("expected some entities")
	}

	// Verify entity types are valid DXF types
	for i, e := range dxfDoc.Entities {
		entityType := e.EntityType()
		validTypes := []string{"LINE", "CIRCLE", "ARC", "ELLIPSE", "POINT", "TEXT", "SOLID", "INSERT"}
		valid := false
		for _, vt := range validTypes {
			if entityType == vt {
				valid = true
				break
			}
		}
		if !valid {
			t.Errorf("entity %d has invalid DXF type: %s", i, entityType)
		}
	}
}

// TestE2E_OutputValidDXF tests that the DXF output is valid format.
func TestE2E_OutputValidDXF(t *testing.T) {
	testFile := filepath.Join("examples", "jww", "敷地図.jww")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found:", testFile)
	}

	// Parse JWW file
	f, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	defer f.Close()

	jwwDoc, err := jww.Parse(f)
	if err != nil {
		t.Fatalf("JWW parse failed: %v", err)
	}

	// Convert to DXF
	dxfDoc := dxf.ConvertDocument(jwwDoc)

	// Write to temporary file
	tmpFile := filepath.Join(t.TempDir(), "output.dxf")
	outFile, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	err = dxf.NewWriter(outFile).WriteDocument(dxfDoc)
	if err != nil {
		t.Fatalf("DXF write failed: %v", err)
	}

	// Verify file was created
	fi, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	if fi.Size() == 0 {
		t.Error("output file is empty")
	}

	t.Logf("Output file size: %d bytes", fi.Size())

	// Read and verify DXF structure
	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	dxfContent := string(content)

	// Check for required DXF sections
	requiredSections := []string{
		"SECTION", "HEADER", "ENDSEC",
		"TABLES", "LAYER",
		"ENTITIES",
		"EOF",
	}

	for _, section := range requiredSections {
		if !strings.Contains(dxfContent, section) {
			t.Errorf("DXF output missing required section/keyword: %s", section)
		}
	}
}

// TestE2E_ConvertAllFiles attempts to convert all parseable files.
func TestE2E_ConvertAllFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	examplesDir := filepath.Join("examples", "jww")

	var files []string
	filepath.Walk(examplesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		ext := filepath.Ext(path)
		if !info.IsDir() && (ext == ".jww" || ext == ".JWW") {
			files = append(files, path)
		}
		return nil
	})

	var successCount, failCount int

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			f, err := os.Open(file)
			if err != nil {
				t.Fatalf("failed to open: %v", err)
			}
			defer f.Close()

			jwwDoc, err := jww.Parse(f)
			if err != nil {
				failCount++
				t.Logf("PARSE FAILED: %v", err)
				return
			}

			dxfDoc := dxf.ConvertDocument(jwwDoc)
			if dxfDoc == nil {
				failCount++
				t.Log("CONVERT FAILED: nil document")
				return
			}

			successCount++
			t.Logf("SUCCESS: %d entities converted", len(dxfDoc.Entities))
		})
	}

	t.Logf("Summary: %d/%d files converted successfully", successCount, successCount+failCount)
}

// BenchmarkE2E_FullPipeline benchmarks the full JWW to DXF conversion.
func BenchmarkE2E_FullPipeline(b *testing.B) {
	testFile := filepath.Join("examples", "jww", "敷地図.jww")
	data, err := os.ReadFile(testFile)
	if err != nil {
		b.Fatalf("failed to read file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f := newBytesReader(data)
		jwwDoc, err := jww.Parse(f)
		if err != nil {
			b.Fatalf("parse failed: %v", err)
		}
		_ = dxf.ConvertDocument(jwwDoc)
	}
}

type bytesReaderE2E struct {
	data []byte
	pos  int
}

func newBytesReader(data []byte) *bytesReaderE2E {
	return &bytesReaderE2E{data: data}
}

func (r *bytesReaderE2E) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, os.ErrClosed
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
