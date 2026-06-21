package config

import (
	"path/filepath"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/home"
)

func TestResolvePath(t *testing.T) {
	folioHome, err := home.Dir()
	if err != nil {
		t.Fatalf("cannot get folio home: %v", err)
	}
	vaultBase := filepath.Join(folioHome, "vault")

	// Explicit registry mirroring the implicit single-home default, so these
	// expectations stay independent of any real ~/.folio/stores.yml.
	reg := &Registry{
		Stores: map[string]Store{"vault": {Name: "vault", Path: vaultBase, Kind: KindFolio}},
		Order:  []string{"vault"},
	}

	tests := []struct {
		name     string
		folioDir string
		path     string
		want     string
	}{
		{
			name:     "regular relative path",
			folioDir: "/projects/foo",
			path:     "reference/research/bar.md",
			want:     "/projects/foo/reference/research/bar.md",
		},
		{
			name:     "vault research path",
			folioDir: "/projects/foo",
			path:     "vault:research/2026-03-01-test.md",
			want:     filepath.Join(vaultBase, "research/2026-03-01-test.md"),
		},
		{
			name:     "vault domain path",
			folioDir: "/projects/foo",
			path:     "vault:domain/some-topic.md",
			want:     filepath.Join(vaultBase, "domain/some-topic.md"),
		},
		{
			name:     "vault bare prefix",
			folioDir: "/projects/foo",
			path:     "vault:",
			want:     vaultBase,
		},
		{
			name:     "vault with trailing slash",
			folioDir: "/projects/foo",
			path:     "vault:research/",
			want:     filepath.Join(vaultBase, "research"),
		},
		{
			name:     "regular path ignores vault in name",
			folioDir: "/projects/foo",
			path:     "reference/vault-notes.md",
			want:     "/projects/foo/reference/vault-notes.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolvePath(tt.folioDir, tt.path, reg)
			if err != nil {
				t.Fatalf("ResolvePath(%q, %q) returned error: %v", tt.folioDir, tt.path, err)
			}
			if got != tt.want {
				t.Errorf("ResolvePath(%q, %q) = %q, want %q", tt.folioDir, tt.path, got, tt.want)
			}
		})
	}
}

func TestIsVaultPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"vault:research/foo.md", true},
		{"vault:", true},
		{"reference/vault-notes.md", false},
		{"", false},
		{"vault", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsVaultPath(tt.path); got != tt.want {
				t.Errorf("IsVaultPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
