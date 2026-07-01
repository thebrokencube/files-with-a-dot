package fleet

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

func gitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	runOK(t, dir, "git", "init", "-q", "-b", "main")
	runOK(t, dir, "git", "config", "user.email", "t@t")
	runOK(t, dir, "git", "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("hi"), 0644); err != nil {
		t.Fatal(err)
	}
	runOK(t, dir, "git", "add", "-A")
	runOK(t, dir, "git", "commit", "-q", "-m", "init")
	return dir
}

func runOK(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func TestDetectTier(t *testing.T) {
	git := gitRepo(t)
	if got := DetectTier(git); got != TierGitWorktree {
		t.Errorf("git repo tier = %q, want %q", got, TierGitWorktree)
	}
	if got := DetectTier(t.TempDir()); got != TierNone {
		t.Errorf("plain dir tier = %q, want %q", got, TierNone)
	}
}

func TestOpenListReapGitWorktree(t *testing.T) {
	umbrella := t.TempDir()
	repo := gitRepo(t)
	store := config.Store{Name: "code1", Kind: config.KindCode, Path: repo, DefaultBranch: "main"}

	wa, err := Open(umbrella, store, "feature/x", "", "sess1")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if wa.Tier != string(TierGitWorktree) {
		t.Errorf("tier = %q", wa.Tier)
	}
	if !pathExists(wa.Dir) {
		t.Fatalf("work area dir not created: %s", wa.Dir)
	}
	if want := CanonicalDir(umbrella, "code1", "feature/x"); wa.Dir != want {
		t.Errorf("dir = %q, want canonical %q", wa.Dir, want)
	}

	// Ledger has exactly one row.
	areas, _ := ReadLedger(umbrella)
	if len(areas) != 1 {
		t.Fatalf("ledger rows = %d, want 1", len(areas))
	}

	// Reconcile: one ok row, no orphans.
	rows, orphans, err := Reconcile(umbrella)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].State != "ok" {
		t.Fatalf("reconcile rows = %+v", rows)
	}
	if len(orphans) != 0 {
		t.Errorf("unexpected orphans: %v", orphans)
	}

	// Reap the clean area: removed, ledger emptied, no stale .git/worktrees entry.
	actions, err := Reap(umbrella, "feature/x", false)
	if err != nil {
		t.Fatalf("Reap: %v", err)
	}
	if len(actions) == 0 {
		t.Fatal("expected reap actions")
	}
	if pathExists(wa.Dir) {
		t.Errorf("work area dir still present after reap: %s", wa.Dir)
	}
	areas, _ = ReadLedger(umbrella)
	if len(areas) != 0 {
		t.Errorf("ledger not pruned: %+v", areas)
	}
	// The main repo must have no lingering worktree registration.
	out, _ := run(repo, "git", "worktree", "list", "--porcelain")
	if strings.Contains(out, wa.Dir) {
		t.Errorf("stale git worktree registration remains:\n%s", out)
	}
}

func TestReapDirtyGuard(t *testing.T) {
	umbrella := t.TempDir()
	repo := gitRepo(t)
	store := config.Store{Name: "code1", Kind: config.KindCode, Path: repo, DefaultBranch: "main"}
	wa, err := Open(umbrella, store, "wip", "", "")
	if err != nil {
		t.Fatal(err)
	}
	// Make the work area dirty.
	if err := os.WriteFile(filepath.Join(wa.Dir, "new.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	actions, err := Reap(umbrella, "wip", false)
	if err != nil {
		t.Fatal(err)
	}
	if !pathExists(wa.Dir) {
		t.Errorf("dirty work area was removed without --force")
	}
	if !anyContains(actions, "dirty") {
		t.Errorf("expected a dirty-kept action, got %v", actions)
	}
}

// A worktree with committed-but-never-pushed work (no upstream configured) must
// NOT be reaped on a plain reap — otherwise the commits are silently lost.
func TestReapKeepsUnpushedCommitsWithoutUpstream(t *testing.T) {
	umbrella := t.TempDir()
	repo := gitRepo(t)
	store := config.Store{Name: "code1", Kind: config.KindCode, Path: repo, DefaultBranch: "main"}
	wa, err := Open(umbrella, store, "wip", "", "")
	if err != nil {
		t.Fatal(err)
	}
	// Commit work in the worktree (ahead of main, no upstream).
	if err := os.WriteFile(filepath.Join(wa.Dir, "work.txt"), []byte("important"), 0644); err != nil {
		t.Fatal(err)
	}
	runOK(t, wa.Dir, "git", "add", "-A")
	runOK(t, wa.Dir, "git", "commit", "-q", "-m", "wip work")

	actions, err := Reap(umbrella, "wip", false)
	if err != nil {
		t.Fatal(err)
	}
	if !pathExists(wa.Dir) {
		t.Fatal("worktree with unpushed commits was reaped — data loss!")
	}
	if !anyContains(actions, "dirty/unpushed") {
		t.Errorf("expected unpushed-kept action, got %v", actions)
	}
}

func anyContains(xs []string, sub string) bool {
	for _, x := range xs {
		if strings.Contains(x, sub) {
			return true
		}
	}
	return false
}
