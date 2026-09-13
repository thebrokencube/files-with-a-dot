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
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/repo"
)

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

type contextMode uint8

const (
	contextReadOnly contextMode = iota
	contextMutation
	contextStoreSync
	contextFleet
)

type contextRoots struct {
	umbrella         string
	registry         *config.Registry
	workRoot         string
	explicitWorkRoot string
}

// resolveContext is the sole command-level entry point for Folio root and
// project resolution. Lower packages receive the returned context instead of
// rediscovering environment state.
func resolveContext(ref string, mode contextMode) (config.Context, error) {
	roots, err := discoverContextRoots()
	if err != nil {
		return config.Context{}, err
	}

	if mode == contextFleet {
		workRoot := roots.umbrella
		if workRoot == "" {
			workRoot = roots.workRoot
		}
		return config.Context{Umbrella: roots.umbrella, Registry: roots.registry, WorkRoot: workRoot}, nil
	}

	if mode == contextStoreSync {
		return resolveStoreSyncContext(roots, ref)
	}

	if ref == "" {
		return config.Context{}, fmt.Errorf("folio reference is empty")
	}

	if storeName, project, ok := splitStoreTarget(ref); ok {
		store, found := roots.registry.Lookup(storeName)
		if !found {
			return config.Context{}, fmt.Errorf("unknown store %q in --folio %q (not registered in stores.yml)", storeName, ref)
		}
		if store.Kind != config.KindFolio {
			return config.Context{}, fmt.Errorf("store %q is %s, not a Folio store", storeName, store.Kind)
		}
		ctx := config.Context{Umbrella: roots.umbrella, Registry: roots.registry, Store: store, WorkRoot: store.Path}
		if err := rejectMutationMismatch(ctx, roots, mode); err != nil {
			return config.Context{}, err
		}
		path, err := resolveInStore(store.Path, storeName, project)
		if err != nil {
			return config.Context{}, err
		}
		return ctx.ForProject(path)
	}

	if isFilePath(ref) {
		return resolveExplicitProject(roots, ref, mode)
	}

	root, store, err := projectSearchRoot(roots)
	if err != nil {
		return config.Context{}, err
	}
	ctx := config.Context{Umbrella: roots.umbrella, Registry: roots.registry, Store: store, WorkRoot: root}
	if err := rejectMutationMismatch(ctx, roots, mode); err != nil {
		return config.Context{}, err
	}
	path, err := resolveInStore(root, store.Name, ref)
	if err != nil {
		if !strings.HasPrefix(err.Error(), "unknown project ") {
			return config.Context{}, err
		}
		active := activeShortnamesFromRoot(root)
		if len(active) > 0 {
			return config.Context{}, fmt.Errorf("unknown project %q — active projects:\n  %s", ref, strings.Join(active, "\n  "))
		}
		return config.Context{}, fmt.Errorf("unknown project %q (no active projects)", ref)
	}
	return ctx.ForProject(path)
}

