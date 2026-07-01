package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRegistryAbsentFileEmpty(t *testing.T) {
	home := t.TempDir()
	reg, err := LoadRegistryFrom(home)
	if err != nil {
		t.Fatalf("LoadRegistryFrom: %v", err)
	}
	// No stores.yml ⇒ EMPTY registry. vault is NOT a registry store; it is
	// resolved intrinsically (see TestResolvePathIntrinsicVault).
	if len(reg.Order) != 0 {
		t.Fatalf("expected empty registry, got order %v", reg.Order)
	}
	if _, ok := reg.Lookup("vault"); ok {
		t.Error("vault must NOT be a registered store in the implicit default")
	}
	if len(reg.FolioStores()) != 0 {
		t.Error("empty registry should have no folio stores")
	}
}

func TestResolvePathIntrinsicVault(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("FOLIO_HOME", homeDir)
	reg, _ := LoadRegistryFrom(homeDir) // empty

	// vault: resolves to <home>/vault even with no registered vault store.
	got, err := ResolvePath("/proj", "vault:research/x.md", reg)
	if err != nil {
		t.Fatalf("intrinsic vault: %v", err)
	}
	if want := filepath.Join(homeDir, "vault", "research/x.md"); got != want {
		t.Errorf("vault resolve = %q, want %q", got, want)
	}

	// A genuinely unknown store-shaped prefix still fails loud.
	if _, err := ResolvePath("/proj", "bogus:foo.md", reg); err == nil {
		t.Error("expected error for unknown prefix with a real (empty) registry")
	}

	// A registered store named vault overrides the intrinsic path.
	override := &Registry{
		Stores: map[string]Store{"vault": {Name: "vault", Path: "/custom/vault", Kind: KindFolio}},
		Order:  []string{"vault"},
	}
	got, _ = ResolvePath("/proj", "vault:a.md", override)
	if got != "/custom/vault/a.md" {
		t.Errorf("registered vault override = %q, want /custom/vault/a.md", got)
	}
}

