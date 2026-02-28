package status

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

func TestDeriveLocalStatusClean(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "source.md")
	outPath := filepath.Join(dir, "compiled", "output.md")

	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(srcPath, []byte("source"), 0644)
	time.Sleep(50 * time.Millisecond)
	os.WriteFile(outPath, []byte("output"), 0644)

	status := DeriveLocalStatus(dir, "compiled/output.md", []string{"source.md"})
	if status != "clean" {
		t.Errorf("status = %q, want clean", status)
	}
}

func TestDeriveLocalStatusStale(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "compiled", "output.md")
	srcPath := filepath.Join(dir, "source.md")

	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(outPath, []byte("output"), 0644)
	time.Sleep(50 * time.Millisecond)
	os.WriteFile(srcPath, []byte("source newer"), 0644)

	status := DeriveLocalStatus(dir, "compiled/output.md", []string{"source.md"})
	if status != "stale" {
		t.Errorf("status = %q, want stale", status)
	}
}

func TestDeriveLocalStatusMissing(t *testing.T) {
	dir := t.TempDir()
	status := DeriveLocalStatus(dir, "compiled/nonexistent.md", []string{"source.md"})
	if status != "missing" {
		t.Errorf("status = %q, want missing", status)
	}
}

func TestDeriveLocalStatusMissingSource(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "compiled", "output.md"), []byte("output"), 0644)

	status := DeriveLocalStatus(dir, "compiled/output.md", []string{"nonexistent.md"})
	if status != "stale" {
		t.Errorf("status = %q, want stale (source missing)", status)
	}
}

func TestClassifySourcePrimary(t *testing.T) {
	src := config.Source{Path: "README.md"}
	info := ClassifySource(src)
	if info.Kind != "primary" {
		t.Errorf("kind = %q, want primary", info.Kind)
	}
	if info.Label != "README.md" {
		t.Errorf("label = %q, want README.md", info.Label)
	}
}

func TestClassifySourceExternal(t *testing.T) {
	src := config.Source{External: "jira", ID: "ACME-123"}
	info := ClassifySource(src)
	if info.Kind != "external" {
		t.Errorf("kind = %q, want external", info.Kind)
	}
	if info.Label != "jira ACME-123" {
		t.Errorf("label = %q, want 'jira ACME-123'", info.Label)
	}
}

func TestClassifySourceCode(t *testing.T) {
	src := config.Source{External: "github", ID: "Org/repo"}
	info := ClassifySource(src)
	if info.Kind != "code" {
		t.Errorf("kind = %q, want code", info.Kind)
	}
}

func TestClassifySourceDerived(t *testing.T) {
	src := config.Source{
		Path: "cached.md",
		DerivedFrom: []config.DerivedFrom{
			{
				External: "web",
				Cached:   "2026-01-01",
			},
		},
	}
	info := ClassifySource(src)
	if info.Kind != "derived" {
		t.Errorf("kind = %q, want derived", info.Kind)
	}
	// Label should contain path and cache age
	if info.Label == "" {
		t.Error("label is empty")
	}
}

func TestClassifySourceUnknown(t *testing.T) {
	src := config.Source{}
	info := ClassifySource(src)
	if info.Kind != "unknown" {
		t.Errorf("kind = %q, want unknown", info.Kind)
	}
}

