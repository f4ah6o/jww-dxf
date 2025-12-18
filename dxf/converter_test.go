package dxf

import (
	"math"
	"testing"

	"github.com/f4ah6o/jww-dxf/jww"
)

func TestConvertLine(t *testing.T) {
	line := &jww.Line{
		EntityBase: jww.EntityBase{
			PenColor:   1,
			Layer:      0,
			LayerGroup: 0,
		},
		StartX: 0,
		StartY: 0,
		EndX:   100,
		EndY:   100,
	}

	doc := createTestDocument()
	doc.Entities = []jww.Entity{line}

	result := ConvertDocument(doc)

	if len(result.Entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(result.Entities))
	}

	dxfLine, ok := result.Entities[0].(*Line)
	if !ok {
		t.Fatalf("expected *Line, got %T", result.Entities[0])
	}

	if dxfLine.X1 != 0 || dxfLine.Y1 != 0 {
		t.Errorf("start: got (%v, %v), want (0, 0)", dxfLine.X1, dxfLine.Y1)
	}
	if dxfLine.X2 != 100 || dxfLine.Y2 != 100 {
		t.Errorf("end: got (%v, %v), want (100, 100)", dxfLine.X2, dxfLine.Y2)
	}
}

func TestConvertCircle(t *testing.T) {
	arc := &jww.Arc{
		EntityBase: jww.EntityBase{
			PenColor:   1,
			Layer:      0,
			LayerGroup: 0,
		},
		CenterX:      50,
		CenterY:      50,
		Radius:       25,
		IsFullCircle: true,
		Flatness:     1.0, // Circle (not ellipse)
	}

	doc := createTestDocument()
	doc.Entities = []jww.Entity{arc}

	result := ConvertDocument(doc)

	if len(result.Entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(result.Entities))
	}

	circle, ok := result.Entities[0].(*Circle)
	if !ok {
		t.Fatalf("expected *Circle, got %T", result.Entities[0])
	}

	if circle.CenterX != 50 || circle.CenterY != 50 {
		t.Errorf("center: got (%v, %v), want (50, 50)", circle.CenterX, circle.CenterY)
	}
	if circle.Radius != 25 {
		t.Errorf("radius: got %v, want 25", circle.Radius)
	}
}

func TestConvertArc(t *testing.T) {
	arc := &jww.Arc{
		EntityBase: jww.EntityBase{
			PenColor:   1,
			Layer:      0,
			LayerGroup: 0,
		},
		CenterX:      0,
		CenterY:      0,
		Radius:       10,
		StartAngle:   0,
		ArcAngle:     math.Pi / 2, // 90 degrees
		IsFullCircle: false,
		Flatness:     1.0,
	}

	doc := createTestDocument()
	doc.Entities = []jww.Entity{arc}

	result := ConvertDocument(doc)

	if len(result.Entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(result.Entities))
	}

	dxfArc, ok := result.Entities[0].(*Arc)
	if !ok {
		t.Fatalf("expected *Arc, got %T", result.Entities[0])
	}

	if dxfArc.Radius != 10 {
		t.Errorf("radius: got %v, want 10", dxfArc.Radius)
	}
	if math.Abs(dxfArc.StartAngle-0) > 0.001 {
		t.Errorf("startAngle: got %v, want 0", dxfArc.StartAngle)
	}
	if math.Abs(dxfArc.EndAngle-90) > 0.001 {
		t.Errorf("endAngle: got %v, want 90", dxfArc.EndAngle)
	}
}

func TestConvertEllipse(t *testing.T) {
	arc := &jww.Arc{
		EntityBase: jww.EntityBase{
			PenColor:   1,
			Layer:      0,
			LayerGroup: 0,
		},
		CenterX:      0,
		CenterY:      0,
		Radius:       10,  // Major radius
		Flatness:     0.5, // Minor/Major ratio
		TiltAngle:    0,
		IsFullCircle: true,
	}

	doc := createTestDocument()
	doc.Entities = []jww.Entity{arc}

	result := ConvertDocument(doc)

	if len(result.Entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(result.Entities))
	}

	ellipse, ok := result.Entities[0].(*Ellipse)
	if !ok {
		t.Fatalf("expected *Ellipse, got %T", result.Entities[0])
	}

	if ellipse.MinorRatio != 0.5 {
		t.Errorf("minorRatio: got %v, want 0.5", ellipse.MinorRatio)
	}
}

func TestConvertPoint(t *testing.T) {
	pt := &jww.Point{
		EntityBase: jww.EntityBase{
			PenColor:   1,
			Layer:      0,
			LayerGroup: 0,
		},
		X:           25,
		Y:           75,
		IsTemporary: false,
	}

	doc := createTestDocument()
	doc.Entities = []jww.Entity{pt}

	result := ConvertDocument(doc)

	if len(result.Entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(result.Entities))
	}

	dxfPoint, ok := result.Entities[0].(*Point)
	if !ok {
		t.Fatalf("expected *Point, got %T", result.Entities[0])
	}

	if dxfPoint.X != 25 || dxfPoint.Y != 75 {
		t.Errorf("point: got (%v, %v), want (25, 75)", dxfPoint.X, dxfPoint.Y)
	}
}

func TestConvertPoint_Temporary(t *testing.T) {
	// Temporary points should be skipped
	pt := &jww.Point{
		EntityBase: jww.EntityBase{
			PenColor:   1,
			Layer:      0,
			LayerGroup: 0,
		},
		X:           25,
		Y:           75,
		IsTemporary: true,
	}

	doc := createTestDocument()
	doc.Entities = []jww.Entity{pt}

	result := ConvertDocument(doc)

	if len(result.Entities) != 0 {
		t.Errorf("expected 0 entities (temporary point skipped), got %d", len(result.Entities))
	}
}

