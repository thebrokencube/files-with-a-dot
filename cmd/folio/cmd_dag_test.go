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
    how: "Test"
    blocked_by: [upstream]
    sources:
      - path: compiled/up.md
    outputs:
      - path: compiled/down.md
  upstream:
    how: "Test"
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
    how: "Test"
    blocked_by: [upstream]
    sources:
      - path: compiled/up.md
    outputs:
      - path: compiled/down.md
  upstream:
    how: "Test"
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
    how: "Test"
    branch: "feat-tooling"
    pr: "#100"
    sources: []
    outputs:
      - path: compiled/a.md
  docs-proposal:
    how: "Test"
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

func TestRunDagBranchesAndJson(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  docs-tooling:
    how: "Test"
    branch: "feat-tooling"
    pr: "#100"
    sources: []
    outputs:
      - path: compiled/a.md
  docs-proposal:
    how: "Test"
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

	code := runDag([]string{"--folio", yml, "--branches", "--json"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)

	if code != 0 {
		t.Errorf("expected exit code 0 for --branches --json, got %d", code)
	}

	var bt struct {
		Roots []struct {
			Base     string `json:"base"`
			Children []struct {
				ID       string `json:"id"`
				Branch   string `json:"branch"`
				Base     string `json:"base"`
				PR       string `json:"pr"`
				Children []struct {
					ID     string `json:"id"`
					Branch string `json:"branch"`
					Base   string `json:"base"`
				} `json:"children"`
			} `json:"children"`
		} `json:"roots"`
	}
	if err := json.Unmarshal(buf.Bytes(), &bt); err != nil {
		t.Fatalf("failed to parse JSON: %v\noutput: %s", err, buf.String())
	}
	if len(bt.Roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(bt.Roots))
	}
	if bt.Roots[0].Base != "main" {
		t.Errorf("root base = %q, want main", bt.Roots[0].Base)
	}
	if len(bt.Roots[0].Children) != 1 {
		t.Fatalf("expected 1 child of main, got %d", len(bt.Roots[0].Children))
	}
	tooling := bt.Roots[0].Children[0]
	if tooling.ID != "docs-tooling" {
		t.Errorf("child ID = %q, want docs-tooling", tooling.ID)
	}
	if tooling.Branch != "feat-tooling" {
		t.Errorf("branch = %q, want feat-tooling", tooling.Branch)
	}
	if len(tooling.Children) != 1 {
		t.Fatalf("expected 1 child of docs-tooling, got %d", len(tooling.Children))
	}
	if tooling.Children[0].ID != "docs-proposal" {
		t.Errorf("grandchild ID = %q, want docs-proposal", tooling.Children[0].ID)
	}
}

func TestRunDagStatusRequiresBranches(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\n"), 0644)

	code := runDag([]string{"--folio", yml, "--status"})
	if code != 1 {
		t.Errorf("expected exit code 1 for --status without --branches, got %d", code)
	}
}

func TestRunDagBranchesStatus(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "compiled", "a.md"), []byte("# Old"), 0644)

	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  my-target:
    how: "Test"
    branch: "feat-test"
    sources:
      - path: src.md
    outputs:
      - path: compiled/a.md
`), 0644)

	// Create source AFTER output to make it stale
	os.WriteFile(filepath.Join(dir, "src.md"), []byte("# Updated"), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := runDag([]string{"--folio", yml, "--branches", "--status"})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	out := buf.String()

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(out, "stale") {
		t.Error("expected 'stale' status annotation in output")
	}
	if !strings.Contains(out, "feat-test") {
		t.Error("expected branch name in output")
	}
}

func TestRunDagBranchesNoColor(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  my-target:
    how: "Test"
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
    how: "Test"
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
    how: "Test"
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
