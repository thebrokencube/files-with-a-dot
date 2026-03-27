package main

import (
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"os"
	"path/filepath"
	"testing"
)

func TestRewriteTBDLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		newKey  string
		want    string
		changed bool
	}{
		{
			name:    "simple TBD",
			input:   "---\njira: TBD\ntype: Story\n---\n# Content\n",
			newKey:  "BEN-456",
			want:    "---\njira: BEN-456\ntype: Story\n---\n# Content\n",
			changed: true,
		},
		{
			name:    "quoted TBD",
			input:   "---\njira: \"TBD\"\ntype: Story\n---\n# Content\n",
			newKey:  "BEN-789",
			want:    "---\njira: BEN-789\ntype: Story\n---\n# Content\n",
			changed: true,
		},
		{
			name:    "single-quoted TBD",
			input:   "---\njira: 'TBD'\ntype: Story\n---\n# Content\n",
			newKey:  "TEST-1",
			want:    "---\njira: TEST-1\ntype: Story\n---\n# Content\n",
			changed: true,
		},
		{
			name:    "lowercase tbd",
			input:   "---\njira: tbd\ntype: Story\n---\n# Content\n",
			newKey:  "BEN-100",
			want:    "---\njira: BEN-100\ntype: Story\n---\n# Content\n",
			changed: true,
		},
		{
			name:    "other fields preserved",
			input:   "---\njira: TBD\ntype: Epic\nlabel: My Epic\norder: 3\n---\n# My Epic\n\nDescription here.\n",
			newKey:  "BEN-200",
			want:    "---\njira: BEN-200\ntype: Epic\nlabel: My Epic\norder: 3\n---\n# My Epic\n\nDescription here.\n",
			changed: true,
		},
		{
			name:    "content below frontmatter untouched",
			input:   "---\njira: TBD\n---\n# Heading\n\n---\n\nHorizontal rule above.\n",
			newKey:  "BEN-300",
			want:    "---\njira: BEN-300\n---\n# Heading\n\n---\n\nHorizontal rule above.\n",
			changed: true,
		},
		{
			name:    "no frontmatter",
			input:   "# Just content\nNo frontmatter here.\n",
			newKey:  "BEN-400",
			want:    "# Just content\nNo frontmatter here.\n",
			changed: false,
		},
		{
			name:    "no jira TBD line",
			input:   "---\njira: BEN-123\ntype: Story\n---\n# Content\n",
			newKey:  "BEN-500",
			want:    "---\njira: BEN-123\ntype: Story\n---\n# Content\n",
			changed: false,
		},
		{
			name:    "no closing fence",
			input:   "---\njira: TBD\ntype: Story\n# No closing fence\n",
			newKey:  "BEN-600",
			want:    "---\njira: TBD\ntype: Story\n# No closing fence\n",
			changed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed := rewriteTBDLine([]byte(tt.input), tt.newKey)
			if changed != tt.changed {
				t.Errorf("changed: got %v, want %v", changed, tt.changed)
			}
			if string(got) != tt.want {
				t.Errorf("output mismatch\n  got:  %q\n  want: %q", string(got), tt.want)
			}
		})
	}
}

func TestIsTBDLine(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"jira: TBD", true},
		{"jira: tbd", true},
		{"jira: \"TBD\"", true},
		{"jira: 'TBD'", true},
		{"jira:TBD", true},
		{"jira: BEN-123", false},
		{"type: Story", false},
		{"jira: TBDX", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isTBDLine(tt.input)
			if got != tt.want {
				t.Errorf("isTBDLine(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestRunCreateMissingDryRun(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n  project: TEST\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task-a.md"), []byte("---\njira: TBD\ntype: Epic\n---\n# New Epic\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task-b.md"), []byte("---\njira: TBD\ntype: Task\n---\n# New Task\n"), 0644)

	code := runCreateMissing([]string{"--dir", dir, "--dry-run"})
	if code != 0 {
		t.Fatalf("expected exit 0, got %d", code)
	}
}

