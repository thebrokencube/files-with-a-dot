package conventions

// Severity indicates how a lint violation should be treated.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Scope indicates whether a check applies universally or only to dendrik tools.
type Scope string

const (
	ScopeUniversal Scope = "universal"
	ScopeDendrik   Scope = "dendrik"
)

// Layer identifies which linter owns a check.
type Layer string

const (
	LayerGo     Layer = "go"
	LayerSkill  Layer = "skill"
	LayerBridge Layer = "bridge"
)

// ContractEntry describes a single lint check in the dendrik tool contract.
type ContractEntry struct {
	ID          string
	Layer       Layer
	Scope       Scope
	Severity    Severity
	Summary     string // one-line description
	Rationale   string // shown by --explain
	Remediation string // shown in lint output
}

// Contract is the complete dendrik tool contract.
var Contract = []ContractEntry{
	// --- Go Layer ---
	{
		ID: "go-mod-linked", Layer: LayerGo, Scope: ScopeDendrik, Severity: SeverityError,
		Summary:     "go.mod exists and go.work links this tool to pkg/dendrik",
		Rationale:   "Every dendrik tool is a Go module that imports the shared library. Without go.mod the tool can't build; without the go.work link it can't resolve pkg/dendrik locally.",
		Remediation: "Create go.mod with `go mod init` and add a `use` entry for this tool in the root go.work file.",
	},
	{
		ID: "main-dispatch", Layer: LayerGo, Scope: ScopeDendrik, Severity: SeverityError,
		Summary:     "main.go has func main() with at least one os.Exit(run*(...)) call",
		Rationale:   "The run-function pattern isolates command logic from process lifecycle. os.Exit in main ensures defer runs in run functions and exit codes propagate correctly.",
		Remediation: "In main.go, ensure func main() delegates to run*() functions via os.Exit(run*(...)).",
	},
	{
		ID: "cmd-file-exists", Layer: LayerGo, Scope: ScopeDendrik, Severity: SeverityError,
		Summary:     "At least one cmd_*.go file exists",
		Rationale:   "The cmd_*.go naming convention makes commands discoverable and separates command implementations from shared code.",
		Remediation: "Create at least one file matching cmd_*.go (e.g., cmd_lint.go) with a run*() function.",
	},
	{
		ID: "makefile-targets", Layer: LayerGo, Scope: ScopeDendrik, Severity: SeverityError,
		Summary:     "Makefile exists with build, test, check targets",
		Rationale:   "Standardized Makefile targets provide a uniform build interface across all dendrik tools.",
		Remediation: "Create a Makefile with `build`, `test`, and `check` targets. See cmd/jf/Makefile for reference.",
	},
	{
		ID: "readme-exists", Layer: LayerGo, Scope: ScopeDendrik, Severity: SeverityError,
		Summary:     "README.md exists in tool directory",
		Rationale:   "README.md is the primary discovery surface for both humans and agents browsing the repository.",
		Remediation: "Create README.md in the tool directory (cmd/*/).",
	},
	{
		ID: "readme-sections", Layer: LayerGo, Scope: ScopeDendrik, Severity: SeverityWarning,
		Summary:     "README.md contains required sections: Install, Quick Start, Commands, Code Structure",
		Rationale:   "Standard sections ensure minimum documentation quality and provide predictable navigation for readers.",
		Remediation: "Add missing ## sections to README.md: ## Install, ## Quick Start, ## Commands, ## Code Structure.",
	},
	{
		ID: "claude-md-exists", Layer: LayerGo, Scope: ScopeDendrik, Severity: SeverityWarning,
		Summary:     "CLAUDE.md exists in tool directory",
		Rationale:   "CLAUDE.md provides build, test, and code convention context for developers modifying the tool. Claude Code auto-loads it when working in the directory.",
		Remediation: "Create CLAUDE.md with standardized skeleton: Build, Test, Binary Distribution, Code Conventions, Deep Context.",
	},
	{
		ID: "docs-naming", Layer: LayerGo, Scope: ScopeDendrik, Severity: SeverityError,
		Summary:     "All files in docs/ match numbered kebab-case pattern (NN-name.md)",
		Rationale:   "Numbered kebab-case naming (01-getting-started.md, 02-workflows.md) provides consistent reading order and prevents renaming when files are added.",
		Remediation: "Rename docs/ files to match the pattern: NN-kebab-case.md (e.g., 01-getting-started.md, 03-reference.md).",
	},
	{
		ID: "docs-getting-started", Layer: LayerGo, Scope: ScopeDendrik, Severity: SeverityWarning,
		Summary:     "docs/01-getting-started.md exists",
		Rationale:   "Every tool with docs/ should have a getting-started guide as the entry point for new users.",
		Remediation: "Create docs/01-getting-started.md with prerequisites, quick start, and core concepts.",
	},
	{
		ID: "readme-doc-links", Layer: LayerGo, Scope: ScopeDendrik, Severity: SeverityError,
		Summary:     "Links in README.md Documentation section resolve to existing files",
		Rationale:   "Broken documentation links in README.md prevent readers from finding detailed docs.",
		Remediation: "Fix broken links in the ## Documentation section of README.md — ensure referenced files exist.",
	},
	{
		ID: "version-flag", Layer: LayerGo, Scope: ScopeUniversal, Severity: SeverityWarning,
		Summary:     "main.go handles a --version flag (distinct from the version subcommand)",
		Rationale:   "A --version flag (with -V) is a near-universal CLI convention (clig.dev, 12-factor CLI) for scriptable version reporting. A `version` subcommand alone does not satisfy the flag form that users and agents expect, so dendrik tools should handle both.",
		Remediation: "In main()'s dispatch, fold the flag forms into the version case: `case \"version\", \"--version\", \"-V\":` printing the version and exiting 0.",
	},

	// --- Skill Layer ---
	{
		ID: "skill-exists", Layer: LayerSkill, Scope: ScopeUniversal, Severity: SeverityError,
		Summary:     "SKILL.md exists at expected path",
		Rationale:   "SKILL.md is the agent discovery surface. Without it, agents cannot discover or invoke the tool.",
		Remediation: "Create skill/SKILL.md with YAML frontmatter containing at minimum `name` and `description`.",
	},
	{
		ID: "skill-frontmatter", Layer: LayerSkill, Scope: ScopeUniversal, Severity: SeverityError,
		Summary:     "Valid YAML frontmatter with name (1-64 chars, lowercase+hyphens) and description (1-1024 chars)",
		Rationale:   "The Agent Skills standard requires name and description for tool discovery. Name constraints ensure cross-platform compatibility.",
		Remediation: "Add valid `name:` (lowercase, hyphens, 1-64 chars) and `description:` (1-1024 chars) fields to SKILL.md frontmatter.",
	},
	{
		ID: "skill-extra-fields", Layer: LayerSkill, Scope: ScopeUniversal, Severity: SeverityWarning,
		Summary:     "No unexpected frontmatter fields outside the Agent Skills spec",
		Rationale:   "Extension fields outside the spec may cause issues with strict validators. Warning severity because real-world skills commonly use extension fields.",
		Remediation: "Remove or document unexpected frontmatter fields. Known spec fields: name, description, version, compatibility, metadata, user_invocable, argument-hint.",
	},
	{
		ID: "skill-links", Layer: LayerSkill, Scope: ScopeUniversal, Severity: SeverityError,
		Summary:     "Standard markdown link references in SKILL.md resolve to existing files",
		Rationale:   "Broken links in SKILL.md prevent agents from accessing referenced documentation.",
		Remediation: "Fix broken [text](path) links in SKILL.md — ensure referenced files exist at the specified paths.",
	},
	{
		ID: "ref-naming", Layer: LayerSkill, Scope: ScopeUniversal, Severity: SeverityWarning,
		Summary:     "Reference files in references/ follow kebab-case naming",
		Rationale:   "Consistent naming conventions improve discoverability and cross-platform compatibility.",
		Remediation: "Rename reference files to kebab-case (lowercase, hyphens between words, e.g., contract-checks.md).",
	},
	{
		ID: "skill-size", Layer: LayerSkill, Scope: ScopeUniversal, Severity: SeverityError,
		Summary:     "SKILL.md body does not exceed 500 lines (token estimate warning at ~5000 tokens)",
		Rationale:   "Oversized skill files consume excessive context window. The 500-line limit keeps skills focused; token estimate provides additional guidance.",
		Remediation: "Move detailed content to reference files in references/ and link from SKILL.md body.",
	},
	{
		ID: "argument-hint", Layer: LayerSkill, Scope: ScopeDendrik, Severity: SeverityError,
		Summary:     "If user_invocable: true, then argument-hint is present",
		Rationale:   "User-invocable skills need argument hints so users know what parameters to provide.",
		Remediation: "Add `argument-hint:` field to SKILL.md frontmatter (e.g., `argument-hint: \"<command> [flags]\"`).",
	},
	{
		ID: "arrow-refs", Layer: LayerSkill, Scope: ScopeDendrik, Severity: SeverityError,
		Summary:     "All -> arrow references in SKILL.md and references/*.md resolve to existing files",
		Rationale:   "Arrow references (-> Read path) are dendrik's convention for cross-file navigation. Broken arrows prevent agents from following documentation chains.",
		Remediation: "Fix broken -> references — ensure the referenced file path exists relative to the skill directory.",
	},
	{
		ID: "activation-guidance", Layer: LayerSkill, Scope: ScopeDendrik, Severity: SeverityWarning,
		Summary:     "Description includes activation guidance (\"use when\", \"for tasks that\")",
		Rationale:   "Activation guidance helps agents decide when to invoke the skill. Without it, routing relies on name matching alone.",
		Remediation: "Add activation guidance to the description field (e.g., \"Use when building...\", \"For tasks that...\").",
	},
	{
		ID: "activation-metadata", Layer: LayerSkill, Scope: ScopeDendrik, Severity: SeverityError,
		Summary:     "If trigger/skip_when/related fields present, they are valid (non-empty, correct types)",
		Rationale:   "Conditional activation metadata must be well-formed for agent routing to work correctly.",
		Remediation: "Ensure trigger, skip_when, and related fields are non-empty strings or string arrays when present.",
	},

	// --- Bridge Layer ---
	{
		ID: "dendrik-import", Layer: LayerBridge, Scope: ScopeDendrik, Severity: SeverityError,
		Summary:     "At least one .go file imports pkg/dendrik",
		Rationale:   "A dendrik tool must import the shared library. Without this import, it's a standalone Go binary, not a dendrik tool.",
		Remediation: "Import github.com/thebrokencube/files-with-a-dot/pkg/dendrik in at least one .go file.",
	},
	{
		ID: "exit-constants", Layer: LayerBridge, Scope: ScopeDendrik, Severity: SeverityError,
		Summary:     "No bare integer returns in cmd_*.go; no os.Exit() outside main.go",
		Rationale:   "Bare returns (return 0, return 1) bypass the exit code contract. os.Exit outside main prevents defer cleanup.",
		Remediation: "Replace bare `return 0` with `return dendrik.ExitOK`, `return 1` with `return dendrik.ExitUserError`, etc. Move os.Exit calls to main.go only.",
	},
	{
		ID: "json-output", Layer: LayerBridge, Scope: ScopeDendrik, Severity: SeverityError,
		Summary:     "If --json flag exists, at least one code path uses dendrik.WriteResult or dendrik.WriteError",
		Rationale:   "A --json flag that produces no structured output is a broken contract with agents.",
		Remediation: "Add dendrik.WriteResult or dendrik.WriteError calls in commands that register a --json flag.",
	},
	{
		ID: "go-work-sync", Layer: LayerBridge, Scope: ScopeUniversal, Severity: SeverityError,
		Summary:     "go.work use entries match cmd/*/ directories with go.mod (symmetric difference)",
		Rationale:   "go.work and the filesystem must agree. Missing entries cause build failures; stale entries cause confusion.",
		Remediation: "Update go.work `use` block to match cmd/*/ directories that contain go.mod. Remove stale entries, add missing ones.",
	},
	{
		ID: "symlink-entries", Layer: LayerBridge, Scope: ScopeDendrik, Severity: SeverityError,
		Summary:     "symlink_map.txt has entries for binary and skill directory",
		Rationale:   "Without symlink_map.txt entries, `dot sync` won't install the binary or register the skill.",
		Remediation: "Add symlink_map.txt entries for the binary (e.g., cmd/jf/jf -> ~/.local/bin/jf) and skill directory.",
	},
	{
		ID: "makefile-gofiles", Layer: LayerBridge, Scope: ScopeDendrik, Severity: SeverityWarning,
		Summary:     "Makefile GOFILES find path includes ../../pkg/dendrik",
		Rationale:   "Without the shared library in the find path, make won't rebuild when pkg/dendrik changes.",
		Remediation: "Update Makefile GOFILES to: $(shell find . ../../pkg/dendrik -name '*.go')",
	},
	{
		ID: "no-json-encoder", Layer: LayerBridge, Scope: ScopeDendrik, Severity: SeverityError,
		Summary:     "No json.NewEncoder in cmd_*.go files",
		Rationale:   "json.NewEncoder writes directly to stdout, bypassing the ResultEnvelope contract. Use dendrik.Output.Result instead.",
		Remediation: "Replace json.NewEncoder(os.Stdout).Encode(...) with dendrik.Output.Result() or dendrik.WriteResult().",
	},
	{
		ID: "no-raw-json", Layer: LayerBridge, Scope: ScopeDendrik, Severity: SeverityWarning,
		Summary:     "No fmt.Print(string( in cmd_*.go files that register a --json flag",
		Rationale:   "Raw JSON passthrough bypasses the envelope. Sometimes intentional (API passthrough) but usually a contract violation.",
		Remediation: "Wrap raw JSON in a ResultEnvelope via dendrik.Output.Result(), or add //nolint:no-raw-json if passthrough is intentional.",
	},
	{
		ID: "run-has-json", Layer: LayerBridge, Scope: ScopeDendrik, Severity: SeverityWarning,
		Summary:     "All cmd_*.go files with run* functions register a --json flag",
		Rationale:   "Commands without --json are invisible to agents in non-TTY contexts. Not all commands need it, but gaps should be visible.",
		Remediation: "Add a --json flag to the command's flag set, or acknowledge the gap is intentional.",
	},
}

// LookupCheck returns the ContractEntry for the given check ID, or nil if not found.
func LookupCheck(id string) *ContractEntry {
	for i := range Contract {
		if Contract[i].ID == id {
			return &Contract[i]
		}
	}
	return nil
}

// ChecksByLayer returns all contract entries for the given layer.
func ChecksByLayer(layer Layer) []ContractEntry {
	var result []ContractEntry
	for _, c := range Contract {
		if c.Layer == layer {
			result = append(result, c)
		}
	}
	return result
}
