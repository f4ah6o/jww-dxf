.PHONY: build build-wasm test stat clean convert-examples clean-bin clean-dist clean-converted copy-wasm-assets

# Build native binary
build: clean-bin
	go build -o bin/jww-dxf ./cmd/jww-dxf

build-stats: clean-bin
	go build -o bin/jww-stats ./cmd/jww-stats

# Build WebAssembly
build-wasm: clean-dist
	rm -rf dist/
	mkdir -p dist
	GOOS=js GOARCH=wasm go build -o dist/jww-dxf.wasm ./wasm/

# Copy wasm_exec.js from Go installation
copy-wasm-exec:
	mkdir -p dist
	if [ -f "$$(go env GOROOT)/misc/wasm/wasm_exec.js" ]; then \
		cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" dist/; \
	else \
		cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" dist/; \
	fi

# Copy static assets for the WASM demo
copy-wasm-assets:
	mkdir -p dist
	cp wasm/example.html dist/index.html
	cp wasm/styles.css wasm/app.js dist/
	cp -r wasm/vendor dist/

# Build WASM and copy support files
dist: build-wasm copy-wasm-exec copy-wasm-assets

# Run tests
test:
	go test -v ./...

stat: build-stats
	./bin/jww-stats examples/jww

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