func TestConvertText(t *testing.T) {
	txt := &jww.Text{
		EntityBase: jww.EntityBase{
			PenColor:   1,
			Layer:      0,
			LayerGroup: 0,
		},
		StartX:   10,
		StartY:   20,
		SizeY:    5,
		Angle:    45,
		Content:  "Hello World",
		FontName: "Arial",
	}

	doc := createTestDocument()
	doc.Entities = []jww.Entity{txt}

	result := ConvertDocument(doc)

	if len(result.Entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(result.Entities))
	}

	dxfText, ok := result.Entities[0].(*Text)
	if !ok {
		t.Fatalf("expected *Text, got %T", result.Entities[0])
	}

	if dxfText.X != 10 || dxfText.Y != 20 {
		t.Errorf("position: got (%v, %v), want (10, 20)", dxfText.X, dxfText.Y)
	}
	if dxfText.Height != 5 {
		t.Errorf("height: got %v, want 5", dxfText.Height)
	}
	if dxfText.Content != "Hello World" {
		t.Errorf("content: got %q, want %q", dxfText.Content, "Hello World")
	}
}

func TestConvertSolid(t *testing.T) {
	solid := &jww.Solid{
		EntityBase: jww.EntityBase{
			PenColor:   1,
			Layer:      0,
			LayerGroup: 0,
		},
		Point1X: 0, Point1Y: 0,
		Point2X: 10, Point2Y: 0,
		Point3X: 10, Point3Y: 10,
		Point4X: 0, Point4Y: 10,
	}

	doc := createTestDocument()
	doc.Entities = []jww.Entity{solid}

	result := ConvertDocument(doc)

	if len(result.Entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(result.Entities))
	}

	dxfSolid, ok := result.Entities[0].(*Solid)
	if !ok {
		t.Fatalf("expected *Solid, got %T", result.Entities[0])
	}

	if dxfSolid.X1 != 0 || dxfSolid.Y1 != 0 {
		t.Errorf("point1: got (%v, %v), want (0, 0)", dxfSolid.X1, dxfSolid.Y1)
	}
}

func TestConvertBlock(t *testing.T) {
	block := &jww.Block{
		EntityBase: jww.EntityBase{
			PenColor:   1,
			Layer:      0,
			LayerGroup: 0,
		},
		RefX:      100,
		RefY:      100,
		ScaleX:    1.0,
		ScaleY:    1.0,
		Rotation:  math.Pi / 2, // 90 degrees in radians
		DefNumber: 1,
	}

	doc := createTestDocument()
	doc.BlockDefs = []jww.BlockDef{
		{
			EntityBase: jww.EntityBase{},
			Number:     1,
			Name:       "TestBlock",
		},
	}
	doc.Entities = []jww.Entity{block}

	result := ConvertDocument(doc)

	if len(result.Entities) != 1 {
		t.Fatalf("expected 1 entity, got %d", len(result.Entities))
	}

	insert, ok := result.Entities[0].(*Insert)
	if !ok {
		t.Fatalf("expected *Insert, got %T", result.Entities[0])
	}

	if insert.BlockName != "TestBlock" {
		t.Errorf("blockName: got %q, want %q", insert.BlockName, "TestBlock")
	}
	if insert.X != 100 || insert.Y != 100 {
		t.Errorf("position: got (%v, %v), want (100, 100)", insert.X, insert.Y)
	}
	if math.Abs(insert.Rotation-90) > 0.001 {
		t.Errorf("rotation: got %v, want 90", insert.Rotation)
	}
}

func TestMapColor(t *testing.T) {
	tests := []struct {
		jwwColor uint16
		expected int
	}{
		{0, 0},    // BYLAYER
		{1, 1},    // Red
		{2, 2},    // Yellow
		{7, 7},    // White
		{9, 9},    // Gray
		{100, 10}, // Extended color
		{150, 60}, // Extended color
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := mapColor(tt.jwwColor)
			if result != tt.expected {
				t.Errorf("mapColor(%d) = %d, want %d", tt.jwwColor, result, tt.expected)
			}
		})
	}
}

func TestConvertLayers(t *testing.T) {
	doc := createTestDocument()

	result := ConvertDocument(doc)

	// Should have 16 * 16 = 256 layers
	if len(result.Layers) != 256 {
		t.Errorf("expected 256 layers, got %d", len(result.Layers))
	}
}

func TestConvertBlocks(t *testing.T) {
	line := &jww.Line{
		EntityBase: jww.EntityBase{PenColor: 1},
		StartX:     0, StartY: 0,
		EndX: 10, EndY: 10,
	}

	doc := createTestDocument()
	doc.BlockDefs = []jww.BlockDef{
		{
			EntityBase: jww.EntityBase{},
			Number:     1,
			Name:       "Block1",
			Entities:   []jww.Entity{line},
		},
	}

	result := ConvertDocument(doc)

	if len(result.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(result.Blocks))
	}

	if result.Blocks[0].Name != "Block1" {
		t.Errorf("block name: got %q, want %q", result.Blocks[0].Name, "Block1")
	}

	if len(result.Blocks[0].Entities) != 1 {
		t.Errorf("block entities: got %d, want 1", len(result.Blocks[0].Entities))
	}
}

// createTestDocument creates a minimal JWW document for testing.
func createTestDocument() *jww.Document {
	doc := &jww.Document{
		Version: 600,
	}

	// Initialize all layer groups and layers
	for i := 0; i < 16; i++ {
		doc.LayerGroups[i] = jww.LayerGroup{
			State: 2, // Editable
			Scale: 1.0,
		}
		for j := 0; j < 16; j++ {
			doc.LayerGroups[i].Layers[j] = jww.Layer{
				State: 2,
			}
		}
	}

	return doc
}
