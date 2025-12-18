//go:build js && wasm

// Package main provides WebAssembly exports for JWW parsing.
package main

import (
	"bytes"
	"encoding/json"
	"syscall/js"

	"github.com/f4ah6o/jww-dxf/dxf"
	"github.com/f4ah6o/jww-dxf/jww"
)

func main() {
	// Register JavaScript functions
	js.Global().Set("jwwParse", js.FuncOf(jwwParse))
	js.Global().Set("jwwToDxf", js.FuncOf(jwwToDxf))
	js.Global().Set("jwwToDxfString", js.FuncOf(jwwToDxfString))

	// Keep the program running
	<-make(chan struct{})
}

// jwwParse parses JWW binary data and returns JSON representation.
// JS: jwwParse(Uint8Array) -> Promise<string>
func jwwParse(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return makeError("jwwParse requires 1 argument: Uint8Array")
	}

	// Get Uint8Array data
	data := jsArrayToBytes(args[0])

	// Parse JWW data
	doc, err := jww.Parse(bytes.NewReader(data))
	if err != nil {
		return makeError("parse error: " + err.Error())
	}

	// Convert to JSON
	jsonData, err := json.Marshal(doc)
	if err != nil {
		return makeError("JSON marshal error: " + err.Error())
	}

	return makeResult(string(jsonData))
}

// jwwToDxf parses JWW binary data and returns DXF object as JSON.
// JS: jwwToDxf(Uint8Array) -> Promise<string>
func jwwToDxf(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return makeError("jwwToDxf requires 1 argument: Uint8Array")
	}

	// Get Uint8Array data
	data := jsArrayToBytes(args[0])

	// Parse JWW data
	jwwDoc, err := jww.Parse(bytes.NewReader(data))
	if err != nil {
		return makeError("parse error: " + err.Error())
	}

	// Convert to DXF
	dxfDoc := dxf.ConvertDocument(jwwDoc)

	// Convert to JSON
	jsonData, err := json.Marshal(dxfDoc)
	if err != nil {
		return makeError("JSON marshal error: " + err.Error())
	}

	return makeResult(string(jsonData))
}

// jwwToDxfString parses JWW binary data and returns DXF file content as string.
// JS: jwwToDxfString(Uint8Array) -> Promise<string>
func jwwToDxfString(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return makeError("jwwToDxfString requires 1 argument: Uint8Array")
	}

	// Get Uint8Array data
	data := jsArrayToBytes(args[0])

	// Parse JWW data
	jwwDoc, err := jww.Parse(bytes.NewReader(data))
	if err != nil {
		return makeError("parse error: " + err.Error())
	}

	// Convert to DXF
	dxfDoc := dxf.ConvertDocument(jwwDoc)

	// Convert to DXF string
	dxfString := dxf.ToString(dxfDoc)

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
	return map[string]interface{}{
		"ok":    false,
		"error": message,
	}
}
