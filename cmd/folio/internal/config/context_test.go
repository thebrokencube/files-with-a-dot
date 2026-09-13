package config

import (
	"os"
	"path/filepath"
	"testing"
)

func contextRegistry(stores ...Store) *Registry {
	reg := &Registry{Stores: make(map[string]Store)}
	for _, store := range stores {
		reg.Stores[store.Name] = store
		reg.Order = append(reg.Order, store.Name)
	}
	return reg
}

func canonicalForTest(t *testing.T, path string) string {
	t.Helper()
	got, err := canonicalPath(path)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestContextResolvePathVaultPrecedence(t *testing.T) {
	umbrella := canonicalForTest(t, t.TempDir())
	owning := canonicalForTest(t, filepath.Join(umbrella, "work"))
	workRoot := canonicalForTest(t, filepath.Join(umbrella, "fallback"))
	literal := canonicalForTest(t, filepath.Join(umbrella, "literal-vault"))
	reg := contextRegistry(
		Store{Name: "work", Path: owning, Kind: KindFolio},
		Store{Name: "vault", Path: literal, Kind: KindFolio},
	)

	ctx := Context{
		Umbrella:  umbrella,
		Registry:  reg,
		Store:     reg.Stores["work"],
		WorkRoot:  workRoot,
		FolioPath: filepath.Join(owning, "active", "project", "folio.yml"),
	}
	got, err := ctx.ResolvePath("vault:notes.md")
	if err != nil {
		t.Fatal(err)
	}
	if want := canonicalForTest(t, filepath.Join(literal, "notes.md")); got != want {
		t.Fatalf("registered vault = %q, want %q", got, want)
	}

	delete(reg.Stores, "vault")
	got, err = ctx.ResolvePath("vault:notes.md")
	if err != nil {
		t.Fatal(err)
	}
	if want := canonicalForTest(t, filepath.Join(owning, "vault", "notes.md")); got != want {
		t.Fatalf("owning-store vault = %q, want %q", got, want)
	}

	ctx.FolioPath = ""
	got, err = ctx.ResolvePath("vault:notes.md")
	if err != nil {
		t.Fatal(err)
	}
	if want := canonicalForTest(t, filepath.Join(umbrella, "vault", "notes.md")); got != want {
		t.Fatalf("umbrella vault = %q, want %q", got, want)
	}

	ctx.Umbrella = ""
	got, err = ctx.ResolvePath("vault:notes.md")
	if err != nil {
		t.Fatal(err)
	}
	if want := canonicalForTest(t, filepath.Join(workRoot, "vault", "notes.md")); got != want {
		t.Fatalf("work-root vault = %q, want %q", got, want)
	}
}
func TestContextResolvePathWorkspaceVaultUsesStore(t *testing.T) {
	storeRoot := canonicalForTest(t, t.TempDir())
	workspaceRoot := canonicalForTest(t, t.TempDir())
	reg := contextRegistry(Store{Name: "work", Path: storeRoot, Kind: KindFolio})
	ctx := Context{
		Registry:  reg,
		Store:     reg.Stores["work"],
		WorkRoot:  workspaceRoot,
		FolioPath: filepath.Join(workspaceRoot, "active", "project", "folio.yml"),
	}

	got, err := ctx.ResolvePath("vault:notes.md")
	if err != nil {
		t.Fatal(err)
	}
	want := canonicalForTest(t, filepath.Join(storeRoot, "vault", "notes.md"))
	if got != want {
		t.Fatalf("workspace vault = %q, want owning store %q", got, want)
	}
}

func TestContextResolvePathStoreOnlyVault(t *testing.T) {
	workRoot := canonicalForTest(t, t.TempDir())
	ctx := Context{WorkRoot: workRoot, Registry: contextRegistry()}

	got, err := ctx.ResolvePath("vault:research.md")
	if err != nil {
		t.Fatal(err)
	}
	if want := canonicalForTest(t, filepath.Join(workRoot, "vault", "research.md")); got != want {
		t.Fatalf("store-only vault = %q, want %q", got, want)
	}
	if _, err := ctx.ResolvePath("reference/research.md"); err == nil {
		t.Fatal("store-only project-local reference should fail")
	}
}

func TestContextResolvePathRegisteredStore(t *testing.T) {
	root := canonicalForTest(t, t.TempDir())
	reg := contextRegistry(Store{Name: "other", Path: root, Kind: KindFolio})
	ctx := Context{Registry: reg, WorkRoot: canonicalForTest(t, t.TempDir())}

	got, err := ctx.ResolvePath("other:reference/design.md")
	if err != nil {
		t.Fatal(err)
	}
	if want := canonicalForTest(t, filepath.Join(root, "reference", "design.md")); got != want {
		t.Fatalf("registered store = %q, want %q", got, want)
	}
}

func TestContextForProjectPreservesRoots(t *testing.T) {
	umbrella := canonicalForTest(t, t.TempDir())
	storeRoot := canonicalForTest(t, filepath.Join(umbrella, "store"))
	workRoot := canonicalForTest(t, filepath.Join(umbrella, "workspace"))
	reg := contextRegistry(Store{Name: "work", Path: storeRoot, Kind: KindFolio})
	ctx := Context{
		Umbrella: umbrella,
		Registry: reg,
		Store:    reg.Stores["work"],
		WorkRoot: workRoot,
	}

	project := filepath.Join(storeRoot, "active", "project", "folio.yml")
	got, err := ctx.ForProject(project)
	if err != nil {
		t.Fatal(err)
	}
	if got.Umbrella != ctx.Umbrella || got.WorkRoot != ctx.WorkRoot || got.Store != ctx.Store || got.Registry != ctx.Registry {
		t.Fatalf("ForProject changed roots: got %+v from %+v", got, ctx)
	}
	if got.FolioPath != canonicalForTest(t, project) {
		t.Fatalf("FolioPath = %q, want %q", got.FolioPath, canonicalForTest(t, project))
	}
	if got.FolioDir() != canonicalForTest(t, filepath.Dir(project)) {
		t.Fatalf("FolioDir = %q, want %q", got.FolioDir(), canonicalForTest(t, filepath.Dir(project)))
	}
}

func TestContextForProjectCanonicalizesPlannedLeaf(t *testing.T) {
	realRoot := canonicalForTest(t, t.TempDir())
	linkParent := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(realRoot, linkParent); err != nil {
		t.Fatal(err)
	}

	ctx := Context{WorkRoot: realRoot}
	project := filepath.Join(linkParent, "missing", "folio.yml")
	got, err := ctx.ForProject(project)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(realRoot, "missing", "folio.yml")
	if got.FolioPath != want {
		t.Fatalf("canonical planned path = %q, want %q", got.FolioPath, want)
	}
	if got.FolioDir() != filepath.Dir(want) {
		t.Fatalf("canonical planned dir = %q, want %q", got.FolioDir(), filepath.Dir(want))
	}
}
