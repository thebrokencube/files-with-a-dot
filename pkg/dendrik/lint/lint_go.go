package lint

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik/conventions"
)

var docsNamingPattern = regexp.MustCompile(`^\d{2}-[a-z0-9-]+\.md$`)

// GoLint validates Go layer conventions. Pure function — no I/O.
func GoLint(data *ToolData) []Result {
	var results []Result

	if data.GoMod == nil {
		results = append(results, lintResult("go-mod-linked", conventions.SeverityError,
			"go.mod not found",
			"go.mod", 0,
			"Create go.mod with `go mod init` and add a `use` entry in go.work."))
		return results // Structure gate
	}

	goWorkLinked := false
	if data.GoWork != nil {
		needle := "./cmd/" + data.ToolName
		goWorkLinked = strings.Contains(string(data.GoWork), needle)
	}
	if !goWorkLinked {
		results = append(results, lintResult("go-mod-linked", conventions.SeverityError,
			"go.work does not link this tool (expected ./cmd/"+data.ToolName+")",
			"go.work", 0,
			"Add `./cmd/"+data.ToolName+"` to the `use` block in go.work."))
	}

	results = append(results, checkMainDispatch(data)...)
	results = append(results, checkVersionFlag(data)...)
	results = append(results, checkCoreInPkg(data)...)

	hasCmdFile := false
	for _, gf := range data.GoFiles {
		if strings.HasPrefix(gf.Path, "cmd_") && strings.HasSuffix(gf.Path, ".go") {
			hasCmdFile = true
			break
		}
	}
	if !hasCmdFile {
		results = append(results, lintResult("cmd-file-exists", conventions.SeverityError,
			"no cmd_*.go files found",
			"", 0,
			"Create at least one file matching cmd_*.go (e.g., cmd_lint.go) with a run*() function."))
	}

	results = append(results, checkMakefileTargets(data)...)

	if !data.HasREADME {
		results = append(results, lintResult("readme-exists", conventions.SeverityError,
			"README.md not found",
			"README.md", 0,
			"Create README.md in the tool directory."))
	} else {
		results = append(results, checkREADMESections(data)...)
		results = append(results, checkREADMEDocLinks(data)...)
	}

	results = append(results, checkCLAUDEMDExists(data)...)
	results = append(results, checkDocsNaming(data)...)
	results = append(results, checkDocsGettingStarted(data)...)

	return results
}

func checkMainDispatch(data *ToolData) []Result {
	var results []Result

	// Find main.go
	var mainFile *GoFileData
	for i := range data.GoFiles {
		if data.GoFiles[i].Path == "main.go" {
			mainFile = &data.GoFiles[i]
			break
		}
	}

	if mainFile == nil {
		results = append(results, lintResult("main-dispatch", conventions.SeverityError,
			"main.go not found",
			"main.go", 0,
			"Create main.go with func main() that delegates to run*() functions via os.Exit()."))
		return results
	}

	if mainFile.AST == nil {
		results = append(results, lintResult("main-dispatch", conventions.SeverityError,
			"main.go has parse errors",
			"main.go", 0,
			"Fix syntax errors in main.go."))
		return results
	}

	// Walk func main() body looking for os.Exit(run*(...)) calls
	found := false
	ast.Inspect(mainFile.AST, func(n ast.Node) bool {
		if found {
			return false
		}
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "main" {
			return true
		}
		// Walk the body of main() for os.Exit calls with run* args
		ast.Inspect(fn.Body, func(inner ast.Node) bool {
			call, ok := inner.(*ast.CallExpr)
			if !ok {
				return true
			}
			if isOsExitCall(call) && hasRunArg(call) {
				found = true
				return false
			}
			return true
		})
		return false
	})

	if !found {
		results = append(results, lintResult("main-dispatch", conventions.SeverityError,
			"func main() has no os.Exit(run*(...)) call",
			"main.go", 0,
			"In main.go, delegate to run*() functions via os.Exit(run*(...))."))
	}

	return results
}

// isOsExitCall checks if a call expression is os.Exit(...)
func isOsExitCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return ident.Name == "os" && sel.Sel.Name == "Exit"
}

