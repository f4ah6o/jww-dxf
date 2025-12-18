// Package dxf provides types and generation functions for DXF format.
package dxf

// Document represents a DXF document structure.
type Document struct {
	Layers   []Layer
	Entities []Entity
	Blocks   []Block
}

// Layer represents a layer definition.
type Layer struct {
	Name     string
	Color    int // ACI color (1-255)
	LineType string
	Frozen   bool
	Locked   bool
}

// Entity is the interface for all DXF entities.
type Entity interface {
	EntityType() string
	GroupCodes() []GroupCode
}

// GroupCode represents a DXF group code/value pair.
type GroupCode struct {
	Code  int
	Value interface{}
}

// Line represents a LINE entity.
type Line struct {
	Layer    string
	Color    int // 0 = BYLAYER
	X1, Y1   float64
	X2, Y2   float64
	LineType string
}

func (l *Line) EntityType() string { return "LINE" }

func (l *Line) GroupCodes() []GroupCode {
	return []GroupCode{
		{0, "LINE"},
		{8, l.Layer},
		{62, l.Color},
		{10, l.X1},
		{20, l.Y1},
		{30, 0.0},
		{11, l.X2},
		{21, l.Y2},
		{31, 0.0},
	}
}

// Circle represents a CIRCLE entity.
type Circle struct {
	Layer   string
	Color   int
	CenterX float64
	CenterY float64
	Radius  float64
}

func (c *Circle) EntityType() string { return "CIRCLE" }

func (c *Circle) GroupCodes() []GroupCode {
	return []GroupCode{
		{0, "CIRCLE"},
		{8, c.Layer},
		{62, c.Color},
		{10, c.CenterX},
		{20, c.CenterY},
		{30, 0.0},
		{40, c.Radius},
	}
}

// Arc represents an ARC entity.
type Arc struct {
	Layer      string
	Color      int
	CenterX    float64
	CenterY    float64
	Radius     float64
	StartAngle float64 // degrees
	EndAngle   float64 // degrees
}

func (a *Arc) EntityType() string { return "ARC" }

func (a *Arc) GroupCodes() []GroupCode {
	return []GroupCode{
		{0, "ARC"},
		{8, a.Layer},
		{62, a.Color},
		{10, a.CenterX},
		{20, a.CenterY},
		{30, 0.0},
		{40, a.Radius},
		{50, a.StartAngle},
		{51, a.EndAngle},
	}
}

// Ellipse represents an ELLIPSE entity.
type Ellipse struct {
	Layer      string
	Color      int
	CenterX    float64
	CenterY    float64
	MajorAxisX float64 // Endpoint of major axis relative to center
	MajorAxisY float64
	MinorRatio float64 // Ratio of minor to major axis
	StartParam float64 // Start parameter (0.0 for full ellipse)
	EndParam   float64 // End parameter (2*PI for full ellipse)
}

func (e *Ellipse) EntityType() string { return "ELLIPSE" }

func (e *Ellipse) GroupCodes() []GroupCode {
	return []GroupCode{
		{0, "ELLIPSE"},
		{8, e.Layer},
		{62, e.Color},
		{10, e.CenterX},
		{20, e.CenterY},
		{30, 0.0},
		{11, e.MajorAxisX},
		{21, e.MajorAxisY},
		{31, 0.0},
		{40, e.MinorRatio},
		{41, e.StartParam},
		{42, e.EndParam},
	}
}

// Point represents a POINT entity.
type Point struct {
	Layer string
	Color int
	X, Y  float64
}

func (p *Point) EntityType() string { return "POINT" }

func (p *Point) GroupCodes() []GroupCode {
	return []GroupCode{
		{0, "POINT"},
		{8, p.Layer},
		{62, p.Color},
		{10, p.X},
		{20, p.Y},
		{30, 0.0},
	}
}

// Text represents a TEXT entity.
type Text struct {
	Layer    string
	Color    int
	X, Y     float64
	Height   float64
	Rotation float64 // degrees
	Content  string
	Style    string
}

func (t *Text) EntityType() string { return "TEXT" }

func (t *Text) GroupCodes() []GroupCode {
	codes := []GroupCode{
		{0, "TEXT"},
		{8, t.Layer},
		{62, t.Color},
		{10, t.X},
		{20, t.Y},
		{30, 0.0},
		{40, t.Height},
		{1, t.Content},
	}
	if t.Rotation != 0 {
		codes = append(codes, GroupCode{50, t.Rotation})
	}
	if t.Style != "" {
		codes = append(codes, GroupCode{7, t.Style})
	}
	return codes
}

// Solid represents a SOLID entity (filled triangle/quadrilateral).
type Solid struct {
	Layer  string
	Color  int
	X1, Y1 float64
	X2, Y2 float64
	X3, Y3 float64
	X4, Y4 float64
}

func (s *Solid) EntityType() string { return "SOLID" }

func (s *Solid) GroupCodes() []GroupCode {
	return []GroupCode{
		{0, "SOLID"},
		{8, s.Layer},
		{62, s.Color},
		{10, s.X1},
		{20, s.Y1},
		{30, 0.0},
		{11, s.X2},
		{21, s.Y2},
		{31, 0.0},
		{12, s.X3},
		{22, s.Y3},
		{32, 0.0},
		{13, s.X4},
		{23, s.Y4},
		{33, 0.0},
	}
}

// Insert represents an INSERT (block reference) entity.
type Insert struct {
	Layer     string
	Color     int
	BlockName string
	X, Y      float64
	ScaleX    float64
	ScaleY    float64
	Rotation  float64 // degrees
}

func (i *Insert) EntityType() string { return "INSERT" }

func (i *Insert) GroupCodes() []GroupCode {
	return []GroupCode{
		{0, "INSERT"},
		{8, i.Layer},
		{62, i.Color},
		{2, i.BlockName},
		{10, i.X},
		{20, i.Y},
		{30, 0.0},
		{41, i.ScaleX},
		{42, i.ScaleY},
		{43, 1.0}, // ScaleZ
		{50, i.Rotation},
	}
}

// Block represents a block definition.
type Block struct {
	Name     string
	BaseX    float64
	BaseY    float64
	Entities []Entity
}
