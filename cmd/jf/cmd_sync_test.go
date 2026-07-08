package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/engine"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func TestRunSyncNoForest(t *testing.T) {
	dir := t.TempDir()
	code := runSync(dir, "", false, false, false, false, false)
	if code != dendrik.ExitUserError {
		t.Fatalf("expected exit %d for missing forest, got %d", dendrik.ExitUserError, code)
	}
}

func TestRunSyncEmptyForest(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)

	code := runSync(dir, "", true, false, false, false, false)
	if code != dendrik.ExitOK {
		t.Fatalf("expected exit 0 for empty forest, got %d", code)
	}
}

func TestRunSyncTBDOnlyForest(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task.md"), []byte("---\njira: TBD\ntype: Task\n---\n# TBD Task\n"), 0644)

	code := runSync(dir, "", true, false, false, false, false)
	if code != dendrik.ExitOK {
		t.Fatalf("expected exit 0 for TBD-only forest, got %d", code)
	}
}

func TestSyncDryRunShowsPlan(t *testing.T) {
	dir := setupSyncTestForest(t)

	// Dry-run shows plan without executing
	code := runSync(dir, "", true, false, false, false, false)
	if code != dendrik.ExitOK {
		t.Fatalf("expected exit 0 for dry-run, got %d", code)
	}
}

func TestSyncDryRunJson(t *testing.T) {
	dir := setupSyncTestForest(t)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runSync(dir, "", false, false, false, true, false)

	w.Close()
	os.Stdout = old

	if code != dendrik.ExitOK {
		t.Fatalf("expected exit 0 for --json, got %d", code)
	}

	var buf [64 * 1024]byte
	n, _ := r.Read(buf[:])
	output := string(buf[:n])

	// Parse the JSON envelope
	var envelope dendrik.ResultEnvelope
	if err := json.Unmarshal([]byte(output), &envelope); err != nil {
		t.Fatalf("invalid JSON output: %s\nraw: %s", err, output)
	}

	// Re-marshal the data to inspect
	data, _ := json.Marshal(envelope.Data)
	var plan planJSON
	if err := json.Unmarshal(data, &plan); err != nil {
		t.Fatalf("invalid plan JSON: %s", err)
	}

	if len(plan.Plan) == 0 {
		t.Error("expected non-empty plan")
	}

	// Check that each entry has required fields
	for _, entry := range plan.Plan {
		if entry.Action == "" {
			t.Error("plan entry missing action")
		}
		if entry.Key == "" {
			t.Error("plan entry missing key")
		}
		if entry.File == "" {
			t.Error("plan entry missing file")
		}
	}
}

