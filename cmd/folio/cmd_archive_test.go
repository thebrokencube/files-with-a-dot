package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func TestRewritePaths(t *testing.T) {
	t.Run("basic replacement", func(t *testing.T) {
		raw := []byte("path: work/active/2026-01-01-track/README.md\n")
		got, count, err := rewritePaths(raw, "work/active/2026-01-01-track", "work/archive/2026-01-01-track")
		if err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Errorf("count = %d, want 1", count)
		}
		want := "path: work/archive/2026-01-01-track/README.md\n"
		if string(got) != want {
			t.Errorf("got %q, want %q", string(got), want)
		}
	})

	t.Run("no matches", func(t *testing.T) {
		raw := []byte("path: work/active/other-track/README.md\n")
		got, count, err := rewritePaths(raw, "work/active/2026-01-01-track", "work/archive/2026-01-01-track")
		if err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Errorf("count = %d, want 0", count)
		}
		if !bytes.Equal(got, raw) {
			t.Errorf("bytes changed when they shouldn't have")
		}
	})

	t.Run("multiple occurrences", func(t *testing.T) {
		raw := []byte("- path: work/active/2026-01-01-track/src.md\n- path: work/active/2026-01-01-track/out.md\n")
		got, count, err := rewritePaths(raw, "work/active/2026-01-01-track", "work/archive/2026-01-01-track")
		if err != nil {
			t.Fatal(err)
		}
		if count != 2 {
			t.Errorf("count = %d, want 2", count)
		}
		if strings.Contains(string(got), "work/active/2026-01-01-track/") {
			t.Error("old prefix still present in output")
		}
	})

	t.Run("anchor paths preserved", func(t *testing.T) {
		raw := []byte("source_of_truth: work/active/2026-01-01-track/doc.md§heading\n")
		got, count, err := rewritePaths(raw, "work/active/2026-01-01-track", "work/archive/2026-01-01-track")
		if err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Errorf("count = %d, want 1", count)
		}
		want := "source_of_truth: work/archive/2026-01-01-track/doc.md§heading\n"
		if string(got) != want {
			t.Errorf("got %q, want %q", string(got), want)
		}
	})
	t.Run("ambiguous token refuses replacement", func(t *testing.T) {
		raw := []byte("note: work/active/2026-01-01-track-backup/file.md\n")
		got, count, err := rewritePaths(raw, "work/active/2026-01-01-track", "work/archive/2026-01-01-track")
		if err == nil {
			t.Fatal("expected ambiguous token error")
		}
		if got != nil || count != 0 {
			t.Fatalf("rewrite result = %q, %d; want no result", got, count)
		}
	})
}

