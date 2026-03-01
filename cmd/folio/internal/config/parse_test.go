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
tasks: []
pending: []
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
	if summary.Transform != "distill" {
		t.Errorf("summary.transform = %q", summary.Transform)
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
	if len(jiraUpdate.BlockedBy) != 1 || jiraUpdate.BlockedBy[0] != "summary" {
		t.Errorf("jira-update.blocked_by = %v", jiraUpdate.BlockedBy)
	}
	if len(jiraUpdate.Outputs) != 2 {
		t.Errorf("jira-update outputs len = %d", len(jiraUpdate.Outputs))
	}

	// Cross references
	if len(f.CrossReferences) != 1 {
		t.Errorf("cross_references len = %d", len(f.CrossReferences))
	}

	// Tasks and pending
	if len(f.Tasks) != 1 {
		t.Errorf("tasks len = %d", len(f.Tasks))
	}
	if len(f.Pending) != 1 {
		t.Errorf("pending len = %d", len(f.Pending))
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
	if f.Tasks == nil {
		t.Error("Tasks should be initialized, got nil")
	}
	if f.Pending == nil {
		t.Error("Pending should be initialized, got nil")
	}
	if f.CrossReferences == nil {
		t.Error("CrossReferences should be initialized, got nil")
	}
	if f.Repositories == nil {
		t.Error("Repositories should be initialized, got nil")
	}
}

func TestParseRejectsUnknownTopLevelKeys(t *testing.T) {
	data := []byte(`
schema: 1
project: "Test"
targetz:
  my-target:
    transform: distill
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
    instructions: "Batch items"
    transform: distill
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
tasks: []
pending: []
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
    instructions: "Jira epics"
    transform: adapt
    sources: []
    outputs:
      - path: compiled/epics-manifest.md
    batch:
      system: jira
      field: description
      items:
        - id: "PROJ-100"
          source: epics/first.md
tasks: []
pending: []
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

func TestParseTreeTarget(t *testing.T) {
	data := []byte(`
schema: 1
project: "Tree Test"
sources: []
targets:
  initiative:
    instructions: "Jira initiative hierarchy"
    transform: compose
    sources: []
    outputs:
      - path: compiled/initiative-manifest.md
    tree:
      system: jira
      field: description
      root:
        id: "PROJ-1"
        label: "Initiative"
        file: initiative.md
        instructions: "Overview and plan tables"
        children:
          - id: "PROJ-10"
            label: "Project A"
            file: projects/a.md
            children:
              - id: "PROJ-100"
                label: "Epic 1"
                file: epics/e1.md
          - id: "PROJ-20"
            label: "Project B"
            file: projects/b.md
            transform: adapt
tasks: []
pending: []
`)
	f, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tt := f.Targets["initiative"]
	if tt.Tree == nil {
		t.Fatal("tree is nil")
	}
	if tt.Tree.System != "jira" {
		t.Errorf("tree.system = %q, want jira", tt.Tree.System)
	}
	if tt.Tree.Field != "description" {
		t.Errorf("tree.field = %q, want description", tt.Tree.Field)
	}

	root := tt.Tree.Root
	if root.ID != "PROJ-1" {
		t.Errorf("root.id = %q, want PROJ-1", root.ID)
	}
	if root.Label != "Initiative" {
		t.Errorf("root.label = %q, want Initiative", root.Label)
	}
	if root.File != "initiative.md" {
		t.Errorf("root.file = %q, want initiative.md", root.File)
	}
	if root.Instructions != "Overview and plan tables" {
		t.Errorf("root.instructions = %q, want %q", root.Instructions, "Overview and plan tables")
	}
	if len(root.Children) != 2 {
		t.Fatalf("root children len = %d, want 2", len(root.Children))
	}

	// First child with nested grandchild
	projA := root.Children[0]
	if projA.ID != "PROJ-10" {
		t.Errorf("child[0].id = %q, want PROJ-10", projA.ID)
	}
	if len(projA.Children) != 1 {
		t.Fatalf("child[0] children len = %d, want 1", len(projA.Children))
	}
	if projA.Children[0].ID != "PROJ-100" {
		t.Errorf("grandchild.id = %q, want PROJ-100", projA.Children[0].ID)
	}

	// Child without instructions should have empty string
	if projA.Instructions != "" {
		t.Errorf("child[0].instructions = %q, want empty", projA.Instructions)
	}

	// Second child with transform override
	projB := root.Children[1]
	if projB.Transform != "adapt" {
		t.Errorf("child[1].transform = %q, want adapt", projB.Transform)
	}
}
