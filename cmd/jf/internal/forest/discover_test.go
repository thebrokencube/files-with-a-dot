package forest

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func writeForestYml(t *testing.T, dir string) *Forest {
	t.Helper()
	path := filepath.Join(dir, "forest.yml")
	writeFile(t, path, "schema: 1\ndefaults:\n  sync: push\n  type: Story\n  project: BEN\n")
	f, err := ParseForestFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestFindForestFromSubdir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "forest.yml"), "schema: 1\ndefaults:\n  sync: push\n")
	sub := filepath.Join(dir, "epics", "deep")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}

	f, err := FindForest(sub)
	if err != nil {
		t.Fatal(err)
	}
	if f == nil {
		t.Fatal("expected to find forest.yml")
	}
	if f.Dir != dir {
		t.Errorf("expected dir %q, got %q", dir, f.Dir)
	}
}

func TestFindForestNotFound(t *testing.T) {
	dir := t.TempDir()
	f, err := FindForest(dir)
	if err != nil {
		t.Fatal(err)
	}
	if f != nil {
		t.Error("expected nil when no forest.yml exists")
	}
}

func TestDiscoverFlatDirectory(t *testing.T) {
	dir := t.TempDir()
	forest := writeForestYml(t, dir)

	writeFile(t, filepath.Join(dir, "one.md"), "---\njira: BEN-1\n---\n# One")
	writeFile(t, filepath.Join(dir, "two.md"), "---\njira: BEN-2\n---\n# Two")
	writeFile(t, filepath.Join(dir, "three.md"), "---\njira: BEN-3\n---\n# Three")

	roots, err := Discover(forest)
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != 3 {
		t.Fatalf("expected 3 roots, got %d", len(roots))
	}
	// Should be sorted alphabetically by label
	if roots[0].Label != "One" {
		t.Errorf("expected One, got %q", roots[0].Label)
	}
}

func TestDiscoverNestedWithREADME(t *testing.T) {
	dir := t.TempDir()
	forest := writeForestYml(t, dir)

	writeFile(t, filepath.Join(dir, "epics", "README.md"), "---\njira: BEN-100\n---\n# Epics")
	writeFile(t, filepath.Join(dir, "epics", "auth.md"), "---\njira: BEN-101\n---\n# Auth Epic")
	writeFile(t, filepath.Join(dir, "epics", "billing.md"), "---\njira: BEN-102\n---\n# Billing Epic")

	roots, err := Discover(forest)
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	root := roots[0]
	if root.Key != "BEN-100" {
		t.Errorf("root key: expected BEN-100, got %q", root.Key)
	}
	if len(root.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(root.Children))
	}
	if root.Children[0].Key != "BEN-101" {
		t.Errorf("first child: expected BEN-101, got %q", root.Children[0].Key)
	}
	// Children should have parent set
	if root.Children[0].Parent != root {
		t.Error("child parent pointer not set")
	}
}

func TestDiscoverPassthroughDirectory(t *testing.T) {
	dir := t.TempDir()
	forest := writeForestYml(t, dir)

	// Root README
	writeFile(t, filepath.Join(dir, "README.md"), "---\njira: BEN-1\n---\n# Root")
	// passthrough/ has no README.md with jira:
	writeFile(t, filepath.Join(dir, "passthrough", "child.md"), "---\njira: BEN-2\n---\n# Child")

	roots, err := Discover(forest)
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	// Child should attach to root (pass-through)
	if len(roots[0].Children) != 1 {
		t.Fatalf("expected 1 child on root, got %d", len(roots[0].Children))
	}
	if roots[0].Children[0].Key != "BEN-2" {
		t.Errorf("expected BEN-2, got %q", roots[0].Children[0].Key)
	}
}

func TestDiscoverMixedJiraAndNonJira(t *testing.T) {
	dir := t.TempDir()
	forest := writeForestYml(t, dir)

	writeFile(t, filepath.Join(dir, "tracked.md"), "---\njira: BEN-1\n---\n# Tracked")
	writeFile(t, filepath.Join(dir, "notes.md"), "# Just notes\n\nNo frontmatter")
	writeFile(t, filepath.Join(dir, "other.md"), "---\ntitle: Not Jira\n---\n# Other")

	roots, err := Discover(forest)
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != 1 {
		t.Fatalf("expected 1 root (only jira: files), got %d", len(roots))
	}
	if roots[0].Key != "BEN-1" {
		t.Errorf("expected BEN-1, got %q", roots[0].Key)
	}
}

func TestDiscoverMultiRoot(t *testing.T) {
	dir := t.TempDir()
	forest := writeForestYml(t, dir)

	writeFile(t, filepath.Join(dir, "alpha", "README.md"), "---\njira: BEN-1\n---\n# Alpha")
	writeFile(t, filepath.Join(dir, "beta", "README.md"), "---\njira: BEN-2\n---\n# Beta")

	roots, err := Discover(forest)
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(roots))
	}
}

