package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("cannot get home dir: %v", err)
	}
	vaultBase := filepath.Join(home, ".folio", "vault")

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
			got := ResolvePath(tt.folioDir, tt.path)
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
