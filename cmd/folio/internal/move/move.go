package move

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var datePrefixRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-`)

// Archive moves active/<path> to archive/<path> with a date prefix on the leaf
// directory. For example, active/ben/pb-on-call/stride-health-secrets becomes
// archive/ben/pb-on-call/2026-02-28-stride-health-secrets.
func Archive(home, relPath string) error {
	if err := validateRelPath(relPath); err != nil {
		return err
	}

	src := filepath.Join(home, "active", relPath)
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("not found: active/%s", relPath)
	}

	dir := filepath.Dir(relPath)
	base := filepath.Base(relPath)
	dated := time.Now().Format("2006-01-02") + "-" + base

	destDir := filepath.Join(home, "archive", dir)
	dest := filepath.Join(destDir, dated)

	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("already exists: archive/%s", filepath.Join(dir, dated))
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	if err := os.Rename(src, dest); err != nil {
		return fmt.Errorf("rename: %w", err)
	}

	// Clean up empty ancestor directories up to (but not including) active/
	sectionDir := filepath.Join(home, "active")
	pruneEmptyAncestors(filepath.Dir(src), sectionDir)

	return nil
}

// Activate moves archive/<path> to active/<path>, stripping the date prefix
// from the leaf directory. For example, archive/ben/pb-on-call/2026-02-28-stride-health-secrets
// becomes active/ben/pb-on-call/stride-health-secrets.
func Activate(home, relPath string) error {
	if err := validateRelPath(relPath); err != nil {
		return err
	}

	src := filepath.Join(home, "archive", relPath)
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("not found: archive/%s", relPath)
	}

	dir := filepath.Dir(relPath)
	base := filepath.Base(relPath)
	stripped := stripDatePrefix(base)

	destDir := filepath.Join(home, "active", dir)
	dest := filepath.Join(destDir, stripped)

	if _, err := os.Stat(dest); err == nil {
		return fmt.Errorf("already exists: active/%s", filepath.Join(dir, stripped))
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	if err := os.Rename(src, dest); err != nil {
		return fmt.Errorf("rename: %w", err)
	}

	// Clean up empty ancestor directories up to (but not including) archive/
	sectionDir := filepath.Join(home, "archive")
	pruneEmptyAncestors(filepath.Dir(src), sectionDir)

	return nil
}

// validateRelPath checks that a relative path is non-empty and meaningful.
func validateRelPath(relPath string) error {
	cleaned := filepath.Clean(relPath)
	if cleaned == "" || cleaned == "." {
		return fmt.Errorf("invalid path: %q", relPath)
	}
	return nil
}

// pruneEmptyAncestors removes empty directories from dir up to (but not
// including) stopAt. Walks upward, removing each empty directory until it
// hits stopAt or a non-empty directory.
func pruneEmptyAncestors(dir, stopAt string) {
	for dir != stopAt && dir != "." && dir != "/" {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			return
		}
		os.Remove(dir)
		dir = filepath.Dir(dir)
	}
}

// stripDatePrefix removes the YYYY-MM-DD- prefix from a name if present.
func stripDatePrefix(name string) string {
	if datePrefixRe.MatchString(name) {
		return name[11:]
	}
	return name
}
