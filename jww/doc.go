// Package jww parses Jw_cad (JWW) drawings into Go structures that expose
// version metadata, layer information, entities, and block definitions.
//
// The package reads the binary JWW format using the same PID-tracking
// serialization as MFC's CArchive and converts Shift-JIS encoded strings to
// UTF-8. Parsed documents can then be inspected directly or transformed into
// DXF entities via the companion dxf package.
package jww
