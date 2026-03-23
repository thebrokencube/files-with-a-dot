package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPushSingleFileNotFound(t *testing.T) {
	code := pushSingle("TEST-1", "/nonexistent/file.md", false)
	if code != 1 {
		t.Fatalf("expected exit 1 for missing file, got %d", code)
	}
}

func TestPushForestNoForest(t *testing.T) {
	dir := t.TempDir()
	code := pushForest(dir, nil, "", false, false, false, nil, nil, nil)
	if code != 1 {
		t.Fatalf("expected exit 1 for missing forest, got %d", code)
	}
}

func TestPushForestSubtreeNotFound(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, "task.md"), []byte("---\njira: TEST-1\n---\n# Task\n"), 0644)

	code := pushForest(dir, nil, "NONEXISTENT", false, false, false, nil, nil, nil)
	if code != 1 {
		t.Fatalf("expected exit 1 for missing subtree target, got %d", code)
	}
}

func TestPushForestTBDSkipped(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, "task.md"), []byte("---\njira: TBD\n---\n# TBD Task\n"), 0644)

	code := pushForest(dir, nil, "", false, false, false, nil, nil, nil)
	if code != 0 {
		t.Fatalf("expected exit 0 (no pushable nodes), got %d", code)
	}
}

func TestBuildPlainTextPayload(t *testing.T) {
	source := []byte("---\njira: BEN-1\n---\n# Hello\n\nSome content")
	payload := buildPlainTextPayload("BEN-1", source)
	s := string(payload)

	if !strings.Contains(s, `"BEN-1"`) {
		t.Error("payload missing key")
	}

	if strings.Contains(s, "jira:") {
		t.Error("payload should not contain frontmatter")
	}

	if !strings.Contains(s, "Hello") {
		t.Error("payload should contain heading text")
	}
}
