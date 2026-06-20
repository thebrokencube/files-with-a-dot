BINARY    := jf
GOFILES   := $(shell find . ../../pkg/dendrik -name '*.go')
GO        := go
VERSION   := $(shell cat VERSION 2>/dev/null || echo dev)
LDFLAGS   := -buildid= -X main.version=$(VERSION)

.PHONY: build test vet fmt check bundle clean

build: $(BINARY)            ## Compile binary

$(BINARY): $(GOFILES) go.mod $(wildcard go.sum)
	$(GO) build -trimpath -buildvcs=false -ldflags "$(LDFLAGS)" -o $(BINARY) .

test:                        ## Run all tests
	$(GO) test ./... -v

vet:                         ## Static analysis
	$(GO) vet ./...

fmt:                         ## Format all Go source
	gofmt -w .

check: fmt vet test          ## Full pre-commit check: format, vet, test

bundle:                      ## Rebuild JS bundles (requires npm)
	cd scripts && npm install && npx esbuild md2adf.mjs --bundle --format=esm --platform=node --outfile=md2adf.bundle.mjs
	cp scripts/md2adf.bundle.mjs internal/pipeline/md2adf.bundle.mjs
	cd scripts && npx esbuild adf2md.mjs --bundle --format=esm --platform=node --outfile=adf2md.bundle.mjs
	cp scripts/adf2md.bundle.mjs internal/pipeline/adf2md.bundle.mjs

clean:                       ## Remove build artifacts
	rm -f $(BINARY)
