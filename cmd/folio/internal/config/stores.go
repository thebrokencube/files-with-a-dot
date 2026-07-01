package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/home"
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
	LocationContained  = "contained"  // under ~/.folio; folio may create/reconstruct it
	LocationReferenced = "referenced" // lives in place; folio reads it, never rewrites its history
)

// maxSchema is the highest stores.yml schema version this binary understands (0 = unset).
const maxSchema = 3

const storesFile = "stores.yml"

// vaultName is the reserved store name for the home's intrinsic shared-reference
// subdir (<home>/vault). It is NOT a registry store — it has no folio.yml and no
// active/archive lifecycle — so it is resolved intrinsically by ResolvePath and
// never listed as a peer store. A user may still register a real store named
// "vault" to override the intrinsic path.
const vaultName = "vault"

// Store is one entry in the global store registry (~/.folio/stores.yml).
type Store struct {
	Name          string // map key, filled on load
	Path          string // ~/ expanded to absolute
	Kind          string // KindFolio | KindExternal | KindCode | KindDot
	Location      string // LocationContained (default) | LocationReferenced
	DefaultBranch string // branch a code push must refuse; "" → "main"
}

// IsExternal reports whether the store is a read-only external KB.
func (s Store) IsExternal() bool { return s.Kind == KindExternal }

// IsReferenced reports whether the store lives in place (folio never rewrites its history).
func (s Store) IsReferenced() bool { return s.Location == LocationReferenced }

// Registry indexes every folio + external KB the user works across. It is
// global: loaded once from the home dir and consulted by all resolution and
// discovery, so a given <store>: prefix means the same thing everywhere.
type Registry struct {
	Stores  map[string]Store // lookup by name
	Order   []string         // declaration order = precedence (home/self implicitly first)
	Default string           // top-level `default:` store name; "" if unset/implicit

	// implicit is true ONLY for the file-absent sentinel produced by
	// defaultRegistry. It is the single source of truth for back-compat: an
	// explicit single-store registry can be byte-identical to the sentinel, so
	// callers MUST read this flag (via isImplicitDefault) rather than inferring
	// "implicit" from empty contents.
	implicit bool
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

// LoadRegistry loads ~/.folio/stores.yml. An absent file yields the implicit
// default registry {vault: {<home>/vault, folio}}, reproducing today's
// single-home behavior byte-for-byte.
func LoadRegistry() (*Registry, error) {
	homeDir, err := home.Dir()
	if err != nil {
		return nil, err
	}
	return LoadRegistryFrom(homeDir)
}

// LoadRegistryFrom loads the registry from a specific home directory. Exposed
// for testing with an isolated FOLIO_HOME.
func LoadRegistryFrom(homeDir string) (*Registry, error) {
	path := filepath.Join(homeDir, storesFile)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return defaultRegistry(homeDir), nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return parseRegistry(data)
}

// defaultRegistry is the implicit registry used when no stores.yml exists. It is
// EMPTY — there are no registered stores in a single-home setup. The `vault:`
// prefix is not a store; it is resolved intrinsically to <home>/vault by
// ResolvePath. (homeDir is unused now but kept for signature stability / future
// container-mode defaults.)
func defaultRegistry(homeDir string) *Registry {
	_ = homeDir
	return &Registry{
		Stores:   map[string]Store{},
		Order:    nil,
		implicit: true,
	}
}

type rawStore struct {
	Path          string `yaml:"path"`
	Kind          string `yaml:"kind"`
	Location      string `yaml:"location"`
	DefaultBranch string `yaml:"default_branch"`
}

// parseRegistry decodes stores.yml. The stores mapping is decoded via a
// yaml.Node so declaration order (and thus precedence) is preserved — plain
// map decoding loses it.
func parseRegistry(data []byte) (*Registry, error) {
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

	reg := &Registry{Stores: map[string]Store{}, Default: doc.Default}
	content := doc.Stores.Content // mapping node: [key, val, key, val, ...]
	for i := 0; i+1 < len(content); i += 2 {
		name := content[i].Value
		var rs rawStore
		if err := content[i+1].Decode(&rs); err != nil {
			return nil, fmt.Errorf("store %q: %w", name, err)
		}
		kind := rs.Kind
		if kind == "" {
			kind = KindExternal // safe default: read-only until told otherwise (never assume the writable folio push)
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
		reg.Stores[name] = Store{Name: name, Path: expandUser(rs.Path), Kind: kind, Location: loc, DefaultBranch: branch}
		reg.Order = append(reg.Order, name)
	}
	return reg, nil
}

// ActiveStore resolves the content-plane store every `home` subcommand acts on
// (via resolveHomeOrFail). ok=false means "fall back to legacy single-home
// (home.Dir())". Resolution order:
//
//  0. implicit registry (no stores.yml) → ok=false immediately (back-compat)
//  1. cwd inside a registered store wins — walk up, prefix-match store roots.
//     [user-pinned: cwd always overrides the default]
//  2. else the `default:` store (error if it names an unregistered store)
//  3. else ok=false
func ActiveStore(reg *Registry) (Store, bool, error) {
	if reg.isImplicitDefault() {
		return Store{}, false, nil
	}
	if cwd, err := os.Getwd(); err == nil {
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

// storeContaining returns the registered store whose root contains dir (dir is
// the root itself or nested under it). Used both by ActiveStore (cwd → store)
// and by vault: resolution (project dir → store root). The longest matching
// root wins, so a store nested inside another beats the outer one. An implicit
// or empty registry contains nothing.
func storeContaining(dir string, reg *Registry) (Store, bool) {
	if reg == nil {
		return Store{}, false
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = filepath.Clean(dir)
	}
	var best Store
	bestLen := -1
	for _, name := range reg.Order {
		root := filepath.Clean(reg.Stores[name].Path)
		if abs == root || strings.HasPrefix(abs, root+string(filepath.Separator)) {
			if len(root) > bestLen {
				best = reg.Stores[name]
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
