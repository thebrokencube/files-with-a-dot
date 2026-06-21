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
// Resolution order for a "<store>:" reference:
//  1. a registered store wins (join against its path);
//  2. the reserved "vault:" prefix is intrinsic — the home's shared-reference
//     subdir <home>/vault — NOT a store (it has no folio.yml/lifecycle and is
//     never listed as a peer);
//  3. any other store-shaped prefix fails loud (no silent dangling refs).
//
// A nil reg reproduces legacy behavior: only "vault:" resolves, everything else
// falls through to folioDir.
func ResolvePath(folioDir, path string, reg *Registry) (string, error) {
	prefix, remainder, isRef := splitStorePrefix(path)
	if !isRef {
		return filepath.Join(folioDir, path), nil
	}

	// 1. Registered store wins (lets a user override the intrinsic vault if they
	//    ever register a real store named "vault").
	if store, found := reg.Lookup(prefix); found {
		return filepath.Join(store.Path, remainder), nil
	}

	// 2. Intrinsic vault: the home's shared-reference subdir, not a store.
	if prefix == vaultName {
		if folioHome, err := home.Dir(); err == nil {
			return filepath.Join(folioHome, "vault", remainder), nil
		}
		return filepath.Join(folioDir, path), nil
	}

	// 3. Unknown store-shaped prefix. With a real registry, fail loud; with a
	//    nil reg (legacy safety net), fall through unchanged.
	if reg == nil {
		return filepath.Join(folioDir, path), nil
	}
	return "", fmt.Errorf("unknown store prefix %q in %q — not registered in stores.yml", prefix, path)
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
