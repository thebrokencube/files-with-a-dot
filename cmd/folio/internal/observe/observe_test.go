package observe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

func testContext(dir string) config.Context {
	return config.Context{FolioPath: filepath.Join(dir, "folio.yml")}
}

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

func makeRemoveTestFile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "folio.yml")
	os.WriteFile(path, []byte(`schema: 2
project: "Test"
observations:
  - "bug(cli): first item"
  - "task(plan): second item"
  - "idea(cli): third item"
`), 0644)
	return path
}

func removeTestItems() []string {
	return []string{
		"bug(cli): first item",
		"task(plan): second item",
		"idea(cli): third item",
	}
}

func TestResolveByIndexAndRemove(t *testing.T) {
	path := makeRemoveTestFile(t)
	items := removeTestItems()
	remaining, removed, err := Resolve(items, []string{"#2"})
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if len(removed) != 1 || removed[0] != 1 {
		t.Errorf("removed = %v, want [1]", removed)
	}
	if len(remaining) != 2 || remaining[0] != items[0] || remaining[1] != items[2] {
		t.Errorf("remaining = %v, want first and third items", remaining)
	}
	removedItems, err := RemoveIndices(path, items, removed)
	if err != nil {
		t.Fatalf("unexpected remove error: %v", err)
	}
	if len(removedItems) != 1 || removedItems[0] != items[1] {
		t.Errorf("removed items = %v, want [%q]", removedItems, items[1])
	}
	data, _ := os.ReadFile(path)
	content := string(data)
	if strings.Contains(content, "second item") {
		t.Error("second item should be removed from file")
	}
	if !strings.Contains(content, "first item") || !strings.Contains(content, "third item") {
		t.Error("other items should be preserved")
	}
}

func TestResolveBySubstringAndRemove(t *testing.T) {
	path := makeRemoveTestFile(t)
	items := removeTestItems()
	_, removed, err := Resolve(items, []string{"second"})
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	removedItems, err := RemoveIndices(path, items, removed)
	if err != nil {
		t.Fatalf("unexpected remove error: %v", err)
	}
	if len(removedItems) != 1 || removedItems[0] != items[1] {
		t.Errorf("removed items = %v, want [%q]", removedItems, items[1])
	}
}

func TestResolveAmbiguous(t *testing.T) {
	_, _, err := Resolve(removeTestItems(), []string{"cli"})
	if err == nil {
		t.Error("expected ambiguity error")
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Errorf("error should mention ambiguity: %v", err)
	}
}

func TestResolveMultiple(t *testing.T) {
	items := removeTestItems()
	remaining, removed, err := Resolve(items, []string{"#1", "#3"})
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if len(removed) != 2 || removed[0] != 0 || removed[1] != 2 {
		t.Errorf("removed = %v, want [0 2]", removed)
	}
	if len(remaining) != 1 || remaining[0] != items[1] {
		t.Errorf("remaining = %v, want second item", remaining)
	}
}

func TestResolveNotFound(t *testing.T) {
	_, _, err := Resolve(removeTestItems(), []string{"nonexistent"})
	if err == nil {
		t.Error("expected not-found error")
	}
}

func TestRemoveIndicesRefusesChangedManifest(t *testing.T) {
	path := makeRemoveTestFile(t)
	items := removeTestItems()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	data = []byte(strings.Replace(string(data), "second item", `second "changed" item`, 1))
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	before := string(data)
	_, err = RemoveIndices(path, items, []int{1})
	if err == nil {
		t.Fatal("expected changed-manifest error")
	}
	after, _ := os.ReadFile(path)
	if string(after) != before {
		t.Fatal("changed manifest must not be written")
	}
}

func TestLintValid(t *testing.T) {
	dir := t.TempDir()
	items := []string{
		"bug(cli): something broken",
		"task(plan): action item",
	}
	issues := Lint(items, testContext(dir))
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d: %v", len(issues), issues)
	}
}

func TestLintMalformed(t *testing.T) {
	dir := t.TempDir()
	items := []string{
		"freeform text",
		"bug(cli): valid item",
	}
	issues := Lint(items, testContext(dir))
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Index != 1 {
		t.Errorf("issue index = %d, want 1", issues[0].Index)
	}
	if issues[0].Reason != "malformed format" {
		t.Errorf("reason = %q, want malformed format", issues[0].Reason)
	}
}

func TestLintBrokenPath(t *testing.T) {
	dir := t.TempDir()
	items := []string{
		"bug(cli): broken ref (reference/missing.md)",
	}
	issues := Lint(items, testContext(dir))
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if !strings.Contains(issues[0].Reason, "broken path") {
		t.Errorf("reason = %q, want broken path", issues[0].Reason)
	}
}

func TestLintSkipsURLs(t *testing.T) {
	dir := t.TempDir()
	items := []string{
		"idea(cli): see docs (https://example.com/path)",
	}
	issues := Lint(items, testContext(dir))
	if len(issues) != 0 {
		t.Errorf("expected no issues for URL, got %d: %v", len(issues), issues)
	}
}