func discoverContextRoots() (contextRoots, error) {
	var roots contextRoots
	umbrellaOverride, umbrellaSet, err := home.UmbrellaOverride()
	if err != nil {
		return roots, err
	}
	workOverride, workSet, err := home.WorkRootOverride()
	if err != nil {
		return roots, err
	}

	cwd := mustGetwd()
	cwd, err = config.CanonicalPath(cwd)
	if err != nil {
		return roots, fmt.Errorf("canonicalizing cwd: %w", err)
	}

	if umbrellaSet {
		roots.umbrella, err = config.CanonicalPath(umbrellaOverride)
		if err != nil {
			return roots, fmt.Errorf("canonicalizing FOLIO_UMBRELLA: %w", err)
		}
	} else {
		var candidate string
		if workSet {
			candidate = findUmbrella(workOverride)
			if candidate != "" {
				roots.umbrella = candidate
			}
		}
		if roots.umbrella == "" {
			roots.umbrella = findUmbrella(cwd)
		}
		if roots.umbrella == "" && repo.IsJJ(cwd) {
			if defaultRoot, rootErr := repo.DefaultWorkspaceRoot(cwd); rootErr == nil {
				roots.umbrella = findUmbrella(defaultRoot)
			}
		}
	}

	if roots.umbrella != "" {
		roots.registry, err = config.LoadRegistryFrom(roots.umbrella)
		if err != nil {
			return roots, err
		}
	}

	if workSet {
		workRoot, pathErr := config.CanonicalPath(workOverride)
		if pathErr != nil {
			return roots, fmt.Errorf("canonicalizing FOLIO_HOME: %w", pathErr)
		}
		if roots.umbrella != "" && sameCanonicalPath(workRoot, roots.umbrella) && hasStoresFile(workRoot) {
			fmt.Fprintln(os.Stderr, "warning: FOLIO_HOME points to the registry umbrella; use FOLIO_UMBRELLA for control and FOLIO_HOME for a content work root")
		} else {
			roots.explicitWorkRoot = workRoot
			roots.workRoot = workRoot
		}
	}

	if roots.registry == nil {
		if roots.explicitWorkRoot != "" {
			roots.registry, err = config.LoadRegistryFrom(roots.explicitWorkRoot)
			if err != nil {
				return roots, err
			}
			roots.workRoot = roots.explicitWorkRoot
		} else {
			fallback, fallbackErr := home.Dir()
			if fallbackErr != nil {
				return roots, fallbackErr
			}
			roots.workRoot, err = config.CanonicalPath(fallback)
			if err != nil {
				return roots, err
			}
			roots.registry, err = config.LoadRegistryFrom(roots.workRoot)
			if err != nil {
				return roots, err
			}
		}
	}

	if roots.registry.IsImplicit() && roots.workRoot == "" {
		if roots.umbrella != "" {
			roots.workRoot = roots.umbrella
		} else {
			roots.workRoot, err = config.CanonicalPath(cwd)
			if err != nil {
				return roots, err
			}
		}
	}

	if !roots.registry.IsImplicit() && roots.workRoot == "" {
		if store, ok := config.StoreContaining(cwd, roots.registry); ok && store.Kind == config.KindFolio {
			roots.workRoot = store.Path
		} else if workspaceRoot, ok := workspaceStoreRoot(cwd, roots.registry); ok {
			roots.workRoot = workspaceRoot
		}
	}
	return roots, nil
}

func findUmbrella(start string) string {
	if start == "" {
		return ""
	}
	path, err := config.CanonicalPath(start)
	if err != nil {
		return ""
	}
	if info, statErr := os.Stat(path); statErr == nil && !info.IsDir() {
		path = filepath.Dir(path)
	}
	for {
		if hasStoresFile(path) {
			return path
		}
		parent := filepath.Dir(path)
		if parent == path {
			return ""
		}
		path = parent
	}
}

func hasStoresFile(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, "stores.yml"))
	return err == nil && !info.IsDir()
}

