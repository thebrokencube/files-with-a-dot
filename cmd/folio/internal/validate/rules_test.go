package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

func TestValidateMinimal(t *testing.T) {
	f := &config.Folio{
		Schema:  1,
		Project: "Test",
	}
	r := Validate(f, t.TempDir())
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestValidateBadSchema(t *testing.T) {
	f := &config.Folio{
		Schema:  99,
		Project: "Test",
	}
	r := Validate(f, t.TempDir())
	if r.Valid {
		t.Error("expected invalid for schema 99")
	}
	if !containsError(r, "schema version") {
		t.Errorf("expected schema error, got: %v", r.Errors)
	}
}

func TestValidateSchema2Valid(t *testing.T) {
	f := &config.Folio{
		Schema:  2,
		Project: "Test",
	}
	r := Validate(f, t.TempDir())
	if !r.Valid {
		t.Errorf("expected valid for schema 2, got errors: %v", r.Errors)
	}
}

func TestValidateMissingProject(t *testing.T) {
	f := &config.Folio{
		Schema: 1,
	}
	r := Validate(f, t.TempDir())
	if r.Valid {
		t.Error("expected invalid for missing project")
	}
	if !containsError(r, "project") {
		t.Errorf("expected project error, got: %v", r.Errors)
	}
}

func TestValidateDeprecatedContextSources(t *testing.T) {
	f := &config.Folio{
		Schema:         1,
		Project:        "Test",
		ContextSources: "something",
	}
	r := Validate(f, t.TempDir())
	if !r.Valid {
		t.Errorf("expected valid (warning only), got errors: %v", r.Errors)
	}
	if len(r.Warnings) == 0 {
		t.Error("expected deprecation warning")
	}
}

func TestValidateSourceFileNotFound(t *testing.T) {
	dir := t.TempDir()
	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Sources: []config.Source{
			{Path: "nonexistent.md"},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for missing source file")
	}
	if !containsError(r, "file not found") {
		t.Errorf("expected file not found error, got: %v", r.Errors)
	}
}

func TestValidateSourceFileExists(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Sources: []config.Source{
			{Path: "README.md"},
		},
	}
	r := Validate(f, dir)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestValidateExternalSourceMissingID(t *testing.T) {
	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Sources: []config.Source{
			{External: "jira"},
		},
	}
	r := Validate(f, t.TempDir())
	if r.Valid {
		t.Error("expected invalid for external source missing id")
	}
	if !containsError(r, "missing required field: id") {
		t.Errorf("expected id error, got: %v", r.Errors)
	}
}

func TestValidateDerivedSourceMissingExternal(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "cached.md"), []byte("# Cached"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Sources: []config.Source{
			{Path: "cached.md", DerivedFrom: []config.DerivedFrom{{}}},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for derived_from missing external")
	}
	if !containsError(r, "derived_from[0] missing required field: external") {
		t.Errorf("expected derived_from error, got: %v", r.Errors)
	}
}

func TestValidateMissingHowWarnsNotErrors(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"my-target": {
				Outputs: []config.Output{{Path: "compiled/out.md"}},
			},
		},
	}
	r := Validate(f, dir)
	if !r.Valid {
		t.Errorf("expected valid (warning only for missing how), got errors: %v", r.Errors)
	}
	if !containsWarning(r, "missing 'how' field") {
		t.Errorf("expected how warning, got warnings: %v", r.Warnings)
	}
}

func TestValidateTargetDeprecatedInstructions(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"my-target": {
				Instructions: "Test via deprecated field",
				Outputs:      []config.Output{{Path: "compiled/out.md"}},
			},
		},
	}
	r := Validate(f, dir)
	if !r.Valid {
		t.Errorf("expected valid (warning only), got errors: %v", r.Errors)
	}
	if !containsWarning(r, "'instructions' is deprecated") {
		t.Errorf("expected deprecation warning, got warnings: %v", r.Warnings)
	}
}

func TestValidateTargetBothHowAndInstructions(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"my-target": {
				How:          "New field",
				Instructions: "Old field",
				Outputs:      []config.Output{{Path: "compiled/out.md"}},
			},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for both how and instructions")
	}
	if !containsError(r, "has both 'how' and 'instructions'") {
		t.Errorf("expected conflict error, got: %v", r.Errors)
	}
}

