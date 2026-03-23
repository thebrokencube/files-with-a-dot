package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
)

// pushRecordingRunner records push calls and always succeeds.
type pushRecordingRunner struct {
	pushCalls []string // JSON payloads
	viewResp  map[string][]byte
	failKeys  map[string]bool
}

func (r *pushRecordingRunner) run(name string, args ...string) ([]byte, error) {
	if len(args) >= 4 && args[2] == "edit" {
		// Read the temp file payload
		payload, err := os.ReadFile(args[4])
		if err != nil {
			return nil, err
		}
		// Extract key from payload
		var p struct{ Issues []string `json:"issues"` }
		json.Unmarshal(payload, &p)
		if len(p.Issues) > 0 && r.failKeys[p.Issues[0]] {
			return nil, fmt.Errorf("mock push error for %s", p.Issues[0])
		}
		r.pushCalls = append(r.pushCalls, string(payload))
		return []byte("OK"), nil
	}
	if len(args) >= 4 && args[2] == "view" {
		key := args[3]
		if resp, ok := r.viewResp[key]; ok {
			return resp, nil
		}
		return nil, fmt.Errorf("mock: no view for %s", key)
	}
	return nil, fmt.Errorf("mock: unexpected %s %v", name, args)
}

func setupExecuteForest(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	// Create .jf directory for state
	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	for name, content := range files {
		path := filepath.Join(dir, name)
		os.MkdirAll(filepath.Dir(path), 0755)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestExecuteSkipAction(t *testing.T) {
	actions := []Action{
		{Node: &forest.Node{Key: "KEY-1"}, Kind: ActionSkip},
	}
	results, err := Execute(actions, nil, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if !results[0].Success {
		t.Error("skip should be successful")
	}
	if results[0].Kind != ActionSkip {
		t.Errorf("kind = %v, want skip", results[0].Kind)
	}
}

func TestExecuteBlockedNoTTY(t *testing.T) {
	// CI/tests run without TTY, so confirmOverride should return false
	actions := []Action{
		{Node: &forest.Node{Key: "KEY-1"}, Kind: ActionBlocked,
			Block: BlockFirstPush, Reason: "first sync"},
	}
	results, err := Execute(actions, nil, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if results[0].Success {
		t.Error("blocked action should not succeed without TTY")
	}
	if results[0].Kind != ActionBlocked {
		t.Errorf("kind = %v, want blocked", results[0].Kind)
	}
}

func TestExecuteBlockedConflictNoOverride(t *testing.T) {
	actions := []Action{
		{Node: &forest.Node{Key: "KEY-1"}, Kind: ActionBlocked,
			Block: BlockConflict, Reason: "both sides changed"},
	}
	results, err := Execute(actions, nil, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if results[0].Success {
		t.Error("conflict should not succeed")
	}
}

func TestExecuteTier3NeverCallsConfirmOverride(t *testing.T) {
	tier3Blocks := []BlockReason{BlockEmpty, BlockRemoteUnknown}
	for _, block := range tier3Blocks {
		actions := []Action{
			{Node: &forest.Node{Key: "KEY-1"}, Kind: ActionBlocked,
				Block: block, Reason: "tier 3"},
		}
		results, err := Execute(actions, nil, nil, "")
		if err != nil {
			t.Fatal(err)
		}
		if results[0].Success {
			t.Errorf("tier 3 block %v should not succeed", block)
		}
	}
}

func TestExecuteNilStateSkipsRecord(t *testing.T) {
	dir := setupExecuteForest(t, map[string]string{
		"a.md": "---\njira: KEY-1\n---\nContent here",
	})

	runner := &pushRecordingRunner{}
	p := &pipeline.Pipeline{Run: runner.run}

	actions := []Action{
		{Node: &forest.Node{Key: "KEY-1", File: "a.md", Label: "A"},
			Kind: ActionPush, LocalContent: []byte("Content here"), LocalHash: "lh"},
	}

	// nil state — should not panic
	results, err := Execute(actions, p, nil, dir)
	if err != nil {
		t.Fatal(err)
	}
	if !results[0].Success {
		t.Errorf("push should succeed with nil state: %v", results[0].Error)
	}
	if len(runner.pushCalls) != 1 {
		t.Errorf("expected 1 push call, got %d", len(runner.pushCalls))
	}
}

func TestExecutePerNodeStateSave(t *testing.T) {
	dir := setupExecuteForest(t, map[string]string{
		"a.md": "---\njira: KEY-1\n---\nContent A",
		"b.md": "---\njira: KEY-2\n---\nContent B",
		"c.md": "---\njira: KEY-3\n---\nContent C",
	})

	runner := &pushRecordingRunner{}
	p := &pipeline.Pipeline{Run: runner.run}
	state := &forest.State{Nodes: make(map[string]forest.NodeState)}

	actions := []Action{
		{Node: &forest.Node{Key: "KEY-1", File: "a.md", Label: "A"},
			Kind: ActionPush, LocalContent: []byte("Content A"), LocalHash: "lh1"},
		{Node: &forest.Node{Key: "KEY-2", File: "b.md", Label: "B"},
			Kind: ActionPush, LocalContent: []byte("Content B"), LocalHash: "lh2"},
		{Node: &forest.Node{Key: "KEY-3", File: "c.md", Label: "C"},
			Kind: ActionPush, LocalContent: []byte("Content C"), LocalHash: "lh3"},
	}

	results, err := Execute(actions, p, state, dir)
	if err != nil {
		t.Fatal(err)
	}

	for i, r := range results {
		if !r.Success {
			t.Errorf("result[%d] should succeed: %v", i, r.Error)
		}
	}

	// Verify state has all 3 nodes
	if len(state.Nodes) != 3 {
		t.Errorf("expected 3 state entries, got %d", len(state.Nodes))
	}

	// Verify state.json exists on disk
	statePath := filepath.Join(dir, ".jf", "state.json")
	data, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	var saved forest.State
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if len(saved.Nodes) != 3 {
		t.Errorf("saved state has %d nodes, want 3", len(saved.Nodes))
	}
	for _, key := range []string{"KEY-1", "KEY-2", "KEY-3"} {
		ns, ok := saved.Nodes[key]
		if !ok {
			t.Errorf("missing state for %s", key)
			continue
		}
		if ns.Direction != "push" {
			t.Errorf("%s direction = %q, want push", key, ns.Direction)
		}
	}
}

func TestExecutePartialFailurePreservesState(t *testing.T) {
	dir := setupExecuteForest(t, map[string]string{
		"a.md": "---\njira: KEY-1\n---\nContent A",
		"b.md": "---\njira: KEY-2\n---\nContent B",
		"c.md": "---\njira: KEY-3\n---\nContent C",
	})

	runner := &pushRecordingRunner{
		failKeys: map[string]bool{"KEY-2": true},
	}
	p := &pipeline.Pipeline{Run: runner.run}
	state := &forest.State{Nodes: make(map[string]forest.NodeState)}

	actions := []Action{
		{Node: &forest.Node{Key: "KEY-1", File: "a.md", Label: "A"},
			Kind: ActionPush, LocalContent: []byte("Content A"), LocalHash: "lh1"},
		{Node: &forest.Node{Key: "KEY-2", File: "b.md", Label: "B"},
			Kind: ActionPush, LocalContent: []byte("Content B"), LocalHash: "lh2"},
		{Node: &forest.Node{Key: "KEY-3", File: "c.md", Label: "C"},
			Kind: ActionPush, LocalContent: []byte("Content C"), LocalHash: "lh3"},
	}

	results, _ := Execute(actions, p, state, dir)

	if !results[0].Success {
		t.Error("KEY-1 should succeed")
	}
	if results[1].Success {
		t.Error("KEY-2 should fail")
	}
	if !results[2].Success {
		t.Error("KEY-3 should succeed")
	}

	// State should have KEY-1 and KEY-3 but not KEY-2
	if _, ok := state.Nodes["KEY-1"]; !ok {
		t.Error("missing state for KEY-1")
	}
	if _, ok := state.Nodes["KEY-2"]; ok {
		t.Error("KEY-2 should not have state")
	}
	if _, ok := state.Nodes["KEY-3"]; !ok {
		t.Error("missing state for KEY-3")
	}
}

func TestExecuteStateSaveFailureReturnsError(t *testing.T) {
	dir := setupExecuteForest(t, map[string]string{
		"a.md": "---\njira: KEY-1\n---\nContent A",
	})

	runner := &pushRecordingRunner{}
	p := &pipeline.Pipeline{Run: runner.run}
	state := &forest.State{Nodes: make(map[string]forest.NodeState)}

	// Make .jf directory read-only so SaveState fails
	jfDir := filepath.Join(dir, ".jf")
	os.Chmod(jfDir, 0444)
	defer os.Chmod(jfDir, 0755)

	actions := []Action{
		{Node: &forest.Node{Key: "KEY-1", File: "a.md", Label: "A"},
			Kind: ActionPush, LocalContent: []byte("Content A"), LocalHash: "lh1"},
	}

	results, err := Execute(actions, p, state, dir)
	if err == nil {
		t.Error("expected error about state save failure")
	}
	// Mutation should still have succeeded
	if !results[0].Success {
		t.Errorf("push should succeed despite state save failure: %v", results[0].Error)
	}
}

func TestOverridePromptPerBlockReason(t *testing.T) {
	tests := []struct {
		block    BlockReason
		contains string
	}{
		{BlockFirstPush, "PUSH local content, replacing the Jira description"},
		{BlockFirstPull, "PULL remote content, replacing your local file"},
		{BlockOverwrite, "PUSH local content, discarding those remote changes"},
	}

	for _, tt := range tests {
		a := Action{
			Node:  &forest.Node{Key: "KEY-1"},
			Block: tt.block,
		}
		prompt := overridePrompt(a)
		if !strings.Contains(prompt, tt.contains) {
			t.Errorf("block %v prompt missing %q:\n  got: %s", tt.block, tt.contains, prompt)
		}
	}
}

func TestPromoteAction(t *testing.T) {
	tests := []struct {
		name     string
		block    BlockReason
		wantKind ActionKind
	}{
		{"first_push_to_push", BlockFirstPush, ActionPush},
		{"overwrite_to_push", BlockOverwrite, ActionPush},
		{"first_pull_to_pull", BlockFirstPull, ActionPull},
		{"conflict_stays_blocked", BlockConflict, ActionBlocked},
		{"tier3_stays_blocked", BlockEmpty, ActionBlocked},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Action{
				Node:  &forest.Node{Key: "KEY-1"},
				Kind:  ActionBlocked,
				Block: tt.block,
			}
			promoted := promoteAction(a)
			if promoted.Kind != tt.wantKind {
				t.Errorf("promoteAction(%v) = %v, want %v", tt.block, promoted.Kind, tt.wantKind)
			}
		})
	}
}

