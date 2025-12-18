package dxf

import (
	"fmt"
	"io"
	"strings"
)

// Writer serializes DXF documents to an io.Writer.
type Writer struct {
	w io.Writer
}

// NewWriter creates a new DXF writer.
func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

// WriteDocument writes a complete DXF document.
func (w *Writer) WriteDocument(doc *Document) error {
	// HEADER section
	if err := w.writeHeader(); err != nil {
		return err
	}

	// TABLES section
	if err := w.writeTables(doc); err != nil {
		return err
	}

	// BLOCKS section
	if err := w.writeBlocks(doc); err != nil {
		return err
	}

	// ENTITIES section
	if err := w.writeEntities(doc); err != nil {
		return err
	}

	// End of file
	if err := w.writeGroupCode(0, "EOF"); err != nil {
		return err
	}

	return nil
}

func (w *Writer) writeHeader() error {
	// Minimal header for AutoCAD compatibility
	if err := w.writeSection("HEADER"); err != nil {
		return err
	}

	// AutoCAD version variable
	if err := w.writeGroupCode(9, "$ACADVER"); err != nil {
		return err
	}
	if err := w.writeGroupCode(1, "AC1015"); err != nil { // AutoCAD 2000
		return err
	}

	// Measurement units (metric)
	if err := w.writeGroupCode(9, "$MEASUREMENT"); err != nil {
		return err
	}
	if err := w.writeGroupCode(70, 1); err != nil {
		return err
	}

	return w.writeEndSection()
}

func (w *Writer) writeTables(doc *Document) error {
	if err := w.writeSection("TABLES"); err != nil {
		return err
	}

	// LTYPE table
	if err := w.writeLinetypeTable(); err != nil {
		return err
	}

	// LAYER table
	if err := w.writeLayerTable(doc); err != nil {
		return err
	}

	// STYLE table (text styles)
	if err := w.writeStyleTable(); err != nil {
		return err
	}

	return w.writeEndSection()
}

func (w *Writer) writeLinetypeTable() error {
	if err := w.writeGroupCode(0, "TABLE"); err != nil {
		return err
	}
	if err := w.writeGroupCode(2, "LTYPE"); err != nil {
		return err
	}
	if err := w.writeGroupCode(70, 1); err != nil {
		return err
	}

	// CONTINUOUS linetype
	if err := w.writeGroupCode(0, "LTYPE"); err != nil {
		return err
	}
	if err := w.writeGroupCode(2, "CONTINUOUS"); err != nil {
		return err
	}
	if err := w.writeGroupCode(70, 0); err != nil {
		return err
	}
	if err := w.writeGroupCode(3, "Solid line"); err != nil {
		return err
	}
	if err := w.writeGroupCode(72, 65); err != nil {
		return err
	}
	if err := w.writeGroupCode(73, 0); err != nil {
		return err
	}
	if err := w.writeGroupCode(40, 0.0); err != nil {
		return err
	}

	return w.writeGroupCode(0, "ENDTAB")
}

func (w *Writer) writeLayerTable(doc *Document) error {
	if err := w.writeGroupCode(0, "TABLE"); err != nil {
		return err
	}
	if err := w.writeGroupCode(2, "LAYER"); err != nil {
		return err
	}
	if err := w.writeGroupCode(70, len(doc.Layers)); err != nil {
		return err
	}

	for _, layer := range doc.Layers {
		if err := w.writeGroupCode(0, "LAYER"); err != nil {
			return err
		}
		if err := w.writeGroupCode(2, layer.Name); err != nil {
			return err
		}
		flags := 0
		if layer.Frozen {
			flags |= 1
		}
		if layer.Locked {
			flags |= 4
		}
		if err := w.writeGroupCode(70, flags); err != nil {
			return err
		}
		if err := w.writeGroupCode(62, layer.Color); err != nil {
			return err
		}
		if err := w.writeGroupCode(6, layer.LineType); err != nil {
			return err
		}
	}

	return w.writeGroupCode(0, "ENDTAB")
}

