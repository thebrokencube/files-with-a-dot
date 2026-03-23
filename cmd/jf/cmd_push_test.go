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

func TestPushEmptyContentBlocked(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name    string
		content string
	}{
		{"frontmatter-only", "---\njira: TEST-1\n---\n"},
		{"whitespace-only", "---\njira: TEST-1\n---\n  \n\n  \n"},
		{"empty body", "---\njira: TEST-1\n---\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(dir, tt.name+".md")
			os.WriteFile(filePath, []byte(tt.content), 0644)
			code := pushSingle("TEST-1", filePath, false)
			if code != 1 {
				t.Errorf("expected exit 1 for %s, got %d", tt.name, code)
			}
		})
	}
}

func TestPushForestEmptyContentBlocked(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, "task.md"), []byte("---\njira: TEST-1\n---\n"), 0644)

	// Non-dry-run: empty content blocks the push (failed count > 0)
	code := pushForest(dir, nil, "", false, false, false, nil, nil, nil)
	if code != 1 {
		t.Fatalf("expected exit 1 for empty content, got %d", code)
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
