package jww

import (
	"fmt"
	"io"
)

// Parse reads a JWW file and returns a Document.
func Parse(r io.Reader) (*Document, error) {
	jr := NewReader(r)

	// Read and validate signature
	if err := jr.ReadSignature(); err != nil {
		return nil, err
	}

	doc := &Document{}

	// Read version
	version, err := jr.ReadDWORD()
	if err != nil {
		return nil, fmt.Errorf("reading version: %w", err)
	}
	doc.Version = version

	// Read file memo
	memo, err := jr.ReadCString()
	if err != nil {
		return nil, fmt.Errorf("reading memo: %w", err)
	}
	doc.Memo = memo

	// Read paper size
	paperSize, err := jr.ReadDWORD()
	if err != nil {
		return nil, fmt.Errorf("reading paper size: %w", err)
	}
	doc.PaperSize = paperSize

	// Read write layer group
	writeGLay, err := jr.ReadDWORD()
	if err != nil {
		return nil, fmt.Errorf("reading write layer group: %w", err)
	}
	doc.WriteLayerGroup = writeGLay

	// Read layer groups (16 groups)
	for gLay := 0; gLay < 16; gLay++ {
		lg := &doc.LayerGroups[gLay]

		// Layer group state
		state, err := jr.ReadDWORD()
		if err != nil {
			return nil, fmt.Errorf("reading layer group %d state: %w", gLay, err)
		}
		lg.State = state

		// Write layer
		writeLay, err := jr.ReadDWORD()
		if err != nil {
			return nil, fmt.Errorf("reading layer group %d write layer: %w", gLay, err)
		}
		lg.WriteLayer = writeLay

		// Scale
		scale, err := jr.ReadDouble()
		if err != nil {
			return nil, fmt.Errorf("reading layer group %d scale: %w", gLay, err)
		}
		lg.Scale = scale

		// Protection flag
		protect, err := jr.ReadDWORD()
		if err != nil {
			return nil, fmt.Errorf("reading layer group %d protect: %w", gLay, err)
		}
		lg.Protect = protect

		// Read 16 layers
		for lay := 0; lay < 16; lay++ {
			layState, err := jr.ReadDWORD()
			if err != nil {
				return nil, fmt.Errorf("reading layer %d-%d state: %w", gLay, lay, err)
			}
			lg.Layers[lay].State = layState

			layProtect, err := jr.ReadDWORD()
			if err != nil {
				return nil, fmt.Errorf("reading layer %d-%d protect: %w", gLay, lay, err)
			}
			lg.Layers[lay].Protect = layProtect
		}
	}

	// Dummy (14 DWORDs)
	if err := jr.Skip(14 * 4); err != nil {
		return nil, fmt.Errorf("skipping dummy: %w", err)
	}

	// Dimension settings (5 DWORDs)
	if err := jr.Skip(5 * 4); err != nil {
		return nil, fmt.Errorf("skipping dimension settings: %w", err)
	}

	// Dummy (1 DWORD)
	if err := jr.Skip(4); err != nil {
		return nil, fmt.Errorf("skipping dummy2: %w", err)
	}

	// Max line width (1 DWORD)
	if err := jr.Skip(4); err != nil {
		return nil, fmt.Errorf("skipping max line width: %w", err)
	}

	// Printer origin (2 doubles)
	if err := jr.Skip(16); err != nil {
		return nil, fmt.Errorf("skipping printer origin: %w", err)
	}

	// Printer scale (1 double)
	if err := jr.Skip(8); err != nil {
		return nil, fmt.Errorf("skipping printer scale: %w", err)
	}

	// Printer settings (1 DWORD)
	if err := jr.Skip(4); err != nil {
		return nil, fmt.Errorf("skipping printer settings: %w", err)
	}

	// Grid settings (1 DWORD + 5 doubles)
	if err := jr.Skip(4 + 40); err != nil {
		return nil, fmt.Errorf("skipping grid settings: %w", err)
	}

	// Layer names (16 * 16 CStrings)
	for gLay := 0; gLay < 16; gLay++ {
		for lay := 0; lay < 16; lay++ {
			name, err := jr.ReadCString()
			if err != nil {
				return nil, fmt.Errorf("reading layer name %d-%d: %w", gLay, lay, err)
			}
			doc.LayerGroups[gLay].Layers[lay].Name = name
		}
	}

	// Layer group names (16 CStrings)
	for gLay := 0; gLay < 16; gLay++ {
		name, err := jr.ReadCString()
		if err != nil {
			return nil, fmt.Errorf("reading layer group name %d: %w", gLay, err)
		}
		doc.LayerGroups[gLay].Name = name
	}

	// Shadow calculation settings (3 doubles + 1 DWORD + 1 double)
	if err := jr.Skip(32 + 4); err != nil {
		return nil, fmt.Errorf("skipping shadow settings: %w", err)
	}

	// Sky diagram settings (Ver.3.00+) (2 doubles)
	if version >= 300 {
		if err := jr.Skip(16); err != nil {
			return nil, fmt.Errorf("skipping sky settings: %w", err)
		}
	}

	// 2.5D calculation unit (1 DWORD)
	if err := jr.Skip(4); err != nil {
		return nil, fmt.Errorf("skipping 2.5D unit: %w", err)
	}

	// Screen scale and origin (3 doubles)
	if err := jr.Skip(24); err != nil {
		return nil, fmt.Errorf("skipping screen scale: %w", err)
	}

	// Range memory (3 doubles)
	if err := jr.Skip(24); err != nil {
		return nil, fmt.Errorf("skipping range memory: %w", err)
	}

	// Mark jump settings (Ver.3.00+: 8 sets, else: 4 sets)
	if version >= 300 {
		// 8 sets of (3 doubles + 1 DWORD)
		if err := jr.Skip(8 * (24 + 4)); err != nil {
			return nil, fmt.Errorf("skipping mark jump: %w", err)
		}
	} else {
		// 4 sets of (3 doubles)
		if err := jr.Skip(4 * 24); err != nil {
			return nil, fmt.Errorf("skipping mark jump: %w", err)
		}
	}

	// Text drawing settings (Ver.3.00+) (7 doubles + 1 DWORD or 4 + dummies)
	if version >= 300 {
		if err := jr.Skip(7*8 + 4); err != nil {
			return nil, fmt.Errorf("skipping text drawing settings: %w", err)
		}
	}

	// Multiple line spacing (10 doubles)
	if err := jr.Skip(80); err != nil {
		return nil, fmt.Errorf("skipping multiple line spacing: %w", err)
	}

	// Double-sided line end (1 double)
	if err := jr.Skip(8); err != nil {
		return nil, fmt.Errorf("skipping double-sided line end: %w", err)
	}

	// Pen colors and widths (10 sets of 2 DWORDs)
	if err := jr.Skip(10 * 8); err != nil {
		return nil, fmt.Errorf("skipping pen colors: %w", err)
	}

	// Printer pen colors, widths, point radius (10 sets of 2 DWORDs + 1 double)
	if err := jr.Skip(10 * 16); err != nil {
		return nil, fmt.Errorf("skipping printer pen settings: %w", err)
	}

	// Line types 2-9 (8 sets of 4 DWORDs)
	if err := jr.Skip(8 * 16); err != nil {
		return nil, fmt.Errorf("skipping line types: %w", err)
	}

	// Random lines 11-15 (5 sets of 5 DWORDs)
	if err := jr.Skip(5 * 20); err != nil {
		return nil, fmt.Errorf("skipping random lines: %w", err)
	}

	// Double-length line types 16-19 (4 sets of 4 DWORDs)
	if err := jr.Skip(4 * 16); err != nil {
		return nil, fmt.Errorf("skipping double-length line types: %w", err)
	}

	// Various draw settings (8 DWORDs)
	if err := jr.Skip(32); err != nil {
		return nil, fmt.Errorf("skipping draw settings: %w", err)
	}

	// Print settings and draw time (3 DWORDs)
	if err := jr.Skip(12); err != nil {
		return nil, fmt.Errorf("skipping print settings: %w", err)
	}

	// 2.5D view settings (3 DWORDs + 6 doubles)
	if err := jr.Skip(12 + 48); err != nil {
		return nil, fmt.Errorf("skipping 2.5D view settings: %w", err)
	}

	// Line length, box dimension, circle radius (4 doubles)
	if err := jr.Skip(32); err != nil {
		return nil, fmt.Errorf("skipping dimension values: %w", err)
	}

	// Solid color settings (2 DWORDs)
	if err := jr.Skip(8); err != nil {
		return nil, fmt.Errorf("skipping solid color settings: %w", err)
	}

	// SXF extended colors (Ver.4.20+)
	if version >= 420 {
		// 257 sets of (2 DWORDs) for screen colors
		if err := jr.Skip(257 * 8); err != nil {
			return nil, fmt.Errorf("skipping SXF screen colors: %w", err)
		}

		// 257 sets of (CString + 2 DWORDs + 1 double) for printer colors
		for n := 0; n <= 256; n++ {
			if _, err := jr.ReadCString(); err != nil {
				return nil, fmt.Errorf("skipping SXF color name %d: %w", n, err)
			}
			if err := jr.Skip(16); err != nil {
				return nil, fmt.Errorf("skipping SXF printer color %d: %w", n, err)
			}
		}

		// 33 sets of (4 DWORDs) for SXF line types
		if err := jr.Skip(33 * 16); err != nil {
			return nil, fmt.Errorf("skipping SXF line types: %w", err)
		}

		// 33 sets of (CString + 1 DWORD + 10 doubles) for SXF line type params
		for n := 0; n <= 32; n++ {
			if _, err := jr.ReadCString(); err != nil {
				return nil, fmt.Errorf("skipping SXF line type name %d: %w", n, err)
			}
			if err := jr.Skip(4 + 80); err != nil {
				return nil, fmt.Errorf("skipping SXF line type params %d: %w", n, err)
			}
		}
	}

	// Text style settings (10 sets of 3 doubles + 1 DWORD)
	if err := jr.Skip(10 * 28); err != nil {
		return nil, fmt.Errorf("skipping text styles: %w", err)
	}

	// Current text settings (3 doubles + 2 DWORDs)
	if err := jr.Skip(24 + 8); err != nil {
		return nil, fmt.Errorf("skipping current text settings: %w", err)
	}

	// Text line spacing (2 doubles)
	if err := jr.Skip(16); err != nil {
		return nil, fmt.Errorf("skipping text line spacing: %w", err)
	}

	// Text base point offset settings (1 DWORD + 6 doubles)
	if err := jr.Skip(4 + 48); err != nil {
		return nil, fmt.Errorf("skipping text base point offset: %w", err)
	}

	// Now parse the entity data list
	entities, err := parseEntityList(jr, version)
	if err != nil {
		return nil, fmt.Errorf("parsing entity list: %w", err)
	}
	doc.Entities = entities

	// Parse block definitions list
	blockDefs, err := parseBlockDefList(jr, version)
	if err != nil {
		return nil, fmt.Errorf("parsing block def list: %w", err)
	}
	doc.BlockDefs = blockDefs

	return doc, nil
}

