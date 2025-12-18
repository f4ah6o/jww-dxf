# jww-dxf

JWW (Jw_cad) file parser and DXF converter for JavaScript/TypeScript.

This package uses WebAssembly to parse JWW binary files (created by [Jw_cad](https://www.jwcad.net/), a popular Japanese CAD software) and convert them to DXF format.

## Installation

```bash
npm install jww-dxf
```

## Usage

### Node.js

```typescript
import { createParser } from 'jww-dxf';
import { readFileSync } from 'fs';

async function main() {
  // Create and initialize the parser
  const parser = await createParser();

  // Read a JWW file
  const jwwData = readFileSync('drawing.jww');
  const data = new Uint8Array(jwwData);

  // Parse JWW to get document structure
  const doc = parser.parse(data);
  console.log('Entities:', doc.Entities.length);
  console.log('Layers:', doc.LayerGroups.length);

  // Convert to DXF string
  const dxfString = parser.toDxfString(data);
  writeFileSync('output.dxf', dxfString);
}

main();
```

### Browser

```html
<!-- Include wasm_exec.js from Go -->
<script src="wasm_exec.js"></script>
<script type="module">
  import { createParser } from 'jww-dxf';

  async function convertFile(file) {
    const parser = await createParser('path/to/jww-dxf.wasm');

    const buffer = await file.arrayBuffer();
    const data = new Uint8Array(buffer);

    // Get DXF content
    const dxfString = parser.toDxfString(data);

    // Download as file
    const blob = new Blob([dxfString], { type: 'application/dxf' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = file.name.replace('.jww', '.dxf');
    a.click();
  }
</script>
```

## API

### `createParser(wasmPath?: string): Promise<JwwParser>`

Create and initialize a JWW parser instance.

- `wasmPath` - Optional path to the `jww-dxf.wasm` file

### `JwwParser`

#### `init(): Promise<void>`

Initialize the WASM module. Called automatically by `createParser()`.

#### `parse(data: Uint8Array): JwwDocument`

Parse a JWW file and return the document structure.

#### `toDxf(data: Uint8Array): DxfDocument`

Parse a JWW file and convert to DXF document structure (JSON).

#### `toDxfString(data: Uint8Array): string`

Parse a JWW file and convert to DXF file content string.

## Types

### JwwDocument

```typescript
interface JwwDocument {
  Version: number;
  Memo: string;
  PaperSize: number;
  LayerGroups: LayerGroup[];
  Entities: JwwEntity[];
  Blocks: JwwBlock[];
}
```

### JwwEntity

Supported entity types:
- `JwwLine` - Line segments
- `JwwArc` - Arcs and circles
- `JwwPoint` - Points
- `JwwText` - Text annotations
- `JwwSolid` - Solid fills
- `JwwBlockRef` - Block references

### DxfDocument

```typescript
interface DxfDocument {
  Layers: DxfLayer[];
  Entities: DxfEntity[];
  Blocks: DxfBlock[];
}
```

## Requirements

- Node.js >= 18.0.0 (for Node.js usage)
- Modern browser with WebAssembly support (for browser usage)

## License

AGPL-3.0
