package move

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/list"
)

var datePrefixRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-`)

type MoveResult struct {
	Source      string
	Destination string

	createdParents []string
}

// Archive moves active/<path> to archive/<path> with a date prefix on the leaf
// directory.
func Archive(home, relPath string) (MoveResult, error) {
	if err := validateRelPath(relPath); err != nil {
		return MoveResult{}, err
	}
	root, err := config.CanonicalPath(home)
	if err != nil {
		return MoveResult{}, err
	}
	src, err := config.CanonicalPath(filepath.Join(root, "active", relPath))
	if err != nil {
		return MoveResult{}, err
	}
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return MoveResult{}, fmt.Errorf("not found: active/%s", relPath)
	} else if err != nil {
		return MoveResult{}, fmt.Errorf("stat active/%s: %w", relPath, err)
	}

	dir := filepath.Dir(filepath.Clean(relPath))
	base := filepath.Base(filepath.Clean(relPath))
	dated := time.Now().Format("2006-01-02") + "-" + base
	destDir := filepath.Join(root, "archive", dir)
	dest, err := config.CanonicalPath(filepath.Join(destDir, dated))
	if err != nil {
		return MoveResult{}, err
	}
	if _, err := os.Stat(dest); err == nil {
		return MoveResult{}, fmt.Errorf("already exists: archive/%s", filepath.Join(dir, dated))
	} else if !os.IsNotExist(err) {
		return MoveResult{}, fmt.Errorf("stat archive/%s: %w", filepath.Join(dir, dated), err)
	}

	createdDirs := missingParents(destDir, filepath.Join(root, "archive"))
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return MoveResult{}, fmt.Errorf("mkdir: %w", err)
	}
	if err := os.Rename(src, dest); err != nil {
		removeParents(createdDirs)
		return MoveResult{}, fmt.Errorf("rename: %w", err)
	}

	pruneEmptyAncestors(filepath.Dir(src), filepath.Join(root, "active"))
	return MoveResult{
		Source:         src,
		Destination:    dest,
		createdParents: createdDirs,
	}, nil
}

// Activate moves archive/<path> to active/<path>, stripping the date prefix
// from the leaf directory.
func Activate(home, relPath string) (MoveResult, error) {
	if err := validateRelPath(relPath); err != nil {
		return MoveResult{}, err
	}
	root, err := config.CanonicalPath(home)
	if err != nil {
		return MoveResult{}, err
	}
	src, err := config.CanonicalPath(filepath.Join(root, "archive", relPath))
	if err != nil {
		return MoveResult{}, err
	}
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return MoveResult{}, fmt.Errorf("not found: archive/%s", relPath)
	} else if err != nil {
		return MoveResult{}, fmt.Errorf("stat archive/%s: %w", relPath, err)
	}

	cleanRelPath := filepath.Clean(relPath)
	dir := filepath.Dir(cleanRelPath)
	base := filepath.Base(cleanRelPath)
	stripped := stripDatePrefix(base)
	destDir := filepath.Join(root, "active", dir)
	dest, err := config.CanonicalPath(filepath.Join(destDir, stripped))
	if err != nil {
		return MoveResult{}, err
	}
	if _, err := os.Stat(dest); err == nil {
		return MoveResult{}, fmt.Errorf("already exists: active/%s", filepath.Join(dir, stripped))
	} else if !os.IsNotExist(err) {
		return MoveResult{}, fmt.Errorf("stat active/%s: %w", filepath.Join(dir, stripped), err)
	}

	createdDirs := missingParents(destDir, filepath.Join(root, "active"))
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return MoveResult{}, fmt.Errorf("mkdir: %w", err)
	}
	if err := os.Rename(src, dest); err != nil {
		removeParents(createdDirs)
		return MoveResult{}, fmt.Errorf("rename: %w", err)
	}

	pruneEmptyAncestors(filepath.Dir(src), filepath.Join(root, "archive"))
	return MoveResult{
		Source:         src,
		Destination:    dest,
		createdParents: createdDirs,
	}, nil
}

// Rollback reverses a successful archive or activate movement.
func Rollback(result MoveResult) error {
	if result.Source == "" || result.Destination == "" {
		return fmt.Errorf("movement result is incomplete")
	}
	if _, err := os.Stat(result.Source); err == nil {
		return fmt.Errorf("rollback source already exists: %s", result.Source)
	}
	if _, err := os.Stat(result.Destination); err != nil {
		return fmt.Errorf("rollback destination is missing: %s: %w", result.Destination, err)
	}
	if err := os.MkdirAll(filepath.Dir(result.Source), 0755); err != nil {
		return fmt.Errorf("recreating source parent: %w", err)
	}
	if err := os.Rename(result.Destination, result.Source); err != nil {
		return fmt.Errorf("rollback rename: %w", err)
	}
	removeParents(result.createdParents)
	return nil
}

// Preflight refuses a movement when a sibling project's structured manifest
// reference resolves below the candidate moved root.
func Preflight(ctx config.Context, candidateMovedRoot string) error {
	candidate, err := config.CanonicalPath(candidateMovedRoot)
	if err != nil {
		return err
	}
	if ctx.WorkRoot == "" {
		return fmt.Errorf("movement preflight requires a work root")
	}
	entries, err := list.Scan(ctx.WorkRoot)
	if err != nil {
		return fmt.Errorf("scanning sibling folios: %w", err)
	}
	for _, entry := range entries {
		manifest := filepath.Join(ctx.WorkRoot, entry.Section, entry.Path, "folio.yml")
		projectRoot := filepath.Dir(manifest)
		if pathWithin(projectRoot, candidate) {
			continue
		}
		if entry.Error != "" {
			return fmt.Errorf("dependent manifest %s is invalid: %s", entry.Path, entry.Error)
		}
		folio, loadErr := config.Load(manifest)
		if loadErr != nil {
			return fmt.Errorf("dependent manifest %s: %w", entry.Path, loadErr)
		}
		siblingCtx, contextErr := ctx.ForProject(manifest)
		if contextErr != nil {
			return contextErr
		}
		for _, reference := range manifestReferences(folio) {
			resolved, resolveErr := siblingCtx.ResolvePath(reference)
			if resolveErr == nil && pathWithin(candidate, resolved) {
				return fmt.Errorf("dependent %s references moved path %q", entry.Path, reference)
			}
		}
	}
	return nil
}

func manifestReferences(folio *config.Folio) []string {
	var references []string
	addSource := func(source config.Source) {
		if source.Path != "" {
			references = append(references, source.Path)
		}
	}
	for _, source := range folio.Sources {
		addSource(source)
	}
	for _, target := range folio.Targets {
		for _, source := range target.Sources {
			addSource(source)
		}
		if target.Batch != nil {
			for _, item := range target.Batch.Items {
				if item.Source != "" {
					references = append(references, item.Source)
				}
			}
		}
		if target.Forest != nil && target.Forest.Root != "" {
			references = append(references, target.Forest.Root)
		}
	}
	return references
}

func pathWithin(root, path string) bool {
	root = filepath.Clean(root)
	path = filepath.Clean(path)
	return path == root || strings.HasPrefix(path, root+string(filepath.Separator))
}

func missingParents(path, stop string) []string {
	var parents []string
	for current := path; current != stop && current != filepath.Dir(current); current = filepath.Dir(current) {
		if _, err := os.Stat(current); err == nil {
			break
		}
		parents = append(parents, current)
	}
	return parents
}

func removeParents(parents []string) {
	for _, parent := range parents {
		if err := os.Remove(parent); err != nil {
			break
		}
	}
}

func sectionRoot(path string) string {
	for current := filepath.Clean(path); current != filepath.Dir(current); current = filepath.Dir(current) {
		base := filepath.Base(current)
		if base == "active" || base == "archive" {
			return current
		}
	}
	return ""
}

// validateRelPath checks that a relative path is non-empty and meaningful.
func validateRelPath(relPath string) error {
	cleaned := filepath.Clean(relPath)
	if cleaned == "" || cleaned == "." || filepath.IsAbs(relPath) || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return fmt.Errorf("invalid path: %q", relPath)
	}
	return nil
}

// PruneEmptyAncestors removes empty directories from dir up to (but not
// including) stopAt. Exported for archive operations outside this package
// (e.g. work-track archive) that need the same emptied-directory cleanup.
func PruneEmptyAncestors(dir, stopAt string) {
	pruneEmptyAncestors(dir, stopAt)
}

func pruneEmptyAncestors(dir, stopAt string) {
	for dir != stopAt && dir != "." && dir != "/" {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			return
		}
		if err := os.Remove(dir); err != nil {
			return
		}
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