func TestDeriveFullProject(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test"), 0644)
	time.Sleep(50 * time.Millisecond)
	os.WriteFile(filepath.Join(dir, "compiled", "out.md"), []byte("compiled"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Test Project",
		Sources: []config.Source{{Path: "README.md"}},
		Targets: map[string]config.Target{
			"summary": {
				Sources: []config.Source{{Path: "README.md"}},
				Outputs: []config.Output{{Path: "compiled/out.md"}},
			},
		},
		Tasks:   []string{"task1"},
		Pending: []string{"note1", "note2"},
	}

	ps := Derive(f, dir)
	if ps.Project != "Test Project" {
		t.Errorf("project = %q", ps.Project)
	}
	if len(ps.Sources) != 1 {
		t.Errorf("sources len = %d", len(ps.Sources))
	}
	if ps.Tasks != 1 {
		t.Errorf("tasks = %d, want 1", ps.Tasks)
	}
	if ps.Pending != 2 {
		t.Errorf("pending = %d, want 2", ps.Pending)
	}

	ts := ps.Targets["summary"]
	if len(ts.Outputs) != 1 {
		t.Fatalf("outputs len = %d", len(ts.Outputs))
	}
	if ts.Outputs[0].Status != "clean" {
		t.Errorf("output status = %q, want clean", ts.Outputs[0].Status)
	}
}

func TestDeriveTreeClean(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)

	// Write source files first
	os.WriteFile(filepath.Join(dir, "root.md"), []byte("# Root"), 0644)
	os.WriteFile(filepath.Join(dir, "child.md"), []byte("# Child"), 0644)
	time.Sleep(50 * time.Millisecond)

	// Write manifest after sources → everything should be clean
	os.WriteFile(filepath.Join(dir, "compiled", "manifest.md"), []byte("manifest"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Tree Clean",
		Targets: map[string]config.Target{
			"initiative": {
				Outputs: []config.Output{{Path: "compiled/manifest.md"}},
				Tree: &config.Tree{
					System: "jira",
					Root: config.TreeNode{
						ID:   "ROOT-1",
						File: "root.md",
						Children: []config.TreeNode{
							{ID: "CHILD-1", File: "child.md"},
						},
					},
				},
			},
		},
	}

	ps := Derive(f, dir)
	ts := ps.Targets["initiative"]

	if ts.TreeRoot == nil {
		t.Fatal("tree root is nil")
	}
	if ts.TreeRoot.Status != "clean" {
		t.Errorf("root status = %q, want clean", ts.TreeRoot.Status)
	}
	if len(ts.TreeRoot.Children) != 1 {
		t.Fatalf("root children len = %d, want 1", len(ts.TreeRoot.Children))
	}
	if ts.TreeRoot.Children[0].Status != "clean" {
		t.Errorf("child status = %q, want clean", ts.TreeRoot.Children[0].Status)
	}
}

func TestDeriveTreeStaleChild(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)

	// Write root source and manifest first
	os.WriteFile(filepath.Join(dir, "root.md"), []byte("# Root"), 0644)
	time.Sleep(50 * time.Millisecond)
	os.WriteFile(filepath.Join(dir, "compiled", "manifest.md"), []byte("manifest"), 0644)
	time.Sleep(50 * time.Millisecond)

	// Write child source AFTER manifest → child stale, root stale via propagation
	os.WriteFile(filepath.Join(dir, "child.md"), []byte("# Updated child"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Tree Stale",
		Targets: map[string]config.Target{
			"initiative": {
				Outputs: []config.Output{{Path: "compiled/manifest.md"}},
				Tree: &config.Tree{
					System: "jira",
					Root: config.TreeNode{
						ID:   "ROOT-1",
						File: "root.md",
						Children: []config.TreeNode{
							{ID: "CHILD-1", File: "child.md"},
						},
					},
				},
			},
		},
	}

	ps := Derive(f, dir)
	ts := ps.Targets["initiative"]

	if ts.TreeRoot == nil {
		t.Fatal("tree root is nil")
	}

	// Child should be stale (source newer than manifest)
	child := ts.TreeRoot.Children[0]
	if child.Status != "stale" {
		t.Errorf("child status = %q, want stale", child.Status)
	}

	// Root should be stale via bottom-up propagation
	if ts.TreeRoot.Status != "stale" {
		t.Errorf("root status = %q, want stale (propagated from child)", ts.TreeRoot.Status)
	}
	if ts.TreeRoot.CausedBy != "CHILD-1" {
		t.Errorf("root caused_by = %q, want CHILD-1", ts.TreeRoot.CausedBy)
	}
}

func TestDeriveTreeMissingSource(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "compiled", "manifest.md"), []byte("manifest"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Tree Missing",
		Targets: map[string]config.Target{
			"initiative": {
				Outputs: []config.Output{{Path: "compiled/manifest.md"}},
				Tree: &config.Tree{
					System: "jira",
					Root: config.TreeNode{
						ID:   "ROOT-1",
						File: "nonexistent.md",
					},
				},
			},
		},
	}

	ps := Derive(f, dir)
	ts := ps.Targets["initiative"]

	if ts.TreeRoot == nil {
		t.Fatal("tree root is nil")
	}
	if ts.TreeRoot.Status != "missing" {
		t.Errorf("root status = %q, want missing", ts.TreeRoot.Status)
	}
}

