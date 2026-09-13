package move

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

func TestArchive_Basic(t *testing.T) {
	dir := t.TempDir()
	setupDirs(t, dir,
		"active/ben/project-a/folio.yml",
	)

	_, err := Archive(dir, "ben/project-a")
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

	_, err := Archive(dir, "nonexistent")
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
	_, _ = Archive(dir, "project")

	// Re-create source
	setupDirs(t, dir,
		"active/project/folio.yml",
	)

	// Second archive should fail (same date)
	_, err := Archive(dir, "project")
	if err == nil {
		t.Error("expected error for duplicate archive")
	}
}

func TestActivate_Basic(t *testing.T) {
	dir := t.TempDir()
	setupDirs(t, dir,
		"archive/ben/2026-02-20-project-a/folio.yml",
	)

	_, err := Activate(dir, "ben/2026-02-20-project-a")
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

	_, err := Activate(dir, "nonexistent")
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
	_, err := Activate(dir, "no-date-project")
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
	if _, err := Archive(dir, "ben/my-project"); err != nil {
		t.Fatal(err)
	}

	// Find the archived name
	entries, _ := os.ReadDir(filepath.Join(dir, "archive/ben"))
	if len(entries) != 1 {
		t.Fatalf("expected 1 archived entry, got %d", len(entries))
	}
	archivedName := entries[0].Name()

	// Activate
	if _, err := Activate(dir, "ben/"+archivedName); err != nil {
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

	_, err := Archive(dir, "ret/deep/project-a")
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

	_, err := Archive(dir, "ret/project-a")
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

	_, err := Activate(dir, "ret/deep/2026-01-01-project-a")
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
func TestPreflightRejectsStructuredDependentReferences(t *testing.T) {
	t.Run("relative path", func(t *testing.T) {
		root := t.TempDir()
		writeValidManifest(t, filepath.Join(root, "active", "moving"), `schema: 1
project: moving
`)
		writeValidManifest(t, filepath.Join(root, "active", "dependent"), `schema: 1
project: dependent
sources:
  - path: ../moving/README.md
`)
		writeFile(t, filepath.Join(root, "active", "moving", "README.md"), "moving")

		ctx := config.Context{WorkRoot: root}
		err := Preflight(ctx, filepath.Join(root, "active", "moving"))
		if err == nil || !strings.Contains(err.Error(), "dependent") {
			t.Fatalf("Preflight error = %v, want dependent reference refusal", err)
		}
	})

	t.Run("registered store path", func(t *testing.T) {
		root := t.TempDir()
		writeValidManifest(t, filepath.Join(root, "active", "moving"), `schema: 1
project: moving
`)
		writeValidManifest(t, filepath.Join(root, "active", "dependent"), `schema: 1
project: dependent
sources:
  - path: work:active/moving/README.md
`)
		writeFile(t, filepath.Join(root, "active", "moving", "README.md"), "moving")
		registry := &config.Registry{
			Stores: map[string]config.Store{
				"work": {Name: "work", Path: root, Kind: config.KindFolio},
			},
			Order: []string{"work"},
		}

		ctx := config.Context{WorkRoot: root, Registry: registry}
		err := Preflight(ctx, filepath.Join(root, "active", "moving"))
		if err == nil || !strings.Contains(err.Error(), "dependent") {
			t.Fatalf("Preflight error = %v, want dependent reference refusal", err)
		}
	})
}

func TestPreflightIgnoresObservationProse(t *testing.T) {
	root := t.TempDir()
	writeValidManifest(t, filepath.Join(root, "active", "moving"), `schema: 1
project: moving
`)
	writeValidManifest(t, filepath.Join(root, "active", "dependent"), `schema: 1
project: dependent
observations:
  - "See ../moving/README.md for context"
`)

	if err := Preflight(config.Context{WorkRoot: root}, filepath.Join(root, "active", "moving")); err != nil {
		t.Fatalf("Preflight returned an error for prose-only reference: %v", err)
	}
}

func TestArchiveRollbackRestoresExactPaths(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "active", "team", "deep", "project")
	manifest := []byte("schema: 1\nproject: project\n")
	writeValidManifest(t, project, string(manifest))
	writeFile(t, filepath.Join(project, "README.md"), "content")

	result, err := Archive(root, "team/deep/project")
	if err != nil {
		t.Fatal(err)
	}
	if err := Rollback(result); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(project, "folio.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(manifest) {
		t.Fatalf("manifest after rollback = %q, want %q", got, manifest)
	}
	if _, err := os.Stat(filepath.Join(project, "README.md")); err != nil {
		t.Fatalf("project content missing after rollback: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "archive", "team")); !os.IsNotExist(err) {
		t.Fatalf("archive parents remain after rollback: %v", err)
	}
}

func TestActivateRollbackRestoresExactPaths(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "archive", "team", "deep", "2026-01-01-project")
	manifest := []byte("schema: 1\nproject: project\n")
	writeValidManifest(t, project, string(manifest))
	writeFile(t, filepath.Join(project, "README.md"), "content")

	result, err := Activate(root, "team/deep/2026-01-01-project")
	if err != nil {
		t.Fatal(err)
	}
	if err := Rollback(result); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(filepath.Join(project, "folio.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(manifest) {
		t.Fatalf("manifest after rollback = %q, want %q", got, manifest)
	}
	if _, err := os.Stat(filepath.Join(project, "README.md")); err != nil {
		t.Fatalf("project content missing after rollback: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "active", "team")); !os.IsNotExist(err) {
		t.Fatalf("active parents remain after rollback: %v", err)
	}
}

func writeValidManifest(t *testing.T, project, content string) {
	t.Helper()
	if err := os.MkdirAll(project, 0755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(project, "folio.yml"), content)
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
