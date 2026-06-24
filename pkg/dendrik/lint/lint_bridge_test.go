package lint

import (
	"testing"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik/conventions"
)

func TestBridgeLint_DendrikImport(t *testing.T) {
	t.Run("has import", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoFiles = []GoFileData{
			{Path: "main.go", Content: []byte("package main\nimport \"github.com/thebrokencube/files-with-a-dot/pkg/dendrik\"\n")},
		}
		results := filterCheck(BridgeLint(data), "dendrik-import")
		if len(results) > 0 {
			t.Errorf("expected no dendrik-import errors, got %v", results)
		}
	})

	t.Run("no import", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoFiles = []GoFileData{
			{Path: "main.go", Content: []byte("package main\n")},
		}
		results := filterCheck(BridgeLint(data), "dendrik-import")
		assertCheckPresent(t, results, "dendrik-import")
	})
}

func TestBridgeLint_ExitConstants(t *testing.T) {
	t.Run("bare return in cmd file", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoFiles = []GoFileData{
			dendrikGoFile("main.go"),
			parseGoFile("cmd_foo.go", "package main\nfunc runFoo() int { return 0 }\n"),
		}
		results := filterCheck(BridgeLint(data), "exit-constants")
		assertCheckPresent(t, results, "exit-constants")
	})

	t.Run("no bare return", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoFiles = []GoFileData{
			dendrikGoFile("main.go"),
			parseGoFile("cmd_foo.go", "package main\nimport \"github.com/thebrokencube/files-with-a-dot/pkg/dendrik\"\nfunc runFoo() int { return dendrik.ExitOK }\n"),
		}
		results := filterCheck(BridgeLint(data), "exit-constants")
		if len(results) > 0 {
			t.Errorf("expected no exit-constants errors, got %v", results)
		}
	})

	t.Run("os.Exit outside main.go", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoFiles = []GoFileData{
			dendrikGoFile("main.go"),
			parseGoFile("cmd_foo.go", "package main\nimport \"os\"\nfunc runFoo() { os.Exit(1) }\n"),
		}
		results := filterCheck(BridgeLint(data), "exit-constants")
		assertCheckPresent(t, results, "exit-constants")
	})
}

func TestBridgeLint_JSONOutput(t *testing.T) {
	t.Run("json flag with WriteResult", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoFiles = []GoFileData{
			{Path: "cmd_foo.go", Content: []byte("package main\nvar f = fs.BoolLong(\"json\", \"JSON\")\nfunc x() { dendrik.WriteResult(nil) }\n")},
		}
		results := filterCheck(BridgeLint(data), "json-output")
		if len(results) > 0 {
			t.Errorf("expected no json-output errors, got %v", results)
		}
	})

	t.Run("json flag without WriteResult", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoFiles = []GoFileData{
			{Path: "cmd_foo.go", Content: []byte("package main\nvar f = fs.BoolLong(\"json\", \"JSON\")\nfunc x() { fmt.Println(\"hi\") }\n")},
		}
		results := filterCheck(BridgeLint(data), "json-output")
		assertCheckPresent(t, results, "json-output")
	})
}

func TestBridgeLint_GoWorkSync(t *testing.T) {
	t.Run("in sync", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoWork = []byte("use (\n\t./cmd/test\n\t./cmd/other\n)")
		data.CmdDirs = []string{"test", "other"}
		results := filterCheck(BridgeLint(data), "go-work-sync")
		if len(results) > 0 {
			t.Errorf("expected no go-work-sync errors, got %v", results)
		}
	})

	t.Run("in go.work but not on disk", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoWork = []byte("use (\n\t./cmd/test\n\t./cmd/ghost\n)")
		data.CmdDirs = []string{"test"}
		results := filterCheck(BridgeLint(data), "go-work-sync")
		assertCheckPresent(t, results, "go-work-sync")
	})

	t.Run("on disk but not in go.work", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoWork = []byte("use (\n\t./cmd/test\n)")
		data.CmdDirs = []string{"test", "new-tool"}
		results := filterCheck(BridgeLint(data), "go-work-sync")
		assertCheckPresent(t, results, "go-work-sync")
	})

	t.Run("no go.work", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoWork = nil
		results := filterCheck(BridgeLint(data), "go-work-sync")
		if len(results) > 0 {
			t.Errorf("expected no errors when go.work missing, got %v", results)
		}
	})
}

func TestBridgeLint_SymlinkEntries(t *testing.T) {
	t.Run("has skill entry", func(t *testing.T) {
		data := bridgeToolData("test")
		data.SymlinkMap = []byte("cmd/test/skill ~/.config/skills/test\n")
		results := filterCheck(BridgeLint(data), "symlink-entries")
		if len(results) > 0 {
			t.Errorf("expected no symlink-entries errors, got %v", results)
		}
	})

	t.Run("missing skill", func(t *testing.T) {
		data := bridgeToolData("test")
		data.SymlinkMap = []byte("cmd/other/skill ~/.config/skills/other\n")
		results := filterCheck(BridgeLint(data), "symlink-entries")
		assertCheckPresent(t, results, "symlink-entries")
	})

	t.Run("no symlink map", func(t *testing.T) {
		data := bridgeToolData("test")
		data.SymlinkMap = nil
		results := filterCheck(BridgeLint(data), "symlink-entries")
		if len(results) > 0 {
			t.Errorf("expected no errors when symlink_map.txt missing, got %v", results)
		}
	})
}

