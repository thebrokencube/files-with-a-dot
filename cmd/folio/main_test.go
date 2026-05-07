package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunValidateMissingFile(t *testing.T) {
	code := runValidate([]string{"--folio", "/nonexistent/folio.yml"})
	if code != 2 {
		t.Errorf("expected exit code 2 for missing file, got %d", code)
	}
}

func TestRunValidateMinimalValid(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\n"), 0644)

	code := runValidate([]string{"--folio", yml, "--no-color"})
	if code != 0 {
		t.Errorf("expected exit code 0 for valid folio, got %d", code)
	}
}

func TestRunValidateInvalidSchema(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 99\nproject: \"Test\"\n"), 0644)

	code := runValidate([]string{"--folio", yml, "--no-color"})
	if code != 2 {
		t.Errorf("expected exit code 2 for invalid schema, got %d", code)
	}
}

func TestRunValidateJSON(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\n"), 0644)

	code := runValidate([]string{"--folio", yml, "--json"})
	if code != 0 {
		t.Errorf("expected exit code 0 for JSON mode, got %d", code)
	}
}

func TestRunStatusMissingFile(t *testing.T) {
	code := runStatus([]string{"--folio", "/nonexistent/folio.yml"})
	if code != 1 {
		t.Errorf("expected exit code 1 for missing file, got %d", code)
	}
}

func TestRunStatusMinimal(t *testing.T) {
	dir := t.TempDir()
	yml := filepath.Join(dir, "folio.yml")
	os.WriteFile(yml, []byte("schema: 1\nproject: \"Test\"\n"), 0644)

	code := runStatus([]string{"--folio", yml, "--no-color"})
	if code != 0 {
		t.Errorf("expected exit code 0 for minimal folio, got %d", code)
	}
}

func TestRunInitAlreadyExists(t *testing.T) {
	homeDir := t.TempDir()
	activeDir := filepath.Join(homeDir, "active")
	os.MkdirAll(activeDir, 0755)
	// Pre-create the target so init sees a conflict
	os.MkdirAll(filepath.Join(activeDir, "test"), 0755)
	os.WriteFile(filepath.Join(activeDir, "test", "folio.yml"), []byte("schema: 1\n"), 0644)
	t.Setenv("FOLIO_HOME", homeDir)

	code := runInit([]string{"--name", "test"})
	if code != 1 {
		t.Errorf("expected exit code 1 for existing folio.yml, got %d", code)
	}
}

func TestRunInitMissingName(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	code := runInit([]string{})
	if code != 1 {
		t.Errorf("expected exit code 1 for missing --name, got %d", code)
	}
}

func TestRunInitCreatesFile(t *testing.T) {
	// When FOLIO_HOME has an active/ directory, init should place the folio there.
	homeDir := t.TempDir()
	activeDir := filepath.Join(homeDir, "active")
	os.MkdirAll(activeDir, 0755)
	t.Setenv("FOLIO_HOME", homeDir)

	code := runInit([]string{"--name", "my-project"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	data, err := os.ReadFile(filepath.Join(activeDir, "my-project", "folio.yml"))
	if err != nil {
		t.Fatalf("folio.yml not created in FOLIO_HOME/active/my-project/: %v", err)
	}
	if len(data) == 0 {
		t.Error("folio.yml is empty")
	}
}

func TestRunInitFallsBackToCWD(t *testing.T) {
	// When FOLIO_HOME has no active/ directory, init should fall back to CWD.
	homeDir := t.TempDir()
	t.Setenv("FOLIO_HOME", homeDir)

	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	code := runInit([]string{"--name", "my-project"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	data, err := os.ReadFile(filepath.Join(dir, "folio.yml"))
	if err != nil {
		t.Fatalf("folio.yml not created in CWD: %v", err)
	}
	if len(data) == 0 {
		t.Error("folio.yml is empty")
	}
}

func TestRunSetup(t *testing.T) {
	code := runSetup([]string{})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunSetupCheck(t *testing.T) {
	code := runSetup([]string{"--check"})
	if code != 0 {
		t.Errorf("expected exit code 0 for --check, got %d", code)
	}
}
