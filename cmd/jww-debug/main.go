// Command jww-debug dumps the raw binary structure of a JWW file.
package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input.jww>\n", os.Args[0])
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	// Read signature
	sig := make([]byte, 8)
	io.ReadFull(f, sig)
	fmt.Printf("Signature: %q\n", string(sig))

	// Read version
	var version uint32
	binary.Read(f, binary.LittleEndian, &version)
	fmt.Printf("Version: %d\n", version)

	// Track position
	pos := int64(12)

	// Skip to approximate position where entity list starts
	// This is a rough estimate based on the format

	// First, let's see the file size
	fi, _ := f.Stat()
	fmt.Printf("File size: %d bytes\n", fi.Size())

	// Read memo (CString)
	memoLen, _ := readCStringLen(f)
	fmt.Printf("Memo length: %d\n", memoLen)
	pos += int64(1 + memoLen)
	if memoLen >= 255 {
		pos += 2
	}

	// Skip to end and read backwards to find entity count
	// Actually, let's dump the first few bytes after header offset

	// Calculate rough offset to entity list
	// Based on format analysis, entity list should start around position 8000-15000

	// Let's search for entity count by looking for reasonable numbers
	f.Seek(0, 0)
	data, _ := io.ReadAll(f)

	fmt.Println("\n--- Searching for potential entity list markers ---")

	// Look for CDataSen, CDataEnko, etc strings in the file
	searchStrings := []string{"CDataSen", "CDataEnko", "CDataTen", "CDataMoji", "CDataSolid", "CDataBlock", "CDataList"}
	for _, s := range searchStrings {
		for i := 0; i < len(data)-len(s); i++ {
			if string(data[i:i+len(s)]) == s {
				fmt.Printf("Found %q at offset %d (0x%X)\n", s, i, i)
			}
		}
	}

	fmt.Println("\n--- Dump around entity offsets ---")
	// Search for class name patterns (length prefix followed by "CData")
	for i := 0; i < len(data)-10; i++ {
		if i+5 < len(data) && string(data[i:i+5]) == "CData" {
			start := i - 10
			if start < 0 {
				start = 0
			}
			end := i + 20
			if end > len(data) {
				end = len(data)
			}
			fmt.Printf("Context at %d (0x%X): %v\n", i, i, data[start:end])
		}
	}
}

func readCStringLen(r io.Reader) (int, error) {
	var lenByte uint8
	if err := binary.Read(r, binary.LittleEndian, &lenByte); err != nil {
		return 0, err
	}
	if lenByte < 255 {
		return int(lenByte), nil
	}
	var lenWord uint16
	if err := binary.Read(r, binary.LittleEndian, &lenWord); err != nil {
		return 0, err
	}
	if lenWord < 65535 {
		return int(lenWord), nil
	}
	var lenDword uint32
	if err := binary.Read(r, binary.LittleEndian, &lenDword); err != nil {
		return 0, err
	}
	return int(lenDword), nil
}
