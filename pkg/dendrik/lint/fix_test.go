package lint

import (
	"os"
	"path/filepath"
	"testing"
)

// writeSyntheticRoot lays down a minimal repo root (go.work + symlink_map.txt)
// in a temp dir and returns its path.
func writeSyntheticRoot(t *testing.T, goWork, symlinkMap string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.work"), []byte(goWork), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "symlink_map.txt"), []byte(symlinkMap), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestApplyFixes_GoWorkSync(t *testing.T) {
	// go.work is missing ./cmd/newtool, which is on disk (CmdDirs).
	root := writeSyntheticRoot(t, "go 1.25.0\n\nuse (\n\t./cmd/jf\n\t./pkg/dendrik\n)\n", "")
	data := &ToolData{
		RepoRoot: root,
		ToolName: "newtool",
		GoWork:   []byte("go 1.25.0\n\nuse (\n\t./cmd/jf\n\t./pkg/dendrik\n)\n"),
		CmdDirs:  []string{"jf", "newtool"},
	}
	results := []Result{{CheckID: "go-work-sync"}}

	fixed, err := ApplyFixes(data, results)
	if err != nil {
		t.Fatal(err)
	}
	if len(fixed) != 1 || fixed[0] != "go-work-sync" {
		t.Fatalf("expected go-work-sync fixed, got %v", fixed)
	}

	got, _ := os.ReadFile(filepath.Join(root, "go.work"))
	want := "go 1.25.0\n\nuse (\n\t./cmd/jf\n\t./cmd/newtool\n\t./pkg/dendrik\n)\n"
	if string(got) != want {
		t.Errorf("go.work after fix:\n%q\nwant:\n%q", got, want)
	}
}

func TestApplyFixes_SymlinkEntries(t *testing.T) {
	root := writeSyntheticRoot(t, "go 1.25.0\n\nuse (\n)\n", "configs/base/x:$HOME/.x\n")
	data := &ToolData{
		RepoRoot:   root,
		ToolName:   "newtool",
		SymlinkMap: []byte("configs/base/x:$HOME/.x\n"),
	}
	results := []Result{{CheckID: "symlink-entries"}}

	fixed, err := ApplyFixes(data, results)
	if err != nil {
		t.Fatal(err)
	}
	if len(fixed) != 1 {
		t.Fatalf("expected symlink-entries fixed, got %v", fixed)
	}
	got, _ := os.ReadFile(filepath.Join(root, "symlink_map.txt"))
	want := "configs/base/x:$HOME/.x\nplugins/newtool/skills/newtool:$HOME/.claude/skills/newtool\n"
	if string(got) != want {
		t.Errorf("symlink_map after fix:\n%q\nwant:\n%q", got, want)
	}
}

func TestApplyFixes_Idempotent(t *testing.T) {
	root := writeSyntheticRoot(t, "go 1.25.0\n\nuse (\n\t./cmd/jf\n)\n", "plugins/jf/skills/jf:$HOME/.claude/skills/jf\n")
	data := &ToolData{
		RepoRoot:   root,
		ToolName:   "jf",
		GoWork:     []byte("go 1.25.0\n\nuse (\n\t./cmd/jf\n)\n"),
		CmdDirs:    []string{"jf"},
		SymlinkMap: []byte("plugins/jf/skills/jf:$HOME/.claude/skills/jf\n"),
	}
	// Already correct: nothing should be reported fixed.
	fixed, err := ApplyFixes(data, []Result{{CheckID: "go-work-sync"}, {CheckID: "symlink-entries"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(fixed) != 0 {
		t.Errorf("expected no fixes on an already-correct root, got %v", fixed)
	}
}

func TestApplyFixes_DocsNamingNotFixed(t *testing.T) {
	// docs-naming is reported, never auto-fixed (NN- prefix is a judgment).
	root := t.TempDir()
	data := &ToolData{RepoRoot: root, ToolName: "newtool"}
	fixed, err := ApplyFixes(data, []Result{{CheckID: "docs-naming"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(fixed) != 0 {
		t.Errorf("docs-naming must not be auto-fixed, got %v", fixed)
	}
}
