package list

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScan_Mixed(t *testing.T) {
	dir := setupTestHome(t, map[string]string{
		"active/ben/state-retirement/folio.yml": `schema: 1
project: "State Retirement Mandates"
targets:
  tech-spec:
    how: "Test"
    sources: []
  gdoc:
    how: "Test"
    sources: []
pending:
  - "item one"
`,
		"active/career-tracking/folio.yml": `schema: 1
project: "Career Tracking"
`,
		"archive/ben/pb-on-call/2026-02-20-ghost/folio.yml": `schema: 1
project: "HSA Ghost Policies"
`,
	})

	entries, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	// Should be sorted: active first, then alphabetical
	if entries[0].Section != "active" || entries[0].Path != "ben/state-retirement" {
		t.Errorf("entry 0: got %s/%s", entries[0].Section, entries[0].Path)
	}
	if entries[0].Targets != 2 {
		t.Errorf("entry 0: expected 2 targets, got %d", entries[0].Targets)
	}
	if entries[0].Pending != 1 {
		t.Errorf("entry 0: expected 1 pending, got %d", entries[0].Pending)
	}

	if entries[1].Section != "active" || entries[1].Path != "career-tracking" {
		t.Errorf("entry 1: got %s/%s", entries[1].Section, entries[1].Path)
	}

	if entries[2].Section != "archive" {
		t.Errorf("entry 2: expected archive, got %s", entries[2].Section)
	}
	if entries[2].Project != "HSA Ghost Policies" {
		t.Errorf("entry 2: expected 'HSA Ghost Policies', got %q", entries[2].Project)
	}
}

func TestScan_Empty(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "active"), 0755)
	os.MkdirAll(filepath.Join(dir, "archive"), 0755)

	entries, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestScan_SkipsMalformedYaml(t *testing.T) {
	dir := setupTestHome(t, map[string]string{
		"active/good/folio.yml": `schema: 1
project: "Good Project"
`,
		"active/bad/folio.yml": `not: valid: yaml: [[[`,
	})

	entries, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (skipping bad), got %d", len(entries))
	}
	if entries[0].Project != "Good Project" {
		t.Errorf("expected 'Good Project', got %q", entries[0].Project)
	}
}

func setupTestHome(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, "active"), 0755)
	os.MkdirAll(filepath.Join(dir, "archive"), 0755)

	for path, content := range files {
		p := filepath.Join(dir, path)
		os.MkdirAll(filepath.Dir(p), 0755)
		os.WriteFile(p, []byte(content), 0644)
	}

	return dir
}
