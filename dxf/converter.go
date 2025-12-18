package dxf

import (
	"fmt"
	"math"

	"github.com/f4ah6o/jww-dxf/jww"
)

// ConvertDocument converts a JWW (Jw_cad) document to a DXF document.
//
// This function transforms JWW entities into their DXF equivalents:
//   - JWW layers are converted to DXF layers with appropriate mapping
//   - JWW entities (Line, Arc, Point, Text, Solid, Block) are converted to DXF entities
//   - JWW block definitions are converted to DXF blocks
//
// The conversion handles:
//   - Layer group and layer hierarchy mapping
//   - Color index mapping
//   - Coordinate system preservation
//   - Arc and ellipse geometry conversion
//   - Text encoding (Shift-JIS to Unicode)
//
// Returns a DXF Document ready to be written to a file.
func ConvertDocument(doc *jww.Document) *Document {
	dxfDoc := &Document{
		Layers:   convertLayers(doc),
		Entities: convertEntities(doc),
		Blocks:   convertBlocks(doc),
	}
	return dxfDoc
}

// convertLayers creates DXF layers from JWW layer groups.
// JWW has 16 layer groups with 16 layers each (256 total layers).
// Each JWW layer is converted to a single DXF layer with a name like "0-0" or "F-A".
// Layer properties (frozen, locked) are preserved in the conversion.
func convertLayers(doc *jww.Document) []Layer {
	var layers []Layer

	for gLay := 0; gLay < 16; gLay++ {
		lg := &doc.LayerGroups[gLay]
		for lay := 0; lay < 16; lay++ {
			l := &lg.Layers[lay]
			name := l.Name
			if name == "" {
				name = fmt.Sprintf("%X-%X", gLay, lay)
			}

			layers = append(layers, Layer{
				Name:     name,
				Color:    (gLay*16+lay)%255 + 1, // Simple ACI color mapping
				LineType: "CONTINUOUS",
				Frozen:   l.State == 0,
				Locked:   l.Protect != 0,
			})
		}
	}

	return layers
}

// convertEntities converts all JWW entities to DXF entities.
// This function iterates through all entities in the JWW document and
// converts each one based on its type. Unsupported or invalid entities
// are skipped.
func convertEntities(doc *jww.Document) []Entity {
	var entities []Entity

	for _, e := range doc.Entities {
		dxfEntity := convertEntity(e, doc)
		if dxfEntity != nil {
			entities = append(entities, dxfEntity)
		}
	}

	return entities
}

// convertEntity converts a single JWW entity to its DXF equivalent.
//
// Supported conversions:
//   - jww.Line -> dxf.Line
//   - jww.Arc -> dxf.Circle (for full circles) or dxf.Arc (for arcs) or dxf.Ellipse (for ellipses)
//   - jww.Point -> dxf.Point (temporary points are skipped)
//   - jww.Text -> dxf.Text (with Unicode escape conversion)
//   - jww.Solid -> dxf.Solid
//   - jww.Block -> dxf.Insert
//
// Returns nil for unsupported entity types or entities that should be skipped.
func convertEntity(e jww.Entity, doc *jww.Document) Entity {
	base := e.Base()
	layerName := getLayerName(doc, base.LayerGroup, base.Layer)
	color := mapColor(base.PenColor)

	switch v := e.(type) {
	case *jww.Line:
		return &Line{
			Layer: layerName,
			Color: color,
			X1:    v.StartX,
			Y1:    v.StartY,
			X2:    v.EndX,
			Y2:    v.EndY,
		}

	case *jww.Arc:
		if v.IsFullCircle && v.Flatness == 1.0 {
			// Full circle
			return &Circle{
				Layer:   layerName,
				Color:   color,
				CenterX: v.CenterX,
				CenterY: v.CenterY,
				Radius:  v.Radius,
			}
		} else if v.Flatness != 1.0 {
			// Ellipse or elliptical arc
			// DXF requires MinorRatio <= 1.0
			// If Flatness > 1.0, we need to swap major and minor axes
			majorRadius := v.Radius
			minorRatio := v.Flatness
			tiltAngle := v.TiltAngle

			if minorRatio > 1.0 {
				// Swap axes: minor becomes major, rotate by 90°
				majorRadius = v.Radius * v.Flatness
				minorRatio = 1.0 / v.Flatness
				tiltAngle = v.TiltAngle + math.Pi/2
			}

			// Major axis endpoint relative to center
			majorAxisX := majorRadius * math.Cos(tiltAngle)
			majorAxisY := majorRadius * math.Sin(tiltAngle)

			startParam := v.StartAngle
			endParam := v.StartAngle + v.ArcAngle
			if v.IsFullCircle {
				startParam = 0
				endParam = 2 * math.Pi
			}

			return &Ellipse{
				Layer:      layerName,
				Color:      color,
				CenterX:    v.CenterX,
				CenterY:    v.CenterY,
				MajorAxisX: majorAxisX,
				MajorAxisY: majorAxisY,
				MinorRatio: minorRatio,
				StartParam: startParam,
				EndParam:   endParam,
			}
		} else {
			// Arc
			startAngle := radToDeg(v.StartAngle)
			endAngle := radToDeg(v.StartAngle + v.ArcAngle)

			return &Arc{
				Layer:      layerName,
				Color:      color,
				CenterX:    v.CenterX,
				CenterY:    v.CenterY,
				Radius:     v.Radius,
				StartAngle: startAngle,
				EndAngle:   endAngle,
			}
		}

	case *jww.Point:
		if v.IsTemporary {
			return nil // Skip temporary points
		}
		return &Point{
			Layer: layerName,
			Color: color,
			X:     v.X,
			Y:     v.Y,
		}

	case *jww.Text:
		return &Text{
			Layer:    layerName,
			Color:    color,
			X:        v.StartX,
			Y:        v.StartY,
			Height:   v.SizeY,
			Rotation: v.Angle,
			Content:  v.Content,
			Style:    "STANDARD",
		}

	case *jww.Solid:
		return &Solid{
			Layer: layerName,
			Color: color,
			X1:    v.Point1X,
			Y1:    v.Point1Y,
			X2:    v.Point2X,
			Y2:    v.Point2Y,
			X3:    v.Point3X,
			Y3:    v.Point3Y,
			X4:    v.Point4X,
			Y4:    v.Point4Y,
		}

	case *jww.Block:
		blockName := getBlockName(doc, v.DefNumber)
		return &Insert{
			Layer:     layerName,
			Color:     color,
			BlockName: blockName,
			X:         v.RefX,
			Y:         v.RefY,
			ScaleX:    v.ScaleX,
			ScaleY:    v.ScaleY,
			Rotation:  radToDeg(v.Rotation),
		}
	}

	return nil
}

