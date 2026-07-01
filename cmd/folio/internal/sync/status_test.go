package sync

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

// gitInit creates a throwaway git repo with one commit and returns its path.
func gitInit(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"config", "user.email", "t@t"},
		{"config", "user.name", "t"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-q", "-m", "init"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir
}

func TestGitStatusCleanThenDirty(t *testing.T) {
	dir := gitInit(t)
	s := config.Store{Name: "code1", Kind: config.KindCode, Path: dir}

	st := For(s).Status(dir, s)
	if st.Err != "" {
		t.Fatalf("unexpected err: %s", st.Err)
	}
	if !st.Exists || st.Branch != "main" || st.Dirty {
		t.Fatalf("clean repo: exists=%v branch=%q dirty=%v", st.Exists, st.Branch, st.Dirty)
	}

	// Make it dirty.
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("changed"), 0644); err != nil {
		t.Fatal(err)
	}
	st = For(s).Status(dir, s)
	if !st.Dirty {
		t.Errorf("modified repo should be dirty, got %+v", st)
	}
}

func TestStatusMissingPath(t *testing.T) {
	s := config.Store{Name: "gone", Kind: config.KindExternal, Path: "/no/such/dir/here"}
	st := For(s).Status(s.Path, s)
	if st.Exists {
		t.Errorf("missing path should not exist: %+v", st)
	}
}

func TestStatusNeverPanicsOnEmptyPath(t *testing.T) {
	// The legacy zero-value store (Kind=="") must degrade, not panic.
	s := config.Store{}
	st := For(s).Status("", s)
	if st.Err == "" && st.Exists {
		t.Errorf("empty path should report an error or non-existence, got %+v", st)
	}
}
