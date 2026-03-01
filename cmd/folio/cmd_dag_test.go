package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunDagMissingFile(t *testing.T) {
	code := runDag([]string{"--folio", "/nonexistent/folio.yml"})
	if code != 1 {
		t.Errorf("expected exit code 1 for missing file, got %d", code)
	}
}

func TestRunDagNoTargets(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\n"), 0644)

	code := runDag([]string{"--folio", yml})
	if code != 0 {
		t.Errorf("expected exit code 0 for no targets, got %d", code)
	}
}

func TestRunDagWithEdges(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  downstream:
    transform: distill
    blocked_by: [upstream]
    sources:
      - path: compiled/up.md
    outputs:
      - path: compiled/down.md
  upstream:
    transform: distill
    sources:
      - path: README.md
    outputs:
      - path: compiled/up.md
`), 0644)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runDag([]string{"--folio", yml})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	// Should show sources and outputs
	if !strings.Contains(out, "sources:") {
		t.Error("expected 'sources:' label in output")
	}
	if !strings.Contains(out, "outputs:") {
		t.Error("expected 'outputs:' label in output")
	}
	if !strings.Contains(out, "compiled/down.md") {
		t.Error("expected output path 'compiled/down.md' in output")
	}
	if !strings.Contains(out, "README.md") {
		t.Error("expected source path 'README.md' in output")
	}
	// Should show edge
	if !strings.Contains(out, "upstream") {
		t.Error("expected 'upstream' in output")
	}
}

func TestRunDagJSON(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  downstream:
    transform: distill
    blocked_by: [upstream]
    sources:
      - path: compiled/up.md
    outputs:
      - path: compiled/down.md
  upstream:
    transform: distill
    sources:
      - path: README.md
    outputs:
      - path: compiled/up.md
`), 0644)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runDag([]string{"--folio", yml, "--json"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	if code != 0 {
		t.Errorf("expected exit code 0 for JSON mode, got %d", code)
	}

	var dj dagJSON
	if err := json.Unmarshal(buf.Bytes(), &dj); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if len(dj.Nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(dj.Nodes))
	}

	// Find upstream node
	for _, node := range dj.Nodes {
		if node.ID == "upstream" {
			if len(node.Sources) != 1 || node.Sources[0] != "README.md" {
				t.Errorf("upstream sources = %v, want [README.md]", node.Sources)
			}
			if len(node.Outputs) != 1 || node.Outputs[0] != "compiled/up.md" {
				t.Errorf("upstream outputs = %v, want [compiled/up.md]", node.Outputs)
			}
		}
	}

	if len(dj.Edges) == 0 {
		t.Error("expected at least one edge")
	}
}

func TestRunDagBranches(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  docs-tooling:
    transform: compose
    branch: "feat-tooling"
    pr: "#100"
    sources: []
    outputs:
      - path: compiled/a.md
  docs-proposal:
    transform: compose
    branch: "feat-proposal"
    pr: "#200"
    blocked_by: [docs-tooling]
    sources: []
    outputs:
      - path: compiled/b.md
`), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runDag([]string{"--folio", yml, "--branches"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(out, "feat-tooling") {
		t.Error("expected branch name 'feat-tooling' in output")
	}
	if !strings.Contains(out, "feat-proposal") {
		t.Error("expected branch name 'feat-proposal' in output")
	}
	if !strings.Contains(out, "(base: main)") {
		t.Error("expected '(base: main)' for root branch")
	}
	if !strings.Contains(out, "(base: feat-tooling)") {
		t.Error("expected '(base: feat-tooling)' for stacked branch")
	}
	if !strings.Contains(out, "PR: #100") {
		t.Error("expected 'PR: #100' in output")
	}
	// Tree connectors
	if !strings.Contains(out, "├── ") && !strings.Contains(out, "└── ") {
		t.Error("expected tree connectors in output")
	}
}

func TestRunDagBranchesAndJsonError(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\n"), 0644)

	code := runDag([]string{"--folio", yml, "--branches", "--json"})
	if code != 1 {
		t.Errorf("expected exit code 1 for --branches --json, got %d", code)
	}
}

func TestRunDagBranchesNoColor(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  my-target:
    transform: distill
    branch: "feat/no-color"
    pr: "#99"
    sources: []
    outputs:
      - path: compiled/out.md
`), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runDag([]string{"--folio", yml, "--branches", "--no-color"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if strings.Contains(out, "\033[") {
		t.Error("expected no ANSI codes with --no-color")
	}
	if !strings.Contains(out, "feat/no-color") {
		t.Error("expected branch name in output")
	}
}

func TestRunDagBranchesNoBranches(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  my-target:
    transform: distill
    sources:
      - path: README.md
    outputs:
      - path: compiled/out.md
`), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runDag([]string{"--folio", yml, "--branches"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if out != "" {
		t.Errorf("expected empty output when no targets have branch, got: %q", out)
	}
}

func TestRunDagNoColor(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  my-target:
    transform: distill
    sources:
      - path: README.md
    outputs:
      - path: compiled/out.md
`), 0644)

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runDag([]string{"--folio", yml, "--no-color"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	if strings.Contains(out, "\033[") {
		t.Error("expected no ANSI codes with --no-color")
	}
}
