package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractExistingFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNil bool
		wantFM  string
	}{
		{
			name:   "basic frontmatter",
			input:  "---\njira: BEN-1\ntype: Story\n---\n# Content",
			wantFM: "---\njira: BEN-1\ntype: Story\n---\n",
		},
		{
			name:    "no frontmatter",
			input:   "# Just content\nNo frontmatter here",
			wantNil: true,
		},
		{
			name:    "no closing fence",
			input:   "---\njira: BEN-1\n# No closing fence",
			wantNil: true,
		},
		{
			name:   "minimal frontmatter",
			input:  "---\njira: BEN-1\n---\n",
			wantFM: "---\njira: BEN-1\n---\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractExistingFrontmatter([]byte(tt.input))
			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %q", string(got))
				}
				return
			}
			if got == nil {
				t.Fatal("expected frontmatter, got nil")
			}
			if string(got) != tt.wantFM {
				t.Errorf("expected %q, got %q", tt.wantFM, string(got))
			}
		})
	}
}

func TestMergeWithFrontmatter(t *testing.T) {
	dir := t.TempDir()

	t.Run("new file (no existing)", func(t *testing.T) {
		path := filepath.Join(dir, "new.md")
		pulled := []byte("Pulled content from Jira")
		result, err := mergeWithFrontmatter(path, pulled)
		if err != nil {
			t.Fatal(err)
		}
		if string(result) != "Pulled content from Jira" {
			t.Errorf("expected raw content for new file, got %q", string(result))
		}
	})

	t.Run("existing file with frontmatter", func(t *testing.T) {
		path := filepath.Join(dir, "existing.md")
		existing := "---\njira: BEN-1\nsync: pull\n---\n# Old Content\nOld text"
		if err := os.WriteFile(path, []byte(existing), 0644); err != nil {
			t.Fatal(err)
		}

		pulled := []byte("# New Content\nNew text from Jira")
		result, err := mergeWithFrontmatter(path, pulled)
		if err != nil {
			t.Fatal(err)
		}

		s := string(result)
		if !strings.HasPrefix(s, "---\njira: BEN-1\nsync: pull\n---\n") {
			t.Errorf("frontmatter not preserved: %q", s)
		}
		if !strings.Contains(s, "# New Content") {
			t.Error("pulled content not included")
		}
		if strings.Contains(s, "Old Content") {
			t.Error("old content should be replaced")
		}
	})

	t.Run("existing file without frontmatter", func(t *testing.T) {
		path := filepath.Join(dir, "nofront.md")
		if err := os.WriteFile(path, []byte("# Old stuff"), 0644); err != nil {
			t.Fatal(err)
		}

		pulled := []byte("# New stuff")
		result, err := mergeWithFrontmatter(path, pulled)
		if err != nil {
			t.Fatal(err)
		}
		if string(result) != "# New stuff" {
			t.Errorf("expected raw pulled content, got %q", string(result))
		}
	})
}
