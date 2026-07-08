package main

import (
	"encoding/json"
	"os"
	"testing"
)

func TestRunSchemaInput(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := buildRoot().Execute([]string{"schema"})

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	buf := make([]byte, 16384)
	n, _ := r.Read(buf)

	// Verify it's valid JSON
	var parsed map[string]any
	if err := json.Unmarshal(buf[:n], &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %s", err)
	}

	// Check structure
	if parsed["title"] != "jf schemas" {
		t.Errorf("title: got %v, want %q", parsed["title"], "jf schemas")
	}

	defs, ok := parsed["definitions"].(map[string]any)
	if !ok {
		t.Fatal("missing definitions")
	}
	if _, ok := defs["forest.yml"]; !ok {
		t.Error("missing forest.yml definition")
	}
	if _, ok := defs["frontmatter"]; !ok {
		t.Error("missing frontmatter definition")
	}
}

func TestRunSchemaOutput(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := buildRoot().Execute([]string{"schema", "--output"})

	w.Close()
	os.Stdout = old

	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}

	buf := make([]byte, 16384)
	n, _ := r.Read(buf)

	var parsed map[string]any
	if err := json.Unmarshal(buf[:n], &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %s", err)
	}

	defs, ok := parsed["definitions"].(map[string]any)
	if !ok {
		t.Fatal("missing definitions")
	}
	for _, key := range []string{"NodeInfo", "StatusResult", "ValidateResult", "ErrorResult"} {
		if _, ok := defs[key]; !ok {
			t.Errorf("missing %s definition", key)
		}
	}
}
