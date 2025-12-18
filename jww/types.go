package jww

// Document represents a JWW file.
type Document struct {
	Version         uint32
	Memo            string
	PaperSize       uint32 // 0-4: A0-A4, 8: 2A, 9: 3A, etc.
	WriteLayerGroup uint32
	LayerGroups     [16]LayerGroup
	Entities        []Entity
	BlockDefs       []BlockDef
}

// LayerGroup represents a layer group (レイヤグループ).
type LayerGroup struct {
	State      uint32  // 0: hidden, 1: display only, 2: editable, 3: write
	WriteLayer uint32  // Current write layer
	Scale      float64 // Scale denominator
	Protect    uint32  // Protection flag
	Layers     [16]Layer
	Name       string
}

// Layer represents a layer within a layer group.
type Layer struct {
	State   uint32 // 0: hidden, 1: display only, 2: editable, 3: write
	Protect uint32
	Name    string
}

// EntityBase contains common attributes for all entities.
type EntityBase struct {
	Group      uint32 // Curve attribute number
	PenStyle   byte   // Line type number
	PenColor   uint16 // Line color number
	PenWidth   uint16 // Line width (Ver.3.51+)
	Layer      uint16 // Layer number
	LayerGroup uint16 // Layer group number
	Flag       uint16 // Attribute flag
}

// Entity is the interface for all JWW entities.
type Entity interface {
	Base() *EntityBase
	Type() string
}

// Line represents a line entity (CDataSen).
type Line struct {
	EntityBase
	StartX, StartY float64
	EndX, EndY     float64
}

func (l *Line) Base() *EntityBase { return &l.EntityBase }
func (l *Line) Type() string      { return "LINE" }

// Arc represents an arc/circle entity (CDataEnko).
type Arc struct {
	EntityBase
	CenterX, CenterY float64
	Radius           float64
	StartAngle       float64 // radians
	ArcAngle         float64 // radians
	TiltAngle        float64 // radians
	Flatness         float64 // ellipse ratio (1.0 for circle)
	IsFullCircle     bool
}

func (a *Arc) Base() *EntityBase { return &a.EntityBase }
func (a *Arc) Type() string {
	if a.IsFullCircle {
		return "CIRCLE"
	}
	return "ARC"
}

// Point represents a point entity (CDataTen).
type Point struct {
	EntityBase
	X, Y        float64
	IsTemporary bool   // 仮点
	Code        uint32 // Point code (arrow, marker, etc.)
	Angle       float64
	Scale       float64
}

func (p *Point) Base() *EntityBase { return &p.EntityBase }
func (p *Point) Type() string      { return "POINT" }

// Text represents a text entity (CDataMoji).
type Text struct {
	EntityBase
	StartX, StartY float64
	EndX, EndY     float64
	TextType       uint32 // +10000: italic, +20000: bold
	SizeX, SizeY   float64
	Spacing        float64
	Angle          float64 // degrees
	FontName       string
	Content        string
}

func (t *Text) Base() *EntityBase { return &t.EntityBase }
func (t *Text) Type() string      { return "TEXT" }

// Solid represents a solid fill entity (CDataSolid).
type Solid struct {
	EntityBase
	Point1X, Point1Y float64 // First point
	Point2X, Point2Y float64 // Second point
	Point3X, Point3Y float64 // Third point
	Point4X, Point4Y float64 // Fourth point
	Color            uint32  // RGB (when PenColor == 10)
}

func (s *Solid) Base() *EntityBase { return &s.EntityBase }
func (s *Solid) Type() string      { return "SOLID" }

// Block represents a block insert (CDataBlock).
type Block struct {
	EntityBase
	RefX, RefY float64 // Reference point
	ScaleX     float64
	ScaleY     float64
	Rotation   float64 // radians
	DefNumber  uint32  // Block definition number
}

func (b *Block) Base() *EntityBase { return &b.EntityBase }
func (b *Block) Type() string      { return "BLOCK" }

// BlockDef represents a block definition (CDataList).
type BlockDef struct {
	EntityBase
	Number       uint32
	IsReferenced bool
	// Time       time.Time // Creation time
	Name     string
	Entities []Entity
}
