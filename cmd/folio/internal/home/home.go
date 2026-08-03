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

// SessionWorkspacePrefix names the jj workspaces `folio home workspace create`
// makes for KB sessions. They sit under /tmp on their own lifetime (reaped at
// -mtime +2), unlike code work areas, which are durable and ledgered under
// <umbrella>/.worktrees. Every caller that classifies a path by this prefix
// shares the constant so the two roots can't drift apart.
const SessionWorkspacePrefix = "folio-ws-"

// IsSessionWorkspace reports whether dir is a KB session workspace.
func IsSessionWorkspace(dir string) bool {
	return strings.HasPrefix(filepath.Base(dir), SessionWorkspacePrefix)
}

// lookJJ reports whether the jj binary is available on PATH. It is a package
// var so tests can force either VCS branch in Init deterministically.
var lookJJ = func() bool {
	_, err := exec.LookPath("jj")
	return err == nil
}

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
//
// Umbrella guard: a dir containing stores.yml is a multi-store umbrella — a
// plain directory that physically contains independent store repos, NOT a folio
// home. Initializing it (scaffolding active/archive, or VCS-init'ing it into a
// repo) would be wrong, so Init refuses. Stores are created by cloning into the
// umbrella, not by `folio home init`.
func Init(dir string) error {
	if _, err := os.Stat(filepath.Join(dir, "stores.yml")); err == nil {
		return fmt.Errorf("%s is a multi-store umbrella (has stores.yml) — it is a plain directory, not a folio home; do not init or VCS-init it", dir)
	}

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

	// Initialize VCS if not already one. Prefer jj (colocated with git) when jj
	// is available so `folio home workspace` works out of the box; fall back to
	// plain git otherwise. Colocation keeps a normal .git for remote and history.
	_, gitErr := os.Stat(filepath.Join(dir, ".git"))
	_, jjErr := os.Stat(filepath.Join(dir, ".jj"))
	if os.IsNotExist(gitErr) && os.IsNotExist(jjErr) {
		vcs := "git"
		var cmd *exec.Cmd
		if lookJJ() {
			vcs = "jj"
			cmd = exec.Command("jj", "git", "init", "--colocate")
		} else {
			cmd = exec.Command("git", "init")
		}
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s init: %w", vcs, err)
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

	// Validation operates on repo content: a directory is only a folio (or an
	// orphan) if it holds VCS-tracked files. Empty or untracked-only directories
	// (build cruft, leftover skeletons, .DS_Store) are not repo state and are
	// never flagged — they wouldn't be part of a commit anyway.
	tracked := trackedDirs(dir)

	// Check that leaf directories in active/ have folio.yml
	activeDir := filepath.Join(dir, "active")
	if fi, err := os.Stat(activeDir); err == nil && fi.IsDir() {
		errs = append(errs, validateLeaves(activeDir, "active", false, tracked)...)
	}

	// Check that leaf directories in archive/ have folio.yml and date prefix
	archiveDir := filepath.Join(dir, "archive")
	if fi, err := os.Stat(archiveDir); err == nil && fi.IsDir() {
		errs = append(errs, validateLeaves(archiveDir, "archive", true, tracked)...)
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
func validateLeaves(root, section string, requireDatePrefix bool, tracked map[string]bool) []string {
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

		// No folio.yml — a leaf holding tracked files is an orphan (real content
		// outside any folio). A leaf with no tracked content is not repo state.
		if isLeaf(path) && tracked[path] {
			rel, _ := filepath.Rel(root, path)
			errs = append(errs, fmt.Sprintf("%s/%s: missing folio.yml", section, rel))
		}

		return nil
	})

	return errs
}

// trackedDirs returns the set of directories (absolute paths) that contain at
// least one VCS-tracked file, including all ancestors up to root. Validation
// uses this to distinguish real repo content from filesystem cruft (empty
// directories, gitignored files like .DS_Store). Returns nil if the tree is
// not under version control, in which case no directory is treated as tracked.
func trackedDirs(root string) map[string]bool {
	files, ok := listTracked(root)
	if !ok {
		return nil
	}
	dirs := make(map[string]bool)
	for _, rel := range files {
		rel = strings.TrimSpace(rel)
		if rel == "" {
			continue
		}
		d := filepath.Dir(filepath.Join(root, rel))
		for {
			dirs[d] = true
			if d == root {
				break
			}
			parent := filepath.Dir(d)
			if parent == d {
				break
			}
			d = parent
		}
	}
	return dirs
}

// listTracked returns repo-relative paths of all VCS-tracked files under root.
// Prefers jj (authoritative for the working copy) when present, falling back to
// git. The bool is false when the tree is not under version control.
func listTracked(root string) ([]string, bool) {
	if _, err := os.Stat(filepath.Join(root, ".jj")); err == nil {
		cmd := exec.Command("jj", "--no-pager", "file", "list")
		cmd.Dir = root
		if out, err := cmd.Output(); err == nil {
			return strings.Split(strings.TrimSpace(string(out)), "\n"), true
		}
	}
	cmd := exec.Command("git", "ls-files")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return nil, false
	}
	return strings.Split(strings.TrimSpace(string(out)), "\n"), true
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
