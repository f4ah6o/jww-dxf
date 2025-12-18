package dxf

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

// Writer serializes DXF documents to an io.Writer in ASCII DXF format.
// The writer manages handle generation for entities and writes properly
// formatted DXF group codes.
type Writer struct {
	w          io.Writer
	nextHandle int
}

// NewWriter creates a new DXF writer that outputs to the provided io.Writer.
// The writer starts with handle counter at 1 and will auto-increment for each
// entity requiring a unique handle.
func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w, nextHandle: 1}
}

// getHandle returns the next available handle as a hexadecimal string.
// Handles are unique identifiers for DXF objects and are auto-incremented.
func (w *Writer) getHandle() string {
	h := fmt.Sprintf("%X", w.nextHandle)
	w.nextHandle++
	return h
}

// EscapeUnicode converts non-ASCII characters to DXF Unicode escape format.
//
// DXF uses the escape sequence \U+XXXX for Unicode characters, where XXXX is
// the hexadecimal Unicode code point. This function converts any non-ASCII
// or non-printable characters to this format.
//
// Example: "日本語" -> "\U+65E5\U+672C\U+8A9E"
func EscapeUnicode(s string) string {
	var sb strings.Builder
	for _, r := range s {
		if r > 127 || !unicode.IsPrint(r) {
			// DXF uses \U+XXXX format for Unicode
			sb.WriteString(fmt.Sprintf("\\U+%04X", r))
		} else {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// WriteDocument writes a complete DXF document to the output stream.
//
// The DXF file structure consists of the following sections in order:
//  1. HEADER section - document settings and variables
//  2. TABLES section - layer, linetype, and text style definitions
//  3. BLOCKS section - block definitions
//  4. ENTITIES section - drawing entities
//  5. EOF marker
//
// This method orchestrates writing all sections in the correct order
// and with proper DXF formatting.
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
	// Header section with essential variables for ODA compatibility
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

	// Code page
	if err := w.writeGroupCode(9, "$DWGCODEPAGE"); err != nil {
		return err
	}
	if err := w.writeGroupCode(3, "ANSI_1252"); err != nil {
		return err
	}

	// Measurement units (metric)
	if err := w.writeGroupCode(9, "$MEASUREMENT"); err != nil {
		return err
	}
	if err := w.writeGroupCode(70, 1); err != nil {
		return err
	}

	// Text style
	if err := w.writeGroupCode(9, "$TEXTSTYLE"); err != nil {
		return err
	}
	if err := w.writeGroupCode(7, "Standard"); err != nil {
		return err
	}

	// Current layer
	if err := w.writeGroupCode(9, "$CLAYER"); err != nil {
		return err
	}
	if err := w.writeGroupCode(8, "0"); err != nil {
		return err
	}

	// Current linetype
	if err := w.writeGroupCode(9, "$CELTYPE"); err != nil {
		return err
	}
	if err := w.writeGroupCode(6, "BYLAYER"); err != nil {
		return err
	}

	// Current color
	if err := w.writeGroupCode(9, "$CECOLOR"); err != nil {
		return err
	}
	if err := w.writeGroupCode(62, 256); err != nil { // BYLAYER
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
	if err := w.writeGroupCode(5, w.getHandle()); err != nil {
		return err
	}
	if err := w.writeGroupCode(70, 3); err != nil { // 3 linetypes: BYLAYER, BYBLOCK, CONTINUOUS
		return err
	}

	// BYLAYER linetype (required)
	if err := w.writeGroupCode(0, "LTYPE"); err != nil {
		return err
	}
	if err := w.writeGroupCode(5, w.getHandle()); err != nil {
		return err
	}
	if err := w.writeGroupCode(2, "BYLAYER"); err != nil {
		return err
	}
	if err := w.writeGroupCode(70, 0); err != nil {
		return err
	}
	if err := w.writeGroupCode(3, ""); err != nil {
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

	// BYBLOCK linetype (required)
	if err := w.writeGroupCode(0, "LTYPE"); err != nil {
		return err
	}
	if err := w.writeGroupCode(5, w.getHandle()); err != nil {
		return err
	}
	if err := w.writeGroupCode(2, "BYBLOCK"); err != nil {
		return err
	}
	if err := w.writeGroupCode(70, 0); err != nil {
		return err
	}
	if err := w.writeGroupCode(3, ""); err != nil {
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

	// CONTINUOUS linetype
	if err := w.writeGroupCode(0, "LTYPE"); err != nil {
		return err
	}
	if err := w.writeGroupCode(5, w.getHandle()); err != nil {
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
	if err := w.writeGroupCode(5, w.getHandle()); err != nil {
		return err
	}
	if err := w.writeGroupCode(70, len(doc.Layers)+1); err != nil { // +1 for required layer 0
		return err
	}

	// Required Layer 0 (must be first and always present)
	if err := w.writeGroupCode(0, "LAYER"); err != nil {
		return err
	}
	if err := w.writeGroupCode(5, w.getHandle()); err != nil {
		return err
	}
	if err := w.writeGroupCode(2, "0"); err != nil {
		return err
	}
	if err := w.writeGroupCode(70, 0); err != nil {
		return err
	}
	if err := w.writeGroupCode(62, 7); err != nil { // white/black
		return err
	}
	if err := w.writeGroupCode(6, "CONTINUOUS"); err != nil {
		return err
	}

	for _, layer := range doc.Layers {
		if err := w.writeGroupCode(0, "LAYER"); err != nil {
			return err
		}
		if err := w.writeGroupCode(5, w.getHandle()); err != nil {
			return err
		}
		if err := w.writeGroupCode(2, EscapeUnicode(layer.Name)); err != nil {
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
	if err := w.writeGroupCode(5, w.getHandle()); err != nil {
		return err
	}
	if err := w.writeGroupCode(70, 1); err != nil {
		return err
	}

	// STANDARD style
	if err := w.writeGroupCode(0, "STYLE"); err != nil {
		return err
	}
	if err := w.writeGroupCode(5, w.getHandle()); err != nil {
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

// writeGroupCode writes a single DXF group code/value pair.
//
// DXF files are structured as pairs of:
//   - Group code (integer, right-aligned in 3 characters)
//   - Value (string, int, or float64, on the next line)
//
// The group code indicates the type of data (e.g., 0=entity type, 8=layer, 10=X coordinate).
// This method formats the pair according to DXF specifications.
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

// ToString serializes a DXF Document to a string in ASCII DXF format.
// This is a convenience function that creates a Writer with a strings.Builder
// and returns the complete DXF file as a string.
//
// Example:
//
//	dxfContent := dxf.ToString(doc)
//	os.WriteFile("output.dxf", []byte(dxfContent), 0644)
func ToString(doc *Document) string {
	var sb strings.Builder
	w := NewWriter(&sb)
	_ = w.WriteDocument(doc)
	return sb.String()
}
