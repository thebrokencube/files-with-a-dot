package pipeline

import "testing"

func TestCheckRoundtrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"simple paragraph", "Hello world\n", true},
		{"h2 heading", "## Title\n\nBody text\n", true},
		{"bold text", "Some **bold** text\n", true},
		{"italic text", "Some *italic* text\n", true},
		{"inline code", "Use `code` here\n", true},
		{"unordered list", "- item one\n- item two\n", true},
		{"ordered list", "1. first\n2. second\n", true},
		{"link", "[text](https://example.com)\n", true},
		{"horizontal rule", "---\n", true},
		{"multiple paragraphs", "First paragraph.\n\nSecond paragraph.\n", true},
		{"heading and list", "## Title\n\n- one\n- two\n", true},
		{"strikethrough", "Some ~~deleted~~ text\n", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckRoundtrip([]byte(tt.input))
			if err != nil {
				t.Fatalf("CheckRoundtrip error: %v", err)
			}
			if got != tt.want {
				t.Errorf("CheckRoundtrip(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