func TestRunCreateMissingNoTBD(t *testing.T) {
	dir := t.TempDir()

	os.MkdirAll(filepath.Join(dir, ".jf"), 0755)
	os.WriteFile(filepath.Join(dir, ".jf", "forest.yml"), []byte("schema: 1\ndefaults:\n  sync: push\n  type: Story\n  project: TEST\n"), 0644)
	os.WriteFile(filepath.Join(dir, ".jf", "task-a.md"), []byte("---\njira: TEST-1\n---\n# Existing\n"), 0644)

	code := runCreateMissing([]string{"--dir", dir, "--dry-run"})
	if code != 0 {
		t.Fatalf("expected exit 0 for no TBD nodes, got %d", code)
	}
}

func TestRunCreateMissingNoForest(t *testing.T) {
	dir := t.TempDir()
	code := runCreateMissing([]string{"--dir", dir})
	if code != 1 {
		t.Fatalf("expected exit 1, got %d", code)
	}
}

func TestRewriteFrontmatterKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	os.WriteFile(path, []byte("---\njira: TBD\ntype: Story\n---\n# Hello\n"), 0644)

	err := rewriteFrontmatterKey(path, "BEN-999")
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	content, _ := os.ReadFile(path)
	want := "---\njira: BEN-999\ntype: Story\n---\n# Hello\n"
	if string(content) != want {
		t.Errorf("file content mismatch\n  got:  %q\n  want: %q", string(content), want)
	}
}

func TestBuildCreatePayload(t *testing.T) {
	f := &forest.Forest{
		Defaults: forest.ForestDefaults{Project: "TEST", Type: "Story"},
	}

	t.Run("root node", func(t *testing.T) {
		n := &forest.Node{Key: "TBD", Label: "My Task", Type: "Task"}
		payload := string(buildCreatePayload(n, f, nil))
		want := `{"projectKey":"TEST","summary":"My Task","type":"Task"}`
		if payload != want {
			t.Errorf("got %s, want %s", payload, want)
		}
	})

	t.Run("child node with real parent", func(t *testing.T) {
		parent := &forest.Node{Key: "TEST-10", Label: "Parent"}
		n := &forest.Node{Key: "TBD", Label: "Child Task", Type: "Story", Parent: parent}
		payload := string(buildCreatePayload(n, f, nil))
		want := `{"parentIssueId":"TEST-10","projectKey":"TEST","summary":"Child Task","type":"Story"}`
		if payload != want {
			t.Errorf("got %s, want %s", payload, want)
		}
	})

	t.Run("child node with TBD parent", func(t *testing.T) {
		parent := &forest.Node{Key: "TBD", Label: "TBD Parent"}
		n := &forest.Node{Key: "TBD", Label: "Child", Type: "Story", Parent: parent}
		payload := string(buildCreatePayload(n, f, nil))
		// TBD parent should NOT be included
		want := `{"projectKey":"TEST","summary":"Child","type":"Story"}`
		if payload != want {
			t.Errorf("got %s, want %s", payload, want)
		}
	})

	t.Run("with project fields", func(t *testing.T) {
		n := &forest.Node{Key: "TBD", Label: "My Task", Type: "Epic"}
		fields := map[string]any{
			"customfield_10028": map[string]any{"value": "Medium"},
		}
		payload := string(buildCreatePayload(n, f, fields))
		// additionalAttributes should be included
		want := `{"additionalAttributes":{"customfield_10028":{"value":"Medium"}},"projectKey":"TEST","summary":"My Task","type":"Epic"}`
		if payload != want {
			t.Errorf("got %s, want %s", payload, want)
		}
	})

	t.Run("with components formatted correctly", func(t *testing.T) {
		n := &forest.Node{Key: "TBD", Label: "My Task", Type: "Task"}
		fields := map[string]any{
			"components":        []any{"Partnered Benefits"},
			"customfield_10324": map[string]any{"value": "Bottoms Up Initiative"},
		}
		payload := string(buildCreatePayload(n, f, fields))
		// components should be in additionalAttributes with name format
		want := `{"additionalAttributes":{"components":[{"name":"Partnered Benefits"}],"customfield_10324":{"value":"Bottoms Up Initiative"}},"projectKey":"TEST","summary":"My Task","type":"Task"}`
		if payload != want {
			t.Errorf("got %s, want %s", payload, want)
		}
	})
}
