BINARY    := dendrik
GOFILES   := $(shell find . ../../pkg/dendrik -name '*.go')
GO        := go

.PHONY: build test vet fmt check clean

build: $(BINARY)            ## Compile binary

$(BINARY): $(GOFILES) go.mod $(wildcard go.sum)
	$(GO) build -o $(BINARY) .

test:                        ## Run all tests
	$(GO) test ./... -v

vet:                         ## Static analysis
	$(GO) vet ./...

fmt:                         ## Format all Go source
	gofmt -w .

check: fmt vet test          ## Full pre-commit check: format, vet, test

clean:                       ## Remove build artifacts
	rm -f $(BINARY)
