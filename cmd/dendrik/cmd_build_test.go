package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveVersion(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "VERSION"), []byte("1.2.3\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("reads VERSION file", func(t *testing.T) {
		got, err := resolveVersion(dir, "")
		if err != nil {
			t.Fatal(err)
		}
		if got != "1.2.3" {
			t.Errorf("got %q, want 1.2.3", got)
		}
	})

	t.Run("override wins over file", func(t *testing.T) {
		got, err := resolveVersion(dir, "9.9.9")
		if err != nil {
			t.Fatal(err)
		}
		if got != "9.9.9" {
			t.Errorf("got %q, want 9.9.9", got)
		}
	})

	t.Run("missing VERSION and no override errors", func(t *testing.T) {
		if _, err := resolveVersion(t.TempDir(), ""); err == nil {
			t.Error("expected error for missing VERSION with no override")
		}
	})

	t.Run("empty VERSION errors", func(t *testing.T) {
		empty := t.TempDir()
		if err := os.WriteFile(filepath.Join(empty, "VERSION"), []byte("  \n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := resolveVersion(empty, ""); err == nil {
			t.Error("expected error for empty VERSION")
		}
	})
}