// parseEntityList parses MFC CTypedPtrList<CObList, CData*>
func parseEntityList(jr *Reader, version uint32) ([]Entity, error) {
	// MFC CObList serialization format:
	// 1. DWORD: number of elements
	// 2. For each element:
	//    - WORD: class schema (0xFFFF for new class)
	//    - If new class: WORD schema version, WORD name length, string class name
	//    - Object data

	count, err := jr.ReadDWORD()
	if err != nil {
		return nil, fmt.Errorf("reading entity count: %w", err)
	}

	entities := make([]Entity, 0, count)
	classMap := make(map[uint16]string) // Map class ID to class name

	for i := uint32(0); i < count; i++ {
		entity, err := parseEntity(jr, version, classMap)
		if err != nil {
			return nil, fmt.Errorf("parsing entity %d: %w", i, err)
		}
		if entity != nil {
			entities = append(entities, entity)
		}
	}

	return entities, nil
}

// parseEntity parses a single entity from the MFC object stream
func parseEntity(jr *Reader, version uint32, classMap map[uint16]string) (Entity, error) {
	// Read class identifier
	classID, err := jr.ReadWORD()
	if err != nil {
		return nil, err
	}

	var className string

	if classID == 0xFFFF {
		// New class definition
		// Read schema version (WORD)
		_, err := jr.ReadWORD()
		if err != nil {
			return nil, fmt.Errorf("reading schema version: %w", err)
		}

		// Read class name length (WORD)
		nameLen, err := jr.ReadWORD()
		if err != nil {
			return nil, fmt.Errorf("reading class name length: %w", err)
		}

		// Read class name
		nameBuf := make([]byte, nameLen)
		if err := jr.ReadBytes(nameBuf); err != nil {
			return nil, fmt.Errorf("reading class name: %w", err)
		}
		className = string(nameBuf)

		// Assign new class ID (1-based index)
		newID := uint16(len(classMap) + 1)
		classMap[newID] = className
	} else if classID == 0x8000 {
		// Null object
		return nil, nil
	} else {
		// Existing class reference
		refID := classID & 0x7FFF
		var ok bool
		className, ok = classMap[refID]
		if !ok {
			return nil, fmt.Errorf("unknown class ID: %d", refID)
		}
	}

	// Parse based on class name
	switch className {
	case "CDataSen":
		return parseLine(jr, version)
	case "CDataEnko":
		return parseArc(jr, version)
	case "CDataTen":
		return parsePoint(jr, version)
	case "CDataMoji":
		return parseText(jr, version)
	case "CDataSolid":
		return parseSolid(jr, version)
	case "CDataBlock":
		return parseBlock(jr, version)
	case "CDataSunpou":
		return parseDimension(jr, version)
	default:
		return nil, fmt.Errorf("unknown entity class: %s", className)
	}
}

