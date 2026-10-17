BINARY    := folio
GOFILES   := $(shell find . ../../pkg/dendrik -name '*.go')
GO        := go
VERSION   := $(shell cat VERSION 2>/dev/null || echo dev)
LDFLAGS   := -buildid= -X main.version=$(VERSION)

.PHONY: build test lint fmt vet check test-migration test-sketch-diagrams bundle install clean

build: $(BINARY)            ## Compile binary

$(BINARY): $(GOFILES) go.mod go.sum
	$(GO) build -trimpath -buildvcs=false -ldflags "$(LDFLAGS)" -o $(BINARY) .

test:                        ## Run all tests
	$(GO) test ./... -v

lint: vet                    ## Run all linters
	@echo "lint: pass (vet only — add golangci-lint later if needed)"

vet:                         ## Static analysis
	$(GO) vet ./...

fmt:                         ## Format all Go source
	gofmt -w .

check: fmt vet test test-migration test-sketch-diagrams   ## Full pre-commit check: format, vet, test, migration, diagrams

test-migration:             ## Verify container migration root handoff
	bash scripts/migrate-container_test.sh

test-sketch-diagrams:       ## Render the synthetic diagram fixture with the bundled renderer (requires node)
	bash scripts/sketch-diagrams_test.sh

bundle:                      ## Rebuild the skill-local sketch-diagrams bundle (requires npm)
	cd scripts && npm install && npx esbuild sketch-diagrams.mjs --bundle --minify --format=esm --platform=node --outfile=../../../plugins/folio/skills/folio/scripts/sketch-diagrams.mjs

install: build               ## Build (symlink_map.txt handles ~/.local/bin)

clean:                       ## Remove build artifacts
	rm -f $(BINARY)
