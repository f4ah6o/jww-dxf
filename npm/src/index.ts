/**
 * JWW (Jw_cad) file parser and DXF converter
 *
 * This module provides functionality to parse JWW binary files
 * and convert them to DXF format using WebAssembly.
 */

// Type definitions for JWW document structure
export interface JwwDocument {
  Version: number;
  Memo: string;
  PaperSize: number;
  LayerGroups: LayerGroup[];
  Entities: JwwEntity[];
  Blocks: JwwBlock[];
}

export interface LayerGroup {
  Name: string;
  Layers: Layer[];
}

export interface Layer {
  Name: string;
  Visible: boolean;
  Locked: boolean;
}

export interface JwwEntityBase {
  Type: string;
  Group: number;
  PenStyle: number;
  PenColor: number;
  PenWidth: number;
  Layer: number;
  LayerGroup: number;
}

export interface JwwLine extends JwwEntityBase {
  Type: "Line";
  X1: number;
  Y1: number;
  X2: number;
  Y2: number;
}

export interface JwwArc extends JwwEntityBase {
  Type: "Arc";
  CenterX: number;
  CenterY: number;
  Radius: number;
  StartAngle: number;
  EndAngle: number;
  Flatness: number;
}

export interface JwwPoint extends JwwEntityBase {
  Type: "Point";
  X: number;
  Y: number;
  Code: number;
}

export interface JwwText extends JwwEntityBase {
  Type: "Text";
  X: number;
  Y: number;
  Text: string;
  FontName: string;
  Height: number;
  Width: number;
  Angle: number;
}

export interface JwwSolid extends JwwEntityBase {
  Type: "Solid";
  Points: [number, number][];
}

export interface JwwBlockRef extends JwwEntityBase {
  Type: "Block";
  X: number;
  Y: number;
  ScaleX: number;
  ScaleY: number;
  Angle: number;
  BlockNumber: number;
}

export type JwwEntity =
  | JwwLine
  | JwwArc
  | JwwPoint
  | JwwText
  | JwwSolid
  | JwwBlockRef;

export interface JwwBlock {
  Name: string;
  Entities: JwwEntity[];
}

// Type definitions for DXF document structure
export interface DxfDocument {
  Layers: DxfLayer[];
  Entities: DxfEntity[];
  Blocks: DxfBlock[];
}

export interface DxfLayer {
  Name: string;
  Color: number;
  Frozen: boolean;
  Locked: boolean;
}

export interface DxfEntity {
  Type: string;
  Layer: string;
  Color?: number;
  [key: string]: unknown;
}

export interface DxfBlock {
  Name: string;
  Entities: DxfEntity[];
}

// WASM result type
interface WasmResult {
  ok: boolean;
  data?: string;
  error?: string;
}

// Global declarations for WASM functions
declare global {
  var Go: new () => GoInstance;
  var jwwParse: (data: Uint8Array) => WasmResult;
  var jwwToDxf: (data: Uint8Array) => WasmResult;
  var jwwToDxfString: (data: Uint8Array) => WasmResult;
}

interface GoInstance {
  importObject: WebAssembly.Imports;
  run(instance: WebAssembly.Instance): Promise<void>;
}

// Parser class
export class JwwParser {
  private initialized = false;
  private initPromise: Promise<void> | null = null;
  private wasmPath: string;

  /**
   * Create a new JWW parser instance
   * @param wasmPath - Path to the jww-dxf.wasm file
   */
  constructor(wasmPath?: string) {
    this.wasmPath = wasmPath || this.getDefaultWasmPath();
  }

  private getDefaultWasmPath(): string {
    // Try to determine the path based on the environment
    if (typeof process !== "undefined" && process.versions?.node) {
      // Node.js environment
      return new URL("../wasm/jww-dxf.wasm", import.meta.url).pathname;
    }
    // Browser environment
    return "jww-dxf.wasm";
  }

  /**
   * Initialize the WASM module
   * Must be called before using parse methods
   */
  async init(): Promise<void> {
    if (this.initialized) return;
    if (this.initPromise) return this.initPromise;

    this.initPromise = this.loadWasm();
    await this.initPromise;
    this.initialized = true;
  }

