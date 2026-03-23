package status

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

func TestDeriveLifecycleSummary(t *testing.T) {
	f := &config.Folio{
		Observations: []string{"idea 1", "idea 2"},
		Sources: []config.Source{
			{Path: "reference/spike/s1.md"},
			{Path: "reference/spike/s2.md"},
			{Path: "reference/design/d1.md"},
			{Path: "work/active/2026-01-01-plan/README.md"},
			{Path: "reference/retro/r1.md"},
			{Path: "reference/research/survey1.md"},
			{Path: "reference/guide/g1.md"},
			{Path: "README.md"},
		},
	}
	ls := DeriveLifecycleSummary(f)
	if ls.Observations != 2 {
		t.Errorf("observations = %d, want 2", ls.Observations)
	}
	if ls.Spikes != 2 {
		t.Errorf("spikes = %d, want 2", ls.Spikes)
	}
	if ls.Designs != 1 {
		t.Errorf("designs = %d, want 1", ls.Designs)
	}
	if ls.Plans != 1 {
		t.Errorf("plans = %d, want 1", ls.Plans)
	}
	if ls.Retros != 1 {
		t.Errorf("retros = %d, want 1", ls.Retros)
	}
	if ls.References != 3 {
		t.Errorf("references = %d, want 3", ls.References)
	}
}

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

func TestDeriveLocalCauseClean(t *testing.T) {
	dir := t.TempDir()
	srcPath := filepath.Join(dir, "source.md")
	outPath := filepath.Join(dir, "compiled", "output.md")

	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(srcPath, []byte("source"), 0644)
	time.Sleep(50 * time.Millisecond)
	os.WriteFile(outPath, []byte("output"), 0644)

	cause := DeriveLocalCause(dir, "compiled/output.md", []string{"source.md"})
	if cause != "" {
		t.Errorf("cause = %q, want empty (clean)", cause)
	}
}

func TestDeriveLocalCauseStale(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "compiled", "output.md")
	srcPath := filepath.Join(dir, "source.md")

	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(outPath, []byte("output"), 0644)
	time.Sleep(50 * time.Millisecond)
	os.WriteFile(srcPath, []byte("source newer"), 0644)

	cause := DeriveLocalCause(dir, "compiled/output.md", []string{"source.md"})
	if cause != "source source.md newer than output" {
		t.Errorf("cause = %q, want 'source source.md newer than output'", cause)
	}
}

func TestDeriveLocalCauseMissingOutput(t *testing.T) {
	dir := t.TempDir()
	cause := DeriveLocalCause(dir, "compiled/nonexistent.md", []string{"source.md"})
	if cause != "output missing" {
		t.Errorf("cause = %q, want 'output missing'", cause)
	}
}

func TestDeriveLocalCauseMissingSource(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "compiled", "output.md"), []byte("output"), 0644)

	cause := DeriveLocalCause(dir, "compiled/output.md", []string{"nonexistent.md"})
	if cause != "source nonexistent.md missing" {
		t.Errorf("cause = %q, want 'source nonexistent.md missing'", cause)
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
	}

	ps := Derive(f, dir)
	if ps.Project != "Test Project" {
		t.Errorf("project = %q", ps.Project)
	}
	if len(ps.Sources) != 1 {
		t.Errorf("sources len = %d", len(ps.Sources))
	}

	ts := ps.Targets["summary"]
	if len(ts.Outputs) != 1 {
		t.Fatalf("outputs len = %d", len(ts.Outputs))
	}
	if ts.Outputs[0].Status != "clean" {
		t.Errorf("output status = %q, want clean", ts.Outputs[0].Status)
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
