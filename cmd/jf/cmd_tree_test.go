package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunTree(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("---\njira: TEST-1\ntype: Epic\n---\n# Root Epic\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "sub"), 0755)
	os.WriteFile(filepath.Join(dir, "sub", "child.md"), []byte("---\njira: TEST-2\n---\n# Child\n"), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runTree([]string{"-dir", dir})

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
	if !strings.Contains(output, "TEST-2") {
		t.Errorf("output missing TEST-2\nGot:\n%s", output)
	}
	if !strings.Contains(output, "2 nodes") {
		t.Errorf("output missing node count\nGot:\n%s", output)
	}
}

func TestRunTreeNoForest(t *testing.T) {
	dir := t.TempDir()
	code := runTree([]string{"-dir", dir})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestRunTreeEmpty(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)

	code := runTree([]string{"-dir", dir})
	if code != 0 {
		t.Fatalf("expected exit 0 for empty forest, got %d", code)
	}
}
