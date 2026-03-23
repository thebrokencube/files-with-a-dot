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

func TestRunNewDesignCreatesWorkDir(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	code := runNew([]string{"--folio", yml, "design", "my-feature"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	// Design should auto-create work dir and colocate inside it
	matches, _ := filepath.Glob(filepath.Join(dir, "work", "active", "*-my-feature", "reference", "design", "*-my-feature.md"))
	if len(matches) != 1 {
		t.Fatalf("expected 1 file in work/active/*-my-feature/reference/design/, got %d", len(matches))
	}

	data, _ := os.ReadFile(matches[0])
	content := string(data)
	if !strings.Contains(content, "## Architecture") {
		t.Error("design file missing Architecture section")
	}
	if !strings.Contains(content, "## Divergence Decisions") {
		t.Error("design file missing Divergence Decisions section")
	}

	// folio.yml source entry should reference the work dir path
	ymlData, _ := os.ReadFile(yml)
	if !strings.Contains(string(ymlData), "work/active/") {
		t.Error("folio.yml source entry should reference work/active/ path")
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

	// Verify file at nested colocated path (not flat retro.md)
	matches, _ := filepath.Glob(filepath.Join(workDir, "reference", "retro", "*-mytopic.md"))
	if len(matches) != 1 {
		t.Fatalf("expected 1 retro file in work dir reference/retro/, got %d", len(matches))
	}

	data, _ := os.ReadFile(matches[0])
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

func TestRunNewDesignColocatedExistingWorkDir(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	// Pre-existing work dir (e.g., created by a prior plan)
	workDir := filepath.Join(dir, "work", "active", "2026-01-01-mydesign")
	os.MkdirAll(workDir, 0755)
	os.WriteFile(filepath.Join(workDir, "README.md"), []byte("# Brief\n"), 0644)

	code := runNew([]string{"--folio", yml, "design", "mydesign"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	// Design colocates with nested format inside existing work dir
	matches, _ := filepath.Glob(filepath.Join(workDir, "reference", "design", "*-mydesign.md"))
	if len(matches) != 1 {
		t.Fatalf("expected 1 file in work dir reference/design/, got %d", len(matches))
	}

	data, _ := os.ReadFile(matches[0])
	if !strings.Contains(string(data), "## Problem") {
		t.Error("design file missing Problem section")
	}

	// folio.yml source entry should contain work/active/
	ymlData, _ := os.ReadFile(yml)
	if !strings.Contains(string(ymlData), "work/active/") {
		t.Error("folio.yml source entry should reference colocated work path")
	}
}

func TestRunNewDesignAutoCreatesWorkDir(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	// No existing work dir — design should auto-create one
	code := runNew([]string{"--folio", yml, "design", "standalone-design"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	// Should NOT be in reference/design/ (old behavior)
	oldMatches, _ := filepath.Glob(filepath.Join(dir, "reference", "design", "*-standalone-design.md"))
	if len(oldMatches) != 0 {
		t.Errorf("design should not create in reference/design/ (old behavior), found %d files", len(oldMatches))
	}

	// Should be in work/active/*/reference/design/
	newMatches, _ := filepath.Glob(filepath.Join(dir, "work", "active", "*-standalone-design", "reference", "design", "*-standalone-design.md"))
	if len(newMatches) != 1 {
		t.Fatalf("expected 1 file in auto-created work dir, got %d", len(newMatches))
	}
}

func TestRunNewPlanUsesExistingWorkDir(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	// Create work dir (as if design created it on a different date)
	workDir := filepath.Join(dir, "work", "active", "2026-01-15-my-topic")
	os.MkdirAll(workDir, 0755)

	code := runNew([]string{"--folio", yml, "plan", "my-topic"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	// Plan should land in the existing work dir, not create a new one
	readmePath := filepath.Join(workDir, "README.md")
	if _, err := os.Stat(readmePath); err != nil {
		t.Fatalf("expected README.md in existing work dir: %v", err)
	}

	// Should NOT create a second work dir with today's date
	matches, _ := filepath.Glob(filepath.Join(dir, "work", "active", "*-my-topic"))
	if len(matches) != 1 {
		t.Errorf("expected exactly 1 work dir for topic, got %d", len(matches))
	}
}

func TestRunNewRoundCreatesFirstRound(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	workDir := filepath.Join(dir, "work", "active", "2026-01-01-my-topic")
	os.MkdirAll(workDir, 0755)

	code := runNew([]string{"--folio", yml, "round", "my-topic"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	roundDir := filepath.Join(workDir, "agent-research", "0001-round")
	if info, err := os.Stat(roundDir); err != nil || !info.IsDir() {
		t.Fatalf("expected directory at %s", roundDir)
	}
}

func TestRunNewRoundAutoIncrements(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	workDir := filepath.Join(dir, "work", "active", "2026-01-01-my-topic")
	os.MkdirAll(filepath.Join(workDir, "agent-research", "0001-round"), 0755)
	os.MkdirAll(filepath.Join(workDir, "agent-research", "0002-round"), 0755)

	code := runNew([]string{"--folio", yml, "round", "my-topic"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	roundDir := filepath.Join(workDir, "agent-research", "0003-round")
	if info, err := os.Stat(roundDir); err != nil || !info.IsDir() {
		t.Fatalf("expected directory at %s", roundDir)
	}
}

func TestRunNewRoundMissingWorkDir(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	code := runNew([]string{"--folio", yml, "round", "nonexistent"})
	if code != 1 {
		t.Errorf("expected exit code 1 for missing work dir, got %d", code)
	}
}

func TestRunNewRoundDryRun(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	workDir := filepath.Join(dir, "work", "active", "2026-01-01-my-topic")
	os.MkdirAll(workDir, 0755)

	code := runNew([]string{"--folio", yml, "--dry-run", "round", "my-topic"})
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}

	roundDir := filepath.Join(workDir, "agent-research", "0001-round")
	if _, err := os.Stat(roundDir); err == nil {
		t.Error("dry-run should not create the directory")
	}
}

func TestRunNewSlugifiesSpaces(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\nsources: []\n"), 0644)

	workDir := filepath.Join(dir, "work", "active", "2026-01-01-my-topic")
	os.MkdirAll(workDir, 0755)

	code := runNew([]string{"--folio", yml, "round", "my topic"})
	if code != 0 {
		t.Fatalf("expected exit code 0 with spaces in topic, got %d", code)
	}

	roundDir := filepath.Join(workDir, "agent-research", "0001-round")
	if info, err := os.Stat(roundDir); err != nil || !info.IsDir() {
		t.Fatalf("expected directory at %s", roundDir)
	}
}
