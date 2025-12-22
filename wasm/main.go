//go:build js && wasm

// Package main provides WebAssembly exports for JWW parsing.
package main

import (
	"bytes"
	"encoding/json"
	"syscall/js"

	"github.com/f4ah6o/jww-parser/dxf"
	"github.com/f4ah6o/jww-parser/jww"
)

// Version of the WASM module (overridden at build time via -ldflags)
var Version = "dev"

// CommitHash of the WASM module (overridden at build time via -ldflags)
var CommitHash = "unknown"

// debugMode controls verbose logging
var debugMode bool

func main() {
	// Register JavaScript functions
	js.Global().Set("jwwParse", js.FuncOf(jwwParse))
	js.Global().Set("jwwToDxf", js.FuncOf(jwwToDxf))
	js.Global().Set("jwwToDxfString", js.FuncOf(jwwToDxfString))
	js.Global().Set("jwwGetVersion", js.FuncOf(jwwGetVersion))
	js.Global().Set("jwwSetDebug", js.FuncOf(jwwSetDebug))
	js.Global().Set("jwwCommitHash", js.FuncOf(jwwCommitHash))

	// Keep the program running
	<-make(chan struct{})
}

// jwwGetVersion returns the WASM module version.
// JS: jwwGetVersion() -> string
func jwwGetVersion(this js.Value, args []js.Value) interface{} {
	return Version
}

// jwwCommitHash returns the WASM module commit hash.
// JS: jwwCommitHash() -> string
func jwwCommitHash(this js.Value, args []js.Value) interface{} {
	return CommitHash
}

// jwwSetDebug enables or disables debug mode.
// JS: jwwSetDebug(enabled: boolean) -> void
func jwwSetDebug(this js.Value, args []js.Value) interface{} {
	if len(args) >= 1 {
		debugMode = args[0].Bool()
		if debugMode {
			logDebug("Debug mode enabled")
		}
	}
	return nil
}

// logDebug logs a message if debug mode is enabled.
func logDebug(format string, args ...interface{}) {
	if debugMode {
		console := js.Global().Get("console")
		if len(args) == 0 {
			console.Call("log", "[JWW-WASM] "+format)
		} else {
			// Simple formatting
			console.Call("log", "[JWW-WASM] "+format, args)
		}
	}
}

// jwwParse parses JWW binary data and returns JSON representation.
// JS: jwwParse(Uint8Array) -> { ok: boolean, data?: string, error?: string }
func jwwParse(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return makeError("jwwParse requires 1 argument: Uint8Array")
	}

	logDebug("Starting parse operation")

	// Get Uint8Array data
	data := jsArrayToBytes(args[0])
	logDebug("Received %d bytes", len(data))

	// Parse JWW data
	doc, err := jww.Parse(bytes.NewReader(data))
	if err != nil {
		logDebug("Parse error: %v", err.Error())
		return makeError("parse error: " + err.Error())
	}

	logDebug("Parsed document with %d entities", len(doc.Entities))

	// Convert to JSON
	jsonData, err := json.Marshal(doc)
	if err != nil {
		logDebug("JSON marshal error: %v", err.Error())
		return makeError("JSON marshal error: " + err.Error())
	}

	logDebug("Generated %d bytes of JSON", len(jsonData))
	return makeResult(string(jsonData))
}

// jwwToDxf parses JWW binary data and returns DXF object as JSON.
// JS: jwwToDxf(Uint8Array) -> { ok: boolean, data?: string, error?: string }
func jwwToDxf(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return makeError("jwwToDxf requires 1 argument: Uint8Array")
	}

	logDebug("Starting DXF conversion")

	// Get Uint8Array data
	data := jsArrayToBytes(args[0])
	logDebug("Received %d bytes", len(data))

	// Parse JWW data
	jwwDoc, err := jww.Parse(bytes.NewReader(data))
	if err != nil {
		logDebug("Parse error: %v", err.Error())
		return makeError("parse error: " + err.Error())
	}

	logDebug("Parsed JWW document with %d entities", len(jwwDoc.Entities))

	// Convert to DXF
	dxfDoc := dxf.ConvertDocument(jwwDoc)
	logDebug("Converted to DXF with %d entities", len(dxfDoc.Entities))

	// Convert to JSON
	jsonData, err := json.Marshal(dxfDoc)
	if err != nil {
		logDebug("JSON marshal error: %v", err.Error())
		return makeError("JSON marshal error: " + err.Error())
	}

	logDebug("Generated %d bytes of JSON", len(jsonData))
	return makeResult(string(jsonData))
}

// jwwToDxfString parses JWW binary data and returns DXF file content as string.
// JS: jwwToDxfString(Uint8Array) -> { ok: boolean, data?: string, error?: string }
func jwwToDxfString(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return makeError("jwwToDxfString requires 1 argument: Uint8Array")
	}

	logDebug("Starting DXF string generation")

	// Get Uint8Array data
	data := jsArrayToBytes(args[0])
	logDebug("Received %d bytes", len(data))

	// Parse JWW data
	jwwDoc, err := jww.Parse(bytes.NewReader(data))
	if err != nil {
		logDebug("Parse error: %v", err.Error())
		return makeError("parse error: " + err.Error())
	}

	logDebug("Parsed JWW document with %d entities", len(jwwDoc.Entities))

	// Convert to DXF
	dxfDoc := dxf.ConvertDocument(jwwDoc)
	logDebug("Converted to DXF with %d entities", len(dxfDoc.Entities))

	// Convert to DXF string
	dxfString := dxf.ToString(dxfDoc)
	logDebug("Generated %d bytes of DXF string", len(dxfString))

	return makeResult(dxfString)
}

// jsArrayToBytes converts a JavaScript Uint8Array to Go []byte.
func jsArrayToBytes(arr js.Value) []byte {
	length := arr.Length()
	data := make([]byte, length)
	js.CopyBytesToGo(data, arr)
	return data
}

// makeResult creates a successful result object.
func makeResult(data string) map[string]interface{} {
	return map[string]interface{}{
		"ok":   true,
		"data": data,
	}
}

// makeError creates an error result object.
func makeError(message string) map[string]interface{} {
	logDebug("Error: %s", message)
	return map[string]interface{}{
		"ok":    false,
		"error": message,
	}
}
