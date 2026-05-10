package health

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

func TestAnalyzeEmptyProject(t *testing.T) {
	dir := t.TempDir()
	f := &config.Folio{Project: "empty"}

	r := Analyze(f, dir)
	if r.Grade != "Good" {
		t.Errorf("empty project grade = %q, want Good", r.Grade)
	}
	if r.TotalReferenceFiles() != 0 {
		t.Errorf("expected 0 reference files, got %d", r.TotalReferenceFiles())
	}
}

func TestAnalyzeTypedReferences(t *testing.T) {
	dir := t.TempDir()

	// Create typed reference files (spike/retro are no longer reference types)
	for _, typ := range []string{"survey", "design"} {
		typeDir := filepath.Join(dir, "reference", typ)
		os.MkdirAll(typeDir, 0755)
	}
	os.WriteFile(filepath.Join(dir, "reference", "survey", "2026-01-01-foo.md"), []byte("# Foo"), 0644)
	os.WriteFile(filepath.Join(dir, "reference", "survey", "2026-01-02-bar.md"), []byte("# Bar"), 0644)
	os.WriteFile(filepath.Join(dir, "reference", "design", "2026-01-01-baz.md"), []byte("# Baz"), 0644)

	f := &config.Folio{Project: "typed"}
	r := Analyze(f, dir)

	if r.Reference["survey"] != 2 {
		t.Errorf("survey count = %d, want 2", r.Reference["survey"])
	}
	if r.Reference["design"] != 1 {
		t.Errorf("design count = %d, want 1", r.Reference["design"])
	}
	if r.TotalReferenceFiles() != 3 {
		t.Errorf("total = %d, want 3", r.TotalReferenceFiles())
	}
	if r.Grade != "Good" {
		t.Errorf("grade = %q, want Good", r.Grade)
	}
}

func TestAnalyzeUntypedFiles(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "reference"), 0755)
	os.WriteFile(filepath.Join(dir, "reference", "flat-file.md"), []byte("# Flat"), 0644)

	f := &config.Folio{Project: "untyped"}
	r := Analyze(f, dir)

	if len(r.Untyped) != 1 {
		t.Errorf("untyped count = %d, want 1", len(r.Untyped))
	}
	if r.Grade != "Needs Attention" {
		t.Errorf("grade = %q, want Needs Attention", r.Grade)
	}
}

func TestAnalyzeUnrecognizedDirs(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "reference", "unknown-type"), 0755)
	os.WriteFile(filepath.Join(dir, "reference", "unknown-type", "file.md"), []byte("# X"), 0644)

	f := &config.Folio{Project: "unrecognized"}
	r := Analyze(f, dir)

	if len(r.Unrecognized) != 1 {
		t.Errorf("unrecognized count = %d, want 1", len(r.Unrecognized))
	}
	if r.Grade != "Needs Attention" {
		t.Errorf("grade = %q, want Needs Attention", r.Grade)
	}
}

func TestAnalyzeObservations(t *testing.T) {
	f := &config.Folio{
		Project: "observations",
		Observations: []string{
			"Active item one",
			"Active item two",
			"Active item three",
		},
	}

	dir := t.TempDir()
	r := Analyze(f, dir)

	if r.Observations.Active != 3 {
		t.Errorf("observations active = %d, want 3", r.Observations.Active)
	}
}

func TestAnalyzeObservationsLintWarnings(t *testing.T) {
	f := &config.Folio{
		Project: "lint-test",
		Observations: []string{
			"bug(cli): valid item",
			"freeform text that fails validation",
		},
	}

	dir := t.TempDir()
	r := Analyze(f, dir)

	if r.Observations.Active != 2 {
		t.Errorf("observations active = %d, want 2", r.Observations.Active)
	}
	if len(r.Observations.LintWarnings) != 1 {
		t.Fatalf("lint warnings = %d, want 1", len(r.Observations.LintWarnings))
	}
	if r.Observations.LintWarnings[0] != "#2: malformed format" {
		t.Errorf("warning = %q, want #2: malformed format", r.Observations.LintWarnings[0])
	}
}

func TestAnalyzeNaming(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "reference", "survey"), 0755)
	os.WriteFile(filepath.Join(dir, "reference", "survey", "2026-01-01-good.md"), []byte("# Good"), 0644)
	os.WriteFile(filepath.Join(dir, "reference", "survey", "no-date-prefix.md"), []byte("# Bad"), 0644)

	f := &config.Folio{Project: "naming"}
	r := Analyze(f, dir)

	if len(r.Naming) != 1 {
		t.Errorf("naming issues = %d, want 1", len(r.Naming))
	}
	if r.Grade != "Needs Attention" {
		t.Errorf("grade = %q, want Needs Attention", r.Grade)
	}
}

func TestAnalyzeWorkLayer(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "work", "active", "2026-01-01-proj-a"), 0755)
	os.MkdirAll(filepath.Join(dir, "work", "active", "2026-01-02-proj-b"), 0755)
	os.MkdirAll(filepath.Join(dir, "work", "archive", "2026-01-01-old"), 0755)

	f := &config.Folio{Project: "work"}
	r := Analyze(f, dir)

	if r.Work.Active != 2 {
		t.Errorf("work active = %d, want 2", r.Work.Active)
	}
	if r.Work.Archived != 1 {
		t.Errorf("work archived = %d, want 1", r.Work.Archived)
	}
}

func TestAnalyzeRetroReport(t *testing.T) {
	dir := t.TempDir()

	// Create a work dir that matches topic "skill-rewrite"
	os.MkdirAll(filepath.Join(dir, "work", "active", "2026-03-01-skill-rewrite"), 0755)

	// Orphaned retro in reference/retro/ (has matching work dir)
	os.MkdirAll(filepath.Join(dir, "reference", "retro"), 0755)
	os.WriteFile(filepath.Join(dir, "reference", "retro", "2026-03-01-skill-rewrite.md"), []byte("# Retro"), 0644)

	// Non-orphaned retro in reference/retro/ (no matching work dir)
	os.WriteFile(filepath.Join(dir, "reference", "retro", "2026-03-01-no-match.md"), []byte("# Retro"), 0644)

	// Colocated retro inside work dir
	os.MkdirAll(filepath.Join(dir, "work", "active", "2026-03-01-skill-rewrite", "reference", "retro"), 0755)
	os.WriteFile(filepath.Join(dir, "work", "active", "2026-03-01-skill-rewrite", "reference", "retro", "retro.md"), []byte("# Retro"), 0644)

	f := &config.Folio{Project: "retro-test"}
	r := Analyze(f, dir)

	if r.Retro.Total != 3 {
		t.Errorf("retro total = %d, want 3", r.Retro.Total)
	}
	if r.Retro.Colocated != 1 {
		t.Errorf("retro colocated = %d, want 1", r.Retro.Colocated)
	}
	if r.Retro.Orphaned != 1 {
		t.Errorf("retro orphaned = %d, want 1", r.Retro.Orphaned)
	}
}
