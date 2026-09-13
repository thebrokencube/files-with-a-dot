package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Store kinds. Behavior (writability, scan, discovery) is derived from kind,
// not stored as separate fields.
const (
	KindFolio    = "folio"    // a full folio home — listed, structure-aware, writable, validated
	KindExternal = "external" // a non-folio KB — read-only, content-grep, missing target warns
	KindCode     = "code"     // a code repo folio hovers over — branch + delegate, never main
	KindDot      = "dot"      // a dotfiles repo managed by the `dot` tool — delegates wholesale to `dot`
)

const (
	LocationContained  = "contained"  // under the umbrella; folio may create/reconstruct it
	LocationReferenced = "referenced" // lives in place; folio reads it, never rewrites its history
)

// maxSchema is the highest stores.yml schema version this binary understands (0 = unset).
const maxSchema = 3

const storesFile = "stores.yml"

// vaultName is the reserved prefix for a selected store's intrinsic shared
// reference subdir (<work-root>/vault). It is not a registry store — it has no
// folio.yml and no active/archive lifecycle — so ResolvePath handles it
// intrinsically. A user may register a real store named "vault" to override it.
const vaultName = "vault"

// Store is one entry in the registry owned by the control root.
type Store struct {
	Name          string // map key, filled on load
	Path          string // expanded and canonical absolute path
	Kind          string // KindFolio | KindExternal | KindCode | KindDot
	Location      string // LocationContained (default) | LocationReferenced
	DefaultBranch string // branch a code push must refuse; "" → "main"
}

// IsExternal reports whether the store is a read-only external KB.
func (s Store) IsExternal() bool { return s.Kind == KindExternal }

// IsReferenced reports whether the store lives in place (folio never rewrites its history).
func (s Store) IsReferenced() bool { return s.Location == LocationReferenced }

// Registry indexes every Folio + external KB the user works across. It is
// loaded from the selected control root and consulted by all context and
// discovery paths, so a given <store>: prefix means the same thing everywhere.
type Registry struct {
	Stores  map[string]Store // lookup by name
	Order   []string         // declaration order = precedence
	Default string           // top-level `default:` store name; "" if unset/implicit

	// umbrella is the canonical directory that owns stores.yml.
	umbrella string

	// implicit is true ONLY for the file-absent sentinel produced by
	// defaultRegistry. It is the single source of truth for back-compat: an
	// explicit single-store registry can be byte-identical to the sentinel, so
	// callers MUST read this flag (via isImplicitDefault) rather than inferring
	// "implicit" from empty contents.
	implicit bool
}

// IsImplicit reports whether the registry came from an absent stores.yml.
func (r *Registry) IsImplicit() bool {
	return r.isImplicitDefault()
}

// LoadRegistryFrom loads stores.yml from a specific control root. An absent file
// returns the implicit registry used by legacy isolated homes.
func LoadRegistryFrom(homeDir string) (*Registry, error) {
	umbrella, err := canonicalPath(homeDir)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(umbrella, storesFile)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return defaultRegistry(umbrella), nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return parseRegistryAt(data, umbrella)
}

func defaultRegistry(homeDir string) *Registry {
	return &Registry{
		Stores:   map[string]Store{},
		Order:    nil,
		umbrella: homeDir,
		implicit: true,
	}
}

func parseRegistry(data []byte) (*Registry, error) {
	return parseRegistryAt(data, "")
}

func parseRegistryAt(data []byte, umbrella string) (*Registry, error) {
	var doc struct {
		Schema  int       `yaml:"schema"`
		Default string    `yaml:"default"`
		Stores  yaml.Node `yaml:"stores"`
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", storesFile, err)
	}
	if doc.Schema > maxSchema {
		return nil, fmt.Errorf("%s: unsupported schema version %d (max %d)", storesFile, doc.Schema, maxSchema)
	}

	reg := &Registry{Stores: map[string]Store{}, Default: doc.Default, umbrella: umbrella}
	content := doc.Stores.Content
	for i := 0; i+1 < len(content); i += 2 {
		name := content[i].Value
		var rs rawStore
		if err := content[i+1].Decode(&rs); err != nil {
			return nil, fmt.Errorf("store %q: %w", name, err)
		}
		kind := rs.Kind
		if kind == "" {
			kind = KindExternal
		}
		switch kind {
		case KindFolio, KindExternal, KindCode, KindDot:
		default:
			return nil, fmt.Errorf("store %q: invalid kind %q (want %s|%s|%s|%s)", name, kind, KindFolio, KindExternal, KindCode, KindDot)
		}
		loc := rs.Location
		if loc == "" {
			loc = LocationContained
		}
		if loc != LocationContained && loc != LocationReferenced {
			return nil, fmt.Errorf("store %q: invalid location %q (want %s|%s)", name, loc, LocationContained, LocationReferenced)
		}
		branch := rs.DefaultBranch
		if branch == "" {
			branch = "main"
		}
		if rs.Path == "" {
			return nil, fmt.Errorf("store %q: missing path", name)
		}
		if _, dup := reg.Stores[name]; dup {
			return nil, fmt.Errorf("store %q: declared more than once", name)
		}
		storePath := expandUser(rs.Path)
		if umbrella != "" && !filepath.IsAbs(storePath) {
			storePath = filepath.Join(umbrella, storePath)
		}
		canonicalStorePath, err := canonicalPath(storePath)
		if err != nil {
			return nil, fmt.Errorf("store %q path: %w", name, err)
		}
		storePath = canonicalStorePath
		reg.Stores[name] = Store{Name: name, Path: storePath, Kind: kind, Location: loc, DefaultBranch: branch}
		reg.Order = append(reg.Order, name)
	}
	return reg, nil
}

