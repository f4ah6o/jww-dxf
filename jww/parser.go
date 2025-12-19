package jww

import (
	"bytes"
	"fmt"
	"io"
)

// Parse reads a JWW (Jw_cad) file from the provided reader and returns a parsed Document.
//
// The function reads the entire file into memory, validates the JWW signature,
// and parses the binary structure according to the MFC CArchive serialization format.
// It extracts layer information, drawing entities, and block definitions.
//
// The JWW file format uses:
//   - Little-endian byte order
//   - Shift-JIS text encoding (converted to UTF-8)
//   - MFC CArchive serialization with PID tracking
//
// Returns an error if:
//   - The file cannot be read
//   - The file signature is invalid (not "JwwData.")
//   - The file structure is corrupted or unsupported
//
// Example:
//
//	f, err := os.Open("drawing.jww")
//	if err != nil {
//	    return err
//	}
//	defer f.Close()
//
//	doc, err := jww.Parse(f)
//	if err != nil {
//	    return fmt.Errorf("parsing JWW file: %w", err)
//	}
//
//	fmt.Printf("Version: %d, Entities: %d\n", doc.Version, len(doc.Entities))
func Parse(r io.Reader) (*Document, error) {
	// Read entire file into memory for simpler parsing
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	// Validate signature
	if len(data) < 8 || string(data[:8]) != "JwwData." {
		return nil, ErrInvalidSignature
	}

	jr := NewReader(bytes.NewReader(data))

	// Skip signature
	jr.Skip(8)

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

		state, _ := jr.ReadDWORD()
		lg.State = state

		writeLay, _ := jr.ReadDWORD()
		lg.WriteLayer = writeLay

		scale, _ := jr.ReadDouble()
		lg.Scale = scale

		protect, _ := jr.ReadDWORD()
		lg.Protect = protect

		for lay := 0; lay < 16; lay++ {
			layState, _ := jr.ReadDWORD()
			lg.Layers[lay].State = layState

			layProtect, _ := jr.ReadDWORD()
			lg.Layers[lay].Protect = layProtect
		}
	}

	// Find entity list start by scanning for the first CData class pattern
	// Pattern: [count DWORD] [0xFF 0xFF] [schema WORD] [name_len WORD] ["CData..."]
	entityListOffset := findEntityListOffset(data, version)
	if entityListOffset < 0 {
		return nil, fmt.Errorf("could not find entity list in file")
	}

	// Parse entities from found offset
	jr2 := NewReader(bytes.NewReader(data[entityListOffset:]))
	entities, bytesRead, err := parseEntityListWithOffset(jr2, version)
	if err != nil {
		return nil, fmt.Errorf("parsing entity list: %w", err)
	}
	doc.Entities = entities

	// Parse block definitions (immediately after entity list)
	jr3 := NewReader(bytes.NewReader(data[entityListOffset+bytesRead:]))
	blockDefs, err := parseBlockDefList(jr3, version)
	if err != nil {
		// Block definitions might not exist in all files, just continue
		blockDefs = nil
	}
	doc.BlockDefs = blockDefs

	// Parse layer names from earlier in the file
	parseLayerNames(data, doc)

	return doc, nil
}

// findEntityListOffset scans the file for the entity list start position.
// The entity list is preceded by [count DWORD] and starts with a class definition.
func findEntityListOffset(data []byte, version uint32) int {
	// Look for the pattern: DWORD count followed by 0xFF 0xFF (new class marker)
	// followed by version schema and "CData" class name

	schemaBytes := []byte{byte(version & 0xFF), byte((version >> 8) & 0xFF)}

	for i := 100; i < len(data)-20; i++ {
		// Check for 0xFF 0xFF (new class marker)
		if data[i] == 0xFF && data[i+1] == 0xFF {
			// Check schema version matches
			if data[i+2] == schemaBytes[0] && data[i+3] == schemaBytes[1] {
				// Check if class name starts with "CData"
				nameLen := int(data[i+4]) + int(data[i+5])*256
				if nameLen >= 8 && nameLen <= 20 && i+6+nameLen <= len(data) {
					className := string(data[i+6 : i+6+nameLen])
					if len(className) >= 5 && className[:5] == "CData" {
						// Found first entity class definition
						// The count WORD is right before this (2 bytes)
						return i - 2
					}
				}
			}
		}
	}

	return -1
}

