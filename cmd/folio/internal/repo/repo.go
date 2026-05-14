package repo

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// ErrNothingToCommit is returned when there are no changes to commit.
var ErrNothingToCommit = errors.New("nothing to commit")

// ErrInvalidCommitMessage is returned when a commit message doesn't follow conventional commit format.
var ErrInvalidCommitMessage = errors.New("invalid commit message")

// ErrConflict is returned when a rebase produces content conflicts.
var ErrConflict = errors.New("rebase conflict")

var commitMsgRe = regexp.MustCompile(`^(feat|fix|docs|refactor|test|chore|style|perf|auto)\([a-z][a-z0-9._-]*\): [a-z].+$`)

// ValidateCommitMessage checks that message follows conventional commit format:
// type(scope): description
func ValidateCommitMessage(message string) error {
	if message == "" {
		return fmt.Errorf("%w: message is empty", ErrInvalidCommitMessage)
	}

	firstLine := strings.SplitN(message, "\n", 2)[0]

	if !commitMsgRe.MatchString(firstLine) {
		return fmt.Errorf("%w: must match type(scope): description\n  allowed types: feat fix docs refactor test chore style perf auto", ErrInvalidCommitMessage)
	}

	if strings.HasSuffix(firstLine, ".") {
		return fmt.Errorf("%w: description must not end with a period", ErrInvalidCommitMessage)
	}

	return nil
}

// Push stages all changes and commits in the FOLIO_HOME directory.
// If a remote is configured, it also pushes.
func Push(home, message string) error {
	if err := ValidateCommitMessage(message); err != nil {
		return err
	}

	if IsJJ(home) {
		return jjPush(home, message)
	}

	// git add -A
	if err := gitIn(home, "add", "-A"); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	// Check if there's anything to commit (diff --cached is empty = nothing staged)
	if err := gitQuiet(home, "diff", "--cached", "--quiet"); err == nil {
		return ErrNothingToCommit
	}

	// git commit
	if err := gitIn(home, "commit", "-m", message); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	// Push if remote exists
	if hasRemote(home) {
		if err := gitIn(home, "push"); err != nil {
			return fmt.Errorf("git push: %w", err)
		}
	}

	return nil
}

// Pull runs git pull in the FOLIO_HOME directory.
func Pull(home string) error {
	if _, err := os.Stat(home); os.IsNotExist(err) {
		return fmt.Errorf("FOLIO_HOME directory does not exist: %s (workspace may have been cleaned up)", home)
	}

	if IsJJ(home) {
		return jjPull(home)
	}

	if !hasRemote(home) {
		return fmt.Errorf("no remote configured")
	}
	if err := gitIn(home, "pull"); err != nil {
		return fmt.Errorf("git pull: %w", err)
	}
	return nil
}

// isWorkspace returns true if dir is a jj workspace (not the repo root).
// Workspaces are created in /tmp with a "folio-ws-" prefix.
func isWorkspace(dir string) bool {
	return strings.HasPrefix(filepath.Base(dir), "folio-ws-")
}

// isColocated returns true if dir has both .jj and .git (jj+git colocated repo).
func isColocated(dir string) bool {
	_, jjErr := os.Stat(filepath.Join(dir, ".jj"))
	_, gitErr := os.Stat(filepath.Join(dir, ".git"))
	return jjErr == nil && gitErr == nil
}

