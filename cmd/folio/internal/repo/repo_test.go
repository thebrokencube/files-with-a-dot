package repo

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// initTestRepo creates a temporary git repo with user config set.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "Test"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
	return dir
}

func TestPushCreatesCommit(t *testing.T) {
	dir := initTestRepo(t)
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)

	if err := Push(dir, "test(repo): add file"); err != nil {
		t.Fatalf("Push: %v", err)
	}

	out, err := exec.Command("git", "-C", dir, "log", "--oneline").CombinedOutput()
	if err != nil {
		t.Fatalf("git log: %s", out)
	}
	if !strings.Contains(string(out), "test(repo): add file") {
		t.Errorf("git log = %q, want commit message", string(out))
	}
}

func TestPushNothingToCommit(t *testing.T) {
	dir := initTestRepo(t)
	// Create an initial commit so the repo isn't empty
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)
	if err := Push(dir, "test(repo): initial"); err != nil {
		t.Fatalf("initial Push: %v", err)
	}

	// Second push with no changes
	err := Push(dir, "test(repo): no changes")
	if !errors.Is(err, ErrNothingToCommit) {
		t.Errorf("err = %v, want ErrNothingToCommit", err)
	}
}

func TestPushInvalidMessage(t *testing.T) {
	dir := initTestRepo(t)
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)

	err := Push(dir, "bad message")
	if !errors.Is(err, ErrInvalidCommitMessage) {
		t.Errorf("err = %v, want ErrInvalidCommitMessage", err)
	}
}

func TestPushNoRemoteSkipsPush(t *testing.T) {
	dir := initTestRepo(t)
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)

	// Should succeed even without a remote (push step silently skipped)
	if err := Push(dir, "test(repo): no remote"); err != nil {
		t.Fatalf("Push: %v", err)
	}
}

func TestPullNoRemote(t *testing.T) {
	dir := initTestRepo(t)

	err := Pull(dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no remote") {
		t.Errorf("err = %q, want 'no remote'", err)
	}
}

func TestHasRemoteFalseOnLocal(t *testing.T) {
	dir := initTestRepo(t)
	if hasRemote(dir) {
		t.Error("hasRemote = true, want false on local repo")
	}
}

func TestPushScopedStagesOnlySpecifiedPaths(t *testing.T) {
	dir := initTestRepo(t)

	// Create an initial commit so repo isn't empty
	os.WriteFile(filepath.Join(dir, "init.txt"), []byte("init"), 0644)
	if err := Push(dir, "test(repo): initial"); err != nil {
		t.Fatalf("initial Push: %v", err)
	}

	// Create files in two directories
	dirA := filepath.Join(dir, "a")
	dirB := filepath.Join(dir, "b")
	os.MkdirAll(dirA, 0755)
	os.MkdirAll(dirB, 0755)
	os.WriteFile(filepath.Join(dirA, "file.txt"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(dirB, "file.txt"), []byte("b"), 0644)

	// Scope to dir "a" only
	if err := PushScoped(dir, "test(repo): scoped add", []string{"a"}); err != nil {
		t.Fatalf("PushScoped: %v", err)
	}

	// Verify dir b file is still untracked
	out, err := exec.Command("git", "-C", dir, "status", "--porcelain").CombinedOutput()
	if err != nil {
		t.Fatalf("git status: %s", out)
	}
	if !strings.Contains(string(out), "b/") {
		t.Errorf("expected b/ to be untracked, git status:\n%s", out)
	}
}

func TestPushScopedNothingToCommit(t *testing.T) {
	dir := initTestRepo(t)

	// Create an initial commit
	os.WriteFile(filepath.Join(dir, "init.txt"), []byte("init"), 0644)
	if err := Push(dir, "test(repo): initial"); err != nil {
		t.Fatalf("initial Push: %v", err)
	}

	// Scope to a dir with no changes
	os.MkdirAll(filepath.Join(dir, "empty"), 0755)
	err := PushScoped(dir, "test(repo): empty scope", []string{"empty"})
	if !errors.Is(err, ErrNothingToCommit) {
		t.Errorf("err = %v, want ErrNothingToCommit", err)
	}
}

// skipWithoutJJ skips the test if jj is not installed.
func skipWithoutJJ(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("jj"); err != nil {
		t.Skip("jj not installed, skipping jj tests")
	}
}

// initTestJJRepo creates a temporary jj repo (not colocated).
func initTestJJRepo(t *testing.T) string {
	t.Helper()
	skipWithoutJJ(t)
	dir := t.TempDir()
	cmd := exec.Command("jj", "git", "init")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("jj git init: %s", out)
	}
	return dir
}

