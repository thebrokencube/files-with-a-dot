package config

// Folio represents the top-level folio.yml document.
type Folio struct {
	Schema          int               `yaml:"schema"`
	Project         string            `yaml:"project"`
	Sources         []Source          `yaml:"sources"`
	Repositories    map[string]string `yaml:"repositories"`
	Targets         map[string]Target `yaml:"targets"`
	CrossReferences []CrossReference  `yaml:"cross_references"`
	Tasks           []string          `yaml:"tasks"`
	Pending         []string          `yaml:"pending"`
	ContextSources  interface{}       `yaml:"context_sources"` // deprecated detection
}

// Source represents a project-level or target-level source entry.
type Source struct {
	Path        string        `yaml:"path"`
	External    string        `yaml:"external"`
	ID          string        `yaml:"id"`
	Notes       string        `yaml:"notes"`
	DerivedFrom []DerivedFrom `yaml:"derived_from"`
}

// DerivedFrom records the provenance of a derived source.
type DerivedFrom struct {
	External string `yaml:"external"`
	URL      string `yaml:"url"`
	Cached   string `yaml:"cached"`
	Notes    string `yaml:"notes"`
}

// Target represents a compilation target.
type Target struct {
	Instructions string   `yaml:"instructions"`
	Transform    string   `yaml:"transform"`
	BlockedBy    []string `yaml:"blocked_by"`
	Sources      []Source `yaml:"sources"`
	Outputs      []Output `yaml:"outputs"`
	Batch        *Batch   `yaml:"batch"`
	Tree         *Tree    `yaml:"tree"`
}

// Output represents a target's output destination.
type Output struct {
	Path     string `yaml:"path"`
	External string `yaml:"external"`
	ID       string `yaml:"id"`
	Field    string `yaml:"field"`
}

// Batch represents a batch target with multiple items.
type Batch struct {
	Description string      `yaml:"description"`
	System      string      `yaml:"system"` // default output.external for items
	Field       string      `yaml:"field"`  // default output.field for items
	Items       []BatchItem `yaml:"items"`
}

// Tree represents a hierarchical target structure (e.g., Jira initiative → projects → epics).
type Tree struct {
	System string   `yaml:"system"` // default external system for all nodes
	Field  string   `yaml:"field"`  // default output field for all nodes
	Root   TreeNode `yaml:"root"`
}

// TreeNode represents a single node in a target tree.
type TreeNode struct {
	ID           string     `yaml:"id"`           // external system ID
	Label        string     `yaml:"label"`        // human-readable name
	File         string     `yaml:"file"`         // linked file path (optional for grouping nodes)
	Transform    string     `yaml:"transform"`    // optional, inherits target default
	Instructions string     `yaml:"instructions"` // optional, per-node compilation instructions
	Children     []TreeNode `yaml:"children"`     // recursive
}

// BatchItem represents a single item in a batch target.
type BatchItem struct {
	ID     string `yaml:"id"`
	Source string `yaml:"source"`
	Output Output `yaml:"output"`
}

// ResolveItemOutput returns the effective output for a batch item by merging
// batch-level defaults. Item-level values take precedence.
func (b *Batch) ResolveItemOutput(item BatchItem) Output {
	out := item.Output
	if out.External == "" && b.System != "" {
		out.External = b.System
	}
	if out.Field == "" && b.Field != "" {
		out.Field = b.Field
	}
	return out
}

// CrossReference represents a fact tracked across multiple locations.
type CrossReference struct {
	Fact          string   `yaml:"fact"`
	SourceOfTruth string   `yaml:"source_of_truth"`
	AlsoAppearsIn []string `yaml:"also_appears_in"`
}

// ValidTransforms is the set of allowed transform values.
var ValidTransforms = map[string]bool{
	"distill":  true,
	"extract":  true,
	"adapt":    true,
	"compose":  true,
	"scaffold": true,
}
