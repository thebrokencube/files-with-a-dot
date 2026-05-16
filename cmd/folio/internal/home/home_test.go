package home

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDir_EnvVar(t *testing.T) {
	t.Setenv("FOLIO_HOME", "/tmp/test-folio")
	dir, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	if dir != "/tmp/test-folio" {
		t.Errorf("expected /tmp/test-folio, got %s", dir)
	}
}

func TestDir_Default(t *testing.T) {
	t.Setenv("FOLIO_HOME", "")
	dir, err := Dir()
	if err != nil {
		t.Fatal(err)
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".folio")
	if dir != expected {
		t.Errorf("expected %s, got %s", expected, dir)
	}
}

func TestInit_CreatesStructure(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "folio-home")

	if err := Init(dir); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"active", "archive", "CLAUDE.md", "README.md"} {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			t.Errorf("missing: %s", name)
		}
	}
}

func TestInit_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "folio-home")

	Init(dir)

	// Write custom CLAUDE.md
	custom := filepath.Join(dir, "CLAUDE.md")
	os.WriteFile(custom, []byte("custom content"), 0644)

	// Init again should not overwrite
	Init(dir)

	data, _ := os.ReadFile(custom)
	if string(data) != "custom content" {
		t.Error("Init overwrote existing CLAUDE.md")
	}
}

func TestValidate_Clean(t *testing.T) {
	dir := setupTestHome(t,
		"active/project-a/folio.yml",
		"archive/2026-01-15-project-b/folio.yml",
	)

	errs := Validate(dir)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidate_MissingFolioYml(t *testing.T) {
	dir := setupTestHome(t,
		"active/project-a/README.md", // leaf without folio.yml
	)

	errs := Validate(dir)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0] != "active/project-a: missing folio.yml" {
		t.Errorf("unexpected error: %s", errs[0])
	}
}

func TestValidate_ArchiveMissingDatePrefix(t *testing.T) {
	dir := setupTestHome(t,
		"archive/no-date-prefix/folio.yml",
	)

	errs := Validate(dir)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0] != "archive/no-date-prefix: leaf missing YYYY-MM-DD- prefix" {
		t.Errorf("unexpected error: %s", errs[0])
	}
}

func TestValidate_NestedActive(t *testing.T) {
	dir := setupTestHome(t,
		"active/ben/state-retirement/folio.yml",
		"active/ben/pb-on-call/folio.yml",
		"active/career-tracking/folio.yml",
	)

	errs := Validate(dir)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidate_NestedArchive(t *testing.T) {
	dir := setupTestHome(t,
		"archive/ben/pb-on-call/2026-02-20-ghost-policies/folio.yml",
		"archive/ben/pb-on-call/2026-02-20-stride-secrets/folio.yml",
	)

	errs := Validate(dir)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestHasDatePrefix(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"2026-02-20-project", true},
		{"2024-01-01-x", true},
		{"project", false},
		{"2026-02-20project", false},
		{"202-02-20-project", false},
		{"2026-2-20-project", false},
	}
	for _, tt := range tests {
		if got := hasDatePrefix(tt.name); got != tt.want {
			t.Errorf("hasDatePrefix(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestValidate_VaultClean(t *testing.T) {
	dir := setupTestHome(t,
		"active/project-a/folio.yml",
	)
	setupVault(t, dir,
		"research/2026-03-01-topic.md",
		"guide/2026-05-14-setup.md",
	)

	errs := Validate(dir)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidate_VaultMissingDatePrefix(t *testing.T) {
	dir := setupTestHome(t)
	setupVault(t, dir,
		"guide/colima-recovery.md",
	)

	errs := Validate(dir)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0] != "vault/guide/colima-recovery.md: missing YYYY-MM-DD- prefix" {
		t.Errorf("unexpected error: %s", errs[0])
	}
}

func TestValidate_VaultRootFile(t *testing.T) {
	dir := setupTestHome(t)
	setupVault(t, dir,
		"orphan.md",
	)

	errs := Validate(dir)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0] != "vault: file at root level: orphan.md" {
		t.Errorf("unexpected error: %s", errs[0])
	}
}

func TestValidate_VaultUnrecognizedLabel(t *testing.T) {
	dir := setupTestHome(t)
	setupVault(t, dir,
		"notes/2026-01-01-something.md",
	)

	errs := Validate(dir)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0] != "vault: unrecognized label directory: notes" {
		t.Errorf("unexpected error: %s", errs[0])
	}
}

func TestValidate_VaultNestedSubdir(t *testing.T) {
	dir := setupTestHome(t)
	setupVault(t, dir,
		"research/nested/2026-01-01-deep.md",
	)

	errs := Validate(dir)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0] != "vault/research/nested: unexpected subdirectory" {
		t.Errorf("unexpected error: %s", errs[0])
	}
}

func TestValidate_VaultNonMarkdown(t *testing.T) {
	dir := setupTestHome(t)
	setupVault(t, dir,
		"domain/2026-01-01-notes.txt",
	)

	errs := Validate(dir)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if errs[0] != "vault/domain/2026-01-01-notes.txt: non-markdown file" {
		t.Errorf("unexpected error: %s", errs[0])
	}
}

func TestValidate_VaultEmpty(t *testing.T) {
	dir := setupTestHome(t)
	os.MkdirAll(filepath.Join(dir, "vault"), 0755)

	errs := Validate(dir)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidate_VaultMissing(t *testing.T) {
	dir := setupTestHome(t)
	// no vault/ at all

	errs := Validate(dir)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestValidate_VaultDotfilesIgnored(t *testing.T) {
	dir := setupTestHome(t)
	setupVault(t, dir,
		"research/2026-01-01-valid.md",
	)
	// Add a .DS_Store and .obsidian dir that should be ignored
	os.WriteFile(filepath.Join(dir, "vault", ".DS_Store"), []byte{}, 0644)
	os.MkdirAll(filepath.Join(dir, "vault", ".obsidian"), 0755)
	os.WriteFile(filepath.Join(dir, "vault", "research", ".DS_Store"), []byte{}, 0644)

	errs := Validate(dir)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

// setupVault creates vault files in a test FOLIO_HOME.
func setupVault(t *testing.T, dir string, files ...string) {
	t.Helper()
	for _, f := range files {
		p := filepath.Join(dir, "vault", f)
		os.MkdirAll(filepath.Dir(p), 0755)
		os.WriteFile(p, []byte("# test\n"), 0644)
	}
}

// setupTestHome creates a temporary FOLIO_HOME with the given file paths.
func setupTestHome(t *testing.T, files ...string) string {
	t.Helper()
	dir := t.TempDir()

	// Create active/ and archive/
	os.MkdirAll(filepath.Join(dir, "active"), 0755)
	os.MkdirAll(filepath.Join(dir, "archive"), 0755)

	for _, f := range files {
		p := filepath.Join(dir, f)
		os.MkdirAll(filepath.Dir(p), 0755)
		os.WriteFile(p, []byte("schema: 1\nproject: test\n"), 0644)
	}

	return dir
}