// hasRunArg checks if an os.Exit call has a run*(...) call as its argument
func hasRunArg(call *ast.CallExpr) bool {
	if len(call.Args) != 1 {
		return false
	}
	innerCall, ok := call.Args[0].(*ast.CallExpr)
	if !ok {
		return false
	}
	ident, ok := innerCall.Fun.(*ast.Ident)
	if !ok {
		return false
	}
	return strings.HasPrefix(ident.Name, "run")
}

// checkVersionFlag verifies main.go handles a --version flag (distinct from any
// `version` subcommand). Stays silent when main.go is missing or unparseable —
// main-dispatch already reports that, and double-reporting would be noise.
func checkVersionFlag(data *ToolData) []Result {
	var mainFile *GoFileData
	for i := range data.GoFiles {
		if data.GoFiles[i].Path == "main.go" {
			mainFile = &data.GoFiles[i]
			break
		}
	}
	if mainFile == nil || mainFile.AST == nil {
		return nil
	}
	if mainHandlesVersionFlag(mainFile.AST) {
		return nil
	}
	return []Result{lintResult("version-flag", conventions.SeverityWarning,
		"main.go does not handle a --version flag",
		"main.go", 0,
		"In main()'s dispatch, fold the flag forms into the version case: `case \"version\", \"--version\", \"-V\":`.")}
}

// mainHandlesVersionFlag reports whether func main() has a switch case whose
// expression is the string literal "--version" — the flag form, distinct from
// the "version" subcommand (which a naive string match would falsely accept).
func mainHandlesVersionFlag(file *ast.File) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		if found {
			return false
		}
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "main" {
			return true
		}
		ast.Inspect(fn.Body, func(inner ast.Node) bool {
			cc, ok := inner.(*ast.CaseClause)
			if !ok {
				return true
			}
			for _, expr := range cc.List {
				lit, ok := expr.(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				if val, err := strconv.Unquote(lit.Value); err == nil && val == "--version" {
					found = true
					return false
				}
			}
			return true
		})
		return false
	})
	return found
}

// checkCoreInPkg guards against verb logic drifting back into package main: once
// a verb's core lives in pkg/dendrik/<verb>, its cmd_<verb>.go must import that
// core rather than redeclare domain types. It fires only for verbs whose core
// already exists (data.PkgVerbCores) — a new verb with no core yet is silent.
// Import-presence + in-file type-location only; the linter can't read pkg source
// to diff against the core's actual exports (deferred to a verb-core registry).
func checkCoreInPkg(data *ToolData) []Result {
	hasCore := map[string]bool{}
	for _, v := range data.PkgVerbCores {
		hasCore[v] = true
	}
	var results []Result
	for _, gf := range data.GoFiles {
		if !strings.HasPrefix(gf.Path, "cmd_") || !strings.HasSuffix(gf.Path, ".go") {
			continue
		}
		if strings.HasSuffix(gf.Path, "_test.go") || gf.AST == nil {
			continue
		}
		verb := strings.TrimSuffix(strings.TrimPrefix(gf.Path, "cmd_"), ".go")
		if !hasCore[verb] || !fileDeclaresTopLevelType(gf.AST) || fileImportsCore(gf.AST, verb) {
			continue
		}
		results = append(results, lintResult("core-in-pkg", conventions.SeverityError,
			"cmd_"+verb+".go declares a domain type but does not import its pkg/dendrik/"+verb+" core",
			gf.Path, 0,
			"The pkg/dendrik/"+verb+" core exists; import it instead of redeclaring the type — keep cmd_"+verb+".go a thin shell."))
	}
	return results
}

// fileDeclaresTopLevelType reports whether the file has any top-level type decl.
func fileDeclaresTopLevelType(f *ast.File) bool {
	for _, decl := range f.Decls {
		if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.TYPE {
			return true
		}
	}
	return false
}

// fileImportsCore reports whether the file imports pkg/dendrik/<verb>.
func fileImportsCore(f *ast.File, verb string) bool {
	suffix := "/pkg/dendrik/" + verb
	for _, imp := range f.Imports {
		p, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		if strings.HasSuffix(p, suffix) {
			return true
		}
	}
	return false
}