func TestValidateTargetDeprecatedTransform(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"my-target": {
				How:       "Test",
				Transform: "distill",
				Outputs:   []config.Output{{Path: "compiled/out.md"}},
			},
		},
	}
	r := Validate(f, dir)
	if !r.Valid {
		t.Errorf("expected valid (warning only), got errors: %v", r.Errors)
	}
	if !containsWarning(r, "'transform' is deprecated") {
		t.Errorf("expected deprecation warning, got warnings: %v", r.Warnings)
	}
}

func TestValidatePrecompileRule(t *testing.T) {
	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"ext-only": {
				How: "External only",
				Outputs: []config.Output{
					{External: "jira", ID: "PROJ-123", Field: "description"},
				},
			},
		},
	}
	r := Validate(f, t.TempDir())
	if r.Valid {
		t.Error("expected invalid for precompile violation")
	}
	if !containsError(r, "precompile rule") {
		t.Errorf("expected precompile error, got: %v", r.Errors)
	}
}

func TestValidateExternalOutputMissingID(t *testing.T) {
	dir := t.TempDir()

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"my-target": {
				How: "Test",
				Outputs: []config.Output{
					{External: "jira"},
					{Path: "compiled/out.md"},
				},
			},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for external output missing id")
	}
	if !containsError(r, "external output") {
		t.Errorf("expected external output error, got: %v", r.Errors)
	}
}

func TestValidateOutputCollision(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"first": {
				How:     "First",
				Outputs: []config.Output{{Path: "compiled/same.md"}},
			},
			"second": {
				How:     "Second",
				Outputs: []config.Output{{Path: "compiled/same.md"}},
			},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for output collision")
	}
	if !containsError(r, "Output collision") {
		t.Errorf("expected collision error, got: %v", r.Errors)
	}
}

func TestValidateDependsOnValid(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "spike.md"), []byte("# Spike"), 0644)
	os.WriteFile(filepath.Join(dir, "design.md"), []byte("# Design"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Sources: []config.Source{
			{Path: "spike.md"},
			{Path: "design.md", DependsOn: []string{"spike.md"}},
		},
	}
	r := Validate(f, dir)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestValidateDependsOnExternalGuard(t *testing.T) {
	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Sources: []config.Source{
			{External: "jira", ID: "PROJ-1", DependsOn: []string{"spike.md"}},
		},
	}
	r := Validate(f, t.TempDir())
	if r.Valid {
		t.Error("expected invalid for external source with depends_on")
	}
	if !containsError(r, "depends_on is only valid on local path sources") {
		t.Errorf("expected external guard error, got: %v", r.Errors)
	}
}

func TestValidateDependsOnUnresolved(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "design.md"), []byte("# Design"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Sources: []config.Source{
			{Path: "design.md", DependsOn: []string{"nonexistent.md"}},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for unresolved depends_on")
	}
	if !containsError(r, "depends_on references non-existent source") {
		t.Errorf("expected unresolved error, got: %v", r.Errors)
	}
}

func TestValidateDependsOnCycle(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "a.md"), []byte("# A"), 0644)
	os.WriteFile(filepath.Join(dir, "b.md"), []byte("# B"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Sources: []config.Source{
			{Path: "a.md", DependsOn: []string{"b.md"}},
			{Path: "b.md", DependsOn: []string{"a.md"}},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for source dependency cycle")
	}
	if !containsError(r, "Source dependency cycle") {
		t.Errorf("expected cycle error, got: %v", r.Errors)
	}
}

func TestValidateBatchItemMissingID(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "src.md"), []byte("# Source"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"batch-target": {
				How:     "Test batch",
				Outputs: []config.Output{{Path: "compiled/manifest.md"}},
				Batch: &config.Batch{
					System: "gdocs",
					Items: []config.BatchItem{
						{Source: "src.md", Output: config.Output{ID: "tab-1"}},
					},
				},
			},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for batch item missing id")
	}
	if !containsError(r, "missing required field: id") {
		t.Errorf("expected id error, got: %v", r.Errors)
	}
}