// defaultWorkspaceRoot returns the default workspace root for a jj repo.
func defaultWorkspaceRoot(dir string) (string, error) {
	out, err := jjOutput(dir, "workspace", "root", "--name", "default")
	if err != nil {
		return "", fmt.Errorf("jj workspace root: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// jjPush describes @, sets bookmark, pushes, and creates fresh @.
func jjPush(home, message string) error {
	// Check if @ is empty
	out, err := jjOutput(home, "log", "-r", "@", "--no-graph", "-T", `if(empty, "empty", "changed")`)
	if err != nil {
		return fmt.Errorf("jj log: %w", err)
	}
	if strings.TrimSpace(out) == "empty" {
		return ErrNothingToCommit
	}

	// Describe the change
	if err := jjIn(home, "describe", "-m", message); err != nil {
		return fmt.Errorf("jj describe: %w", err)
	}

	// Rebase onto main to prevent bookmark divergence (skip if main doesn't exist yet)
	if hasBookmark(home, "main") {
		if err := jjIn(home, "rebase", "-d", "main"); err != nil {
			return fmt.Errorf("jj rebase: %w", err)
		}
		// jj rebase exits 0 even when it produces conflicts, so check explicitly
		cout, cerr := jjOutput(home, "log", "-r", "@", "--no-graph", "-T", "conflict")
		if cerr == nil && strings.TrimSpace(cout) == "true" {
			return fmt.Errorf("%w: resolve conflicts in %s, then retry folio home push", ErrConflict, home)
		}
	}

	// Set bookmark to current change
	if err := jjIn(home, "bookmark", "set", "main", "-r", "@"); err != nil {
		return fmt.Errorf("jj bookmark set: %w", err)
	}

	// Push if remote exists
	if hasJJRemote(home) {
		if err := jjIn(home, "git", "push", "--bookmark", "main"); err != nil {
			return fmt.Errorf("jj git push: %w", err)
		}
	}

	// Fresh @ for next operation
	if err := jjIn(home, "new"); err != nil {
		return fmt.Errorf("jj new: %w", err)
	}

	// If pushing from a workspace, rebase the default workspace onto main
	// so it stays current without manual intervention.
	if isWorkspace(home) {
		if root, err := defaultWorkspaceRoot(home); err == nil {
			if rebErr := jjIn(root, "rebase", "-r", "@", "-d", "main"); rebErr != nil {
				fmt.Fprintf(os.Stderr, "warning: could not rebase default workspace: %s\n", rebErr)
			}
			// Re-attach git HEAD to main in colocated repos so git status
			// shows "on branch main" instead of detached HEAD.
			// Uses symbolic-ref (not checkout) to avoid touching the working tree.
			if isColocated(root) {
				_ = gitIn(root, "symbolic-ref", "HEAD", "refs/heads/main")
			}
		}
	}

	return nil
}

// jjPull fetches and rebases onto main@origin, or rebases onto local main if no remote.
func jjPull(home string) error {
	if hasJJRemote(home) {
		if err := jjIn(home, "git", "fetch"); err != nil {
			return fmt.Errorf("jj git fetch: %w", err)
		}
		if err := jjIn(home, "rebase", "-d", "main@origin"); err != nil {
			return fmt.Errorf("jj rebase: %w", err)
		}
	} else if hasBookmark(home, "main") {
		if err := jjIn(home, "rebase", "-d", "main"); err != nil {
			return fmt.Errorf("jj rebase: %w", err)
		}
	} else {
		return fmt.Errorf("no remote configured and no main bookmark")
	}
	return nil
}

// hasRemote checks if a git remote is configured.
func hasRemote(dir string) bool {
	out, err := gitOutput(dir, "remote")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) != ""
}

// gitIn runs a git command in the specified directory with stdout/stderr.
func gitIn(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// gitQuiet runs a git command silently (no output).
func gitQuiet(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	return cmd.Run()
}

// GitOutput runs a git command in dir and captures stdout.
func GitOutput(dir string, args ...string) (string, error) {
	return gitOutput(dir, args...)
}

// gitOutput runs a git command and captures stdout.
func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return string(out), err
}

// IsJJ returns true if dir contains a .jj directory (jj-managed repo).
func IsJJ(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".jj"))
	return err == nil
}

// jjIn runs a jj command in dir with --no-pager --quiet.
// Stdout and stderr go to os.Stdout/os.Stderr.
func jjIn(dir string, args ...string) error {
	full := append([]string{"--no-pager", "--quiet"}, args...)
	cmd := exec.Command("jj", full...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// jjOutput runs a jj command in dir with --no-pager and captures stdout.
// Does NOT use --quiet (callers need to parse output).
func jjOutput(dir string, args ...string) (string, error) {
	full := append([]string{"--no-pager"}, args...)
	cmd := exec.Command("jj", full...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return string(out), err
}

// hasBookmark checks if a jj bookmark exists.
func hasBookmark(dir, name string) bool {
	out, err := jjOutput(dir, "bookmark", "list", name)
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) != ""
}

// hasJJRemote checks if a jj git remote is configured.
func hasJJRemote(dir string) bool {
	out, err := jjOutput(dir, "git", "remote", "list")
	if err != nil {
		return false
	}
	return strings.TrimSpace(out) != ""
}
