package observe

import (
	"fmt"
	"os"
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
