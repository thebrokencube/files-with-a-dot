package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/touch"
)

func TestRunStaleMissingFile(t *testing.T) {
	code := buildRoot().Execute([]string{"stale", "--folio", "/nonexistent/folio.yml"})
	if code != 1 {
		t.Errorf("expected exit code 1 for missing file, got %d", code)
	}
}

func TestRunStaleNoTargets(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\n"), 0644)

	code := buildRoot().Execute([]string{"stale", "--folio", yml})
	if code != 0 {
		t.Errorf("expected exit code 0 for no targets, got %d", code)
	}
}

func TestRunStaleWithStaleTarget(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "compiled", "out.md"), []byte("# Old"), 0644)
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  my-target:
    how: "Test"
    sources:
      - path: src.md
    outputs:
      - path: compiled/out.md
`), 0644)
	os.WriteFile(filepath.Join(dir, "src.md"), []byte("# Initial Source"), 0644)
	stampStaleTarget(t, yml, "my-target", dir)
	os.WriteFile(filepath.Join(dir, "src.md"), []byte("# Updated Source"), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	code := buildRoot().Execute([]string{"stale", "--folio", yml})
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
	if !strings.Contains(out, "source content changed since compose") {
		t.Error("expected content-change cause")
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

	code := buildRoot().Execute([]string{"stale", "--folio", yml, "--json"})

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
    how: "Test"
    sources:
      - path: src.md
    outputs:
      - path: compiled/out.md
`), 0644)
	os.WriteFile(filepath.Join(dir, "src.md"), []byte("# Initial"), 0644)
	stampStaleTarget(t, yml, "my-target", dir)
	os.WriteFile(filepath.Join(dir, "src.md"), []byte("# Updated"), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	code := buildRoot().Execute([]string{"stale", "--folio", yml, "--json"})
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}

	var env struct {
		Data struct {
			Stale []struct {
				ID      string   `json:"id"`
				Status  string   `json:"status"`
				Outputs []string `json:"outputs"`
				Cause   string   `json:"cause"`
			} `json:"stale"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(env.Data.Stale) != 1 {
		t.Fatalf("expected 1 stale entry, got %d", len(env.Data.Stale))
	}
	entry := env.Data.Stale[0]
	if entry.ID != "my-target" || entry.Status != "stale" {
		t.Fatalf("entry = %+v", entry)
	}
	if len(entry.Outputs) != 1 || entry.Outputs[0] != "compiled/out.md" {
		t.Errorf("outputs = %v, want [compiled/out.md]", entry.Outputs)
	}
	if entry.Cause == "" {
		t.Error("expected non-empty cause")
	}
}
func TestRunStaleUnknownSnapshotExitsZero(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "compiled", "out.md"), []byte("# Output"), 0644)
	os.WriteFile(filepath.Join(dir, "src.md"), []byte("# Source"), 0644)
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  my-target:
    how: "Test"
    sources:
      - path: src.md
    outputs:
      - path: compiled/out.md
`), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	code := buildRoot().Execute([]string{"stale", "--folio", yml, "--json"})
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	if code != 0 {
		t.Fatalf("unknown snapshot exit code = %d, want zero", code)
	}
	var env struct {
		Data struct {
			Stale []struct {
				Status string `json:"status"`
				Cause  string `json:"cause"`
			} `json:"stale"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data.Stale) != 1 || env.Data.Stale[0].Status != "unknown" || env.Data.Stale[0].Cause != "composition not recorded" {
		t.Fatalf("unknown stale entry = %+v", env.Data.Stale)
	}
}
func TestRunStaleForestDelegatedExcluded(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	os.WriteFile(filepath.Join(dir, "compiled", "forest.md"), []byte("# Forest"), 0644)
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Forest"
targets:
  forest-target:
    how: "jf"
    forest:
      root: forest
    outputs:
      - path: compiled/forest.md
`), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	code := buildRoot().Execute([]string{"stale", "--folio", yml, "--json"})
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	if code != 0 {
		t.Fatalf("forest stale exit code = %d, want zero", code)
	}
	var env struct {
		Data struct {
			Stale []struct {
				ID string `json:"id"`
			} `json:"stale"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data.Stale) != 0 {
		t.Fatalf("forest stale entries = %+v, want none", env.Data.Stale)
	}
}

func TestRunStaleMissingLocalOutput(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "compiled"), 0755)
	outputPath := filepath.Join(dir, "compiled", "out.md")
	os.WriteFile(outputPath, []byte("# Output"), 0644)
	os.WriteFile(filepath.Join(dir, "src.md"), []byte("# Source"), 0644)
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte(`schema: 1
project: "Test"
targets:
  my-target:
    how: "Test"
    sources:
      - path: src.md
    outputs:
      - path: compiled/out.md
`), 0644)
	stampStaleTarget(t, yml, "my-target", dir)
	if err := os.Remove(outputPath); err != nil {
		t.Fatal(err)
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	code := buildRoot().Execute([]string{"stale", "--folio", yml, "--json"})
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	if code != 1 {
		t.Fatalf("missing output exit code = %d, want one", code)
	}
	var env struct {
		Data struct {
			Stale []struct {
				Status string `json:"status"`
				Cause  string `json:"cause"`
			} `json:"stale"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatal(err)
	}
	if len(env.Data.Stale) != 1 || env.Data.Stale[0].Status != "missing" || env.Data.Stale[0].Cause != "output missing" {
		t.Fatalf("missing output entry = %+v", env.Data.Stale)
	}
}

func TestStaleOmitsOnlyCleanTerminalTarget(t *testing.T) {
	dir := t.TempDir()
	outputPath := filepath.Join(dir, "compiled", "review.md")
	writeTouchCommandFile(t, outputPath, "# Review")
	writeTouchCommandFile(t, filepath.Join(dir, "source.md"), "# Source")
	yml := filepath.Join(dir, "folio.yml")
	writeTouchCommandFile(t, yml, `schema: 3
project: "Test"
targets:
  publish:
    how: "Test"
    sources:
      - path: source.md
    outputs:
      - path: compiled/review.md
      - external: jira
        id: "PROJ-123"
`)
	if code := buildRoot().Execute([]string{"touch", "--folio", yml, "--final", "publish"}); code != 0 {
		t.Fatalf("final touch = %d, want zero", code)
	}
	if code, entries := runStaleJSONForTest(t, yml); code != 0 || len(entries) != 0 {
		t.Fatalf("clean terminal stale result = %d/%+v, want zero/empty", code, entries)
	}

	writeTouchCommandFile(t, filepath.Join(dir, "source.md"), "# Changed")
	if code, entries := runStaleJSONForTest(t, yml); code != 1 || len(entries) != 1 || entries[0].Status != "stale" {
		t.Fatalf("stale terminal result = %d/%+v, want one/stale", code, entries)
	}

	if err := os.Remove(outputPath); err != nil {
		t.Fatal(err)
	}
	if code, entries := runStaleJSONForTest(t, yml); code != 1 || len(entries) != 1 || entries[0].Status != "missing" {
		t.Fatalf("missing terminal result = %d/%+v, want one/missing", code, entries)
	}
}

type staleTestEntry struct {
	Status string `json:"status"`
}

func runStaleJSONForTest(t *testing.T, folioPath string) (int, []staleTestEntry) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	code := buildRoot().Execute([]string{"stale", "--folio", folioPath, "--json"})
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = old
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		Data struct {
			Stale []staleTestEntry `json:"stale"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	return code, envelope.Data.Stale
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
    how: "Test"
    sources:
      - path: src.md
    outputs:
      - path: compiled/out.md
`), 0644)
	os.WriteFile(filepath.Join(dir, "src.md"), []byte("# Initial"), 0644)
	stampStaleTarget(t, yml, "my-target", dir)
	os.WriteFile(filepath.Join(dir, "src.md"), []byte("# Updated"), 0644)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	code := buildRoot().Execute([]string{"stale", "--folio", yml, "--no-color"})
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

func stampStaleTarget(t *testing.T, folioPath, targetID, dir string) {
	t.Helper()
	f, err := config.Load(folioPath)
	if err != nil {
		t.Fatal(err)
	}
	target := f.Targets[targetID]
	ctx := config.Context{FolioPath: folioPath, WorkRoot: dir}
	state, err := touch.Snapshot(ctx, &target, false, time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if err := touch.WriteState(folioPath, targetID, state); err != nil {
		t.Fatal(err)
	}
}
