.PHONY: build build-wasm test clean

# Build native binary
build:
	go build -o bin/jww-dxf ./cmd/jww-dxf

# Build WebAssembly
build-wasm:
	GOOS=js GOARCH=wasm go build -o dist/jww-dxf.wasm ./wasm/

# Copy wasm_exec.js from Go installation
copy-wasm-exec:
	cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" dist/

# Build WASM and copy support files
dist: build-wasm copy-wasm-exec

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf bin/ dist/