func TestDiscoverDefaultsApplied(t *testing.T) {
	dir := t.TempDir()
	forest := writeForestYml(t, dir) // defaults: sync=push, type=Story

	writeFile(t, filepath.Join(dir, "bare.md"), "---\njira: BEN-1\n---\n# Bare")

	roots, err := Discover(forest)
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != 1 {
		t.Fatal("expected 1 root")
	}
	if roots[0].Sync != "push" {
		t.Errorf("sync: expected push (from defaults), got %q", roots[0].Sync)
	}
	if roots[0].Type != "Story" {
		t.Errorf("type: expected Story (from defaults), got %q", roots[0].Type)
	}
}

func TestDiscoverFrontmatterOverridesDefaults(t *testing.T) {
	dir := t.TempDir()
	forest := writeForestYml(t, dir) // defaults: sync=push, type=Story

	writeFile(t, filepath.Join(dir, "override.md"), "---\njira: BEN-1\nsync: pull\ntype: Epic\n---\n# Override")

	roots, err := Discover(forest)
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != 1 {
		t.Fatal("expected 1 root")
	}
	if roots[0].Sync != "pull" {
		t.Errorf("sync: expected pull (from frontmatter), got %q", roots[0].Sync)
	}
	if roots[0].Type != "Epic" {
		t.Errorf("type: expected Epic (from frontmatter), got %q", roots[0].Type)
	}
}

func TestDiscoverDeepNesting(t *testing.T) {
	dir := t.TempDir()
	forest := writeForestYml(t, dir)

	writeFile(t, filepath.Join(dir, "a", "README.md"), "---\njira: BEN-1\n---\n# A")
	writeFile(t, filepath.Join(dir, "a", "b", "README.md"), "---\njira: BEN-2\n---\n# B")
	writeFile(t, filepath.Join(dir, "a", "b", "leaf.md"), "---\njira: BEN-3\n---\n# Leaf")

	roots, err := Discover(forest)
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	if roots[0].Key != "BEN-1" {
		t.Errorf("root: expected BEN-1, got %q", roots[0].Key)
	}
	if len(roots[0].Children) != 1 {
		t.Fatalf("expected 1 child of root, got %d", len(roots[0].Children))
	}
	b := roots[0].Children[0]
	if b.Key != "BEN-2" {
		t.Errorf("mid: expected BEN-2, got %q", b.Key)
	}
	if len(b.Children) != 1 {
		t.Fatalf("expected 1 child of B, got %d", len(b.Children))
	}
	if b.Children[0].Key != "BEN-3" {
		t.Errorf("leaf: expected BEN-3, got %q", b.Children[0].Key)
	}
}

func TestDiscoverOrderSort(t *testing.T) {
	dir := t.TempDir()
	forest := writeForestYml(t, dir)

	writeFile(t, filepath.Join(dir, "README.md"), "---\njira: BEN-1\n---\n# Root")
	writeFile(t, filepath.Join(dir, "z-last.md"), "---\njira: BEN-2\norder: 1\n---\n# Z Last")
	writeFile(t, filepath.Join(dir, "a-first.md"), "---\njira: BEN-3\norder: 2\n---\n# A First")

	roots, err := Discover(forest)
	if err != nil {
		t.Fatal(err)
	}
	root := roots[0]
	if len(root.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(root.Children))
	}
	// order 1 before order 2, regardless of label
	if root.Children[0].Key != "BEN-2" {
		t.Errorf("first child: expected BEN-2 (order 1), got %q", root.Children[0].Key)
	}
	if root.Children[1].Key != "BEN-3" {
		t.Errorf("second child: expected BEN-3 (order 2), got %q", root.Children[1].Key)
	}
}

func TestDiscoverSkipsHiddenDirs(t *testing.T) {
	dir := t.TempDir()
	forest := writeForestYml(t, dir)

	writeFile(t, filepath.Join(dir, "visible.md"), "---\njira: BEN-1\n---\n# Visible")
	writeFile(t, filepath.Join(dir, ".hidden", "secret.md"), "---\njira: BEN-2\n---\n# Secret")

	roots, err := Discover(forest)
	if err != nil {
		t.Fatal(err)
	}
	if len(roots) != 1 {
		t.Fatalf("expected 1 root (hidden dir skipped), got %d", len(roots))
	}
	if roots[0].Key != "BEN-1" {
		t.Errorf("expected BEN-1, got %q", roots[0].Key)
	}
}

func TestDiscoverNilForest(t *testing.T) {
	_, err := Discover(nil)
	if err == nil {
		t.Error("expected error for nil forest")
	}
}
