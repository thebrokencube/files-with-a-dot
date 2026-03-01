package pending

import (
	"fmt"
	"os"
	"strings"
)

// Append adds an item to the pending list in a folio.yml file.
// It uses line-level manipulation to avoid reformatting the entire file.
func Append(path string, item string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading %s: %w", path, err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")

	// Find the pending: line and determine where to insert
	insertIdx := -1
	indent := "  "

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "pending:" || trimmed == "pending: []" {
			// Handle empty list case: pending: []
			if trimmed == "pending: []" {
				lines[i] = "pending:"
			}
			// Insert after this line, or after the last pending item
			insertIdx = i + 1
			continue
		}

		// If we're inside the pending block, track the last item
		if insertIdx > 0 && i >= insertIdx {
			if strings.HasPrefix(line, "  - ") || strings.HasPrefix(line, "  #") {
				insertIdx = i + 1
				// Preserve existing indent
				if strings.HasPrefix(line, "  - ") {
					indent = "  "
				}
			} else if trimmed == "" {
				// Blank line within pending is ok, keep scanning
				continue
			} else {
				// Hit a new top-level key, stop
				break
			}
		}
	}

	if insertIdx < 0 {
		return fmt.Errorf("no 'pending:' key found in %s", path)
	}

	// Build the new line. Always quote — pending items are descriptive sentences
	// that typically contain YAML-special characters (colons, dashes, etc.).
	newLine := fmt.Sprintf("%s- %q", indent, item)

	// Insert the new line
	result := make([]string, 0, len(lines)+1)
	result = append(result, lines[:insertIdx]...)
	result = append(result, newLine)
	result = append(result, lines[insertIdx:]...)

	return os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644)
}
