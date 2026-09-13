package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// Context is the resolved control-plane and content-plane state for one command.
type Context struct {
	Umbrella  string
	Registry  *Registry
	Store     Store
	WorkRoot  string
	FolioPath string
}

// FolioDir returns the directory containing FolioPath. Store-only contexts have
// no project directory and return an empty string.
func (c Context) FolioDir() string {
	if c.FolioPath == "" {
		return ""
	}
	return filepath.Dir(c.FolioPath)
}

// ResolvePath resolves a project-local, vault, or registered-store reference
// using only the roots already selected for this context.
func (c Context) ResolvePath(ref string) (string, error) {
	prefix, remainder, isRef := splitStorePrefix(ref)
	if !isRef {
		if c.FolioDir() == "" {
			return "", fmt.Errorf("project-local path %q requires a project context", ref)
		}
		return canonicalPath(filepath.Join(c.FolioDir(), ref))
	}

	if store, ok := c.Registry.Lookup(prefix); ok {
		return canonicalPath(filepath.Join(store.Path, remainder))
	}

	if prefix == vaultName {
		if c.FolioDir() == "" {
			if c.Store.Name == "" && c.WorkRoot != "" {
				return canonicalPath(filepath.Join(c.WorkRoot, "vault", remainder))
			}
			if c.Umbrella != "" {
				return canonicalPath(filepath.Join(c.Umbrella, "vault", remainder))
			}
			if c.WorkRoot != "" {
				return canonicalPath(filepath.Join(c.WorkRoot, "vault", remainder))
			}
		} else if c.Store.Kind == KindFolio && c.Store.Path != "" {
			return canonicalPath(filepath.Join(c.Store.Path, "vault", remainder))
		} else if store, ok := storeContaining(c.FolioDir(), c.Registry); ok {
			return canonicalPath(filepath.Join(store.Path, "vault", remainder))
		}
		if c.Umbrella != "" {
			return canonicalPath(filepath.Join(c.Umbrella, "vault", remainder))
		}
		if c.WorkRoot != "" {
			return canonicalPath(filepath.Join(c.WorkRoot, "vault", remainder))
		}
		return "", fmt.Errorf("cannot resolve vault reference %q without an umbrella or work root", ref)
	}

	if c.Registry == nil {
		if c.FolioDir() == "" {
			return "", fmt.Errorf("unknown store prefix %q in %q", prefix, ref)
		}
		return canonicalPath(filepath.Join(c.FolioDir(), ref))
	}
	return "", fmt.Errorf("unknown store prefix %q in %q — not registered in stores.yml", prefix, ref)
}

// ForProject returns this context with a canonical project path selected.
func (c Context) ForProject(folioPath string) (Context, error) {
	if folioPath == "" {
		return Context{}, fmt.Errorf("project path is empty")
	}
	path, err := canonicalPath(folioPath)
	if err != nil {
		return Context{}, err
	}
	c.FolioPath = path
	return c, nil
}

// canonicalPath resolves existing path components through symlinks and keeps
// missing planned leaves below the nearest existing canonical parent.
func canonicalPath(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	abs = filepath.Clean(abs)
	candidate := abs
	var suffix []string
	for {
		_, statErr := os.Lstat(candidate)
		if statErr == nil {
			resolved, err := filepath.EvalSymlinks(candidate)
			if err != nil {
				return "", fmt.Errorf("canonicalizing %s: %w", path, err)
			}
			for i := len(suffix) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, suffix[i])
			}
			return filepath.Clean(resolved), nil
		}
		if !os.IsNotExist(statErr) {
			return "", fmt.Errorf("checking %s: %w", candidate, statErr)
		}
		parent := filepath.Dir(candidate)
		if parent == candidate {
			return abs, nil
		}
		suffix = append(suffix, filepath.Base(candidate))
		candidate = parent
	}
}

// CanonicalPath exposes the shared path normalization for command-level
// resolution without duplicating symlink and planned-leaf handling.
func CanonicalPath(path string) (string, error) {
	return canonicalPath(path)
}

// StoreContaining returns the longest canonical registered-store match.
func StoreContaining(dir string, reg *Registry) (Store, bool) {
	return storeContaining(dir, reg)
}
