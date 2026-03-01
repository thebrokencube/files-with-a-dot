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
		Schema:  2,
		Project: "Test",
	}
	r := Validate(f, t.TempDir())
	if r.Valid {
		t.Error("expected invalid for schema 2")
	}
	if !containsError(r, "schema version") {
		t.Errorf("expected schema error, got: %v", r.Errors)
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

func TestValidateTargetMissingTransform(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"my-target": {
				Instructions: "Test",
				Outputs:      []config.Output{{Path: "compiled/out.md"}},
			},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for missing transform")
	}
	if !containsError(r, "missing required field: transform") {
		t.Errorf("expected transform error, got: %v", r.Errors)
	}
}

func TestValidateTargetInvalidTransform(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"my-target": {
				Instructions: "Test",
				Transform:    "bogus",
				Outputs:      []config.Output{{Path: "compiled/out.md"}},
			},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for bad transform")
	}
	if !containsError(r, "invalid transform") {
		t.Errorf("expected transform error, got: %v", r.Errors)
	}
}

func TestValidatePrecompileRule(t *testing.T) {
	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"ext-only": {
				Instructions: "External only",
				Transform:    "distill",
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

func TestValidateBlockedByNonexistent(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"my-target": {
				Instructions: "Test",
				Transform:    "distill",
				BlockedBy:    []string{"nonexistent"},
				Outputs:      []config.Output{{Path: "compiled/out.md"}},
			},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for nonexistent blocked_by")
	}
	if !containsError(r, "non-existent target") {
		t.Errorf("expected blocked_by error, got: %v", r.Errors)
	}
}

func TestValidateExternalOutputMissingID(t *testing.T) {
	dir := t.TempDir()

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"my-target": {
				Instructions: "Test",
				Transform:    "distill",
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
				Instructions: "First",
				Transform:    "distill",
				Outputs:      []config.Output{{Path: "compiled/same.md"}},
			},
			"second": {
				Instructions: "Second",
				Transform:    "distill",
				Outputs:      []config.Output{{Path: "compiled/same.md"}},
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

func TestValidateCycleDetection(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"a": {
				Transform: "distill",
				BlockedBy: []string{"b"},
				Outputs:   []config.Output{{Path: "compiled/a.md"}},
			},
			"b": {
				Transform: "distill",
				BlockedBy: []string{"a"},
				Outputs:   []config.Output{{Path: "compiled/b.md"}},
			},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for cycle")
	}
	if !containsError(r, "cycle") {
		t.Errorf("expected cycle error, got: %v", r.Errors)
	}
}

func TestValidateAllTransforms(t *testing.T) {
	transforms := []string{"distill", "extract", "adapt", "compose"}
	for _, tr := range transforms {
		dir := t.TempDir()
		os.MkdirAll(filepath.Join(dir, "compiled"), 0755)

		f := &config.Folio{
			Schema:  1,
			Project: "Test",
			Targets: map[string]config.Target{
				"target": {
					Transform: tr,
					Outputs:   []config.Output{{Path: "compiled/out.md"}},
				},
			},
		}
		r := Validate(f, dir)
		if containsError(r, "transform") {
			t.Errorf("transform %q should be valid, got errors: %v", tr, r.Errors)
		}
	}
}

func TestValidateTreeMissingSystem(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "root.md"), []byte("# Root"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"tree-target": {
				Transform: "compose",
				Outputs:   []config.Output{{Path: "compiled/manifest.md"}},
				Tree: &config.Tree{
					Root: config.TreeNode{
						ID:   "ROOT-1",
						File: "root.md",
					},
				},
			},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for tree missing system")
	}
	if !containsError(r, "tree missing required field: system") {
		t.Errorf("expected system error, got: %v", r.Errors)
	}
}

func TestValidateTreeNodeMissingID(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "root.md"), []byte("# Root"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"tree-target": {
				Transform: "compose",
				Outputs:   []config.Output{{Path: "compiled/manifest.md"}},
				Tree: &config.Tree{
					System: "jira",
					Root: config.TreeNode{
						File: "root.md",
					},
				},
			},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for tree node missing id")
	}
	if !containsError(r, "missing required field: id") {
		t.Errorf("expected id error, got: %v", r.Errors)
	}
}

