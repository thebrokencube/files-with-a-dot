package main

import (
	"go/ast"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik/conventions"
)

// GoLint validates Go layer conventions. Pure function — no I/O.
func GoLint(data *ToolData) []LintResult {
	var results []LintResult

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
	}

	return results
}

func checkMainDispatch(data *ToolData) []LintResult {
	var results []LintResult

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

func checkMakefileTargets(data *ToolData) []LintResult {
	var results []LintResult

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

func checkREADMESections(data *ToolData) []LintResult {
	var results []LintResult

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

func lintResult(id string, severity conventions.Severity, msg, file string, line int, remediation string) LintResult {
	return LintResult{
		CheckID:     id,
		Severity:    severity,
		Message:     msg,
		File:        file,
		Line:        line,
		Remediation: remediation,
	}
}
