package main

import (
	"fmt"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSyncNoForest(t *testing.T) {
	dir := t.TempDir()
	code := runSync([]string{"--dir", dir})
	if code != 1 {
		t.Fatalf("expected exit 1 for missing forest, got %d", code)
	}
}

func TestRunSyncEmptyForest(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)

	// No jira: nodes — both push and pull find nothing → exit 0
	code := runSync([]string{"--dir", dir})
	if code != 0 {
		t.Fatalf("expected exit 0 for empty forest, got %d", code)
	}
}

func TestRunSyncTBDOnlyForest(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, "task.md"), []byte("---\njira: TBD\ntype: Task\n---\n# TBD Task\n"), 0644)

	// TBD nodes are filtered out of push and pull → exit 0
	code := runSync([]string{"--dir", dir})
	if code != 0 {
		t.Fatalf("expected exit 0 for TBD-only forest, got %d", code)
	}
}

func TestParseCompletenessResults(t *testing.T) {
	input := `[{"key":"BEN-100","fields":{"summary":"Child One","issuetype":{"name":"Story"}}},{"key":"BEN-101","fields":{"summary":"Child Two","issuetype":{"name":"Task"}}}]`
	results, err := parseCompletenessResults([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Key != "BEN-100" || results[0].Summary != "Child One" {
		t.Errorf("unexpected first result: %+v", results[0])
	}
	if results[1].Key != "BEN-101" || results[1].Type != "Task" {
		t.Errorf("unexpected second result: %+v", results[1])
	}
}

func TestParseCompletenessResultsEmpty(t *testing.T) {
	results, err := parseCompletenessResults([]byte("[]"))
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestCheckCompletenessAllPresent(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "parent"), 0755)
	os.WriteFile(filepath.Join(dir, "parent", "README.md"), []byte("---\njira: BEN-1\n---\n"), 0644)
	os.WriteFile(filepath.Join(dir, "parent", "BEN-2.md"), []byte("---\njira: BEN-2\n---\n"), 0644)

	f, roots, err := loadForest(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Mock pipeline that returns BEN-2 as the only child of BEN-1
	mockRunner := func(name string, args ...string) ([]byte, error) {
		for i, a := range args {
			if a == "--jql" && i+1 < len(args) && strings.Contains(args[i+1], "BEN-1") {
				return []byte(`[{"key":"BEN-2","fields":{"summary":"Existing Child","issuetype":{"name":"Story"}}}]`), nil
			}
		}
		return nil, fmt.Errorf("unexpected call: %s %v", name, args)
	}
	p := &pipeline.Pipeline{Run: mockRunner}

	count := checkCompleteness(f, roots, p, false, false)
	if count != 0 {
		t.Errorf("expected 0 new children, got %d", count)
	}
}

func TestCheckCompletenessFindsNew(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "parent"), 0755)
	os.WriteFile(filepath.Join(dir, "parent", "README.md"), []byte("---\njira: BEN-1\n---\n"), 0644)
	os.WriteFile(filepath.Join(dir, "parent", "BEN-2.md"), []byte("---\njira: BEN-2\n---\n"), 0644)

	f, roots, err := loadForest(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Mock: BEN-1 has children BEN-2 (known) and BEN-3 (new)
	mockRunner := func(name string, args ...string) ([]byte, error) {
		for i, a := range args {
			if a == "--jql" && i+1 < len(args) && strings.Contains(args[i+1], "BEN-1") {
				return []byte(`[{"key":"BEN-2","fields":{"summary":"Known","issuetype":{"name":"Story"}}},{"key":"BEN-3","fields":{"summary":"New Child","issuetype":{"name":"Task"}}}]`), nil
			}
		}
		return nil, fmt.Errorf("unexpected call")
	}
	p := &pipeline.Pipeline{Run: mockRunner}

	count := checkCompleteness(f, roots, p, false, false)
	if count != 1 {
		t.Errorf("expected 1 new child, got %d", count)
	}
}

func TestCheckCompletenessScaffold(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: both\n"), 0644)
	os.MkdirAll(filepath.Join(dir, "parent"), 0755)
	os.WriteFile(filepath.Join(dir, "parent", "README.md"), []byte("---\njira: BEN-1\n---\n"), 0644)
	os.WriteFile(filepath.Join(dir, "parent", "BEN-2.md"), []byte("---\njira: BEN-2\n---\n"), 0644)

	f, roots, err := loadForest(dir)
	if err != nil {
		t.Fatal(err)
	}

	mockRunner := func(name string, args ...string) ([]byte, error) {
		for i, a := range args {
			if a == "--jql" && i+1 < len(args) && strings.Contains(args[i+1], "BEN-1") {
				return []byte(`[{"key":"BEN-2","fields":{"summary":"Known","issuetype":{"name":"Story"}}},{"key":"BEN-3","fields":{"summary":"New Child","issuetype":{"name":"Task"}}}]`), nil
			}
		}
		return nil, fmt.Errorf("unexpected call")
	}
	p := &pipeline.Pipeline{Run: mockRunner}

	count := checkCompleteness(f, roots, p, false, true)
	if count != 1 {
		t.Errorf("expected 1 new child, got %d", count)
	}

	// Check stub file was created
	stubPath := filepath.Join(dir, "parent", "BEN-3.md")
	data, err := os.ReadFile(stubPath)
	if err != nil {
		t.Fatalf("expected stub file at %s: %v", stubPath, err)
	}
	if !strings.Contains(string(data), "jira: BEN-3") {
		t.Error("stub missing jira key")
	}
	if !strings.Contains(string(data), `label: "New Child"`) {
		t.Error("stub missing label")
	}
	if !strings.Contains(string(data), "sync: both") {
		t.Error("stub should inherit forest default sync mode")
	}
}

func TestScaffoldStub(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "sub"), 0755)
	os.WriteFile(filepath.Join(dir, "sub", "README.md"), []byte("---\njira: BEN-1\n---\n"), 0644)

	f := &forest.Forest{
		Dir:      dir,
		Defaults: forest.ForestDefaults{Sync: "push"},
	}
	parent := &forest.Node{Key: "BEN-1", File: "sub/README.md"}
	child := completenessChild{Key: "BEN-99", Summary: "Test Child", Type: "Story"}

	rel := scaffoldStub(f, parent, child)
	if rel == "" {
		t.Fatal("expected non-empty path")
	}

	data, err := os.ReadFile(filepath.Join(dir, rel))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "jira: BEN-99") {
		t.Error("missing jira key")
	}
}

func TestScaffoldStubNoOverwrite(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("---\njira: BEN-1\n---\n"), 0644)
	existing := filepath.Join(dir, "BEN-5.md")
	os.WriteFile(existing, []byte("---\njira: BEN-5\n---\n# Existing"), 0644)

	f := &forest.Forest{
		Dir:      dir,
		Defaults: forest.ForestDefaults{Sync: "push"},
	}
	parent := &forest.Node{Key: "BEN-1", File: "README.md"}
	child := completenessChild{Key: "BEN-5", Summary: "Duplicate", Type: "Story"}

	rel := scaffoldStub(f, parent, child)
	if rel != "" {
		t.Error("should not overwrite existing file")
	}

	// Verify original content unchanged
	data, err := os.ReadFile(existing)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "# Existing") {
		t.Error("existing file was modified")
	}
}