func TestLintSkipsProseInParens(t *testing.T) {
	dir := t.TempDir()
	// Parentheticals that are prose, not path refs: one with whitespace, one a
	// bare dotted filename. Neither should be flagged as a broken path.
	items := []string{
		"task(session-isolation): workspace shared jj repos (esp ~/.dotfiles); codify as a rule (dotfiles-awareness.md)",
	}
	issues := Lint(items, testContext(dir))
	if len(issues) != 0 {
		t.Errorf("expected no issues for prose parentheticals, got %d: %v", len(issues), issues)
	}
}

func TestLintStripsLineReferences(t *testing.T) {
	dir := t.TempDir()
	// A bare "file.ext:NN" is a code reference, not a "<store>:<path>" ref — the colon must not make
	// it look like an unknown store. A line range and a bare ":NN-NN" are the same class.
	items := []string{
		"bug(review): the transcript is Redis-only (activity_log.rb:7) so it expires (Show.tsx:393)",
		"bug(ui): the empty state renders wrong (ActivityStream.tsx:18-21) and again (:87-98)",
	}
	issues := Lint(items, testContext(dir))
	if len(issues) != 0 {
		t.Errorf("expected no issues for line references, got %d: %v", len(issues), issues)
	}
}

func TestLintChecksPathHalfOfALineReference(t *testing.T) {
	dir := t.TempDir()
	// Stripping the line suffix must not stop the path itself being verified.
	items := []string{
		"bug(cli): broken ref (reference/missing.md:12)",
	}
	issues := Lint(items, testContext(dir))
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
	}
	if !strings.Contains(issues[0].Reason, "broken path") {
		t.Errorf("reason = %q, want broken path", issues[0].Reason)
	}
}

func TestLintSkipsProseSlashRunsAndNamespaces(t *testing.T) {
	dir := t.TempDir()
	// Prose the parenthetical matcher captures but which are not path refs: a
	// slash-delimited word run (no extension) and a "::" language namespace.
	// Neither should be flagged (regression for validation false-positives that
	// blocked pushes and forced hand-edits to unrelated projects).
	items := []string{
		"idea(ui): show states (active/completed/failed/skipped) not the run breakdown",
		"idea(color): give each concept (grove/canopies/wardens/repos) one registry color",
		"idea(jobs): Gloppy sweep (CodeReviews::CodeReviewJob) runs every 10 min",
	}
	issues := Lint(items, testContext(dir))
	if len(issues) != 0 {
		t.Errorf("expected no issues for prose slash-runs and :: namespaces, got %d: %v", len(issues), issues)
	}
}

func TestLintSeePath(t *testing.T) {
	dir := t.TempDir()
	items := []string{
		"task(plan): needs work. See reference/design/missing.md",
	}
	issues := Lint(items, testContext(dir))
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if !strings.Contains(issues[0].Reason, "broken path") {
		t.Errorf("reason = %q, want broken path", issues[0].Reason)
	}
}

func TestLintPositionalReferencesAreWarnings(t *testing.T) {
	items := []string{
		"idea(cli): obs #1 is stale",
		"idea(cli): observation #2 is stale",
		"idea(cli): refines #3",
		"idea(cli): confirms #4",
		"idea(cli): per #5",
		"idea(cli): see #6",
		"idea(cli): literal #7 is not classified",
	}
	issues := Lint(items, testContext(t.TempDir()))
	if len(issues) != 6 {
		t.Fatalf("issues = %d, want six positional warnings: %v", len(issues), issues)
	}
	for _, issue := range issues {
		if issue.Severity != SeverityWarning {
			t.Errorf("issue severity = %d, want warning", issue.Severity)
		}
		if issue.Code != "observation.positional-reference" {
			t.Errorf("issue code = %q, want positional-reference", issue.Code)
		}
	}
}

func TestLintRecognizesExtensionlessWorkReferences(t *testing.T) {
	dir := t.TempDir()
	workPath := filepath.Join(dir, "work", "active", "topic", "notes")
	if err := os.MkdirAll(filepath.Dir(workPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(workPath, []byte("notes"), 0644); err != nil {
		t.Fatal(err)
	}
	items := []string{
		"idea(cli): use (work/active/topic/notes)",
		"idea(cli): prose (active/completed/failed/skipped)",
	}
	issues := Lint(items, testContext(dir))
	if len(issues) != 0 {
		t.Fatalf("issues = %v, want existing extensionless work path and prose to pass", issues)
	}
}

func TestDuplicateCandidatesNormalizeWhitespaceWithoutChangingCase(t *testing.T) {
	existing := []string{
		"idea(cli): keep this",
		"idea(cli): Keep this",
	}
	candidates := DuplicateCandidates(" idea(cli):  keep   this ", existing)
	if len(candidates) != 1 || candidates[0].Index != 1 {
		t.Fatalf("candidates = %+v, want only first exact normalized match", candidates)
	}
}

func TestRemoveIndicesDecodesQuotedScalars(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "folio.yml")
	items := []string{`idea(cli): quote "inside"`}
	content := "schema: 2\nproject: Test\nobservations:\n  - \"idea(cli): quote \\\"inside\\\"\"\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := RemoveIndices(path, items, []int{0})
	if err != nil {
		t.Fatalf("unexpected remove error: %v", err)
	}
	after, _ := os.ReadFile(path)
	if strings.Contains(string(after), "quote") {
		t.Fatalf("quoted scalar was not removed:\n%s", after)
	}
}
