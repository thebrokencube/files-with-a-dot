package config

// Folio represents the top-level folio.yml document.
type Folio struct {
	Schema          int               `yaml:"schema"`
	Project         string            `yaml:"project"`
	Sources         []Source          `yaml:"sources"`
	Repositories    map[string]string `yaml:"repositories"`
	Targets         map[string]Target `yaml:"targets"`
	CrossReferences []CrossReference  `yaml:"cross_references"`
	Observations    []string          `yaml:"observations"`
	ContextSources  interface{}       `yaml:"context_sources"` // deprecated detection
}

// Source represents a project-level or target-level source entry.
type Source struct {
	Path        string        `yaml:"path"`
	External    string        `yaml:"external"`
	ID          string        `yaml:"id"`
	Notes       string        `yaml:"notes"`
	DerivedFrom []DerivedFrom `yaml:"derived_from"`
	DependsOn   []string      `yaml:"depends_on"`
}

// DerivedFrom records the provenance of a derived source.
type DerivedFrom struct {
	External string `yaml:"external"`
	ID       string `yaml:"id"`
	URL      string `yaml:"url"`
	Cached   string `yaml:"cached"`
	Notes    string `yaml:"notes"`
}

// Target represents a compilation target.
type Target struct {
	How          string   `yaml:"how"`
	Instructions string   `yaml:"instructions"` // deprecated: use how
	Transform    string   `yaml:"transform"`    // deprecated: ignored
	Sources      []Source `yaml:"sources"`
	Outputs      []Output `yaml:"outputs"`
	Batch        *Batch   `yaml:"batch"`
	Forest       *Forest  `yaml:"forest"`
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

// Forest represents a jf-managed forest target. The hierarchy lives on disk
// (directory layout + YAML frontmatter); folio stores per-node compilation
// instructions and a pointer to the forest root.
type Forest struct {
	Root         string            `yaml:"root"`          // path to forest root directory (relative to folio.yml)
	HowDefault   string            `yaml:"how_default"`   // default compilation instruction for all nodes
	HowOverrides map[string]string `yaml:"how_overrides"` // per-node overrides keyed by Jira key
}

// BatchItem represents a single item in a batch target.
type BatchItem struct {
	ID     string `yaml:"id"`
	Label  string `yaml:"label"` // optional human-readable name
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
