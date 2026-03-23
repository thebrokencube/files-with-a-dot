package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik/conventions"
)

// LintResult is a single lint finding.
type LintResult struct {
	CheckID     string              `json:"check_id"`
	Severity    conventions.Severity `json:"severity"`
	Message     string              `json:"message"`
	File        string              `json:"file,omitempty"`
	Line        int                 `json:"line,omitempty"`
	Remediation string              `json:"remediation"`
}

// GoFileData holds parsed data for a single Go file.
type GoFileData struct {
	Path    string   // relative path within tool dir
	Content []byte   // raw file content
	AST     *ast.File // parsed AST (nil on parse error)
	Err     error    // parse error, if any
}

// ToolData is the I/O-free data bundle passed to linters.
type ToolData struct {
	ToolDir     string // absolute path to cmd/*/ directory
	ToolName    string // directory name (e.g., "jf", "folio", "dendrik")
	RepoRoot    string // absolute path to repo root

	// Go layer
	GoMod       []byte       // go.mod content (nil if missing)
	GoWork      []byte       // go.work content (nil if missing)
	Makefile    []byte       // Makefile content (nil if missing)
	GoFiles     []GoFileData // all .go files in tool dir
	HasREADME   bool         // README.md exists
	READMEBytes []byte       // README.md content

	// Skill layer
	SkillMD      []byte            // skill/SKILL.md content (nil if missing)
	SkillDir     string            // absolute path to skill/ directory
	RefFiles     []string          // filenames in skill/references/
	RefContents  map[string][]byte // reference file name -> content

	// Bridge layer
	SymlinkMap  []byte   // symlink_map.txt content (nil if missing)
	CmdDirs     []string // cmd/*/ directories that contain go.mod
}

func runLint(args []string) int {
	fs := dendrik.NewFlagSet("dendrik lint")
	jsonFlag := fs.BoolLong("json", "JSON output")
	strictFlag := fs.BoolLong("strict", "Promote warnings to errors")
	explainFlag := fs.StringLong("explain", "", "Show rationale for a check ID")
	noColor := fs.BoolLong("no-color", "Disable color output")

	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	// --explain mode
	if *explainFlag != "" {
		return handleExplain(*explainFlag, *jsonFlag, *noColor)
	}

	// Positional arg: tool directory path
	remaining := fs.GetArgs()
	if len(remaining) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: dendrik lint <path> [--json] [--strict]")
		return dendrik.ExitUserError
	}

	toolDir, err := filepath.Abs(remaining[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return dendrik.ExitUserError
	}

	// Build ToolData
	data, err := gatherToolData(toolDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error gathering tool data: %v\n", err)
		return dendrik.ExitExternalErr
	}

	// Run linters
	var results []LintResult
	results = append(results, GoLint(data)...)
	results = append(results, SkillLint(data)...)
	results = append(results, BridgeLint(data)...)

	// Apply --strict
	if *strictFlag {
		for i := range results {
			if results[i].Severity == conventions.SeverityWarning {
				results[i].Severity = conventions.SeverityError
			}
		}
	}

	// Sort by severity (errors first), then check ID
	sort.Slice(results, func(i, j int) bool {
		if results[i].Severity != results[j].Severity {
			return results[i].Severity == conventions.SeverityError
		}
		return results[i].CheckID < results[j].CheckID
	})

	// Output
	out := dendrik.NewOutput(*jsonFlag, *noColor)

	if out.IsJSON() {
		type jsonOutput struct {
			Tool     string       `json:"tool"`
			Errors   int          `json:"errors"`
			Warnings int          `json:"warnings"`
			Results  []LintResult `json:"results"`
		}
		errors, warnings := countSeverities(results)
		fmt.Print(string(out.MustResult(jsonOutput{
			Tool:     data.ToolName,
			Errors:   errors,
			Warnings: warnings,
			Results:  results,
		})))
		if errors > 0 {
			return dendrik.ExitUserError
		}
		return dendrik.ExitOK
	}

	// Human output
	errors, warnings := countSeverities(results)
	if len(results) == 0 {
		fmt.Println(out.Success("All 25 checks passed for %s", data.ToolName))
		return dendrik.ExitOK
	}

	for _, r := range results {
		icon := out.Pal.Yellow + "W" + out.Pal.Reset
		if r.Severity == conventions.SeverityError {
			icon = out.Pal.Red + "E" + out.Pal.Reset
		}
		loc := r.File
		if r.Line > 0 {
			loc = fmt.Sprintf("%s:%d", r.File, r.Line)
		}
		fmt.Printf("  %s [%s] %s", icon, r.CheckID, r.Message)
		if loc != "" {
			fmt.Printf(" (%s)", loc)
		}
		fmt.Println()
		if r.Remediation != "" {
			fmt.Printf("    %s%s%s\n", out.Pal.Dim, r.Remediation, out.Pal.Reset)
		}
	}

	fmt.Printf("\n%s: %d error(s), %d warning(s)\n", data.ToolName, errors, warnings)
	if errors > 0 {
		return dendrik.ExitUserError
	}
	return dendrik.ExitOK
}