  private async loadWasm(): Promise<void> {
    // Load wasm_exec.js if Go is not defined
    if (typeof Go === "undefined") {
      await this.loadWasmExec();
    }

    const go = new Go();

    let wasmInstance: WebAssembly.Instance;

    if (typeof process !== "undefined" && process.versions?.node) {
      // Node.js environment
      const fs = await import("fs");
      const path = await import("path");
      const wasmBuffer = fs.readFileSync(this.wasmPath);
      const wasmModule = await WebAssembly.compile(wasmBuffer);
      wasmInstance = await WebAssembly.instantiate(wasmModule, go.importObject);
    } else {
      // Browser environment
      const result = await WebAssembly.instantiateStreaming(
        fetch(this.wasmPath),
        go.importObject
      ).catch(async () => {
        // Fallback for browsers that don't support instantiateStreaming
        const response = await fetch(this.wasmPath);
        const bytes = await response.arrayBuffer();
        return WebAssembly.instantiate(bytes, go.importObject);
      });
      wasmInstance = result.instance;
    }

    // Don't await - Go.run() blocks until the program exits
    go.run(wasmInstance);

    // Wait for functions to be available
    await this.waitForWasmFunctions();
  }

  private async loadWasmExec(): Promise<void> {
    if (typeof process !== "undefined" && process.versions?.node) {
      // Node.js - require wasm_exec.js
      const wasmExecPath = new URL("../wasm/wasm_exec.js", import.meta.url)
        .pathname;
      await import(wasmExecPath);
    } else {
      throw new Error(
        "Go runtime not loaded. Please include wasm_exec.js in your HTML."
      );
    }
  }

  private async waitForWasmFunctions(
    timeout = 5000,
    interval = 50
  ): Promise<void> {
    const start = Date.now();
    while (Date.now() - start < timeout) {
      if (
        typeof globalThis.jwwParse === "function" &&
        typeof globalThis.jwwToDxf === "function" &&
        typeof globalThis.jwwToDxfString === "function"
      ) {
        return;
      }
      await new Promise((resolve) => setTimeout(resolve, interval));
    }
    throw new Error("WASM functions not available after timeout");
  }

  private ensureInitialized(): void {
    if (!this.initialized) {
      throw new Error("Parser not initialized. Call init() first.");
    }
  }

  /**
   * Parse a JWW file and return the document structure
   * @param data - JWW file content as Uint8Array
   * @returns Parsed JWW document
   */
  parse(data: Uint8Array): JwwDocument {
    this.ensureInitialized();
    const result = globalThis.jwwParse(data);
    if (!result.ok) {
      throw new Error(result.error || "Parse failed");
    }
    return JSON.parse(result.data!) as JwwDocument;
  }

  /**
   * Parse a JWW file and convert to DXF document structure
   * @param data - JWW file content as Uint8Array
   * @returns DXF document object
   */
  toDxf(data: Uint8Array): DxfDocument {
    this.ensureInitialized();
    const result = globalThis.jwwToDxf(data);
    if (!result.ok) {
      throw new Error(result.error || "Conversion failed");
    }
    return JSON.parse(result.data!) as DxfDocument;
  }

  /**
   * Parse a JWW file and convert to DXF file content string
   * @param data - JWW file content as Uint8Array
   * @returns DXF file content as string
   */
  toDxfString(data: Uint8Array): string {
    this.ensureInitialized();
    const result = globalThis.jwwToDxfString(data);
    if (!result.ok) {
      throw new Error(result.error || "Conversion failed");
    }
    return result.data!;
  }
}

/**
 * Create and initialize a JWW parser instance
 * @param wasmPath - Optional path to the jww-dxf.wasm file
 * @returns Initialized JwwParser instance
 */
export async function createParser(wasmPath?: string): Promise<JwwParser> {
  const parser = new JwwParser(wasmPath);
  await parser.init();
  return parser;
}

// Default export
export default { JwwParser, createParser };