func TestIsJJTrue(t *testing.T) {
	skipWithoutJJ(t)
	dir := initTestJJRepo(t)
	if !IsJJ(dir) {
		t.Error("IsJJ = false, want true for jj repo")
	}
}

func TestIsJJFalseOnGit(t *testing.T) {
	dir := initTestRepo(t)
	if IsJJ(dir) {
		t.Error("IsJJ = true, want false for git-only repo")
	}
}

func TestIsJJFalseOnEmpty(t *testing.T) {
	dir := t.TempDir()
	if IsJJ(dir) {
		t.Error("IsJJ = true, want false for empty dir")
	}
}

func TestHasJJRemoteFalse(t *testing.T) {
	dir := initTestJJRepo(t)
	if hasJJRemote(dir) {
		t.Error("hasJJRemote = true, want false on local jj repo")
	}
}

func TestJJPushCreatesCommit(t *testing.T) {
	dir := initTestJJRepo(t)
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)

	if err := Push(dir, "test(repo): add file"); err != nil {
		t.Fatalf("Push: %v", err)
	}

	// Verify commit exists on main bookmark
	cmd := exec.Command("jj", "--no-pager", "log", "-r", "main", "--no-graph", "-T", "description")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("jj log: %s", out)
	}
	if !strings.Contains(string(out), "test(repo): add file") {
		t.Errorf("jj log = %q, want commit message", string(out))
	}
}

func TestJJPushNothingToCommit(t *testing.T) {
	dir := initTestJJRepo(t)
	// Write, push, then try again with no changes
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)
	if err := Push(dir, "test(repo): initial"); err != nil {
		t.Fatalf("initial Push: %v", err)
	}

	err := Push(dir, "test(repo): no changes")
	if !errors.Is(err, ErrNothingToCommit) {
		t.Errorf("err = %v, want ErrNothingToCommit", err)
	}
}

func TestJJPushCreatesNewChange(t *testing.T) {
	dir := initTestJJRepo(t)
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)

	if err := Push(dir, "test(repo): first"); err != nil {
		t.Fatalf("Push: %v", err)
	}

	// After push, @ should be empty (jj new created fresh change)
	cmd := exec.Command("jj", "--no-pager", "log", "-r", "@", "--no-graph",
		"-T", `if(empty, "empty", "changed")`)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("jj log: %s", out)
	}
	if strings.TrimSpace(string(out)) != "empty" {
		t.Errorf("@ after push = %q, want empty", string(out))
	}
}

func TestJJPullNoRemoteNoBookmark(t *testing.T) {
	dir := initTestJJRepo(t)
	err := Pull(dir)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no remote configured and no main bookmark") {
		t.Errorf("err = %q, want 'no remote configured and no main bookmark'", err)
	}
}

func TestJJPullLocalOnly(t *testing.T) {
	// After a push creates the main bookmark, pull (no remote) should rebase onto it
	dir := initTestJJRepo(t)
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)
	if err := Push(dir, "test(repo): initial"); err != nil {
		t.Fatalf("Push: %v", err)
	}

	// Pull should succeed (rebase onto local main)
	if err := Pull(dir); err != nil {
		t.Fatalf("Pull: %v", err)
	}
}

func TestJJPushThenModifyAndPush(t *testing.T) {
	// Two sequential pushes with modifications between them
	dir := initTestJJRepo(t)

	// First push
	os.WriteFile(filepath.Join(dir, "a.txt"), []byte("first"), 0644)
	if err := Push(dir, "test(repo): first file"); err != nil {
		t.Fatalf("first Push: %v", err)
	}

	// Second push
	os.WriteFile(filepath.Join(dir, "b.txt"), []byte("second"), 0644)
	if err := Push(dir, "test(repo): second file"); err != nil {
		t.Fatalf("second Push: %v", err)
	}

	// Verify both commits exist on main
	cmd := exec.Command("jj", "--no-pager", "log", "-r", "root()..main", "--no-graph", "-T", "description ++ \"\\n\"")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("jj log: %s", out)
	}
	if !strings.Contains(string(out), "first file") || !strings.Contains(string(out), "second file") {
		t.Errorf("jj log = %q, want both commit messages", string(out))
	}
}

