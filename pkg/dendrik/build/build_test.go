package build

import (
	"runtime"
	"testing"
)

func TestParseVersion(t *testing.T) {
	t.Run("trims file content", func(t *testing.T) {
		got, err := ParseVersion("1.2.3\n", "")
		if err != nil {
			t.Fatal(err)
		}
		if got != "1.2.3" {
			t.Errorf("got %q, want 1.2.3", got)
		}
	})

	t.Run("override wins over content", func(t *testing.T) {
		got, err := ParseVersion("1.2.3", "9.9.9")
		if err != nil {
			t.Fatal(err)
		}
		if got != "9.9.9" {
			t.Errorf("got %q, want 9.9.9", got)
		}
	})

	t.Run("empty content and no override errors", func(t *testing.T) {
		if _, err := ParseVersion("  \n", ""); err == nil {
			t.Error("expected error for empty content")
		}
	})
}

func TestTargets(t *testing.T) {
	host := Targets(false)
	if len(host) != 1 || host[0].OS != runtime.GOOS || host[0].Arch != runtime.GOARCH {
		t.Errorf("host targets = %+v, want single host pair", host)
	}

	m := Targets(true)
	if len(m) != len(ReleaseMatrix) {
		t.Fatalf("matrix len = %d, want %d", len(m), len(ReleaseMatrix))
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
	if got := ArtifactName("folio", "darwin", "arm64"); got != "folio-darwin-arm64" {
		t.Errorf("got %q", got)
	}
}

func TestLDFlags(t *testing.T) {
	got := LDFlags("0.6.0")
	want := "-buildid= -X main.version=0.6.0"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
