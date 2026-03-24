package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func TestPullForestNoForest(t *testing.T) {
	dir := t.TempDir()
	code := pullForest(filepath.Join(dir, ".jf"), nil, "", false, false, false)
	if code != dendrik.ExitUserError {
		t.Fatalf("expected exit %d for missing forest, got %d", dendrik.ExitUserError, code)
	}
}

func TestPullForestTargetNotFound(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task.md"), []byte("---\njira: TEST-1\n---\n# Task\n"), 0644)

	code := pullForest(filepath.Join(dir, ".jf"), []string{"NONEXISTENT"}, "", false, false, false)
	if code != dendrik.ExitUserError {
		t.Fatalf("expected exit %d for missing target, got %d", dendrik.ExitUserError, code)
	}
}

func TestPullForestIncludesSyncBoth(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "both.md"), []byte("---\njira: TEST-1\nsync: both\n---\n# Both\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "push.md"), []byte("---\njira: TEST-2\nsync: push\n---\n# Push\n"), 0644)

	// dry-run mode: sync:both should appear in pull list
	code := pullForest(filepath.Join(dir, ".jf"), nil, "", true, false, false)
	if code != dendrik.ExitOK {
		t.Fatalf("expected exit 0 for dry-run, got %d", code)
	}
}

func TestPullForestNoPullNodes(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task.md"), []byte("---\njira: TEST-1\n---\n# Push Only\n"), 0644)

	code := pullForest(filepath.Join(dir, ".jf"), nil, "", false, false, false)
	if code != dendrik.ExitOK {
		t.Fatalf("expected exit 0 (no pull nodes), got %d", code)
	}
}
