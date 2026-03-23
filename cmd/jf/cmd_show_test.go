package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunShow(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, "task-a.md"), []byte("---\njira: TEST-1\ntype: Epic\n---\n# My Epic\n"), 0644)

	// By key
	code := runShow([]string{"--dir", dir, "TEST-1"})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	// By filename stem
	code = runShow([]string{"--dir", dir, "task-a"})
	if code != 0 {
		t.Fatalf("expected exit 0 for stem match, got %d", code)
	}
}

func TestRunShowNotFound(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, "task-a.md"), []byte("---\njira: TEST-1\n---\n# Task\n"), 0644)

	code := runShow([]string{"--dir", dir, "NONEXISTENT"})
	if code != 1 {
		t.Fatalf("expected exit 1 for not-found, got %d", code)
	}
}

func TestRunShowNoTarget(t *testing.T) {
	code := runShow([]string{})
	if code != 1 {
		t.Fatalf("expected exit 1 for missing target, got %d", code)
	}
}

func TestRunShowWithChildren(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("---\njira: TEST-1\ntype: Epic\n---\n# Parent\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "sub"), 0755)
	os.WriteFile(filepath.Join(dir, "sub", "child.md"), []byte("---\njira: TEST-2\n---\n# Child\n"), 0644)

	// Show parent — should have 1 child
	code := runShow([]string{"--dir", dir, "TEST-1"})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

func TestRunShowPullNode(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, "pulled.md"), []byte("---\njira: TEST-5\nsync: pull\n---\n# Pulled\n"), 0644)

	code := runShow([]string{"--dir", dir, "TEST-5"})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

func TestRunShowOutputFormat(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("---\njira: TEST-1\ntype: Epic\n---\n# Parent Epic\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "sub"), 0755)
	os.WriteFile(filepath.Join(dir, "sub", "child.md"), []byte("---\njira: TEST-2\nsync: pull\n---\n# Child Story\n"), 0644)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runShow([]string{"--dir", dir, "TEST-1"})

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	// Verify all required fields appear in output
	for _, want := range []string{
		"Key:      TEST-1",
		"Label:    Parent Epic",
		"Type:     Epic",
		"Sync:     push",
		"File:     README.md",
		"Parent:   (root)",
		"Children: 1",
		"Status:   stale",
	} {
		if !contains(output, want) {
			t.Errorf("output missing %q\nGot:\n%s", want, output)
		}
	}
}

func TestRunShowTBDNode(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, "tbd.md"), []byte("---\njira: TBD\ntype: Epic\n---\n# Upcoming\n"), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runShow([]string{"--dir", dir, "tbd"})

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if !contains(output, "Status:   unknown") {
		t.Errorf("expected unknown status for TBD node\nGot:\n%s", output)
	}
}

func TestRunShowNoForest(t *testing.T) {
	dir := t.TempDir()
	code := runShow([]string{"--dir", dir, "TEST-1"})
	if code != 1 {
		t.Fatalf("expected exit 1 for no forest, got %d", code)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
