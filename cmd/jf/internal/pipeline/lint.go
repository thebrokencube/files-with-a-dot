package pipeline

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// LintIssue represents a lint violation at a specific line.
type LintIssue struct {
	Line    int
	Message string
}

var (
	reFencedCode = regexp.MustCompile("^\\s*```")
	reImage      = regexp.MustCompile(`!\[`)
	reCheckbox   = regexp.MustCompile(`^\s*[-*+]\s+\[([ xX])\]`)
	reLink       = regexp.MustCompile(`\]\(([^)]+)\)`)
)

// Lint validates that markdown uses only the restricted subset supported by
// clean ADF roundtripping. Returns a slice of issues (empty if valid).
// Input should be frontmatter-stripped content.
func Lint(input []byte, filename string) []LintIssue {
	var issues []LintIssue
	lines := bytes.Split(input, []byte("\n"))

	for i, line := range lines {
		lineNum := i + 1
		trimmed := bytes.TrimSpace(line)

		if len(trimmed) == 0 {
			continue
		}

		// h1 heading: line starts with "# " (single # + space)
		if bytes.HasPrefix(trimmed, []byte("# ")) || bytes.Equal(trimmed, []byte("#")) {
			issues = append(issues, LintIssue{lineNum, "h1 heading not supported"})
			continue
		}

		// h3+ heading: line starts with "###" or more
		if bytes.HasPrefix(trimmed, []byte("### ")) || bytes.HasPrefix(trimmed, []byte("###")) && (len(trimmed) == 3 || trimmed[3] == '#' || trimmed[3] == ' ') {
			issues = append(issues, LintIssue{lineNum, "h3+ heading not supported"})
			continue
		}

		// Table: line starts with |
		if trimmed[0] == '|' {
			issues = append(issues, LintIssue{lineNum, "table syntax not supported"})
			continue
		}

		// Fenced code: line starts with ```
		if reFencedCode.Match(line) {
			issues = append(issues, LintIssue{lineNum, "code block (fenced) not supported"})
			continue
		}

		// Blockquote: line starts with >
		if trimmed[0] == '>' {
			issues = append(issues, LintIssue{lineNum, "blockquote not supported"})
			continue
		}

		// Nested list: 2+ leading spaces/tabs then list marker
		if isNestedListLine(line) {
			issues = append(issues, LintIssue{lineNum, "nested list not supported"})
			continue
		}

		// Checkbox: list marker followed by [ ] or [x]
		if reCheckbox.Match(line) {
			issues = append(issues, LintIssue{lineNum, "checkbox not supported"})
			continue
		}

		// Image: ![
		if reImage.Match(line) {
			issues = append(issues, LintIssue{lineNum, "image not supported"})
			continue
		}

		// Relative link: ](path) where path doesn't start with http(s)://
		if containsRelativeLink(line) {
			issues = append(issues, LintIssue{lineNum, "relative link not supported"})
			continue
		}

		// Bare bracket: [ not part of a complete link or image
		if containsBareBracket(line) {
			issues = append(issues, LintIssue{lineNum, "bare bracket not supported"})
		}
	}

	return issues
}

// FormatLintIssues formats lint issues for human-readable error output.
func FormatLintIssues(issues []LintIssue) string {
	var parts []string
	for _, iss := range issues {
		parts = append(parts, fmt.Sprintf("line %d: %s", iss.Line, iss.Message))
	}
	return strings.Join(parts, "; ")
}

// isNestedListLine checks if a line has 2+ leading whitespace chars followed by a list marker.
func isNestedListLine(line []byte) bool {
	indent := 0
	for _, b := range line {
		if b == ' ' || b == '\t' {
			indent++
		} else {
			break
		}
	}
	if indent < 2 {
		return false
	}
	rest := bytes.TrimLeft(line, " \t")
	return isListMarkerPrefix(rest)
}

// isListMarkerPrefix checks if bytes start with a list marker (-, *, +, or digit+./))
func isListMarkerPrefix(b []byte) bool {
	if len(b) < 2 {
		return false
	}
	// Bullet markers: -, *, +
	if (b[0] == '-' || b[0] == '*' || b[0] == '+') && b[1] == ' ' {
		return true
	}
	// Ordered: digits then . or ) then space
	for j := 0; j < len(b); j++ {
		if b[j] >= '0' && b[j] <= '9' {
			continue
		}
		if (b[j] == '.' || b[j] == ')') && j > 0 && j+1 < len(b) && b[j+1] == ' ' {
			return true
		}
		break
	}
	return false
}

// containsRelativeLink checks for ](path) where path doesn't start with http:// or https://.
func containsRelativeLink(line []byte) bool {
	matches := reLink.FindAllSubmatch(line, -1)
	for _, m := range matches {
		url := string(m[1])
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			return true
		}
	}
	return false
}

// containsBareBracket checks for [ that is not part of ![ or a complete [text](url) link.
func containsBareBracket(line []byte) bool {
	s := string(line)
	idx := 0
	for {
		pos := strings.IndexByte(s[idx:], '[')
		if pos < 0 {
			return false
		}
		abs := idx + pos

		// Skip ![ (image — caught by separate rule)
		if abs > 0 && s[abs-1] == '!' {
			idx = abs + 1
			continue
		}

		// Check if this is a complete [text](url) link
		closePos := strings.IndexByte(s[abs:], ']')
		if closePos < 0 {
			return true // unclosed bracket
		}
		afterClose := abs + closePos + 1
		if afterClose < len(s) && s[afterClose] == '(' {
			// Complete link syntax — skip past it
			parenClose := strings.IndexByte(s[afterClose:], ')')
			if parenClose >= 0 {
				idx = afterClose + parenClose + 1
				continue
			}
		}
		return true // bare bracket
	}
}