func TestSyncDryRunJsonBlockedHasTier(t *testing.T) {
	actions := []engine.Action{
		{Node: &forest.Node{Key: "TEST-1", File: "a.md"}, Kind: engine.ActionBlocked,
			Block: engine.BlockEmpty, Reason: "empty local content"},
		{Node: &forest.Node{Key: "TEST-2", File: "b.md"}, Kind: engine.ActionBlocked,
			Block: engine.BlockFirstPush, Reason: "first sync — remote has content"},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	writePlanJSON(actions)

	w.Close()
	os.Stdout = old

	var buf [64 * 1024]byte
	n, _ := r.Read(buf[:])
	output := string(buf[:n])

	var envelope dendrik.ResultEnvelope
	json.Unmarshal([]byte(output), &envelope)
	data, _ := json.Marshal(envelope.Data)
	var plan planJSON
	json.Unmarshal(data, &plan)

	for _, entry := range plan.Plan {
		if entry.Tier == 0 {
			t.Errorf("blocked entry %s missing tier", entry.Key)
		}
		if entry.Hint == "" {
			t.Errorf("blocked entry %s missing hint", entry.Key)
		}
	}
}

func TestSyncWithResolveLocal(t *testing.T) {
	reading := engine.NodeReading{
		Node:         &forest.Node{Key: "TEST-1", File: "a.md", Sync: "both"},
		LocalContent: []byte("local content here"),
		LocalHash:    "abc",
		RemoteADF:    json.RawMessage(`{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"remote"}]}]}`),
		RemoteHash:   "def",
		Baseline:     &forest.NodeState{LocalHash: "old-local", RemoteHash: "old-remote"},
		Mutable:      true,
	}

	plan := engine.Plan([]engine.NodeReading{reading}, engine.PlanOpts{Direction: "both", Resolve: "local"})
	if len(plan) != 1 {
		t.Fatalf("expected 1 action, got %d", len(plan))
	}
	if plan[0].Kind != engine.ActionPush {
		t.Errorf("expected push (resolve local), got %s", plan[0].Kind)
	}
}

func TestSyncWithResolveRemote(t *testing.T) {
	reading := engine.NodeReading{
		Node:         &forest.Node{Key: "TEST-1", File: "a.md", Sync: "both"},
		LocalContent: []byte("local content here"),
		LocalHash:    "abc",
		RemoteADF:    json.RawMessage(`{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"remote"}]}]}`),
		RemoteHash:   "def",
		Baseline:     &forest.NodeState{LocalHash: "old-local", RemoteHash: "old-remote"},
		Mutable:      true,
	}

	plan := engine.Plan([]engine.NodeReading{reading}, engine.PlanOpts{Direction: "both", Resolve: "remote"})
	if len(plan) != 1 {
		t.Fatalf("expected 1 action, got %d", len(plan))
	}
	if plan[0].Kind != engine.ActionPull {
		t.Errorf("expected pull (resolve remote), got %s", plan[0].Kind)
	}
}

func TestSyncEmptyContentNeverPushes(t *testing.T) {
	// Empty stub -> Plan -> ActionBlocked -> no push
	reading := engine.NodeReading{
		Node:         &forest.Node{Key: "TEST-1", File: "stub.md", Sync: "push"},
		LocalContent: []byte(""),
		LocalHash:    pipeline.ComputeLocalHash([]byte("")),
	}

	plan := engine.Plan([]engine.NodeReading{reading}, engine.PlanOpts{Direction: "push"})
	if len(plan) != 1 {
		t.Fatalf("expected 1 action, got %d", len(plan))
	}
	if plan[0].Kind != engine.ActionBlocked {
		t.Errorf("expected blocked for empty content, got %s", plan[0].Kind)
	}
	if plan[0].Block != engine.BlockEmpty {
		t.Errorf("expected BlockEmpty, got %s", plan[0].Block)
	}
}

func TestSyncBatchNoTTYNoYes(t *testing.T) {
	actions := []engine.Action{
		{Node: &forest.Node{Key: "A"}, Kind: engine.ActionPush},
		{Node: &forest.Node{Key: "B"}, Kind: engine.ActionPush},
	}
	if !isBatch(actions) {
		t.Fatal("expected isBatch=true for 2 push actions")
	}
}

func TestSyncBatchTTYNoYes(t *testing.T) {
	// Single mutation is never a batch
	actions := []engine.Action{
		{Node: &forest.Node{Key: "A"}, Kind: engine.ActionPush},
		{Node: &forest.Node{Key: "B"}, Kind: engine.ActionSkip},
	}
	if isBatch(actions) {
		t.Fatal("expected isBatch=false for 1 push + 1 skip")
	}
}

func TestSyncSingleNodeNoGate(t *testing.T) {
	actions := []engine.Action{
		{Node: &forest.Node{Key: "A"}, Kind: engine.ActionPush},
	}
	if isBatch(actions) {
		t.Fatal("expected isBatch=false for single push")
	}
}

func TestDisplayPlanSortOrder(t *testing.T) {
	actions := []engine.Action{
		{Node: &forest.Node{Key: "A", File: "a.md"}, Kind: engine.ActionSkip, Reason: "no changes"},
		{Node: &forest.Node{Key: "B", File: "b.md"}, Kind: engine.ActionPush, Reason: "local changed"},
		{Node: &forest.Node{Key: "C", File: "c.md"}, Kind: engine.ActionBlocked, Block: engine.BlockEmpty, Reason: "empty"},
		{Node: &forest.Node{Key: "D", File: "d.md"}, Kind: engine.ActionPull, Reason: "remote changed"},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	displayPlan(actions)

	w.Close()
	os.Stdout = old

	var buf [4096]byte
	n, _ := r.Read(buf[:])
	output := string(buf[:n])
	lines := strings.Split(output, "\n")

	// Find the action lines (skip header and summary)
	var actionLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "BLOCKED") || strings.HasPrefix(trimmed, "PUSH") ||
			strings.HasPrefix(trimmed, "PULL") || strings.HasPrefix(trimmed, "SKIP") {
			actionLines = append(actionLines, trimmed)
		}
	}

	if len(actionLines) != 4 {
		t.Fatalf("expected 4 action lines, got %d: %v", len(actionLines), actionLines)
	}

	// BLOCKED should be first
	if !strings.HasPrefix(actionLines[0], "BLOCKED") {
		t.Errorf("first action should be BLOCKED, got: %s", actionLines[0])
	}
	// SKIP should be last
	if !strings.HasPrefix(actionLines[3], "SKIP") {
		t.Errorf("last action should be SKIP, got: %s", actionLines[3])
	}
}

