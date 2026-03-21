package forest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFrontmatterMinimal(t *testing.T) {
	content := []byte("---\njira: BEN-123\n---\n# Content")
	fm, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatal(err)
	}
	if fm == nil {
		t.Fatal("expected frontmatter, got nil")
	}
	if fm.Jira != "BEN-123" {
		t.Errorf("expected BEN-123, got %q", fm.Jira)
	}
}

func TestParseFrontmatterFull(t *testing.T) {
	content := []byte("---\njira: BEN-456\nlabel: My Epic\ntype: Story\nsync: pull\norder: 5\n---\n# Content")
	fm, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatal(err)
	}
	if fm == nil {
		t.Fatal("expected frontmatter, got nil")
	}
	if fm.Jira != "BEN-456" {
		t.Errorf("jira: expected BEN-456, got %q", fm.Jira)
	}
	if fm.Label != "My Epic" {
		t.Errorf("label: expected My Epic, got %q", fm.Label)
	}
	if fm.Type != "Story" {
		t.Errorf("type: expected Story, got %q", fm.Type)
	}
	if fm.Sync != "pull" {
		t.Errorf("sync: expected pull, got %q", fm.Sync)
	}
	if fm.Order != 5 {
		t.Errorf("order: expected 5, got %d", fm.Order)
	}
}

func TestParseFrontmatterTBD(t *testing.T) {
	content := []byte("---\njira: TBD\ntype: Story\n---\n# New Feature")
	fm, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatal(err)
	}
	if fm == nil {
		t.Fatal("expected frontmatter, got nil")
	}
	if fm.Jira != "TBD" {
		t.Errorf("expected TBD, got %q", fm.Jira)
	}
	if fm.Type != "Story" {
		t.Errorf("expected Story, got %q", fm.Type)
	}
}

func TestParseFrontmatterNoFrontmatter(t *testing.T) {
	content := []byte("# Just a heading\n\nSome text")
	fm, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatal(err)
	}
	if fm != nil {
		t.Error("expected nil for file without frontmatter")
	}
}

func TestParseFrontmatterNoJira(t *testing.T) {
	content := []byte("---\ntitle: Notes\n---\n# Notes")
	fm, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatal(err)
	}
	if fm != nil {
		t.Error("expected nil for frontmatter without jira: field")
	}
}

func TestParseFrontmatterMalformed(t *testing.T) {
	content := []byte("---\n: bad yaml [[[}\n---\n# Content")
	_, err := ParseFrontmatter(content)
	if err == nil {
		t.Error("expected error for malformed YAML")
	}
}

func TestParseFrontmatterUnknownFields(t *testing.T) {
	content := []byte("---\njira: BEN-789\ncustom_field: hello\nanother: 42\n---\n# Content")
	fm, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatal(err)
	}
	if fm == nil {
		t.Fatal("expected frontmatter, got nil")
	}
	if fm.Jira != "BEN-789" {
		t.Errorf("expected BEN-789, got %q", fm.Jira)
	}
}

func TestParseFrontmatterQuotedTBD(t *testing.T) {
	content := []byte("---\njira: \"TBD\"\ntype: Task\n---\n# Content")
	fm, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatal(err)
	}
	if fm == nil {
		t.Fatal("expected frontmatter, got nil")
	}
	if fm.Jira != "TBD" {
		t.Errorf("expected TBD, got %q", fm.Jira)
	}
}

func TestParseForestFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "forest.yml")
	content := `schema: 1

defaults:
  sync: push
  type: Story
  project: BEN
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	f, err := ParseForestFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if f.Schema != 1 {
		t.Errorf("schema: expected 1, got %d", f.Schema)
	}
	if f.Defaults.Sync != "push" {
		t.Errorf("sync: expected push, got %q", f.Defaults.Sync)
	}
	if f.Defaults.Type != "Story" {
		t.Errorf("type: expected Story, got %q", f.Defaults.Type)
	}
	if f.Defaults.Project != "BEN" {
		t.Errorf("project: expected BEN, got %q", f.Defaults.Project)
	}
	if f.Dir != dir {
		t.Errorf("dir: expected %q, got %q", dir, f.Dir)
	}
}

func TestDeriveLabelFromFrontmatter(t *testing.T) {
	fm := &Frontmatter{Label: "My Label"}
	label := DeriveLabel(fm, []byte("# Heading"), "path/file.md")
	if label != "My Label" {
		t.Errorf("expected My Label, got %q", label)
	}
}

func TestDeriveLabelFromHeading(t *testing.T) {
	fm := &Frontmatter{Jira: "BEN-1"}
	content := []byte("---\njira: BEN-1\n---\n# My Heading\n\nText")
	label := DeriveLabel(fm, content, "path/file.md")
	if label != "My Heading" {
		t.Errorf("expected My Heading, got %q", label)
	}
}

func TestDeriveLabelFromFilename(t *testing.T) {
	fm := &Frontmatter{Jira: "BEN-1"}
	content := []byte("---\njira: BEN-1\n---\nNo heading here")
	label := DeriveLabel(fm, content, "path/syncer-migration.md")
	if label != "syncer-migration" {
		t.Errorf("expected syncer-migration, got %q", label)
	}
}

func TestDeriveLabelFromREADMEUsesDir(t *testing.T) {
	fm := &Frontmatter{Jira: "BEN-1"}
	content := []byte("---\njira: BEN-1\n---\nNo heading")
	label := DeriveLabel(fm, content, "epics/README.md")
	if label != "epics" {
		t.Errorf("expected epics, got %q", label)
	}
}

func TestFirstHeadingSkipsFrontmatter(t *testing.T) {
	content := []byte("---\ntitle: test\n---\n# Real Heading")
	heading := firstHeading(content)
	if heading != "Real Heading" {
		t.Errorf("expected Real Heading, got %q", heading)
	}
}

func TestFirstHeadingNoHeading(t *testing.T) {
	content := []byte("Just some text\nNo headings here")
	heading := firstHeading(content)
	if heading != "" {
		t.Errorf("expected empty, got %q", heading)
	}
}
