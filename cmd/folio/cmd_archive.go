package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/home"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/repo"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/validate"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runArchive(args []string) int {
	fs := dendrik.NewFlagSet("archive")
	folioPath := fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
	dryRun := fs.Bool('n', "dry-run", "Print what would happen, no side effects")
	noPush := fs.BoolLong("no-push", "Skip auto-commit")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if len(fs.GetArgs()) < 1 {
		fmt.Fprintln(os.Stderr, output.Errf("usage: folio archive <track-name> [--folio PATH] [--dry-run] [--no-push]"))
		return dendrik.ExitUserError
	}
	trackName := fs.GetArgs()[0]

	if !resolveOrDie(folioPath) {
		return dendrik.ExitUserError
	}

	folioDir := filepath.Dir(*folioPath)
	activeDir := filepath.Join(folioDir, "work", "active", trackName)
	archiveDir := filepath.Join(folioDir, "work", "archive", trackName)

	// Verify active dir exists
	info, err := os.Stat(activeDir)
	if err != nil || !info.IsDir() {
		fmt.Fprintln(os.Stderr, output.Errf("track not found: %s", activeDir))
		return dendrik.ExitUserError
	}

	// Verify archive dir doesn't exist
	if _, err := os.Stat(archiveDir); err == nil {
		fmt.Fprintln(os.Stderr, output.Errf("archive already exists: %s", archiveDir))
		return dendrik.ExitUserError
	}

	// Read raw folio.yml bytes
	raw, err := os.ReadFile(*folioPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("reading folio.yml: %s", err))
		return dendrik.ExitUserError
	}

	oldPrefix := filepath.Join("work", "active", trackName)
	newPrefix := filepath.Join("work", "archive", trackName)

	rewritten, count := rewritePaths(raw, oldPrefix, newPrefix)

	if *dryRun {
		fmt.Printf("Would move: %s → %s\n", activeDir, archiveDir)
		fmt.Printf("Would rewrite %d path reference(s) in folio.yml\n", count)
		return 0
	}

	// Ensure archive parent directory exists
	if err := os.MkdirAll(filepath.Dir(archiveDir), 0755); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("creating archive directory: %s", err))
		return dendrik.ExitUserError
	}

	// Move directory
	if err := os.Rename(activeDir, archiveDir); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("moving directory: %s", err))
		return dendrik.ExitUserError
	}

	// Validate rewritten config
	parsed, err := config.Parse(rewritten)
	if err != nil {
		rollbackDirMove(activeDir, archiveDir)
		fmt.Fprintln(os.Stderr, output.Errf("rewritten folio.yml failed to parse: %s", err))
		return dendrik.ExitUserError
	}
	result := validate.Validate(parsed, folioDir)
	if !result.Valid {
		rollbackDirMove(activeDir, archiveDir)
		fmt.Fprintln(os.Stderr, output.Errf("rewritten folio.yml failed validation: %s", strings.Join(result.Errors, "; ")))
		return dendrik.ExitUserError
	}

	// Atomic write: tmp file then rename
	tmpPath := *folioPath + ".tmp"
	if err := os.WriteFile(tmpPath, rewritten, 0644); err != nil {
		rollbackDirMove(activeDir, archiveDir)
		fmt.Fprintln(os.Stderr, output.Errf("writing temp file: %s", err))
		return dendrik.ExitUserError
	}
	if err := os.Rename(tmpPath, *folioPath); err != nil {
		rollbackDirMove(activeDir, archiveDir)
		os.Remove(tmpPath) // best-effort cleanup
		fmt.Fprintln(os.Stderr, output.Errf("replacing folio.yml: %s", err))
		return dendrik.ExitUserError
	}

	fmt.Printf("Archived: %s → %s\n", activeDir, archiveDir)
	fmt.Printf("Rewrote %d path reference(s) in folio.yml\n", count)

	// Auto-commit unless --no-push
	if !*noPush {
		homeDir, err := home.Dir()
		if err != nil {
			fmt.Fprintln(os.Stderr, output.Errf("resolving FOLIO_HOME: %s", err))
			return dendrik.ExitUserError
		}
		rel, err := filepath.Rel(homeDir, folioDir)
		if err != nil || strings.HasPrefix(rel, "..") {
			fmt.Fprintln(os.Stderr, output.Errf("folio directory %s is outside FOLIO_HOME %s — skipping auto-push", folioDir, homeDir))
			return 0
		}
		msg := fmt.Sprintf("chore(archive): archive %s", trackName)
		if pushErr := repo.PushScoped(homeDir, msg, []string{rel}); pushErr != nil {
			fmt.Fprintln(os.Stderr, output.Errf("auto-commit: %s", pushErr))
			return dendrik.ExitUserError
		}
		fmt.Println("Committed and pushed.")
	}

	return 0
}

func rewritePaths(raw []byte, oldPrefix, newPrefix string) ([]byte, int) {
	oldToken := []byte(oldPrefix + "/")
	newToken := []byte(newPrefix + "/")
	count := bytes.Count(raw, oldToken)
	rewritten := bytes.ReplaceAll(raw, oldToken, newToken)
	return rewritten, count
}

func rollbackDirMove(activeDir, archiveDir string) {
	if err := os.Rename(archiveDir, activeDir); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: rollback failed: %s\n", err)
	}
}
