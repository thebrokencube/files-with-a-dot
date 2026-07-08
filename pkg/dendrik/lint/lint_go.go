package lint

import (
	"go/ast"
	"go/parser"
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

	results = append(results, checkDispatchRouter(data)...)
	results = append(results, checkLeafStrictness(data)...)
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

// checkDispatchRouter enforces that subcommand dispatch is data (a
// dendrik.Command tree), not a hand-rolled switch on os.Args. Two assertions:
// (a) func main() must dispatch via <expr>.Execute(...); (b) no top-level
// switch on os.Args[N] may appear in a non-test file unless annotated.
func checkDispatchRouter(data *ToolData) []Result {
	var results []Result

	var mainFile *GoFileData
	for i := range data.GoFiles {
		if data.GoFiles[i].Path == "main.go" {
			mainFile = &data.GoFiles[i]
			break
		}
	}
	switch {
	case mainFile == nil:
		results = append(results, lintResult("dispatch-router", conventions.SeverityError,
			"main.go not found",
			"main.go", 0,
			"Create main.go with func main() that calls os.Exit(buildRoot().Execute(os.Args[1:]))."))
	case mainFile.AST == nil:
		results = append(results, lintResult("dispatch-router", conventions.SeverityError,
			"main.go has parse errors",
			"main.go", 0,
			"Fix syntax errors in main.go."))
	case !mainDispatchesViaExecute(mainFile.AST):
		results = append(results, lintResult("dispatch-router", conventions.SeverityError,
			"func main() must dispatch via dendrik.Command.Execute",
			"main.go", 0,
			"In main(), call os.Exit(buildRoot().Execute(os.Args[1:])) instead of a hand-rolled switch."))
	}

	for _, gf := range data.GoFiles {
		if !strings.HasSuffix(gf.Path, ".go") || strings.HasSuffix(gf.Path, "_test.go") {
			continue
		}
		results = append(results, dispatchSwitchViolations(gf)...)
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

// mainDispatchesViaExecute reports whether func main() contains an
// os.Exit(<expr>.Execute(...)) call — the dendrik.Command dispatch entrypoint.
func mainDispatchesViaExecute(file *ast.File) bool {
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		if found {
			return false
		}
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "main" || fn.Body == nil {
			return true
		}
		ast.Inspect(fn.Body, func(inner ast.Node) bool {
			call, ok := inner.(*ast.CallExpr)
			if !ok || !isOsExitCall(call) || len(call.Args) != 1 {
				return true
			}
			arg, ok := call.Args[0].(*ast.CallExpr)
			if !ok {
				return true
			}
			if sel, ok := arg.Fun.(*ast.SelectorExpr); ok && sel.Sel.Name == "Execute" {
				found = true
				return false
			}
			return true
		})
		return false
	})
	return found
}

// dispatchSwitchViolations flags any switch whose tag indexes os.Args — the
// hand-rolled top-level dispatch the Command tree replaces. RunRaw leaves that
// switch on a pre-sliced `args` param are the sanctioned escape hatch and are
// governed by leaf-strictness at the RunRaw site, not here. A switch is exempt
// when its source line (or the line above) carries //nolint:dispatch-router.
func dispatchSwitchViolations(gf GoFileData) []Result {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, gf.Path, gf.Content, parser.AllErrors)
	if file == nil || err != nil {
		return nil
	}
	lines := strings.Split(string(gf.Content), "\n")
	var results []Result
	ast.Inspect(file, func(n ast.Node) bool {
		sw, ok := n.(*ast.SwitchStmt)
		if !ok || sw.Tag == nil || !isArgsIndexTag(sw.Tag) {
			return true
		}
		line := fset.Position(sw.Pos()).Line
		if lineHasNolintDispatch(lines, line) {
			return true
		}
		results = append(results, lintResult("dispatch-router", conventions.SeverityError,
			"hand-rolled subcommand dispatch in "+gf.Path,
			gf.Path, line,
			"Build a dendrik.Command tree or annotate the switch with //nolint:dispatch-router."))
		return true
	})
	return results
}

// isArgsIndexTag reports whether the switch tag is an index into os.Args — e.g.
// `switch os.Args[1]`. Kept conservative: only the os.Args base counts so a
// RunRaw leaf's `switch args[0]` (a pre-sliced param) is not misflagged.
func isArgsIndexTag(tag ast.Expr) bool {
	idx, ok := tag.(*ast.IndexExpr)
	if !ok {
		return false
	}
	sel, ok := idx.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	base, ok := sel.X.(*ast.Ident)
	return ok && base.Name == "os" && sel.Sel.Name == "Args"
}

// lineHasNolintDispatch reports whether the 1-indexed source line or the line
// directly above it contains a //nolint:dispatch-router annotation.
func lineHasNolintDispatch(lines []string, line int) bool {
	for _, idx := range []int{line - 1, line - 2} {
		if idx >= 0 && idx < len(lines) && strings.Contains(lines[idx], "//nolint:dispatch-router") {
			return true
		}
	}
	return false
}

// checkLeafStrictness warns when a RunRaw escape hatch is used without an
// explicit //nolint:dispatch-router annotation (same line or the line above).
// Every escape from strict Command-tree dispatch must be deliberately marked.
func checkLeafStrictness(data *ToolData) []Result {
	var results []Result
	for _, gf := range data.GoFiles {
		if !strings.HasSuffix(gf.Path, ".go") || strings.HasSuffix(gf.Path, "_test.go") {
			continue
		}
		lines := strings.Split(string(gf.Content), "\n")
		for i, line := range lines {
			if !strings.Contains(line, "RunRaw:") {
				continue
			}
			if lineHasNolintDispatch(lines, i+1) {
				continue
			}
			results = append(results, lintResult("leaf-strictness", conventions.SeverityWarning,
				"RunRaw escape hatch in "+gf.Path+" lacks //nolint:dispatch-router",
				gf.Path, i+1,
				"Annotate the RunRaw field with //nolint:dispatch-router explaining why strict dispatch does not apply."))
		}
	}
	return results
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
