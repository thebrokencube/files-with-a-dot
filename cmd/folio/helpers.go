package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/home"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/list"
)

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

// isFilePath returns true if the value should be treated as a file path (not a shortname).
func isFilePath(s string) bool {
	if strings.HasSuffix(s, ".yml") || strings.HasSuffix(s, ".yaml") {
		return true
	}
	if strings.HasPrefix(s, ".") || strings.HasPrefix(s, "/") || strings.HasPrefix(s, "~/") {
		return true
	}
	fi, err := os.Stat(s)
	return err == nil && fi.Mode().IsRegular()
}

// matchShortname finds an active entry matching the shortname.
// Pass 1: exact match on entry.Path. Pass 2: exact match on final path component.
func matchShortname(entries []list.Entry, shortname string) (list.Entry, error) {
	// Pass 1: exact match on full path
	for _, e := range entries {
		if e.Section == "active" && e.Path == shortname {
			return e, nil
		}
	}

	// Pass 2: match on final path component
	var matches []list.Entry
	for _, e := range entries {
		if e.Section == "active" {
			base := filepath.Base(e.Path)
			if base == shortname {
				matches = append(matches, e)
			}
		}
	}

	switch len(matches) {
	case 0:
		return list.Entry{}, nil
	case 1:
		return matches[0], nil
	default:
		var paths []string
		for _, m := range matches {
			paths = append(paths, m.Path)
		}
		return list.Entry{}, fmt.Errorf("ambiguous shortname %q matches: %s", shortname, strings.Join(paths, ", "))
	}
}

// activeShortnames returns all active entry paths, sorted alphabetically.
func activeShortnames(entries []list.Entry) []string {
	var paths []string
	for _, e := range entries {
		if e.Section == "active" {
			paths = append(paths, e.Path)
		}
	}
	sort.Strings(paths)
	return paths
}

// resolveFolioPath resolves a --folio flag value to an absolute folio.yml path.
func resolveFolioPath(value string) (string, error) {
	if isFilePath(value) {
		return value, nil
	}

	// Registry-aware write-routing: `--folio <store>:<project>` targets a
	// project in any registered folio store. `isFilePath` gates first, so a
	// bare `work:proj` (no path markers) reaches here; `ben/foo` (no colon)
	// falls through to the home-only path below, unchanged.
	if storeName, project, ok := splitStoreTarget(value); ok {
		reg, err := config.LoadRegistry()
		if err != nil {
			return "", fmt.Errorf("cannot resolve store target: %w", err)
		}
		store, found := reg.Lookup(storeName)
		if !found {
			return "", fmt.Errorf("unknown store %q in --folio %q (not registered in stores.yml)", storeName, value)
		}
		if store.IsExternal() {
			return "", fmt.Errorf("store %q is external (read-only) — not a write target", storeName)
		}
		return resolveInStore(store.Path, storeName, project)
	}
	// Store-shaped but malformed (e.g. `work:` with no project) — error clearly
	// rather than falling through to a confusing home-only shortname scan.
	if i := strings.IndexByte(value, ':'); i > 0 && !strings.ContainsAny(value[:i], "/\\. \t") {
		return "", fmt.Errorf("invalid store target %q — use --folio <store>:<project>", value)
	}

	homeDir, err := home.Dir()
	if err != nil {
		return "", fmt.Errorf("cannot resolve shortname: %w", err)
	}

	entries, err := list.Scan(homeDir)
	if err != nil {
		return "", fmt.Errorf("scanning FOLIO_HOME: %w", err)
	}

	match, err := matchShortname(entries, value)
	if err != nil {
		return "", err
	}
	if match.Path != "" {
		return filepath.Join(homeDir, "active", match.Path, "folio.yml"), nil
	}

	active := activeShortnames(entries)
	if len(active) > 0 {
		return "", fmt.Errorf("unknown project %q — active projects:\n  %s", value, strings.Join(active, "\n  "))
	}
	return "", fmt.Errorf("unknown project %q (no active projects)", value)
}

// splitStoreTarget splits a `<store>:<project>` --folio value. Returns ok=false
// unless the text before the first colon is a bare store identifier (no slash,
// dot, or space) — so `work:my-project` routes to a store but a stray colon in
// a path does not.
func splitStoreTarget(value string) (store, project string, ok bool) {
	i := strings.IndexByte(value, ':')
	if i <= 0 || i == len(value)-1 {
		return "", "", false
	}
	store = value[:i]
	if strings.ContainsAny(store, "/\\. \t") {
		return "", "", false
	}
	return store, value[i+1:], true
}

// resolveInStore finds a project shortname within a folio store root, searching
// both active/ and archive/, and returns the absolute folio.yml path joined
// against the store root.
func resolveInStore(storeRoot, storeName, project string) (string, error) {
	entries, err := list.Scan(storeRoot)
	if err != nil {
		return "", fmt.Errorf("scanning store %q: %w", storeName, err)
	}
	// list.Scan returns entries sorted active-before-archive, so iterating in
	// order means active wins on a same-name active/archive collision.

	// Pass 1: exact path match (active wins by sort order).
	for _, e := range entries {
		if e.Path == project {
			return filepath.Join(storeRoot, e.Section, e.Path, "folio.yml"), nil
		}
	}

	// Pass 2: final-component match. Collect to detect genuine ambiguity
	// (different paths), while treating a same-path active/archive pair as a
	// single project that active resolves.
	var matches []list.Entry
	distinct := map[string]bool{}
	for _, e := range entries {
		if filepath.Base(e.Path) == project {
			matches = append(matches, e)
			distinct[e.Path] = true
		}
	}
	switch {
	case len(matches) == 0:
		return "", fmt.Errorf("unknown project %q in store %q", project, storeName)
	case len(distinct) > 1:
		var labels []string
		for _, m := range matches {
			labels = append(labels, m.Section+"/"+m.Path)
		}
		return "", fmt.Errorf("ambiguous project %q in store %q matches: %s", project, storeName, strings.Join(labels, ", "))
	default:
		e := matches[0] // active-first by sort order
		return filepath.Join(storeRoot, e.Section, e.Path, "folio.yml"), nil
	}
}

// resolveOrDie resolves *folioPath in place. Returns false and prints error on failure.
func resolveOrDie(folioPath *string) bool {
	resolved, err := resolveFolioPath(*folioPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	*folioPath = resolved
	return true
}