// isImplicitDefault reports whether this registry is the file-absent sentinel
// (no stores.yml on disk). ActiveStore short-circuits to legacy single-home in
// that case. Reads the implicit flag, never infers from contents.
func (r *Registry) isImplicitDefault() bool {
	return r != nil && r.implicit
}

// Lookup returns the store registered under prefix.
func (r *Registry) Lookup(prefix string) (Store, bool) {
	if r == nil {
		return Store{}, false
	}
	s, ok := r.Stores[prefix]
	return s, ok
}

// FolioStores returns kind=="folio" stores in declaration order — the set the
// discovery fan-out iterates structure-aware.
func (r *Registry) FolioStores() []Store {
	if r == nil {
		return nil
	}
	var out []Store
	for _, name := range r.Order {
		if s := r.Stores[name]; s.Kind == KindFolio {
			out = append(out, s)
		}
	}
	return out
}

// AllStores returns every registered store in declaration order, regardless of
// kind — the set the fleet fan-out (fleet status) iterates. Unlike FolioStores
// it includes external, code, and dot stores.
func (r *Registry) AllStores() []Store {
	if r == nil {
		return nil
	}
	out := make([]Store, 0, len(r.Order))
	for _, name := range r.Order {
		out = append(out, r.Stores[name])
	}
	return out
}

type rawStore struct {
	Path          string `yaml:"path"`
	Kind          string `yaml:"kind"`
	Location      string `yaml:"location"`
	DefaultBranch string `yaml:"default_branch"`
}

// ActiveStore resolves the content-plane store every `home` subcommand acts on
// (via the command-level context resolver). ok=false means callers should use
// the legacy isolated content root.
func ActiveStore(reg *Registry) (Store, bool, error) {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = ""
	}
	return ActiveStoreAt(reg, cwd)
}

// ActiveStoreAt resolves a registered store using an explicit working
// directory, avoiding ambient process state for context-sensitive callers.
func ActiveStoreAt(reg *Registry, cwd string) (Store, bool, error) {
	if reg.isImplicitDefault() {
		return Store{}, false, nil
	}
	if cwd != "" {
		if s, ok := storeContaining(cwd, reg); ok {
			return s, true, nil
		}
	}
	if reg.Default != "" {
		s, ok := reg.Lookup(reg.Default)
		if !ok {
			return Store{}, false, fmt.Errorf("default store %q is not registered in %s", reg.Default, storesFile)
		}
		return s, true, nil
	}
	return Store{}, false, nil
}

func storeContaining(dir string, reg *Registry) (Store, bool) {
	if reg == nil || dir == "" {
		return Store{}, false
	}
	canonicalDir, err := canonicalPath(dir)
	if err != nil {
		return Store{}, false
	}
	var best Store
	bestLen := -1
	for _, name := range reg.Order {
		store := reg.Stores[name]
		root, err := canonicalPath(store.Path)
		if err != nil {
			continue
		}
		if canonicalDir == root || strings.HasPrefix(canonicalDir, root+string(filepath.Separator)) {
			if len(root) > bestLen {
				store.Path = root
				best = store
				bestLen = len(root)
			}
		}
	}
	return best, bestLen >= 0
}

// expandUser expands a leading ~ or ~/ to the OS user home directory. Other
// paths are returned cleaned; downstream resolution treats them as-is.
func expandUser(path string) string {
	if path == "~" {
		if h, err := os.UserHomeDir(); err == nil {
			return h
		}
		return path
	}
	if strings.HasPrefix(path, "~/") {
		if h, err := os.UserHomeDir(); err == nil {
			return filepath.Join(h, path[2:])
		}
		return path
	}
	return filepath.Clean(path)
}
