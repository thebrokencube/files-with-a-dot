package jira

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
)

// ADFNode is a generic ADF node (heading, paragraph, list, text, etc.).
type ADFNode map[string]any

// ADFDoc is the top-level Atlassian Document Format document.
type ADFDoc struct {
	Version int       `json:"version"`
	Type    string    `json:"type"`
	Content []ADFNode `json:"content"`
}

var (
	reHeading     = regexp.MustCompile(`^## (.+)`)
	reBullet      = regexp.MustCompile(`^- (.+)`)
	reOrdered     = regexp.MustCompile(`^(\d+)\. (.+)`)
	reBlank       = regexp.MustCompile(`^\s*$`)
	reBold        = regexp.MustCompile(`\*\*(.+?)\*\*`)
	reCode        = regexp.MustCompile("`(.+?)`")
	reLink        = regexp.MustCompile(`\[(.+?)\]\((https?://[^\)]+)\)`)
	reLookahead   = regexp.MustCompile(`^.+?(?:\*\*|` + "`" + `|\[)`)
	reFrontmatter = regexp.MustCompile(`^---\s*$`)
)

// StripFrontmatter removes YAML frontmatter from input if present.
// Only strips when: line 0 is exactly "---", a closing "---" appears
// within lines 1-19, and at least one line between fences contains ":".
func StripFrontmatter(input []byte) []byte {
	lines := bytes.SplitN(input, []byte("\n"), -1)
	if len(lines) == 0 || !reFrontmatter.Match(lines[0]) {
		return input
	}

	limit := len(lines)
	if limit > 20 {
		limit = 20
	}

	for i := 1; i < limit; i++ {
		if reFrontmatter.Match(lines[i]) {
			hasColon := false
			for j := 1; j < i; j++ {
				if bytes.Contains(lines[j], []byte(":")) {
					hasColon = true
					break
				}
			}
			if hasColon {
				return bytes.Join(lines[i+1:], []byte("\n"))
			}
			return input
		}
	}

	return input
}

// Convert parses a restricted markdown subset into an ADF document.
// Strips frontmatter first, then processes line-by-line.
func Convert(input []byte) (*ADFDoc, error) {
	stripped := StripFrontmatter(input)
	lines := strings.Split(string(stripped), "\n")
	var content []ADFNode
	i := 0

	for i < len(lines) {
		line := lines[i]

		if m := reHeading.FindStringSubmatch(line); m != nil {
			content = append(content, ADFNode{
				"type":    "heading",
				"attrs":   map[string]any{"level": 2},
				"content": ParseInline(m[1]),
			})
			i++
		} else if reBullet.MatchString(line) {
			var items []any
			for i < len(lines) {
				if m := reBullet.FindStringSubmatch(lines[i]); m != nil {
					items = append(items, ADFNode{
						"type": "listItem",
						"content": []any{ADFNode{
							"type":    "paragraph",
							"content": ParseInline(m[1]),
						}},
					})
					i++
				} else {
					break
				}
			}
			content = append(content, ADFNode{
				"type":    "bulletList",
				"content": items,
			})
		} else if reOrdered.MatchString(line) {
			var items []any
			for i < len(lines) {
				if m := reOrdered.FindStringSubmatch(lines[i]); m != nil {
					items = append(items, ADFNode{
						"type": "listItem",
						"content": []any{ADFNode{
							"type":    "paragraph",
							"content": ParseInline(m[2]),
						}},
					})
					i++
				} else {
					break
				}
			}
			content = append(content, ADFNode{
				"type":    "orderedList",
				"content": items,
			})
		} else if reBlank.MatchString(line) {
			i++
		} else {
			content = append(content, ADFNode{
				"type":    "paragraph",
				"content": ParseInline(line),
			})
			i++
		}
	}

	return &ADFDoc{Version: 1, Type: "doc", Content: content}, nil
}

// ParseInline converts inline markdown (bold, code, links) into ADF text nodes.
func ParseInline(text string) []any {
	if strings.TrimSpace(text) == "" {
		return nil
	}

	var segments []any
	pos := 0

	for pos < len(text) {
		remaining := text[pos:]

		// Try bold: **text**
		if loc := reBold.FindStringSubmatchIndex(remaining); loc != nil && loc[0] == 0 {
			segments = append(segments, ADFNode{
				"type":  "text",
				"text":  remaining[loc[2]:loc[3]],
				"marks": []any{ADFNode{"type": "strong"}},
			})
			pos += loc[1]
			continue
		}

		// Try code: `text`
		if loc := reCode.FindStringSubmatchIndex(remaining); loc != nil && loc[0] == 0 {
			segments = append(segments, ADFNode{
				"type":  "text",
				"text":  remaining[loc[2]:loc[3]],
				"marks": []any{ADFNode{"type": "code"}},
			})
			pos += loc[1]
			continue
		}

		// Try link: [text](https://url)
		if loc := reLink.FindStringSubmatchIndex(remaining); loc != nil && loc[0] == 0 {
			segments = append(segments, ADFNode{
				"type": "text",
				"text": remaining[loc[2]:loc[3]],
				"marks": []any{ADFNode{
					"type":  "link",
					"attrs": map[string]any{"href": remaining[loc[4]:loc[5]]},
				}},
			})
			pos += loc[1]
			continue
		}

		// Plain text up to next special character
		if loc := reLookahead.FindStringIndex(remaining); loc != nil && loc[0] == 0 {
			// reLookahead matches up to (but including) the delimiter prefix.
			// We need to find where the delimiter starts and take text before it.
			plain := findPlainPrefix(remaining)
			if plain != "" {
				segments = append(segments, ADFNode{
					"type": "text",
					"text": plain,
				})
				pos += len(plain)
				continue
			}
		}

		// Fallback: rest of string is plain text
		segments = append(segments, ADFNode{
			"type": "text",
			"text": remaining,
		})
		break
	}

	return segments
}

// findPlainPrefix returns the plain text before the first inline marker.
func findPlainPrefix(s string) string {
	markers := []string{"**", "`", "["}
	minIdx := len(s)
	for _, m := range markers {
		if idx := strings.Index(s, m); idx > 0 && idx < minIdx {
			minIdx = idx
		}
	}
	if minIdx > 0 && minIdx < len(s) {
		return s[:minIdx]
	}
	return ""
}

// MarshalJSON produces pretty-printed JSON for an ADFDoc.
func (d *ADFDoc) MarshalJSON() ([]byte, error) {
	type alias ADFDoc
	return json.MarshalIndent((*alias)(d), "", "  ")
}
