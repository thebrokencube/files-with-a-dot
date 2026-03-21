package pipeline

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

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

func TestStripFrontmatter30Lines(t *testing.T) {
	lines := []string{"---"}
	for i := 0; i < 30; i++ {
		lines = append(lines, fmt.Sprintf("key%d: value%d", i, i))
	}
	lines = append(lines, "---", "content")
	input := strings.Join(lines, "\n")
	result := StripFrontmatter([]byte(input))
	if string(result) != "content" {
		t.Error("expected 30-line frontmatter to be stripped (limit is 50)")
	}
}

func TestStripFrontmatterTooLong(t *testing.T) {
	lines := []string{"---"}
	for i := 0; i < 55; i++ {
		lines = append(lines, "key: value")
	}
	lines = append(lines, "---", "content")
	input := strings.Join(lines, "\n")
	result := StripFrontmatter([]byte(input))
	if string(result) != input {
		t.Error("expected input unchanged when closing --- is beyond line 49")
	}
}

func TestStripFrontmatterPreservesContent(t *testing.T) {
	input := "---\ntitle: test\n---\n## Heading\n\nParagraph"
	result := StripFrontmatter([]byte(input))
	if string(result) != "## Heading\n\nParagraph" {
		t.Errorf("expected content after frontmatter, got %q", string(result))
	}
}

// --- CompileMarkdown integration tests ---

func requireNode(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not found, skipping integration test")
	}
}

func TestCompileMarkdownBasic(t *testing.T) {
	requireNode(t)
	raw, err := CompileMarkdown([]byte("## Hello\n\nSome text"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if doc["type"] != "doc" {
		t.Errorf("expected type doc, got %v", doc["type"])
	}
	if doc["version"] != float64(1) {
		t.Errorf("expected version 1, got %v", doc["version"])
	}
}

func TestCompileMarkdownWithFrontmatter(t *testing.T) {
	requireNode(t)
	raw, err := CompileMarkdown([]byte("---\ntitle: test\n---\n## Heading"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	content := doc["content"].([]any)
	if len(content) == 0 {
		t.Fatal("expected content after frontmatter strip")
	}
	first := content[0].(map[string]any)
	if first["type"] != "heading" {
		t.Errorf("expected heading, got %v", first["type"])
	}
}

func TestCompileMarkdownTable(t *testing.T) {
	requireNode(t)
	input := "| A | B |\n|---|---|\n| 1 | 2 |"
	raw, err := CompileMarkdown([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	content := doc["content"].([]any)
	if len(content) == 0 {
		t.Fatal("expected table content")
	}
	first := content[0].(map[string]any)
	if first["type"] != "table" {
		t.Errorf("expected table, got %v", first["type"])
	}
}

func TestCompileMarkdownCodeBlock(t *testing.T) {
	requireNode(t)
	input := "```go\nfmt.Println(\"hello\")\n```"
	raw, err := CompileMarkdown([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	content := doc["content"].([]any)
	if len(content) == 0 {
		t.Fatal("expected code block content")
	}
	first := content[0].(map[string]any)
	if first["type"] != "codeBlock" {
		t.Errorf("expected codeBlock, got %v", first["type"])
	}
}

func TestCompileMarkdownBlockquote(t *testing.T) {
	requireNode(t)
	raw, err := CompileMarkdown([]byte("> quoted text"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	content := doc["content"].([]any)
	if len(content) == 0 {
		t.Fatal("expected blockquote content")
	}
	first := content[0].(map[string]any)
	if first["type"] != "blockquote" {
		t.Errorf("expected blockquote, got %v", first["type"])
	}
}

func TestCompileMarkdownEmpty(t *testing.T) {
	requireNode(t)
	raw, err := CompileMarkdown([]byte(""))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if doc["type"] != "doc" {
		t.Errorf("expected type doc, got %v", doc["type"])
	}
}

func TestCompileMarkdownMixed(t *testing.T) {
	requireNode(t)
	input := "## Title\n\nSome **bold** and `code`\n\n- item one\n- item two\n\n> quote\n\n| A | B |\n|---|---|\n| 1 | 2 |"
	raw, err := CompileMarkdown([]byte(input))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	content := doc["content"].([]any)
	if len(content) < 4 {
		t.Errorf("expected at least 4 content nodes, got %d", len(content))
	}
}