func TestValidateBatchItemResolvedOutput(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "src.md"), []byte("# Source"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"batch-target": {
				How:     "Test batch",
				Outputs: []config.Output{{Path: "compiled/manifest.md"}},
				Batch: &config.Batch{
					System: "gdocs",
					Field:  "body",
					Items: []config.BatchItem{
						{ID: "tab-1", Source: "src.md", Output: config.Output{ID: "doc-tab-1"}},
					},
				},
			},
		},
	}
	r := Validate(f, dir)
	if !r.Valid {
		t.Errorf("expected valid batch target with defaults, got errors: %v", r.Errors)
	}
}

func TestValidateBatchItemNoSystemAnywhere(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "src.md"), []byte("# Source"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"batch-target": {
				How:     "Test batch",
				Outputs: []config.Output{{Path: "compiled/manifest.md"}},
				Batch: &config.Batch{
					// No system at batch level
					Items: []config.BatchItem{
						{ID: "tab-1", Source: "src.md", Output: config.Output{ID: "doc-tab-1"}},
					},
				},
			},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for batch item with no system anywhere")
	}
	if !containsError(r, "resolved output missing required field: external") {
		t.Errorf("expected external error, got: %v", r.Errors)
	}
}

func TestValidateEmptySource(t *testing.T) {
	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Sources: []config.Source{
			{}, // neither path nor external
		},
	}
	r := Validate(f, t.TempDir())
	if r.Valid {
		t.Error("expected invalid for empty source")
	}
	if !containsError(r, "must have either 'path' or 'external' set") {
		t.Errorf("expected empty source error, got: %v", r.Errors)
	}
}

func TestValidateAmbiguousSource(t *testing.T) {
	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Sources: []config.Source{
			{Path: "README.md", External: "jira", ID: "PROJ-123"},
		},
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Test"), 0644)

	r := Validate(f, dir)
	// Should still be valid (warning only)
	if !r.Valid {
		t.Errorf("expected valid (warning only), got errors: %v", r.Errors)
	}
	if !containsWarning(r, "both 'path' and 'external' set") {
		t.Errorf("expected ambiguity warning, got warnings: %v", r.Warnings)
	}
}

func TestValidateRepositoryValid(t *testing.T) {
	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Repositories: map[string]string{
			"dotfiles": "https://github.com/org/repo/blob/main/{path}",
		},
	}
	r := Validate(f, t.TempDir())
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
	if containsWarning(r, "Repository") {
		t.Errorf("expected no repo warnings, got: %v", r.Warnings)
	}
}

func TestValidateRepositoryEmptyURL(t *testing.T) {
	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Repositories: map[string]string{
			"dotfiles": "",
		},
	}
	r := Validate(f, t.TempDir())
	if r.Valid {
		t.Error("expected invalid for empty URL")
	}
	if !containsError(r, "URL template is empty") {
		t.Errorf("expected empty URL error, got: %v", r.Errors)
	}
}

func TestValidateRepositoryMissingPlaceholder(t *testing.T) {
	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Repositories: map[string]string{
			"dotfiles": "https://github.com/org/repo",
		},
	}
	r := Validate(f, t.TempDir())
	if !r.Valid {
		t.Errorf("expected valid (warning only), got errors: %v", r.Errors)
	}
	if !containsWarning(r, "{path} placeholder") {
		t.Errorf("expected placeholder warning, got warnings: %v", r.Warnings)
	}
}

func TestValidateCrossRefValid(t *testing.T) {
	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		CrossReferences: []config.CrossReference{
			{Fact: "Some fact", SourceOfTruth: "path/to/source.md"},
		},
	}
	r := Validate(f, t.TempDir())
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestValidateCrossRefMissingFact(t *testing.T) {
	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		CrossReferences: []config.CrossReference{
			{Fact: "", SourceOfTruth: "path/to/source.md"},
		},
	}
	r := Validate(f, t.TempDir())
	if r.Valid {
		t.Error("expected invalid for missing fact")
	}
	if !containsError(r, "missing required field: fact") {
		t.Errorf("expected fact error, got: %v", r.Errors)
	}
}

