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

// A worktree made by hand — beside the repo, not under .worktrees — is invisible
// to Reconcile, which walks only the canonical root. ScanVCS is what surfaces it,
// so this is the load-bearing case: without it, `workarea list` reports "no work
// areas" while checkouts pile up next to the repo.
func TestScanVCSSurfacesHandMadeGitWorktree(t *testing.T) {
	umbrella := t.TempDir()
	repo := gitRepo(t)
	store := config.Store{Name: "code1", Kind: config.KindCode, Path: repo, DefaultBranch: "main"}

	beside := filepath.Join(filepath.Dir(repo), "code1-sibling")
	runOK(t, repo, "git", "worktree", "add", "-q", beside, "-b", "sibling")

	// Reconcile alone sees nothing — it only knows the ledger and .worktrees.
	rows, orphans, err := Reconcile(umbrella)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 0 || len(orphans) != 0 {
		t.Fatalf("Reconcile should not see a hand-made sibling: rows=%+v orphans=%v", rows, orphans)
	}

	areas := ScanVCS(umbrella, []config.Store{store})
	if len(areas) != 1 {
		t.Fatalf("ScanVCS areas = %+v, want exactly the sibling", areas)
	}
	if areas[0].State != StateStray {
		t.Errorf("state = %q, want %q", areas[0].State, StateStray)
	}
	if !sameDir(areas[0].Dir, beside) {
		t.Errorf("dir = %q, want %q", areas[0].Dir, beside)
	}
	if areas[0].Tier != string(TierGitWorktree) {
		t.Errorf("tier = %q, want %q", areas[0].Tier, TierGitWorktree)
	}
}

// A registration whose directory was deleted behind the VCS's back is the residue
// that accumulates when sessions exit without deregistering.
func TestScanVCSReportsDanglingRegistration(t *testing.T) {
	umbrella := t.TempDir()
	repo := gitRepo(t)
	store := config.Store{Name: "code1", Kind: config.KindCode, Path: repo, DefaultBranch: "main"}

	beside := filepath.Join(filepath.Dir(repo), "code1-deleted")
	runOK(t, repo, "git", "worktree", "add", "-q", beside, "-b", "deleted")
	if err := os.RemoveAll(beside); err != nil {
		t.Fatal(err)
	}

	areas := ScanVCS(umbrella, []config.Store{store})
	if len(areas) != 1 || areas[0].State != StateDangling {
		t.Fatalf("areas = %+v, want one %q", areas, StateDangling)
	}
}

// An area folio placed itself is already covered by Reconcile, so ScanVCS must
// not report it a second time.
func TestScanVCSSkipsLedgeredAndMainWorkingCopy(t *testing.T) {
	umbrella := t.TempDir()
	repo := gitRepo(t)
	store := config.Store{Name: "code1", Kind: config.KindCode, Path: repo, DefaultBranch: "main"}

	if _, err := Open(umbrella, store, "feature/y", "", ""); err != nil {
		t.Fatal(err)
	}
	if areas := ScanVCS(umbrella, []config.Store{store}); len(areas) != 0 {
		t.Fatalf("ScanVCS reported %+v; a ledgered area and the main copy must be skipped", areas)
	}
}

func TestScanVCSSurfacesHandMadeJJWorkspace(t *testing.T) {
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj not on PATH")
	}
	umbrella := t.TempDir()
	repo := t.TempDir()
	runOK(t, repo, "jj", "git", "init", "--colocate")
	store := config.Store{Name: "code1", Kind: config.KindCode, Path: repo, DefaultBranch: "main"}

	beside := filepath.Join(filepath.Dir(repo), "code1-jj-sibling")
	runOK(t, repo, "jj", "--no-pager", "workspace", "add", "--name", "sibling", beside)

	areas := ScanVCS(umbrella, []config.Store{store})
	// A colocated repo is probed for both tiers; only the jj workspace is an area.
	var jjAreas []Unledgered
	for _, a := range areas {
		if a.Tier == string(TierJJ) {
			jjAreas = append(jjAreas, a)
		}
	}
	if len(jjAreas) != 1 {
		t.Fatalf("jj areas = %+v, want exactly the sibling (default must be skipped)", areas)
	}
	if jjAreas[0].Name != "sibling" || jjAreas[0].State != StateStray {
		t.Errorf("area = %+v, want name=sibling state=%s", jjAreas[0], StateStray)
	}
	if !sameDir(jjAreas[0].Dir, beside) {
		t.Errorf("dir = %q, want %q", jjAreas[0].Dir, beside)
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
