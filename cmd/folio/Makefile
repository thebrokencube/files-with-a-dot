BINARY    := folio
GOFILES   := $(shell find . -name '*.go')
GO        := go

.PHONY: build test lint fmt vet check install clean bundle

build: $(BINARY)            ## Compile binary

$(BINARY): $(GOFILES) go.mod go.sum
	$(GO) build -o $(BINARY) .

test:                        ## Run all tests
	$(GO) test ./... -v

lint: vet                    ## Run all linters
	@echo "lint: pass (vet only — add golangci-lint later if needed)"

vet:                         ## Static analysis
	$(GO) vet ./...

fmt:                         ## Format all Go source
	gofmt -w .

check: fmt vet test          ## Full pre-commit check: format, vet, test

install: build               ## Build (symlink_map.txt handles ~/.local/bin)

bundle:                      ## Rebuild marklassian bundle (requires npm)
	cd scripts && npm install && npx esbuild md2adf.mjs --bundle --format=esm --platform=node --outfile=md2adf.bundle.mjs
	cp scripts/md2adf.bundle.mjs internal/jira/md2adf.bundle.mjs

clean:                       ## Remove build artifacts
	rm -f $(BINARY)
