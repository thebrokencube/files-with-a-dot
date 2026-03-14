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
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\ntargets: {}\nobservations: []\n"), 0644)

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

func TestRunGatherMaterializeRequiresType(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\ntargets: {}\nobservations: []\n"), 0644)

	code := runGather([]string{"--folio", yml, "--materialize", "https://example.com/api-spec"})
	if code != 1 {
		t.Errorf("expected exit code 1 for --materialize without --type, got %d", code)
	}
}

func TestRunGatherMaterialize(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\ntargets: {}\nobservations: []\n"), 0644)

	code := runGather([]string{"--folio", yml, "--materialize", "--type", "survey", "https://example.com/api-spec"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	// Verify reference file was created in type directory with date prefix
	matches, _ := filepath.Glob(filepath.Join(dir, "reference", "survey", "*-api-spec.md"))
	if len(matches) != 1 {
		t.Fatalf("expected 1 file in reference/survey/, got %d", len(matches))
	}

	// Verify folio.yml has typed path entry
	data, _ := os.ReadFile(yml)
	content := string(data)
	if !strings.Contains(content, "path: reference/survey/") {
		t.Error("expected typed path entry in folio.yml")
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
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\ntargets: {}\nobservations: []\n"), 0644)

	code := runGather([]string{"--folio", yml, "--materialize", "--type", "spike", "--name", "my-ref", "https://example.com/page"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	matches, _ := filepath.Glob(filepath.Join(dir, "reference", "spike", "*-my-ref.md"))
	if len(matches) != 1 {
		t.Fatal("reference file was not created with custom name in type directory")
	}

	data, _ := os.ReadFile(yml)
	if !strings.Contains(string(data), "path: reference/spike/") {
		t.Error("expected typed path in folio.yml")
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
    how: "Test"
    sources:
      - path: existing.md
    outputs:
      - path: compiled/out.md

cross_references:
  - fact: "Some fact"
    source_of_truth: "existing.md"
    also_appears_in: ["other.md"]

observations:
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
		`how: "Test"`,
		`fact: "Some fact"`,
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
