package agentskills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateLayer1_ValidSkill(t *testing.T) {
	dir := testdataDir(t, "valid-skill")
	results := ValidateLayer1(dir, "valid-skill")
	errors := filterErrors(results)
	if len(errors) > 0 {
		for _, r := range errors {
			t.Errorf("[%s] %s: %s", r.CheckID, r.Severity, r.Message)
		}
	}
}

func TestValidateLayer1_NoSkillFile(t *testing.T) {
	dir := t.TempDir()
	results := ValidateLayer1(dir, "missing")
	assertHasCheck(t, results, "K1", SeverityError)
	if len(results) != 1 {
		t.Fatalf("expected 1 result (structure gate), got %d", len(results))
	}
}

func TestValidateLayer1_NoFrontmatter(t *testing.T) {
	dir := testdataDir(t, "no-frontmatter")
	results := ValidateLayer1(dir, "no-frontmatter")
	assertHasCheck(t, results, "K2", SeverityError)
}

func TestValidateLayer1_BadName(t *testing.T) {
	dir := testdataDir(t, "bad-name")
	results := ValidateLayer1(dir, "bad-name")
	k2s := filterByCheck(results, "K2")
	if len(k2s) == 0 {
		t.Fatal("expected K2 errors for bad name")
	}
	// Should have pattern violation and directory mismatch
	var hasPattern, hasMismatch bool
	for _, r := range k2s {
		if strings.Contains(r.Message, "does not match pattern") {
			hasPattern = true
		}
		if strings.Contains(r.Message, "does not match directory") {
			hasMismatch = true
		}
	}
	if !hasPattern {
		t.Error("expected name pattern violation")
	}
	if !hasMismatch {
		t.Error("expected directory name mismatch")
	}
}

func TestValidateLayer1_MissingName(t *testing.T) {
	dir := writeSkill(t, "---\ndescription: \"test\"\n---\n# Test\n")
	results := ValidateLayer1(dir, "test")
	assertHasCheck(t, results, "K2", SeverityError)
	found := false
	for _, r := range results {
		if r.CheckID == "K2" && strings.Contains(r.Message, "missing required field: name") {
			found = true
		}
	}
	if !found {
		t.Error("expected missing name error")
	}
}

func TestValidateLayer1_MissingDescription(t *testing.T) {
	dir := writeSkill(t, "---\nname: test\n---\n# Test\n")
	results := ValidateLayer1(dir, "test")
	assertHasCheck(t, results, "K2", SeverityError)
	found := false
	for _, r := range results {
		if r.CheckID == "K2" && strings.Contains(r.Message, "missing required field: description") {
			found = true
		}
	}
	if !found {
		t.Error("expected missing description error")
	}
}

func TestValidateLayer1_NameTooLong(t *testing.T) {
	longName := strings.Repeat("a", 65)
	dir := writeSkill(t, "---\nname: "+longName+"\ndescription: test\n---\n# Test\n")
	results := ValidateLayer1(dir, longName)
	found := false
	for _, r := range results {
		if r.CheckID == "K2" && strings.Contains(r.Message, "exceeds 64") {
			found = true
		}
	}
	if !found {
		t.Error("expected name length error")
	}
}

func TestValidateLayer1_DescriptionTooLong(t *testing.T) {
	longDesc := strings.Repeat("a", 1025)
	dir := writeSkill(t, "---\nname: test\ndescription: \""+longDesc+"\"\n---\n# Test\n")
	results := ValidateLayer1(dir, "test")
	found := false
	for _, r := range results {
		if r.CheckID == "K2" && strings.Contains(r.Message, "exceeds 1024") {
			found = true
		}
	}
	if !found {
		t.Error("expected description length error")
	}
}

func TestValidateLayer1_ExtraFields(t *testing.T) {
	dir := testdataDir(t, "extra-fields")
	results := ValidateLayer1(dir, "extra-fields")
	k2ext := filterByCheck(results, "K2-EXT")
	if len(k2ext) != 2 {
		t.Fatalf("expected 2 K2-EXT warnings, got %d", len(k2ext))
	}
	for _, r := range k2ext {
		if r.Severity != SeverityWarning {
			t.Errorf("K2-EXT should be warning, got %s", r.Severity)
		}
	}
}

func TestValidateLayer1_BrokenLinks(t *testing.T) {
	dir := testdataDir(t, "broken-links")
	results := ValidateLayer1(dir, "broken-links")
	k3s := filterByCheck(results, "K3")
	if len(k3s) != 2 {
		t.Fatalf("expected 2 broken links (not counting URL), got %d", len(k3s))
	}
}

func TestValidateLayer1_BadRefNaming(t *testing.T) {
	dir := testdataDir(t, "bad-ref-naming")
	results := ValidateLayer1(dir, "bad-ref-naming")
	k4s := filterByCheck(results, "K4")
	if len(k4s) != 1 {
		t.Fatalf("expected 1 K4 warning (BadName.md), got %d", len(k4s))
	}
	if k4s[0].Severity != SeverityWarning {
		t.Errorf("K4 should be warning, got %s", k4s[0].Severity)
	}
}

func TestValidateLayer1_Oversized(t *testing.T) {
	// Build a SKILL.md with >500 body lines
	var sb strings.Builder
	sb.WriteString("---\nname: oversized\ndescription: test\n---\n")
	for i := 0; i < 510; i++ {
		sb.WriteString("line content here\n")
	}
	dir := writeSkill(t, sb.String())
	results := ValidateLayer1(dir, "oversized")
	k5s := filterByCheck(results, "K5")
	hasLineError := false
	for _, r := range k5s {
		if r.Severity == SeverityError && strings.Contains(r.Message, "lines") {
			hasLineError = true
		}
	}
	if !hasLineError {
		t.Error("expected K5 line count error")
	}
}

