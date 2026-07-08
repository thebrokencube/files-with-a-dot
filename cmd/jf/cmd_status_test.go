package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	gh "github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/github"
)

func TestRunStatus(t *testing.T) {
	dir := t.TempDir()

	// forest.yml
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)

	// push node (will be stale — no state recorded)
	os.WriteFile(filepath.Join(dir, ".jf", "task-a.md"), []byte("---\njira: TEST-1\n---\n# Task A\n"), 0644)

	// pull node
	os.WriteFile(filepath.Join(dir, ".jf", "task-b.md"), []byte("---\njira: TEST-2\nsync: pull\n---\n# Task B\n"), 0644)

	// TBD node
	os.WriteFile(filepath.Join(dir, ".jf", "task-c.md"), []byte("---\njira: TBD\n---\n# Task C\n"), 0644)

	code := buildRoot().Execute([]string{"status", "--dir", dir})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

func TestRunStatusWithState(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task-a.md"), []byte("---\njira: TEST-1\n---\n# Task A\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task-b.md"), []byte("---\njira: TEST-2\n---\n# Task B\n"), 0644)

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

	code := buildRoot().Execute([]string{"status", "--dir", dir})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
	// TEST-1 is clean (pushed in future), TEST-2 is stale (never pushed)
}

func TestRunStatusCorruptState(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task.md"), []byte("---\njira: TEST-1\n---\n# Task\n"), 0644)

	// Write corrupt state — should degrade gracefully
	jfDir := filepath.Join(dir, ".jf")
	os.MkdirAll(jfDir, 0755)
	os.WriteFile(filepath.Join(jfDir, "state.json"), []byte("{bad json"), 0644)

	code := buildRoot().Execute([]string{"status", "--dir", dir})
	if code != 0 {
		t.Fatalf("expected exit 0 with corrupt state (graceful degradation), got %d", code)
	}
}

func TestRunStatusJSON(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task.md"), []byte("---\njira: TEST-1\n---\n# Task\n"), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := buildRoot().Execute([]string{"status", "--dir", dir, "--json"})

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
	if envelope.Data["total"] != float64(1) {
		t.Errorf("expected total=1, got %v", envelope.Data["total"])
	}
}

func TestRunStatusNoForest(t *testing.T) {
	dir := t.TempDir()
	code := buildRoot().Execute([]string{"status", "--dir", dir})
	if code != 1 {
		t.Fatalf("expected exit 1 for missing forest, got %d", code)
	}
}

func TestStatusWithPRBadges(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n  repos:\n    - Org/repo\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task-a.md"), []byte("---\njira: TEST-1\n---\n# Task A\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task-b.md"), []byte("---\njira: TEST-2\n---\n# Task B\n"), 0644)

	success := "SUCCESS"
	oldFetcher := prFetcher
	prFetcher = func(repos []string, keys []string) ([]gh.PR, error) {
		return []gh.PR{
			{
				Number:      42,
				State:       "OPEN",
				IsDraft:     true,
				HeadRefName: "test-1-fix-stuff",
				Title:       "fix stuff",
				StatusCheckRollup: []gh.StatusCheckRun{
					{Conclusion: &success, State: "COMPLETED"},
				},
			},
			{
				Number:      43,
				State:       "MERGED",
				HeadRefName: "test-2-feature",
				Title:       "[TEST-2] feature",
			},
		}, nil
	}
	t.Cleanup(func() { prFetcher = oldFetcher })

	// Capture JSON output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := buildRoot().Execute([]string{"status", "--dir", dir, "--json"})

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	buf := make([]byte, 8192)
	n, _ := r.Read(buf)

	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(buf[:n], &envelope); err != nil {
		t.Fatalf("invalid JSON: %v\nraw: %s", err, buf[:n])
	}

	var result map[string]any
	json.Unmarshal(envelope.Data, &result)

	// Verify repos field
	reposVal, ok := result["repos"].([]any)
	if !ok || len(reposVal) != 1 || reposVal[0] != "Org/repo" {
		t.Errorf("repos: got %v, want [Org/repo]", result["repos"])
	}

	// Verify PRs array exists
	prs, ok := result["prs"].([]any)
	if !ok || len(prs) != 2 {
		t.Fatalf("expected 2 PR badges, got %v", result["prs"])
	}

	// Verify first PR (branch match for TEST-1)
	pr0 := prs[0].(map[string]any)
	if pr0["key"] != "TEST-1" {
		t.Errorf("pr[0].key: got %v, want TEST-1", pr0["key"])
	}
	if pr0["state"] != "draft" {
		t.Errorf("pr[0].state: got %v, want draft", pr0["state"])
	}
	if pr0["ci"] != "pass" {
		t.Errorf("pr[0].ci: got %v, want pass", pr0["ci"])
	}

	// Verify second PR (title match for TEST-2)
	pr1 := prs[1].(map[string]any)
	if pr1["key"] != "TEST-2" {
		t.Errorf("pr[1].key: got %v, want TEST-2", pr1["key"])
	}
	if pr1["state"] != "merged" {
		t.Errorf("pr[1].state: got %v, want merged", pr1["state"])
	}

	// Verify state.json was NOT modified
	statePath := filepath.Join(dir, ".jf", "state.json")
	if _, err := os.Stat(statePath); err == nil {
		// state.json exists — check it wasn't modified by PR enrichment
		stateData, _ := os.ReadFile(statePath)
		var stateMap map[string]any
		json.Unmarshal(stateData, &stateMap)
		if _, hasPRs := stateMap["prs"]; hasPRs {
			t.Error("state.json should not contain PR data")
		}
	}
}

func TestStatusWithoutRepo(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	// No repo in defaults
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task.md"), []byte("---\njira: TEST-1\n---\n# Task\n"), 0644)

	fetcherCalled := false
	oldFetcher := prFetcher
	prFetcher = func(repos []string, keys []string) ([]gh.PR, error) {
		fetcherCalled = true
		return nil, nil
	}
	t.Cleanup(func() { prFetcher = oldFetcher })

	// Capture JSON output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := buildRoot().Execute([]string{"status", "--dir", dir, "--json"})

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	if fetcherCalled {
		t.Error("fetcher should not be called when repos is empty")
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)

	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	json.Unmarshal(buf[:n], &envelope)

	var result map[string]any
	json.Unmarshal(envelope.Data, &result)

	// No repos or prs fields in output
	if _, hasRepos := result["repos"]; hasRepos {
		t.Error("repos should not appear in output when not configured")
	}
	if _, hasPRs := result["prs"]; hasPRs {
		t.Error("prs should not appear in output when repos is not configured")
	}
}
