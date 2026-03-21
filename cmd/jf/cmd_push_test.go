package main

import (
	"strings"
	"testing"
)

func TestBuildPlainTextPayload(t *testing.T) {
	source := []byte("---\njira: BEN-1\n---\n# Hello\n\nSome content")
	payload := buildPlainTextPayload("BEN-1", source)
	s := string(payload)

	if !strings.Contains(s, `"BEN-1"`) {
		t.Error("payload missing key")
	}

	if strings.Contains(s, "jira:") {
		t.Error("payload should not contain frontmatter")
	}

	if !strings.Contains(s, "Hello") {
		t.Error("payload should contain heading text")
	}
}