// parseBlockDefList parses the block definition list
func parseBlockDefList(jr *Reader, version uint32) ([]BlockDef, error) {
	count, err := jr.ReadDWORD()
	if err != nil {
		return nil, fmt.Errorf("reading block def count: %w", err)
	}

	blockDefs := make([]BlockDef, 0, count)
	classMap := make(map[uint16]string)

	for i := uint32(0); i < count; i++ {
		bd, err := parseBlockDef(jr, version, classMap)
		if err != nil {
			return nil, fmt.Errorf("parsing block def %d: %w", i, err)
		}
		if bd != nil {
			blockDefs = append(blockDefs, *bd)
		}
	}

	return blockDefs, nil
}

// parseBlockDef parses a single block definition
func parseBlockDef(jr *Reader, version uint32, classMap map[uint16]string) (*BlockDef, error) {
	// Read class identifier
	classID, err := jr.ReadWORD()
	if err != nil {
		return nil, err
	}

	if classID == 0xFFFF {
		// New class definition
		_, err := jr.ReadWORD() // schema
		if err != nil {
			return nil, err
		}
		nameLen, err := jr.ReadWORD()
		if err != nil {
			return nil, err
		}
		nameBuf := make([]byte, nameLen)
		if err := jr.ReadBytes(nameBuf); err != nil {
			return nil, err
		}
		newID := uint16(len(classMap) + 1)
		classMap[newID] = string(nameBuf)
	} else if classID == 0x8000 {
		return nil, nil
	}

	// Parse CDataList (block definition)
	base, err := parseEntityBase(jr, version)
	if err != nil {
		return nil, err
	}

	bd := &BlockDef{EntityBase: *base}

	// Block definition number
	bd.Number, err = jr.ReadDWORD()
	if err != nil {
		return nil, err
	}

	// Referenced flag
	ref, err := jr.ReadDWORD()
	if err != nil {
		return nil, err
	}
	bd.IsReferenced = ref != 0

	// Creation time (CTime - skip)
	if err := jr.Skip(4); err != nil {
		return nil, err
	}

	// Block name
	bd.Name, err = jr.ReadCString()
	if err != nil {
		return nil, err
	}

	// Block entities
	bd.Entities, err = parseEntityList(jr, version)
	if err != nil {
		return nil, err
	}

	return bd, nil
}

