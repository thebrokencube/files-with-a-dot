package config

import (
	"os"
	"path/filepath"
	"testing"
)

// writeRegistry writes a stores.yml into home and loads it.
func writeRegistry(t *testing.T, home, yaml string) *Registry {
	t.Helper()
	if err := os.WriteFile(filepath.Join(home, "stores.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	reg, err := LoadRegistryFrom(home)
	if err != nil {
		t.Fatalf("LoadRegistryFrom: %v", err)
	}
	return reg
}

func TestParseRegistryDefaultAndSchema2(t *testing.T) {
	home := t.TempDir()
	reg := writeRegistry(t, home, `schema: 2
default: work
stores:
  work:  { path: /tmp/work-folio, kind: folio }
  vault: { path: /tmp/vault, kind: folio }
`)
	if reg.Default != "work" {
		t.Errorf("Default = %q, want work", reg.Default)
	}
	if reg.isImplicitDefault() {
		t.Error("a parsed registry must not be the implicit sentinel")
	}
}

func TestImplicitDefaultFlagVsExplicitEmpty(t *testing.T) {
	// File-absent → implicit sentinel.
	home := t.TempDir()
	reg, err := LoadRegistryFrom(home)
	if err != nil {
		t.Fatal(err)
	}
	if !reg.isImplicitDefault() {
		t.Error("absent stores.yml must yield the implicit sentinel")
	}

	// An explicit, content-empty stores.yml is byte-identical in contents but
	// must NOT read as implicit — the flag, not the contents, decides.
	home2 := t.TempDir()
	reg2 := writeRegistry(t, home2, "schema: 2\nstores: {}\n")
	if reg2.isImplicitDefault() {
		t.Error("an explicit empty registry must not be the implicit sentinel")
	}
}

func TestActiveStoreImplicitFallsBack(t *testing.T) {
	reg, _ := LoadRegistryFrom(t.TempDir()) // implicit
	s, ok, err := ActiveStore(reg)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Errorf("implicit registry must yield ok=false, got store %q", s.Name)
	}
}

func TestActiveStoreDefault(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	reg := writeRegistry(t, home, "schema: 2\ndefault: work\nstores:\n  work: { path: "+work+", kind: folio }\n")

	// cwd is neither store, so default wins.
	t.Chdir(home)
	s, ok, err := ActiveStore(reg)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || s.Name != "work" {
		t.Errorf("default resolution = (%q, %v), want (work, true)", s.Name, ok)
	}
}

func TestActiveStoreCwdOverridesDefault(t *testing.T) {
	home := t.TempDir()
	work := t.TempDir()
	vault := t.TempDir()
	reg := writeRegistry(t, home, "schema: 2\ndefault: work\nstores:\n"+
		"  work:  { path: "+work+", kind: folio }\n"+
		"  vault: { path: "+vault+", kind: folio }\n")

	// cwd inside vault must override the work default.
	sub := filepath.Join(vault, "active", "proj")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(sub)
	s, ok, err := ActiveStore(reg)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || s.Name != "vault" {
		t.Errorf("cwd-in-store = (%q, %v), want (vault, true)", s.Name, ok)
	}
}

func TestActiveStoreUnknownDefaultErrors(t *testing.T) {
	home := t.TempDir()
	reg := writeRegistry(t, home, "schema: 2\ndefault: ghost\nstores:\n  work: { path: /tmp/w, kind: folio }\n")
	t.Chdir(home)
	if _, _, err := ActiveStore(reg); err == nil {
		t.Fatal("default naming an unregistered store must error")
	}
}

func TestActiveStoreNoDefaultNoCwdFallsBack(t *testing.T) {
	home := t.TempDir()
	reg := writeRegistry(t, home, "schema: 2\nstores:\n  work: { path: /tmp/w, kind: folio }\n")
	t.Chdir(home) // home is the umbrella, not a store
	_, ok, err := ActiveStore(reg)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("no default + cwd outside any store must yield ok=false")
	}
}

func TestStoreContainingLongestMatchWins(t *testing.T) {
	outer := filepath.Clean("/tmp/outer")
	inner := filepath.Join(outer, "nested")
	reg := &Registry{
		Stores: map[string]Store{
			"outer": {Name: "outer", Path: outer, Kind: KindFolio},
			"inner": {Name: "inner", Path: inner, Kind: KindFolio},
		},
		Order: []string{"outer", "inner"},
	}
	s, ok := storeContaining(filepath.Join(inner, "a", "b"), reg)
	if !ok || s.Name != "inner" {
		t.Errorf("nested dir = (%q, %v), want (inner, true)", s.Name, ok)
	}
	// A sibling that only shares a string prefix must NOT match.
	if _, ok := storeContaining("/tmp/outer-sibling/x", reg); ok {
		t.Error("string-prefix sibling must not match a store root")
	}
}

// vault: resolves relative to the active folio store (the one containing
// folioDir) in container mode, not the umbrella home.
func TestResolveVaultRelativeToActiveStore(t *testing.T) {
	work := filepath.Clean("/tmp/work-folio")
	reg := &Registry{
		Stores: map[string]Store{"work": {Name: "work", Path: work, Kind: KindFolio}},
		Order:  []string{"work"},
	}
	folioDir := filepath.Join(work, "active", "proj")
	got, err := ResolvePath(folioDir, "vault:research/x.md", reg)
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(work, "vault", "research/x.md"); got != want {
		t.Errorf("vault resolve = %q, want %q", got, want)
	}
}