func TestBlockHints(t *testing.T) {
	tests := []struct {
		block engine.BlockReason
		want  string
	}{
		{engine.BlockEmpty, "empty content — no override"},
		{engine.BlockRemoteUnknown, "remote unreachable — no override"},
		{engine.BlockFirstPush, "first sync, remote has content — resolve in terminal"},
		{engine.BlockFirstPull, "first sync, local has content — resolve in terminal"},
		{engine.BlockOverwrite, "remote changed — resolve in terminal"},
		{engine.BlockConflict, "conflict — use --resolve local|remote"},
	}

	for _, tt := range tests {
		a := engine.Action{Block: tt.block}
		got := blockHint(a)
		if got != tt.want {
			t.Errorf("blockHint(%s) = %q, want %q", tt.block, got, tt.want)
		}
	}
}

// --- Completeness tests (unchanged) ---

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
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n"), 0644)
	os.MkdirAll(filepath.Join(dir, ".jf", "parent"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "parent", "README.md"), []byte("---\njira: BEN-1\n---\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "parent", "BEN-2.md"), []byte("---\njira: BEN-2\n---\n"), 0644)

	f, roots, err := loadForest(filepath.Join(dir, ".jf"))
	if err != nil {
		t.Fatal(err)
	}

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
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n"), 0644)
	os.MkdirAll(filepath.Join(dir, ".jf", "parent"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "parent", "README.md"), []byte("---\njira: BEN-1\n---\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "parent", "BEN-2.md"), []byte("---\njira: BEN-2\n---\n"), 0644)

	f, roots, err := loadForest(filepath.Join(dir, ".jf"))
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

	count := checkCompleteness(f, roots, p, false, false)
	if count != 1 {
		t.Errorf("expected 1 new child, got %d", count)
	}
}

func TestCheckCompletenessScaffold(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: both\n"), 0644)
	os.MkdirAll(filepath.Join(dir, ".jf", "parent"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "parent", "README.md"), []byte("---\njira: BEN-1\n---\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "parent", "BEN-2.md"), []byte("---\njira: BEN-2\n---\n"), 0644)

	f, roots, err := loadForest(filepath.Join(dir, ".jf"))
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

	stubPath := filepath.Join(dir, ".jf", "parent", "BEN-3.md")
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

	data, err := os.ReadFile(existing)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "# Existing") {
		t.Error("existing file was modified")
	}
}

// --- Helpers ---

// setupSyncTestForest creates a minimal forest with a mock-compatible runner.
// The pipeline uses DefaultRunner, so this forest is for dry-run/json tests only
// (no actual Jira calls). Nodes have real keys to exercise engine.Read.
func setupSyncTestForest(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task.md"), []byte("---\njira: TBD\ntype: Task\n---\n# TBD Task\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "empty.md"), []byte("---\njira: TEST-1\n---\n"), 0644)
	return dir
}
