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
)

const storesFile = "stores.yml"

// vaultName is the reserved store name for the home's intrinsic shared-reference
// subdir (<home>/vault). It is NOT a registry store — it has no folio.yml and no
// active/archive lifecycle — so it is resolved intrinsically by ResolvePath and
// never listed as a peer store. A user may still register a real store named
// "vault" to override the intrinsic path.
const vaultName = "vault"

// Store is one entry in the global store registry (~/.folio/stores.yml).
type Store struct {
	Name string // map key, filled on load
	Path string // ~/ expanded to absolute
	Kind string // KindFolio | KindExternal
}

// IsExternal reports whether the store is a read-only external KB.
func (s Store) IsExternal() bool { return s.Kind == KindExternal }

// Registry indexes every folio + external KB the user works across. It is
// global: loaded once from the home dir and consulted by all resolution and
// discovery, so a given <store>: prefix means the same thing everywhere.
type Registry struct {
	Stores map[string]Store // lookup by name
	Order  []string         // declaration order = precedence (home/self implicitly first)
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
		Stores: map[string]Store{},
		Order:  nil,
	}
}

type rawStore struct {
	Path string `yaml:"path"`
	Kind string `yaml:"kind"`
}

// parseRegistry decodes stores.yml. The stores mapping is decoded via a
// yaml.Node so declaration order (and thus precedence) is preserved — plain
// map decoding loses it.
func parseRegistry(data []byte) (*Registry, error) {
	var doc struct {
		Schema int       `yaml:"schema"`
		Stores yaml.Node `yaml:"stores"`
	}
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", storesFile, err)
	}

	reg := &Registry{Stores: map[string]Store{}}
	content := doc.Stores.Content // mapping node: [key, val, key, val, ...]
	for i := 0; i+1 < len(content); i += 2 {
		name := content[i].Value
		var rs rawStore
		if err := content[i+1].Decode(&rs); err != nil {
			return nil, fmt.Errorf("store %q: %w", name, err)
		}
		kind := rs.Kind
		if kind == "" {
			kind = KindFolio // default: a registered store is a folio unless told otherwise
		}
		if kind != KindFolio && kind != KindExternal {
			return nil, fmt.Errorf("store %q: invalid kind %q (want %s|%s)", name, kind, KindFolio, KindExternal)
		}
		if rs.Path == "" {
			return nil, fmt.Errorf("store %q: missing path", name)
		}
		if _, dup := reg.Stores[name]; dup {
			return nil, fmt.Errorf("store %q: declared more than once", name)
		}
		reg.Stores[name] = Store{Name: name, Path: expandUser(rs.Path), Kind: kind}
		reg.Order = append(reg.Order, name)
	}
	return reg, nil
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
