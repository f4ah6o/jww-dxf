package dxf

import (
	"fmt"
	"math"

	"github.com/f4ah6o/jww-dxf/jww"
)

// ConvertDocument converts a JWW document to a DXF document.
func ConvertDocument(doc *jww.Document) *Document {
	dxfDoc := &Document{
		Layers:   convertLayers(doc),
		Entities: convertEntities(doc),
		Blocks:   convertBlocks(doc),
	}
	return dxfDoc
}

// convertLayers creates DXF layers from JWW layer groups.
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

// convertEntities converts JWW entities to DXF entities.
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

// convertEntity converts a single JWW entity to a DXF entity.
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
			majorRadius := v.Radius
			minorRadius := v.Radius * v.Flatness

			// Major axis endpoint relative to center
			majorAxisX := majorRadius * math.Cos(v.TiltAngle)
			majorAxisY := majorRadius * math.Sin(v.TiltAngle)

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
				MinorRatio: minorRadius / majorRadius,
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

// getLayerName returns the layer name for a given layer group and layer.
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

// getBlockName returns the block name for a given definition number.
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

// mapColor maps JWW color to DXF ACI color.
func mapColor(jwwColor uint16) int {
	// JWW uses 1-9 for colors, 0 is background
	// DXF ACI: 1=red, 2=yellow, 3=green, 4=cyan, 5=blue, 6=magenta, 7=white
	if jwwColor == 0 {
		return 0 // BYLAYER
	}
	if jwwColor <= 9 {
		return int(jwwColor)
	}
	// Extended colors (SXF): offset by 100
	if jwwColor >= 100 {
		return int(jwwColor - 100 + 10)
	}
	return int(jwwColor)
}

// radToDeg converts radians to degrees.
func radToDeg(rad float64) float64 {
	return rad * 180.0 / math.Pi
}