func TestDeriveTreeNoManifest(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "root.md"), []byte("# Root"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Tree No Manifest",
		Targets: map[string]config.Target{
			"initiative": {
				Outputs: []config.Output{{Path: "compiled/manifest.md"}},
				Tree: &config.Tree{
					System: "jira",
					Root: config.TreeNode{
						ID:   "ROOT-1",
						File: "root.md",
					},
				},
			},
		},
	}

	ps := Derive(f, dir)
	ts := ps.Targets["initiative"]

	if ts.TreeRoot == nil {
		t.Fatal("tree root is nil")
	}
	// No manifest → zero time → everything stale
	if ts.TreeRoot.Status != "stale" {
		t.Errorf("root status = %q, want stale (no manifest)", ts.TreeRoot.Status)
	}
}

func TestDeriveTreeNoSource(t *testing.T) {
	dir := t.TempDir()

	f := &config.Folio{
		Schema:  1,
		Project: "Tree No Source",
		Targets: map[string]config.Target{
			"initiative": {
				Outputs: []config.Output{{Path: "compiled/manifest.md"}},
				Tree: &config.Tree{
					System: "jira",
					Root: config.TreeNode{
						ID: "ROOT-1",
					},
				},
			},
		},
	}

	ps := Derive(f, dir)
	ts := ps.Targets["initiative"]

	if ts.TreeRoot == nil {
		t.Fatal("tree root is nil")
	}
	// No source → unknown status
	if ts.TreeRoot.Status != "unknown" {
		t.Errorf("root status = %q, want unknown (no source)", ts.TreeRoot.Status)
	}
}

func TestDeriveBatchClean(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)

	// Write sources first
	os.WriteFile(filepath.Join(dir, "tab1.md"), []byte("# Tab 1"), 0644)
	os.WriteFile(filepath.Join(dir, "tab2.md"), []byte("# Tab 2"), 0644)
	time.Sleep(50 * time.Millisecond)

	// Write manifest after sources → everything should be clean
	os.WriteFile(filepath.Join(dir, "compiled", "manifest.md"), []byte("manifest"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Batch Clean",
		Targets: map[string]config.Target{
			"batch-target": {
				Outputs: []config.Output{{Path: "compiled/manifest.md"}},
				Batch: &config.Batch{
					System: "gdocs",
					Items: []config.BatchItem{
						{ID: "tab-1", Source: "tab1.md", Output: config.Output{ID: "doc-tab-1"}},
						{ID: "tab-2", Source: "tab2.md", Output: config.Output{ID: "doc-tab-2"}},
					},
				},
			},
		},
	}

	ps := Derive(f, dir)
	ts := ps.Targets["batch-target"]

	if len(ts.BatchItems) != 2 {
		t.Fatalf("batch items len = %d, want 2", len(ts.BatchItems))
	}
	for _, item := range ts.BatchItems {
		if item.Status != "clean" {
			t.Errorf("item %q status = %q, want clean", item.ID, item.Status)
		}
		if item.System != "gdocs" {
			t.Errorf("item %q system = %q, want gdocs", item.ID, item.System)
		}
	}
}

