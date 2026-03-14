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
	if len(f.Targets) != 3 {
		t.Fatalf("targets len = %d, want 3", len(f.Targets))
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
	if len(jiraUpdate.BlockedBy) != 1 || jiraUpdate.BlockedBy[0] != "summary" {
		t.Errorf("jira-update.blocked_by = %v", jiraUpdate.BlockedBy)
	}
	if len(jiraUpdate.Outputs) != 2 {
		t.Errorf("jira-update outputs len = %d", len(jiraUpdate.Outputs))
	}

	featureBranch, ok := f.Targets["feature-branch"]
	if !ok {
		t.Fatal("missing target 'feature-branch'")
	}
	if featureBranch.Branch != "feat/my-feature" {
		t.Errorf("feature-branch.branch = %q, want %q", featureBranch.Branch, "feat/my-feature")
	}
	if featureBranch.PR != "#42" {
		t.Errorf("feature-branch.pr = %q, want %q", featureBranch.PR, "#42")
	}
	if featureBranch.Tree == nil {
		t.Fatal("feature-branch.tree is nil")
	}
	if featureBranch.Tree.Root.Sync != "pull" {
		t.Errorf("feature-branch tree root sync = %q, want %q", featureBranch.Tree.Root.Sync, "pull")
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

func TestNormalizeTasksToObservations(t *testing.T) {
	data := []byte(`
schema: 1
project: "Test"
sources: []
targets: {}
tasks:
  - "task 1"
  - "task 2"
pending: []
`)
	f, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Observations) != 2 {
		t.Errorf("observations len = %d, want 2 (auto-upgraded from tasks)", len(f.Observations))
	}
	if f.Observations[0] != "task 1" {
		t.Errorf("observations[0] = %q, want %q", f.Observations[0], "task 1")
	}
}

func TestNormalizeObservationsNotOverwritten(t *testing.T) {
	data := []byte(`
schema: 2
project: "Test"
sources: []
targets: {}
tasks:
  - "old task"
observations:
  - "explicit obs"
`)
	f, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Observations) != 1 {
		t.Errorf("observations len = %d, want 1 (should not merge tasks when observations present)", len(f.Observations))
	}
	if f.Observations[0] != "explicit obs" {
		t.Errorf("observations[0] = %q, want %q", f.Observations[0], "explicit obs")
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
    how: "Jira initiative hierarchy"
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
        how: "Overview and plan tables"
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
	if root.How != "Overview and plan tables" {
		t.Errorf("root.how = %q, want %q", root.How, "Overview and plan tables")
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

	// Child without how should have empty string
	if projA.How != "" {
		t.Errorf("child[0].how = %q, want empty", projA.How)
	}
}

func TestParseTreeNodeWithSync(t *testing.T) {
	data := []byte(`
schema: 1
project: "Sync Test"
sources: []
targets:
  tree-target:
    how: "Tree with sync"
    sources: []
    outputs: []
    tree:
      system: jira
      field: description
      root:
        id: "ROOT-1"
        label: "Root"
        sync: pull
        children:
          - id: "CHILD-1"
            label: "Child"
            sync: both
tasks: []
pending: []
`)
	f, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tt := f.Targets["tree-target"]
	if tt.Tree == nil {
		t.Fatal("tree is nil")
	}
	if tt.Tree.Root.Sync != "pull" {
		t.Errorf("root.sync = %q, want pull", tt.Tree.Root.Sync)
	}
	if len(tt.Tree.Root.Children) != 1 {
		t.Fatalf("root children len = %d, want 1", len(tt.Tree.Root.Children))
	}
	if tt.Tree.Root.Children[0].Sync != "both" {
		t.Errorf("child.sync = %q, want both", tt.Tree.Root.Children[0].Sync)
	}
}

func TestParseTargetWithBranchAndPR(t *testing.T) {
	data := []byte(`
schema: 1
project: "Branch Test"
sources: []
targets:
  my-target:
    how: "Target with branch"
    branch: "feat/my-branch"
    pr: "#123"
    sources: []
    outputs: []
tasks: []
pending: []
`)
	f, err := Parse(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tgt := f.Targets["my-target"]
	if tgt.Branch != "feat/my-branch" {
		t.Errorf("branch = %q, want %q", tgt.Branch, "feat/my-branch")
	}
	if tgt.PR != "#123" {
		t.Errorf("pr = %q, want %q", tgt.PR, "#123")
	}
}
