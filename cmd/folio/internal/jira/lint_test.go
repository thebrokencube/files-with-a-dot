package jira

import (
	"testing"
)

func TestLintClean(t *testing.T) {
	input := "## Heading\n\n- bullet\n\nparagraph"
	issues := Lint([]byte(input), "test.md")
	if len(issues) != 0 {
		t.Errorf("expected 0 issues, got %d: %v", len(issues), issues)
	}
}

func TestLintH1(t *testing.T) {
	issues := Lint([]byte("# Bad heading"), "test.md")
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Line != 1 {
		t.Errorf("expected line 1, got %d", issues[0].Line)
	}
	if issues[0].Message != "h1 heading not supported (only ## supported)" {
		t.Errorf("unexpected message: %s", issues[0].Message)
	}
}

func TestLintH3(t *testing.T) {
	issues := Lint([]byte("### Bad heading"), "test.md")
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Message != "level 3+ heading not supported (only ## supported)" {
		t.Errorf("unexpected message: %s", issues[0].Message)
	}
}

func TestLintH4(t *testing.T) {
	issues := Lint([]byte("#### Also bad"), "test.md")
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Message != "level 3+ heading not supported (only ## supported)" {
		t.Errorf("unexpected message: %s", issues[0].Message)
	}
}

func TestLintCodeBlock(t *testing.T) {
	input := "```\ncode here\n```"
	issues := Lint([]byte(input), "test.md")
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d: %v", len(issues), issues)
	}
	if issues[0].Line != 1 {
		t.Errorf("expected line 1, got %d", issues[0].Line)
	}
	if issues[0].Message != "code block (fenced) not supported" {
		t.Errorf("unexpected message: %s", issues[0].Message)
	}
}

func TestLintCodeBlockContentsSkipped(t *testing.T) {
	// Content inside code blocks should NOT trigger additional issues
	input := "```\n# h1 inside code\n| table | inside |\n```"
	issues := Lint([]byte(input), "test.md")
	if len(issues) != 1 {
		t.Errorf("expected 1 issue (code block opener only), got %d: %v", len(issues), issues)
	}
}

func TestLintTable(t *testing.T) {
	issues := Lint([]byte("| a | b |"), "test.md")
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Message != "table syntax not supported" {
		t.Errorf("unexpected message: %s", issues[0].Message)
	}
}

func TestLintBlockquote(t *testing.T) {
	issues := Lint([]byte("> quote"), "test.md")
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Message != "blockquote not supported" {
		t.Errorf("unexpected message: %s", issues[0].Message)
	}
}

func TestLintNestedBullet(t *testing.T) {
	issues := Lint([]byte("  - nested"), "test.md")
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Message != "nested list not supported" {
		t.Errorf("unexpected message: %s", issues[0].Message)
	}
}

func TestLintNestedOrdered(t *testing.T) {
	issues := Lint([]byte("  1. nested"), "test.md")
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Message != "nested list not supported" {
		t.Errorf("unexpected message: %s", issues[0].Message)
	}
}

func TestLintTabNestedList(t *testing.T) {
	issues := Lint([]byte("\t- tab nested"), "test.md")
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Message != "nested list not supported" {
		t.Errorf("unexpected message: %s", issues[0].Message)
	}
}

func TestLintCheckbox(t *testing.T) {
	// Checkbox triggers both "checkbox" and "bare [" (the [ in [ ] survives link stripping)
	issues := Lint([]byte("- [ ] todo"), "test.md")
	if len(issues) != 2 {
		t.Fatalf("expected 2 issues (checkbox + bare [), got %d: %v", len(issues), issues)
	}
	if issues[0].Message != "checkbox not supported" {
		t.Errorf("unexpected message: %s", issues[0].Message)
	}
}

func TestLintCheckedCheckbox(t *testing.T) {
	issues := Lint([]byte("- [x] done"), "test.md")
	if len(issues) != 2 {
		t.Fatalf("expected 2 issues (checkbox + bare [), got %d: %v", len(issues), issues)
	}
}

func TestLintImage(t *testing.T) {
	issues := Lint([]byte("![alt](https://example.com/img.png)"), "test.md")
	found := false
	for _, i := range issues {
		if i.Message == "image syntax not supported (![alt](url) produces a plain link, not an image)" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected image issue, got %v", issues)
	}
}

func TestLintBareLink(t *testing.T) {
	issues := Lint([]byte("see [this"), "test.md")
	found := false
	for _, i := range issues {
		if i.Message == "bare [ or relative link not supported (only [text](https://...) links)" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected bare [ issue, got %v", issues)
	}
}

func TestLintValidLink(t *testing.T) {
	issues := Lint([]byte("see [text](https://example.com)"), "test.md")
	if len(issues) != 0 {
		t.Errorf("expected 0 issues for valid link, got %d: %v", len(issues), issues)
	}
}

func TestLintHTTPLink(t *testing.T) {
	issues := Lint([]byte("see [text](http://example.com)"), "test.md")
	if len(issues) != 0 {
		t.Errorf("expected 0 issues for http link, got %d: %v", len(issues), issues)
	}
}

func TestLintCodeSpanExclusion(t *testing.T) {
	issues := Lint([]byte("`![not image]`"), "test.md")
	if len(issues) != 0 {
		t.Errorf("expected 0 issues (inside code span), got %d: %v", len(issues), issues)
	}
}

func TestLintMultipleIssues(t *testing.T) {
	input := "# h1 bad\n\n### h3 bad\n\n> quote\n\n| a | b |"
	issues := Lint([]byte(input), "test.md")
	if len(issues) != 4 {
		t.Errorf("expected 4 issues, got %d: %v", len(issues), issues)
	}
	// Verify line numbers are correct
	expectedLines := []int{1, 3, 5, 7}
	for i, want := range expectedLines {
		if i < len(issues) && issues[i].Line != want {
			t.Errorf("issue %d: expected line %d, got %d", i, want, issues[i].Line)
		}
	}
}

func TestLintRelativeLink(t *testing.T) {
	issues := Lint([]byte("[text](./relative.md)"), "test.md")
	found := false
	for _, i := range issues {
		if i.Message == "bare [ or relative link not supported (only [text](https://...) links)" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected bare [ issue for relative link, got %v", issues)
	}
}

func TestLintH2Allowed(t *testing.T) {
	issues := Lint([]byte("## This is fine"), "test.md")
	if len(issues) != 0 {
		t.Errorf("expected 0 issues for h2, got %d: %v", len(issues), issues)
	}
}
