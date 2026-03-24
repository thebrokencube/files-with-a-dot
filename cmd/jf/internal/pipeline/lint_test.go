package pipeline

import (
	"testing"
)

func TestLint(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantN   int
		wantMsg string // substring in first issue message (empty = no check)
	}{
		// h1
		{"h1 rejected", "# Title\n", 1, "h1 heading"},
		{"h2 allowed", "## Title\n", 0, ""},
		{"h1 bare hash", "#\n", 1, "h1 heading"},

		// h3+
		{"h3 rejected", "### Title\n", 1, "h3+"},
		{"h4 rejected", "#### Title\n", 1, "h3+"},
		{"h2 allowed in heading check", "## OK\n", 0, ""},

		// Tables
		{"table rejected", "| a | b |\n", 1, "table"},
		{"table separator", "|---|---|\n", 1, "table"},

		// Fenced code
		{"fenced code rejected", "```go\nfmt.Println()\n```\n", 2, "code block"},
		{"backtick in text ok", "use `code` inline\n", 0, ""},

		// Blockquote
		{"blockquote rejected", "> quoted text\n", 1, "blockquote"},

		// Nested list
		{"nested list rejected", "- top\n  - nested\n", 1, "nested list"},
		{"flat list allowed", "- one\n- two\n", 0, ""},
		{"deep nested rejected", "    - deep\n", 1, "nested list"},
		{"nested ordered", "  1. nested\n", 1, "nested list"},

		// Checkbox
		{"checkbox rejected", "- [ ] unchecked\n", 1, "checkbox"},
		{"checked checkbox rejected", "- [x] checked\n", 1, "checkbox"},

		// Image
		{"image rejected", "![alt](img.png)\n", 1, "image"},

		// Relative link
		{"relative link rejected", "[text](./local.md)\n", 1, "relative link"},
		{"relative link no prefix", "[text](local.md)\n", 1, "relative link"},
		{"absolute link allowed", "[text](https://example.com)\n", 0, ""},
		{"http link allowed", "[text](http://example.com)\n", 0, ""},

		// Bare bracket
		{"bare bracket rejected", "some [incomplete text\n", 1, "bare bracket"},
		{"complete link ok", "[text](https://example.com)\n", 0, ""},

		// Clean content
		{"clean paragraph", "Hello world\n", 0, ""},
		{"clean h2 and list", "## Title\n\n- item one\n- item two\n", 0, ""},
		{"clean bold", "Some **bold** text\n", 0, ""},
		{"clean link", "[click here](https://example.com)\n", 0, ""},
		{"empty input", "", 0, ""},

		// Multiple issues
		{"multiple issues", "# Bad\n| table |\n> quote\n", 3, "h1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := Lint([]byte(tt.input), "test.md")
			if len(issues) != tt.wantN {
				t.Errorf("got %d issues, want %d: %v", len(issues), tt.wantN, issues)
			}
			if tt.wantMsg != "" && len(issues) > 0 {
				found := false
				for _, iss := range issues {
					if contains(iss.Message, tt.wantMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected message containing %q, got %v", tt.wantMsg, issues)
				}
			}
		})
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestFormatLintIssues(t *testing.T) {
	issues := []LintIssue{
		{1, "h1 heading not supported"},
		{5, "table syntax not supported"},
	}
	got := FormatLintIssues(issues)
	want := "line 1: h1 heading not supported; line 5: table syntax not supported"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