func sameCanonicalPath(left, right string) bool {
	leftPath, leftErr := config.CanonicalPath(left)
	rightPath, rightErr := config.CanonicalPath(right)
	return leftErr == nil && rightErr == nil && leftPath == rightPath
}
func canonicalOr(path string) string {
	resolved, err := config.CanonicalPath(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return resolved
}

func workspaceStoreRoot(dir string, reg *config.Registry) (string, bool) {
	if !repo.IsJJ(dir) {
		return "", false
	}
	defaultRoot, err := repo.DefaultWorkspaceRoot(dir)
	if err != nil {
		return "", false
	}
	_, ok := config.StoreContaining(defaultRoot, reg)
	if !ok {
		return "", false
	}
	return canonicalOr(dir), true
}

func resolveStoreSyncContext(roots contextRoots, storeName string) (config.Context, error) {
	if storeName != "" {
		store, ok := roots.registry.Lookup(storeName)
		if !ok {
			return config.Context{}, fmt.Errorf("store %q is not registered in stores.yml", storeName)
		}
		return config.Context{Umbrella: roots.umbrella, Registry: roots.registry, Store: store, WorkRoot: store.Path}, nil
	}

	if roots.explicitWorkRoot != "" {
		store := storeForWorkRoot(roots.explicitWorkRoot, roots.registry)
		return config.Context{Umbrella: roots.umbrella, Registry: roots.registry, Store: store, WorkRoot: roots.explicitWorkRoot}, nil
	}

	cwd := mustGetwd()
	if store, ok := config.StoreContaining(cwd, roots.registry); ok {
		return config.Context{Umbrella: roots.umbrella, Registry: roots.registry, Store: store, WorkRoot: store.Path}, nil
	}
	if store, ok := workspaceStore(cwd, roots.registry); ok {
		return config.Context{Umbrella: roots.umbrella, Registry: roots.registry, Store: store, WorkRoot: canonicalOr(cwd)}, nil
	}
	if roots.registry != nil && roots.registry.Default != "" {
		store, ok := roots.registry.Lookup(roots.registry.Default)
		if !ok {
			return config.Context{}, fmt.Errorf("default store %q is not registered in stores.yml", roots.registry.Default)
		}
		return config.Context{Umbrella: roots.umbrella, Registry: roots.registry, Store: store, WorkRoot: store.Path}, nil
	}
	workRoot := roots.workRoot
	if workRoot == "" {
		workRoot = roots.umbrella
	}
	return config.Context{Umbrella: roots.umbrella, Registry: roots.registry, WorkRoot: workRoot}, nil
}

func storeForWorkRoot(workRoot string, reg *config.Registry) config.Store {
	if store, ok := config.StoreContaining(workRoot, reg); ok {
		return store
	}
	if store, ok := workspaceStore(workRoot, reg); ok {
		return store
	}
	return config.Store{}
}

func workspaceStore(dir string, reg *config.Registry) (config.Store, bool) {
	if !repo.IsJJ(dir) {
		return config.Store{}, false
	}
	defaultRoot, err := repo.DefaultWorkspaceRoot(dir)
	if err != nil {
		return config.Store{}, false
	}
	return config.StoreContaining(defaultRoot, reg)
}

func workspaceStoreBase(dir string, reg *config.Registry) (string, config.Store, bool) {
	if !repo.IsJJ(dir) {
		return "", config.Store{}, false
	}
	defaultRoot, err := repo.DefaultWorkspaceRoot(dir)
	if err != nil {
		return "", config.Store{}, false
	}
	store, ok := config.StoreContaining(defaultRoot, reg)
	if !ok {
		return "", config.Store{}, false
	}
	return canonicalOr(defaultRoot), store, true
}

func projectSearchRoot(roots contextRoots) (string, config.Store, error) {
	if roots.explicitWorkRoot != "" {
		store := storeForWorkRoot(roots.explicitWorkRoot, roots.registry)
		if store.Name != "" && store.Kind != config.KindFolio {
			return "", config.Store{}, fmt.Errorf("FOLIO_HOME points to non-Folio store %q (%s); select a Folio work root or use --folio <store>:<project>", store.Name, store.Kind)
		}
		return roots.explicitWorkRoot, store, nil
	}

	cwd := mustGetwd()
	if store, ok := config.StoreContaining(cwd, roots.registry); ok && store.Kind == config.KindFolio {
		return store.Path, store, nil
	}
	if store, ok := workspaceStore(cwd, roots.registry); ok && store.Kind == config.KindFolio {
		return canonicalOr(cwd), store, nil
	}
	if roots.registry != nil && roots.registry.Default != "" {
		store, ok := roots.registry.Lookup(roots.registry.Default)
		if !ok {
			return "", config.Store{}, fmt.Errorf("default store %q is not registered in stores.yml", roots.registry.Default)
		}
		if store.Kind != config.KindFolio {
			return "", config.Store{}, fmt.Errorf("default store %q is %s, not a Folio store; configure a Folio default or set FOLIO_HOME", store.Name, store.Kind)
		}
		return store.Path, store, nil
	}
	if roots.registry == nil || roots.registry.IsImplicit() {
		if roots.workRoot != "" {
			return roots.workRoot, config.Store{}, nil
		}
	}
	return "", config.Store{}, fmt.Errorf("no Folio work root selected; set FOLIO_HOME to a Folio store or configure default: in stores.yml")
}
func resolveFolioRootForInit() (string, error) {
	roots, err := discoverContextRoots()
	if err != nil {
		return "", err
	}
	if roots.explicitWorkRoot != "" {
		if store := storeForWorkRoot(roots.explicitWorkRoot, roots.registry); store.Name != "" && store.Kind != config.KindFolio {
			return "", fmt.Errorf("FOLIO_HOME points to non-Folio store %q (%s)", store.Name, store.Kind)
		}
		return roots.explicitWorkRoot, nil
	}
	if store, ok := config.StoreContaining(mustGetwd(), roots.registry); ok && store.Kind == config.KindFolio {
		return store.Path, nil
	}
	if store, ok := workspaceStore(mustGetwd(), roots.registry); ok && store.Kind == config.KindFolio {
		return canonicalOr(mustGetwd()), nil
	}
	if roots.registry != nil && roots.registry.Default != "" {
		store, ok := roots.registry.Lookup(roots.registry.Default)
		if !ok {
			return "", fmt.Errorf("default store %q is not registered in stores.yml", roots.registry.Default)
		}
		if store.Kind != config.KindFolio {
			return "", fmt.Errorf("default store %q is %s, not a Folio store", store.Name, store.Kind)
		}
		return store.Path, nil
	}
	if roots.workRoot != "" {
		return roots.workRoot, nil
	}
	return "", fmt.Errorf("no Folio work root selected")
}

func resolveExplicitProject(roots contextRoots, ref string, mode contextMode) (config.Context, error) {
	path, err := config.CanonicalPath(ref)
	if err != nil {
		return config.Context{}, err
	}
	store := config.Store{}
	workRoot := roots.workRoot
	if roots.registry != nil && !roots.registry.IsImplicit() {
		if owning, ok := config.StoreContaining(filepath.Dir(path), roots.registry); ok {
			store = owning
			workRoot = owning.Path
			if roots.explicitWorkRoot != "" && configPathWithin(owning.Path, roots.explicitWorkRoot) {
				workRoot = roots.explicitWorkRoot
			}
		} else if roots.explicitWorkRoot != "" && configPathWithin(roots.explicitWorkRoot, path) {
			store = storeForWorkRoot(roots.explicitWorkRoot, roots.registry)
			workRoot = roots.explicitWorkRoot
		} else {
			workRoot = filepath.Dir(path)
		}
	}
	if workRoot == "" {
		workRoot = filepath.Dir(path)
	}
	ctx := config.Context{Umbrella: roots.umbrella, Registry: roots.registry, Store: store, WorkRoot: workRoot, FolioPath: path}
	if err := rejectMutationMismatch(ctx, roots, mode); err != nil {
		return config.Context{}, err
	}
	return ctx, nil
}

func configPathWithin(root, path string) bool {
	rootPath, rootErr := config.CanonicalPath(root)
	pathValue, pathErr := config.CanonicalPath(path)
	if rootErr != nil || pathErr != nil {
		return false
	}
	return pathValue == rootPath || strings.HasPrefix(pathValue, rootPath+string(filepath.Separator))
}

func selectedAmbientStore(roots contextRoots) config.Store {
	if roots.explicitWorkRoot != "" {
		if store := storeForWorkRoot(roots.explicitWorkRoot, roots.registry); store.Name != "" {
			return store
		}
	}
	if store, ok := config.StoreContaining(mustGetwd(), roots.registry); ok {
		return store
	}
	if store, ok := workspaceStore(mustGetwd(), roots.registry); ok {
		return store
	}
	return config.Store{}
}

func rejectMutationMismatch(ctx config.Context, roots contextRoots, mode contextMode) error {
	if mode != contextMutation || ctx.Store.Name == "" {
		return nil
	}
	ambient := selectedAmbientStore(roots)
	if ambient.Name != "" && ambient.Name != ctx.Store.Name {
		return fmt.Errorf("mutation target store %q conflicts with current store %q; select the target workspace explicitly", ctx.Store.Name, ambient.Name)
	}
	if ambient.Name == ctx.Store.Name {
		return nil
	}
	if roots.explicitWorkRoot != "" && !configPathWithin(ctx.WorkRoot, roots.explicitWorkRoot) && !configPathWithin(roots.explicitWorkRoot, ctx.WorkRoot) {
		return fmt.Errorf("mutation target work root %q conflicts with FOLIO_HOME work root %q", ctx.WorkRoot, roots.explicitWorkRoot)
	}
	return nil
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

func activeShortnamesFromRoot(root string) []string {
	entries, err := list.Scan(root)
	if err != nil {
		return nil
	}
	return activeShortnames(entries)
}

// resolveFolioPath resolves a --folio flag value through the command context.
func resolveFolioPath(value string) (string, error) {
	ctx, err := resolveContext(value, contextReadOnly)
	if err != nil {
		return "", err
	}
	return ctx.FolioPath, nil
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

// resolveInStore finds an active project shortname within a folio store root
// and returns the absolute folio.yml path joined against that store root.
func resolveInStore(storeRoot, storeName, project string) (string, error) {
	canonicalRoot, err := config.CanonicalPath(storeRoot)
	if err != nil {
		return "", fmt.Errorf("canonicalizing store %q: %w", storeName, err)
	}
	storeRoot = canonicalRoot
	entries, err := list.Scan(storeRoot)
	if err != nil {
		return "", fmt.Errorf("scanning store %q: %w", storeName, err)
	}
	// list.Scan returns entries sorted deterministically; only active entries
	// participate in shortname resolution.

	// Pass 1: exact active path match.
	for _, e := range entries {
		if e.Section != "active" {
			continue
		}
		if e.Path == project {
			return filepath.Join(storeRoot, e.Section, e.Path, "folio.yml"), nil
		}
	}

	// Pass 2: final-component match among active entries.
	var matches []list.Entry
	distinct := map[string]bool{}
	for _, e := range entries {
		if e.Section != "active" {
			continue
		}
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
		e := matches[0]
		return filepath.Join(storeRoot, e.Section, e.Path, "folio.yml"), nil
	}
}
