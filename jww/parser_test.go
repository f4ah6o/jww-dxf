package jww

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestParse_ValidSignature(t *testing.T) {
	// Create minimal valid JWW data with signature and version
	data := createMinimalJWWData()
	r := bytes.NewReader(data)

	doc, err := Parse(r)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if doc == nil {
		t.Fatal("doc is nil")
	}
}

func TestParse_InvalidSignature(t *testing.T) {
	data := []byte("NotValid")
	r := bytes.NewReader(data)

	_, err := Parse(r)
	if err == nil {
		t.Fatal("expected error for invalid signature")
	}
	if err != ErrInvalidSignature {
		t.Errorf("expected ErrInvalidSignature, got: %v", err)
	}
}

func TestParse_Version(t *testing.T) {
	// Test with version 600 file
	testFile := filepath.Join("..", "examples", "jww", "敷地図.jww")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found:", testFile)
	}

	f, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	doc, err := Parse(f)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// 敷地図.jww has version 600
	if doc.Version != 600 {
		t.Errorf("got version %d, want 600", doc.Version)
	}
}

func TestParse_SampleFile_Shikichizu(t *testing.T) {
	// Test the only file that currently parses correctly
	testFile := filepath.Join("..", "examples", "jww", "敷地図.jww")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found:", testFile)
	}

	f, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	doc, err := Parse(f)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Basic validation
	if doc.Version == 0 {
		t.Error("version should not be 0")
	}

	// Check entities were parsed
	if len(doc.Entities) == 0 {
		t.Error("expected some entities")
	}

	// Count entity types
	lineCount := 0
	for _, e := range doc.Entities {
		switch e.Type() {
		case "LINE":
			lineCount++
		}
	}

	t.Logf("Parsed %d entities (Lines: %d)", len(doc.Entities), lineCount)

	// We know from jww-stats that 敷地図.jww has 9 lines
	if lineCount != 9 {
		t.Errorf("got %d lines, want 9", lineCount)
	}
}

func TestParse_LayerGroups(t *testing.T) {
	testFile := filepath.Join("..", "examples", "jww", "敷地図.jww")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("test file not found:", testFile)
	}

	f, err := os.Open(testFile)
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	doc, err := Parse(f)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Should have 16 layer groups
	if len(doc.LayerGroups) != 16 {
		t.Errorf("got %d layer groups, want 16", len(doc.LayerGroups))
	}

	// Each layer group should have 16 layers
	for i, lg := range doc.LayerGroups {
		if len(lg.Layers) != 16 {
			t.Errorf("layer group %d has %d layers, want 16", i, len(lg.Layers))
		}
	}
}

func TestParseLine(t *testing.T) {
	// Create minimal line entity data
	// EntityBase (version 600): DWORD group + BYTE penStyle + WORD penColor + WORD penWidth + WORD layer + WORD layerGroup + WORD flag
	// Line: 4 doubles (startX, startY, endX, endY)
	data := make([]byte, 0)

	// EntityBase
	data = append(data, 0, 0, 0, 0) // group = 0
	data = append(data, 1)          // penStyle = 1
	data = append(data, 1, 0)       // penColor = 1
	data = append(data, 1, 0)       // penWidth = 1 (ver >= 351)
	data = append(data, 0, 0)       // layer = 0
	data = append(data, 0, 0)       // layerGroup = 0
	data = append(data, 0, 0)       // flag = 0

	// Line coordinates (doubles)
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0)    // startX = 0
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0)    // startY = 0
	data = append(data, 0, 0, 0, 0, 0, 0, 240, 63) // endX = 1.0
	data = append(data, 0, 0, 0, 0, 0, 0, 240, 63) // endY = 1.0

	r := NewReader(bytes.NewReader(data))
	line, err := parseLine(r, 600)
	if err != nil {
		t.Fatalf("parseLine failed: %v", err)
	}

	if line.StartX != 0 || line.StartY != 0 {
		t.Errorf("start point: got (%v, %v), want (0, 0)", line.StartX, line.StartY)
	}
	if line.EndX != 1.0 || line.EndY != 1.0 {
		t.Errorf("end point: got (%v, %v), want (1, 1)", line.EndX, line.EndY)
	}
}

