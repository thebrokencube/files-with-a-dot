package agentskills

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var namePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)
var mdLinkPattern = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
var kebabPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*\.[a-z]+$`)

//dendrik:block agentskills
//dendrik:kind component
//dendrik:layer skill
//dendrik:status shipped
//dendrik:definition skill-layer validator — frontmatter, links, size gate (Layer 1)
//dendrik:intent one validator so every agent skill is well-formed and discoverable
//dendrik:conformance skill-frontmatter skill-links skill-size

// ValidateLayer1 runs Layer 1 checks (skill-exists through skill-size) against a skill directory.
// skillDir is the absolute path to the directory containing SKILL.md.
// dirName is the expected skill name (must match frontmatter name field).
func ValidateLayer1(skillDir, dirName string) []ValidationResult {
	var results []ValidationResult

	// skill-exists: SKILL.md exists
	skillPath := filepath.Join(skillDir, "SKILL.md")
	content, err := os.ReadFile(skillPath)
	if err != nil {
		results = append(results, ValidationResult{
			CheckID:     "skill-exists",
			Severity:    SeverityError,
			Message:     "SKILL.md not found",
			File:        "skill/SKILL.md",
			Remediation: "Create skill/SKILL.md with YAML frontmatter containing `name` and `description`.",
		})
		return results // Structure gate: skip remaining checks
	}

	// Parse frontmatter
	fm, rawMap, bodyStart, parseErr := ParseFrontmatter(content)
	if parseErr != nil {
		results = append(results, ValidationResult{
			CheckID:     "skill-frontmatter",
			Severity:    SeverityError,
			Message:     fmt.Sprintf("frontmatter parse error: %s", parseErr),
			File:        "skill/SKILL.md",
			Line:        1,
			Remediation: "Fix YAML frontmatter syntax — ensure SKILL.md starts with `---`, has valid YAML, and ends with `---`.",
		})
		return results
	}

	// skill-frontmatter: Valid frontmatter fields
	results = append(results, validateName(fm.Name, dirName)...)
	results = append(results, validateDescription(fm.Description)...)

	// skill-extra-fields: No unexpected fields
	results = append(results, validateUnexpectedFields(rawMap)...)

	// skill-links: Standard markdown links resolve
	results = append(results, validateMarkdownLinks(content, skillDir)...)

	// ref-naming: Reference file naming
	refsDir := filepath.Join(skillDir, "references")
	results = append(results, validateReferenceNaming(refsDir)...)

	// skill-size: Size constraints
	results = append(results, validateSize(content, bodyStart)...)

	return results
}

// ValidatePortable applies Layer 1 plus the standard Agent Skills frontmatter boundary.
func ValidatePortable(skillDir, dirName string) []ValidationResult {
	results := ValidateLayer1(skillDir, dirName)
	content, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		return results
	}
	_, rawMap, _, err := ParseFrontmatter(content)
	if err != nil {
		return results
	}
	for field := range rawMap {
		if !PortableFields[field] {
			results = append(results, ValidationResult{
				CheckID: "portable-skill-fields", Severity: SeverityError,
				Message: fmt.Sprintf("harness-specific frontmatter field in portable skill: %q", field),
				File:    "SKILL.md", Remediation: "Move the field to a native adapter or remove it.",
			})
		}
	}
	return results
}

func validateName(name, dirName string) []ValidationResult {
	var results []ValidationResult

	if name == "" {
		results = append(results, ValidationResult{
			CheckID:     "skill-frontmatter",
			Severity:    SeverityError,
			Message:     "frontmatter missing required field: name",
			File:        "skill/SKILL.md",
			Line:        2,
			Remediation: "Add `name:` field to SKILL.md frontmatter (lowercase, hyphens, 1-64 chars).",
		})
		return results
	}

	if len(name) > 64 {
		results = append(results, ValidationResult{
			CheckID:     "skill-frontmatter",
			Severity:    SeverityError,
			Message:     fmt.Sprintf("name exceeds 64 characters (%d)", len(name)),
			File:        "skill/SKILL.md",
			Line:        2,
			Remediation: "Shorten the `name` field to 64 characters or fewer.",
		})
	}

	if !namePattern.MatchString(name) {
		results = append(results, ValidationResult{
			CheckID:     "skill-frontmatter",
			Severity:    SeverityError,
			Message:     fmt.Sprintf("name %q does not match pattern [a-z0-9-]", name),
			File:        "skill/SKILL.md",
			Line:        2,
			Remediation: "Use only lowercase letters, digits, and hyphens in the `name` field.",
		})
	}

	if name != dirName {
		results = append(results, ValidationResult{
			CheckID:     "skill-frontmatter",
			Severity:    SeverityError,
			Message:     fmt.Sprintf("name %q does not match directory name %q", name, dirName),
			File:        "skill/SKILL.md",
			Line:        2,
			Remediation: fmt.Sprintf("Change `name:` to %q to match the directory name.", dirName),
		})
	}

	return results
}

func validateDescription(desc string) []ValidationResult {
	var results []ValidationResult

	if desc == "" {
		results = append(results, ValidationResult{
			CheckID:     "skill-frontmatter",
			Severity:    SeverityError,
			Message:     "frontmatter missing required field: description",
			File:        "skill/SKILL.md",
			Line:        3,
			Remediation: "Add `description:` field to SKILL.md frontmatter (1-1024 chars).",
		})
		return results
	}

	if len(desc) > 1024 {
		results = append(results, ValidationResult{
			CheckID:     "skill-frontmatter",
			Severity:    SeverityError,
			Message:     fmt.Sprintf("description exceeds 1024 characters (%d)", len(desc)),
			File:        "skill/SKILL.md",
			Line:        3,
			Remediation: "Shorten the `description` field to 1024 characters or fewer.",
		})
	}

	return results
}

func validateUnexpectedFields(rawMap map[string]any) []ValidationResult {
	var results []ValidationResult
	for field := range rawMap {
		if !KnownFields[field] {
			results = append(results, ValidationResult{
				CheckID:     "skill-extra-fields",
				Severity:    SeverityWarning,
				Message:     fmt.Sprintf("unexpected frontmatter field: %q", field),
				File:        "skill/SKILL.md",
				Remediation: fmt.Sprintf("Remove or document the %q field — it is not part of the Agent Skills spec.", field),
			})
		}
	}
	return results
}

func validateMarkdownLinks(content []byte, skillDir string) []ValidationResult {
	var results []ValidationResult
	lines := bytes.Split(content, []byte("\n"))

	for i, line := range lines {
		matches := mdLinkPattern.FindAllSubmatch(line, -1)
		for _, match := range matches {
			target := string(match[2])
			// Skip URLs
			if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
				continue
			}
			// Skip anchors
			if strings.HasPrefix(target, "#") {
				continue
			}
			// Strip anchor fragment from path
			if idx := strings.Index(target, "#"); idx >= 0 {
				target = target[:idx]
			}
			// Resolve relative to skill dir
			absTarget := filepath.Join(skillDir, target)
			if _, err := os.Stat(absTarget); err != nil {
				results = append(results, ValidationResult{
					CheckID:     "skill-links",
					Severity:    SeverityError,
					Message:     fmt.Sprintf("broken link: [%s](%s)", match[1], target),
					File:        "skill/SKILL.md",
					Line:        i + 1,
					Remediation: fmt.Sprintf("Fix or remove the link to %q — file does not exist.", target),
				})
			}
		}
	}
	return results
}

func validateReferenceNaming(refsDir string) []ValidationResult {
	var results []ValidationResult

	entries, err := os.ReadDir(refsDir)
	if err != nil {
		return results // No references/ dir is fine
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !kebabPattern.MatchString(name) {
			results = append(results, ValidationResult{
				CheckID:     "ref-naming",
				Severity:    SeverityWarning,
				Message:     fmt.Sprintf("reference file %q does not follow kebab-case naming", name),
				File:        filepath.Join("references", name),
				Remediation: fmt.Sprintf("Rename %q to kebab-case (lowercase, hyphens between words).", name),
			})
		}
	}
	return results
}

func validateSize(content []byte, bodyStart int) []ValidationResult {
	var results []ValidationResult

	lines := bytes.Split(content, []byte("\n"))
	bodyLines := len(lines) - bodyStart
	if bodyLines < 0 {
		bodyLines = 0
	}

	if bodyLines > 500 {
		results = append(results, ValidationResult{
			CheckID:     "skill-size",
			Severity:    SeverityError,
			Message:     fmt.Sprintf("SKILL.md body is %d lines (max 500)", bodyLines),
			File:        "skill/SKILL.md",
			Remediation: "Move detailed content to reference files in references/ and link from SKILL.md.",
		})
	}

	// Token estimate: ~4 chars per token, warn at ~5000 tokens
	bodyBytes := 0
	for _, line := range lines[bodyStart:] {
		bodyBytes += len(line) + 1
	}
	estimatedTokens := bodyBytes / 4
	if estimatedTokens > 5000 {
		results = append(results, ValidationResult{
			CheckID:     "skill-size",
			Severity:    SeverityWarning,
			Message:     fmt.Sprintf("SKILL.md body estimated at ~%d tokens (guideline: ~5000)", estimatedTokens),
			File:        "skill/SKILL.md",
			Remediation: "Consider moving detailed content to reference files to reduce token consumption.",
		})
	}

	return results
}
