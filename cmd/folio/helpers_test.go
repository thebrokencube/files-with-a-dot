package main

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/list"
)

// setupTestHome creates a temporary FOLIO_HOME with the given active folio.yml files.
func setupTestHome(t *testing.T, paths ...string) string {
	t.Helper()
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "active"), 0755)
	os.MkdirAll(filepath.Join(dir, "archive"), 0755)
	for _, p := range paths {
		full := filepath.Join(dir, "active", p, "folio.yml")
		os.MkdirAll(filepath.Dir(full), 0755)
		os.WriteFile(full, []byte("schema: 1\nproject: test\nsources: []\ntargets: {}\nobservations: []\n"), 0644)
	}
	return dir
}

func TestIsFilePath(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"./folio.yml", true},
		{"/abs/path.yml", true},
		{"~/foo.yaml", true},
		{"../relative.yml", true},
		{"ben/state-retirement-mandates", false},
		{"state-retirement-mandates", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isFilePath(tt.input)
			if got != tt.want {
				t.Errorf("isFilePath(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestMatchShortname(t *testing.T) {
	entries := []list.Entry{
		{Section: "active", Path: "ben/state-retirement-mandates"},
		{Section: "active", Path: "ben/onboarding-redesign"},
		{Section: "active", Path: "personal/side-project"},
		{Section: "archive", Path: "2024-01-01_old-project"},
	}

	t.Run("exact path match", func(t *testing.T) {
		match, err := matchShortname(entries, "ben/state-retirement-mandates")
		if err != nil {
			t.Fatal(err)
		}
		if match.Path != "ben/state-retirement-mandates" {
			t.Errorf("got %q, want ben/state-retirement-mandates", match.Path)
		}
	})

	t.Run("final component match", func(t *testing.T) {
		match, err := matchShortname(entries, "side-project")
		if err != nil {
			t.Fatal(err)
		}
		if match.Path != "personal/side-project" {
			t.Errorf("got %q, want personal/side-project", match.Path)
		}
	})

	t.Run("no match", func(t *testing.T) {
		match, err := matchShortname(entries, "nonexistent")
		if err != nil {
			t.Fatal(err)
		}
		if match.Path != "" {
			t.Errorf("expected empty match, got %q", match.Path)
		}
	})

	t.Run("archive entries ignored", func(t *testing.T) {
		match, err := matchShortname(entries, "old-project")
		if err != nil {
			t.Fatal(err)
		}
		if match.Path != "" {
			t.Errorf("expected empty match for archive entry, got %q", match.Path)
		}
	})

	t.Run("ambiguous final component", func(t *testing.T) {
		ambiguous := []list.Entry{
			{Section: "active", Path: "team-a/shared-name"},
			{Section: "active", Path: "team-b/shared-name"},
		}
		_, err := matchShortname(ambiguous, "shared-name")
		if err == nil {
			t.Fatal("expected ambiguity error")
		}
		if !strings.Contains(err.Error(), "ambiguous") {
			t.Errorf("expected ambiguity error, got: %s", err)
		}
	})
}

func TestActiveShortnames(t *testing.T) {
	entries := []list.Entry{
		{Section: "active", Path: "ben/state-retirement-mandates"},
		{Section: "active", Path: "ben/onboarding-redesign"},
		{Section: "archive", Path: "2024-01-01_old-state-thing"},
	}

	t.Run("returns active only", func(t *testing.T) {
		got := activeShortnames(entries)
		if len(got) != 2 {
			t.Fatalf("expected 2 active paths, got %d: %v", len(got), got)
		}
	})

	t.Run("sorted alphabetically", func(t *testing.T) {
		got := activeShortnames(entries)
		if got[0] != "ben/onboarding-redesign" || got[1] != "ben/state-retirement-mandates" {
			t.Errorf("expected sorted order, got %v", got)
		}
	})

	t.Run("empty when no active", func(t *testing.T) {
		got := activeShortnames([]list.Entry{{Section: "archive", Path: "old"}})
		if len(got) != 0 {
			t.Fatalf("expected 0, got %d", len(got))
		}
	})
}

func TestParseFlagsNormalOrder(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("folio", "", "test flag")
	parseFlags(fs, []string{"--folio", "/tmp/folio.yml", "positional"})
	if fs.Args()[0] != "positional" {
		t.Fatalf("expected positional arg, got %q", fs.Args()[0])
	}
}

func TestParseFlagsNoArgs(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	parseFlags(fs, []string{})
	if len(fs.Args()) != 0 {
		t.Fatalf("expected no args, got %d", len(fs.Args()))
	}
}

func TestParseFlagsPositionalOnly(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	parseFlags(fs, []string{"hello", "world"})
	if len(fs.Args()) != 2 {
		t.Fatalf("expected 2 args, got %d", len(fs.Args()))
	}
}

// TestParseFlagsTrailingDetected verifies that trailing flags cause os.Exit(1).
// We run ourselves as a subprocess to catch the exit.
func TestParseFlagsTrailingDetected(t *testing.T) {
	if os.Getenv("TEST_PARSEFLAGS_EXIT") == "1" {
		fs := flag.NewFlagSet("test", flag.ContinueOnError)
		fs.String("folio", "", "test flag")
		parseFlags(fs, []string{"positional", "--folio", "/tmp"})
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestParseFlagsTrailingDetected")
	cmd.Env = append(os.Environ(), "TEST_PARSEFLAGS_EXIT=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return // expected non-zero exit
	}
	t.Fatal("expected process to exit with non-zero status for trailing flags")
}

func TestResolveFolioPath(t *testing.T) {
	t.Run("file path passthrough", func(t *testing.T) {
		got, err := resolveFolioPath("./folio.yml")
		if err != nil {
			t.Fatal(err)
		}
		if got != "./folio.yml" {
			t.Errorf("got %q, want ./folio.yml", got)
		}
	})

	t.Run("shortname exact match", func(t *testing.T) {
		homeDir := setupTestHome(t, "ben/state-retirement-mandates", "ben/onboarding")
		t.Setenv("FOLIO_HOME", homeDir)

		got, err := resolveFolioPath("ben/state-retirement-mandates")
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(homeDir, "active", "ben/state-retirement-mandates", "folio.yml")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("shortname final component match", func(t *testing.T) {
		homeDir := setupTestHome(t, "ben/state-retirement-mandates", "ben/onboarding")
		t.Setenv("FOLIO_HOME", homeDir)

		got, err := resolveFolioPath("state-retirement-mandates")
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(homeDir, "active", "ben/state-retirement-mandates", "folio.yml")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("miss lists active projects", func(t *testing.T) {
		homeDir := setupTestHome(t, "ben/state-retirement-mandates", "ben/state-tax-engine")
		t.Setenv("FOLIO_HOME", homeDir)

		_, err := resolveFolioPath("nonexistent")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "active projects") {
			t.Errorf("expected 'active projects' in error, got: %s", err)
		}
		if !strings.Contains(err.Error(), "ben/state-retirement-mandates") {
			t.Errorf("expected active projects listed in error, got: %s", err)
		}
	})

	t.Run("miss with no active projects", func(t *testing.T) {
		homeDir := setupTestHome(t) // no active projects
		t.Setenv("FOLIO_HOME", homeDir)

		_, err := resolveFolioPath("zzzzz")
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "no active projects") {
			t.Errorf("expected 'no active projects' in error, got: %s", err)
		}
	})

	t.Run("FOLIO_HOME missing", func(t *testing.T) {
		t.Setenv("FOLIO_HOME", "/nonexistent/path/that/does/not/exist")

		_, err := resolveFolioPath("anything")
		if err == nil {
			t.Fatal("expected error when FOLIO_HOME is invalid")
		}
	})
}
