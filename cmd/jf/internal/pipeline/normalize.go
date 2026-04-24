package pipeline

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
)

var reTrailingSpace = regexp.MustCompile(`[ \t]+\n`)
var reMultipleBlanks = regexp.MustCompile(`\n{3,}`)

// NormalizeMarkdown applies basic text normalization for content comparison.
// Strips trailing whitespace per line, collapses 3+ consecutive newlines to 2
// (one blank line), and trims leading/trailing whitespace.
// Does NOT modify files on disk — used only for hash comparison.
func NormalizeMarkdown(content []byte) []byte {
	// Ensure trailing newline for consistent regex matching
	b := content
	if len(b) > 0 && b[len(b)-1] != '\n' {
		b = append(b, '\n')
	}

	// Strip trailing whitespace per line
	b = reTrailingSpace.ReplaceAll(b, []byte("\n"))

	// Collapse multiple blank lines to one
	b = reMultipleBlanks.ReplaceAll(b, []byte("\n\n"))

	// Normalize smart quotes/dashes to ASCII equivalents
	b = []byte(normalizeQuotes(string(b)))

	// Canonicalize block boundaries between paragraphs and lists
	b = normalizeBlockBoundaries(b)

	return bytes.TrimSpace(b)
}

// normalizeBlockBoundaries inserts blank lines at paragraph/list boundaries.
// Absorbs Pattern 2: ADF joins top-level blocks with \n\n, but authors may
// omit the blank line between paragraphs and lists.
// Line-by-line, NOT regex — a regex false-positives between consecutive list items.
func normalizeBlockBoundaries(b []byte) []byte {
	lines := bytes.Split(b, []byte("\n"))
	var result [][]byte
	for i, line := range lines {
		if i > 0 {
			prev := lines[i-1]
			// Insert blank line before list marker if preceded by non-list, non-blank
			if isListMarker(line) && !isListMarker(prev) && !isBlank(prev) {
				result = append(result, nil)
			}
			// Insert blank line after list block if followed by non-list, non-blank
			if !isListMarker(line) && !isBlank(line) && isListMarker(prev) {
				result = append(result, nil)
			}
		}
		result = append(result, line)
	}
	return bytes.Join(result, []byte("\n"))
}

func isListMarker(line []byte) bool {
	trimmed := bytes.TrimLeft(line, " \t")
	if len(trimmed) == 0 {
		return false
	}
	// All CommonMark bullet markers: -, *, +
	if len(trimmed) >= 2 && (trimmed[0] == '-' || trimmed[0] == '*' || trimmed[0] == '+') && trimmed[1] == ' ' {
		return true
	}
	// Ordered list: digits followed by . or ) then space
	for j := 0; j < len(trimmed); j++ {
		if trimmed[j] >= '0' && trimmed[j] <= '9' {
			continue
		}
		if (trimmed[j] == '.' || trimmed[j] == ')') && j > 0 && j+1 < len(trimmed) && trimmed[j+1] == ' ' {
			return true
		}
		break
	}
	return false
}

func isBlank(line []byte) bool {
	return len(bytes.TrimSpace(line)) == 0
}

// normalizeQuotes replaces Unicode smart quotes and dashes with ASCII equivalents.
// Marklassian converts straight quotes to smart quotes during ADF compilation;
// this absorbs that difference so roundtrip comparison passes.
func normalizeQuotes(s string) string {
	r := strings.NewReplacer(
		"\u2018", "'", // left single quotation mark
		"\u2019", "'", // right single quotation mark
		"\u201C", `"`, // left double quotation mark
		"\u201D", `"`, // right double quotation mark
		"\u2013", "-", // en dash
		"\u2014", "--", // em dash
	)
	return r.Replace(s)
}

// ComputeLocalHash normalizes markdown content then computes its sha256 hash.
// This is the ONLY way local content hashes should be computed — ensures
// consistent comparison across push/pull/sync cycles by absorbing whitespace
// noise from the ADF roundtrip.
func ComputeLocalHash(content []byte) string {
	return forest.ComputeHash(NormalizeMarkdown(content))
}
