package main

import (
	"bytes"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik/agentskills"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik/conventions"
)

var arrowRefPattern = regexp.MustCompile(`->\s+[Rr]ead\s+(references/[^\s]+)`)

// SkillLint validates skill layer conventions. Pure function — no I/O.
func SkillLint(data *ToolData) []LintResult {
	var results []LintResult

	// Layer 1 checks via agentskills package
	if data.SkillMD == nil {
		results = append(results, lintResult("skill-exists", conventions.SeverityError,
			"SKILL.md not found at skill/SKILL.md",
			"skill/SKILL.md", 0,
			"Create skill/SKILL.md with YAML frontmatter containing `name` and `description`."))
		return results // Structure gate
	}

	layer1Results := agentskills.ValidateLayer1(data.SkillDir, data.ToolName)
	for _, r := range layer1Results {
		results = append(results, LintResult{
			CheckID:     r.CheckID,
			Severity:    conventions.Severity(r.Severity),
			Message:     r.Message,
			File:        r.File,
			Line:        r.Line,
			Remediation: r.Remediation,
		})
	}

	// Parse frontmatter for Layer 2 checks
	fm, _, _, parseErr := agentskills.ParseFrontmatter(data.SkillMD)
	if parseErr != nil {
		return results // Can't run Layer 2 without valid frontmatter
	}

	// Layer 2 checks (dendrik-specific)
	results = append(results, checkArgumentHint(fm)...)
	results = append(results, checkArrowRefs(data)...)
	results = append(results, checkActivationGuidance(fm)...)
	results = append(results, checkActivationMetadata(fm)...)

	return results
}

func checkArgumentHint(fm *agentskills.SkillFrontmatter) []LintResult {
	// Check if user_invocable is truthy
	invocable := false
	switch v := fm.UserInvocable.(type) {
	case bool:
		invocable = v
	case string:
		invocable = strings.EqualFold(v, "true")
	}

	if invocable && fm.ArgumentHint == "" {
		return []LintResult{lintResult("argument-hint", conventions.SeverityError,
			"user_invocable is true but argument-hint is missing",
			"skill/SKILL.md", 0,
			"Add `argument-hint:` field to SKILL.md frontmatter (e.g., `argument-hint: \"<command> [flags]\"`).")}
	}
	return nil
}

func checkArrowRefs(data *ToolData) []LintResult {
	var results []LintResult

	// Check SKILL.md for arrow references
	results = append(results, findBrokenArrowRefs(data.SkillMD, "skill/SKILL.md", data.RefContents)...)

	// Check references/*.md for arrow references (using pre-read content)
	for _, name := range data.RefFiles {
		content, ok := data.RefContents[name]
		if !ok {
			continue
		}
		results = append(results, findBrokenArrowRefs(content, filepath.Join("skill/references", name), data.RefContents)...)
	}

	return results
}

func findBrokenArrowRefs(content []byte, relFile string, refContents map[string][]byte) []LintResult {
	var results []LintResult
	lines := bytes.Split(content, []byte("\n"))

	for i, line := range lines {
		matches := arrowRefPattern.FindAllSubmatch(line, -1)
		for _, match := range matches {
			target := string(match[1])
			// Arrow refs are "references/foo.md" — extract filename
			refName := strings.TrimPrefix(target, "references/")
			if _, exists := refContents[refName]; !exists {
				results = append(results, lintResult("arrow-refs", conventions.SeverityError,
					"broken arrow reference: -> "+target,
					relFile, i+1,
					"Fix the arrow reference — ensure "+target+" exists relative to the skill directory."))
			}
		}
	}
	return results
}

func checkActivationGuidance(fm *agentskills.SkillFrontmatter) []LintResult {
	desc := strings.ToLower(fm.Description)
	patterns := []string{"use when", "for tasks that", "use this"}
	for _, p := range patterns {
		if strings.Contains(desc, p) {
			return nil
		}
	}
	return []LintResult{lintResult("activation-guidance", conventions.SeverityWarning,
		"description lacks activation guidance (e.g., \"Use when...\", \"For tasks that...\")",
		"skill/SKILL.md", 0,
		"Add activation guidance to the description field.")}
}

func checkActivationMetadata(fm *agentskills.SkillFrontmatter) []LintResult {
	var results []LintResult

	results = append(results, validateOptionalField("trigger", fm.Trigger)...)
	results = append(results, validateOptionalField("skip_when", fm.SkipWhen)...)
	results = append(results, validateOptionalField("related", fm.Related)...)

	return results
}

func validateOptionalField(name string, value any) []LintResult {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return []LintResult{lintResult("activation-metadata", conventions.SeverityError,
				name+" field is present but empty",
				"skill/SKILL.md", 0,
				"Remove the empty `"+name+"` field or provide a valid value.")}
		}
	case []any:
		if len(v) == 0 {
			return []LintResult{lintResult("activation-metadata", conventions.SeverityError,
				name+" field is present but empty array",
				"skill/SKILL.md", 0,
				"Remove the empty `"+name+"` field or provide values.")}
		}
	}
	return nil
}
