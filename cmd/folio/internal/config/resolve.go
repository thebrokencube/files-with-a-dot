package config

import (
	"path/filepath"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/home"
)

const vaultPrefix = "vault:"

// ResolvePath resolves a source path to an absolute filesystem path.
// Handles vault: prefix (expands to ~/.folio/vault/) and regular paths
// (relative to folioDir).
func ResolvePath(folioDir, path string) string {
	if strings.HasPrefix(path, vaultPrefix) {
		remainder := strings.TrimPrefix(path, vaultPrefix)
		folioHome, err := home.Dir()
		if err != nil {
			return filepath.Join(folioDir, path)
		}
		return filepath.Join(folioHome, "vault", remainder)
	}
	return filepath.Join(folioDir, path)
}

// IsVaultPath returns true if the path uses the vault: prefix.
func IsVaultPath(path string) bool {
	return strings.HasPrefix(path, vaultPrefix)
}
