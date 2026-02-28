BINARY    := folio
INSTALL   := $(HOME)/.local/bin/$(BINARY)
GOFILES   := $(shell find . -name '*.go')
GO        := go

.PHONY: build test lint fmt vet check install clean

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

install: build               ## Build and install to ~/.local/bin
	@mkdir -p $(dir $(INSTALL))
	cp $(BINARY) $(INSTALL)

clean:                       ## Remove build artifacts
	rm -f $(BINARY)