func TestValidateTreeNodeOptionalSource(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"tree-target": {
				Transform: "compose",
				Outputs:   []config.Output{{Path: "compiled/manifest.md"}},
				Tree: &config.Tree{
					System: "jira",
					Root: config.TreeNode{
						ID: "ROOT-1",
					},
				},
			},
		},
	}
	r := Validate(f, dir)
	if !r.Valid {
		t.Errorf("expected valid for sourceless tree node, got errors: %v", r.Errors)
	}
}

func TestValidateTreeNodeWithDescription(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "root.md"), []byte("# Root"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"tree-target": {
				Transform: "compose",
				Outputs:   []config.Output{{Path: "compiled/manifest.md"}},
				Tree: &config.Tree{
					System: "jira",
					Root: config.TreeNode{
						ID:           "ROOT-1",
						File:         "root.md",
						Instructions: "Include overview and plan tables",
					},
				},
			},
		},
	}
	r := Validate(f, dir)
	if !r.Valid {
		t.Errorf("expected valid tree node with instructions, got errors: %v", r.Errors)
	}
}

func TestValidateTreeNodeSourceNotFound(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"tree-target": {
				Transform: "compose",
				Outputs:   []config.Output{{Path: "compiled/manifest.md"}},
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
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for tree node source not found")
	}
	if !containsError(r, "file not found") {
		t.Errorf("expected file not found error, got: %v", r.Errors)
	}
}

func TestValidateTreeDuplicateNodeID(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "root.md"), []byte("# Root"), 0644)
	os.WriteFile(filepath.Join(dir, "child.md"), []byte("# Child"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"tree-target": {
				Transform: "compose",
				Outputs:   []config.Output{{Path: "compiled/manifest.md"}},
				Tree: &config.Tree{
					System: "jira",
					Root: config.TreeNode{
						ID:   "SAME-ID",
						File: "root.md",
						Children: []config.TreeNode{
							{ID: "SAME-ID", File: "child.md"},
						},
					},
				},
			},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for duplicate tree node ID")
	}
	if !containsError(r, "duplicate tree node ID") {
		t.Errorf("expected duplicate ID error, got: %v", r.Errors)
	}
}

func TestValidateTreeNodeInvalidTransform(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "root.md"), []byte("# Root"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"tree-target": {
				Transform: "compose",
				Outputs:   []config.Output{{Path: "compiled/manifest.md"}},
				Tree: &config.Tree{
					System: "jira",
					Root: config.TreeNode{
						ID:        "ROOT-1",
						File:      "root.md",
						Transform: "bogus",
					},
				},
			},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for bad tree node transform")
	}
	if !containsError(r, "invalid transform") {
		t.Errorf("expected transform error, got: %v", r.Errors)
	}
}

func TestValidateTreeBatchMutualExclusion(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "root.md"), []byte("# Root"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"both": {
				Transform: "compose",
				Outputs:   []config.Output{{Path: "compiled/manifest.md"}},
				Tree: &config.Tree{
					System: "jira",
					Root: config.TreeNode{
						ID:   "ROOT-1",
						File: "root.md",
					},
				},
				Batch: &config.Batch{
					Items: []config.BatchItem{
						{ID: "item-1", Source: "root.md"},
					},
				},
			},
		},
	}
	r := Validate(f, dir)
	if r.Valid {
		t.Error("expected invalid for tree+batch on same target")
	}
	if !containsError(r, "mutually exclusive") {
		t.Errorf("expected mutual exclusion error, got: %v", r.Errors)
	}
}

func TestValidateTreeValid(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "root.md"), []byte("# Root"), 0644)
	os.WriteFile(filepath.Join(dir, "child.md"), []byte("# Child"), 0644)

	f := &config.Folio{
		Schema:  1,
		Project: "Test",
		Targets: map[string]config.Target{
			"tree-target": {
				Transform: "compose",
				Outputs:   []config.Output{{Path: "compiled/manifest.md"}},
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
	r := Validate(f, dir)
	if !r.Valid {
		t.Errorf("expected valid tree target, got errors: %v", r.Errors)
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
				Transform: "compose",
				Outputs:   []config.Output{{Path: "compiled/manifest.md"}},
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
				Transform: "compose",
				Outputs:   []config.Output{{Path: "compiled/manifest.md"}},
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
				Transform: "compose",
				Outputs:   []config.Output{{Path: "compiled/manifest.md"}},
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

func containsError(r *Result, substr string) bool {
	for _, e := range r.Errors {
		if strings.Contains(e, substr) {
			return true
		}
	}
	return false
}
