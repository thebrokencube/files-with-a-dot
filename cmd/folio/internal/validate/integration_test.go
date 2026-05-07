package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

// Integration tests: YAML on disk → config.Load → Validate → assert result.
// Each test creates a self-consistent temp directory with folio.yml and any
// referenced files.

func writeFixture(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	os.MkdirAll(filepath.Dir(path), 0755)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writing %s: %v", name, err)
	}
}

func loadAndValidate(t *testing.T, dir string) *Result {
	t.Helper()
	folioPath := filepath.Join(dir, "folio.yml")
	f, err := config.Load(folioPath)
	if err != nil {
		t.Fatalf("loading folio.yml: %v", err)
	}
	return Validate(f, dir)
}

func TestIntegrationMinimalValid(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "folio.yml", `
schema: 1
project: "Test"
sources: []
targets: {}
observations: []
`)
	r := loadAndValidate(t, dir)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestIntegrationWithSourcesAndTargets(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "README.md", "# Test")
	writeFixture(t, dir, "compiled/.gitkeep", "")
	writeFixture(t, dir, "folio.yml", `
schema: 1
project: "Full Test"
sources:
  - path: README.md
  - external: jira
    id: "ACME-123"
targets:
  summary:
    how: "Condense"
    sources:
      - path: README.md
    outputs:
      - path: compiled/summary.md
observations: []
`)
	r := loadAndValidate(t, dir)
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestIntegrationBadSchema(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "folio.yml", `
schema: 99
project: "Bad"
sources: []
targets: {}
`)
	r := loadAndValidate(t, dir)
	if r.Valid {
		t.Error("expected invalid for schema 99")
	}
	if !hasError(r, "schema version") {
		t.Errorf("expected schema error, got: %v", r.Errors)
	}
}

func TestIntegrationMissingProject(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "folio.yml", `
schema: 1
sources: []
targets: {}
`)
	r := loadAndValidate(t, dir)
	if r.Valid {
		t.Error("expected invalid for missing project")
	}
	if !hasError(r, "project") {
		t.Errorf("expected project error, got: %v", r.Errors)
	}
}

func TestIntegrationCycleDetection(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "compiled/.gitkeep", "")
	writeFixture(t, dir, "folio.yml", `
schema: 1
project: "Cycle"
sources: []
targets:
  a:
    how: "A"
    sources:
      - path: compiled/b.md
    outputs:
      - path: compiled/a.md
  b:
    how: "B"
    sources:
      - path: compiled/a.md
    outputs:
      - path: compiled/b.md
`)
	r := loadAndValidate(t, dir)
	if r.Valid {
		t.Error("expected invalid for cycle")
	}
	if !hasError(r, "cycle") {
		t.Errorf("expected cycle error, got: %v", r.Errors)
	}
}

func TestIntegrationOutputCollision(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "compiled/.gitkeep", "")
	writeFixture(t, dir, "folio.yml", `
schema: 1
project: "Collision"
sources: []
targets:
  first:
    how: "First"
    sources: []
    outputs:
      - path: compiled/same.md
  second:
    how: "Second"
    sources: []
    outputs:
      - path: compiled/same.md
`)
	r := loadAndValidate(t, dir)
	if r.Valid {
		t.Error("expected invalid for collision")
	}
	if !hasError(r, "Output collision") {
		t.Errorf("expected collision error, got: %v", r.Errors)
	}
}

func TestIntegrationPrecompileViolation(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "folio.yml", `
schema: 1
project: "Precompile"
sources: []
targets:
  external-only:
    how: "No local sibling"
    sources: []
    outputs:
      - external: jira
        id: "PROJ-123"
        field: description
`)
	r := loadAndValidate(t, dir)
	if r.Valid {
		t.Error("expected invalid for precompile violation")
	}
	if !hasError(r, "precompile rule") {
		t.Errorf("expected precompile error, got: %v", r.Errors)
	}
}

func TestIntegrationDeprecatedWarning(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "folio.yml", `
schema: 1
project: "Deprecated"
context_sources:
  - name: "old"
sources: []
targets: {}
`)
	r := loadAndValidate(t, dir)
	if !r.Valid {
		t.Errorf("expected valid (warning only), got errors: %v", r.Errors)
	}
	if len(r.Warnings) == 0 || !strings.Contains(r.Warnings[0], "context_sources") {
		t.Errorf("expected deprecation warning, got: %v", r.Warnings)
	}
}

