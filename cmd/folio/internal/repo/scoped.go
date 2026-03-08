package repo

import (
	"fmt"
	"path/filepath"
)

// PushScoped stages only files under the given paths, then commits and pushes.
// paths are relative to home (e.g., "active/files-with-a-dot").
func PushScoped(home, message string, paths []string) error {
	if err := ValidateCommitMessage(message); err != nil {
		return err
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