// parseEntityListWithOffset parses the entity list and returns bytes consumed.
func parseEntityListWithOffset(jr *Reader, version uint32) ([]Entity, int, error) {
	startBytes := jr.BytesRead()

	countWord, err := jr.ReadWORD()
	if err != nil {
		return nil, 0, fmt.Errorf("reading entity count: %w", err)
	}
	count := uint32(countWord)

	entities := make([]Entity, 0, count)

	// MFC CArchive PID tracking:
	// - Each new class definition gets a PID
	// - Each object also gets a PID
	// - PIDs are assigned sequentially starting from 1
	// - Class references use 0x8000 | class_PID
	pidToClassName := make(map[uint32]string) // PID -> class name
	nextPID := uint32(1)

	for i := uint32(0); i < count; i++ {
		entity, newPID, err := parseEntityWithPIDTracking(jr, version, pidToClassName, nextPID)
		if err != nil {
			return entities, 0, fmt.Errorf("parsing entity %d/%d: %w", i+1, count, err)
		}
		nextPID = newPID
		if entity != nil {
			entities = append(entities, entity)
		}
	}

	bytesConsumed := jr.BytesRead() - startBytes
	return entities, int(bytesConsumed), nil
}

// parseEntityWithPIDTracking parses an entity using MFC CArchive PID tracking.
// In MFC serialization:
// - 0xFFFF = new class definition follows (schema + name), then assign PID to class
// - 0x8000 = null object
// - 0x8000 | n = reference to class with PID n
// After parsing each object, assign a new PID to that object too.
func parseEntityWithPIDTracking(jr *Reader, version uint32, pidToClassName map[uint32]string, nextPID uint32) (Entity, uint32, error) {
	classID, err := jr.ReadWORD()
	if err != nil {
		return nil, nextPID, err
	}

	var className string

	if classID == 0xFFFF {
		// New class definition
		schemaVer, err := jr.ReadWORD()
		if err != nil {
			return nil, nextPID, fmt.Errorf("reading schema version: %w", err)
		}
		_ = schemaVer

		nameLen, err := jr.ReadWORD()
		if err != nil {
			return nil, nextPID, fmt.Errorf("reading class name length: %w", err)
		}

		nameBuf := make([]byte, nameLen)
		if err := jr.ReadBytes(nameBuf); err != nil {
			return nil, nextPID, fmt.Errorf("reading class name: %w", err)
		}
		className = string(nameBuf)

		// Assign PID to this class definition
		pidToClassName[nextPID] = className
		nextPID++
	} else if classID == 0x8000 {
		// Null object
		return nil, nextPID, nil
	} else {
		// Class reference: 0x8000 | class_PID
		// The lower bits contain the PID of the class definition
		classPID := uint32(classID & 0x7FFF)
		var ok bool
		className, ok = pidToClassName[classPID]
		if !ok {
			return nil, nextPID, fmt.Errorf("unknown class PID: %d (have PIDs: %v)", classPID, getKeys(pidToClassName))
		}
	}

	// Parse the object based on class name
	var entity Entity
	switch className {
	case "CDataSen":
		entity, err = parseLine(jr, version)
	case "CDataEnko":
		entity, err = parseArc(jr, version)
	case "CDataTen":
		entity, err = parsePoint(jr, version)
	case "CDataMoji":
		entity, err = parseText(jr, version)
	case "CDataSolid":
		entity, err = parseSolid(jr, version)
	case "CDataBlock":
		entity, err = parseBlock(jr, version)
	case "CDataSunpou":
		entity, err = parseDimension(jr, version)
	default:
		return nil, nextPID, fmt.Errorf("unknown entity class: %s", className)
	}

	if err != nil {
		return nil, nextPID, err
	}

	// Assign PID to this object
	nextPID++

	return entity, nextPID, nil
}