func TestJJPushScopedInScope(t *testing.T) {
	// PushScoped under jj succeeds when all changes are within allowed paths
	dir := initTestJJRepo(t)
	os.MkdirAll(filepath.Join(dir, "a"), 0755)
	os.WriteFile(filepath.Join(dir, "a", "file.txt"), []byte("scoped"), 0644)

	if err := PushScoped(dir, "test(repo): scoped push", []string{"a"}); err != nil {
		t.Fatalf("PushScoped: %v", err)
	}
}

func TestJJPushScopedOutOfScope(t *testing.T) {
	// PushScoped under jj errors when changes exist outside allowed paths
	dir := initTestJJRepo(t)
	os.MkdirAll(filepath.Join(dir, "a"), 0755)
	os.MkdirAll(filepath.Join(dir, "b"), 0755)
	os.WriteFile(filepath.Join(dir, "a", "file.txt"), []byte("in scope"), 0644)
	os.WriteFile(filepath.Join(dir, "b", "file.txt"), []byte("out of scope"), 0644)

	err := PushScoped(dir, "test(repo): should fail", []string{"a"})
	if err == nil {
		t.Fatal("expected error for out-of-scope changes, got nil")
	}
	if !strings.Contains(err.Error(), "change outside scope") {
		t.Errorf("err = %q, want 'change outside scope'", err)
	}
}

func TestJJPushInvalidMessage(t *testing.T) {
	dir := initTestJJRepo(t)
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)

	err := Push(dir, "bad message")
	if !errors.Is(err, ErrInvalidCommitMessage) {
		t.Errorf("err = %v, want ErrInvalidCommitMessage", err)
	}
}

func TestGitPushStillWorks(t *testing.T) {
	// Regression guard: git repos must continue working
	dir := initTestRepo(t)
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)

	if err := Push(dir, "test(repo): git still works"); err != nil {
		t.Fatalf("git Push after jj code added: %v", err)
	}

	// Verify it used git (not jj)
	out, err := exec.Command("git", "-C", dir, "log", "--oneline").CombinedOutput()
	if err != nil {
		t.Fatalf("git log: %s", out)
	}
	if !strings.Contains(string(out), "git still works") {
		t.Errorf("git log = %q, want commit message", string(out))
	}
}

func TestIsJJDispatchCorrectly(t *testing.T) {
	// Verify that a git-only repo doesn't accidentally take jj path
	dir := initTestRepo(t)
	if IsJJ(dir) {
		t.Fatal("IsJJ returned true for git-only repo")
	}

	// Verify push uses git path (no jj commands should run)
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("dispatch"), 0644)
	if err := Push(dir, "test(repo): dispatch check"); err != nil {
		t.Fatalf("Push: %v", err)
	}
}

func TestValidateCommitMessage(t *testing.T) {
	tests := []struct {
		name    string
		msg     string
		wantErr bool
	}{
		// Valid
		{"basic feat", "feat(srm): add spike sources", false},
		{"docs update", "docs(folio): update readme", false},
		{"auto codegen", "auto(codegen): graphqlme", false},
		{"fix with dots in scope", "fix(ben.payroll): handle nil", false},
		{"chore with hyphen scope", "chore(ci-lint): add step", false},
		{"multi-line with body", "feat(folio): add validation\n\nMore details here.", false},

		// Invalid
		{"empty", "", true},
		{"old default", "folio: update", true},
		{"no scope", "feat: add thing", true},
		{"uppercase description", "feat(bar): Thing", true},
		{"trailing period", "docs(folio): update readme.", true},
		{"bare text", "update stuff", true},
		{"bad type", "yolo(foo): do thing", true},
		{"uppercase scope", "feat(Foo): bar", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommitMessage(tt.msg)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for %q, got nil", tt.msg)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for %q: %v", tt.msg, err)
			}
			if tt.wantErr && err != nil && !errors.Is(err, ErrInvalidCommitMessage) {
				t.Errorf("expected ErrInvalidCommitMessage, got: %v", err)
			}
		})
	}
}