func TestIntegrationEdgeInference(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "README.md", "# Source")
	writeFixture(t, dir, "compiled/.gitkeep", "")
	writeFixture(t, dir, "compiled/summary.md", "# Compiled upstream output")
	writeFixture(t, dir, "folio.yml", `
schema: 1
project: "Edge Inference"
sources: []
targets:
  upstream:
    how: "Produces summary"
    sources:
      - path: README.md
    outputs:
      - path: compiled/summary.md
  downstream:
    how: "Consumes summary"
    sources:
      - path: compiled/summary.md
    outputs:
      - path: compiled/final.md
`)
	r := loadAndValidate(t, dir)
	// Should be valid — inferred edge (downstream depends on upstream) with no cycle
	if !r.Valid {
		t.Errorf("expected valid, got errors: %v", r.Errors)
	}
}

func TestIntegrationInferredCycle(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "compiled/.gitkeep", "")
	writeFixture(t, dir, "compiled/a.md", "")
	writeFixture(t, dir, "compiled/b.md", "")
	writeFixture(t, dir, "folio.yml", `
schema: 1
project: "Inferred Cycle"
sources: []
targets:
  a:
    how: "A consumes B's output"
    sources:
      - path: compiled/b.md
    outputs:
      - path: compiled/a.md
  b:
    how: "B consumes A's output"
    sources:
      - path: compiled/a.md
    outputs:
      - path: compiled/b.md
`)
	r := loadAndValidate(t, dir)
	if r.Valid {
		t.Error("expected invalid for inferred cycle")
	}
	if !hasError(r, "cycle") {
		t.Errorf("expected cycle error, got: %v", r.Errors)
	}
}

func TestIntegrationMissingHow(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "compiled/.gitkeep", "")
	writeFixture(t, dir, "folio.yml", `
schema: 1
project: "No How"
sources: []
targets:
  oops:
    sources: []
    outputs:
      - path: compiled/out.md
`)
	r := loadAndValidate(t, dir)
	if !r.Valid {
		t.Errorf("expected valid (warning only for missing how), got errors: %v", r.Errors)
	}
	if !hasWarning(r, "missing 'how' field") {
		t.Errorf("expected how warning, got warnings: %v", r.Warnings)
	}
}

func TestIntegrationDeprecatedInstructions(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "compiled/.gitkeep", "")
	writeFixture(t, dir, "folio.yml", `
schema: 1
project: "Deprecated Instructions"
sources: []
targets:
  target:
    instructions: "Using old field"
    sources: []
    outputs:
      - path: compiled/out.md
`)
	r := loadAndValidate(t, dir)
	if !r.Valid {
		t.Errorf("expected valid (warning only), got errors: %v", r.Errors)
	}
	if !hasWarning(r, "'instructions' is deprecated") {
		t.Errorf("expected deprecation warning, got warnings: %v", r.Warnings)
	}
}

func TestIntegrationDeprecatedTransform(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "compiled/.gitkeep", "")
	writeFixture(t, dir, "folio.yml", `
schema: 1
project: "Deprecated Transform"
sources: []
targets:
  target:
    how: "Test"
    transform: distill
    sources: []
    outputs:
      - path: compiled/out.md
`)
	r := loadAndValidate(t, dir)
	if !r.Valid {
		t.Errorf("expected valid (warning only), got errors: %v", r.Errors)
	}
	if !hasWarning(r, "'transform' is deprecated") {
		t.Errorf("expected deprecation warning, got warnings: %v", r.Warnings)
	}
}

func TestIntegrationSourceDependsOnCycle(t *testing.T) {
	dir := t.TempDir()
	writeFixture(t, dir, "a.md", "# A")
	writeFixture(t, dir, "b.md", "# B")
	writeFixture(t, dir, "folio.yml", `
schema: 1
project: "Source Cycle"
sources:
  - path: a.md
    depends_on: [b.md]
  - path: b.md
    depends_on: [a.md]
targets: {}
`)
	r := loadAndValidate(t, dir)
	if r.Valid {
		t.Error("expected invalid for source dependency cycle")
	}
	if !hasError(r, "Source dependency cycle") {
		t.Errorf("expected cycle error, got: %v", r.Errors)
	}
}


func hasWarning(r *Result, substr string) bool {
	for _, w := range r.Warnings {
		if strings.Contains(w, substr) {
			return true
		}
	}
	return false
}

func hasError(r *Result, substr string) bool {
	for _, e := range r.Errors {
		if strings.Contains(e, substr) {
			return true
		}
	}
	return false
}