// parseDimension parses a dimension entity (CDataSunpou)
// For now, we skip the complex dimension data
func parseDimension(jr *Reader, version uint32) (Entity, error) {
	base, err := parseEntityBase(jr, version)
	if err != nil {
		return nil, err
	}
	_ = base

	// Parse the line member (CDataSen without class header)
	line, err := parseLine(jr, version)
	if err != nil {
		return nil, err
	}

	// Parse the text member (CDataMoji without class header)
	_, err = parseText(jr, version)
	if err != nil {
		return nil, err
	}

	// Ver.4.20+ has additional SXF mode data
	if version >= 420 {
		// SXF mode (WORD)
		_, err := jr.ReadWORD()
		if err != nil {
			return nil, err
		}

		// 2 helper lines, 2 points, 2 base points
		for i := 0; i < 2; i++ {
			if _, err := parseLine(jr, version); err != nil {
				return nil, err
			}
		}
		for i := 0; i < 4; i++ {
			if _, err := parsePoint(jr, version); err != nil {
				return nil, err
			}
		}
	}

	// Return the main line as the dimension representation
	return line, nil
}

// parseEntityBase reads the common entity base fields.
func parseEntityBase(jr *Reader, version uint32) (*EntityBase, error) {
	base := &EntityBase{}

	// Curve attribute number
	group, err := jr.ReadDWORD()
	if err != nil {
		return nil, err
	}
	base.Group = group

	// Line type
	penStyle, err := jr.ReadBYTE()
	if err != nil {
		return nil, err
	}
	base.PenStyle = penStyle

	// Line color
	penColor, err := jr.ReadWORD()
	if err != nil {
		return nil, err
	}
	base.PenColor = penColor

	// Line width (Ver.3.51+, version >= 351)
	if version >= 351 {
		penWidth, err := jr.ReadWORD()
		if err != nil {
			return nil, err
		}
		base.PenWidth = penWidth
	}

	// Layer
	layer, err := jr.ReadWORD()
	if err != nil {
		return nil, err
	}
	base.Layer = layer

	// Layer group
	layerGroup, err := jr.ReadWORD()
	if err != nil {
		return nil, err
	}
	base.LayerGroup = layerGroup

	// Attribute flag
	flag, err := jr.ReadWORD()
	if err != nil {
		return nil, err
	}
	base.Flag = flag

	return base, nil
}

