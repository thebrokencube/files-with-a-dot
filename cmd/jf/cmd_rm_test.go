package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunRmSuccess(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task-a.md"), []byte("---\njira: TEST-1\n---\n# Task A\n"), 0644)

	code := runRm([]string{"--dir", dir, "TEST-1"})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	if _, err := os.Stat(filepath.Join(dir, ".jf", "task-a.md")); !os.IsNotExist(err) {
		t.Fatal("expected file to be removed")
	}
}

func TestRunRmMultipleKeys(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "a.md"), []byte("---\njira: TEST-1\n---\n# A\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "b.md"), []byte("---\njira: TEST-2\n---\n# B\n"), 0644)

	code := runRm([]string{"--dir", dir, "TEST-1", "TEST-2"})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	for _, f := range []string{"a.md", "b.md"} {
		if _, err := os.Stat(filepath.Join(dir, ".jf", f)); !os.IsNotExist(err) {
			t.Fatalf("expected %s to be removed", f)
		}
	}
}

func TestRunRmNotFound(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task-a.md"), []byte("---\njira: TEST-1\n---\n# Task\n"), 0644)

	code := runRm([]string{"--dir", dir, "NONEXISTENT"})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestRunRmChildGuard(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "README.md"), []byte("---\njira: TEST-1\ntype: Epic\n---\n# Parent\n"), 0644)
	os.MkdirAll(filepath.Join(dir, ".jf", "sub"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "sub", "child.md"), []byte("---\njira: TEST-2\n---\n# Child\n"), 0644)

	code := runRm([]string{"--dir", dir, "TEST-1"})
	if code != 1 {
		t.Fatalf("expected exit 1 for child guard, got %d", code)
	}

	// Parent file should still exist
	if _, err := os.Stat(filepath.Join(dir, ".jf", "README.md")); os.IsNotExist(err) {
		t.Fatal("parent file should not have been removed")
	}
}

func TestRunRmNoArgs(t *testing.T) {
	code := runRm([]string{})
	if code != 1 {
		t.Fatalf("expected exit 1 for no args, got %d", code)
	}
}

func TestRunRmNoForest(t *testing.T) {
	dir := t.TempDir()
	code := runRm([]string{"--dir", dir, "TEST-1"})
	if code != 1 {
		t.Fatalf("expected exit 1 for no forest, got %d", code)
	}
}

func TestRunRmPartialFailure(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "a.md"), []byte("---\njira: TEST-1\n---\n# A\n"), 0644)

	// TEST-1 exists, NOPE does not — should return 1 but still remove TEST-1
	code := runRm([]string{"--dir", dir, "TEST-1", "NOPE"})
	if code != 1 {
		t.Fatalf("expected exit 1 for partial failure, got %d", code)
	}

	if _, err := os.Stat(filepath.Join(dir, ".jf", "a.md")); !os.IsNotExist(err) {
		t.Fatal("expected TEST-1 file to be removed despite partial failure")
	}
}
