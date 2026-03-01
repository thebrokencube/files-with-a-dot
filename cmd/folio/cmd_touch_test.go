package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunTouchNoArgs(t *testing.T) {
	code := runTouch([]string{})
	if code != 1 {
		t.Errorf("expected exit code 1 for no args, got %d", code)
	}
}

func TestRunTouchTargetNotFound(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\n"), 0644)

	code := runTouch([]string{"--folio", yml, "nonexistent"})
	if code != 1 {
		t.Errorf("expected exit code 1 for missing target, got %d", code)
	}
}

func TestRunTouchNoLocalOutput(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  my-target:
    transform: distill
    outputs:
      - external: jira
        id: "PROJ-123"
        field: description
`), 0644)

	code := runTouch([]string{"--folio", yml, "my-target"})
	if code != 1 {
		t.Errorf("expected exit code 1 for no local output, got %d", code)
	}
}

func TestRunTouchSuccess(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	outPath := filepath.Join(dir, "compiled", "out.md")
	os.WriteFile(outPath, []byte("# Output"), 0644)

	// Set mtime to the past
	past := time.Now().Add(-24 * time.Hour)
	os.Chtimes(outPath, past, past)

	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  my-target:
    transform: distill
    outputs:
      - path: compiled/out.md
`), 0644)

	before := time.Now()
	code := runTouch([]string{"--folio", yml, "my-target"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.ModTime().Before(before) {
		t.Error("mtime was not updated")
	}
}
