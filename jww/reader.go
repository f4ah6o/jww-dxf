// Package jww provides types and parsing functions for Jw_cad (JWW) files.
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
	ErrInvalidSignature   = errors.New("invalid JWW signature: expected 'JwwData.'")
	ErrUnsupportedVersion = errors.New("unsupported JWW version")
)

// Reader wraps an io.Reader to read JWW binary data in little-endian format.
type Reader struct {
	r   io.Reader
	buf []byte
}

// NewReader creates a new JWW binary reader.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		r:   r,
		buf: make([]byte, 8),
	}
}

// ReadSignature reads and validates the JWW file signature "JwwData.".
func (r *Reader) ReadSignature() error {
	sig := make([]byte, 8)
	if _, err := io.ReadFull(r.r, sig); err != nil {
		return err
	}
	if string(sig) != "JwwData." {
		return ErrInvalidSignature
	}
	return nil
}

// ReadDWORD reads a 32-bit unsigned integer (little-endian).
func (r *Reader) ReadDWORD() (uint32, error) {
	if _, err := io.ReadFull(r.r, r.buf[:4]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(r.buf[:4]), nil
}

// ReadWORD reads a 16-bit unsigned integer (little-endian).
func (r *Reader) ReadWORD() (uint16, error) {
	if _, err := io.ReadFull(r.r, r.buf[:2]); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(r.buf[:2]), nil
}

// ReadBYTE reads a single byte.
func (r *Reader) ReadBYTE() (byte, error) {
	if _, err := io.ReadFull(r.r, r.buf[:1]); err != nil {
		return 0, err
	}
	return r.buf[0], nil
}

// ReadDouble reads a 64-bit floating point number (little-endian).
func (r *Reader) ReadDouble() (float64, error) {
	if _, err := io.ReadFull(r.r, r.buf[:8]); err != nil {
		return 0, err
	}
	bits := binary.LittleEndian.Uint64(r.buf[:8])
	return float64FromBits(bits), nil
}

// ReadCString reads a length-prefixed string (MFC CString format).
// Format: 1 byte length if < 255, otherwise 2 bytes if < 65535, otherwise 4 bytes.
// Converts Shift-JIS to UTF-8.
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
	if _, err := io.ReadFull(r.r, strBuf); err != nil {
		return "", err
	}

	// Convert Shift-JIS to UTF-8
	return shiftJISToUTF8(strBuf), nil
}

// ReadBytes reads n bytes into the provided buffer.
func (r *Reader) ReadBytes(buf []byte) error {
	_, err := io.ReadFull(r.r, buf)
	return err
}

// Skip skips n bytes.
func (r *Reader) Skip(n int) error {
	buf := make([]byte, n)
	_, err := io.ReadFull(r.r, buf)
	return err
}

// float64FromBits converts uint64 bits to float64.
func float64FromBits(bits uint64) float64 {
	return *(*float64)(unsafe.Pointer(&bits))
}

// shiftJISToUTF8 converts Shift-JIS encoded bytes to UTF-8 string.
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