func TestValidateLayer1_TokenWarning(t *testing.T) {
	// Build a SKILL.md with many long lines (~5000+ tokens = ~20000+ chars in body)
	var sb strings.Builder
	sb.WriteString("---\nname: token-test\ndescription: test\n---\n")
	longLine := strings.Repeat("word ", 100) // 500 chars per line
	for i := 0; i < 50; i++ {
		sb.WriteString(longLine + "\n")
	}
	dir := writeSkill(t, sb.String())
	results := ValidateLayer1(dir, "token-test")
	k5s := filterByCheck(results, "K5")
	hasTokenWarning := false
	for _, r := range k5s {
		if r.Severity == SeverityWarning && strings.Contains(r.Message, "tokens") {
			hasTokenWarning = true
		}
	}
	if !hasTokenWarning {
		t.Error("expected K5 token warning")
	}
}

func TestValidateLayer1_ValidNamePatterns(t *testing.T) {
	valid := []string{"a", "abc", "my-tool", "tool-123", "a-b-c"}
	for _, name := range valid {
		t.Run(name, func(t *testing.T) {
			dir := writeSkill(t, "---\nname: "+name+"\ndescription: test\n---\n# Test\n")
			results := ValidateLayer1(dir, name)
			for _, r := range results {
				if r.CheckID == "K2" && strings.Contains(r.Message, "pattern") {
					t.Errorf("name %q should be valid", name)
				}
			}
		})
	}
}

func TestValidateLayer1_InvalidNamePatterns(t *testing.T) {
	invalid := []string{"Bad", "has space", "UPPER", "has_underscore", "-leading", "trailing-"}
	for _, name := range invalid {
		t.Run(name, func(t *testing.T) {
			dir := writeSkill(t, "---\nname: \""+name+"\"\ndescription: test\n---\n# Test\n")
			results := ValidateLayer1(dir, name)
			found := false
			for _, r := range results {
				if r.CheckID == "K2" && strings.Contains(r.Message, "pattern") {
					found = true
				}
			}
			if !found {
				t.Errorf("name %q should be invalid", name)
			}
		})
	}
}

func TestValidateLayer1_URLLinksNotFlagged(t *testing.T) {
	content := "---\nname: url-test\ndescription: test\n---\n\n[Google](https://google.com)\n[HTTP](http://example.com)\n"
	dir := writeSkill(t, content)
	results := ValidateLayer1(dir, "url-test")
	k3s := filterByCheck(results, "K3")
	if len(k3s) != 0 {
		t.Fatalf("URL links should not be flagged, got %d K3 results", len(k3s))
	}
}

func TestValidateLayer1_AnchorLinksNotFlagged(t *testing.T) {
	content := "---\nname: anchor-test\ndescription: test\n---\n\n[Section](#section)\n"
	dir := writeSkill(t, content)
	results := ValidateLayer1(dir, "anchor-test")
	k3s := filterByCheck(results, "K3")
	if len(k3s) != 0 {
		t.Fatalf("anchor links should not be flagged, got %d K3 results", len(k3s))
	}
}

func TestParseFrontmatter_ValidYAML(t *testing.T) {
	content := []byte("---\nname: test\ndescription: hello\n---\n# Body\n")
	fm, raw, bodyStart, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Name != "test" {
		t.Fatalf("got name %q, want test", fm.Name)
	}
	if fm.Description != "hello" {
		t.Fatalf("got description %q, want hello", fm.Description)
	}
	if raw["name"] != "test" {
		t.Fatalf("raw map missing name")
	}
	if bodyStart < 4 {
		t.Fatalf("bodyStart %d seems too low", bodyStart)
	}
}

func TestParseFrontmatter_NoOpening(t *testing.T) {
	content := []byte("# Just markdown\n")
	_, _, _, err := ParseFrontmatter(content)
	if err == nil {
		t.Fatal("expected error for missing frontmatter")
	}
}

func TestParseFrontmatter_NoClosing(t *testing.T) {
	content := []byte("---\nname: test\n")
	_, _, _, err := ParseFrontmatter(content)
	if err == nil {
		t.Fatal("expected error for unclosed frontmatter")
	}
}

func TestParseFrontmatter_MultilineDescription(t *testing.T) {
	content := []byte("---\nname: test\ndescription: Line one\n  continued here.\n---\n# Body\n")
	fm, _, _, err := ParseFrontmatter(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(fm.Description, "Line one") {
		t.Fatalf("description should contain 'Line one', got %q", fm.Description)
	}
}

// --- helpers ---

func testdataDir(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join("testdata", name)
}

func writeSkill(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}
	return dir
}

func filterErrors(results []ValidationResult) []ValidationResult {
	var out []ValidationResult
	for _, r := range results {
		if r.Severity == SeverityError {
			out = append(out, r)
		}
	}
	return out
}

func filterByCheck(results []ValidationResult, checkID string) []ValidationResult {
	var out []ValidationResult
	for _, r := range results {
		if r.CheckID == checkID {
			out = append(out, r)
		}
	}
	return out
}

func assertHasCheck(t *testing.T, results []ValidationResult, checkID string, severity Severity) {
	t.Helper()
	for _, r := range results {
		if r.CheckID == checkID && r.Severity == severity {
			return
		}
	}
	t.Errorf("expected %s %s result, got: %v", checkID, severity, results)
}
