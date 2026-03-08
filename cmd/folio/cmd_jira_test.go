package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAutoTouchFindsTreeTarget(t *testing.T) {
	dir := t.TempDir()

	// Create source file referenced by tree node
	srcDir := filepath.Join(dir, "sources")
	os.MkdirAll(srcDir, 0755)
	srcFile := filepath.Join(srcDir, "epic.md")
	os.WriteFile(srcFile, []byte("# Epic"), 0644)

	// Create output file with old mtime
	outDir := filepath.Join(dir, "compiled")
	os.MkdirAll(outDir, 0755)
	outFile := filepath.Join(outDir, "out.json")
	os.WriteFile(outFile, []byte("{}"), 0644)
	past := time.Now().Add(-24 * time.Hour)
	os.Chtimes(outFile, past, past)

	// Create folio.yml with a tree target referencing the source
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
sources: []
targets:
  my-tree:
    how: "Test tree"
    outputs:
      - path: compiled/out.json
    tree:
      system: jira
      compiled_ext: ".json"
      root:
        id: "PROJ-1"
        file: sources/epic.md
`), 0644)

	before := time.Now()
	touched, err := autoTouch(srcFile)
	if err != nil {
		t.Fatalf("autoTouch error: %v", err)
	}
	if touched != 1 {
		t.Errorf("expected 1 touched, got %d", touched)
	}

	info, _ := os.Stat(outFile)
	if info.ModTime().Before(before) {
		t.Error("output mtime was not updated")
	}
}

func TestAutoTouchNoFolioYml(t *testing.T) {
	dir := t.TempDir()
	srcFile := filepath.Join(dir, "orphan.md")
	os.WriteFile(srcFile, []byte("# Orphan"), 0644)

	_, err := autoTouch(srcFile)
	if err == nil {
		t.Error("expected error when no folio.yml exists")
	}
}

func TestAutoTouchNoMatchingTreeNode(t *testing.T) {
	dir := t.TempDir()
	srcFile := filepath.Join(dir, "unlinked.md")
	os.WriteFile(srcFile, []byte("# Unlinked"), 0644)

	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
sources: []
targets:
  my-tree:
    how: "Test tree"
    outputs:
      - path: out.md
    tree:
      system: jira
      root:
        id: "PROJ-1"
        file: other.md
`), 0644)

	_, err := autoTouch(srcFile)
	if err == nil {
		t.Error("expected error when no tree node matches source")
	}
}

func TestTreeNodeMatchesChild(t *testing.T) {
	dir := t.TempDir()

	// Create source file referenced by a child node
	srcFile := filepath.Join(dir, "child.md")
	os.WriteFile(srcFile, []byte("# Child"), 0644)

	outDir := filepath.Join(dir, "compiled")
	os.MkdirAll(outDir, 0755)
	outFile := filepath.Join(outDir, "out.json")
	os.WriteFile(outFile, []byte("{}"), 0644)
	past := time.Now().Add(-24 * time.Hour)
	os.Chtimes(outFile, past, past)

	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
sources: []
targets:
  my-tree:
    how: "Test tree"
    outputs:
      - path: compiled/out.json
    tree:
      system: jira
      root:
        id: "ROOT-1"
        children:
          - id: "CHILD-1"
            file: child.md
`), 0644)

	touched, err := autoTouch(srcFile)
	if err != nil {
		t.Fatalf("autoTouch error: %v", err)
	}
	if touched != 1 {
		t.Errorf("expected 1 touched, got %d", touched)
	}
}
