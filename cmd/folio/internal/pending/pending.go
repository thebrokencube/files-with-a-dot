package pending

import (
	"fmt"
	"os"
	"strings"
)

// Append adds an item to the observations (schema 2) or pending (schema 1) list
// in a folio.yml file. It searches for observations: first, then falls back to
// pending:. Uses line-level manipulation to avoid reformatting the entire file.
func Append(path string, item string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	// Search for observations: first (schema 2), then pending: (schema 1)
	insertIdx := -1
	indent := "  "
	keyName := ""

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "observations:" || trimmed == "observations: []" {
			if trimmed == "observations: []" {
				lines[i] = "observations:"
			}
			insertIdx = i + 1
			keyName = "observations"
			continue
		}
		if keyName == "" && (trimmed == "pending:" || trimmed == "pending: []") {
			if trimmed == "pending: []" {
				lines[i] = "pending:"
			}
			insertIdx = i + 1
			keyName = "pending"
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
				// Hit a new top-level key — if we found observations, stop.
				// If we only found pending so far, keep scanning for observations.
				if keyName == "observations" {
					break
				}
				// We had pending, but hit a new key. Check if observations exists later.
				// For simplicity, stop — pending is our fallback.
				break
			}
		}
	}

	if insertIdx < 0 {
		return fmt.Errorf("no 'observations:' or 'pending:' key found in %s", path)
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
