package status

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/touch"
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

func TestDeriveLocalStatusUsesDigestNotMtime(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "source.md")
	outputPath := filepath.Join(dir, "compiled", "output.md")
	writeStatusFile(t, sourcePath, "source")
	writeStatusFile(t, outputPath, "output")
	ctx := config.Context{FolioPath: filepath.Join(dir, "folio.yml")}
	target := &config.Target{
		Sources: []config.Source{{Path: "source.md"}},
		Outputs: []config.Output{{Path: "compiled/output.md"}},
	}
	mtime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(sourcePath, mtime, mtime); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(outputPath, mtime, mtime); err != nil {
		t.Fatal(err)
	}
	stampStatusTarget(t, ctx, target)

	got, cause := DeriveLocalStatus(ctx, target, target.Outputs[0])
	if got != "clean" || cause != "" {
		t.Fatalf("status = %q cause = %q, want clean/empty", got, cause)
	}
}

func TestDeriveLocalStatusStaleWhenInputChanges(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "source.md")
	outputPath := filepath.Join(dir, "compiled", "output.md")
	writeStatusFile(t, sourcePath, "source")
	writeStatusFile(t, outputPath, "output")
	ctx := config.Context{FolioPath: filepath.Join(dir, "folio.yml")}
	target := &config.Target{
		Sources: []config.Source{{Path: "source.md"}},
		Outputs: []config.Output{{Path: "compiled/output.md"}},
	}
	stampStatusTarget(t, ctx, target)
	writeStatusFile(t, sourcePath, "changed")
	future := time.Now().Add(time.Hour)
	if err := os.Chtimes(outputPath, future, future); err != nil {
		t.Fatal(err)
	}

	got, cause := DeriveLocalStatus(ctx, target, target.Outputs[0])
	if got != "stale" || cause != "source content changed since compose" {
		t.Fatalf("status = %q cause = %q, want stale/content change", got, cause)
	}
}

func TestDeriveLocalStatusMissingOutput(t *testing.T) {
	dir := t.TempDir()
	ctx := config.Context{FolioPath: filepath.Join(dir, "folio.yml")}
	target := &config.Target{Outputs: []config.Output{{Path: "compiled/missing.md"}}}

	got, cause := DeriveLocalStatus(ctx, target, target.Outputs[0])
	if got != "missing" || cause != "output missing" {
		t.Fatalf("status = %q cause = %q, want missing/output missing", got, cause)
	}
}

func TestDeriveLocalStatusMissingSource(t *testing.T) {
	dir := t.TempDir()
	writeStatusFile(t, filepath.Join(dir, "source.md"), "source")
	writeStatusFile(t, filepath.Join(dir, "compiled", "output.md"), "output")
	ctx := config.Context{FolioPath: filepath.Join(dir, "folio.yml")}
	target := &config.Target{
		Sources: []config.Source{{Path: "source.md"}},
		Outputs: []config.Output{{Path: "compiled/output.md"}},
	}
	stampStatusTarget(t, ctx, target)
	if err := os.Remove(filepath.Join(dir, "source.md")); err != nil {
		t.Fatal(err)
	}

	got, cause := DeriveLocalStatus(ctx, target, target.Outputs[0])
	if got != "stale" || cause != "source source.md missing" {
		t.Fatalf("status = %q cause = %q, want stale/source missing", got, cause)
	}
}

func TestDeriveLocalStatusUnknownWithoutSnapshot(t *testing.T) {
	dir := t.TempDir()
	writeStatusFile(t, filepath.Join(dir, "source.md"), "source")
	writeStatusFile(t, filepath.Join(dir, "compiled", "output.md"), "output")
	ctx := config.Context{FolioPath: filepath.Join(dir, "folio.yml")}
	target := &config.Target{
		Sources: []config.Source{{Path: "source.md"}},
		Outputs: []config.Output{{Path: "compiled/output.md"}},
	}

	got, cause := DeriveLocalStatus(ctx, target, target.Outputs[0])
	if got != "unknown" || cause != "composition not recorded" {
		t.Fatalf("status = %q cause = %q, want unknown/not recorded", got, cause)
	}
}

