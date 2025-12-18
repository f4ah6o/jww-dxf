package jww

import (
	"bytes"
	"math"
	"testing"
)

func TestReader_ReadDWORD(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint32
	}{
		{"zero", []byte{0, 0, 0, 0}, 0},
		{"one", []byte{1, 0, 0, 0}, 1},
		{"max byte", []byte{255, 0, 0, 0}, 255},
		{"version 600", []byte{88, 2, 0, 0}, 600},
		{"version 700", []byte{188, 2, 0, 0}, 700},
		{"max uint32", []byte{255, 255, 255, 255}, 0xFFFFFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(bytes.NewReader(tt.data))
			val, err := r.ReadDWORD()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if val != tt.expected {
				t.Errorf("got %d, want %d", val, tt.expected)
			}
		})
	}
}

func TestReader_ReadWORD(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint16
	}{
		{"zero", []byte{0, 0}, 0},
		{"one", []byte{1, 0}, 1},
		{"max byte", []byte{255, 0}, 255},
		{"max uint16", []byte{255, 255}, 0xFFFF},
		{"class ID marker", []byte{255, 255}, 0xFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(bytes.NewReader(tt.data))
			val, err := r.ReadWORD()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if val != tt.expected {
				t.Errorf("got %d, want %d", val, tt.expected)
			}
		})
	}
}

func TestReader_ReadBYTE(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected byte
	}{
		{"zero", []byte{0}, 0},
		{"one", []byte{1}, 1},
		{"max", []byte{255}, 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(bytes.NewReader(tt.data))
			val, err := r.ReadBYTE()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if val != tt.expected {
				t.Errorf("got %d, want %d", val, tt.expected)
			}
		})
	}
}

func TestReader_ReadDouble(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected float64
	}{
		{
			"zero",
			[]byte{0, 0, 0, 0, 0, 0, 0, 0},
			0.0,
		},
		{
			"one",
			[]byte{0, 0, 0, 0, 0, 0, 240, 63}, // 1.0 in little-endian IEEE 754
			1.0,
		},
		{
			"negative one",
			[]byte{0, 0, 0, 0, 0, 0, 240, 191}, // -1.0 in little-endian IEEE 754
			-1.0,
		},
		{
			"pi approx",
			[]byte{24, 45, 68, 84, 251, 33, 9, 64}, // 3.141592653589793
			3.141592653589793,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewReader(bytes.NewReader(tt.data))
			val, err := r.ReadDouble()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if math.Abs(val-tt.expected) > 1e-10 {
				t.Errorf("got %v, want %v", val, tt.expected)
			}
		})
	}
}

func TestReader_ReadCString_Short(t *testing.T) {
	// Short string (length < 255): 1-byte length prefix
	// "test" in Shift-JIS (ASCII compatible for basic chars)
	data := []byte{4, 't', 'e', 's', 't'}
	r := NewReader(bytes.NewReader(data))
	val, err := r.ReadCString()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "test" {
		t.Errorf("got %q, want %q", val, "test")
	}
}

func TestReader_ReadCString_Empty(t *testing.T) {
	// Empty string: length = 0
	data := []byte{0}
	r := NewReader(bytes.NewReader(data))
	val, err := r.ReadCString()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "" {
		t.Errorf("got %q, want empty string", val)
	}
}

func TestReader_ReadCString_Medium(t *testing.T) {
	// Medium string (length >= 255): 0xFF prefix + 2-byte length
	// Create a 300-byte string
	strLen := 300
	expectedStr := make([]byte, strLen)
	for i := range expectedStr {
		expectedStr[i] = 'a' // Fill with 'a'
	}

	data := make([]byte, 1+2+strLen)
	data[0] = 0xFF              // 2-byte length marker
	data[1] = byte(strLen)      // Low byte
	data[2] = byte(strLen >> 8) // High byte
	copy(data[3:], expectedStr)

	r := NewReader(bytes.NewReader(data))
	val, err := r.ReadCString()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != string(expectedStr) {
		t.Errorf("got string of length %d, want %d", len(val), strLen)
	}
}

func TestReader_ReadBytes(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	r := NewReader(bytes.NewReader(data))
	buf := make([]byte, 5)
	err := r.ReadBytes(buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, b := range buf {
		if b != data[i] {
			t.Errorf("buf[%d] = %d, want %d", i, b, data[i])
		}
	}
}

func TestReader_Skip(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	r := NewReader(bytes.NewReader(data))

	// Skip first 4 bytes
	err := r.Skip(4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read next byte should be 5
	val, err := r.ReadBYTE()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 5 {
		t.Errorf("got %d, want 5", val)
	}
}

func TestReader_ReadSignature_Valid(t *testing.T) {
	data := []byte("JwwData.")
	r := NewReader(bytes.NewReader(data))
	err := r.ReadSignature()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestReader_ReadSignature_Invalid(t *testing.T) {
	data := []byte("NotJwwD.")
	r := NewReader(bytes.NewReader(data))
	err := r.ReadSignature()
	if err != ErrInvalidSignature {
		t.Errorf("expected ErrInvalidSignature, got: %v", err)
	}
}
