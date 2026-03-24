package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/engine"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func TestPushSingleFileNotFound(t *testing.T) {
	code := pushSingle("TEST-1", "/nonexistent/file.md", false)
	if code != dendrik.ExitUserError {
		t.Fatalf("expected exit %d for missing file, got %d", dendrik.ExitUserError, code)
	}
}

func TestPushForestNoForest(t *testing.T) {
	dir := t.TempDir()
	code := pushForest(filepath.Join(dir, ".jf"), nil, "", false, false, false, false)
	if code != dendrik.ExitUserError {
		t.Fatalf("expected exit %d for missing forest, got %d", dendrik.ExitUserError, code)
	}
}

func TestPushForestSubtreeNotFound(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task.md"), []byte("---\njira: TEST-1\n---\n# Task\n"), 0644)

	code := pushForest(filepath.Join(dir, ".jf"), nil, "NONEXISTENT", false, false, false, false)
	if code != dendrik.ExitUserError {
		t.Fatalf("expected exit %d for missing subtree target, got %d", dendrik.ExitUserError, code)
	}
}

func TestPushForestTBDSkipped(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task.md"), []byte("---\njira: TBD\n---\n# TBD Task\n"), 0644)

	code := pushForest(filepath.Join(dir, ".jf"), nil, "", false, false, false, false)
	if code != dendrik.ExitOK {
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
			// pushSingle reads the file and routes through engine — empty content
			// will be blocked by engine.Plan
			code := pushSingle("TEST-1", filePath, false)
			// Empty content is blocked by engine plan; the display shows blocked
			// and returns ExitOK (no failures, just blocked)
			if code != dendrik.ExitOK {
				t.Errorf("expected exit 0 for %s (blocked by engine), got %d", tt.name, code)
			}
		})
	}
}

func TestPushForestEmptyContentBlocked(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task.md"), []byte("---\njira: TEST-1\n---\n"), 0644)

	// Empty content gets blocked by engine plan; dry-run shows plan and exits OK
	code := pushForest(filepath.Join(dir, ".jf"), nil, "", false, true, false, false)
	if code != dendrik.ExitOK {
		t.Fatalf("expected exit 0 for dry-run with empty content, got %d", code)
	}
}

func TestBuildPlainTextPayload(t *testing.T) {
	source := []byte("---\njira: BEN-1\n---\n# Hello\n\nSome content")
	payload := engine.BuildPlainTextPayload("BEN-1", source)
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
