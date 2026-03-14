package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

func TestRunNewNoArgs(t *testing.T) {
	code := runNew([]string{})
	if code != 1 {
		t.Errorf("expected exit code 1 for no args, got %d", code)
	}
}

func TestRunNewOneArg(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	code := runNew([]string{"--folio", yml, "spike"})
	if code != 1 {
		t.Errorf("expected exit code 1 for one arg, got %d", code)
	}
}

func TestRunNewInvalidType(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	code := runNew([]string{"--folio", yml, "invalid", "topic"})
	if code != 1 {
		t.Errorf("expected exit code 1 for invalid type, got %d", code)
	}
}

func TestRunNewSpikeCreatesFile(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	code := runNew([]string{"--folio", yml, "spike", "test-topic"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	// Check file exists in reference/spike/
	matches, _ := filepath.Glob(filepath.Join(dir, "reference", "spike", "*-test-topic.md"))
	if len(matches) != 1 {
		t.Fatalf("expected 1 file in reference/spike/, got %d", len(matches))
	}

	// Check content has spike template
	data, _ := os.ReadFile(matches[0])
	if !strings.Contains(string(data), "## Purpose") {
		t.Error("spike file missing Purpose section")
	}

	// Check folio.yml was updated
	ymlData, _ := os.ReadFile(yml)
	if !strings.Contains(string(ymlData), "reference/spike/") {
		t.Error("folio.yml missing source entry")
	}

	// Verify it still parses
	if _, err := config.Load(yml); err != nil {
		t.Errorf("folio.yml no longer parses: %v", err)
	}
}

func TestRunNewDesignCreatesFile(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	code := runNew([]string{"--folio", yml, "design", "my-feature"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	matches, _ := filepath.Glob(filepath.Join(dir, "reference", "design", "*-my-feature.md"))
	if len(matches) != 1 {
		t.Fatalf("expected 1 file in reference/design/, got %d", len(matches))
	}

	data, _ := os.ReadFile(matches[0])
	content := string(data)
	if !strings.Contains(content, "## Architecture") {
		t.Error("design file missing Architecture section")
	}
	if !strings.Contains(content, "## Divergence Decisions") {
		t.Error("design file missing Divergence Decisions section")
	}
}

func TestRunNewBriefCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	code := runNew([]string{"--folio", yml, "brief", "my-project"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	// Brief creates work/active/YYYY-MM-DD-<topic>/README.md
	matches, _ := filepath.Glob(filepath.Join(dir, "work", "active", "*-my-project", "README.md"))
	if len(matches) != 1 {
		t.Fatalf("expected 1 README.md in work/active/*-my-project/, got %d", len(matches))
	}

	data, _ := os.ReadFile(matches[0])
	if !strings.Contains(string(data), "## Objective") {
		t.Error("brief file missing Objective section")
	}
}

func TestRunNewDuplicateErrors(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	code := runNew([]string{"--folio", yml, "spike", "dup-test"})
	if code != 0 {
		t.Fatalf("first create expected exit code 0, got %d", code)
	}

	code = runNew([]string{"--folio", yml, "spike", "dup-test"})
	if code != 1 {
		t.Errorf("duplicate create expected exit code 1, got %d", code)
	}
}

func TestRunNewNoRegister(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	code := runNew([]string{"--folio", yml, "--no-register", "spike", "no-reg"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	// File should exist
	matches, _ := filepath.Glob(filepath.Join(dir, "reference", "spike", "*-no-reg.md"))
	if len(matches) != 1 {
		t.Fatal("expected file to be created")
	}

	// folio.yml should NOT have the entry
	data, _ := os.ReadFile(yml)
	if strings.Contains(string(data), "reference/spike/") {
		t.Error("folio.yml should not contain source entry with --no-register")
	}
}

func TestRunNewPreservesExistingSources(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	original := "schema: 1\nproject: \"Test\"\n\nsources:\n  - path: existing.md\n\ntargets: {}\nobservations: []\n"
	os.WriteFile(yml, []byte(original), 0644)

	code := runNew([]string{"--folio", yml, "spike", "my-spike"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	data, _ := os.ReadFile(yml)
	content := string(data)
	if !strings.Contains(content, "path: existing.md") {
		t.Error("existing source entry was lost")
	}
	if !strings.Contains(content, "reference/spike/") {
		t.Error("new source entry not added")
	}

	f, err := config.Load(yml)
	if err != nil {
		t.Fatalf("folio.yml no longer parses: %v", err)
	}
	if len(f.Sources) != 2 {
		t.Errorf("expected 2 sources, got %d", len(f.Sources))
	}
}

func TestRunNewNoteDeprecation(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	code := runNew([]string{"--folio", yml, "note", "topic"})
	if code != 1 {
		t.Errorf("expected exit code 1 for deprecated note type, got %d", code)
	}
}

func TestRunNewRetroColocated(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	// Create a work dir matching the topic
	workDir := filepath.Join(dir, "work", "active", "2026-01-01-mytopic")
	os.MkdirAll(workDir, 0755)
	os.WriteFile(filepath.Join(workDir, "README.md"), []byte("# Brief\n"), 0644)

	code := runNew([]string{"--folio", yml, "retro", "mytopic"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	// Verify file at colocated path
	retroPath := filepath.Join(workDir, "retro.md")
	if _, err := os.Stat(retroPath); err != nil {
		t.Fatalf("expected retro.md at colocated path %s: %v", retroPath, err)
	}

	data, _ := os.ReadFile(retroPath)
	if !strings.Contains(string(data), "## Context") {
		t.Error("retro file missing Context section")
	}

	// folio.yml source entry should contain work/active/
	ymlData, _ := os.ReadFile(yml)
	if !strings.Contains(string(ymlData), "work/active/") {
		t.Error("folio.yml source entry should reference colocated work path")
	}
}

func TestRunNewRetroStandalone(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	code := runNew([]string{"--folio", yml, "retro", "standalone-topic"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	matches, _ := filepath.Glob(filepath.Join(dir, "reference", "retro", "*-standalone-topic.md"))
	if len(matches) != 1 {
		t.Fatalf("expected 1 file in reference/retro/, got %d", len(matches))
	}
}

func TestRunNewDesignColocated(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	workDir := filepath.Join(dir, "work", "active", "2026-01-01-mydesign")
	os.MkdirAll(workDir, 0755)
	os.WriteFile(filepath.Join(workDir, "README.md"), []byte("# Brief\n"), 0644)

	code := runNew([]string{"--folio", yml, "design", "mydesign"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	designPath := filepath.Join(workDir, "design.md")
	if _, err := os.Stat(designPath); err != nil {
		t.Fatalf("expected design.md at colocated path %s: %v", designPath, err)
	}

	data, _ := os.ReadFile(designPath)
	if !strings.Contains(string(data), "## Problem") {
		t.Error("design file missing Problem section")
	}

	// folio.yml source entry should contain work/active/
	ymlData, _ := os.ReadFile(yml)
	if !strings.Contains(string(ymlData), "work/active/") {
		t.Error("folio.yml source entry should reference colocated work path")
	}
}

func TestRunNewDesignStandalone(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	code := runNew([]string{"--folio", yml, "design", "standalone-design"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	matches, _ := filepath.Glob(filepath.Join(dir, "reference", "design", "*-standalone-design.md"))
	if len(matches) != 1 {
		t.Fatalf("expected 1 file in reference/design/, got %d", len(matches))
	}
}
