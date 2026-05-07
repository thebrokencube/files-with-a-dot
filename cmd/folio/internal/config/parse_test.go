package config

import (
	"os"
	"testing"
)

func TestParseMinimal(t *testing.T) {
	data := []byte(`
schema: 1
project: "Test"
sources: []
targets: {}
observations: []
`)
	f, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Schema != 1 {
		t.Errorf("schema = %d, want 1", f.Schema)
	}
	if f.Project != "Test" {
		t.Errorf("project = %q, want %q", f.Project, "Test")
	}
	if len(f.Sources) != 0 {
		t.Errorf("sources len = %d, want 0", len(f.Sources))
	}
	if len(f.Targets) != 0 {
		t.Errorf("targets len = %d, want 0", len(f.Targets))
	}
}

func TestParseFull(t *testing.T) {
	data, err := os.ReadFile("../../testdata/full.yml")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	f, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if f.Project != "Full Featured Project" {
		t.Errorf("project = %q", f.Project)
	}

	// Sources
	if len(f.Sources) != 4 {
		t.Fatalf("sources len = %d, want 4", len(f.Sources))
	}
	if f.Sources[0].Path != "README.md" {
		t.Errorf("sources[0].path = %q", f.Sources[0].Path)
	}
	if f.Sources[1].External != "jira" {
		t.Errorf("sources[1].external = %q", f.Sources[1].External)
	}
	if f.Sources[1].ID != "ACME-123" {
		t.Errorf("sources[1].id = %q", f.Sources[1].ID)
	}
	if len(f.Sources[2].DerivedFrom) != 1 {
		t.Fatalf("sources[2].derived_from len = %d, want 1", len(f.Sources[2].DerivedFrom))
	}
	if f.Sources[2].DerivedFrom[0].External != "web" {
		t.Errorf("sources[2].derived_from[0].external = %q", f.Sources[2].DerivedFrom[0].External)
	}
	if f.Sources[3].External != "github" {
		t.Errorf("sources[3].external = %q", f.Sources[3].External)
	}

	// Targets
	if len(f.Targets) != 2 {
		t.Fatalf("targets len = %d, want 2", len(f.Targets))
	}
	summary, ok := f.Targets["summary"]
	if !ok {
		t.Fatal("missing target 'summary'")
	}
	if summary.How != "Condense research into summary" {
		t.Errorf("summary.how = %q", summary.How)
	}
	if len(summary.Sources) != 1 {
		t.Errorf("summary sources len = %d", len(summary.Sources))
	}
	if len(summary.Outputs) != 1 {
		t.Errorf("summary outputs len = %d", len(summary.Outputs))
	}

	jiraUpdate, ok := f.Targets["jira-update"]
	if !ok {
		t.Fatal("missing target 'jira-update'")
	}
	if len(jiraUpdate.Outputs) != 2 {
		t.Errorf("jira-update outputs len = %d", len(jiraUpdate.Outputs))
	}

	// Cross references
	if len(f.CrossReferences) != 1 {
		t.Errorf("cross_references len = %d", len(f.CrossReferences))
	}

	// Observations
	if len(f.Observations) != 2 {
		t.Errorf("observations len = %d, want 2", len(f.Observations))
	}

	// Repositories
	if len(f.Repositories) != 1 {
		t.Errorf("repositories len = %d", len(f.Repositories))
	}
}

func TestParseDeprecatedContextSources(t *testing.T) {
	data, err := os.ReadFile("../../testdata/deprecated.yml")
	if err != nil {
		t.Fatalf("reading fixture: %v", err)
	}
	f, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.ContextSources == nil {
		t.Error("expected context_sources to be non-nil for deprecated detection")
	}
}

func TestParseInvalidYAML(t *testing.T) {
	data := []byte(`{invalid: yaml: [}`)
	_, err := Parse(data)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestParseNormalizesNilSlices(t *testing.T) {
	data := []byte(`
schema: 1
project: "Bare"
`)
	f, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Sources == nil {
		t.Error("Sources should be initialized, got nil")
	}
	if f.Targets == nil {
		t.Error("Targets should be initialized, got nil")
	}
	if f.CrossReferences == nil {
		t.Error("CrossReferences should be initialized, got nil")
	}
	if f.Repositories == nil {
		t.Error("Repositories should be initialized, got nil")
	}
	if f.Observations == nil {
		t.Error("Observations should be initialized, got nil")
	}
}

func TestParseObservationsField(t *testing.T) {
	data := []byte(`
schema: 2
project: "Test"
sources: []
targets: {}
observations:
  - "idea 1"
  - "idea 2"
`)
	f, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Observations) != 2 {
		t.Errorf("observations len = %d, want 2", len(f.Observations))
	}
}

func TestParseRejectsUnknownTopLevelKeys(t *testing.T) {
	data := []byte(`
schema: 1
project: "Test"
targetz:
  my-target:
    how: "Test"
`)
	_, err := Parse(data)
	if err == nil {
		t.Error("expected error for unknown key 'targetz', got nil")
	}
}

func TestParseRejectsUnknownNestedKeys(t *testing.T) {
	data := []byte(`
schema: 1
project: "Test"
targets:
  my-target:
    transfrom: distill
    outputs: []
`)
	_, err := Parse(data)
	if err == nil {
		t.Error("expected error for unknown nested key 'transfrom', got nil")
	}
}

func TestParseBatchTarget(t *testing.T) {
	data := []byte(`
schema: 1
project: "Batch Test"
sources: []
targets:
  batch-target:
    how: "Batch items"
    sources: []
    outputs: []
    batch:
      description: "Pattern"
      items:
        - id: "item-1"
          source: source1.md
          output:
            external: jira
            id: "PROJ-1"
            field: description
observations: []
`)
	f, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bt := f.Targets["batch-target"]
	if bt.Batch == nil {
		t.Fatal("batch is nil")
	}
	if len(bt.Batch.Items) != 1 {
		t.Fatalf("batch items len = %d, want 1", len(bt.Batch.Items))
	}
	if bt.Batch.Items[0].ID != "item-1" {
		t.Errorf("batch item id = %q", bt.Batch.Items[0].ID)
	}
}

func TestParseBatchWithDefaults(t *testing.T) {
	data := []byte(`
schema: 1
project: "Batch Defaults"
sources: []
targets:
  jira-epics:
    how: "Jira epics"
    sources: []
    outputs:
      - path: compiled/epics-manifest.md
    batch:
      system: jira
      field: description
      items:
        - id: "PROJ-100"
          source: epics/first.md
observations: []
`)
	f, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	bt := f.Targets["jira-epics"]
	if bt.Batch == nil {
		t.Fatal("batch is nil")
	}
	if bt.Batch.System != "jira" {
		t.Errorf("batch.system = %q, want jira", bt.Batch.System)
	}
	if bt.Batch.Field != "description" {
		t.Errorf("batch.field = %q, want description", bt.Batch.Field)
	}
}