func TestValidateCrossRefMissingSourceOfTruth(t *testing.T) {
	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		CrossReferences: []config.CrossReference{
			{Fact: "Some fact", SourceOfTruth: ""},
		},
	}
	r := Validate(f, t.TempDir())
	if r.Valid {
		t.Error("expected invalid for missing source_of_truth")
	}
	if !containsError(r, "missing required field: source_of_truth") {
		t.Errorf("expected source_of_truth error, got: %v", r.Errors)
	}
}

func TestValidateCrossRefDuplicateFact(t *testing.T) {
	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		CrossReferences: []config.CrossReference{
			{Fact: "Same fact", SourceOfTruth: "path/a.md"},
			{Fact: "Same fact", SourceOfTruth: "path/b.md"},
		},
	}
	r := Validate(f, t.TempDir())
	if !r.Valid {
		t.Errorf("expected valid (warning only), got errors: %v", r.Errors)
	}
	if !containsWarning(r, "duplicate fact") {
		t.Errorf("expected duplicate warning, got warnings: %v", r.Warnings)
	}
}

func TestValidateMinimalWithEmptyCollections(t *testing.T) {
	f := &config.Folio{
		Schema:          1,
		Project:         "Test",
		Repositories:    map[string]string{},
		CrossReferences: []config.CrossReference{},
	}
	r := Validate(f, t.TempDir())
	if !r.Valid {
		t.Errorf("expected valid for empty collections, got errors: %v", r.Errors)
	}
}

func TestValidateSchema3Valid(t *testing.T) {
	f := &config.Folio{
		Schema:  3,
		Project: "Test",
	}
	r := Validate(f, t.TempDir())
	if !r.Valid {
		t.Errorf("expected valid for schema 3, got errors: %v", r.Errors)
	}
}

func TestValidateSchema4Rejected(t *testing.T) {
	f := &config.Folio{
		Schema:  4,
		Project: "Test",
	}
	r := Validate(f, t.TempDir())
	if r.Valid {
		t.Error("expected invalid for schema 4")
	}
	if !containsError(r, "schema version") {
		t.Errorf("expected schema error, got: %v", r.Errors)
	}
}

func TestValidateSchema0Rejected(t *testing.T) {
	f := &config.Folio{
		Schema:  0,
		Project: "Test",
	}
	r := Validate(f, t.TempDir())
	if r.Valid {
		t.Error("expected invalid for schema 0")
	}
	if !containsError(r, "schema version") {
		t.Errorf("expected schema error, got: %v", r.Errors)
	}
}

func TestValidateTypeStatusDesignActive(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "design.md"), []byte("# Design"), 0644)

	f := &config.Folio{
		Schema:  3,
		Project: "Test",
		Sources: []config.Source{
			{Path: "design.md", Type: "design", Status: "active"},
		},
	}
	r := Validate(f, dir)
	if !r.Valid {
		t.Errorf("expected valid for design+active, got errors: %v", r.Errors)
	}
}

func TestValidateTypeStatusPlanDone(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "plan.md"), []byte("# Plan"), 0644)

	f := &config.Folio{
		Schema:  3,
		Project: "Test",
		Sources: []config.Source{
			{Path: "plan.md", Type: "plan", Status: "done"},
		},
	}
	r := Validate(f, dir)
	if !r.Valid {
		t.Errorf("expected valid for plan+done, got errors: %v", r.Errors)
	}
}

func TestValidateTypeStatusTrackActive(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "track.md"), []byte("# Track"), 0644)

	f := &config.Folio{
		Schema:  3,
		Project: "Test",
		Sources: []config.Source{
			{Path: "track.md", Type: "track", Status: "active"},
		},
	}
	r := Validate(f, dir)
	if !r.Valid {
		t.Errorf("expected valid for track+active, got errors: %v", r.Errors)
	}
}

func TestValidateInvalidType(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "foo.md"), []byte("# Foo"), 0644)

	f := &config.Folio{
		Schema:  3,
		Project: "Test",
		Sources: []config.Source{
			{Path: "foo.md", Type: "foo"},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for type 'foo'")
	}
	if !containsError(r, "invalid type 'foo'") {
		t.Errorf("expected type error, got: %v", r.Errors)
	}
}