func TestParseRegistryOrderAndKinds(t *testing.T) {
	home := t.TempDir()
	yaml := `schema: 1
stores:
  vault: { path: ~/.folio/vault, kind: folio }
  work:  { path: /tmp/work-folio, kind: folio }
  radr:  { path: /tmp/guideline/radrs, kind: external }
`
	if err := os.WriteFile(filepath.Join(home, "stores.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	reg, err := LoadRegistryFrom(home)
	if err != nil {
		t.Fatalf("LoadRegistryFrom: %v", err)
	}

	wantOrder := []string{"vault", "work", "radr"}
	if len(reg.Order) != len(wantOrder) {
		t.Fatalf("order = %v, want %v", reg.Order, wantOrder)
	}
	for i, name := range wantOrder {
		if reg.Order[i] != name {
			t.Errorf("order[%d] = %q, want %q", i, reg.Order[i], name)
		}
	}

	folios := reg.FolioStores()
	if len(folios) != 2 || folios[0].Name != "vault" || folios[1].Name != "work" {
		t.Errorf("FolioStores = %v, want [vault work] in order", folios)
	}

	radr, _ := reg.Lookup("radr")
	if !radr.IsExternal() {
		t.Errorf("radr should be external")
	}

	home2, _ := os.UserHomeDir()
	if v, _ := reg.Lookup("vault"); v.Path != filepath.Join(home2, ".folio/vault") {
		t.Errorf("~/ expansion failed: %q", v.Path)
	}
}

func TestParseRegistryRejectsBadKind(t *testing.T) {
	home := t.TempDir()
	yaml := "schema: 1\nstores:\n  bad: { path: /x, kind: bogus }\n"
	os.WriteFile(filepath.Join(home, "stores.yml"), []byte(yaml), 0644)
	if _, err := LoadRegistryFrom(home); err == nil {
		t.Fatal("expected error for invalid kind, got nil")
	}
}

func TestParseRegistryCodeAndDot(t *testing.T) {
	home := t.TempDir()
	yaml := `schema: 3
stores:
  zp:  { path: ~/workspace/zenpayroll, kind: code, location: referenced, default_branch: main }
  dot: { path: ~/.dotfiles, kind: dot, location: referenced }
  kb:  { path: ~/.folio/kb }
`
	os.WriteFile(filepath.Join(home, "stores.yml"), []byte(yaml), 0644)
	reg, err := LoadRegistryFrom(home)
	if err != nil {
		t.Fatalf("LoadRegistryFrom: %v", err)
	}
	if zp, _ := reg.Lookup("zp"); zp.Kind != KindCode || !zp.IsReferenced() || zp.DefaultBranch != "main" {
		t.Errorf("zp = %+v, want code/referenced/main", zp)
	}
	if dot, _ := reg.Lookup("dot"); dot.Kind != KindDot {
		t.Errorf("dot kind = %q, want dot", dot.Kind)
	}
	// kb omits kind → safe default is external (read-only), NOT folio; location/default_branch → contained/main
	if kb, _ := reg.Lookup("kb"); kb.Kind != KindExternal || kb.Location != LocationContained || kb.DefaultBranch != "main" {
		t.Errorf("kb defaults = %q/%q/%q, want external/contained/main", kb.Kind, kb.Location, kb.DefaultBranch)
	}
}

func TestParseRegistryRejectsFutureSchema(t *testing.T) {
	home := t.TempDir()
	os.WriteFile(filepath.Join(home, "stores.yml"), []byte("schema: 99\nstores: {}\n"), 0644)
	if _, err := LoadRegistryFrom(home); err == nil {
		t.Fatal("expected error for unsupported schema, got nil")
	}
}

func TestResolvePathStoreRefs(t *testing.T) {
	reg := &Registry{
		Stores: map[string]Store{
			"vault": {Name: "vault", Path: "/home/.folio/vault", Kind: KindFolio},
			"work":  {Name: "work", Path: "/work-folio", Kind: KindFolio},
			"radr":  {Name: "radr", Path: "/guideline/radrs", Kind: KindExternal},
		},
		Order: []string{"vault", "work", "radr"},
	}

	tests := []struct {
		name     string
		path     string
		want     string
		wantErr  bool
		folioDir string
	}{
		{name: "work folio ref", path: "work:reference/design/x.md", want: "/work-folio/reference/design/x.md"},
		{name: "external ref", path: "radr:0042-foo.md", want: "/guideline/radrs/0042-foo.md"},
		{name: "vault ref", path: "vault:research/a.md", want: "/home/.folio/vault/research/a.md"},
		{name: "unknown registered-looking prefix errors", path: "bogus:foo.md", wantErr: true},
		{name: "path with colon but slash falls through", path: "reference/a:b.md", want: "/proj/reference/a:b.md", folioDir: "/proj"},
		{name: "plain relative path", path: "reference/x.md", want: "/proj/reference/x.md", folioDir: "/proj"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.folioDir
			if dir == "" {
				dir = "/proj"
			}
			got, err := ResolvePath(dir, tt.path, reg)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got %q", tt.path, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tt.path, err)
			}
			if got != tt.want {
				t.Errorf("ResolvePath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

// A nil registry must reproduce legacy behavior: only vault: is special-cased,
// everything else joins against folioDir. This is the back-compat safety net
// used when LoadRegistry errors during validation.
func TestResolvePathNilRegLegacy(t *testing.T) {
	got, err := ResolvePath("/proj", "reference/x.md", nil)
	if err != nil {
		t.Fatalf("nil reg plain path: %v", err)
	}
	if got != "/proj/reference/x.md" {
		t.Errorf("nil reg plain = %q", got)
	}
	// An unknown prefix under nil reg must NOT fail loud — it falls through
	// (no registry to consult), preserving pre-registry behavior.
	got, err = ResolvePath("/proj", "bogus:foo.md", nil)
	if err != nil {
		t.Fatalf("nil reg unknown prefix should not error: %v", err)
	}
	if got != "/proj/bogus:foo.md" {
		t.Errorf("nil reg unknown prefix = %q, want fall-through", got)
	}
}

func TestIsExternalStorePath(t *testing.T) {
	reg := &Registry{
		Stores: map[string]Store{
			"work": {Name: "work", Path: "/w", Kind: KindFolio},
			"radr": {Name: "radr", Path: "/r", Kind: KindExternal},
		},
		Order: []string{"work", "radr"},
	}
	if !IsExternalStorePath("radr:x.md", reg) {
		t.Error("radr:x.md should be external")
	}
	if IsExternalStorePath("work:x.md", reg) {
		t.Error("work:x.md should not be external (folio)")
	}
	if IsExternalStorePath("reference/x.md", reg) {
		t.Error("plain path should not be external")
	}
}
