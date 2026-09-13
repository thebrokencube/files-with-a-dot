package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/status"
)

func TestRunTouchNoArgs(t *testing.T) {
	code := buildRoot().Execute([]string{"touch"})
	if code != 1 {
		t.Errorf("expected exit code 1 for no args, got %d", code)
	}
}

func TestRunTouchTargetNotFound(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	writeTouchCommandFile(t, yml, "schema: 1\nproject: \"Test\"\n")

	code := buildRoot().Execute([]string{"touch", "--folio", yml, "nonexistent"})
	if code != 1 {
		t.Errorf("expected exit code 1 for missing target, got %d", code)
	}
}

func TestRunTouchNoLocalOutput(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	writeTouchCommandFile(t, yml, `schema: 1
project: "Test"
targets:
  my-target:
    how: "Test"
    outputs:
      - external: jira
        id: "PROJ-123"
        field: description
`)

	code := buildRoot().Execute([]string{"touch", "--folio", yml, "my-target"})
	if code != 1 {
		t.Errorf("expected exit code 1 for no local output, got %d", code)
	}
}

func TestTouchWritesSnapshotWithoutChangingOutputMtimes(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source.md")
	output := filepath.Join(dir, "compiled", "out.md")
	writeTouchCommandFile(t, source, "# Source")
	writeTouchCommandFile(t, output, "# Output")
	mtime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(source, mtime, mtime); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(output, mtime, mtime); err != nil {
		t.Fatal(err)
	}
	yml := filepath.Join(dir, "folio.yml")
	writeTouchCommandFile(t, yml, `schema: 3
project: "Test"
targets:
  my-target:
    how: "Test"
    sources:
      - path: source.md
    outputs:
      - path: compiled/out.md
`)
	beforeBytes, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	beforeInfo, err := os.Stat(output)
	if err != nil {
		t.Fatal(err)
	}

	code := buildRoot().Execute([]string{"touch", "--folio", yml, "my-target"})
	if code != 0 {
		t.Fatalf("touch exit code = %d, want zero", code)
	}
	folio, err := config.Load(yml)
	if err != nil {
		t.Fatal(err)
	}
	target := folio.Targets["my-target"]
	if target.ComposedAt == "" || target.InputsSHA256 == "" || target.Final {
		t.Fatalf("recorded target state = %+v", target)
	}
	afterBytes, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	afterInfo, err := os.Stat(output)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(afterBytes, beforeBytes) || !afterInfo.ModTime().Equal(beforeInfo.ModTime()) {
		t.Fatal("touch changed output bytes or mtime")
	}
	freshness, cause := status.DeriveLocalStatus(config.Context{FolioPath: yml, WorkRoot: dir}, &target, target.Outputs[0])
	if freshness != "clean" || cause != "" {
		t.Fatalf("freshness = %q cause = %q, want clean/empty", freshness, cause)
	}
}

func TestTouchFailurePreservesManifestAndOutputMtimes(t *testing.T) {
	dir := t.TempDir()
	output := filepath.Join(dir, "compiled", "out.md")
	writeTouchCommandFile(t, output, "# Output")
	yml := filepath.Join(dir, "folio.yml")
	manifest := []byte(`schema: 3
project: "Test"
targets:
  my-target:
    how: "Test"
    sources:
      - path: missing.md
    outputs:
      - path: compiled/out.md
`)
	writeTouchCommandFile(t, yml, string(manifest))
	beforeInfo, err := os.Stat(output)
	if err != nil {
		t.Fatal(err)
	}

	code := buildRoot().Execute([]string{"touch", "--folio", yml, "my-target"})
	if code != 1 {
		t.Fatalf("touch exit code = %d, want one", code)
	}
	afterManifest, err := os.ReadFile(yml)
	if err != nil {
		t.Fatal(err)
	}
	afterInfo, err := os.Stat(output)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(afterManifest, manifest) {
		t.Fatal("failed touch changed manifest")
	}
	if !afterInfo.ModTime().Equal(beforeInfo.ModTime()) {
		t.Fatal("failed touch changed output mtime")
	}
}

func TestTouchAllowsSurvivingBaselineFindingWarning(t *testing.T) {
	dir := t.TempDir()
	writeTouchCommandFile(t, filepath.Join(dir, "source.md"), "# Source")
	writeTouchCommandFile(t, filepath.Join(dir, "compiled", "out.md"), "# Output")
	yml := filepath.Join(dir, "folio.yml")
	writeTouchCommandFile(t, yml, `schema: 3
project: "Test"
sources:
  - path: source.md
  - path: source.md
targets:
  my-target:
    how: "Test"
    sources:
      - path: source.md
    outputs:
      - path: compiled/out.md
`)

	code := buildRoot().Execute([]string{"touch", "--folio", yml, "my-target"})
	if code != 0 {
		t.Fatalf("touch with surviving baseline finding = %d, want zero", code)
	}
	folio, err := config.Load(yml)
	if err != nil {
		t.Fatal(err)
	}
	if folio.Targets["my-target"].InputsSHA256 == "" {
		t.Fatal("touch did not record state")
	}
}

