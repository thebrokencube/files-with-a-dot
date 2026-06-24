// Package lint holds the functional core of the dendrik `lint` verb: the parsed
// tool-data bundle, the per-layer check functions (Go, skill, bridge), and the
// Run orchestrator. The checks are pure over a pre-gathered ToolData; the
// imperative shell (cmd/dendrik) gathers the data via GatherToolData and renders
// the results. GatherToolData (gather.go) is the package's sole I/O boundary.
package lint

import (
	"go/ast"
	"sort"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik/conventions"
)

// Result is a single lint finding.
type Result struct {
	CheckID     string               `json:"check_id"`
	Severity    conventions.Severity `json:"severity"`
	Message     string               `json:"message"`
	File        string               `json:"file,omitempty"`
	Line        int                  `json:"line,omitempty"`
	Remediation string               `json:"remediation"`
}

// GoFileData holds parsed data for a single Go file.
type GoFileData struct {
	Path    string    // relative path within tool dir
	Content []byte    // raw file content
	AST     *ast.File // parsed AST (nil on parse error)
	Err     error     // parse error, if any
}

// ToolData is the I/O-free data bundle passed to linters.
type ToolData struct {
	ToolDir  string // absolute path to cmd/*/ directory
	ToolName string // directory name (e.g., "jf", "folio", "dendrik")
	RepoRoot string // absolute path to repo root

	// Go layer
	GoMod       []byte       // go.mod content (nil if missing)
	GoWork      []byte       // go.work content (nil if missing)
	Makefile    []byte       // Makefile content (nil if missing)
	GoFiles     []GoFileData // all .go files in tool dir
	HasREADME   bool         // README.md exists
	READMEBytes []byte       // README.md content

	// Docs layer
	HasCLAUDEMD bool     // CLAUDE.md exists
	DocsFiles   []string // filenames in docs/ (e.g., "01-getting-started.md")

	// Skill layer
	SkillMD     []byte            // skill/SKILL.md content (nil if missing)
	SkillDir    string            // absolute path to skill/ directory
	RefFiles    []string          // filenames in skill/references/
	RefContents map[string][]byte // reference file name -> content

	// Bridge layer
	SymlinkMap []byte   // symlink_map.txt content (nil if missing)
	CmdDirs    []string // cmd/*/ directories that contain go.mod

	PkgVerbCores []string // pkg/dendrik subdirs holding a verb core (e.g. "build", "lint")
}

// Options tunes a lint run.
type Options struct {
	Strict bool // promote warnings to errors
}

// Run executes every layer's checks over data, applies Options, and returns the
// findings sorted by severity (errors first) then check ID. Pure — no I/O.
func Run(data *ToolData, opts Options) []Result {
	var results []Result
	results = append(results, GoLint(data)...)
	results = append(results, SkillLint(data)...)
	results = append(results, BridgeLint(data)...)

	if opts.Strict {
		for i := range results {
			if results[i].Severity == conventions.SeverityWarning {
				results[i].Severity = conventions.SeverityError
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Severity != results[j].Severity {
			return results[i].Severity == conventions.SeverityError
		}
		return results[i].CheckID < results[j].CheckID
	})
	return results
}
