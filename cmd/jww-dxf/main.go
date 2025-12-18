// Command jww-dxf parses JWW files and optionally outputs DXF.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/f4ah6o/jww-dxf/dxf"
	"github.com/f4ah6o/jww-dxf/jww"
)

func main() {
	outputDxf := flag.Bool("dxf", false, "Output DXF format")
	outputFile := flag.String("o", "", "Output file (default: stdout)")
	verbose := flag.Bool("v", false, "Verbose output")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <input.jww>\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	inputFile := flag.Arg(0)

	// Open input file
	f, err := os.Open(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	// Parse JWW file
	doc, err := jww.Parse(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JWW: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "JWW File: %s\n", inputFile)
		fmt.Fprintf(os.Stderr, "  Version: %d\n", doc.Version)
		fmt.Fprintf(os.Stderr, "  Memo: %s\n", doc.Memo)
		fmt.Fprintf(os.Stderr, "  Paper Size: %d\n", doc.PaperSize)
		fmt.Fprintf(os.Stderr, "  Entities: %d\n", len(doc.Entities))
		fmt.Fprintf(os.Stderr, "  Blocks: %d\n", len(doc.BlockDefs))
	}

	if *outputDxf {
		// Convert to DXF
		dxfDoc := dxf.ConvertDocument(doc)
		dxfStr := dxf.ToString(dxfDoc)

		// Output
		if *outputFile != "" {
			if err := os.WriteFile(*outputFile, []byte(dxfStr), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
				os.Exit(1)
			}
			if *verbose {
				fmt.Fprintf(os.Stderr, "DXF written to: %s\n", *outputFile)
			}
		} else {
			fmt.Print(dxfStr)
		}
	} else if !*verbose {
		// Default: show summary
		fmt.Printf("JWW File: %s\n", inputFile)
		fmt.Printf("  Version: %d\n", doc.Version)
		fmt.Printf("  Memo: %s\n", doc.Memo)
		fmt.Printf("  Paper Size: %d\n", doc.PaperSize)
		fmt.Printf("  Entities: %d\n", len(doc.Entities))
		fmt.Printf("  Blocks: %d\n", len(doc.BlockDefs))
	}
}
