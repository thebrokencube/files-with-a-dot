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
