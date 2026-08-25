// Package agentskills provides Layer 1 validation for the Agent Skills standard (agentskills.io).
package agentskills

// Severity indicates how a validation issue should be treated.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// ValidationResult represents a single validation issue.
type ValidationResult struct {
	CheckID     string   // e.g. "skill-exists", "skill-frontmatter"
	Severity    Severity // error or warning
	Message     string   // what's wrong
	File        string   // file path (relative to skill dir)
	Line        int      // line number (0 = unknown)
	Remediation string   // how to fix
}

// SkillFrontmatter represents parsed YAML frontmatter from SKILL.md.
type SkillFrontmatter struct {
	Name          string `yaml:"name"`
	Description   string `yaml:"description"`
	License       string `yaml:"license"`
	AllowedTools  any    `yaml:"allowed-tools"`
	UserInvocable any    `yaml:"user_invocable"`
	ArgumentHint  string `yaml:"argument-hint"`
	Compatibility any    `yaml:"compatibility"`
	Metadata      any    `yaml:"metadata"`
	Version       string `yaml:"version"`

	// Dendrik extensions (Layer 2 — not validated here but parsed for passthrough)
	Trigger  any `yaml:"trigger"`
	SkipWhen any `yaml:"skip_when"`
	Related  any `yaml:"related"`
}

// KnownFields is the allowlist of frontmatter field names recognized by the Agent Skills spec
// plus dendrik extensions. Used by skill-extra-fields to detect unexpected fields.
var KnownFields = map[string]bool{
	"name":           true,
	"description":    true,
	"license":        true,
	"allowed-tools":  true,
	"user_invocable": true,
	"argument-hint":  true,
	"compatibility":  true,
	"metadata":       true,
	"version":        true,
	// dendrik extensions
	"trigger":   true,
	"skip_when": true,
	"related":   true,
}

var PortableFields = map[string]bool{
	"name": true, "description": true, "license": true, "compatibility": true,
	"metadata": true, "allowed-tools": true,
}
