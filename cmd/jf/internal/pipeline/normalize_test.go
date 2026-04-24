package pipeline

import (
	"testing"
)

func TestNormalizeMarkdown(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"trailing spaces", "hello  \nworld \n", "hello\nworld"},
		{"trailing tabs", "hello\t\nworld\t\n", "hello\nworld"},
		{"multiple blank lines", "a\n\n\n\nb\n", "a\n\nb"},
		{"three blank lines", "a\n\n\nb\n", "a\n\nb"},
		{"single blank line preserved", "a\n\nb\n", "a\n\nb"},
		{"leading trailing whitespace", "\n\n  hello  \n\n", "hello"},
		{"mixed", "# Title  \n\n\n\nParagraph \n\n\n\n## Next  \n", "# Title\n\nParagraph\n\n## Next"},
		{"empty", "", ""},
		{"only whitespace", "   \n\n\n  \n", ""},
		{"no changes needed", "clean\ncontent", "clean\ncontent"},

		// Block boundary normalization
		{"paragraph then list", "Hello\n- item\n", "Hello\n\n- item"},
		{"list then paragraph", "- item\nHello\n", "- item\n\nHello"},
		{"tight list stays tight", "- a\n- b\n- c\n", "- a\n- b\n- c"},
		{"heading then list", "## Title\n- item\n", "## Title\n\n- item"},
		{"already has blank", "Hello\n\n- item\n", "Hello\n\n- item"},
		{"ordered list marker", "Hello\n1. item\n", "Hello\n\n1. item"},
		{"star bullet marker", "Hello\n* item\n", "Hello\n\n* item"},
		{"plus bullet marker", "Hello\n+ item\n", "Hello\n\n+ item"},
		{"paren ordered marker", "Hello\n1) item\n", "Hello\n\n1) item"},
		{"list then heading", "- item\n## Title\n", "- item\n\n## Title"},
		{"bidirectional", "Hello\n- a\n- b\nWorld\n", "Hello\n\n- a\n- b\n\nWorld"},

		// Smart quote normalization
		{"smart single quotes", "it\u2019s a test\n", "it's a test"},
		{"smart double quotes", "\u201CHello\u201D\n", `"Hello"`},
		{"left single quote", "\u2018word\u2019\n", "'word'"},
		{"en dash", "pages 1\u20135\n", "pages 1-5"},
		{"em dash", "word\u2014another\n", "word--another"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(NormalizeMarkdown([]byte(tt.input)))
			if got != tt.want {
				t.Errorf("NormalizeMarkdown(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
