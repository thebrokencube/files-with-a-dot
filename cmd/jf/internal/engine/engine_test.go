package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
)

// mockRunner returns canned responses keyed by Jira issue key.
type mockRunner struct {
	responses map[string][]byte
	errors    map[string]error
}

func (m *mockRunner) run(name string, args ...string) ([]byte, error) {
	// Extract issue key from args (acli jira workitem view <KEY> ...)
	if len(args) >= 4 && args[0] == "jira" && args[1] == "workitem" && args[2] == "view" {
		key := args[3]
		if err, ok := m.errors[key]; ok {
			return nil, err
		}
		if resp, ok := m.responses[key]; ok {
			return resp, nil
		}
		return nil, fmt.Errorf("mock: no response for %s", key)
	}
	return nil, fmt.Errorf("mock: unexpected command %s %v", name, args)
}

func makeViewJSON(adf string) []byte {
	if adf == "" {
		return []byte(`{"fields":{"description":null}}`)
	}
	return []byte(fmt.Sprintf(`{"fields":{"description":%s}}`, adf))
}

func setupTestForest(t *testing.T, nodes []struct{ key, file, content string }) string {
	t.Helper()
	dir := t.TempDir()
	for _, n := range nodes {
		path := filepath.Join(dir, n.file)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(n.content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestReadParallelFetch(t *testing.T) {
	adfContent := `{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"text":"Hello"}]}]}`
	nodeDefs := []struct{ key, file, content string }{
		{"KEY-1", "a.md", "---\njira: KEY-1\n---\nContent 1"},
		{"KEY-2", "b.md", "---\njira: KEY-2\n---\nContent 2"},
		{"KEY-3", "c.md", "---\njira: KEY-3\n---\nContent 3"},
		{"KEY-4", "d.md", "---\njira: KEY-4\n---\nContent 4"},
		{"KEY-5", "e.md", "---\njira: KEY-5\n---\nContent 5"},
	}
	dir := setupTestForest(t, nodeDefs)

	mock := &mockRunner{
		responses: map[string][]byte{
			"KEY-1": makeViewJSON(adfContent),
			"KEY-2": makeViewJSON(adfContent),
			"KEY-3": makeViewJSON(adfContent),
			"KEY-5": makeViewJSON(""),
		},
		errors: map[string]error{
			"KEY-4": fmt.Errorf("connection timeout"),
		},
	}

	nodes := make([]*forest.Node, len(nodeDefs))
	for i, n := range nodeDefs {
		nodes[i] = &forest.Node{Key: n.key, File: n.file, Sync: "push"}
	}

	p := &pipeline.Pipeline{Run: mock.run}
	state := &forest.State{Nodes: make(map[string]forest.NodeState)}

	readings, err := Read(nodes, p, state, dir)
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if len(readings) != 5 {
		t.Fatalf("got %d readings, want 5", len(readings))
	}

	// KEY-1,2,3: should have remote ADF
	for i := 0; i < 3; i++ {
		if readings[i].RemoteADF == nil {
			t.Errorf("readings[%d] (%s): expected RemoteADF", i, readings[i].Node.Key)
		}
		if readings[i].RemoteErr != nil {
			t.Errorf("readings[%d] (%s): unexpected RemoteErr: %v", i, readings[i].Node.Key, readings[i].RemoteErr)
		}
	}

	// KEY-4: should have RemoteErr
	if readings[3].RemoteErr == nil {
		t.Error("readings[3] (KEY-4): expected RemoteErr")
	}

	// KEY-5: should have nil ADF (null description)
	if readings[4].RemoteADF != nil {
		t.Errorf("readings[4] (KEY-5): expected nil RemoteADF, got %s", string(readings[4].RemoteADF))
	}

	// All should have local content
	for i, r := range readings {
		if r.LocalContent == nil {
			t.Errorf("readings[%d]: expected LocalContent", i)
		}
		if r.LocalHash == "" {
			t.Errorf("readings[%d]: expected LocalHash", i)
		}
	}
}

func TestReadEmptyForest(t *testing.T) {
	readings, err := Read(nil, nil, nil, "")
	if err != nil {
		t.Fatalf("Read returned error: %v", err)
	}
	if readings != nil {
		t.Errorf("expected nil readings, got %d", len(readings))
	}
}

func TestReadBaselinePopulated(t *testing.T) {
	nodeDefs := []struct{ key, file, content string }{
		{"KEY-1", "a.md", "Content 1"},
		{"KEY-2", "b.md", "Content 2"},
		{"KEY-3", "c.md", "Content 3"},
	}
	dir := setupTestForest(t, nodeDefs)

	mock := &mockRunner{
		responses: map[string][]byte{
			"KEY-1": makeViewJSON(""),
			"KEY-2": makeViewJSON(""),
			"KEY-3": makeViewJSON(""),
		},
	}

	state := &forest.State{Nodes: map[string]forest.NodeState{
		"KEY-1": {LocalHash: "abc"},
		"KEY-3": {LocalHash: "def"},
	}}

	nodes := make([]*forest.Node, len(nodeDefs))
	for i, n := range nodeDefs {
		nodes[i] = &forest.Node{Key: n.key, File: n.file, Sync: "push"}
	}

	p := &pipeline.Pipeline{Run: mock.run}
	readings, err := Read(nodes, p, state, dir)
	if err != nil {
		t.Fatal(err)
	}

	if readings[0].Baseline == nil {
		t.Error("KEY-1: expected non-nil Baseline")
	}
	if readings[1].Baseline != nil {
		t.Error("KEY-2: expected nil Baseline")
	}
	if readings[2].Baseline == nil {
		t.Error("KEY-3: expected non-nil Baseline")
	}
}

func TestReadMoreNodesThanWorkers(t *testing.T) {
	nodeDefs := make([]struct{ key, file, content string }, 12)
	for i := range nodeDefs {
		key := fmt.Sprintf("KEY-%d", i+1)
		nodeDefs[i] = struct{ key, file, content string }{
			key, fmt.Sprintf("%d.md", i+1), fmt.Sprintf("Content %d", i+1),
		}
	}
	dir := setupTestForest(t, nodeDefs)

	responses := make(map[string][]byte)
	for _, n := range nodeDefs {
		responses[n.key] = makeViewJSON("")
	}
	mock := &mockRunner{responses: responses}

	nodes := make([]*forest.Node, len(nodeDefs))
	for i, n := range nodeDefs {
		nodes[i] = &forest.Node{Key: n.key, File: n.file, Sync: "push"}
	}

	p := &pipeline.Pipeline{Run: mock.run}
	readings, err := Read(nodes, p, &forest.State{Nodes: make(map[string]forest.NodeState)}, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(readings) != 12 {
		t.Fatalf("got %d readings, want 12", len(readings))
	}
}

func TestReadMajorityFailure(t *testing.T) {
	nodeDefs := []struct{ key, file, content string }{
		{"KEY-1", "a.md", "Content 1"},
		{"KEY-2", "b.md", "Content 2"},
		{"KEY-3", "c.md", "Content 3"},
		{"KEY-4", "d.md", "Content 4"},
		{"KEY-5", "e.md", "Content 5"},
	}
	dir := setupTestForest(t, nodeDefs)

	mock := &mockRunner{
		responses: map[string][]byte{
			"KEY-3": makeViewJSON(""),
		},
		errors: map[string]error{
			"KEY-1": fmt.Errorf("fail 1"),
			"KEY-2": fmt.Errorf("fail 2"),
			"KEY-4": fmt.Errorf("fail 4"),
			"KEY-5": fmt.Errorf("fail 5"),
		},
	}

	nodes := make([]*forest.Node, len(nodeDefs))
	for i, n := range nodeDefs {
		nodes[i] = &forest.Node{Key: n.key, File: n.file, Sync: "push"}
	}

	p := &pipeline.Pipeline{Run: mock.run}
	readings, err := Read(nodes, p, &forest.State{Nodes: make(map[string]forest.NodeState)}, dir)
	if err != nil {
		t.Fatalf("Read should not return top-level error on partial failure: %v", err)
	}
	if len(readings) != 5 {
		t.Fatalf("got %d readings, want 5", len(readings))
	}

	failCount := 0
	for _, r := range readings {
		if r.RemoteErr != nil {
			failCount++
		}
	}
	if failCount != 4 {
		t.Errorf("expected 4 failed readings, got %d", failCount)
	}
}

func TestReadSingleNode(t *testing.T) {
	nodeDefs := []struct{ key, file, content string }{
		{"KEY-1", "a.md", "Content 1"},
	}
	dir := setupTestForest(t, nodeDefs)

	mock := &mockRunner{
		responses: map[string][]byte{"KEY-1": makeViewJSON("")},
	}

	nodes := []*forest.Node{{Key: "KEY-1", File: "a.md", Sync: "push"}}
	p := &pipeline.Pipeline{Run: mock.run}
	readings, err := Read(nodes, p, &forest.State{Nodes: make(map[string]forest.NodeState)}, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(readings) != 1 {
		t.Fatalf("got %d readings, want 1", len(readings))
	}
}

