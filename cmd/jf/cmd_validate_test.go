package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunValidate(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, "task-a.md"), []byte("---\njira: TEST-1\n---\n# Task A\n"), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runValidate([]string{"--dir", dir})

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !strings.Contains(output, "Forest valid") {
		t.Errorf("expected valid message\nGot:\n%s", output)
	}
}

func TestRunValidateNoForest(t *testing.T) {
	dir := t.TempDir()
	code := runValidate([]string{"--dir", dir})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestRunValidateWithErrors(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	// Duplicate keys
	os.WriteFile(filepath.Join(dir, "a.md"), []byte("---\njira: TEST-1\n---\n# A\n"), 0644)
	os.WriteFile(filepath.Join(dir, "b.md"), []byte("---\njira: TEST-1\n---\n# B\n"), 0644)

	code := runValidate([]string{"--dir", dir})
	if code != 1 {
		t.Fatalf("expected exit 1 for duplicate keys, got %d", code)
	}
}

func TestRunValidateJSON(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, "task-a.md"), []byte("---\njira: TEST-1\n---\n# Task A\n"), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runValidate([]string{"--dir", dir, "--json"})

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)

	var envelope struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(buf[:n], &envelope); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}
	if envelope.Data["valid"] != true {
		t.Errorf("expected valid=true, got %v", envelope.Data["valid"])
	}
}