// getKeys returns the keys of a map for debugging
func getKeys(m map[uint32]string) []uint32 {
	keys := make([]uint32, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// parseLayerNames extracts layer names from the file.
func parseLayerNames(data []byte, doc *Document) {
	// Layer names appear earlier in the file
	// For now, use default names if we can't find them
	for gLay := 0; gLay < 16; gLay++ {
		if doc.LayerGroups[gLay].Name == "" {
			doc.LayerGroups[gLay].Name = fmt.Sprintf("Group%X", gLay)
		}
		for lay := 0; lay < 16; lay++ {
			if doc.LayerGroups[gLay].Layers[lay].Name == "" {
				doc.LayerGroups[gLay].Layers[lay].Name = fmt.Sprintf("%X-%X", gLay, lay)
			}
		}
	}
}

// parseBlockDefList parses the block definition list
func parseBlockDefList(jr *Reader, version uint32) ([]BlockDef, error) {
	count, err := jr.ReadDWORD()
	if err != nil {
		return nil, fmt.Errorf("reading block def count: %w", err)
	}

	if count > 10000 {
		// Probably not a valid block count, skip
		return nil, nil
	}

	blockDefs := make([]BlockDef, 0, count)
	classMap := make(map[uint16]string)
	nextID := uint16(1)

	for i := uint32(0); i < count; i++ {
		bd, newID, err := parseBlockDefWithTracking(jr, version, classMap, nextID)
		if err != nil {
			return blockDefs, nil // Return what we have
		}
		nextID = newID
		if bd != nil {
			blockDefs = append(blockDefs, *bd)
		}
	}

	return blockDefs, nil
}

// parseBlockDefWithTracking parses a single block definition with class tracking.
func parseBlockDefWithTracking(jr *Reader, version uint32, classMap map[uint16]string, nextID uint16) (*BlockDef, uint16, error) {
	classID, err := jr.ReadWORD()
	if err != nil {
		return nil, nextID, err
	}

	if classID == 0xFFFF {
		_, _ = jr.ReadWORD() // schema
		nameLen, _ := jr.ReadWORD()
		nameBuf := make([]byte, nameLen)
		jr.ReadBytes(nameBuf)
		classMap[nextID] = string(nameBuf)
		nextID++
	} else if classID == 0x8000 {
		return nil, nextID, nil
	}

	base, err := parseEntityBase(jr, version)
	if err != nil {
		return nil, nextID, err
	}

	bd := &BlockDef{EntityBase: *base}

	bd.Number, _ = jr.ReadDWORD()

	ref, _ := jr.ReadDWORD()
	bd.IsReferenced = ref != 0

	jr.Skip(4) // CTime

	bd.Name, _ = jr.ReadCString()

	// Parse nested entities
	nestedEntities, _, err := parseEntityListWithOffset(jr, version)
	if err != nil {
		return bd, nextID, nil
	}
	bd.Entities = nestedEntities

	return bd, nextID, nil
}

// parseDimension parses a dimension entity from the JWW file (JWW class: CDataSunpou).
// Dimensions are complex entities composed of lines and text to show measurements.
// This function extracts the dimension data and returns the associated line entity.
// Version 4.20 and later include additional SXF mode data.
func parseDimension(jr *Reader, version uint32) (Entity, error) {
	base, err := parseEntityBase(jr, version)
	if err != nil {
		return nil, err
	}
	_ = base

	// Parse the line member
	line, err := parseLine(jr, version)
	if err != nil {
		return nil, err
	}

	// Parse the text member
	_, err = parseText(jr, version)
	if err != nil {
		return nil, err
	}

	// Ver.4.20+ has additional SXF mode data
	if version >= 420 {
		_, _ = jr.ReadWORD() // SXF mode

		for i := 0; i < 2; i++ {
			parseLine(jr, version)
		}
		for i := 0; i < 4; i++ {
			parsePoint(jr, version)
		}
	}

	return line, nil
}

// parseEntityBase reads the common entity base fields shared by all entity types.
// This function extracts attributes like layer, color, line style, and flags
// that are present at the beginning of every JWW entity structure.
//
// The structure varies slightly based on the file version:
//   - Ver.3.51+: includes PenWidth field
//   - Earlier versions: no PenWidth field
func parseEntityBase(jr *Reader, version uint32) (*EntityBase, error) {
	base := &EntityBase{}

	group, err := jr.ReadDWORD()
	if err != nil {
		return nil, err
	}
	base.Group = group

	penStyle, err := jr.ReadBYTE()
	if err != nil {
		return nil, err
	}
	base.PenStyle = penStyle

	penColor, err := jr.ReadWORD()
	if err != nil {
		return nil, err
	}
	base.PenColor = penColor

	if version >= 351 {
		penWidth, err := jr.ReadWORD()
		if err != nil {
			return nil, err
		}
		base.PenWidth = penWidth
	}

	layer, err := jr.ReadWORD()
	if err != nil {
		return nil, err
	}
	base.Layer = layer

	layerGroup, err := jr.ReadWORD()
	if err != nil {
		return nil, err
	}
	base.LayerGroup = layerGroup

	flag, err := jr.ReadWORD()
	if err != nil {
		return nil, err
	}
	base.Flag = flag

	return base, nil
}

// parseLine reads a line entity from the JWW file (JWW class: CDataSen).
// Lines are represented by start and end points in 2D coordinate space.
func parseLine(jr *Reader, version uint32) (*Line, error) {
	base, err := parseEntityBase(jr, version)
	if err != nil {
		return nil, err
	}

	line := &Line{EntityBase: *base}

	line.StartX, _ = jr.ReadDouble()
	line.StartY, _ = jr.ReadDouble()
	line.EndX, _ = jr.ReadDouble()
	line.EndY, _ = jr.ReadDouble()

	return line, nil
}

// parseArc reads an arc or circle entity from the JWW file (JWW class: CDataEnko).
// This entity type can represent circles, ellipses, arcs, or elliptical arcs
// based on the Flatness and IsFullCircle properties.
func parseArc(jr *Reader, version uint32) (*Arc, error) {
	base, err := parseEntityBase(jr, version)
	if err != nil {
		return nil, err
	}

	arc := &Arc{EntityBase: *base}

	arc.CenterX, _ = jr.ReadDouble()
	arc.CenterY, _ = jr.ReadDouble()
	arc.Radius, _ = jr.ReadDouble()
	arc.StartAngle, _ = jr.ReadDouble()
	arc.ArcAngle, _ = jr.ReadDouble()
	arc.TiltAngle, _ = jr.ReadDouble()
	arc.Flatness, _ = jr.ReadDouble()
	fullCircle, _ := jr.ReadDWORD()
	arc.IsFullCircle = fullCircle != 0

	return arc, nil
}

// parsePoint reads a point entity from the JWW file (JWW class: CDataTen).
// Points can be temporary construction points or permanent marker points with symbols.
func parsePoint(jr *Reader, version uint32) (*Point, error) {
	base, err := parseEntityBase(jr, version)
	if err != nil {
		return nil, err
	}

	pt := &Point{EntityBase: *base}

	pt.X, _ = jr.ReadDouble()
	pt.Y, _ = jr.ReadDouble()
	tmp, _ := jr.ReadDWORD()
	pt.IsTemporary = tmp != 0

	if base.PenStyle == 100 {
		pt.Code, _ = jr.ReadDWORD()
		pt.Angle, _ = jr.ReadDouble()
		pt.Scale, _ = jr.ReadDouble()
	}

	return pt, nil
}

// parseText reads a text entity from the JWW file (JWW class: CDataMoji).
// Text content is stored in Shift-JIS encoding and converted to UTF-8.
// Text can have various fonts, sizes, and styles including bold and italic.
func parseText(jr *Reader, version uint32) (*Text, error) {
	base, err := parseEntityBase(jr, version)
	if err != nil {
		return nil, err
	}

	txt := &Text{EntityBase: *base}

	txt.StartX, _ = jr.ReadDouble()
	txt.StartY, _ = jr.ReadDouble()
	txt.EndX, _ = jr.ReadDouble()
	txt.EndY, _ = jr.ReadDouble()
	txt.TextType, _ = jr.ReadDWORD()
	txt.SizeX, _ = jr.ReadDouble()
	txt.SizeY, _ = jr.ReadDouble()
	txt.Spacing, _ = jr.ReadDouble()
	txt.Angle, _ = jr.ReadDouble()
	txt.FontName, _ = jr.ReadCString()
	txt.Content, _ = jr.ReadCString()

	return txt, nil
}

// parseSolid reads a solid fill entity from the JWW file (JWW class: CDataSolid).
// Solids are quadrilaterals or triangles used for filled areas, hatching, and shading.
func parseSolid(jr *Reader, version uint32) (*Solid, error) {
	base, err := parseEntityBase(jr, version)
	if err != nil {
		return nil, err
	}

	solid := &Solid{EntityBase: *base}

	solid.Point1X, _ = jr.ReadDouble()
	solid.Point1Y, _ = jr.ReadDouble()
	solid.Point4X, _ = jr.ReadDouble()
	solid.Point4Y, _ = jr.ReadDouble()
	solid.Point2X, _ = jr.ReadDouble()
	solid.Point2Y, _ = jr.ReadDouble()
	solid.Point3X, _ = jr.ReadDouble()
	solid.Point3Y, _ = jr.ReadDouble()

	if base.PenColor == 10 {
		solid.Color, _ = jr.ReadDWORD()
	}

	return solid, nil
}

// parseBlock reads a block insert entity from the JWW file (JWW class: CDataBlock).
// Block inserts reference a block definition and can have independent scale and rotation.
func parseBlock(jr *Reader, version uint32) (*Block, error) {
	base, err := parseEntityBase(jr, version)
	if err != nil {
		return nil, err
	}

	block := &Block{EntityBase: *base}

	block.RefX, _ = jr.ReadDouble()
	block.RefY, _ = jr.ReadDouble()
	block.ScaleX, _ = jr.ReadDouble()
	block.ScaleY, _ = jr.ReadDouble()
	block.Rotation, _ = jr.ReadDouble()
	block.DefNumber, _ = jr.ReadDWORD()

	return block, nil
}
