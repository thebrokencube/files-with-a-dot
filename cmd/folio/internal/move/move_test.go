package move

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestArchive_Basic(t *testing.T) {
	dir := t.TempDir()
	setupDirs(t, dir,
		"active/ben/project-a/folio.yml",
	)

	err := Archive(dir, "ben/project-a")
	if err != nil {
		t.Fatal(err)
	}

	// Source should be gone
	if _, err := os.Stat(filepath.Join(dir, "active/ben/project-a")); !os.IsNotExist(err) {
		t.Error("source still exists after archive")
	}

	// Dest should exist with date prefix
	entries, _ := os.ReadDir(filepath.Join(dir, "archive/ben"))
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry in archive/ben, got %d", len(entries))
	}
	name := entries[0].Name()
	if !strings.HasSuffix(name, "-project-a") {
		t.Errorf("expected *-project-a, got %s", name)
	}
	// Check date prefix format
	if len(name) < 11 || name[4] != '-' || name[7] != '-' || name[10] != '-' {
		t.Errorf("bad date prefix format: %s", name)
	}
}

func TestArchive_NotFound(t *testing.T) {
	dir := t.TempDir()
	setupDirs(t, dir)

	err := Archive(dir, "nonexistent")
	if err == nil {
		t.Error("expected error for missing path")
	}
}

func TestArchive_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	setupDirs(t, dir,
		"active/project/folio.yml",
	)

	// First archive
	Archive(dir, "project")

	// Re-create source
	setupDirs(t, dir,
		"active/project/folio.yml",
	)

	// Second archive should fail (same date)
	err := Archive(dir, "project")
	if err == nil {
		t.Error("expected error for duplicate archive")
	}
}

func TestActivate_Basic(t *testing.T) {
	dir := t.TempDir()
	setupDirs(t, dir,
		"archive/ben/2026-02-20-project-a/folio.yml",
	)

	err := Activate(dir, "ben/2026-02-20-project-a")
	if err != nil {
		t.Fatal(err)
	}

	// Source should be gone
	if _, err := os.Stat(filepath.Join(dir, "archive/ben/2026-02-20-project-a")); !os.IsNotExist(err) {
		t.Error("archive source still exists")
	}

	// Dest should exist without date prefix
	dest := filepath.Join(dir, "active/ben/project-a")
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		t.Error("active/ben/project-a not found")
	}

	// folio.yml should be there
	if _, err := os.Stat(filepath.Join(dest, "folio.yml")); os.IsNotExist(err) {
		t.Error("folio.yml not found in activated project")
	}
}

func TestActivate_NotFound(t *testing.T) {
	dir := t.TempDir()
	setupDirs(t, dir)

	err := Activate(dir, "nonexistent")
	if err == nil {
		t.Error("expected error for missing path")
	}
}

func TestActivate_NoDatePrefix(t *testing.T) {
	dir := t.TempDir()
	setupDirs(t, dir,
		"archive/no-date-project/folio.yml",
	)

	// Should still work — just moves without stripping
	err := Activate(dir, "no-date-project")
	if err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join(dir, "active/no-date-project")
	if _, err := os.Stat(dest); os.IsNotExist(err) {
		t.Error("expected active/no-date-project to exist")
	}
}

func TestStripDatePrefix(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"2026-02-20-project", "project"},
		{"2024-01-01-a", "a"},
		{"project", "project"},
		{"2026-02-20project", "2026-02-20project"},
		{"20-project", "20-project"},
	}
	for _, tt := range tests {
		if got := stripDatePrefix(tt.input); got != tt.want {
			t.Errorf("stripDatePrefix(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestArchiveActivateRoundTrip(t *testing.T) {
	dir := t.TempDir()
	setupDirs(t, dir,
		"active/ben/my-project/folio.yml",
		"active/ben/my-project/README.md",
	)

	// Archive
	if err := Archive(dir, "ben/my-project"); err != nil {
		t.Fatal(err)
	}

	// Find the archived name
	entries, _ := os.ReadDir(filepath.Join(dir, "archive/ben"))
	if len(entries) != 1 {
		t.Fatalf("expected 1 archived entry, got %d", len(entries))
	}
	archivedName := entries[0].Name()

	// Activate
	if err := Activate(dir, "ben/"+archivedName); err != nil {
		t.Fatal(err)
	}

	// Should be back in active/ben/my-project
	dest := filepath.Join(dir, "active/ben/my-project")
	if _, err := os.Stat(filepath.Join(dest, "folio.yml")); os.IsNotExist(err) {
		t.Error("folio.yml not restored")
	}
	if _, err := os.Stat(filepath.Join(dest, "README.md")); os.IsNotExist(err) {
		t.Error("README.md not restored")
	}
}

func TestArchive_PrunesEmptyAncestors(t *testing.T) {
	dir := t.TempDir()
	setupDirs(t, dir,
		"active/ret/deep/project-a/folio.yml",
	)

	err := Archive(dir, "ret/deep/project-a")
	if err != nil {
		t.Fatal(err)
	}

	// active/ret/deep/ and active/ret/ should be pruned (empty after move)
	if _, err := os.Stat(filepath.Join(dir, "active/ret/deep")); !os.IsNotExist(err) {
		t.Error("active/ret/deep should have been pruned")
	}
	if _, err := os.Stat(filepath.Join(dir, "active/ret")); !os.IsNotExist(err) {
		t.Error("active/ret should have been pruned")
	}
	// active/ itself must remain
	if _, err := os.Stat(filepath.Join(dir, "active")); os.IsNotExist(err) {
		t.Error("active/ should NOT have been pruned")
	}
}

func TestArchive_PreservesSiblings(t *testing.T) {
	dir := t.TempDir()
	setupDirs(t, dir,
		"active/ret/project-a/folio.yml",
		"active/ret/project-b/folio.yml",
	)

	err := Archive(dir, "ret/project-a")
	if err != nil {
		t.Fatal(err)
	}

	// active/ret/ should still exist because project-b is there
	if _, err := os.Stat(filepath.Join(dir, "active/ret")); os.IsNotExist(err) {
		t.Error("active/ret should still exist (has sibling)")
	}
}

func TestActivate_PrunesEmptyAncestors(t *testing.T) {
	dir := t.TempDir()
	setupDirs(t, dir,
		"archive/ret/deep/2026-01-01-project-a/folio.yml",
	)

	err := Activate(dir, "ret/deep/2026-01-01-project-a")
	if err != nil {
		t.Fatal(err)
	}

	// archive/ret/deep/ and archive/ret/ should be pruned
	if _, err := os.Stat(filepath.Join(dir, "archive/ret/deep")); !os.IsNotExist(err) {
		t.Error("archive/ret/deep should have been pruned")
	}
	if _, err := os.Stat(filepath.Join(dir, "archive/ret")); !os.IsNotExist(err) {
		t.Error("archive/ret should have been pruned")
	}
	// archive/ itself must remain
	if _, err := os.Stat(filepath.Join(dir, "archive")); os.IsNotExist(err) {
		t.Error("archive/ should NOT have been pruned")
	}
}

func setupDirs(t *testing.T, dir string, files ...string) {
	t.Helper()
	os.MkdirAll(filepath.Join(dir, "active"), 0755)
	os.MkdirAll(filepath.Join(dir, "archive"), 0755)
	for _, f := range files {
		p := filepath.Join(dir, f)
		os.MkdirAll(filepath.Dir(p), 0755)
		os.WriteFile(p, []byte("test"), 0644)
	}
}
