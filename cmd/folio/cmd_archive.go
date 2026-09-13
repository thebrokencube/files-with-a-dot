package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/move"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/repo"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/validate"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runArchive(folioPath string, dryRun, noPush bool, trackName string) int {
	pal := dendrik.NewPalette(true)

	ctx, err := resolveContext(folioPath, contextMutation)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}
	folioPath = ctx.FolioPath
	folioDir := ctx.FolioDir()
	activeDir := filepath.Join(folioDir, "work", "active", trackName)
	archiveDir := filepath.Join(folioDir, "work", "archive", trackName)

	// Verify active dir exists
	info, err := os.Stat(activeDir)
	if err != nil || !info.IsDir() {
		fmt.Fprintln(os.Stderr, pal.Errf("track not found: %s", activeDir))
		return dendrik.ExitUserError
	}

	// Verify archive dir doesn't exist
	if _, err := os.Stat(archiveDir); err == nil {
		fmt.Fprintln(os.Stderr, pal.Errf("archive already exists: %s", archiveDir))
		return dendrik.ExitUserError
	}

	// Read raw folio.yml bytes and establish the baseline before movement.
	raw, err := os.ReadFile(folioPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("reading folio.yml: %s", err))
		return dendrik.ExitUserError
	}
	baselineFolio, err := config.Parse(raw)
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}
	before := validate.Validate(baselineFolio, ctx, validate.Mutation)

	oldPrefix := filepath.Join("work", "active", trackName)
	newPrefix := filepath.Join("work", "archive", trackName)
	if err := move.Preflight(ctx, activeDir); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("archive refused: %s", err))
		return dendrik.ExitUserError
	}
	rewritten, count, err := rewritePaths(raw, oldPrefix, newPrefix)
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("archive refused: %s", err))
		return dendrik.ExitUserError
	}

	roots, rootsErr := discoverContextRoots()
	if rootsErr != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", rootsErr))
		return dendrik.ExitUserError
	}
	promisedPush := !noPush && (ctx.Store.Name != "" ||
		(ctx.Store.Name == "" && roots.workRoot != "" && configPathWithin(roots.workRoot, folioDir)))
	var scopedRel string
	if promisedPush {
		if !repo.IsJJ(ctx.WorkRoot) {
			fmt.Fprintln(os.Stderr, pal.Errf("archive requires a jj work root for its promised scoped push: %s", ctx.WorkRoot))
			return dendrik.ExitUserError
		}
		scopedRel, err = filepath.Rel(ctx.WorkRoot, folioDir)
		if err != nil || scopedRel == ".." || strings.HasPrefix(scopedRel, ".."+string(filepath.Separator)) {
			fmt.Fprintln(os.Stderr, pal.Errf("folio directory %s is outside selected work root %s — cannot promise auto-push", folioDir, ctx.WorkRoot))
			return dendrik.ExitUserError
		}
	}

	if dryRun {
		fmt.Printf("Would move: %s → %s\n", activeDir, archiveDir)
		fmt.Printf("Would rewrite %d path reference(s) in folio.yml\n", count)
		return dendrik.ExitOK
	}

	// Record only archive parents this command creates so rollback preserves
	// pre-existing empty directories.
	createdArchiveParents := missingParentDirs(filepath.Dir(archiveDir), filepath.Join(folioDir, "work"))
	if err := os.MkdirAll(filepath.Dir(archiveDir), 0755); err != nil {
		return dendrik.ExitUserError
	}

	// Move directory
	if err := os.Rename(activeDir, archiveDir); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("moving directory: %s", err))
		return dendrik.ExitUserError
	}

	// Validate rewritten config before replacing the manifest.
	parsed, err := config.Parse(rewritten)
	if err != nil {
		rollbackDirMove(activeDir, archiveDir, createdArchiveParents)
		fmt.Fprintln(os.Stderr, pal.Errf("rewritten folio.yml failed to parse: %s", err))
		return dendrik.ExitUserError
	}
	after := validate.Validate(parsed, ctx, validate.Mutation)
	delta := validate.Delta(before, after, map[string]string{oldPrefix: newPrefix})
	if !delta.Valid {
		rollbackDirMove(activeDir, archiveDir, createdArchiveParents)
		fmt.Fprintln(os.Stderr, pal.Errf("validation failed — archive rolled back:"))
		for _, validationErr := range delta.Errors {
			fmt.Fprintf(os.Stderr, "  - %s\n", validationErr)
		}
		return dendrik.ExitUserError
	}

	// Atomic write: tmp file then rename
	tmpPath := folioPath + ".tmp"
	if err := os.WriteFile(tmpPath, rewritten, 0644); err != nil {
		rollbackDirMove(activeDir, archiveDir, createdArchiveParents)
		fmt.Fprintln(os.Stderr, pal.Errf("writing temp file: %s", err))
		return dendrik.ExitUserError
	}
	if err := os.Rename(tmpPath, folioPath); err != nil {
		rollbackDirMove(activeDir, archiveDir, createdArchiveParents)
		os.Remove(tmpPath) // best-effort cleanup
		fmt.Fprintln(os.Stderr, pal.Errf("replacing folio.yml: %s", err))
		return dendrik.ExitUserError
	}

	// Prune the emptied work/active dir so archiving the last active track
	// doesn't leave a hollow skeleton behind. Stop at work/ to preserve the
	// project's directory structure.
	move.PruneEmptyAncestors(filepath.Dir(activeDir), filepath.Join(folioDir, "work"))

	fmt.Printf("Archived: %s → %s\n", activeDir, archiveDir)
	fmt.Printf("Rewrote %d path reference(s) in folio.yml\n", count)

	// Auto-commit unless --no-push. Explicit projects outside the selected
	// content root have no promised push and complete without one.
	if promisedPush {
		msg := fmt.Sprintf("chore(archive): archive %s", trackName)
		if pushErr := repo.PushScoped(ctx.WorkRoot, msg, []string{scopedRel}); pushErr != nil {
			fmt.Fprintln(os.Stderr, pal.Errf("auto-commit: %s", pushErr))
			return dendrik.ExitUserError
		}
		fmt.Println("Committed and pushed.")
	}

	return dendrik.ExitOK
}

