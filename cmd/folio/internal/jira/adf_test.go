package jira

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestConvertHeading(t *testing.T) {
	doc, err := Convert([]byte("## Hello"))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 node, got %d", len(doc.Content))
	}
	n := doc.Content[0]
	if n["type"] != "heading" {
		t.Errorf("expected heading, got %s", n["type"])
	}
	attrs := n["attrs"].(map[string]any)
	if attrs["level"] != 2 {
		t.Errorf("expected level 2, got %v", attrs["level"])
	}
}

func TestConvertBulletList(t *testing.T) {
	doc, err := Convert([]byte("- one\n- two"))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 node, got %d", len(doc.Content))
	}
	n := doc.Content[0]
	if n["type"] != "bulletList" {
		t.Errorf("expected bulletList, got %s", n["type"])
	}
	items := n["content"].([]any)
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestConvertOrderedList(t *testing.T) {
	doc, err := Convert([]byte("1. one\n2. two"))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 node, got %d", len(doc.Content))
	}
	n := doc.Content[0]
	if n["type"] != "orderedList" {
		t.Errorf("expected orderedList, got %s", n["type"])
	}
	items := n["content"].([]any)
	if len(items) != 2 {
		t.Errorf("expected 2 items, got %d", len(items))
	}
}

func TestConvertParagraph(t *testing.T) {
	doc, err := Convert([]byte("plain text"))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 node, got %d", len(doc.Content))
	}
	if doc.Content[0]["type"] != "paragraph" {
		t.Errorf("expected paragraph, got %s", doc.Content[0]["type"])
	}
}

func TestConvertBlankLines(t *testing.T) {
	doc, err := Convert([]byte("## A\n\n## B"))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Content) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(doc.Content))
	}
	for _, n := range doc.Content {
		if n["type"] != "heading" {
			t.Errorf("expected heading, got %s", n["type"])
		}
	}
}

