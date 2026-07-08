package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunList(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task-a.md"), []byte("---\njira: TEST-1\n---\n# Task A\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task-b.md"), []byte("---\njira: TEST-2\nsync: pull\n---\n# Task B\n"), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := buildRoot().Execute([]string{"list", "--dir", dir})

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
	if !strings.Contains(output, "pull") {
		t.Errorf("output missing pull sync mode\nGot:\n%s", output)
	}
}

func TestRunListNoForest(t *testing.T) {
	dir := t.TempDir()
	code := buildRoot().Execute([]string{"list", "--dir", dir})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestRunListJSON(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task-a.md"), []byte("---\njira: TEST-1\n---\n# Task A\n"), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := buildRoot().Execute([]string{"list", "--dir", dir, "--json"})

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)

	var envelope struct{ Data []map[string]any }
	if err := json.Unmarshal(buf[:n], &envelope); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if len(envelope.Data) != 1 {
		t.Fatalf("expected 1 item, got %d", len(envelope.Data))
	}
	if envelope.Data[0]["key"] != "TEST-1" {
		t.Errorf("expected key TEST-1, got %v", envelope.Data[0]["key"])
	}
}