func TestDeriveBatchStale(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)

	// Write manifest first
	os.WriteFile(filepath.Join(dir, "compiled", "manifest.md"), []byte("manifest"), 0644)
	time.Sleep(50 * time.Millisecond)

	// Write source AFTER manifest → stale
	os.WriteFile(filepath.Join(dir, "tab1.md"), []byte("# Updated"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Batch Stale",
		Targets: map[string]config.Target{
			"batch-target": {
				Outputs: []config.Output{{Path: "compiled/manifest.md"}},
				Batch: &config.Batch{
					System: "gdocs",
					Items: []config.BatchItem{
						{ID: "tab-1", Source: "tab1.md", Output: config.Output{ID: "doc-tab-1"}},
					},
				},
			},
		},
	}

	ps := Derive(f, dir)
	ts := ps.Targets["batch-target"]

	if len(ts.BatchItems) != 1 {
		t.Fatalf("batch items len = %d, want 1", len(ts.BatchItems))
	}
	if ts.BatchItems[0].Status != "stale" {
		t.Errorf("item status = %q, want stale", ts.BatchItems[0].Status)
	}
}

func TestDeriveBatchMissing(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "compiled", "manifest.md"), []byte("manifest"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Batch Missing",
		Targets: map[string]config.Target{
			"batch-target": {
				Outputs: []config.Output{{Path: "compiled/manifest.md"}},
				Batch: &config.Batch{
					System: "gdocs",
					Items: []config.BatchItem{
						{ID: "tab-1", Source: "nonexistent.md", Output: config.Output{ID: "doc-tab-1"}},
					},
				},
			},
		},
	}

	ps := Derive(f, dir)
	ts := ps.Targets["batch-target"]

	if len(ts.BatchItems) != 1 {
		t.Fatalf("batch items len = %d, want 1", len(ts.BatchItems))
	}
	if ts.BatchItems[0].Status != "missing" {
		t.Errorf("item status = %q, want missing", ts.BatchItems[0].Status)
	}
}

func TestDeriveWithDAG(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)

	// Create source file
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Source"), 0644)
	time.Sleep(50 * time.Millisecond)

	// upstream output is clean (written after source)
	os.WriteFile(filepath.Join(dir, "compiled", "summary.md"), []byte("summary"), 0644)
	time.Sleep(50 * time.Millisecond)

	// downstream output is clean (written after upstream output)
	os.WriteFile(filepath.Join(dir, "compiled", "final.md"), []byte("final"), 0644)

	// Now make upstream stale by touching source after outputs
	time.Sleep(50 * time.Millisecond)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Updated source"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "DAG Test",
		Targets: map[string]config.Target{
			"upstream": {
				Sources: []config.Source{{Path: "README.md"}},
				Outputs: []config.Output{{Path: "compiled/summary.md"}},
			},
			"downstream": {
				Sources: []config.Source{{Path: "compiled/summary.md"}},
				Outputs: []config.Output{{Path: "compiled/final.md"}},
			},
		},
	}

	ps, causedBy := DeriveWithDAG(f, dir)

	// upstream should be stale (source newer than output)
	upOut := ps.Targets["upstream"].Outputs[0]
	if upOut.Status != "stale" {
		t.Errorf("upstream status = %q, want stale", upOut.Status)
	}

	// downstream should be transitively stale via propagation
	downOut := ps.Targets["downstream"].Outputs[0]
	if downOut.Status != "stale" {
		t.Errorf("downstream status = %q, want stale (propagated)", downOut.Status)
	}

	// causedBy should record the propagation chain
	if cause, ok := causedBy["downstream"]; !ok || cause != "upstream" {
		t.Errorf("causedBy[downstream] = %q, want upstream", cause)
	}
}