// parseLine reads a line entity (CDataSen).
func parseLine(jr *Reader, version uint32) (*Line, error) {
	base, err := parseEntityBase(jr, version)
	if err != nil {
		return nil, err
	}

	line := &Line{EntityBase: *base}

	// Start point X
	line.StartX, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Start point Y
	line.StartY, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// End point X
	line.EndX, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// End point Y
	line.EndY, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	return line, nil
}

// parseArc reads an arc entity (CDataEnko).
func parseArc(jr *Reader, version uint32) (*Arc, error) {
	base, err := parseEntityBase(jr, version)
	if err != nil {
		return nil, err
	}

	arc := &Arc{EntityBase: *base}

	// Center point
	arc.CenterX, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}
	arc.CenterY, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Radius
	arc.Radius, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Start angle (radians)
	arc.StartAngle, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Arc angle (radians)
	arc.ArcAngle, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Tilt angle (radians)
	arc.TiltAngle, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Flatness ratio
	arc.Flatness, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Full circle flag
	fullCircle, err := jr.ReadDWORD()
	if err != nil {
		return nil, err
	}
	arc.IsFullCircle = fullCircle != 0

	return arc, nil
}

// parsePoint reads a point entity (CDataTen).
func parsePoint(jr *Reader, version uint32) (*Point, error) {
	base, err := parseEntityBase(jr, version)
	if err != nil {
		return nil, err
	}

	pt := &Point{EntityBase: *base}

	// Point coordinates
	pt.X, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}
	pt.Y, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Temporary point flag
	tmp, err := jr.ReadDWORD()
	if err != nil {
		return nil, err
	}
	pt.IsTemporary = tmp != 0

	// Extended point data (when PenStyle == 100)
	if base.PenStyle == 100 {
		code, err := jr.ReadDWORD()
		if err != nil {
			return nil, err
		}
		pt.Code = code

		pt.Angle, err = jr.ReadDouble()
		if err != nil {
			return nil, err
		}

		pt.Scale, err = jr.ReadDouble()
		if err != nil {
			return nil, err
		}
	}

	return pt, nil
}

