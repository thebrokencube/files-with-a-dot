package graph

import (
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

func TestBuildOutputMap(t *testing.T) {
	f := &config.Folio{
		Targets: map[string]config.Target{
			"summary": {
				Outputs: []config.Output{
					{Path: "compiled/summary.md"},
				},
			},
			"jira-update": {
				Outputs: []config.Output{
					{External: "jira", ID: "PROJ-123"},
					{Path: "compiled/jira.md"},
				},
			},
		},
	}

	m := BuildOutputMap(f)

	if len(m) != 3 {
		t.Fatalf("output map len = %d, want 3", len(m))
	}
	assertSingleProducer(t, m, "path:compiled/summary.md", "summary")
	assertSingleProducer(t, m, "path:compiled/jira.md", "jira-update")
	assertSingleProducer(t, m, "ext:jira:PROJ-123:", "jira-update")
}

func TestBuildOutputMapCollision(t *testing.T) {
	f := &config.Folio{
		Targets: map[string]config.Target{
			"first": {
				Outputs: []config.Output{{Path: "compiled/same.md"}},
			},
			"second": {
				Outputs: []config.Output{{Path: "compiled/same.md"}},
			},
		},
	}

	m := BuildOutputMap(f)
	producers := m["path:compiled/same.md"]
	if len(producers) != 2 {
		t.Errorf("expected 2 producers for collision, got %d", len(producers))
	}
}

func TestSingleProducerMap(t *testing.T) {
	outputMap := map[string][]string{
		"path:a.md": {"target-a"},
		"path:b.md": {"target-b1", "target-b2"}, // collision
	}
	single := SingleProducerMap(outputMap)
	if len(single) != 1 {
		t.Fatalf("single producer map len = %d, want 1", len(single))
	}
	if single["path:a.md"] != "target-a" {
		t.Errorf("unexpected producer for path:a.md")
	}
}

func TestInferEdges(t *testing.T) {
	f := &config.Folio{
		Targets: map[string]config.Target{
			"summary": {
				Sources: []config.Source{{Path: "README.md"}},
				Outputs: []config.Output{{Path: "compiled/summary.md"}},
			},
			"jira-update": {
				Sources: []config.Source{{Path: "compiled/summary.md"}},
				Outputs: []config.Output{{Path: "compiled/jira.md"}},
			},
		},
	}

	producerMap := map[string]string{
		"path:compiled/summary.md": "summary",
	}

	edges := InferEdges(f, producerMap)
	deps, ok := edges["jira-update"]
	if !ok {
		t.Fatal("expected edge from jira-update to summary")
	}
	if len(deps) != 1 || deps[0] != "summary" {
		t.Errorf("jira-update deps = %v, want [summary]", deps)
	}
}

func TestInferEdgesNoSelfEdge(t *testing.T) {
	f := &config.Folio{
		Targets: map[string]config.Target{
			"self": {
				Sources: []config.Source{{Path: "compiled/out.md"}},
				Outputs: []config.Output{{Path: "compiled/out.md"}},
			},
		},
	}
	producerMap := map[string]string{
		"path:compiled/out.md": "self",
	}
	edges := InferEdges(f, producerMap)
	if len(edges["self"]) != 0 {
		t.Errorf("expected no self-edge, got: %v", edges["self"])
	}
}

func TestMergeEdges(t *testing.T) {
	f := &config.Folio{
		Targets: map[string]config.Target{
			"a": {BlockedBy: []string{"b"}},
			"b": {},
			"c": {},
		},
	}
	inferred := map[string][]string{
		"a": {"c"},
	}
	merged := MergeEdges(f, inferred)
	deps := merged["a"]
	if len(deps) != 2 {
		t.Fatalf("merged deps for a = %v, want [b, c]", deps)
	}
	found := map[string]bool{}
	for _, d := range deps {
		found[d] = true
	}
	if !found["b"] || !found["c"] {
		t.Errorf("expected b and c in deps, got %v", deps)
	}
}

func TestMergeEdgesDedup(t *testing.T) {
	f := &config.Folio{
		Targets: map[string]config.Target{
			"a": {BlockedBy: []string{"b"}},
			"b": {},
		},
	}
	inferred := map[string][]string{
		"a": {"b"}, // same as explicit
	}
	merged := MergeEdges(f, inferred)
	if len(merged["a"]) != 1 {
		t.Errorf("expected deduped, got %v", merged["a"])
	}
}

func TestBuildOutputMapTree(t *testing.T) {
	f := &config.Folio{
		Targets: map[string]config.Target{
			"initiative": {
				Outputs: []config.Output{{Path: "compiled/manifest.md"}},
				Tree: &config.Tree{
					System: "jira",
					Root: config.TreeNode{
						ID:   "PROJ-1",
						File: "root.md",
						Children: []config.TreeNode{
							{ID: "PROJ-10", File: "child.md"},
							{ID: "PROJ-100", File: "grandchild.md"},
						},
					},
				},
			},
		},
	}

	m := BuildOutputMap(f)

	// Should have: path:compiled/manifest.md + ext:jira:PROJ-1: + ext:jira:PROJ-10: + ext:jira:PROJ-100:
	if len(m) != 4 {
		t.Fatalf("output map len = %d, want 4 (got keys: %v)", len(m), mapKeys(m))
	}
	assertSingleProducer(t, m, "path:compiled/manifest.md", "initiative")
	assertSingleProducer(t, m, "ext:jira:PROJ-1:", "initiative")
	assertSingleProducer(t, m, "ext:jira:PROJ-10:", "initiative")
	assertSingleProducer(t, m, "ext:jira:PROJ-100:", "initiative")
}

func TestInferEdgesFromBatchSources(t *testing.T) {
	f := &config.Folio{
		Targets: map[string]config.Target{
			"upstream": {
				Outputs: []config.Output{{Path: "compiled/summary.md"}},
			},
			"batch-target": {
				Outputs: []config.Output{{Path: "compiled/manifest.md"}},
				Batch: &config.Batch{
					Items: []config.BatchItem{
						{ID: "item-1", Source: "compiled/summary.md"},
					},
				},
			},
		},
	}
	producerMap := map[string]string{
		"path:compiled/summary.md": "upstream",
	}
	edges := InferEdges(f, producerMap)
	deps := edges["batch-target"]
	if len(deps) != 1 || deps[0] != "upstream" {
		t.Errorf("batch-target deps = %v, want [upstream]", deps)
	}
}

func TestInferEdgesFromTreeSources(t *testing.T) {
	f := &config.Folio{
		Targets: map[string]config.Target{
			"upstream": {
				Outputs: []config.Output{{Path: "compiled/summary.md"}},
			},
			"tree-target": {
				Outputs: []config.Output{{Path: "compiled/manifest.md"}},
				Tree: &config.Tree{
					System: "jira",
					Root: config.TreeNode{
						ID:   "ROOT-1",
						File: "root.md",
						Children: []config.TreeNode{
							{ID: "CHILD-1", File: "compiled/summary.md"},
						},
					},
				},
			},
		},
	}
	producerMap := map[string]string{
		"path:compiled/summary.md": "upstream",
	}
	edges := InferEdges(f, producerMap)
	deps := edges["tree-target"]
	if len(deps) != 1 || deps[0] != "upstream" {
		t.Errorf("tree-target deps = %v, want [upstream]", deps)
	}
}

func TestWalkTree(t *testing.T) {
	root := config.TreeNode{
		ID: "root",
		Children: []config.TreeNode{
			{ID: "a", Children: []config.TreeNode{
				{ID: "a1"},
			}},
			{ID: "b"},
		},
	}
	var visited []string
	WalkTree(&root, func(n *config.TreeNode) {
		visited = append(visited, n.ID)
	})
	expected := []string{"root", "a", "a1", "b"}
	if len(visited) != len(expected) {
		t.Fatalf("visited = %v, want %v", visited, expected)
	}
	for i, v := range visited {
		if v != expected[i] {
			t.Errorf("visited[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

func TestBuildOutputMapBatch(t *testing.T) {
	f := &config.Folio{
		Targets: map[string]config.Target{
			"batch-target": {
				Outputs: []config.Output{{Path: "compiled/manifest.md"}},
				Batch: &config.Batch{
					System: "gdocs",
					Items: []config.BatchItem{
						{ID: "tab-1", Output: config.Output{ID: "doc-tab-1"}},
						{ID: "tab-2", Output: config.Output{ID: "doc-tab-2"}},
					},
				},
			},
		},
	}

	m := BuildOutputMap(f)

	// path:compiled/manifest.md + ext:gdocs:doc-tab-1: + ext:gdocs:doc-tab-2:
	if len(m) != 3 {
		t.Fatalf("output map len = %d, want 3 (got keys: %v)", len(m), mapKeys(m))
	}
	assertSingleProducer(t, m, "path:compiled/manifest.md", "batch-target")
	assertSingleProducer(t, m, "ext:gdocs:doc-tab-1:", "batch-target")
	assertSingleProducer(t, m, "ext:gdocs:doc-tab-2:", "batch-target")
}

func TestBuildOutputMapBatchWithDefaults(t *testing.T) {
	f := &config.Folio{
		Targets: map[string]config.Target{
			"batch-target": {
				Outputs: []config.Output{{Path: "compiled/manifest.md"}},
				Batch: &config.Batch{
					System: "gdocs",
					Items: []config.BatchItem{
						{ID: "tab-1", Output: config.Output{ID: "doc-tab-1"}},
						// Item with explicit external override
						{ID: "tab-2", Output: config.Output{External: "jira", ID: "PROJ-99"}},
					},
				},
			},
		},
	}

	m := BuildOutputMap(f)

	// path:compiled/manifest.md + ext:gdocs:doc-tab-1: + ext:jira:PROJ-99:
	if len(m) != 3 {
		t.Fatalf("output map len = %d, want 3 (got keys: %v)", len(m), mapKeys(m))
	}
	assertSingleProducer(t, m, "ext:gdocs:doc-tab-1:", "batch-target")
	assertSingleProducer(t, m, "ext:jira:PROJ-99:", "batch-target")
}

func mapKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func assertSingleProducer(t *testing.T, m map[string][]string, key, expected string) {
	t.Helper()
	producers, ok := m[key]
	if !ok {
		t.Errorf("missing key %q in output map", key)
		return
	}
	if len(producers) != 1 || producers[0] != expected {
		t.Errorf("key %q: got %v, want [%s]", key, producers, expected)
	}
}