func TestBridgeLint_MakefileGofiles(t *testing.T) {
	t.Run("includes dendrik path", func(t *testing.T) {
		data := bridgeToolData("test")
		data.Makefile = []byte("GOFILES = $(shell find . ../../pkg/dendrik -name '*.go')\nbuild:\n")
		results := filterCheck(BridgeLint(data), "makefile-gofiles")
		if len(results) > 0 {
			t.Errorf("expected no makefile-gofiles warnings, got %v", results)
		}
	})

	t.Run("missing dendrik path", func(t *testing.T) {
		data := bridgeToolData("test")
		data.Makefile = []byte("GOFILES = $(shell find . -name '*.go')\nbuild:\n")
		results := filterCheck(BridgeLint(data), "makefile-gofiles")
		assertCheckPresent(t, results, "makefile-gofiles")
		if results[0].Severity != conventions.SeverityWarning {
			t.Errorf("makefile-gofiles should be warning, got %s", results[0].Severity)
		}
	})
}

func TestBridgeLint_NoJsonEncoder(t *testing.T) {
	t.Run("has encoder in cmd file", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoFiles = []GoFileData{
			{Path: "cmd_foo.go", Content: []byte("package main\njson.NewEncoder(os.Stdout)\n")},
		}
		results := filterCheck(BridgeLint(data), "no-json-encoder")
		assertCheckPresent(t, results, "no-json-encoder")
	})

	t.Run("no encoder", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoFiles = []GoFileData{
			{Path: "cmd_foo.go", Content: []byte("package main\nfunc x() {}\n")},
		}
		results := filterCheck(BridgeLint(data), "no-json-encoder")
		if len(results) > 0 {
			t.Errorf("expected no no-json-encoder errors, got %v", results)
		}
	})

	t.Run("encoder in non-cmd file", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoFiles = []GoFileData{
			{Path: "helpers.go", Content: []byte("package main\njson.NewEncoder(os.Stdout)\n")},
		}
		results := filterCheck(BridgeLint(data), "no-json-encoder")
		if len(results) > 0 {
			t.Errorf("expected no errors for non-cmd file, got %v", results)
		}
	})
}

func TestBridgeLint_NoRawJSON(t *testing.T) {
	t.Run("raw json passthrough with json flag", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoFiles = []GoFileData{
			{Path: "cmd_foo.go", Content: []byte("package main\nvar f = fs.BoolLong(\"json\", \"JSON\")\nfmt.Print(string(data))\n")},
		}
		results := filterCheck(BridgeLint(data), "no-raw-json")
		assertCheckPresent(t, results, "no-raw-json")
		if results[0].Severity != conventions.SeverityWarning {
			t.Errorf("no-raw-json should be warning, got %s", results[0].Severity)
		}
	})

	t.Run("no raw json", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoFiles = []GoFileData{
			{Path: "cmd_foo.go", Content: []byte("package main\nvar f = fs.BoolLong(\"json\", \"JSON\")\nfunc x() {}\n")},
		}
		results := filterCheck(BridgeLint(data), "no-raw-json")
		if len(results) > 0 {
			t.Errorf("expected no no-raw-json warnings, got %v", results)
		}
	})

	t.Run("raw print without json flag", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoFiles = []GoFileData{
			{Path: "cmd_foo.go", Content: []byte("package main\nfmt.Print(string(data))\n")},
		}
		results := filterCheck(BridgeLint(data), "no-raw-json")
		if len(results) > 0 {
			t.Errorf("should not flag without json flag, got %v", results)
		}
	})

	t.Run("MustResult is not flagged", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoFiles = []GoFileData{
			{Path: "cmd_foo.go", Content: []byte("package main\nvar f = fs.BoolLong(\"json\", \"JSON\")\nfmt.Print(string(out.MustResult(x)))\n")},
		}
		results := filterCheck(BridgeLint(data), "no-raw-json")
		if len(results) > 0 {
			t.Errorf("MustResult goes through envelope, should not flag, got %v", results)
		}
	})

	t.Run("nolint comment is not flagged", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoFiles = []GoFileData{
			{Path: "cmd_foo.go", Content: []byte("package main\nvar f = fs.BoolLong(\"json\", \"JSON\")\nfmt.Print(string(data)) //nolint:no-raw-json\n")},
		}
		results := filterCheck(BridgeLint(data), "no-raw-json")
		if len(results) > 0 {
			t.Errorf("nolint should suppress, got %v", results)
		}
	})
}

func TestBridgeLint_RunHasJSON(t *testing.T) {
	t.Run("run function with json flag", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoFiles = []GoFileData{
			{Path: "cmd_foo.go", Content: []byte("package main\nfunc runFoo() int { return 0 }\nvar f = fs.BoolLong(\"json\", \"JSON\")\n")},
		}
		results := filterCheck(BridgeLint(data), "run-has-json")
		if len(results) > 0 {
			t.Errorf("expected no run-has-json warnings, got %v", results)
		}
	})

	t.Run("run function without json flag", func(t *testing.T) {
		data := bridgeToolData("test")
		data.GoFiles = []GoFileData{
			{Path: "cmd_foo.go", Content: []byte("package main\nfunc runFoo() int { return 0 }\n")},
		}
		results := filterCheck(BridgeLint(data), "run-has-json")
		assertCheckPresent(t, results, "run-has-json")
		if results[0].Severity != conventions.SeverityWarning {
			t.Errorf("run-has-json should be warning, got %s", results[0].Severity)
		}
	})
}

// --- helpers ---

func bridgeToolData(name string) *ToolData {
	return &ToolData{
		ToolName: name,
		GoFiles: []GoFileData{
			dendrikGoFile("main.go"),
		},
	}
}

func dendrikGoFile(path string) GoFileData {
	return GoFileData{
		Path:    path,
		Content: []byte("package main\nimport \"github.com/thebrokencube/files-with-a-dot/pkg/dendrik\"\n"),
	}
}
