package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDiscover(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, "task-a.md"), []byte("---\njira: TEST-1\n---\n# Task A\n"), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runDiscover([]string{"-dir", dir})

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "TEST-1") {
		t.Errorf("output missing TEST-1\nGot:\n%s", output)
	}
	if !strings.Contains(output, "Task A") {
		t.Errorf("output missing label\nGot:\n%s", output)
	}
}

func TestRunDiscoverNoForest(t *testing.T) {
	dir := t.TempDir()
	code := runDiscover([]string{"-dir", dir})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestRunDiscoverJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, "task-a.md"), []byte("---\njira: TEST-1\n---\n# Task A\n"), 0644)
	os.WriteFile(filepath.Join(dir, "task-b.md"), []byte("---\njira: TEST-2\nsync: pull\n---\n# Task B\n"), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runDiscover([]string{"-dir", dir, "-json"})

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)

	var result []map[string]any
	if err := json.Unmarshal(buf[:n], &result); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}
}

func TestRunDiscoverEmpty(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)

	code := runDiscover([]string{"-dir", dir})
	if code != 0 {
		t.Fatalf("expected exit 0 for empty forest, got %d", code)
	}
}
