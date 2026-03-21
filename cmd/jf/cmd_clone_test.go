package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"[SRM] Unify Compliance Engine", "unify-compliance-engine"},
		{"Simple Title", "simple-title"},
		{"  spaces  ", "spaces"},
		{"Hello World!", "hello-world"},
	}
	for _, tt := range tests {
		got := slugify(tt.input)
		if got != tt.want {
			t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSlugifyEmpty(t *testing.T) {
	if slugify("") != "forest" {
		t.Error("expected 'forest' for empty input")
	}
	if slugify("   ") != "forest" {
		t.Error("expected 'forest' for whitespace input")
	}
}

func TestCountNodes(t *testing.T) {
	tree := &cloneNode{
		Key: "ROOT",
		Children: []*cloneNode{
			{Key: "A"},
			{Key: "B", Children: []*cloneNode{{Key: "C"}}},
		},
	}
	if countNodes(tree) != 4 {
		t.Errorf("expected 4, got %d", countNodes(tree))
	}
}

func TestScaffoldTree(t *testing.T) {
	dir := t.TempDir()
	tree := &cloneNode{
		Key:     "BEN-1",
		Summary: "Root Epic",
		Children: []*cloneNode{
			{Key: "BEN-2", Summary: "Leaf Story"},
		},
	}

	if err := scaffoldTree(dir, tree, ""); err != nil {
		t.Fatal(err)
	}

	// Root README.md should exist
	readme, err := os.ReadFile(filepath.Join(dir, "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(readme), "jira: BEN-1") {
		t.Error("expected jira key in root README")
	}

	// Leaf file should exist
	leaf, err := os.ReadFile(filepath.Join(dir, "BEN-2.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(leaf), "jira: BEN-2") {
		t.Error("expected jira key in leaf file")
	}
}

func TestGenerateForestYAML(t *testing.T) {
	dir := t.TempDir()
	root := &cloneNode{Key: "BEN-123", Summary: "Test"}

	if err := generateForestYAML(dir, root); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "forest.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "project: BEN") {
		t.Errorf("expected project BEN in forest.yml, got %s", string(data))
	}
}
