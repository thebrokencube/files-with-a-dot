package engine

import (
	"encoding/json"
	"testing"
)

func TestIsSubstantiveLocal(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty", "", false},
		{"whitespace_only", "   \n  ", false},
		{"tbd_upper", "TBD", false},
		{"todo_lower", "todo", false},
		{"wip_upper", "WIP", false},
		{"tbd_mixed_case", "Tbd", false},
		{"heading_only", "# Heading\n## Sub", false},
		{"heading_with_content", "# Heading\nContent here", true},
		{"plain_content", "Some real content", true},
		{"tbd_prefix_with_more", "TBD: but with more text", true},
		{"single_heading", "# Just a heading", false},
		{"headings_with_blank_lines", "# H1\n\n## H2\n\n### H3", false},
		{"content_after_headings", "# H1\n## H2\nActual content", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSubstantiveLocal([]byte(tt.input))
			if got != tt.want {
				t.Errorf("IsSubstantiveLocal(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsSubstantiveADF(t *testing.T) {
	tests := []struct {
		name  string
		input json.RawMessage
		want  bool
	}{
		{"nil", nil, false},
		{"json_null", json.RawMessage(`null`), false},
		{"empty_bytes", json.RawMessage(``), false},
		{"empty_content_array", json.RawMessage(`{"content":[]}`), false},
		{"empty_paragraph", json.RawMessage(`{"content":[{"type":"paragraph","content":[]}]}`), false},
		{"whitespace_text_substantive", json.RawMessage(`{"content":[{"type":"paragraph","content":[{"text":"  "}]}]}`), true},
		{"real_content", json.RawMessage(`{"content":[{"type":"paragraph","content":[{"text":"Hello"}]}]}`), true},
		{"non_paragraph_node", json.RawMessage(`{"content":[{"type":"table","content":[]}]}`), true},
		{"invalid_json", json.RawMessage(`invalid json`), true},
		{"mixed_with_non_paragraph", json.RawMessage(`{"content":[{"type":"paragraph","content":[{"text":"  "}]},{"type":"heading","content":[]}]}`), true},
		{"multiple_empty_paragraphs", json.RawMessage(`{"content":[{"type":"paragraph","content":[]},{"type":"paragraph","content":[]}]}`), false},
		{"paragraph_no_content_field", json.RawMessage(`{"content":[{"type":"paragraph"}]}`), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSubstantiveADF(tt.input)
			if got != tt.want {
				t.Errorf("IsSubstantiveADF(%s) = %v, want %v", string(tt.input), got, tt.want)
			}
		})
	}
}
