package observe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	valid := []string{
		"bug(cli): something broken",
		"gap(health): missing feature",
		"idea(plan): potential improvement",
		"debt(cli): known shortcut",
		"task(roadmap): action item",
		"bug(my-scope2): with numbers",
	}
	for _, item := range valid {
		if err := Validate(item); err != nil {
			t.Errorf("Validate(%q) = %v, want nil", item, err)
		}
	}

	invalid := []string{
		"freeform text",
		"BUG(cli): uppercase type",
		"bug(CLI): uppercase scope",
		"bug(cli) missing colon-space",
		"bug(): empty scope",
		"bug(cli):",
		"bug(cli): ",
		"unknown(cli): bad type",
		"bug(9bad): scope starts with digit",
	}
	for _, item := range invalid {
		if err := Validate(item); err == nil {
			t.Errorf("Validate(%q) = nil, want error", item)
		}
	}
}

func TestValidateAllTypes(t *testing.T) {
	types := ValidTypes()
	if len(types) != 5 {
		t.Fatalf("ValidTypes() returned %d entries, want 5", len(types))
	}
	for _, ti := range types {
		item := ti.Name + "(test): description"
		if err := Validate(item); err != nil {
			t.Errorf("type %q should be valid: %v", ti.Name, err)
		}
	}
}

func TestParseObservation(t *testing.T) {
	typ, scope, desc, err := ParseObservation("bug(cli): something broken")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if typ != "bug" {
		t.Errorf("type = %q, want bug", typ)
	}
	if scope != "cli" {
		t.Errorf("scope = %q, want cli", scope)
	}
	if desc != "something broken" {
		t.Errorf("desc = %q, want 'something broken'", desc)
	}
}

func TestParseObservationInvalid(t *testing.T) {
	_, _, _, err := ParseObservation("freeform text")
	if err == nil {
		t.Error("expected error for malformed input")
	}
}

func TestValidTypes(t *testing.T) {
	types := ValidTypes()
	if len(types) != 5 {
		t.Fatalf("ValidTypes() returned %d entries, want 5", len(types))
	}
	names := map[string]bool{}
	for _, ti := range types {
		names[ti.Name] = true
		if ti.Description == "" {
			t.Errorf("type %q has empty description", ti.Name)
		}
	}
	for _, expected := range []string{"bug", "gap", "idea", "debt", "task"} {
		if !names[expected] {
			t.Errorf("missing type %q", expected)
		}
	}
}

func TestAppendValidatesFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "folio.yml")
	os.WriteFile(path, []byte("schema: 2\nproject: \"Test\"\nobservations: []\n"), 0644)

	err := Append(path, "freeform text")
	if err == nil {
		t.Error("expected Append to reject invalid format")
	}
}

func TestAppendToEmptyList(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "folio.yml")
	os.WriteFile(path, []byte("schema: 2\nproject: \"Test\"\nobservations: []\n"), 0644)

	if err := Append(path, "bug(testing): new item"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, `"bug(testing): new item"`) {
		t.Errorf("expected new item in observations, got:\n%s", content)
	}
	if strings.Contains(content, "observations: []") {
		t.Error("observations: [] should have been replaced with observations:")
	}
}

func TestAppendToExistingList(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "folio.yml")
	os.WriteFile(path, []byte("schema: 2\nproject: \"Test\"\nobservations:\n  - \"task(cli): existing item\"\n"), 0644)

	if err := Append(path, "bug(testing): new item"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, `"task(cli): existing item"`) {
		t.Error("existing item should be preserved")
	}
	if !strings.Contains(content, `"bug(testing): new item"`) {
		t.Error("new item should be appended")
	}
}

func TestAppendWithComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "folio.yml")
	os.WriteFile(path, []byte("schema: 2\nproject: \"Test\"\nobservations:\n  # Category A\n  - \"bug(cli): item a\"\n  # Category B\n  - \"task(cli): item b\"\n"), 0644)

	if err := Append(path, "idea(plan): item c"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	bIdx := strings.Index(content, `"task(cli): item b"`)
	cIdx := strings.Index(content, `"idea(plan): item c"`)
	if cIdx < bIdx {
		t.Errorf("new item should be after item b:\n%s", content)
	}
}

func TestAppendNoObservationsKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "folio.yml")
	os.WriteFile(path, []byte("schema: 2\nproject: \"Test\"\n"), 0644)

	err := Append(path, "bug(cli): item")
	if err == nil {
		t.Error("expected error for missing observations key")
	}
}

func TestAppendToObservations(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "folio.yml")
	os.WriteFile(path, []byte("schema: 2\nproject: \"Test\"\nobservations:\n  - \"task(cli): existing obs\"\n"), 0644)

	if err := Append(path, "gap(health): new obs"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, `"gap(health): new obs"`) {
		t.Errorf("expected new obs in observations, got:\n%s", content)
	}
}

func TestAppendPreservesFollowingKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "folio.yml")
	os.WriteFile(path, []byte("schema: 2\nproject: \"Test\"\nobservations:\n  - \"task(cli): existing\"\ncross_references:\n  - fact: \"test\"\n"), 0644)

	if err := Append(path, "bug(cli): new item"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if !strings.Contains(content, "cross_references:") {
		t.Error("cross_references key should be preserved")
	}
	if !strings.Contains(content, `"test"`) {
		t.Error("cross_references item should be preserved")
	}
}