func handleExplain(checkID string, jsonFlag, noColor bool) int {
	entry := conventions.LookupCheck(checkID)
	if entry == nil {
		fmt.Fprintf(os.Stderr, "Unknown check: %s\n", checkID)
		return dendrik.ExitUserError
	}

	out := dendrik.NewOutput(jsonFlag, noColor)
	if out.IsJSON() {
		fmt.Print(string(out.MustResult(entry)))
		return dendrik.ExitOK
	}

	fmt.Printf("%s%s%s [%s] %s\n", out.Pal.Bold, entry.ID, out.Pal.Reset, entry.Severity, entry.Summary)
	fmt.Printf("\n%sRationale:%s %s\n", out.Pal.Bold, out.Pal.Reset, entry.Rationale)
	fmt.Printf("\n%sRemediation:%s %s\n", out.Pal.Bold, out.Pal.Reset, entry.Remediation)
	return dendrik.ExitOK
}

func gatherToolData(toolDir string) (*ToolData, error) {
	toolName := filepath.Base(toolDir)

	// Find repo root by walking up to find go.work
	repoRoot := toolDir
	for {
		if _, err := os.Stat(filepath.Join(repoRoot, "go.work")); err == nil {
			break
		}
		parent := filepath.Dir(repoRoot)
		if parent == repoRoot {
			return nil, fmt.Errorf("could not find go.work in parent directories of %s", toolDir)
		}
		repoRoot = parent
	}

	data := &ToolData{
		ToolDir:  toolDir,
		ToolName: toolName,
		RepoRoot: repoRoot,
	}

	// Read files (nil if missing — that's fine, linters check for it)
	data.GoMod, _ = os.ReadFile(filepath.Join(toolDir, "go.mod"))
	data.GoWork, _ = os.ReadFile(filepath.Join(repoRoot, "go.work"))
	data.Makefile, _ = os.ReadFile(filepath.Join(toolDir, "Makefile"))
	data.SymlinkMap, _ = os.ReadFile(filepath.Join(repoRoot, "symlink_map.txt"))

	// README
	readmePath := filepath.Join(toolDir, "README.md")
	if content, err := os.ReadFile(readmePath); err == nil {
		data.HasREADME = true
		data.READMEBytes = content
	}

	// Skill layer
	data.SkillDir = filepath.Join(toolDir, "skill")
	data.SkillMD, _ = os.ReadFile(filepath.Join(data.SkillDir, "SKILL.md"))

	data.RefContents = map[string][]byte{}
	refsDir := filepath.Join(data.SkillDir, "references")
	if entries, err := os.ReadDir(refsDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				data.RefFiles = append(data.RefFiles, e.Name())
				if content, err := os.ReadFile(filepath.Join(refsDir, e.Name())); err == nil {
					data.RefContents[e.Name()] = content
				}
			}
		}
	}

	// Go files — parse all .go files in tool dir
	fset := token.NewFileSet()
	entries, err := os.ReadDir(toolDir)
	if err != nil {
		return nil, fmt.Errorf("reading tool directory: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		path := filepath.Join(toolDir, e.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		gf := GoFileData{
			Path:    e.Name(),
			Content: content,
		}
		astFile, parseErr := parser.ParseFile(fset, path, content, parser.AllErrors)
		if parseErr != nil {
			gf.Err = parseErr
		}
		gf.AST = astFile
		data.GoFiles = append(data.GoFiles, gf)
	}

	// cmd/*/ directories with go.mod (for go-work-sync)
	cmdDir := filepath.Join(repoRoot, "cmd")
	if cmdEntries, err := os.ReadDir(cmdDir); err == nil {
		for _, e := range cmdEntries {
			if !e.IsDir() {
				continue
			}
			if _, err := os.Stat(filepath.Join(cmdDir, e.Name(), "go.mod")); err == nil {
				data.CmdDirs = append(data.CmdDirs, e.Name())
			}
		}
	}

	return data, nil
}

func countSeverities(results []LintResult) (errors, warnings int) {
	for _, r := range results {
		if r.Severity == conventions.SeverityError {
			errors++
		} else {
			warnings++
		}
	}
	return
}