func TestParseArc(t *testing.T) {
	data := make([]byte, 0)

	// EntityBase
	data = append(data, 0, 0, 0, 0) // group = 0
	data = append(data, 1)          // penStyle = 1
	data = append(data, 1, 0)       // penColor = 1
	data = append(data, 1, 0)       // penWidth = 1
	data = append(data, 0, 0)       // layer = 0
	data = append(data, 0, 0)       // layerGroup = 0
	data = append(data, 0, 0)       // flag = 0

	// Arc data
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0)         // centerX = 0
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0)         // centerY = 0
	data = append(data, 0, 0, 0, 0, 0, 0, 240, 63)      // radius = 1.0
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0)         // startAngle = 0
	data = append(data, 24, 45, 68, 84, 251, 33, 9, 64) // arcAngle = PI
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0)         // tiltAngle = 0
	data = append(data, 0, 0, 0, 0, 0, 0, 240, 63)      // flatness = 1.0 (circle)
	data = append(data, 0, 0, 0, 0)                     // fullCircle = false

	r := NewReader(bytes.NewReader(data))
	arc, err := parseArc(r, 600)
	if err != nil {
		t.Fatalf("parseArc failed: %v", err)
	}

	if arc.CenterX != 0 || arc.CenterY != 0 {
		t.Errorf("center: got (%v, %v), want (0, 0)", arc.CenterX, arc.CenterY)
	}
	if arc.Radius != 1.0 {
		t.Errorf("radius: got %v, want 1.0", arc.Radius)
	}
	if arc.IsFullCircle {
		t.Error("expected not full circle")
	}
}

func TestParsePoint(t *testing.T) {
	data := make([]byte, 0)

	// EntityBase
	data = append(data, 0, 0, 0, 0) // group = 0
	data = append(data, 1)          // penStyle = 1 (regular point, not 100)
	data = append(data, 1, 0)       // penColor = 1
	data = append(data, 1, 0)       // penWidth = 1
	data = append(data, 0, 0)       // layer = 0
	data = append(data, 0, 0)       // layerGroup = 0
	data = append(data, 0, 0)       // flag = 0

	// Point data
	data = append(data, 0, 0, 0, 0, 0, 0, 36, 64) // x = 10.0
	data = append(data, 0, 0, 0, 0, 0, 0, 52, 64) // y = 20.0
	data = append(data, 0, 0, 0, 0)               // isTemporary = false

	r := NewReader(bytes.NewReader(data))
	pt, err := parsePoint(r, 600)
	if err != nil {
		t.Fatalf("parsePoint failed: %v", err)
	}

	if pt.X != 10.0 || pt.Y != 20.0 {
		t.Errorf("point: got (%v, %v), want (10, 20)", pt.X, pt.Y)
	}
	if pt.IsTemporary {
		t.Error("expected not temporary")
	}
}

func TestParseText(t *testing.T) {
	data := make([]byte, 0)

	// EntityBase
	data = append(data, 0, 0, 0, 0) // group = 0
	data = append(data, 1)          // penStyle = 1
	data = append(data, 1, 0)       // penColor = 1
	data = append(data, 1, 0)       // penWidth = 1
	data = append(data, 0, 0)       // layer = 0
	data = append(data, 0, 0)       // layerGroup = 0
	data = append(data, 0, 0)       // flag = 0

	// Text data
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0)    // startX = 0
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0)    // startY = 0
	data = append(data, 0, 0, 0, 0, 0, 0, 240, 63) // endX = 1.0
	data = append(data, 0, 0, 0, 0, 0, 0, 240, 63) // endY = 1.0
	data = append(data, 1, 0, 0, 0)                // textType = 1
	data = append(data, 0, 0, 0, 0, 0, 0, 20, 64)  // sizeX = 5.0
	data = append(data, 0, 0, 0, 0, 0, 0, 20, 64)  // sizeY = 5.0
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0)    // spacing = 0
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0)    // angle = 0

	// Font name (CString): "Arial"
	data = append(data, 5, 'A', 'r', 'i', 'a', 'l')

	// Content (CString): "Hello"
	data = append(data, 5, 'H', 'e', 'l', 'l', 'o')

	r := NewReader(bytes.NewReader(data))
	txt, err := parseText(r, 600)
	if err != nil {
		t.Fatalf("parseText failed: %v", err)
	}

	if txt.FontName != "Arial" {
		t.Errorf("fontName: got %q, want %q", txt.FontName, "Arial")
	}
	if txt.Content != "Hello" {
		t.Errorf("content: got %q, want %q", txt.Content, "Hello")
	}
}

