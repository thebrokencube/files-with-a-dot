package config

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/home"
)

const vaultPrefix = "vault:"

// ResolvePath resolves a source path to an absolute filesystem path, consulting
// the store registry. A registered "<store>:" prefix joins against that store's
// root; an unregistered-but-store-shaped prefix fails loud (the survey's central
// anti-pattern — silent dangling refs); any other path (including one that
// merely contains a colon) joins against folioDir, unchanged.
//
// A nil reg reproduces legacy behavior: only "vault:" is special-cased.
func ResolvePath(folioDir, path string, reg *Registry) (string, error) {
	prefix, remainder, isRef := splitStorePrefix(path)
	if !isRef {
		return filepath.Join(folioDir, path), nil
	}

	if reg == nil {
		// Back-compat: pre-registry behavior special-cased only vault:.
		if prefix == "vault" {
			if folioHome, err := home.Dir(); err == nil {
				return filepath.Join(folioHome, "vault", remainder), nil
			}
		}
		return filepath.Join(folioDir, path), nil
	}

	store, found := reg.Lookup(prefix)
	if !found {
		return "", fmt.Errorf("unknown store prefix %q in %q — not registered in stores.yml", prefix, path)
	}
	return filepath.Join(store.Path, remainder), nil
}

// splitStorePrefix splits a "<store>:<remainder>" reference. A path is only a
// store reference if the text before the first colon is a bare identifier (no
// slash, dot, or space) — so "vault:research/x.md" is a ref but
// "reference/a:b.md" is just a path that happens to contain a colon.
func splitStorePrefix(path string) (prefix, remainder string, isRef bool) {
	i := strings.IndexByte(path, ':')
	if i <= 0 {
		return "", "", false
	}
	prefix = path[:i]
	if strings.ContainsAny(prefix, "/\\. \t") {
		return "", "", false
	}
	return prefix, path[i+1:], true
}

// IsExternalStorePath reports whether path is a "<store>:" reference into a
// registered external (read-only, non-folio) store. Used to downgrade a
// missing-file error to a warning during validation.
func IsExternalStorePath(path string, reg *Registry) bool {
	prefix, _, isRef := splitStorePrefix(path)
	if !isRef {
		return false
	}
	store, found := reg.Lookup(prefix)
	return found && store.IsExternal()
}

// IsVaultPath returns true if the path uses the vault: prefix.
func IsVaultPath(path string) bool {
	return strings.HasPrefix(path, vaultPrefix)
}