func rewritePaths(raw []byte, oldPrefix, newPrefix string) ([]byte, int, error) {
	oldToken := []byte(oldPrefix)
	newToken := []byte(newPrefix)
	var rewritten bytes.Buffer
	cursor := 0
	count := 0
	for cursor < len(raw) {
		offset := bytes.Index(raw[cursor:], oldToken)
		if offset < 0 {
			rewritten.Write(raw[cursor:])
			break
		}
		start := cursor + offset
		end := start + len(oldToken)
		if !safeTokenBoundary(raw, start, end) {
			return nil, 0, fmt.Errorf("cannot safely rewrite path reference %q", oldPrefix)
		}
		rewritten.Write(raw[cursor:start])
		rewritten.Write(newToken)
		cursor = end
		count++
	}
	return rewritten.Bytes(), count, nil
}

func safeTokenBoundary(raw []byte, start, end int) bool {
	if start > 0 && !isBoundaryByte(raw[start-1]) {
		return false
	}
	if end < len(raw) && raw[end] != '/' && !isBoundaryByte(raw[end]) {
		return false
	}
	return true
}

func isBoundaryByte(value byte) bool {
	return value <= ' ' || strings.ContainsRune(`"'([{,:;.)]}`, rune(value))
}

func rollbackDirMove(activeDir, archiveDir string, createdParents []string) {
	if err := os.MkdirAll(filepath.Dir(activeDir), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: rollback failed recreating parent: %s\n", err)
		return
	}
	if err := os.Rename(archiveDir, activeDir); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: rollback failed: %s\n", err)
		return
	}
	for _, parent := range createdParents {
		if err := os.Remove(parent); err != nil {
			break
		}
	}
}