func TestParseSolid(t *testing.T) {
	data := make([]byte, 0)

	// EntityBase
	data = append(data, 0, 0, 0, 0) // group = 0
	data = append(data, 1)          // penStyle = 1
	data = append(data, 1, 0)       // penColor = 1 (not 10, so no RGB)
	data = append(data, 1, 0)       // penWidth = 1
	data = append(data, 0, 0)       // layer = 0
	data = append(data, 0, 0)       // layerGroup = 0
	data = append(data, 0, 0)       // flag = 0

	// Solid data (8 doubles = 64 bytes)
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0)    // point1X = 0
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0)    // point1Y = 0
	data = append(data, 0, 0, 0, 0, 0, 0, 240, 63) // point4X = 1.0
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0)    // point4Y = 0
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0)    // point2X = 0
	data = append(data, 0, 0, 0, 0, 0, 0, 240, 63) // point2Y = 1.0
	data = append(data, 0, 0, 0, 0, 0, 0, 240, 63) // point3X = 1.0
	data = append(data, 0, 0, 0, 0, 0, 0, 240, 63) // point3Y = 1.0

	r := NewReader(bytes.NewReader(data))
	solid, err := parseSolid(r, 600)
	if err != nil {
		t.Fatalf("parseSolid failed: %v", err)
	}

	if solid.Point1X != 0 || solid.Point1Y != 0 {
		t.Errorf("point1: got (%v, %v), want (0, 0)", solid.Point1X, solid.Point1Y)
	}
	if solid.Point3X != 1.0 || solid.Point3Y != 1.0 {
		t.Errorf("point3: got (%v, %v), want (1, 1)", solid.Point3X, solid.Point3Y)
	}
}

func TestParseBlock(t *testing.T) {
	data := make([]byte, 0)

	// EntityBase
	data = append(data, 0, 0, 0, 0) // group = 0
	data = append(data, 1)          // penStyle = 1
	data = append(data, 1, 0)       // penColor = 1
	data = append(data, 1, 0)       // penWidth = 1
	data = append(data, 0, 0)       // layer = 0
	data = append(data, 0, 0)       // layerGroup = 0
	data = append(data, 0, 0)       // flag = 0

	// Block data
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0)    // refX = 0
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0)    // refY = 0
	data = append(data, 0, 0, 0, 0, 0, 0, 240, 63) // scaleX = 1.0
	data = append(data, 0, 0, 0, 0, 0, 0, 240, 63) // scaleY = 1.0
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0)    // rotation = 0
	data = append(data, 1, 0, 0, 0)                // defNumber = 1

	r := NewReader(bytes.NewReader(data))
	block, err := parseBlock(r, 600)
	if err != nil {
		t.Fatalf("parseBlock failed: %v", err)
	}

	if block.ScaleX != 1.0 || block.ScaleY != 1.0 {
		t.Errorf("scale: got (%v, %v), want (1, 1)", block.ScaleX, block.ScaleY)
	}
	if block.DefNumber != 1 {
		t.Errorf("defNumber: got %d, want 1", block.DefNumber)
	}
}

