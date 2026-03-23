package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunSyncNoForest(t *testing.T) {
	dir := t.TempDir()
	code := runSync([]string{"--dir", dir})
	if code != 1 {
		t.Fatalf("expected exit 1 for missing forest, got %d", code)
	}
}

func TestRunSyncEmptyForest(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)

	// No jira: nodes — both push and pull find nothing → exit 0
	code := runSync([]string{"--dir", dir})
	if code != 0 {
		t.Fatalf("expected exit 0 for empty forest, got %d", code)
	}
}

func TestRunSyncTBDOnlyForest(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, "task.md"), []byte("---\njira: TBD\ntype: Task\n---\n# TBD Task\n"), 0644)

	// TBD nodes are filtered out of push and pull → exit 0
	code := runSync([]string{"--dir", dir})
	if code != 0 {
		t.Fatalf("expected exit 0 for TBD-only forest, got %d", code)
	}
}