func checkMakefileTargets(data *ToolData) []Result {
	var results []Result

	if data.Makefile == nil {
		results = append(results, lintResult("makefile-targets", conventions.SeverityError,
			"Makefile not found",
			"Makefile", 0,
			"Create a Makefile with `build`, `test`, and `check` targets."))
		return results
	}

	content := string(data.Makefile)
	for _, target := range []string{"build", "test", "check"} {
		// Look for target: at start of line
		if !strings.Contains(content, target+":") {
			results = append(results, lintResult("makefile-targets", conventions.SeverityError,
				"Makefile missing `"+target+"` target",
				"Makefile", 0,
				"Add a `"+target+":` target to the Makefile."))
		}
	}

	return results
}

func checkREADMESections(data *ToolData) []Result {
	var results []Result

	content := string(data.READMEBytes)
	requiredSections := []string{"## Install", "## Quick Start", "## Commands", "## Code Structure"}

	for _, section := range requiredSections {
		if !strings.Contains(content, section) {
			results = append(results, lintResult("readme-sections", conventions.SeverityWarning,
				"README.md missing section: "+section,
				"README.md", 0,
				"Add a `"+section+"` section to README.md."))
		}
	}

	return results
}

func checkCLAUDEMDExists(data *ToolData) []Result {
	if data.HasCLAUDEMD {
		return nil
	}
	return []Result{lintResult("claude-md-exists", conventions.SeverityWarning,
		"CLAUDE.md not found",
		"CLAUDE.md", 0,
		"Create CLAUDE.md with standardized skeleton: Build, Test, Binary Distribution, Code Conventions, Deep Context.")}
}

func checkDocsNaming(data *ToolData) []Result {
	var results []Result
	for _, name := range data.DocsFiles {
		if !docsNamingPattern.MatchString(name) {
			results = append(results, lintResult("docs-naming", conventions.SeverityError,
				"docs/"+name+" does not match numbered kebab-case pattern (NN-name.md)",
				"docs/"+name, 0,
				"Rename to match pattern: NN-kebab-case.md (e.g., 01-getting-started.md)."))
		}
	}
	return results
}

func checkDocsGettingStarted(data *ToolData) []Result {
	if len(data.DocsFiles) == 0 {
		return nil // No docs/ directory — not an issue for this check
	}
	for _, name := range data.DocsFiles {
		if name == "01-getting-started.md" {
			return nil
		}
	}
	return []Result{lintResult("docs-getting-started", conventions.SeverityWarning,
		"docs/01-getting-started.md not found",
		"docs/", 0,
		"Create docs/01-getting-started.md as the entry point for new users.")}
}

// checkREADMEDocLinks parses the ## Documentation section of README.md and
// verifies that markdown links resolve to existing files relative to the tool dir.
func checkREADMEDocLinks(data *ToolData) []Result {
	content := string(data.READMEBytes)

	// Find ## Documentation section
	idx := strings.Index(content, "## Documentation")
	if idx == -1 {
		return nil // No Documentation section — not an issue for this check
	}

	// Extract section content (until next ## or end)
	section := content[idx:]
	if nextH2 := strings.Index(section[1:], "\n## "); nextH2 != -1 {
		section = section[:nextH2+1]
	}

	// Find markdown links: [text](path)
	var results []Result
	linkPattern := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	for _, match := range linkPattern.FindAllStringSubmatch(section, -1) {
		linkPath := match[2]
		// Skip URLs
		if strings.HasPrefix(linkPath, "http://") || strings.HasPrefix(linkPath, "https://") {
			continue
		}
		// Resolve relative to tool dir
		absPath := filepath.Join(data.ToolDir, linkPath)
		if _, err := os.Stat(absPath); err != nil {
			results = append(results, lintResult("readme-doc-links", conventions.SeverityError,
				"README.md Documentation link broken: "+linkPath,
				"README.md", 0,
				"Fix the link to "+linkPath+" in the ## Documentation section."))
		}
	}
	return results
}

func lintResult(id string, severity conventions.Severity, msg, file string, line int, remediation string) Result {
	return Result{
		CheckID:     id,
		Severity:    severity,
		Message:     msg,
		File:        file,
		Line:        line,
		Remediation: remediation,
	}
}
