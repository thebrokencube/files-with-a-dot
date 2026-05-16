package home

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/taxonomy"
)

const defaultHome = ".folio"

// Dir resolves the FOLIO_HOME directory. Uses FOLIO_HOME env var if set,
// otherwise defaults to ~/.folio/.
func Dir() (string, error) {
	if env := os.Getenv("FOLIO_HOME"); env != "" {
		return env, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, defaultHome), nil
}

// Init scaffolds the FOLIO_HOME directory structure. Idempotent — only creates
// directories and files that don't already exist.
func Init(dir string) error {
	dirs := []string{
		filepath.Join(dir, "active"),
		filepath.Join(dir, "archive"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("mkdir %s: %w", d, err)
		}
	}

	// Create CLAUDE.md if missing
	claudePath := filepath.Join(dir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); os.IsNotExist(err) {
		if err := os.WriteFile(claudePath, []byte(TemplateClaude), 0644); err != nil {
			return fmt.Errorf("write CLAUDE.md: %w", err)
		}
	}

	// Create README.md if missing
	readmePath := filepath.Join(dir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		if err := os.WriteFile(readmePath, []byte(TemplateReadme), 0644); err != nil {
			return fmt.Errorf("write README.md: %w", err)
		}
	}

	// Create .gitignore if missing
	gitignorePath := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		if err := os.WriteFile(gitignorePath, []byte(TemplateGitignore), 0644); err != nil {
			return fmt.Errorf("write .gitignore: %w", err)
		}
	}

	// Initialize git repo if not already one
	if _, err := os.Stat(filepath.Join(dir, ".git")); os.IsNotExist(err) {
		cmd := exec.Command("git", "init")
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("git init: %w", err)
		}
	}

	return nil
}

// Validate checks the structural integrity of a FOLIO_HOME directory.
// Returns a list of errors found.
func Validate(dir string) []string {
	var errs []string

	// Check top-level dirs exist
	for _, name := range []string{"active", "archive"} {
		p := filepath.Join(dir, name)
		fi, err := os.Stat(p)
		if err != nil {
			errs = append(errs, fmt.Sprintf("missing directory: %s", name))
			continue
		}
		if !fi.IsDir() {
			errs = append(errs, fmt.Sprintf("not a directory: %s", name))
		}
	}

	// Check that leaf directories in active/ have folio.yml
	activeDir := filepath.Join(dir, "active")
	if fi, err := os.Stat(activeDir); err == nil && fi.IsDir() {
		errs = append(errs, validateLeaves(activeDir, "active", false)...)
	}

	// Check that leaf directories in archive/ have folio.yml and date prefix
	archiveDir := filepath.Join(dir, "archive")
	if fi, err := os.Stat(archiveDir); err == nil && fi.IsDir() {
		errs = append(errs, validateLeaves(archiveDir, "archive", true)...)
	}

	// Check vault structure
	errs = append(errs, validateVault(dir)...)

	return errs
}

// validateVault checks the structural integrity of the vault/ directory.
// Enforces: only recognized label subdirectories, only .md files with
// YYYY-MM-DD- prefix, no nested subdirectories, no root-level files.
func validateVault(dir string) []string {
	var errs []string
	vaultDir := filepath.Join(dir, "vault")

	fi, err := os.Stat(vaultDir)
	if os.IsNotExist(err) {
		return nil // no vault is fine
	}
	if !fi.IsDir() {
		return []string{"vault: not a directory"}
	}

	entries, err := os.ReadDir(vaultDir)
	if err != nil {
		return []string{fmt.Sprintf("vault: %s", err)}
	}

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue // skip dotfiles (.DS_Store, .obsidian, etc.)
		}

		if !entry.IsDir() {
			errs = append(errs, fmt.Sprintf("vault: file at root level: %s", name))
			continue
		}

		if !taxonomy.ReferenceLabels[name] {
			errs = append(errs, fmt.Sprintf("vault: unrecognized label directory: %s", name))
			continue
		}

		// Check contents of label directory
		labelDir := filepath.Join(vaultDir, name)
		files, err := os.ReadDir(labelDir)
		if err != nil {
			continue
		}
		for _, f := range files {
			fname := f.Name()
			if strings.HasPrefix(fname, ".") {
				continue
			}
			if f.IsDir() {
				errs = append(errs, fmt.Sprintf("vault/%s/%s: unexpected subdirectory", name, fname))
				continue
			}
			if !strings.HasSuffix(fname, ".md") {
				errs = append(errs, fmt.Sprintf("vault/%s/%s: non-markdown file", name, fname))
				continue
			}
			if !hasDatePrefix(fname) {
				errs = append(errs, fmt.Sprintf("vault/%s/%s: missing YYYY-MM-DD- prefix", name, fname))
			}
		}
	}

	return errs
}

// validateLeaves walks a section directory and checks that every folio directory
// (identified by containing a folio.yml) meets structural requirements.
// Directories containing folio.yml are treated as project roots — their children
// are internal structure, not separate folios. If requireDatePrefix is true,
// folio directory names must start with YYYY-MM-DD-.
func validateLeaves(root, section string, requireDatePrefix bool) []string {
	var errs []string

	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() || path == root {
			return nil
		}

		folioPath := filepath.Join(path, "folio.yml")
		hasFolio := false
		if _, err := os.Stat(folioPath); err == nil {
			hasFolio = true
		}

		if hasFolio {
			// This is a folio project root — validate it, then skip children
			if requireDatePrefix {
				base := filepath.Base(path)
				if !hasDatePrefix(base) {
					rel, _ := filepath.Rel(root, path)
					errs = append(errs, fmt.Sprintf("%s/%s: leaf missing YYYY-MM-DD- prefix", section, rel))
				}
			}
			return fs.SkipDir
		}

		// No folio.yml — if this is a leaf, it's an orphan
		if isLeaf(path) {
			rel, _ := filepath.Rel(root, path)
			errs = append(errs, fmt.Sprintf("%s/%s: missing folio.yml", section, rel))
		}

		return nil
	})

	return errs
}

// isLeaf returns true if the directory has no subdirectories.
func isLeaf(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return true
	}
	for _, e := range entries {
		if e.IsDir() {
			return false
		}
	}
	return true
}

// hasDatePrefix checks if a name starts with YYYY-MM-DD-.
func hasDatePrefix(name string) bool {
	if len(name) < 11 {
		return false
	}
	// Check pattern: DDDD-DD-DD-
	for i, c := range name[:10] {
		switch {
		case i == 4 || i == 7:
			if c != '-' {
				return false
			}
		default:
			if c < '0' || c > '9' {
				return false
			}
		}
	}
	return name[10] == '-'
}