func (w *Writer) writeStyleTable() error {
	if err := w.writeGroupCode(0, "TABLE"); err != nil {
		return err
	}
	if err := w.writeGroupCode(2, "STYLE"); err != nil {
		return err
	}
	if err := w.writeGroupCode(70, 1); err != nil {
		return err
	}

	// STANDARD style
	if err := w.writeGroupCode(0, "STYLE"); err != nil {
		return err
	}
	if err := w.writeGroupCode(2, "STANDARD"); err != nil {
		return err
	}
	if err := w.writeGroupCode(70, 0); err != nil {
		return err
	}
	if err := w.writeGroupCode(40, 0.0); err != nil {
		return err
	}
	if err := w.writeGroupCode(41, 1.0); err != nil {
		return err
	}
	if err := w.writeGroupCode(50, 0.0); err != nil {
		return err
	}
	if err := w.writeGroupCode(71, 0); err != nil {
		return err
	}
	if err := w.writeGroupCode(42, 2.5); err != nil {
		return err
	}
	if err := w.writeGroupCode(3, "txt"); err != nil {
		return err
	}
	if err := w.writeGroupCode(4, ""); err != nil {
		return err
	}

	return w.writeGroupCode(0, "ENDTAB")
}

func (w *Writer) writeBlocks(doc *Document) error {
	if err := w.writeSection("BLOCKS"); err != nil {
		return err
	}

	for _, block := range doc.Blocks {
		// Block header
		if err := w.writeGroupCode(0, "BLOCK"); err != nil {
			return err
		}
		if err := w.writeGroupCode(8, "0"); err != nil {
			return err
		}
		if err := w.writeGroupCode(2, block.Name); err != nil {
			return err
		}
		if err := w.writeGroupCode(70, 0); err != nil {
			return err
		}
		if err := w.writeGroupCode(10, block.BaseX); err != nil {
			return err
		}
		if err := w.writeGroupCode(20, block.BaseY); err != nil {
			return err
		}
		if err := w.writeGroupCode(30, 0.0); err != nil {
			return err
		}
		if err := w.writeGroupCode(3, block.Name); err != nil {
			return err
		}

		// Block entities
		for _, entity := range block.Entities {
			if err := w.writeEntity(entity); err != nil {
				return err
			}
		}

		// Block end
		if err := w.writeGroupCode(0, "ENDBLK"); err != nil {
			return err
		}
		if err := w.writeGroupCode(8, "0"); err != nil {
			return err
		}
	}

	return w.writeEndSection()
}

func (w *Writer) writeEntities(doc *Document) error {
	if err := w.writeSection("ENTITIES"); err != nil {
		return err
	}

	for _, entity := range doc.Entities {
		if err := w.writeEntity(entity); err != nil {
			return err
		}
	}

	return w.writeEndSection()
}

func (w *Writer) writeEntity(entity Entity) error {
	for _, gc := range entity.GroupCodes() {
		if err := w.writeGroupCode(gc.Code, gc.Value); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) writeSection(name string) error {
	if err := w.writeGroupCode(0, "SECTION"); err != nil {
		return err
	}
	return w.writeGroupCode(2, name)
}

func (w *Writer) writeEndSection() error {
	return w.writeGroupCode(0, "ENDSEC")
}

func (w *Writer) writeGroupCode(code int, value interface{}) error {
	var line string
	switch v := value.(type) {
	case string:
		line = fmt.Sprintf("%3d\n%s\n", code, v)
	case int:
		line = fmt.Sprintf("%3d\n%d\n", code, v)
	case float64:
		line = fmt.Sprintf("%3d\n%f\n", code, v)
	default:
		line = fmt.Sprintf("%3d\n%v\n", code, v)
	}
	_, err := io.WriteString(w.w, line)
	return err
}

// ToString serializes a Document to a DXF string.
func ToString(doc *Document) string {
	var sb strings.Builder
	w := NewWriter(&sb)
	_ = w.WriteDocument(doc)
	return sb.String()
}