func TestConvertMixed(t *testing.T) {
	input := "## Title\n\nSome text\n\n- bullet one\n- bullet two\n\n1. first\n2. second\n\nEnd paragraph"
	doc, err := Convert([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{"heading", "paragraph", "bulletList", "orderedList", "paragraph"}
	if len(doc.Content) != len(expected) {
		t.Fatalf("expected %d nodes, got %d", len(expected), len(doc.Content))
	}
	for i, want := range expected {
		if doc.Content[i]["type"] != want {
			t.Errorf("node %d: expected %s, got %s", i, want, doc.Content[i]["type"])
		}
	}
}

func TestConvertDocStructure(t *testing.T) {
	doc, err := Convert([]byte("## Hello"))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Version != 1 {
		t.Errorf("expected version 1, got %d", doc.Version)
	}
	if doc.Type != "doc" {
		t.Errorf("expected type doc, got %s", doc.Type)
	}
}

func TestParseInlineBold(t *testing.T) {
	segs := ParseInline("**bold**")
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	n := segs[0].(ADFNode)
	if n["text"] != "bold" {
		t.Errorf("expected bold, got %s", n["text"])
	}
	marks := n["marks"].([]any)
	m := marks[0].(ADFNode)
	if m["type"] != "strong" {
		t.Errorf("expected strong mark, got %s", m["type"])
	}
}

func TestParseInlineCode(t *testing.T) {
	segs := ParseInline("`code`")
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	n := segs[0].(ADFNode)
	if n["text"] != "code" {
		t.Errorf("expected code, got %s", n["text"])
	}
	marks := n["marks"].([]any)
	m := marks[0].(ADFNode)
	if m["type"] != "code" {
		t.Errorf("expected code mark, got %s", m["type"])
	}
}

func TestParseInlineLink(t *testing.T) {
	segs := ParseInline("[text](https://example.com)")
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	n := segs[0].(ADFNode)
	if n["text"] != "text" {
		t.Errorf("expected text, got %s", n["text"])
	}
	marks := n["marks"].([]any)
	m := marks[0].(ADFNode)
	if m["type"] != "link" {
		t.Errorf("expected link mark, got %s", m["type"])
	}
	attrs := m["attrs"].(map[string]any)
	if attrs["href"] != "https://example.com" {
		t.Errorf("expected https://example.com, got %s", attrs["href"])
	}
}

func TestParseInlineMixed(t *testing.T) {
	segs := ParseInline("Hello **bold** and `code`")
	if len(segs) != 4 {
		t.Fatalf("expected 4 segments, got %d: %v", len(segs), segs)
	}
	// "Hello " (plain), "bold" (strong), " and " (plain), "code" (code)
	n0 := segs[0].(ADFNode)
	if n0["text"] != "Hello " {
		t.Errorf("seg 0: expected 'Hello ', got %q", n0["text"])
	}
	n1 := segs[1].(ADFNode)
	if n1["text"] != "bold" {
		t.Errorf("seg 1: expected 'bold', got %q", n1["text"])
	}
	n2 := segs[2].(ADFNode)
	if n2["text"] != " and " {
		t.Errorf("seg 2: expected ' and ', got %q", n2["text"])
	}
	n3 := segs[3].(ADFNode)
	if n3["text"] != "code" {
		t.Errorf("seg 3: expected 'code', got %q", n3["text"])
	}
}

func TestParseInlinePlain(t *testing.T) {
	segs := ParseInline("just text")
	if len(segs) != 1 {
		t.Fatalf("expected 1 segment, got %d", len(segs))
	}
	n := segs[0].(ADFNode)
	if n["text"] != "just text" {
		t.Errorf("expected 'just text', got %q", n["text"])
	}
	if n["marks"] != nil {
		t.Errorf("expected no marks on plain text")
	}
}

func TestParseInlineEmpty(t *testing.T) {
	segs := ParseInline("")
	if segs != nil {
		t.Errorf("expected nil for empty, got %v", segs)
	}
	segs = ParseInline("   ")
	if segs != nil {
		t.Errorf("expected nil for whitespace, got %v", segs)
	}
}

func TestStripFrontmatterValid(t *testing.T) {
	input := "---\ntitle: foo\n---\ncontent here"
	result := StripFrontmatter([]byte(input))
	if string(result) != "content here" {
		t.Errorf("expected 'content here', got %q", string(result))
	}
}

func TestStripFrontmatterMultipleFields(t *testing.T) {
	input := "---\ntitle: foo\ndate: 2026-01-01\n---\ncontent"
	result := StripFrontmatter([]byte(input))
	if string(result) != "content" {
		t.Errorf("expected 'content', got %q", string(result))
	}
}

func TestStripFrontmatterNoClosing(t *testing.T) {
	input := "---\ntitle: foo\nno closing"
	result := StripFrontmatter([]byte(input))
	if string(result) != input {
		t.Error("expected input unchanged when no closing ---")
	}
}

func TestStripFrontmatterNoColon(t *testing.T) {
	input := "---\nno colon here\n---\ncontent"
	result := StripFrontmatter([]byte(input))
	if string(result) != input {
		t.Error("expected input unchanged when no colon between fences")
	}
}

func TestStripFrontmatterHR(t *testing.T) {
	input := "---\n\ntext after HR"
	result := StripFrontmatter([]byte(input))
	if string(result) != input {
		t.Error("expected input unchanged for horizontal rule (no colon)")
	}
}

func TestStripFrontmatterTooLong(t *testing.T) {
	lines := []string{"---"}
	for i := 0; i < 25; i++ {
		lines = append(lines, "key: value")
	}
	lines = append(lines, "---", "content")
	input := strings.Join(lines, "\n")
	result := StripFrontmatter([]byte(input))
	if string(result) != input {
		t.Error("expected input unchanged when closing --- is beyond line 19")
	}
}

func TestStripFrontmatterPreservesContent(t *testing.T) {
	input := "---\ntitle: test\n---\n## Heading\n\nParagraph"
	result := StripFrontmatter([]byte(input))
	if string(result) != "## Heading\n\nParagraph" {
		t.Errorf("expected content after frontmatter, got %q", string(result))
	}
}

func TestConvertWithFrontmatter(t *testing.T) {
	input := "---\ntitle: test\n---\n## Heading"
	doc, err := Convert([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 node after frontmatter strip, got %d", len(doc.Content))
	}
	if doc.Content[0]["type"] != "heading" {
		t.Errorf("expected heading, got %s", doc.Content[0]["type"])
	}
}

func TestConvertProducesValidJSON(t *testing.T) {
	input := "## Title\n\nHello **bold** and `code`\n\n- one\n- two"
	doc, err := Convert([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(b, &parsed); err != nil {
		t.Fatalf("produced invalid JSON: %v", err)
	}
	if parsed["version"] != float64(1) {
		t.Errorf("expected version 1, got %v", parsed["version"])
	}
	if parsed["type"] != "doc" {
		t.Errorf("expected type doc, got %v", parsed["type"])
	}
}
