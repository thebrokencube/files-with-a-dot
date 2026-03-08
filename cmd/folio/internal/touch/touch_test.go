package touch

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

func TestTargetTouchesOutputFiles(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	outPath := filepath.Join(dir, "compiled", "out.md")
	os.WriteFile(outPath, []byte("# Output"), 0644)

	past := time.Now().Add(-24 * time.Hour)
	os.Chtimes(outPath, past, past)

	target := &config.Target{
		Outputs: []config.Output{
			{Path: "compiled/out.md"},
		},
	}

	before := time.Now()
	touched, err := Target(dir, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if touched != 1 {
		t.Errorf("expected 1 touched, got %d", touched)
	}

	info, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.ModTime().Before(before) {
		t.Error("mtime was not updated")
	}
}

func TestTargetSkipsEmptyPaths(t *testing.T) {
	dir := t.TempDir()
	target := &config.Target{
		Outputs: []config.Output{
			{External: "jira", ID: "PROJ-123"},
		},
	}

	touched, err := Target(dir, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if touched != 0 {
		t.Errorf("expected 0 touched, got %d", touched)
	}
}

func TestTargetErrorOnMissingFile(t *testing.T) {
	dir := t.TempDir()
	target := &config.Target{
		Outputs: []config.Output{
			{Path: "nonexistent.md"},
		},
	}

	_, err := Target(dir, target)
	if err == nil {
		t.Error("expected error for missing file")
	}
}
