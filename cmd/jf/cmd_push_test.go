package main

import (
	"testing"
)

func TestBuildPlainTextPayload(t *testing.T) {
	source := []byte("---\njira: BEN-1\n---\n# Hello\n\nSome content")
	payload := buildPlainTextPayload("BEN-1", source)
	s := string(payload)

	// Should contain the key
	if want := `"BEN-1"`; !contains(s, want) {
		t.Errorf("payload missing key %s", want)
	}

	// Should strip frontmatter from text
	if contains(s, "jira:") {
		t.Error("payload should not contain frontmatter")
	}

	// Should contain the content
	if !contains(s, "Hello") {
		t.Error("payload should contain heading text")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