// parseText reads a text entity (CDataMoji).
func parseText(jr *Reader, version uint32) (*Text, error) {
	base, err := parseEntityBase(jr, version)
	if err != nil {
		return nil, err
	}

	txt := &Text{EntityBase: *base}

	// Start point
	txt.StartX, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}
	txt.StartY, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// End point
	txt.EndX, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}
	txt.EndY, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Text type (+10000: italic, +20000: bold)
	textType, err := jr.ReadDWORD()
	if err != nil {
		return nil, err
	}
	txt.TextType = textType

	// Text size X
	txt.SizeX, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Text size Y
	txt.SizeY, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Spacing
	txt.Spacing, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Angle (degrees)
	txt.Angle, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Font name
	txt.FontName, err = jr.ReadCString()
	if err != nil {
		return nil, err
	}

	// Text content
	txt.Content, err = jr.ReadCString()
	if err != nil {
		return nil, err
	}

	return txt, nil
}

// parseSolid reads a solid entity (CDataSolid).
func parseSolid(jr *Reader, version uint32) (*Solid, error) {
	base, err := parseEntityBase(jr, version)
	if err != nil {
		return nil, err
	}

	solid := &Solid{EntityBase: *base}

	// Point 1 (start)
	solid.Point1X, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}
	solid.Point1Y, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Point 4 (end)
	solid.Point4X, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}
	solid.Point4Y, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Point 2
	solid.Point2X, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}
	solid.Point2Y, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Point 3
	solid.Point3X, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}
	solid.Point3Y, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// RGB color (only when PenColor == 10)
	if base.PenColor == 10 {
		color, err := jr.ReadDWORD()
		if err != nil {
			return nil, err
		}
		solid.Color = color
	}

	return solid, nil
}

// parseBlock reads a block insert (CDataBlock).
func parseBlock(jr *Reader, version uint32) (*Block, error) {
	base, err := parseEntityBase(jr, version)
	if err != nil {
		return nil, err
	}

	block := &Block{EntityBase: *base}

	// Reference point
	block.RefX, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}
	block.RefY, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Scale X
	block.ScaleX, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Scale Y
	block.ScaleY, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Rotation angle (radians)
	block.Rotation, err = jr.ReadDouble()
	if err != nil {
		return nil, err
	}

	// Block definition number
	defNum, err := jr.ReadDWORD()
	if err != nil {
		return nil, err
	}
	block.DefNumber = defNum

	return block, nil
}
