package engine

import (
	"bytes"
	"encoding/json"
	"strings"
)

// IsSubstantiveLocal returns true if stripped content (below frontmatter)
// contains meaningful content — not empty, not a stub marker, not heading-only.
func IsSubstantiveLocal(stripped []byte) bool {
	trimmed := bytes.TrimSpace(stripped)
	if len(trimmed) == 0 {
		return false
	}
	// Stub markers
	lower := strings.ToLower(string(trimmed))
	if lower == "tbd" || lower == "todo" || lower == "wip" {
		return false
	}
	// Heading-only: all non-empty lines start with #
	lines := bytes.Split(trimmed, []byte("\n"))
	allHeadings := true
	for _, line := range lines {
		t := bytes.TrimSpace(line)
		if len(t) == 0 {
			continue
		}
		if t[0] != '#' {
			allHeadings = false
			break
		}
	}
	if allHeadings {
		return false
	}
	return true
}

// IsSubstantiveADF returns true if ADF JSON contains meaningful content —
// not nil, not empty content array, not all-empty-paragraphs.
// Conservative: unparseable ADF is treated as substantive.
func IsSubstantiveADF(adf json.RawMessage) bool {
	if adf == nil || string(adf) == "null" || len(adf) == 0 {
		return false
	}

	var doc struct {
		Content []struct {
			Type    string `json:"type"`
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"content"`
	}
	if err := json.Unmarshal(adf, &doc); err != nil {
		return true // conservative: can't parse = assume substantive
	}
	if len(doc.Content) == 0 {
		return false
	}

	// Check if all nodes are empty paragraphs (no children at all)
	// Conservative: whitespace text nodes are treated as substantive (per design doc).
	for _, node := range doc.Content {
		if node.Type != "paragraph" {
			return true // non-paragraph = substantive
		}
		if len(node.Content) > 0 {
			return true // has children (even whitespace-only) = substantive
		}
	}
	return false
}