// scaffoldTestFolio creates a minimal valid folio project with a work track.
// Returns the folio.yml path.
func scaffoldTestFolio(t *testing.T, dir string) string {
	t.Helper()

	trackDir := filepath.Join(dir, "work", "active", "2026-01-01-test-track")
	os.MkdirAll(trackDir, 0755)
	os.WriteFile(filepath.Join(trackDir, "README.md"), []byte("# Test Track\n"), 0644)

	// Ensure archive parent exists
	os.MkdirAll(filepath.Join(dir, "work", "archive"), 0755)

	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test Project"
sources:
  - path: work/active/2026-01-01-test-track/README.md
`), 0644)

	return yml
}

func TestRunArchive(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		dir := t.TempDir()
		yml := scaffoldTestFolio(t, dir)

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		code := buildRoot().Execute([]string{"archive", "--folio", yml, "--no-push", "2026-01-01-test-track"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		buf.ReadFrom(r)

		if code != 0 {
			t.Fatalf("expected exit code 0, got %d", code)
		}

		// Verify directory moved
		if _, err := os.Stat(filepath.Join(dir, "work", "active", "2026-01-01-test-track")); !os.IsNotExist(err) {
			t.Error("active directory should not exist after archive")
		}
		if _, err := os.Stat(filepath.Join(dir, "work", "archive", "2026-01-01-test-track", "README.md")); err != nil {
			t.Error("archive directory should contain README.md")
		}

		// Verify emptied work/active was pruned (no hollow skeleton left behind),
		// while work/ itself is preserved
		if _, err := os.Stat(filepath.Join(dir, "work", "active")); !os.IsNotExist(err) {
			t.Error("empty work/active should be pruned after archiving the last track")
		}
		if _, err := os.Stat(filepath.Join(dir, "work")); err != nil {
			t.Error("work/ should be preserved")
		}

		// Verify folio.yml rewritten
		data, _ := os.ReadFile(yml)
		if strings.Contains(string(data), "work/active/2026-01-01-test-track/") {
			t.Error("folio.yml still contains old active path")
		}
		if !strings.Contains(string(data), "work/archive/2026-01-01-test-track/") {
			t.Error("folio.yml should contain new archive path")
		}
	})

	t.Run("preserves active dir with remaining tracks", func(t *testing.T) {
		dir := t.TempDir()
		yml := scaffoldTestFolio(t, dir)

		// Add a second active track that should survive the archive
		otherTrack := filepath.Join(dir, "work", "active", "2026-02-02-other-track")
		os.MkdirAll(otherTrack, 0755)
		os.WriteFile(filepath.Join(otherTrack, "README.md"), []byte("# Other\n"), 0644)

		code := buildRoot().Execute([]string{"archive", "--folio", yml, "--no-push", "2026-01-01-test-track"})
		if code != 0 {
			t.Fatalf("expected exit code 0, got %d", code)
		}

		// work/active must remain because another track still lives there
		if _, err := os.Stat(otherTrack); err != nil {
			t.Error("remaining track should be untouched")
		}
		if _, err := os.Stat(filepath.Join(dir, "work", "active")); err != nil {
			t.Error("work/active should be preserved when other tracks remain")
		}
	})

	t.Run("missing track", func(t *testing.T) {
		dir := t.TempDir()
		yml := filepath.Join(dir, "folio.yml")
		os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\n"), 0644)

		code := buildRoot().Execute([]string{"archive", "--folio", yml, "--no-push", "nonexistent-track"})
		if code != 1 {
			t.Errorf("expected exit code 1 for missing track, got %d", code)
		}
	})

	t.Run("already archived", func(t *testing.T) {
		dir := t.TempDir()
		yml := scaffoldTestFolio(t, dir)

		// Pre-create archive directory
		os.MkdirAll(filepath.Join(dir, "work", "archive", "2026-01-01-test-track"), 0755)

		code := buildRoot().Execute([]string{"archive", "--folio", yml, "--no-push", "2026-01-01-test-track"})
		if code != 1 {
			t.Errorf("expected exit code 1 for already archived, got %d", code)
		}
	})

	t.Run("dry run", func(t *testing.T) {
		dir := t.TempDir()
		yml := scaffoldTestFolio(t, dir)

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		code := buildRoot().Execute([]string{"archive", "--folio", yml, "--dry-run", "2026-01-01-test-track"})

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		buf.ReadFrom(r)
		out := buf.String()

		if code != 0 {
			t.Fatalf("expected exit code 0 for dry-run, got %d", code)
		}

		// Verify no side effects
		if _, err := os.Stat(filepath.Join(dir, "work", "active", "2026-01-01-test-track")); err != nil {
			t.Error("dry-run should not move directory")
		}

		origData, _ := os.ReadFile(yml)
		if strings.Contains(string(origData), "work/archive/") {
			t.Error("dry-run should not modify folio.yml")
		}

		// Verify output mentions count
		if !strings.Contains(out, "1") {
			t.Error("dry-run output should mention replacement count")
		}
	})

	t.Run("rollback on validation failure", func(t *testing.T) {
		dir := t.TempDir()

		trackDir := filepath.Join(dir, "work", "active", "2026-01-01-bad-track")
		os.MkdirAll(trackDir, 0755)
		os.WriteFile(filepath.Join(trackDir, "README.md"), []byte("# Bad\n"), 0644)
		os.MkdirAll(filepath.Join(dir, "work", "archive"), 0755)

		// Create folio.yml that references a source which won't exist after rewrite
		// The source points to a file that only exists in active, and the rewrite
		// will change the path to archive — but we'll make a second source that
		// breaks validation (missing path field entirely is caught by parse)
		yml := filepath.Join(dir, "folio.yml")
		manifestBefore := []byte(`schema: 1
project: "Test"
sources:
  - path: work/active/2026-01-01-bad-track/README.md
  - path: work/active/2026-01-01-bad-track/missing.md
  - path: work/archive/2026-01-01-bad-track/README.md
`)
		os.WriteFile(yml, manifestBefore, 0644)

		code := buildRoot().Execute([]string{"archive", "--folio", yml, "--no-push", "2026-01-01-bad-track"})
		if code != 1 {
			t.Errorf("expected exit code 1 for validation failure, got %d", code)
		}

		// Verify rollback: active directory should be restored
		if _, err := os.Stat(filepath.Join(dir, "work", "active", "2026-01-01-bad-track")); err != nil {
			t.Error("directory should be rolled back to active after validation failure")
		}

		// folio.yml should be unchanged
		data, _ := os.ReadFile(yml)
		if string(data) != string(manifestBefore) {
			t.Errorf("folio.yml changed after rollback:\n%s", data)
		}
	})
}
func TestArchiveCommandRefusesDependentBeforeMove(t *testing.T) {
	store := t.TempDir()
	t.Setenv("FOLIO_HOME", store)
	t.Setenv("FOLIO_UMBRELLA", "")
	t.Chdir(store)

	target := filepath.Join(store, "active", "target")
	writeArchiveManifest(t, target, `schema: 1
project: target
sources:
  - path: work/active/track/README.md
`)
	writeArchiveFile(t, filepath.Join(target, "work", "active", "track", "README.md"), "track")

	dependent := filepath.Join(store, "active", "dependent")
	writeArchiveManifest(t, dependent, `schema: 1
project: dependent
sources:
  - path: ../target/work/active/track/README.md
`)
	before, err := os.ReadFile(filepath.Join(target, "folio.yml"))
	if err != nil {
		t.Fatal(err)
	}

	if code := buildRoot().Execute([]string{"archive", "--folio", filepath.Join(target, "folio.yml"), "--no-push", "track"}); code == dendrik.ExitOK {
		t.Fatal("archive succeeded despite a dependent project reference")
	}
	if _, err := os.Stat(filepath.Join(target, "work", "active", "track")); err != nil {
		t.Fatalf("track moved despite preflight refusal: %v", err)
	}
	after, err := os.ReadFile(filepath.Join(target, "folio.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatal("target manifest changed during preflight refusal")
	}
}

func TestArchiveRefusesUnfulfillableScopedPush(t *testing.T) {
	store := t.TempDir()
	t.Setenv("FOLIO_HOME", store)
	t.Setenv("FOLIO_UMBRELLA", "")
	t.Chdir(store)

	target := filepath.Join(store, "folio.yml")
	writeArchiveManifest(t, store, `schema: 1
project: target
`)
	writeArchiveFile(t, filepath.Join(store, "work", "active", "track", "README.md"), "track")

	if code := buildRoot().Execute([]string{"archive", "--folio", target, "track"}); code == dendrik.ExitOK {
		t.Fatal("archive succeeded without a jj work root for its promised push")
	}
	if _, err := os.Stat(filepath.Join(store, "work", "active", "track")); err != nil {
		t.Fatalf("track moved before scoped push refusal: %v", err)
	}
}

func writeArchiveManifest(t *testing.T, projectDir, content string) {
	t.Helper()
	writeArchiveFile(t, filepath.Join(projectDir, "folio.yml"), content)
}

func writeArchiveFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
