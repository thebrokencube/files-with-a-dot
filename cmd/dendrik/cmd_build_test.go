package main

import (
	"os"
	"path/filepath"
	"runtime"
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

func TestBuildTargets(t *testing.T) {
	host := buildTargets(false)
	if len(host) != 1 || host[0].OS != runtime.GOOS || host[0].Arch != runtime.GOARCH {
		t.Errorf("host targets = %+v, want single host pair", host)
	}

	m := buildTargets(true)
	if len(m) != len(releaseMatrix) {
		t.Fatalf("matrix len = %d, want %d", len(m), len(releaseMatrix))
	}
	// The matrix must include the two platforms dot sync / the marketplace consume.
	want := map[string]bool{"darwin/arm64": false, "linux/amd64": false}
	for _, target := range m {
		want[target.OS+"/"+target.Arch] = true
	}
	for k, seen := range want {
		if !seen {
			t.Errorf("matrix missing %s", k)
		}
	}
}

func TestArtifactName(t *testing.T) {
	if got := artifactName("folio", "darwin", "arm64"); got != "folio-darwin-arm64" {
		t.Errorf("got %q", got)
	}
}

func TestBuildLDFlags(t *testing.T) {
	got := buildLDFlags("0.6.0")
	want := "-buildid= -X main.version=0.6.0"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