func TestIsWorkspace(t *testing.T) {
	tests := []struct {
		dir  string
		want bool
	}{
		{"/tmp/folio-ws-1234567890-12345", true},
		{"/tmp/folio-ws-0-0", true},
		{"/Users/me/.folio", false},
		{"/tmp/other-dir", false},
		{"", false},
		{"/tmp/folio-ws-", true}, // edge: empty suffix still has prefix
	}
	for _, tt := range tests {
		t.Run(tt.dir, func(t *testing.T) {
			if got := isWorkspace(tt.dir); got != tt.want {
				t.Errorf("isWorkspace(%q) = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}

// initTestJJRepoWithWorkspace creates a jj repo and a workspace named "folio-ws-test"
// so isWorkspace() returns true. Returns (repoDir, workspaceDir).
func initTestJJRepoWithWorkspace(t *testing.T) (string, string) {
	t.Helper()
	return initTestJJRepoWithNamedWorkspace(t, "folio-ws-test")
}

// initTestJJRepoWithNamedWorkspace creates a jj repo and a workspace with the given name.
// Returns (repoDir, workspaceDir).
func initTestJJRepoWithNamedWorkspace(t *testing.T, wsName string) (string, string) {
	t.Helper()
	skipWithoutJJ(t)

	base := t.TempDir()
	repoDir := filepath.Join(base, "repo")
	wsDir := filepath.Join(base, wsName)

	os.MkdirAll(repoDir, 0755)

	cmd := exec.Command("jj", "git", "init")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("jj git init: %s", out)
	}

	// Create initial commit so main bookmark exists
	os.WriteFile(filepath.Join(repoDir, "init.txt"), []byte("init"), 0644)
	if err := Push(repoDir, "test(repo): initial"); err != nil {
		t.Fatalf("initial Push: %v", err)
	}

	// Create workspace
	cmd = exec.Command("jj", "--no-pager", "workspace", "add", wsDir, "-r", "main", "-R", repoDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("jj workspace add: %v", err)
	}

	return repoDir, wsDir
}

func TestJJPushFromWorkspaceRebasesDefault(t *testing.T) {
	repoDir, wsDir := initTestJJRepoWithWorkspace(t)

	// Write a file in the workspace and push
	os.WriteFile(filepath.Join(wsDir, "from-ws.txt"), []byte("workspace change"), 0644)
	if err := Push(wsDir, "test(repo): push from workspace"); err != nil {
		t.Fatalf("Push from workspace: %v", err)
	}

	// Verify default workspace's @ parent is main
	cmd := exec.Command("jj", "--no-pager", "log", "-r", "@", "--no-graph",
		"-T", `separate(" ", parents.map(|p| if(p.bookmarks().len() > 0, p.bookmarks(), "none")))`)
	cmd.Dir = repoDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("jj log: %s", out)
	}
	if !strings.Contains(string(out), "main") {
		t.Errorf("default workspace @ parent = %q, want parent to be main", strings.TrimSpace(string(out)))
	}

	// Verify git HEAD is attached to main (not detached) in the colocated repo
	cmd = exec.Command("git", "symbolic-ref", "HEAD")
	cmd.Dir = repoDir
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git symbolic-ref HEAD: %s", out)
	}
	if strings.TrimSpace(string(out)) != "refs/heads/main" {
		t.Errorf("git HEAD = %q, want refs/heads/main", strings.TrimSpace(string(out)))
	}
}

func TestIsColocated(t *testing.T) {
	// jj git init creates a colocated repo (.jj + .git)
	skipWithoutJJ(t)
	dir := initTestJJRepo(t)
	if !isColocated(dir) {
		t.Error("isColocated = false, want true for jj git init repo")
	}

	// plain git repo is not colocated
	gitDir := initTestRepo(t)
	if isColocated(gitDir) {
		t.Error("isColocated = true, want false for git-only repo")
	}

	// empty dir is not colocated
	emptyDir := t.TempDir()
	if isColocated(emptyDir) {
		t.Error("isColocated = true, want false for empty dir")
	}
}

func TestJJPushFromRepoRootSkipsRebase(t *testing.T) {
	// When pushing from the repo root (not a workspace), no rebase should happen
	// This is a regression guard — isWorkspace should return false for repo root
	dir := initTestJJRepo(t)
	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("hello"), 0644)

	// Should succeed without errors (no workspace rebase attempted)
	if err := Push(dir, "test(repo): from root"); err != nil {
		t.Fatalf("Push: %v", err)
	}
}

func TestJJPushFromTwoWorkspacesSequential(t *testing.T) {
	// Simulate two sessions pushing sequentially — default should track main after each
	skipWithoutJJ(t)

	base := t.TempDir()
	repoDir := filepath.Join(base, "repo")
	wsA := filepath.Join(base, "folio-ws-session-a")
	wsB := filepath.Join(base, "folio-ws-session-b")

	os.MkdirAll(repoDir, 0755)

	cmd := exec.Command("jj", "git", "init")
	cmd.Dir = repoDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("jj git init: %s", out)
	}

	// Bootstrap main
	os.WriteFile(filepath.Join(repoDir, "init.txt"), []byte("init"), 0644)
	if err := Push(repoDir, "test(repo): initial"); err != nil {
		t.Fatalf("initial Push: %v", err)
	}

	// Create two workspaces
	for _, ws := range []string{wsA, wsB} {
		cmd = exec.Command("jj", "--no-pager", "workspace", "add", ws, "-r", "main", "-R", repoDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			t.Fatalf("jj workspace add %s: %v", ws, err)
		}
	}

	// Session A pushes
	os.WriteFile(filepath.Join(wsA, "a.txt"), []byte("from A"), 0644)
	if err := Push(wsA, "test(repo): session a push"); err != nil {
		t.Fatalf("Push from A: %v", err)
	}

	// Verify default is on main after A's push
	cmd = exec.Command("jj", "--no-pager", "log", "-r", "@", "--no-graph",
		"-T", `separate(" ", parents.map(|p| if(p.bookmarks().len() > 0, p.bookmarks(), "none")))`)
	cmd.Dir = repoDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("jj log after A: %s", out)
	}
	if !strings.Contains(string(out), "main") {
		t.Errorf("after A: default @ parent = %q, want main", strings.TrimSpace(string(out)))
	}

	// Session B pushes (main has moved since B was created)
	os.WriteFile(filepath.Join(wsB, "b.txt"), []byte("from B"), 0644)
	if err := Push(wsB, "test(repo): session b push"); err != nil {
		t.Fatalf("Push from B: %v", err)
	}

	// Verify default is still on main after B's push
	cmd = exec.Command("jj", "--no-pager", "log", "-r", "@", "--no-graph",
		"-T", `separate(" ", parents.map(|p| if(p.bookmarks().len() > 0, p.bookmarks(), "none")))`)
	cmd.Dir = repoDir
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("jj log after B: %s", out)
	}
	if !strings.Contains(string(out), "main") {
		t.Errorf("after B: default @ parent = %q, want main", strings.TrimSpace(string(out)))
	}
}

