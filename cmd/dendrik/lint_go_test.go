package main

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik/conventions"
)

func TestGoLint_StructureGate(t *testing.T) {
	data := &ToolData{ToolName: "test", GoMod: nil}
	results := GoLint(data)
	assertCheckPresent(t, results, "go-mod-linked")
	// Structure gate: no other checks should fire
	for _, r := range results {
		if r.CheckID != "go-mod-linked" {
			t.Errorf("expected only go-mod-linked after structure gate, got %s", r.CheckID)
		}
	}
}

func TestGoLint_GoWorkLink(t *testing.T) {
	t.Run("linked", func(t *testing.T) {
		data := minimalToolData("mytool")
		data.GoWork = []byte("use (\n\t./cmd/mytool\n\t./pkg/dendrik\n)")
		results := filterCheck(GoLint(data), "go-mod-linked")
		if len(results) > 0 {
			t.Errorf("expected no go-mod-linked errors, got %v", results)
		}
	})

	t.Run("not linked", func(t *testing.T) {
		data := minimalToolData("mytool")
		data.GoWork = []byte("use (\n\t./cmd/other\n)")
		results := filterCheck(GoLint(data), "go-mod-linked")
		assertCheckPresent(t, results, "go-mod-linked")
	})
}

func TestGoLint_MainDispatch(t *testing.T) {
	t.Run("valid dispatch", func(t *testing.T) {
		data := minimalToolData("test")
		data.GoFiles = []GoFileData{
			parseGoFile("main.go", `package main
import "os"
func main() { os.Exit(runFoo(os.Args[1:])) }
func runFoo(args []string) int { return 0 }
`),
			cmdFile("cmd_foo.go"),
		}
		results := filterCheck(GoLint(data), "main-dispatch")
		if len(results) > 0 {
			t.Errorf("expected no main-dispatch errors, got %v", results)
		}
	})

	t.Run("no os.Exit(run*())", func(t *testing.T) {
		data := minimalToolData("test")
		data.GoFiles = []GoFileData{
			parseGoFile("main.go", `package main
import "fmt"
func main() { fmt.Println("hello") }
`),
			cmdFile("cmd_foo.go"),
		}
		results := filterCheck(GoLint(data), "main-dispatch")
		assertCheckPresent(t, results, "main-dispatch")
	})

	t.Run("missing main.go", func(t *testing.T) {
		data := minimalToolData("test")
		data.GoFiles = []GoFileData{cmdFile("cmd_foo.go")}
		results := filterCheck(GoLint(data), "main-dispatch")
		assertCheckPresent(t, results, "main-dispatch")
	})
}

func TestGoLint_CmdFiles(t *testing.T) {
	t.Run("has cmd file", func(t *testing.T) {
		data := minimalToolData("test")
		data.GoFiles = []GoFileData{
			validMainFile(),
			cmdFile("cmd_foo.go"),
		}
		results := filterCheck(GoLint(data), "cmd-file-exists")
		if len(results) > 0 {
			t.Errorf("expected no cmd-file-exists errors, got %v", results)
		}
	})

	t.Run("no cmd file", func(t *testing.T) {
		data := minimalToolData("test")
		data.GoFiles = []GoFileData{validMainFile()}
		results := filterCheck(GoLint(data), "cmd-file-exists")
		assertCheckPresent(t, results, "cmd-file-exists")
	})
}

func TestGoLint_MakefileTargets(t *testing.T) {
	t.Run("all targets", func(t *testing.T) {
		data := minimalToolData("test")
		data.Makefile = []byte("build:\ntest:\ncheck:")
		results := filterCheck(GoLint(data), "makefile-targets")
		if len(results) > 0 {
			t.Errorf("expected no makefile-targets errors, got %v", results)
		}
	})

	t.Run("missing target", func(t *testing.T) {
		data := minimalToolData("test")
		data.Makefile = []byte("build:\ntest:")
		results := filterCheck(GoLint(data), "makefile-targets")
		assertCheckPresent(t, results, "makefile-targets")
		if !strings.Contains(results[0].Message, "check") {
			t.Errorf("expected missing check target, got %s", results[0].Message)
		}
	})

	t.Run("no makefile", func(t *testing.T) {
		data := minimalToolData("test")
		data.Makefile = nil
		results := filterCheck(GoLint(data), "makefile-targets")
		assertCheckPresent(t, results, "makefile-targets")
	})
}

