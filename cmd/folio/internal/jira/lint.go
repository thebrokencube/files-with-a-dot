package jira

import (
	"regexp"
	"strings"
)

// Issue represents a lint violation at a specific line.
type Issue struct {
	Line    int
	Message string
}

var (
	reH1         = regexp.MustCompile(`^# [^#]`)
	reH3Plus     = regexp.MustCompile(`^###+ `)
	reCodeFence  = regexp.MustCompile("^```")
	reTable      = regexp.MustCompile(`^\|.+\|`)
	reBlockquote = regexp.MustCompile(`^>`)
	reNestedDash = regexp.MustCompile(`^[ \t]+- `)
	reNestedNum  = regexp.MustCompile(`^[ \t]+\d+\. `)
	reCheckbox   = regexp.MustCompile(`- \[[ x]\]`)
	reImage      = regexp.MustCompile(`!\[`)
	reBareOpen   = regexp.MustCompile(`\[`)
	reCodeSpan   = regexp.MustCompile("`.+?`")
	reValidLink  = regexp.MustCompile(`\[.+?\]\(https?://[^\)]{1,1000}\)`)
)

// Lint validates that markdown uses only the restricted subset supported by
// the ADF converter. Returns a slice of issues (empty if valid).
func Lint(input []byte, filename string) []Issue {
	var issues []Issue
	lines := strings.Split(string(input), "\n")
	inCodeBlock := false

	for idx, line := range lines {
		lineno := idx + 1

		if reCodeFence.MatchString(line) {
			if !inCodeBlock {
				issues = append(issues, Issue{lineno, "code block (fenced) not supported"})
				inCodeBlock = true
			} else {
				inCodeBlock = false
			}
			continue
		}
		if inCodeBlock {
			continue
		}

		if reH1.MatchString(line) {
			issues = append(issues, Issue{lineno, "h1 heading not supported (only ## supported)"})
		} else if reH3Plus.MatchString(line) {
			issues = append(issues, Issue{lineno, "level 3+ heading not supported (only ## supported)"})
		} else if reTable.MatchString(line) {
			issues = append(issues, Issue{lineno, "table syntax not supported"})
		} else if reBlockquote.MatchString(line) {
			issues = append(issues, Issue{lineno, "blockquote not supported"})
		} else if reNestedDash.MatchString(line) || reNestedNum.MatchString(line) {
			issues = append(issues, Issue{lineno, "nested list not supported"})
		} else if reCheckbox.MatchString(line) {
			issues = append(issues, Issue{lineno, "checkbox not supported"})
		}

		// Inline checks: strip code spans first to avoid false positives
		inline := reCodeSpan.ReplaceAllString(line, "")
		if reImage.MatchString(inline) {
			issues = append(issues, Issue{lineno, "image syntax not supported (![alt](url) produces a plain link, not an image)"})
		}
		// Strip valid links, then check for bare [
		inline = reValidLink.ReplaceAllString(inline, "")
		if reBareOpen.MatchString(inline) {
			issues = append(issues, Issue{lineno, "bare [ or relative link not supported (only [text](https://...) links)"})
		}
	}

	return issues
}
