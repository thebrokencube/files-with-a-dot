package forest

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Forest represents a parsed forest.yml configuration.
type Forest struct {
	Schema   int            `yaml:"schema"`
	Defaults ForestDefaults `yaml:"defaults"`
	Acli     string         `yaml:"acli"` // version constraint, optional
	Dir      string         `yaml:"-"`    // runtime: directory containing forest.yml
}

// ForestDefaults are inherited by all nodes unless overridden.
type ForestDefaults struct {
	Sync    string `yaml:"sync"`    // push | pull
	Type    string `yaml:"type"`    // Jira issue type (for creation)
	Field   string `yaml:"field"`   // description | comment
	Project string `yaml:"project"` // Jira project key
}

// Node represents a single jira-forest node discovered from the filesystem.
type Node struct {
	Key      string  // from jira: field ("BEN-123" or "TBD")
	Label    string  // from label: field, first # heading, or filename stem
	Type     string  // from type: field or defaults
	Sync     string  // from sync: field or defaults (push | pull)
	Order    int     // from order: field (0 = unset, alphabetical)
	File     string  // relative path from forest root
	Parent   *Node   `yaml:"-" json:"-"` // nil for root nodes
	Children []*Node // child nodes
}

// Frontmatter represents the YAML frontmatter of a jira-forest node file.
type Frontmatter struct {
	Jira  string `yaml:"jira"`
	Label string `yaml:"label"`
	Type  string `yaml:"type"`
	Sync  string `yaml:"sync"`
	Order int    `yaml:"order"`
}

// ParseForestFile reads and parses a forest.yml file.
func ParseForestFile(path string) (*Forest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f Forest
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	f.Dir = filepath.Dir(path)
	return &f, nil
}

// ParseFrontmatter extracts jira-forest frontmatter from file contents.
// Returns nil if the file has no YAML frontmatter or no jira: field.
func ParseFrontmatter(content []byte) (*Frontmatter, error) {
	fm := extractFrontmatterBytes(content)
	if fm == nil {
		return nil, nil
	}

	var parsed Frontmatter
	if err := yaml.Unmarshal(fm, &parsed); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}

	if parsed.Jira == "" {
		return nil, nil
	}

	return &parsed, nil
}

// ParseFrontmatterFile reads a file and extracts its frontmatter.
// Returns nil if the file has no jira: frontmatter.
func ParseFrontmatterFile(path string) (*Frontmatter, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	fm, err := ParseFrontmatter(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return fm, nil
}

// DeriveLabel returns a display label for a node from frontmatter, file
// content, or filename stem (in that priority order).
func DeriveLabel(fm *Frontmatter, content []byte, filePath string) string {
	if fm != nil && fm.Label != "" {
		return fm.Label
	}

	if heading := firstHeading(content); heading != "" {
		return heading
	}

	stem := filepath.Base(filePath)
	stem = strings.TrimSuffix(stem, filepath.Ext(stem))
	if stem == "README" {
		stem = filepath.Base(filepath.Dir(filePath))
	}
	return stem
}

// extractFrontmatterBytes returns the raw YAML between --- fences, or nil.
func extractFrontmatterBytes(content []byte) []byte {
	lines := bytes.SplitN(content, []byte("\n"), -1)
	if len(lines) == 0 || strings.TrimSpace(string(lines[0])) != "---" {
		return nil
	}

	limit := len(lines)
	if limit > 50 {
		limit = 50
	}

	for i := 1; i < limit; i++ {
		if strings.TrimSpace(string(lines[i])) == "---" {
			hasColon := false
			for j := 1; j < i; j++ {
				if bytes.Contains(lines[j], []byte(":")) {
					hasColon = true
					break
				}
			}
			if !hasColon {
				return nil
			}
			return bytes.Join(lines[1:i], []byte("\n"))
		}
	}

	return nil
}

// firstHeading returns the text of the first # heading in the content.
func firstHeading(content []byte) string {
	scanner := bufio.NewScanner(bytes.NewReader(content))
	inFrontmatter := false
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "---" {
			inFrontmatter = !inFrontmatter
			continue
		}
		if inFrontmatter {
			continue
		}

		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(trimmed, "#"))
		}
	}
	return ""
}
