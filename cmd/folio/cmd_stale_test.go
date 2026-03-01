package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunStaleMissingFile(t *testing.T) {
	code := runStale([]string{"--folio", "/nonexistent/folio.yml"})
	if code != 1 {
		t.Errorf("expected exit code 1 for missing file, got %d", code)
	}
}

func TestRunStaleNoTargets(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\n"), 0644)

	code := runStale([]string{"--folio", yml})
	if code != 0 {
		t.Errorf("expected exit code 0 for no targets, got %d", code)
	}
}

func TestRunStaleWithStaleTarget(t *testing.T) {
	dir := t.TempDir()

	// Create source file
	os.WriteFile(filepath.Join(dir, "src.md"), []byte("# Source"), 0644)

	// Create output directory and file with old mtime
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	outPath := filepath.Join(dir, "compiled", "out.md")
	os.WriteFile(outPath, []byte("# Old"), 0644)

	// Touch source to be newer than output
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  my-target:
    transform: distill
    sources:
      - path: src.md
    outputs:
      - path: compiled/out.md
`), 0644)

	// Ensure source is newer by writing it again
	os.WriteFile(filepath.Join(dir, "src.md"), []byte("# Updated Source"), 0644)

	code := runStale([]string{"--folio", yml})
	if code != 1 {
		t.Errorf("expected exit code 1 for stale targets, got %d", code)
	}
}

func TestRunStaleJSON(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\n"), 0644)

	code := runStale([]string{"--folio", yml, "--json"})
	if code != 0 {
		t.Errorf("expected exit code 0 for JSON mode with no targets, got %d", code)
	}
}
