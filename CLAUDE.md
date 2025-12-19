# jww-parser - Claude Code Project Guide

## Project Overview

jww-parser is a Go library and toolset for parsing JWW (Jw_cad) files and converting them to DXF format. The project includes:

- **Go Library**: Core parser and DXF converter
- **CLI Tools**: Command-line utilities for file conversion and inspection
- **WebAssembly**: Browser-compatible WASM build
- **NPM Package**: JavaScript/TypeScript bindings for the WASM module

## Development Requirements

### Core Dependencies

- **Go**: Version 1.25.0 or higher
- **Node.js**: Version 18.0.0 or higher (for npm package development)
- **pnpm**: For npm package dependency management

### Go Dependencies

The project uses minimal external dependencies:
- `golang.org/x/text v0.32.0` - Text encoding/decoding support

## Project Structure

```
jww-parser/
├── jww/                  # Core JWW parser library
├── dxf/                  # DXF converter and entity builders
├── cmd/
│   ├── jww-parser/       # CLI tool for conversion
│   ├── jww-stats/        # CLI tool for file inspection
│   └── jww-debug/        # Debug utilities
├── wasm/                 # WebAssembly frontend and bindings
├── npm/                  # NPM package for JavaScript/TypeScript
├── examples/             # Example JWW files
├── refs/                 # Reference documentation
│   └── DXFFileStructure/ # DXF format specifications
└── AGENTS.md            # DXF file structure reference (see @AGENTS.md)
```

## Common Development Commands

### Building

```bash
# Build native CLI binary
make build

# Build statistics tool
make build-stats

# Build WebAssembly module
make build-wasm

# Build complete distribution (WASM + assets)
make dist

# Build NPM package
make build-npm
```

### Testing

```bash
# Run all Go tests
make test
# or
go test -v ./...

# Show JWW file statistics
make stat
# or
./bin/jww-stats examples/jww
```

### File Conversion

```bash
# Convert single JWW file to DXF
./bin/jww-parser -o output.dxf input.jww

# Convert all example files
make convert-examples
```

### Cleaning

```bash
make clean           # Clean all build artifacts
make clean-bin       # Clean binary files only
make clean-dist      # Clean distribution files only
make clean-converted # Clean converted example files
```

## Development Workflow

### Working with Go Code

1. **Parser Development** (`jww/` directory):
   - Modify parser logic in `parser.go`
   - Update type definitions in `types.go`
   - Add tests in `*_test.go` files
   - Run tests: `go test ./jww/...`

2. **DXF Converter** (`dxf/` directory):
   - Modify conversion logic in `converter.go`
   - Add entity builders using functional options pattern
   - Update tests: `go test ./dxf/...`

### Working with WebAssembly

1. Modify WASM bindings in `wasm/` directory
2. Build: `make build-wasm`
3. Test locally by serving `dist/` directory
4. Copy to npm: `make build-npm`

### Working with NPM Package

1. Navigate to `npm/` directory
2. Install dependencies: `pnpm install`
3. Build TypeScript: `npm run build:js`
4. Full build: `npm run build` (includes WASM)

## Testing Strategy

### Unit Tests
- Each package has comprehensive unit tests
- Use table-driven tests for parser validation
- Mock I/O for reliable testing

### Integration Tests
- `e2e_test.go` validates full conversion pipeline
- Test files from Jw_cad official distribution
- Verify entity counts match source files

### Validation Tools
- **ezdxf**: Python library for DXF validation
- **ODA FileConverter**: Industry-standard converter (see Known Issues)

## DXF File Format Reference

For detailed information about DXF file structure, see @AGENTS.md which contains:
- DXF file organization (HEADER, CLASSES, TABLES, BLOCKS, ENTITIES, OBJECTS sections)
- Group code specifications
- Entity definitions and attributes
- Based on AutoCAD 2024 Developer and ObjectARX Help documentation

## Known Issues

### ODA FileConverter Compatibility

Generated DXF files pass ezdxf validation but produce errors in ODA FileConverter:
- "Record name is empty - Ignored" (layer table)
- "Syntax error or premature end of file"
- "Null object Id"

The cause is under investigation. Basic DXF structure (HEADER, TABLES, ENTITIES) is correct, but ODA may expect stricter formatting.

## API Usage Examples

### Parsing JWW Files

```go
import "github.com/f4ah6o/jww-parser/jww"

f, _ := os.Open("example.jww")
defer f.Close()

doc, err := jww.Parse(f)
if err != nil {
    panic(err)
}
```

### Creating DXF Entities

```go
import "github.com/f4ah6o/jww-parser/dxf"

// Create entities with functional options
line := dxf.NewLine(0, 0, 100, 100,
    dxf.WithLineLayer("MyLayer"),
    dxf.WithLineColor(1))

circle := dxf.NewCircle(50, 50, 25,
    dxf.WithCircleLayer("MyLayer"))

// Build document with fluent API
doc := dxf.NewDocument().
    AddLayer("Layer1", 1, "CONTINUOUS").
    AddLine(0, 0, 100, 100).
    AddCircle(50, 50, 25)

// Export to DXF
dxfString := dxf.ToString(doc)
```

## Continuous Integration

GitHub Actions workflows:
- `gh-pages.yml`: Deploys WASM demo to GitHub Pages
- `npm-publish.yaml`: Publishes npm package on release

## License

This project is licensed under the [GNU Affero General Public License v3.0](https://www.gnu.org/licenses/agpl-3.0.html).

## Related Resources

- [Jw_cad Official Site](https://www.jwcad.net/)
- [AutoCAD DXF Reference](https://help.autodesk.com/view/OARX/2024/ENU/)
- [ezdxf Documentation](https://ezdxf.readthedocs.io/)
- [ODA File Converter](https://www.opendesign.com/guestfiles/oda_file_Converter)