func TestValidateInvalidStatus(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "design.md"), []byte("# Design"), 0644)

	f := &config.Folio{
		Schema:  3,
		Project: "Test",
		Sources: []config.Source{
			{Path: "design.md", Type: "design", Status: "pending"},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for status 'pending'")
	}
	if !containsError(r, "invalid status 'pending'") {
		t.Errorf("expected status error, got: %v", r.Errors)
	}
}

func TestValidateStatusOnSpike(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "spike.md"), []byte("# Spike"), 0644)

	f := &config.Folio{
		Schema:  3,
		Project: "Test",
		Sources: []config.Source{
			{Path: "spike.md", Type: "spike", Status: "active"},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for status on spike")
	}
	if !containsError(r, "status is only valid on design, plan, or track") {
		t.Errorf("expected tier1 error, got: %v", r.Errors)
	}
}

func TestValidateStatusOnRetro(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "retro.md"), []byte("# Retro"), 0644)

	f := &config.Folio{
		Schema:  3,
		Project: "Test",
		Sources: []config.Source{
			{Path: "retro.md", Type: "retro", Status: "done"},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for status on retro")
	}
	if !containsError(r, "status is only valid on design, plan, or track") {
		t.Errorf("expected tier1 error, got: %v", r.Errors)
	}
}

func TestValidateStatusWithoutType(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "doc.md"), []byte("# Doc"), 0644)

	f := &config.Folio{
		Schema:  3,
		Project: "Test",
		Sources: []config.Source{
			{Path: "doc.md", Status: "active"},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for status without type")
	}
	if !containsError(r, "status requires an explicit type field") {
		t.Errorf("expected type-required error, got: %v", r.Errors)
	}
}

func TestValidateTypeWithoutStatus(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "spike.md"), []byte("# Spike"), 0644)

	f := &config.Folio{
		Schema:  3,
		Project: "Test",
		Sources: []config.Source{
			{Path: "spike.md", Type: "spike"},
		},
	}
	r := Validate(f, dir)
	if !r.Valid {
		t.Errorf("expected valid for type without status, got errors: %v", r.Errors)
	}
}

func TestValidateOmittedTypeAndStatus(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "readme.md"), []byte("# Readme"), 0644)

	f := &config.Folio{
		Schema:  3,
		Project: "Test",
		Sources: []config.Source{
			{Path: "readme.md"},
		},
	}
	r := Validate(f, dir)
	if !r.Valid {
		t.Errorf("expected valid for omitted type and status, got errors: %v", r.Errors)
	}
}

func TestValidateSchema2WithTypeField(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "design.md"), []byte("# Design"), 0644)

	f := &config.Folio{
		Schema:  2,
		Project: "Test",
		Sources: []config.Source{
			{Path: "design.md", Type: "design"},
		},
	}
	r := Validate(f, dir)
	if !r.Valid {
		t.Errorf("expected valid for schema 2 with type field, got errors: %v", r.Errors)
	}
}

func TestValidateTypeOnExternalSource(t *testing.T) {
	f := &config.Folio{
		Schema:  3,
		Project: "Test",
		Sources: []config.Source{
			{External: "github", ID: "repo-1", Type: "design"},
		},
	}
	r := Validate(f, t.TempDir())
	// Type on external sources is silently ignored (no path to validate against)
	// Should not produce type/status errors — only the normal external source validation
	if containsError(r, "invalid type") {
		t.Errorf("should not validate type on external sources, got: %v", r.Errors)
	}
}

func TestValidateMultipleErrorsSameSource(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "bad.md"), []byte("# Bad"), 0644)

	f := &config.Folio{
		Schema:  3,
		Project: "Test",
		Sources: []config.Source{
			{Path: "bad.md", Type: "foo", Status: "pending"},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid")
	}
	typeErr := containsError(r, "invalid type")
	statusErr := containsError(r, "invalid status")
	if !typeErr || !statusErr {
		t.Errorf("expected both type and status errors, got: %v", r.Errors)
	}
}

func containsWarning(r *Result, substr string) bool {
	for _, w := range r.Warnings {
		if strings.Contains(w, substr) {
			return true
		}
	}
	return false
}

func containsError(r *Result, substr string) bool {
	for _, e := range r.Errors {
		if strings.Contains(e, substr) {
			return true
		}
	}
	return false
}
