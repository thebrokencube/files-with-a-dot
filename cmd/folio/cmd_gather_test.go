package main

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

func TestRunGatherNoArgs(t *testing.T) {
	code := runGather([]string{})
	if code != 1 {
		t.Errorf("expected exit code 1 for no args, got %d", code)
	}
}

func TestRunGatherInvalidURL(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	code := runGather([]string{"--folio", yml, "not-a-url"})
	if code != 1 {
		t.Errorf("expected exit code 1 for invalid URL, got %d", code)
	}
}

func TestRunGatherReadFlag(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	code := runGather([]string{"--folio", yml, "--read", "https://example.com/doc"})
	if code != 1 {
		t.Errorf("expected exit code 1 for --read flag, got %d", code)
	}
}

func TestRunGatherURLOnly(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\ntargets: {}\ntasks: []\npending: []\n"), 0644)

	code := runGather([]string{"--folio", yml, "https://example.com/guide.html"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	// Verify folio.yml was modified
	data, _ := os.ReadFile(yml)
	content := string(data)
	if !strings.Contains(content, "external: web") {
		t.Error("expected 'external: web' in folio.yml")
	}
	if !strings.Contains(content, "https://example.com/guide.html") {
		t.Error("expected URL in folio.yml")
	}
	if !strings.Contains(content, "not yet materialized") {
		t.Error("expected 'not yet materialized' note in folio.yml")
	}

	// Verify it still parses
	if _, err := config.Load(yml); err != nil {
		t.Errorf("folio.yml no longer parses: %v", err)
	}
}

func TestRunGatherMaterialize(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\ntargets: {}\ntasks: []\npending: []\n"), 0644)

	code := runGather([]string{"--folio", yml, "--materialize", "https://example.com/api-spec"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	// Verify reference file was created
	refPath := filepath.Join(dir, "reference", "api-spec.md")
	if _, err := os.Stat(refPath); os.IsNotExist(err) {
		t.Error("reference file was not created")
	}

	// Verify folio.yml has path entry
	data, _ := os.ReadFile(yml)
	content := string(data)
	if !strings.Contains(content, "path: reference/api-spec.md") {
		t.Error("expected path entry in folio.yml")
	}
	if !strings.Contains(content, "derived_from:") {
		t.Error("expected derived_from entry in folio.yml")
	}

	// Verify it still parses
	if _, err := config.Load(yml); err != nil {
		t.Errorf("folio.yml no longer parses: %v", err)
	}
}

func TestRunGatherMaterializeWithName(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\ntargets: {}\ntasks: []\npending: []\n"), 0644)

	code := runGather([]string{"--folio", yml, "--materialize", "--name", "my-ref", "https://example.com/page"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	refPath := filepath.Join(dir, "reference", "my-ref.md")
	if _, err := os.Stat(refPath); os.IsNotExist(err) {
		t.Error("reference file was not created with custom name")
	}

	data, _ := os.ReadFile(yml)
	if !strings.Contains(string(data), "path: reference/my-ref.md") {
		t.Error("expected custom name path in folio.yml")
	}
}

func TestRunGatherPreservesOtherFields(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")

	original := `schema: 1
project: "Complex"

sources:
  - path: existing.md

targets:
  my-target:
    transform: distill
    sources:
      - path: existing.md
    outputs:
      - path: compiled/out.md

cross_references:
  - fact: "Some fact"
    source_of_truth: "existing.md"
    also_appears_in: ["other.md"]

tasks:
  - "Task one"

pending:
  - "Pending one"
`
	os.WriteFile(yml, []byte(original), 0644)

	code := runGather([]string{"--folio", yml, "https://example.com/new-source"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	data, _ := os.ReadFile(yml)
	content := string(data)

	// Verify source was added
	if !strings.Contains(content, "https://example.com/new-source") {
		t.Error("expected new URL in folio.yml")
	}

	// Verify other sections are intact
	checks := []string{
		`project: "Complex"`,
		"path: existing.md",
		"my-target:",
		"transform: distill",
		`fact: "Some fact"`,
		`"Task one"`,
		`"Pending one"`,
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("expected %q to be preserved in folio.yml", check)
		}
	}

	// Verify it still parses
	f, err := config.Load(yml)
	if err != nil {
		t.Fatalf("folio.yml no longer parses: %v", err)
	}

	// Sources should have grown by 1
	if len(f.Sources) != 2 {
		t.Errorf("expected 2 sources, got %d", len(f.Sources))
	}
}

func TestDeriveNameFromURL(t *testing.T) {
	tests := []struct {
		url  string
		want string
	}{
		{"https://example.com/guide.html", "guide"},
		{"https://example.com/api-spec", "api-spec"},
		{"https://docs.example.com/v2/getting-started.md", "getting-started"},
		{"https://example.com/", "example-com"},
		{"https://example.com/path/to/deep/page", "page"},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			u, _ := url.Parse(tt.url)
			got := deriveNameFromURL(u)
			if got != tt.want {
				t.Errorf("deriveNameFromURL(%s) = %s, want %s", tt.url, got, tt.want)
			}
		})
	}
}