// convertBlocks converts JWW block definitions to DXF blocks.
// Each JWW block definition is converted to a DXF block with all its
// entities converted to DXF equivalents.
func convertBlocks(doc *jww.Document) []Block {
	var blocks []Block

	for _, bd := range doc.BlockDefs {
		block := Block{
			Name:  bd.Name,
			BaseX: 0,
			BaseY: 0,
		}

		for _, e := range bd.Entities {
			dxfEntity := convertEntity(e, doc)
			if dxfEntity != nil {
				block.Entities = append(block.Entities, dxfEntity)
			}
		}

		blocks = append(blocks, block)
	}

	return blocks
}

// getLayerName returns the DXF layer name for a given JWW layer group and layer.
// If the layer has a custom name, it is used. Otherwise, a default name
// in the format "G-L" (e.g., "0-0", "F-A") is generated using hexadecimal notation.
func getLayerName(doc *jww.Document, layerGroup, layer uint16) string {
	if int(layerGroup) < 16 && int(layer) < 16 {
		lg := &doc.LayerGroups[layerGroup]
		l := &lg.Layers[layer]
		if l.Name != "" {
			return l.Name
		}
	}
	return fmt.Sprintf("%X-%X", layerGroup, layer)
}

// getBlockName returns the block name for a given JWW block definition number.
// If the block has a custom name, it is used. Otherwise, a default name
// like "BLOCK_1" is generated.
func getBlockName(doc *jww.Document, defNumber uint32) string {
	for _, bd := range doc.BlockDefs {
		if bd.Number == defNumber {
			if bd.Name != "" {
				return bd.Name
			}
			break
		}
	}
	return fmt.Sprintf("BLOCK_%d", defNumber)
}

// mapColor maps JWW color codes to DXF ACI (AutoCAD Color Index) values.
//
// JWW color mapping (standard Jw_cad colors):
//   - 0: background color -> 0 (BYLAYER in DXF)
//   - 1: 水色 (cyan) -> 4 (cyan)
//   - 2: 白 (white) -> 7 (white)
//   - 3: 緑 (green) -> 3 (green)
//   - 4: 黄色 (yellow) -> 2 (yellow)
//   - 5: ピンク (magenta) -> 6 (magenta)
//   - 6: 青 (blue) -> 5 (blue)
//   - 7: 黒/白 (foreground) -> 7 (white/black)
//   - 8: 赤 (red) -> 1 (red)
//   - 9: グレー (gray) -> 8 (dark gray)
//   - 100+: extended SXF colors -> mapped to DXF colors 10+
//
// DXF ACI color reference:
//   - 0: BYLAYER (inherits layer color)
//   - 1: red, 2: yellow, 3: green, 4: cyan, 5: blue, 6: magenta, 7: white/black
//   - 8-255: additional colors
func mapColor(jwwColor uint16) int {
	// JWW uses different color assignments than DXF ACI
	switch jwwColor {
	case 0:
		return 0 // BYLAYER
	case 1:
		return 4 // JWW 水色 (cyan) -> DXF cyan
	case 2:
		return 7 // JWW 白 (white) -> DXF white
	case 3:
		return 3 // JWW 緑 (green) -> DXF green
	case 4:
		return 2 // JWW 黄色 (yellow) -> DXF yellow
	case 5:
		return 6 // JWW ピンク (magenta) -> DXF magenta
	case 6:
		return 5 // JWW 青 (blue) -> DXF blue
	case 7:
		return 7 // JWW 黒/白 (foreground) -> DXF white/black
	case 8:
		return 1 // JWW 赤 (red) -> DXF red
	case 9:
		return 8 // JWW グレー (gray) -> DXF gray
	default:
		// Extended colors (SXF): 100+ -> map to DXF 10+
		if jwwColor >= 100 {
			return int(jwwColor - 100 + 10)
		}
		return int(jwwColor)
	}
}

// radToDeg converts an angle from radians to degrees.
// This is used for converting JWW angle values (in radians) to DXF angle values (in degrees).
func radToDeg(rad float64) float64 {
	return rad * 180.0 / math.Pi
}
