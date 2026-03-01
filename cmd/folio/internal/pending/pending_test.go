package pending

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendToEmptyList(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "folio.yml")
	os.WriteFile(path, []byte(`schema: 1
project: "Test"
pending: []
`), 0644)

	if err := Append(path, "new item"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, `- "new item"`) {
		t.Errorf("expected new item in pending, got:\n%s", content)
	}
	if strings.Contains(content, "pending: []") {
		t.Error("pending: [] should have been replaced with pending:")
	}
}

func TestAppendToExistingList(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "folio.yml")
	os.WriteFile(path, []byte(`schema: 1
project: "Test"
pending:
  - "existing item"
`), 0644)

	if err := Append(path, "new item"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, `- "existing item"`) {
		t.Error("existing item should be preserved")
	}
	if !strings.Contains(content, `- "new item"`) {
		t.Error("new item should be appended")
	}
}

func TestAppendWithComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "folio.yml")
	os.WriteFile(path, []byte(`schema: 1
project: "Test"
pending:
  # Category A
  - "item a"
  # Category B
  - "item b"
`), 0644)

	if err := Append(path, "item c"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	// New item should be after item b (the last pending entry)
	bIdx := strings.Index(content, `"item b"`)
	cIdx := strings.Index(content, `"item c"`)
	if cIdx < bIdx {
		t.Errorf("new item should be after item b:\n%s", content)
	}
}

func TestAppendNoPendingKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "folio.yml")
	os.WriteFile(path, []byte(`schema: 1
project: "Test"
`), 0644)

	err := Append(path, "item")
	if err == nil {
		t.Error("expected error for missing pending key")
	}
}

func TestAppendPreservesFollowingKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "folio.yml")
	os.WriteFile(path, []byte(`schema: 1
project: "Test"
pending:
  - "existing"
tasks:
  - "task 1"
`), 0644)

	if err := Append(path, "new item"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "tasks:") {
		t.Error("tasks key should be preserved")
	}
	if !strings.Contains(content, `"task 1"`) {
		t.Error("task item should be preserved")
	}
}
