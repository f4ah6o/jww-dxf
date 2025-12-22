.PHONY: build build-wasm test stat clean convert-examples clean-bin clean-dist clean-converted copy-wasm-assets build-npm

COMMIT_HASH := $(shell git rev-parse --short HEAD)

# VERSION will be taken from the latest tag (without leading 'v') when available,
# otherwise falls back to 'dev'. You can also override by invoking `make VERSION=...`.
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//')
VERSION := $(if $(VERSION),$(VERSION),dev)

# Build native binary
build: clean-bin
	go build -o bin/jww-parser ./cmd/jww-parser

build-stats: clean-bin
	go build -o bin/jww-stats ./cmd/jww-stats

# Build WebAssembly
build-wasm: clean-dist
	rm -rf dist/
	mkdir -p dist
	# Embed Version and CommitHash into the WASM binary via -ldflags
	GOOS=js GOARCH=wasm go build -ldflags="-s -w -X main.Version=$(VERSION) -X main.CommitHash=$(COMMIT_HASH)" -o dist/jww-parser.wasm ./wasm/

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
	cp wasm/styles.css dist/
	sed 's/__COMMIT_HASH__/$(COMMIT_HASH)/g' wasm/app.js > dist/app.js
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
			./bin/jww-parser -o "examples/converted/$$(basename "$$f" .jww).dxf" "$$f"; \
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

# Build npm package
build-npm: build-wasm copy-wasm-exec
	mkdir -p npm/wasm
	cp dist/jww-parser.wasm npm/wasm/
	cp dist/wasm_exec.js npm/wasm/
	cd npm && npm install && npm run build:js
