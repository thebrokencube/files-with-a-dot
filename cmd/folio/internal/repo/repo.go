package repo

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// ErrNothingToCommit is returned when there are no changes to commit.
var ErrNothingToCommit = errors.New("nothing to commit")

// ErrInvalidCommitMessage is returned when a commit message doesn't follow conventional commit format.
var ErrInvalidCommitMessage = errors.New("invalid commit message")

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
	if !hasRemote(home) {
		return fmt.Errorf("no remote configured")
	}
	if err := gitIn(home, "pull"); err != nil {
		return fmt.Errorf("git pull: %w", err)
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

// gitOutput runs a git command and captures stdout.
func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return string(out), err
}
