package observe

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var observationRe = regexp.MustCompile(`^(bug|gap|idea|debt|task)\([a-z][a-z0-9-]*\): .+$`)

// TypeInfo describes a valid observation type.
type TypeInfo struct {
	Name        string
	Description string
}

// ValidTypes returns the closed vocabulary of observation types.
func ValidTypes() []TypeInfo {
	return []TypeInfo{
		{"bug", "Something broken"},
		{"gap", "Something missing"},
		{"idea", "Potential improvement"},
		{"debt", "Known shortcuts / cleanup needed"},
		{"task", "Action item to do"},
	}
}

// Validate checks that an observation matches the required format.
func Validate(item string) error {
	if !observationRe.MatchString(item) {
		return fmt.Errorf("invalid format %q — expected type(scope): description (types: bug, gap, idea, debt, task)", item)
	}
	return nil
}

// ParseObservation extracts type, scope, and description from a validated observation string.
func ParseObservation(item string) (typ, scope, desc string, err error) {
	if err = Validate(item); err != nil {
		return
	}
	parenOpen := strings.Index(item, "(")
	parenClose := strings.Index(item, ")")
	typ = item[:parenOpen]
	scope = item[parenOpen+1 : parenClose]
	desc = item[parenClose+3:] // skip "): "
	return
}

// Append adds an item to the observations: list in a folio.yml file.
// Uses line-level manipulation to avoid reformatting the entire file.
func Append(path string, item string) error {
	if err := Validate(item); err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	insertIdx := -1
	indent := "  "

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "observations:" || trimmed == "observations: []" {
			if trimmed == "observations: []" {
				lines[i] = "observations:"
			}
			insertIdx = i + 1
			continue
		}

		// If we're inside the target block, track the last item
		if insertIdx > 0 && i >= insertIdx {
			if strings.HasPrefix(line, "  - ") || strings.HasPrefix(line, "  #") {
				insertIdx = i + 1
				if strings.HasPrefix(line, "  - ") {
					indent = "  "
				}
			} else if trimmed == "" {
				continue
			} else {
				break
			}
		}
	}

	if insertIdx < 0 {
		return fmt.Errorf("no 'observations:' key found in %s", path)
	}

	// Build the new line. Always quote — items are descriptive sentences
	// that typically contain YAML-special characters (colons, dashes, etc.).
	newLine := fmt.Sprintf("%s- %q", indent, item)

	// Insert the new line
	result := make([]string, 0, len(lines)+1)
	result = append(result, lines[:insertIdx]...)
	result = append(result, newLine)
	result = append(result, lines[insertIdx:]...)

	return os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644)
}

// Remove deletes observations from a folio.yml file by index (#N) or substring match.
// Returns the list of removed item texts. Errors on ambiguity or not-found.
func Remove(path string, matches []string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}

	lines := strings.Split(string(data), "\n")

	// Collect observation line indices and their unquoted text
	type obsLine struct {
		lineIdx int
		text    string
	}
	var obs []obsLine
	inObs := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "observations:" || trimmed == "observations: []" {
			inObs = true
			continue
		}
		if inObs {
			if strings.HasPrefix(line, "  - ") {
				// Extract unquoted text
				raw := strings.TrimPrefix(line, "  - ")
				text := strings.Trim(raw, "\"")
				obs = append(obs, obsLine{i, text})
			} else if strings.HasPrefix(line, "  #") || trimmed == "" {
				continue
			} else {
				break
			}
		}
	}

	// Resolve each match to observation indices
	toRemove := map[int]bool{}
	var removed []string
	for _, m := range matches {
		if strings.HasPrefix(m, "#") {
			// Index-based: #N (1-indexed)
			var n int
			if _, err := fmt.Sscanf(m, "#%d", &n); err != nil || n < 1 || n > len(obs) {
				return nil, fmt.Errorf("invalid index %q (have %d observations)", m, len(obs))
			}
			idx := n - 1
			toRemove[idx] = true
			removed = append(removed, obs[idx].text)
		} else {
			// Substring match
			var candidates []int
			for j, o := range obs {
				if strings.Contains(o.text, m) {
					candidates = append(candidates, j)
				}
			}
			if len(candidates) == 0 {
				return nil, fmt.Errorf("no match for %q", m)
			}
			if len(candidates) > 1 {
				lines := make([]string, len(candidates))
				for k, c := range candidates {
					lines[k] = fmt.Sprintf("  #%d: %s", c+1, obs[c].text)
				}
				return nil, fmt.Errorf("ambiguous match %q — matches %d items:\n%s", m, len(candidates), strings.Join(lines, "\n"))
			}
			toRemove[candidates[0]] = true
			removed = append(removed, obs[candidates[0]].text)
		}
	}

	// Collect line indices to delete
	deleteLines := map[int]bool{}
	for idx := range toRemove {
		deleteLines[obs[idx].lineIdx] = true
	}

	// Rebuild file without deleted lines
	var result []string
	for i, line := range lines {
		if !deleteLines[i] {
			result = append(result, line)
		}
	}

	return removed, os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644)
}

// LintIssue describes a single lint finding for an observation.
type LintIssue struct {
	Index  int
	Item   string
	Reason string
}

var (
	parenPathRe = regexp.MustCompile(`\(([^)]+)\)`)
	seePathRe   = regexp.MustCompile(`(?:See|see) ([^\s,]+)`)
)

// Lint checks observations for format errors and broken inline path references.
func Lint(folioDir string, items []string) []LintIssue {
	var issues []LintIssue
	for i, item := range items {
		if err := Validate(item); err != nil {
			issues = append(issues, LintIssue{Index: i + 1, Item: item, Reason: "malformed format"})
			continue
		}

		// Extract and check inline path references
		paths := extractPaths(item)
		for _, p := range paths {
			full := filepath.Join(folioDir, p)
			if _, err := os.Stat(full); os.IsNotExist(err) {
				issues = append(issues, LintIssue{Index: i + 1, Item: item, Reason: fmt.Sprintf("broken path: %s", p)})
			}
		}
	}
	return issues
}

func extractPaths(item string) []string {
	var paths []string
	seen := map[string]bool{}

	add := func(candidate string) {
		// Strip trailing punctuation
		candidate = strings.TrimRight(candidate, ".,)")
		// Skip URLs
		if strings.Contains(candidate, "://") {
			return
		}
		// Path-likeness: must contain / or .
		if !strings.Contains(candidate, "/") && !strings.Contains(candidate, ".") {
			return
		}
		if !seen[candidate] {
			seen[candidate] = true
			paths = append(paths, candidate)
		}
	}

	// Parenthetical references
	for _, m := range parenPathRe.FindAllStringSubmatch(item, -1) {
		add(m[1])
	}

	// "See path" references
	for _, m := range seePathRe.FindAllStringSubmatch(item, -1) {
		add(m[1])
	}

	return paths
}
