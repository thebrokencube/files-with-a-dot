package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunStatus(t *testing.T) {
	dir := t.TempDir()

	// forest.yml
	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)

	// push node (will be stale — no state recorded)
	os.WriteFile(filepath.Join(dir, "task-a.md"), []byte("---\njira: TEST-1\n---\n# Task A\n"), 0644)

	// pull node
	os.WriteFile(filepath.Join(dir, "task-b.md"), []byte("---\njira: TEST-2\nsync: pull\n---\n# Task B\n"), 0644)

	// TBD node
	os.WriteFile(filepath.Join(dir, "task-c.md"), []byte("---\njira: TBD\n---\n# Task C\n"), 0644)

	code := runStatus([]string{"-dir", dir})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

func TestRunStatusWithState(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, "task-a.md"), []byte("---\njira: TEST-1\n---\n# Task A\n"), 0644)
	os.WriteFile(filepath.Join(dir, "task-b.md"), []byte("---\njira: TEST-2\n---\n# Task B\n"), 0644)

	// Record push for TEST-1 in the future so it's not stale
	jfDir := filepath.Join(dir, ".jf")
	os.MkdirAll(jfDir, 0755)
	stateData := map[string]interface{}{
		"nodes": map[string]interface{}{
			"TEST-1": map[string]interface{}{
				"last_push": time.Now().Add(time.Hour).Format(time.RFC3339Nano),
			},
		},
	}
	data, _ := json.Marshal(stateData)
	os.WriteFile(filepath.Join(jfDir, "state.json"), data, 0644)

	code := runStatus([]string{"-dir", dir})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	// TEST-1 is clean (pushed in future), TEST-2 is stale (never pushed)
}

func TestRunStatusNoForest(t *testing.T) {
	dir := t.TempDir()
	code := runStatus([]string{"-dir", dir})
	if code != 1 {
		t.Fatalf("expected exit 1 for missing forest, got %d", code)
	}
}
