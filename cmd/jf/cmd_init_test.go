package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunInit(t *testing.T) {
	dir := t.TempDir()
	code := runInit([]string{"--dir", dir, "--project", "TEST"})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	content, err := os.ReadFile(filepath.Join(dir, ".jf", "forest.yml"))
	if err != nil {
		t.Fatalf(".jf/forest.yml not created: %s", err)
	}

	want := "schema: 1\n\ndefaults:\n  sync: both\n  type: Story\n  project: TEST\n"
	if string(content) != want {
		t.Errorf("content mismatch\n  got:  %q\n  want: %q", string(content), want)
	}
}

func TestRunInitAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	jfDir := filepath.Join(dir, ".jf")
	os.MkdirAll(jfDir, 0755)
	os.WriteFile(filepath.Join(jfDir, "forest.yml"), []byte("existing"), 0644)

	code := runInit([]string{"--dir", dir})
	if code != 0 {
		t.Fatalf("expected exit 0 for existing .jf/forest.yml, got %d", code)
	}

	// Verify original file not overwritten
	content, _ := os.ReadFile(filepath.Join(jfDir, "forest.yml"))
	if string(content) != "existing" {
		t.Errorf("forest.yml was overwritten, got %q", string(content))
	}
}

func TestRunInitDefaultProject(t *testing.T) {
	dir := t.TempDir()
	code := runInit([]string{"--dir", dir})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	content, _ := os.ReadFile(filepath.Join(dir, ".jf", "forest.yml"))
	want := "schema: 1\n\ndefaults:\n  sync: both\n  type: Story\n  project: BEN\n"
	if string(content) != want {
		t.Errorf("content mismatch\n  got:  %q\n  want: %q", string(content), want)
	}
}