func TestGoLint_README(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		data := minimalToolData("test")
		data.HasREADME = true
		data.READMEBytes = []byte("# Test\n## Install\n## Quick Start\n## Commands\n## Code Structure\n")
		results := filterCheck(GoLint(data), "readme-exists")
		if len(results) > 0 {
			t.Errorf("expected no readme-exists errors, got %v", results)
		}
	})

	t.Run("missing", func(t *testing.T) {
		data := minimalToolData("test")
		data.HasREADME = false
		results := filterCheck(GoLint(data), "readme-exists")
		assertCheckPresent(t, results, "readme-exists")
	})
}

func TestGoLint_READMESections(t *testing.T) {
	t.Run("all sections", func(t *testing.T) {
		data := minimalToolData("test")
		data.HasREADME = true
		data.READMEBytes = []byte("# Test\n## Install\n## Quick Start\n## Commands\n## Code Structure\n")
		results := filterCheck(GoLint(data), "readme-sections")
		if len(results) > 0 {
			t.Errorf("expected no readme-sections warnings, got %v", results)
		}
	})

	t.Run("missing sections", func(t *testing.T) {
		data := minimalToolData("test")
		data.HasREADME = true
		data.READMEBytes = []byte("# Test\n## Install\n")
		results := filterCheck(GoLint(data), "readme-sections")
		if len(results) != 3 {
			t.Errorf("expected 3 missing sections, got %d", len(results))
		}
		for _, r := range results {
			if r.Severity != conventions.SeverityWarning {
				t.Errorf("readme-sections should be warning, got %s", r.Severity)
			}
		}
	})

	t.Run("skipped when no README", func(t *testing.T) {
		data := minimalToolData("test")
		data.HasREADME = false
		results := filterCheck(GoLint(data), "readme-sections")
		if len(results) > 0 {
			t.Errorf("readme-sections should not fire when README is missing, got %v", results)
		}
	})
}

// --- helpers ---

func minimalToolData(name string) *ToolData {
	return &ToolData{
		ToolName:    name,
		GoMod:       []byte("module example.com/" + name),
		GoWork:      []byte("use (\n\t./cmd/" + name + "\n)"),
		Makefile:    []byte("build:\ntest:\ncheck:"),
		GoFiles:     []GoFileData{validMainFile(), cmdFile("cmd_foo.go")},
		HasREADME:   true,
		READMEBytes: []byte("# Test\n## Install\n## Quick Start\n## Commands\n## Code Structure\n"),
	}
}

func parseGoFile(name, src string) GoFileData {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, name, src, parser.AllErrors)
	return GoFileData{Path: name, Content: []byte(src), AST: f, Err: err}
}

func validMainFile() GoFileData {
	return parseGoFile("main.go", `package main
import "os"
func main() { os.Exit(runFoo(os.Args[1:])) }
`)
}

func cmdFile(name string) GoFileData {
	src := `package main
func runFoo(args []string) int { return 0 }
`
	return parseGoFile(name, src)
}

func filterCheck(results []LintResult, checkID string) []LintResult {
	var out []LintResult
	for _, r := range results {
		if r.CheckID == checkID {
			out = append(out, r)
		}
	}
	return out
}

func assertCheckPresent(t *testing.T, results []LintResult, checkID string) {
	t.Helper()
	for _, r := range results {
		if r.CheckID == checkID {
			return
		}
	}
	t.Errorf("expected %s result, not found in %v", checkID, results)
}