func TestDeriveLocalStatusUnknownForExternalInput(t *testing.T) {
	dir := t.TempDir()
	writeStatusFile(t, filepath.Join(dir, "compiled", "output.md"), "output")
	ctx := config.Context{FolioPath: filepath.Join(dir, "folio.yml")}
	target := &config.Target{
		Sources: []config.Source{{External: "jira", ID: "ONE"}},
		Outputs: []config.Output{{Path: "compiled/output.md"}},
	}

	got, cause := DeriveLocalStatus(ctx, target, target.Outputs[0])
	if got != "unknown" || cause != "external source freshness unverified" {
		t.Fatalf("status = %q cause = %q, want unknown/external", got, cause)
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
	writeStatusFile(t, filepath.Join(dir, "README.md"), "# Test")
	writeStatusFile(t, filepath.Join(dir, "compiled", "out.md"), "compiled")

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
	ctx := config.Context{FolioPath: filepath.Join(dir, "folio.yml"), WorkRoot: dir}
	target := f.Targets["summary"]
	stampStatusTarget(t, ctx, &target)
	f.Targets["summary"] = target

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
	writeStatusFile(t, filepath.Join(dir, "tab1.md"), "# Tab 1")
	writeStatusFile(t, filepath.Join(dir, "tab2.md"), "# Tab 2")

	f := &config.Folio{
		Schema:  1,
		Project: "Batch Clean",
		Targets: map[string]config.Target{
			"batch-target": {
				Batch: &config.Batch{
					System: "gdocs",
					Items: []config.BatchItem{
						{ID: "tab-1", Source: "tab1.md", Output: config.Output{ID: "doc-tab-1"}},
						{ID: "tab-2", Source: "tab2.md", Output: config.Output{ID: "doc-tab-2"}},
						{ID: "manual", Output: config.Output{ID: "doc-manual"}},
					},
				},
			},
		},
	}
	ctx := config.Context{FolioPath: filepath.Join(dir, "folio.yml"), WorkRoot: dir}
	target := f.Targets["batch-target"]
	stampStatusTarget(t, ctx, &target)
	f.Targets["batch-target"] = target

	ts := Derive(f, dir).Targets["batch-target"]
	if len(ts.BatchItems) != 3 {
		t.Fatalf("batch items len = %d, want 3", len(ts.BatchItems))
	}
	for _, item := range ts.BatchItems {
		want := "clean"
		if item.ID == "manual" {
			want = "unknown"
		}
		if item.Status != want {
			t.Errorf("item %q status = %q, want %s", item.ID, item.Status, want)
		}
		if item.ID != "manual" && item.System != "gdocs" {
			t.Errorf("item %q system = %q, want gdocs", item.ID, item.System)
		}
	}
}

func TestDeriveBatchStale(t *testing.T) {
	dir := t.TempDir()
	writeStatusFile(t, filepath.Join(dir, "tab1.md"), "# Original")

	f := &config.Folio{
		Schema:  1,
		Project: "Batch Stale",
		Targets: map[string]config.Target{
			"batch-target": {
				Batch: &config.Batch{
					System: "gdocs",
					Items: []config.BatchItem{
						{ID: "tab-1", Source: "tab1.md", Output: config.Output{ID: "doc-tab-1"}},
					},
				},
			},
		},
	}
	ctx := config.Context{FolioPath: filepath.Join(dir, "folio.yml"), WorkRoot: dir}
	target := f.Targets["batch-target"]
	stampStatusTarget(t, ctx, &target)
	f.Targets["batch-target"] = target
	writeStatusFile(t, filepath.Join(dir, "tab1.md"), "# Updated")

	ts := Derive(f, dir).Targets["batch-target"]
	if len(ts.BatchItems) != 1 {
		t.Fatalf("batch items len = %d, want 1", len(ts.BatchItems))
	}
	if ts.BatchItems[0].Status != "stale" {
		t.Errorf("item status = %q, want stale", ts.BatchItems[0].Status)
	}
}
func TestDeriveBatchMissing(t *testing.T) {
	dir := t.TempDir()
	writeStatusFile(t, filepath.Join(dir, "compiled", "manifest.md"), "manifest")
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

	ts := Derive(f, dir).Targets["batch-target"]
	if len(ts.BatchItems) != 1 {
		t.Fatalf("batch items len = %d, want 1", len(ts.BatchItems))
	}
	if ts.BatchItems[0].Status != "unknown" {
		t.Errorf("item status = %q, want unknown without persisted digest", ts.BatchItems[0].Status)
	}
}

func TestStaleOrMissingUpstreamPropagatesStale(t *testing.T) {
	dir := t.TempDir()
	writeStatusFile(t, filepath.Join(dir, "README.md"), "# Source")
	writeStatusFile(t, filepath.Join(dir, "compiled", "summary.md"), "summary")
	writeStatusFile(t, filepath.Join(dir, "compiled", "final.md"), "final")

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
	ctx := config.Context{FolioPath: filepath.Join(dir, "folio.yml"), WorkRoot: dir}
	upstream := f.Targets["upstream"]
	stampStatusTarget(t, ctx, &upstream)
	f.Targets["upstream"] = upstream
	downstream := f.Targets["downstream"]
	stampStatusTarget(t, ctx, &downstream)
	f.Targets["downstream"] = downstream
	writeStatusFile(t, filepath.Join(dir, "README.md"), "# Updated source")

	ps, causedBy, causedStatus := DeriveWithDAG(f, dir)
	upOut := ps.Targets["upstream"].Outputs[0]
	if upOut.Status != "stale" {
		t.Errorf("upstream status = %q, want stale", upOut.Status)
	}
	downOut := ps.Targets["downstream"].Outputs[0]
	if downOut.Status != "stale" {
		t.Errorf("downstream status = %q, want stale (propagated)", downOut.Status)
	}
	if cause, ok := causedBy["downstream"]; !ok || cause != "upstream" {
		t.Errorf("causedBy[downstream] = %q, want upstream", cause)
	}
	if causedStatus["downstream"] != "stale" {
		t.Errorf("causedStatus[downstream] = %q, want stale", causedStatus["downstream"])
	}
}
func TestUnknownUpstreamPropagatesUnknown(t *testing.T) {
	dir := t.TempDir()
	writeStatusFile(t, filepath.Join(dir, "README.md"), "# Source")
	writeStatusFile(t, filepath.Join(dir, "compiled", "summary.md"), "summary")
	writeStatusFile(t, filepath.Join(dir, "compiled", "final.md"), "final")
	f := &config.Folio{
		Schema:  1,
		Project: "Unknown DAG",
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
	ctx := config.Context{FolioPath: filepath.Join(dir, "folio.yml"), WorkRoot: dir}
	downstream := f.Targets["downstream"]
	stampStatusTarget(t, ctx, &downstream)
	f.Targets["downstream"] = downstream

	ps, causedBy, causedStatus := DeriveWithDAG(f, dir)
	if got := ps.Targets["downstream"].Outputs[0].Status; got != "unknown" {
		t.Fatalf("downstream status = %q, want unknown", got)
	}
	if causedBy["downstream"] != "upstream" {
		t.Fatalf("causedBy[downstream] = %q, want upstream", causedBy["downstream"])
	}
	if causedStatus["downstream"] != "unknown" {
		t.Fatalf("causedStatus[downstream] = %q, want unknown", causedStatus["downstream"])
	}
}

func TestFinalTargetEmitsStaleDownstreamWithoutReceivingPropagation(t *testing.T) {
	dir := t.TempDir()
	writeStatusFile(t, filepath.Join(dir, "README.md"), "# Source")
	writeStatusFile(t, filepath.Join(dir, "compiled", "upstream.md"), "upstream")
	writeStatusFile(t, filepath.Join(dir, "compiled", "terminal.md"), "terminal")
	writeStatusFile(t, filepath.Join(dir, "compiled", "downstream.md"), "downstream")
	f := &config.Folio{
		Schema:  1,
		Project: "Terminal DAG",
		Targets: map[string]config.Target{
			"upstream": {
				Sources: []config.Source{{Path: "README.md"}},
				Outputs: []config.Output{{Path: "compiled/upstream.md"}},
			},
			"terminal": {
				Sources: []config.Source{{Path: "compiled/upstream.md"}},
				Outputs: []config.Output{{Path: "compiled/terminal.md"}},
				Final:   true,
			},
			"downstream": {
				Sources: []config.Source{{Path: "compiled/terminal.md"}},
				Outputs: []config.Output{{Path: "compiled/downstream.md"}},
			},
		},
	}
	ctx := config.Context{FolioPath: filepath.Join(dir, "folio.yml"), WorkRoot: dir}
	for tid, target := range f.Targets {
		stampStatusTarget(t, ctx, &target)
		f.Targets[tid] = target
	}
	writeStatusFile(t, filepath.Join(dir, "README.md"), "# Updated")
	writeStatusFile(t, filepath.Join(dir, "compiled", "upstream.md"), "upstream changed")

	ps, causedBy, causedStatus := DeriveWithDAG(f, dir)
	if got := ps.Targets["terminal"].Outputs[0].Status; got != "stale" {
		t.Fatalf("terminal status = %q, want stale", got)
	}
	if _, ok := causedBy["terminal"]; ok {
		t.Fatalf("terminal unexpectedly received propagation from %q", causedBy["terminal"])
	}
	if got := ps.Targets["downstream"].Outputs[0].Status; got != "stale" {
		t.Fatalf("downstream status = %q, want stale", got)
	}
	if causedBy["downstream"] != "terminal" || causedStatus["downstream"] != "stale" {
		t.Fatalf("downstream cause = %q/%q, want terminal/stale", causedBy["downstream"], causedStatus["downstream"])
	}
}

func TestExternalOutputsDoNotAggregateOrReceivePropagation(t *testing.T) {
	dir := t.TempDir()
	writeStatusFile(t, filepath.Join(dir, "README.md"), "# Source")
	writeStatusFile(t, filepath.Join(dir, "stable.md"), "# Stable")
	writeStatusFile(t, filepath.Join(dir, "compiled", "upstream.md"), "upstream")
	writeStatusFile(t, filepath.Join(dir, "compiled", "mixed.md"), "mixed")
	f := &config.Folio{
		Schema:  1,
		Project: "External Outputs",
		Targets: map[string]config.Target{
			"upstream": {
				Sources: []config.Source{{Path: "README.md"}},
				Outputs: []config.Output{{Path: "compiled/upstream.md"}},
			},
			"external-only": {
				Sources: []config.Source{{Path: "compiled/upstream.md"}},
				Outputs: []config.Output{{External: "jira", ID: "EXT-1"}},
			},
			"mixed": {
				Sources: []config.Source{{Path: "stable.md"}},
				Outputs: []config.Output{
					{Path: "compiled/mixed.md"},
					{External: "jira", ID: "EXT-2"},
				},
			},
		},
	}
	ctx := config.Context{FolioPath: filepath.Join(dir, "folio.yml"), WorkRoot: dir}
	mixed := f.Targets["mixed"]
	stampStatusTarget(t, ctx, &mixed)
	f.Targets["mixed"] = mixed
	upstream := f.Targets["upstream"]
	stampStatusTarget(t, ctx, &upstream)
	f.Targets["upstream"] = upstream
	writeStatusFile(t, filepath.Join(dir, "README.md"), "# Updated")

	ps, causedBy, _ := DeriveWithDAG(f, dir)
	externalOnly := ps.Targets["external-only"]
	if externalOnly.Outputs[0].Status != "unknown" || externalOnly.Outputs[0].Cause != "" {
		t.Fatalf("external-only output = %+v", externalOnly.Outputs[0])
	}
	if _, ok := causedBy["external-only"]; ok {
		t.Fatal("external-only target received propagated status")
	}
	mixedStatus := ps.Targets["mixed"]
	if mixedStatus.Outputs[0].Status != "clean" || mixedStatus.Outputs[1].Status != "unknown" {
		t.Fatalf("mixed outputs = %+v", mixedStatus.Outputs)
	}
}

func TestForestTargetIsDelegatedAndExcludedFromStaleQueue(t *testing.T) {
	dir := t.TempDir()
	writeStatusFile(t, filepath.Join(dir, "compiled", "forest.md"), "forest")
	ctx := config.Context{FolioPath: filepath.Join(dir, "folio.yml"), WorkRoot: dir}
	target := &config.Target{
		Forest:  &config.Forest{Root: "forest"},
		Outputs: []config.Output{{Path: "compiled/forest.md"}},
	}

	got, cause := DeriveLocalStatus(ctx, target, target.Outputs[0])
	if got != "unknown" || cause != "freshness delegated to jf" {
		t.Fatalf("forest status = %q cause = %q, want unknown/delegated", got, cause)
	}
}

func stampStatusTarget(t *testing.T, ctx config.Context, target *config.Target) {
	t.Helper()
	state, err := touch.Snapshot(ctx, target, false, time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	target.ComposedAt = state.ComposedAt
	target.InputsSHA256 = state.InputsSHA256
	if target.Batch != nil {
		for _, itemState := range state.Items {
			for i := range target.Batch.Items {
				if target.Batch.Items[i].ID == itemState.ID {
					target.Batch.Items[i].InputsSHA256 = itemState.InputsSHA256
				}
			}
		}
	}
}

func writeStatusFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
