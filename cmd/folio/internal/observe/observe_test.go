package observe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendToEmptyList(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "folio.yml")
	os.WriteFile(path, []byte(`schema: 2
project: "Test"
observations: []
`), 0644)

	if err := Append(path, "new item"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, `- "new item"`) {
		t.Errorf("expected new item in observations, got:\n%s", content)
	}
	if strings.Contains(content, "observations: []") {
		t.Error("observations: [] should have been replaced with observations:")
	}
}

func TestAppendToExistingList(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "folio.yml")
	os.WriteFile(path, []byte(`schema: 2
project: "Test"
observations:
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
	os.WriteFile(path, []byte(`schema: 2
project: "Test"
observations:
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
	// New item should be after item b (the last entry)
	bIdx := strings.Index(content, `"item b"`)
	cIdx := strings.Index(content, `"item c"`)
	if cIdx < bIdx {
		t.Errorf("new item should be after item b:\n%s", content)
	}
}

func TestAppendNoObservationsKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "folio.yml")
	os.WriteFile(path, []byte(`schema: 2
project: "Test"
`), 0644)

	err := Append(path, "item")
	if err == nil {
		t.Error("expected error for missing observations key")
	}
}

func TestAppendToObservations(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "folio.yml")
	os.WriteFile(path, []byte(`schema: 2
project: "Test"
observations:
  - "existing obs"
`), 0644)

	if err := Append(path, "new obs"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, `"new obs"`) {
		t.Errorf("expected new obs in observations, got:\n%s", content)
	}
}

func TestAppendPreservesFollowingKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "folio.yml")
	os.WriteFile(path, []byte(`schema: 2
project: "Test"
observations:
  - "existing"
cross_references:
  - fact: "test"
`), 0644)

	if err := Append(path, "new item"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "cross_references:") {
		t.Error("cross_references key should be preserved")
	}
	if !strings.Contains(content, `"test"`) {
		t.Error("cross_references item should be preserved")
	}
}
