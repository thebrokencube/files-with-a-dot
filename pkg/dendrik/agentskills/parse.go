package agentskills

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

// ParseFrontmatter extracts and parses YAML frontmatter from SKILL.md content.
// Returns the parsed frontmatter, the raw field map (for extra-fields detection), the body start line, and any error.
func ParseFrontmatter(content []byte) (*SkillFrontmatter, map[string]any, int, error) {
	// Find frontmatter delimiters (--- ... ---)
	if !bytes.HasPrefix(content, []byte("---\n")) && !bytes.HasPrefix(content, []byte("---\r\n")) {
		return nil, nil, 0, fmt.Errorf("SKILL.md does not start with YAML frontmatter (---)")
	}

	// Find closing ---
	rest := content[4:] // skip opening "---\n"
	lineNum := 2        // first line of frontmatter content
	idx := bytes.Index(rest, []byte("\n---"))
	if idx < 0 {
		return nil, nil, 0, fmt.Errorf("SKILL.md frontmatter not closed (missing closing ---)")
	}

	fmBytes := rest[:idx]

	// Count lines to find body start
	bodyStart := lineNum
	for _, b := range fmBytes {
		if b == '\n' {
			bodyStart++
		}
	}
	bodyStart += 2 // closing --- line + next line

	// Parse into struct
	var fm SkillFrontmatter
	if err := yaml.Unmarshal(fmBytes, &fm); err != nil {
		return nil, nil, 0, fmt.Errorf("invalid YAML frontmatter: %w", err)
	}

	// Parse into raw map for skill-extra-fields detection
	var rawMap map[string]any
	if err := yaml.Unmarshal(fmBytes, &rawMap); err != nil {
		return nil, nil, 0, fmt.Errorf("invalid YAML frontmatter: %w", err)
	}

	return &fm, rawMap, bodyStart, nil
}
