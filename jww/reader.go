// Package jww provides types and parsing functions for Jw_cad (JWW) files.
//
// Jw_cad is a popular 2D CAD software in Japan that uses the JWW binary file format.
// This package handles the parsing of JWW files and conversion to Go data structures.
//
// The JWW file format characteristics:
//   - Binary format using MFC CArchive serialization
//   - Little-endian byte order
//   - Shift-JIS text encoding
//   - Supports layers, blocks, and various entity types (lines, arcs, text, etc.)
//
// Basic usage:
//
//	file, _ := os.Open("drawing.jww")
//	defer file.Close()
//
//	doc, err := jww.Parse(file)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, entity := range doc.Entities {
//	    fmt.Printf("Entity type: %s\n", entity.Type())
//	}
package jww

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"unsafe"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

var (
	// ErrInvalidSignature is returned when the file does not start with the JWW signature "JwwData.".
	ErrInvalidSignature = errors.New("invalid JWW signature: expected 'JwwData.'")

	// ErrUnsupportedVersion is returned when the JWW file version is not supported by this parser.
	ErrUnsupportedVersion = errors.New("unsupported JWW version")
)

// Reader wraps an io.Reader to provide convenient methods for reading JWW binary data.
// All multi-byte values are read in little-endian format, and text strings are
// decoded from Shift-JIS to UTF-8.
type Reader struct {
	r         io.Reader
	buf       []byte
	bytesRead int64
}

// NewReader creates a new JWW binary reader that wraps the provided io.Reader.
// The reader maintains an internal buffer for efficient binary data reading.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		r:         r,
		buf:       make([]byte, 8),
		bytesRead: 0,
	}
}

// ReadSignature reads and validates the JWW file signature.
// The signature must be the 8-byte string "JwwData.".
// Returns ErrInvalidSignature if the signature is invalid.
func (r *Reader) ReadSignature() error {
	sig := make([]byte, 8)
	n, err := io.ReadFull(r.r, sig)
	r.bytesRead += int64(n)
	if err != nil {
		return err
	}
	if string(sig) != "JwwData." {
		return ErrInvalidSignature
	}
	return nil
}

// ReadDWORD reads a 32-bit unsigned integer in little-endian format.
// This corresponds to the Windows DWORD type used in the JWW file format.
func (r *Reader) ReadDWORD() (uint32, error) {
	n, err := io.ReadFull(r.r, r.buf[:4])
	r.bytesRead += int64(n)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(r.buf[:4]), nil
}

// ReadWORD reads a 16-bit unsigned integer in little-endian format.
// This corresponds to the Windows WORD type used in the JWW file format.
func (r *Reader) ReadWORD() (uint16, error) {
	n, err := io.ReadFull(r.r, r.buf[:2])
	r.bytesRead += int64(n)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(r.buf[:2]), nil
}

// ReadBYTE reads a single unsigned byte.
// This corresponds to the Windows BYTE type used in the JWW file format.
func (r *Reader) ReadBYTE() (byte, error) {
	n, err := io.ReadFull(r.r, r.buf[:1])
	r.bytesRead += int64(n)
	if err != nil {
		return 0, err
	}
	return r.buf[0], nil
}

// ReadDouble reads a 64-bit IEEE 754 floating point number in little-endian format.
// This is used for coordinate values and other measurements in JWW files.
func (r *Reader) ReadDouble() (float64, error) {
	n, err := io.ReadFull(r.r, r.buf[:8])
	r.bytesRead += int64(n)
	if err != nil {
		return 0, err
	}
	bits := binary.LittleEndian.Uint64(r.buf[:8])
	return float64FromBits(bits), nil
}

// ReadCString reads a length-prefixed string in MFC CString format.
//
// The string format is:
//   - If length < 255: 1 byte length prefix
//   - If length < 65535: 1 byte 0xFF marker + 2 byte length
//   - Otherwise: 1 byte 0xFF marker + 2 byte 0xFFFF marker + 4 byte length
//
// The string data is encoded in Shift-JIS and automatically converted to UTF-8.
func (r *Reader) ReadCString() (string, error) {
	// Read length prefix
	lenByte, err := r.ReadBYTE()
	if err != nil {
		return "", err
	}

	var length uint32
	if lenByte < 0xFF {
		length = uint32(lenByte)
	} else {
		// Read 2-byte length
		lenWord, err := r.ReadWORD()
		if err != nil {
			return "", err
		}
		if lenWord < 0xFFFF {
			length = uint32(lenWord)
		} else {
			// Read 4-byte length
			length, err = r.ReadDWORD()
			if err != nil {
				return "", err
			}
		}
	}

	if length == 0 {
		return "", nil
	}

	// Read string bytes
	strBuf := make([]byte, length)
	n, err := io.ReadFull(r.r, strBuf)
	r.bytesRead += int64(n)
	if err != nil {
		return "", err
	}

	// Convert Shift-JIS to UTF-8
	return shiftJISToUTF8(strBuf), nil
}

// ReadBytes reads exactly len(buf) bytes into the provided buffer.
// Returns an error if fewer bytes are available.
func (r *Reader) ReadBytes(buf []byte) error {
	n, err := io.ReadFull(r.r, buf)
	r.bytesRead += int64(n)
	return err
}

// Skip skips n bytes in the input stream.
// This is useful for skipping over unknown or unneeded data structures.
func (r *Reader) Skip(n int) error {
	buf := make([]byte, n)
	read, err := io.ReadFull(r.r, buf)
	r.bytesRead += int64(read)
	return err
}

// BytesRead returns the total number of bytes read from the underlying reader.
// This is useful for tracking offsets while parsing nested structures.
func (r *Reader) BytesRead() int64 {
	return r.bytesRead
}

// float64FromBits converts a uint64 bit pattern to a float64 value.
// This uses unsafe pointer conversion to reinterpret the bits as a float64.
func float64FromBits(bits uint64) float64 {
	return *(*float64)(unsafe.Pointer(&bits))
}

// shiftJISToUTF8 converts Shift-JIS encoded bytes to a UTF-8 string.
// Shift-JIS is the legacy Japanese character encoding used by JWW files.
// Null bytes are trimmed from the result.
// If conversion fails, the raw bytes are returned as a fallback.
func shiftJISToUTF8(data []byte) string {
	decoder := japanese.ShiftJIS.NewDecoder()
	result, _, err := transform.Bytes(decoder, data)
	if err != nil {
		// Fallback to raw bytes if conversion fails
		return string(data)
	}
	// Remove null bytes from the result
	return string(bytes.TrimRight(result, "\x00"))
}
