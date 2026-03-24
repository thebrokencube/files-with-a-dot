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
		var p struct {
			Issues []string `json:"issues"`
		}
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

// localHashOf computes the expected local hash from file content (with frontmatter).
func localHashOf(content string) string {
	return pipeline.ComputeLocalHash(pipeline.StripFrontmatter([]byte(content)))
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
			Kind: ActionPush, LocalContent: []byte("Content here"),
			LocalHash: localHashOf("---\njira: KEY-1\n---\nContent here")},
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
			Kind: ActionPush, LocalContent: []byte("Content A"),
			LocalHash: localHashOf("---\njira: KEY-1\n---\nContent A")},
		{Node: &forest.Node{Key: "KEY-2", File: "b.md", Label: "B"},
			Kind: ActionPush, LocalContent: []byte("Content B"),
			LocalHash: localHashOf("---\njira: KEY-2\n---\nContent B")},
		{Node: &forest.Node{Key: "KEY-3", File: "c.md", Label: "C"},
			Kind: ActionPush, LocalContent: []byte("Content C"),
			LocalHash: localHashOf("---\njira: KEY-3\n---\nContent C")},
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
			Kind: ActionPush, LocalContent: []byte("Content A"),
			LocalHash: localHashOf("---\njira: KEY-1\n---\nContent A")},
		{Node: &forest.Node{Key: "KEY-2", File: "b.md", Label: "B"},
			Kind: ActionPush, LocalContent: []byte("Content B"),
			LocalHash: localHashOf("---\njira: KEY-2\n---\nContent B")},
		{Node: &forest.Node{Key: "KEY-3", File: "c.md", Label: "C"},
			Kind: ActionPush, LocalContent: []byte("Content C"),
			LocalHash: localHashOf("---\njira: KEY-3\n---\nContent C")},
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
			Kind: ActionPush, LocalContent: []byte("Content A"),
			LocalHash: localHashOf("---\njira: KEY-1\n---\nContent A")},
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

func TestExecutePushRecordsPostPushRemoteHash(t *testing.T) {
	dir := setupExecuteForest(t, map[string]string{
		"a.md": "---\njira: KEY-1\n---\nContent A",
	})

	// The post-push ADF differs from the pre-push ADF.
	// This simulates: pre-push remote was old content, after push Jira has new ADF.
	prePushADF := `{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"old"}]}]}`
	postPushADF := `{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"Content A"}]}]}`

	runner := &pushRecordingRunner{
		viewResp: map[string][]byte{
			"KEY-1": []byte(fmt.Sprintf(`{"fields":{"description":%s}}`, postPushADF)),
		},
	}
	p := &pipeline.Pipeline{Run: runner.run}
	state := &forest.State{Nodes: make(map[string]forest.NodeState)}

	prePushHash := forest.ComputeHash([]byte(prePushADF))
	postPushHash := forest.ComputeHash([]byte(postPushADF))

	actions := []Action{
		{Node: &forest.Node{Key: "KEY-1", File: "a.md", Label: "A"},
			Kind: ActionPush, LocalContent: []byte("Content A"),
			LocalHash: localHashOf("---\njira: KEY-1\n---\nContent A"), RemoteHash: prePushHash},
	}

	results, err := Execute(actions, p, state, dir)
	if err != nil {
		t.Fatal(err)
	}
	if !results[0].Success {
		t.Fatalf("push should succeed: %v", results[0].Error)
	}

	// State should have the POST-push remote hash, not the pre-push one
	ns := state.Nodes["KEY-1"]
	if ns.RemoteHash == prePushHash {
		t.Error("state recorded pre-push remote hash — should have re-read after push")
	}
	if ns.RemoteHash != postPushHash {
		t.Errorf("state remote hash = %q, want post-push hash %q", ns.RemoteHash, postPushHash)
	}
}

func TestExecutePushFallbackOnReReadFailure(t *testing.T) {
	dir := setupExecuteForest(t, map[string]string{
		"a.md": "---\njira: KEY-1\n---\nContent A",
	})

	// Runner succeeds for push (edit) but has no view response — simulates re-read failure
	runner := &pushRecordingRunner{}
	p := &pipeline.Pipeline{Run: runner.run}
	state := &forest.State{Nodes: make(map[string]forest.NodeState)}

	actions := []Action{
		{Node: &forest.Node{Key: "KEY-1", File: "a.md", Label: "A"},
			Kind: ActionPush, LocalContent: []byte("Content A"),
			LocalHash: localHashOf("---\njira: KEY-1\n---\nContent A"), RemoteHash: "old-remote-hash"},
	}

	results, err := Execute(actions, p, state, dir)
	if err != nil {
		t.Fatal(err)
	}
	if !results[0].Success {
		t.Fatalf("push should succeed despite re-read failure: %v", results[0].Error)
	}

	// State should NOT have the pre-push hash if the fallback worked.
	// The fallback extracts ADF from the compiled payload — so the hash should
	// differ from "old-remote-hash" (the pre-push hash).
	ns := state.Nodes["KEY-1"]
	if ns.RemoteHash == "old-remote-hash" {
		t.Error("state recorded pre-push remote hash — fallback should have used compiled ADF hash")
	}
	if ns.RemoteHash == "" {
		t.Error("state remote hash should not be empty")
	}
}

func TestFetchRemoteHashInternal(t *testing.T) {
	adfContent := `{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":"hello"}]}]}`
	mock := &mockRunner{
		responses: map[string][]byte{
			"KEY-1": makeViewJSON(adfContent),
			"KEY-2": makeViewJSON(""), // null description
		},
		errors: map[string]error{
			"KEY-3": fmt.Errorf("connection refused"),
		},
	}
	p := &pipeline.Pipeline{Run: mock.run}

	// Successful fetch
	adf, hash, err := fetchRemoteHash("KEY-1", p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if adf == nil {
		t.Error("expected non-nil ADF")
	}
	expectedHash := forest.ComputeHash([]byte(adfContent))
	if hash != expectedHash {
		t.Errorf("hash = %q, want %q", hash, expectedHash)
	}

	// Null description
	adf2, hash2, err2 := fetchRemoteHash("KEY-2", p)
	if err2 != nil {
		t.Fatalf("unexpected error for null desc: %v", err2)
	}
	if adf2 != nil {
		t.Error("expected nil ADF for null description")
	}
	if hash2 != "" {
		t.Errorf("expected empty hash for null desc, got %q", hash2)
	}

	// Error
	_, _, err3 := fetchRemoteHash("KEY-3", p)
	if err3 == nil {
		t.Error("expected error for KEY-3")
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

func TestSimpleDiff(t *testing.T) {
	tests := []struct {
		name string
		from []string
		to   []string
		want []string
	}{
		{"identical", []string{"a", "b"}, []string{"a", "b"}, []string{" a", " b"}},
		{"addition", []string{"a"}, []string{"a", "b"}, []string{" a", "+b"}},
		{"deletion", []string{"a", "b"}, []string{"a"}, []string{" a", "-b"}},
		{"change", []string{"a", "old", "c"}, []string{"a", "new", "c"}, []string{" a", "-old", "+new", " c"}},
		{"empty_from", []string{}, []string{"a"}, []string{"+a"}},
		{"empty_to", []string{"a"}, []string{}, []string{"-a"}},
		{"both_empty", []string{}, []string{}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SimpleDiff(tt.from, tt.to)
			if len(got) != len(tt.want) {
				t.Fatalf("SimpleDiff len=%d, want %d\n  got:  %v\n  want: %v", len(got), len(tt.want), got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("line %d: got %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
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
