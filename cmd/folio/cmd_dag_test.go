package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunDagMissingFile(t *testing.T) {
	code := runDag([]string{"--folio", "/nonexistent/folio.yml"})
	if code != 1 {
		t.Errorf("expected exit code 1 for missing file, got %d", code)
	}
}

func TestRunDagNoTargets(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\n"), 0644)

	code := runDag([]string{"--folio", yml})
	if code != 0 {
		t.Errorf("expected exit code 0 for no targets, got %d", code)
	}
}

func TestRunDagWithEdges(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  downstream:
    transform: distill
    blocked_by: [upstream]
    outputs:
      - path: compiled/down.md
  upstream:
    transform: distill
    outputs:
      - path: compiled/up.md
`), 0644)

	code := runDag([]string{"--folio", yml})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunDagJSON(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  downstream:
    transform: distill
    blocked_by: [upstream]
    outputs:
      - path: compiled/down.md
  upstream:
    transform: distill
    outputs:
      - path: compiled/up.md
`), 0644)

	code := runDag([]string{"--folio", yml, "--json"})
	if code != 0 {
		t.Errorf("expected exit code 0 for JSON mode, got %d", code)
	}
}