func TestParse_BlockDefinitionsAreParsedAfterEntities(t *testing.T) {
	data := createMinimalJWWDataWithBlockDef()

	doc, err := Parse(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(doc.BlockDefs) != 1 {
		t.Fatalf("expected 1 block definition, got %d", len(doc.BlockDefs))
	}

	def := doc.BlockDefs[0]
	if def.Number != 1 {
		t.Errorf("block def number: got %d, want 1", def.Number)
	}
	if def.Name != "BLK" {
		t.Errorf("block def name: got %q, want %q", def.Name, "BLK")
	}
}

// createMinimalJWWData creates minimal valid JWW file data for testing
func createMinimalJWWData() []byte {
	data := make([]byte, 0, 15000)

	// Signature
	data = append(data, []byte("JwwData.")...)

	// Version (600)
	data = append(data, 88, 2, 0, 0)

	// Memo (empty)
	data = append(data, 0)

	// Paper size
	data = append(data, 3, 0, 0, 0) // A3

	// Write layer group
	data = append(data, 0, 0, 0, 0)

	// 16 layer groups x (state + writeLayer + scale + protect + 16 layers x (state + protect))
	for i := 0; i < 16; i++ {
		data = append(data, 2, 0, 0, 0) // state = editable
		data = append(data, 0, 0, 0, 0) // writeLayer = 0
		// scale = 1.0 (double)
		data = append(data, 0, 0, 0, 0, 0, 0, 240, 63)
		data = append(data, 0, 0, 0, 0) // protect = 0

		for j := 0; j < 16; j++ {
			data = append(data, 2, 0, 0, 0) // layer state = editable
			data = append(data, 0, 0, 0, 0) // layer protect = 0
		}
	}

	// Pad with zeros to get past header to entity list area
	padding := make([]byte, 10000)
	data = append(data, padding...)

	// Entity list: count = 1, followed by one line entity
	// Position this where the scanner will find it
	// Count (WORD)
	data = append(data, 1, 0)

	// New class definition (0xFFFF)
	data = append(data, 0xFF, 0xFF)
	// Schema (600 in little-endian)
	data = append(data, 88, 2)
	// Class name length + name
	data = append(data, 8, 0) // length = 8
	data = append(data, []byte("CDataSen")...)

	// EntityBase for line
	data = append(data, 0, 0, 0, 0) // group
	data = append(data, 1)          // penStyle
	data = append(data, 1, 0)       // penColor
	data = append(data, 1, 0)       // penWidth
	data = append(data, 0, 0)       // layer
	data = append(data, 0, 0)       // layerGroup
	data = append(data, 0, 0)       // flag

	// Line coordinates
	for i := 0; i < 4; i++ {
		data = append(data, 0, 0, 0, 0, 0, 0, 0, 0) // 4 doubles = 0
	}

	return data
}

func createMinimalJWWDataWithBlockDef() []byte {
	data := createMinimalJWWData()

	var buf bytes.Buffer

	// Block definition count
	_ = binary.Write(&buf, binary.LittleEndian, uint32(1))

	// Class definition header for block definition (CDataList)
	_ = binary.Write(&buf, binary.LittleEndian, uint16(0xFFFF))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(600)) // schema version
	nameBytes := []byte("CDataList")
	_ = binary.Write(&buf, binary.LittleEndian, uint16(len(nameBytes)))
	buf.Write(nameBytes)

	// EntityBase
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0)) // group
	buf.WriteByte(1)                                       // penStyle
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1)) // penColor
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1)) // penWidth
	_ = binary.Write(&buf, binary.LittleEndian, uint16(0)) // layer
	_ = binary.Write(&buf, binary.LittleEndian, uint16(0)) // layerGroup
	_ = binary.Write(&buf, binary.LittleEndian, uint16(0)) // flag

	// Block definition fields
	_ = binary.Write(&buf, binary.LittleEndian, uint32(1)) // Number
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0)) // IsReferenced
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0)) // CTime

	// Name CString: "BLK"
	blockName := []byte("BLK")
	buf.WriteByte(byte(len(blockName)))
	buf.Write(blockName)

	// Nested entity list (count = 0)
	_ = binary.Write(&buf, binary.LittleEndian, uint16(0))

	return append(data, buf.Bytes()...)
}