func TestTouchFinalWritesTargetFlagWithoutProviderClaim(t *testing.T) {
	dir := t.TempDir()
	writeTouchCommandFile(t, filepath.Join(dir, "source.md"), "# Source")
	writeTouchCommandFile(t, filepath.Join(dir, "compiled", "review.md"), "# Review")
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
        field: description
`)

	code, stdout := captureTouchStdout(t, []string{"touch", "--folio", yml, "--final", "publish"})
	if code != 0 {
		t.Fatalf("final touch exit code = %d, want zero", code)
	}
	if !strings.Contains(stdout, "Finalized target publish") || strings.Contains(strings.ToLower(stdout), "verified") {
		t.Fatalf("final output makes the wrong claim: %q", stdout)
	}
	folio, err := config.Load(yml)
	if err != nil {
		t.Fatal(err)
	}
	target := folio.Targets["publish"]
	if !target.Final || target.ComposedAt == "" || target.InputsSHA256 == "" {
		t.Fatalf("final target state = %+v", target)
	}
}

func TestTouchRejectsAlreadyFinalUntilFlagRemoved(t *testing.T) {
	dir := t.TempDir()
	writeTouchCommandFile(t, filepath.Join(dir, "source.md"), "# Source")
	writeTouchCommandFile(t, filepath.Join(dir, "compiled", "review.md"), "# Review")
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
		t.Fatalf("initial final touch = %d, want zero", code)
	}
	before, err := os.ReadFile(yml)
	if err != nil {
		t.Fatal(err)
	}
	if code := buildRoot().Execute([]string{"touch", "--folio", yml, "publish"}); code != 1 {
		t.Fatalf("ordinary touch on final target = %d, want one", code)
	}
	after, err := os.ReadFile(yml)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, before) {
		t.Fatal("ordinary touch changed an already-final manifest")
	}
}

func TestTouchFinalRejectsLocalOnlyExternalOnlyBatchAndForest(t *testing.T) {
	tests := []struct {
		name  string
		yaml  string
		setup func(*testing.T, string)
	}{
		{
			name: "local only",
			yaml: `schema: 3
project: "Test"
targets:
  target:
    how: "Test"
    outputs:
      - path: compiled/review.md
`,
			setup: func(t *testing.T, dir string) {
				writeTouchCommandFile(t, filepath.Join(dir, "compiled", "review.md"), "# Review")
			},
		},
		{
			name: "external only",
			yaml: `schema: 3
project: "Test"
targets:
  target:
    how: "Test"
    outputs:
      - external: jira
        id: "PROJ-123"
`,
			setup: func(*testing.T, string) {},
		},
		{
			name: "batch",
			yaml: `schema: 3
project: "Test"
targets:
  target:
    how: "Test"
    outputs:
      - path: compiled/review.md
    batch:
      system: jira
`,
			setup: func(t *testing.T, dir string) {
				writeTouchCommandFile(t, filepath.Join(dir, "compiled", "review.md"), "# Review")
			},
		},
		{
			name: "forest",
			yaml: `schema: 3
project: "Test"
targets:
  target:
    how: "Test"
    outputs:
      - path: compiled/review.md
    forest:
      root: forest
`,
			setup: func(t *testing.T, dir string) {
				writeTouchCommandFile(t, filepath.Join(dir, "compiled", "review.md"), "# Review")
				if err := os.MkdirAll(filepath.Join(dir, "forest"), 0755); err != nil {
					t.Fatal(err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(t, dir)
			yml := filepath.Join(dir, "folio.yml")
			writeTouchCommandFile(t, yml, tt.yaml)
			before, err := os.ReadFile(yml)
			if err != nil {
				t.Fatal(err)
			}
			if code := buildRoot().Execute([]string{"touch", "--folio", yml, "--final", "target"}); code != 1 {
				t.Fatalf("final touch = %d, want one", code)
			}
			after, err := os.ReadFile(yml)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(after, before) {
				t.Fatal("invalid final touch changed manifest")
			}
		})
	}
}

func captureTouchStdout(t *testing.T, args []string) (int, string) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	code := buildRoot().Execute(args)
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
	return code, buf.String()
}

func writeTouchCommandFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
