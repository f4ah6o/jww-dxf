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

# Install frontend dependencies for the WASM demo
install-wasm-deps:
	cd wasm && npm ci

# Build WebAssembly
build-wasm: clean-dist
	mkdir -p wasm/public
# Embed Version and CommitHash into the WASM binary via -ldflags
	GOOS=js GOARCH=wasm go build -ldflags="-s -w -X main.Version=$(VERSION) -X main.CommitHash=$(COMMIT_HASH)" -o wasm/public/jww-parser.wasm ./wasm/

# Copy wasm_exec.js from Go installation
copy-wasm-exec:
	mkdir -p wasm/public
	if [ -f "$$(go env GOROOT)/misc/wasm/wasm_exec.js" ]; then \
		cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" wasm/public/; \
	else \
		cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" wasm/public/; \
	fi

# Build static assets for the WASM demo (Vite)
copy-wasm-assets: install-wasm-deps build-wasm copy-wasm-exec
	cd wasm && VITE_COMMIT_HASH=$(COMMIT_HASH) npm run build
	rm -rf dist
	cp -r wasm/dist dist

# Build WASM and copy support files
dist: copy-wasm-assets

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
build-npm: dist
	mkdir -p npm/wasm
	cp dist/jww-parser.wasm npm/wasm/
	cp dist/wasm_exec.js npm/wasm/
	cd npm && npm install && npm run build:js
