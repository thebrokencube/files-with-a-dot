package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunStaleMissingFile(t *testing.T) {
	code := runStale([]string{"--folio", "/nonexistent/folio.yml"})
	if code != 1 {
		t.Errorf("expected exit code 1 for missing file, got %d", code)
	}
}

func TestRunStaleNoTargets(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\n"), 0644)

	code := runStale([]string{"--folio", yml})
	if code != 0 {
		t.Errorf("expected exit code 0 for no targets, got %d", code)
	}
}

func TestRunStaleWithStaleTarget(t *testing.T) {
	dir := t.TempDir()

	// Create output directory and file with old mtime
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	outPath := filepath.Join(dir, "compiled", "out.md")
	os.WriteFile(outPath, []byte("# Old"), 0644)

	// Create source file AFTER output so it's newer
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  my-target:
    transform: distill
    sources:
      - path: src.md
    outputs:
      - path: compiled/out.md
`), 0644)

	os.WriteFile(filepath.Join(dir, "src.md"), []byte("# Updated Source"), 0644)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runStale([]string{"--folio", yml})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if code != 1 {
		t.Errorf("expected exit code 1 for stale targets, got %d", code)
	}

	if !strings.Contains(out, "my-target") {
		t.Error("expected target name in output")
	}
	if !strings.Contains(out, "stale") {
		t.Error("expected 'stale' status in output")
	}
	if !strings.Contains(out, "compiled/out.md") {
		t.Error("expected output path in output")
	}
	if !strings.Contains(out, "cause:") {
		t.Error("expected 'cause:' label in output")
	}
	if !strings.Contains(out, "src.md newer than output") {
		t.Error("expected cause mentioning source file")
	}
}

func TestRunStaleJSON(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\n"), 0644)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runStale([]string{"--folio", yml, "--json"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	if code != 0 {
		t.Errorf("expected exit code 0 for JSON mode with no targets, got %d", code)
	}

	var sj struct {
		Stale []struct {
			ID      string   `json:"id"`
			Status  string   `json:"status"`
			Outputs []string `json:"outputs"`
			Cause   string   `json:"cause"`
		} `json:"stale"`
	}
	if err := json.Unmarshal(buf.Bytes(), &sj); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(sj.Stale) != 0 {
		t.Errorf("expected empty stale list, got %d entries", len(sj.Stale))
	}
}

func TestRunStaleJSONWithStaleTarget(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "compiled", "out.md"), []byte("# Old"), 0644)

	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  my-target:
    transform: distill
    sources:
      - path: src.md
    outputs:
      - path: compiled/out.md
`), 0644)

	os.WriteFile(filepath.Join(dir, "src.md"), []byte("# Updated"), 0644)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runStale([]string{"--folio", yml, "--json"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}

	var sj struct {
		Stale []struct {
			ID      string   `json:"id"`
			Status  string   `json:"status"`
			Outputs []string `json:"outputs"`
			Cause   string   `json:"cause"`
		} `json:"stale"`
	}
	if err := json.Unmarshal(buf.Bytes(), &sj); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(sj.Stale) != 1 {
		t.Fatalf("expected 1 stale entry, got %d", len(sj.Stale))
	}
	entry := sj.Stale[0]
	if entry.ID != "my-target" {
		t.Errorf("id = %q, want my-target", entry.ID)
	}
	if entry.Status != "stale" {
		t.Errorf("status = %q, want stale", entry.Status)
	}
	if len(entry.Outputs) != 1 || entry.Outputs[0] != "compiled/out.md" {
		t.Errorf("outputs = %v, want [compiled/out.md]", entry.Outputs)
	}
	if entry.Cause == "" {
		t.Error("expected non-empty cause")
	}
}

func TestRunStaleJSONWithBranch(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "compiled", "out.md"), []byte("# Old"), 0644)

	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  my-target:
    transform: distill
    branch: "feat-my-branch"
    sources:
      - path: src.md
    outputs:
      - path: compiled/out.md
`), 0644)

	os.WriteFile(filepath.Join(dir, "src.md"), []byte("# Updated"), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runStale([]string{"--folio", yml, "--json"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}

	var sj struct {
		Stale []struct {
			ID     string `json:"id"`
			Branch string `json:"branch"`
		} `json:"stale"`
	}
	if err := json.Unmarshal(buf.Bytes(), &sj); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(sj.Stale) != 1 {
		t.Fatalf("expected 1 stale entry, got %d", len(sj.Stale))
	}
	if sj.Stale[0].Branch != "feat-my-branch" {
		t.Errorf("branch = %q, want feat-my-branch", sj.Stale[0].Branch)
	}
}

func TestRunStaleNoColor(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "compiled", "out.md"), []byte("# Old"), 0644)

	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  my-target:
    transform: distill
    sources:
      - path: src.md
    outputs:
      - path: compiled/out.md
`), 0644)

	os.WriteFile(filepath.Join(dir, "src.md"), []byte("# Updated"), 0644)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runStale([]string{"--folio", yml, "--no-color"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}

	if strings.Contains(out, "\033[") {
		t.Error("expected no ANSI codes with --no-color")
	}
}
