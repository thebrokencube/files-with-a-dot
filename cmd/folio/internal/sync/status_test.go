package sync

import (
	"errors"
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

func TestCodePushRefusesMainEmitsDelegate(t *testing.T) {
	dir := gitInit(t) // starts on main
	s := config.Store{Name: "code1", Kind: config.KindCode, Path: dir, DefaultBranch: "main"}

	// On main → refuse, and NOT an ErrDelegate.
	_, err := For(s).Push(dir, s, "feat(x): y", PushOpts{})
	if err == nil {
		t.Fatal("expected refusal on main")
	}
	var del *ErrDelegate
	if errors.As(err, &del) {
		t.Fatalf("must not emit a delegate on main, got %v", err)
	}

	// On a feature branch → ErrDelegate to /commit.
	cmd := exec.Command("git", "checkout", "-q", "-b", "feature/x")
	cmd.Dir = dir
	if out, cerr := cmd.CombinedOutput(); cerr != nil {
		t.Fatalf("checkout: %v\n%s", cerr, out)
	}
	_, err = For(s).Push(dir, s, "feat(x): y", PushOpts{})
	if !errors.As(err, &del) {
		t.Fatalf("expected ErrDelegate on feature branch, got %v", err)
	}
	if del.Branch != "feature/x" || len(del.Next) != 1 || del.Next[0] != "/commit" {
		t.Errorf("delegate = %+v, want branch feature/x, next [/commit]", del)
	}
}

// jj-colocated code repo: the branch is the bookmark on @, not the (detached)
// git ref. Push resolves it, refuses without a bookmark, delegates with one.
func TestCodePushJJColocatedUsesBookmark(t *testing.T) {
	dir := gitInit(t)
	if out, err := jjRun(dir, "git", "init", "--colocate"); err != nil {
		t.Skipf("jj unavailable or colocate failed: %v (%s)", err, out)
	}
	s := config.Store{Name: "c", Kind: config.KindCode, Path: dir, DefaultBranch: "main"}

	// No bookmark on @ → refuse with jj guidance (not a git 'detached' error).
	if _, err := For(s).Push(dir, s, "feat(x): y", PushOpts{}); err == nil {
		t.Fatal("expected refusal when @ has no bookmark")
	}

	// Feature bookmark on @ → delegate on that bookmark.
	if out, jerr := jjRun(dir, "bookmark", "create", "feature-x", "-r", "@"); jerr != nil {
		t.Skipf("jj bookmark create failed: %v (%s)", jerr, out)
	}
	var del *ErrDelegate
	if _, e := For(s).Push(dir, s, "feat(x): y", PushOpts{}); !errors.As(e, &del) {
		t.Fatalf("expected ErrDelegate via bookmark, got %v", e)
	}
	if del.Branch != "feature-x" {
		t.Errorf("delegate branch = %q, want feature-x", del.Branch)
	}
}

func jjRun(dir string, args ...string) (string, error) {
	cmd := exec.Command("jj", append([]string{"--no-pager"}, args...)...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
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