func TestJJPushFromWorkspaceDefaultDirty(t *testing.T) {
	// If default workspace has uncommitted changes, push should still succeed
	// (rebase is best-effort — warn but don't fail)
	repoDir, wsDir := initTestJJRepoWithWorkspace(t)

	// Dirty the default workspace
	os.WriteFile(filepath.Join(repoDir, "dirty.txt"), []byte("uncommitted"), 0644)

	// Push from workspace — should succeed despite dirty default
	os.WriteFile(filepath.Join(wsDir, "ws.txt"), []byte("workspace"), 0644)
	if err := Push(wsDir, "test(repo): push with dirty default"); err != nil {
		t.Fatalf("Push should succeed even with dirty default: %v", err)
	}
}

func TestJJPushFromWorkspaceNoDefault(t *testing.T) {
	// If the default workspace has been forgotten, push should still succeed
	// (repoRoot fails, best-effort skips rebase)
	repoDir, wsDir := initTestJJRepoWithWorkspace(t)

	// Forget the default workspace
	cmd := exec.Command("jj", "--no-pager", "--quiet", "workspace", "forget", "default", "-R", repoDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("jj workspace forget: %v", err)
	}

	// Push from workspace — should succeed without default workspace
	os.WriteFile(filepath.Join(wsDir, "ws.txt"), []byte("workspace"), 0644)
	if err := Push(wsDir, "test(repo): push without default"); err != nil {
		t.Fatalf("Push should succeed even without default workspace: %v", err)
	}
}
