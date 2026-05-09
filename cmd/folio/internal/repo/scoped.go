package repo

import (
	"fmt"
	"path/filepath"
	"strings"
)

// PushScoped stages only files under the given paths, then commits and pushes.
// paths are relative to home (e.g., "active/files-with-a-dot").
func PushScoped(home, message string, paths []string) error {
	if err := ValidateCommitMessage(message); err != nil {
		return err
	}

	if IsJJ(home) {
		return jjPushScoped(home, message, paths)
	}

	for _, p := range paths {
		abs := filepath.Join(home, p)
		if err := gitIn(home, "add", abs); err != nil {
			return fmt.Errorf("git add %s: %w", p, err)
		}
	}

	if err := gitQuiet(home, "diff", "--cached", "--quiet"); err == nil {
		return ErrNothingToCommit
	}

	if err := gitIn(home, "commit", "-m", message); err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	if hasRemote(home) {
		if err := gitIn(home, "push"); err != nil {
			return fmt.Errorf("git push: %w", err)
		}
	}

	return nil
}

// jjPushScoped pushes only if @ changes are within the specified paths.
// Errors if @ has changes outside the specified paths.
func jjPushScoped(home, message string, paths []string) error {
	// Get changed files in @
	out, err := jjOutput(home, "diff", "--summary")
	if err != nil {
		return fmt.Errorf("jj diff --summary: %w", err)
	}

	// Parse changed file paths and check scope
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		// jj diff --summary format: "M path/to/file" or "A path/to/file"
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		filePath := parts[1]
		inScope := false
		for _, p := range paths {
			if strings.HasPrefix(filePath, p) {
				inScope = true
				break
			}
		}
		if !inScope {
			return fmt.Errorf("change outside scope: %s (allowed: %v)", filePath, paths)
		}
	}

	return jjPush(home, message)
}
