.PHONY: build build-wasm test stat clean convert-examples clean-bin clean-dist clean-converted

# Build native binary
build: clean-bin
	go build -o bin/jww-dxf ./cmd/jww-dxf

# Build WebAssembly
build-wasm: clean-dist
	rm -rf dist/
	mkdir -p dist
	GOOS=js GOARCH=wasm go build -o dist/jww-dxf.wasm ./wasm/

# Copy wasm_exec.js from Go installation
copy-wasm-exec:
	cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" dist/

# Build WASM and copy support files
dist: build-wasm copy-wasm-exec

# Run tests
test:
	go test -v ./...

stat:
	go run ./cmd/jww-stats/ examples/jww

# Convert all JWW files in examples/jww to DXF and save to examples/converted
convert-examples: build clean-converted
	@mkdir -p examples/converted
	@for f in examples/jww/*.jww; do \
		if [ -f "$$f" ]; then \
			echo "Converting $$f..."; \
			./bin/jww-dxf -o "examples/converted/$$(basename "$$f" .jww).dxf" "$$f"; \
		fi \
	done
	@echo "Done. Converted files are in examples/converted/"

# Clean build artifacts
clean: clean-bin clean-dist clean-converted

clean-bin:
	rm -rf bin/

clean-dist:
	rm -rf dist/

clean-converted:
	rm -rf examples/converted
