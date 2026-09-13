package observe

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
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

type Severity uint8

const (
	SeverityError Severity = iota
	SeverityWarning
)

// LintIssue describes a single observation finding.
type LintIssue struct {
	Code     string
	Subject  string
	Severity Severity
	Index    int
	Item     string
	Reason   string
}

// DuplicateCandidate identifies an existing observation that exactly matches
// the normalized new item.
type DuplicateCandidate struct {
	Index int
	Item  string
}

// Normalize collapses whitespace without changing case or punctuation.
func Normalize(item string) string {
	return strings.Join(strings.Fields(item), " ")
}

// DuplicateCandidates returns exact normalized matches in existing-list order.
func DuplicateCandidates(item string, existing []string) []DuplicateCandidate {
	normalized := Normalize(item)
	var candidates []DuplicateCandidate
	for i, current := range existing {
		if Normalize(current) == normalized {
			candidates = append(candidates, DuplicateCandidate{Index: i + 1, Item: current})
		}
	}
	return candidates
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

// Resolve plans observation removals without touching the manifest.
// Returned indices are zero-based and preserve selector order.
func Resolve(items, matches []string) (remaining []string, removed []int, err error) {
	selected := make(map[int]bool)
	for _, match := range matches {
		if strings.HasPrefix(match, "#") {
			var n int
			if _, scanErr := fmt.Sscanf(match, "#%d", &n); scanErr != nil || n < 1 || n > len(items) {
				return nil, nil, fmt.Errorf("invalid index %q (have %d observations)", match, len(items))
			}
			index := n - 1
			if !selected[index] {
				selected[index] = true
				removed = append(removed, index)
			}
			continue
		}

		var candidates []int
		for index, item := range items {
			if strings.Contains(item, match) {
				candidates = append(candidates, index)
			}
		}
		if len(candidates) == 0 {
			return nil, nil, fmt.Errorf("no match for %q", match)
		}
		if len(candidates) > 1 {
			labels := make([]string, len(candidates))
			for i, index := range candidates {
				labels[i] = fmt.Sprintf("  #%d: %s", index+1, items[index])
			}
			return nil, nil, fmt.Errorf("ambiguous match %q — matches %d items:\n%s", match, len(candidates), strings.Join(labels, "\n"))
		}
		index := candidates[0]
		if !selected[index] {
			selected[index] = true
			removed = append(removed, index)
		}
	}

	remaining = make([]string, 0, len(items)-len(removed))
	for index, item := range items {
		if !selected[index] {
			remaining = append(remaining, item)
		}
	}
	return remaining, removed, nil
}

// RemoveIndices deletes only the selected observation scalar lines after
// proving that the file still contains the expected decoded item list.
func RemoveIndices(path string, items []string, removed []int) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	lines := strings.Split(string(data), "\n")
	observations, err := scanObservationScalars(lines)
	if err != nil {
		return nil, err
	}
	if len(observations) != len(items) {
		return nil, fmt.Errorf("observation list changed: file has %d items, expected %d", len(observations), len(items))
	}
	for index, item := range items {
		if observations[index].text != item {
			return nil, fmt.Errorf("observation #%d changed before removal", index+1)
		}
	}

	toRemove := make(map[int]bool, len(removed))
	var removedItems []string
	for _, index := range removed {
		if index < 0 || index >= len(observations) {
			return nil, fmt.Errorf("invalid observation index %d (have %d observations)", index, len(observations))
		}
		if toRemove[index] {
			continue
		}
		toRemove[index] = true
		removedItems = append(removedItems, items[index])
	}

	deleteLines := make(map[int]bool, len(toRemove))
	for index := range toRemove {
		deleteLines[observations[index].line] = true
	}
	result := make([]string, 0, len(lines)-len(deleteLines))
	for index, line := range lines {
		if !deleteLines[index] {
			result = append(result, line)
		}
	}
	if err := os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644); err != nil {
		return nil, fmt.Errorf("writing %s: %w", path, err)
	}
	return removedItems, nil
}

type observationScalar struct {
	line int
	text string
}

func scanObservationScalars(lines []string) ([]observationScalar, error) {
	var observations []observationScalar
	inObservations := false
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "observations:" || trimmed == "observations: []" {
			inObservations = true
			continue
		}
		if !inObservations {
			continue
		}
		if strings.HasPrefix(line, "  - ") {
			raw := strings.TrimPrefix(line, "  - ")
			var decoded string
			if err := yaml.Unmarshal([]byte(raw), &decoded); err != nil {
				return nil, fmt.Errorf("observation #%d is not a supported scalar: %w", len(observations)+1, err)
			}
			observations = append(observations, observationScalar{line: index, text: decoded})
			continue
		}
		if strings.HasPrefix(line, "  #") || trimmed == "" {
			continue
		}
		break
	}
	return observations, nil
}

var (
	parenPathRe     = regexp.MustCompile(`\(([^)]+)\)`)
	seePathRe       = regexp.MustCompile(`(?i)\bsee\s+([^\s,]+)`)
	positionalRefRe = regexp.MustCompile(`(?i)\b(obs|observation|refines|confirms|per|see)\s+#([0-9]+)\b`)
	// A trailing line or line-range reference: "foo.rb:7", "Show.tsx:12-20".
	lineRefSuffixRe = regexp.MustCompile(`:\d+(?:-\d+)?$`)
)

// Lint checks observations for format errors, path references, and unstable
// positional references using the caller-resolved context.
func Lint(items []string, ctx config.Context) []LintIssue {
	var issues []LintIssue
	for i, item := range items {
		if err := Validate(item); err != nil {
			issues = append(issues, LintIssue{
				Code:     "observation.format",
				Subject:  Normalize(item),
				Severity: SeverityError,
				Index:    i + 1,
				Item:     item,
				Reason:   "malformed format",
			})
			continue
		}

		for _, match := range positionalRefRe.FindAllStringSubmatch(item, -1) {
			phrase := strings.ToLower(strings.TrimSpace(match[0]))
			issues = append(issues, LintIssue{
				Code:     "observation.positional-reference",
				Subject:  Normalize(item),
				Severity: SeverityWarning,
				Index:    i + 1,
				Item:     item,
				Reason:   fmt.Sprintf("positional observation reference %s can drift; quote the observation description instead", phrase),
			})
		}

		for _, path := range extractPaths(item) {
			full, err := ctx.ResolvePath(path)
			if err != nil {
				issues = append(issues, LintIssue{
					Code:     "observation.path",
					Subject:  path,
					Severity: SeverityError,
					Index:    i + 1,
					Item:     item,
					Reason:   err.Error(),
				})
				continue
			}
			if _, statErr := os.Stat(full); os.IsNotExist(statErr) {
				issues = append(issues, LintIssue{
					Code:     "observation.path",
					Subject:  full,
					Severity: SeverityError,
					Index:    i + 1,
					Item:     item,
					Reason:   fmt.Sprintf("broken path: %s", path),
				})
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
		// A path reference has no whitespace — reject prose the parenthetical
		// matcher captures, e.g. "(esp ~/.dotfiles)".
		if strings.ContainsAny(candidate, " \t") {
			return
		}
		// A trailing ":12" or ":12-20" is a line reference, not part of the path. Strip it so the
		// path half is what gets resolved — otherwise "app/models/foo.rb:7" is looked up verbatim
		// and reported broken, and a bare "Show.tsx:393" reads as an unknown "<store>:<path>".
		candidate = lineRefSuffixRe.ReplaceAllString(candidate, "")
		// Path-likeness: a separator, an explicit path indicator, or a
		// recognized Folio content root makes a candidate intentional.
		if strings.Contains(candidate, "::") {
			return
		}
		if strings.Contains(candidate, ":") {
			separator := strings.IndexByte(candidate, ':')
			if separator == 0 || separator == len(candidate)-1 ||
				!pathLooksIntentional(candidate[separator+1:]) {
				return
			}
		} else if !strings.Contains(candidate, "/") || !pathLooksIntentional(candidate) {
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

// hasPathShape reports whether a candidate carries an explicit path marker or
// a file extension on its final segment.
func hasPathShape(s string) bool {
	if strings.HasPrefix(s, "/") || strings.HasPrefix(s, "./") ||
		strings.HasPrefix(s, "../") || strings.HasPrefix(s, "~") {
		return true
	}
	if strings.HasSuffix(s, "/") {
		return true
	}
	last := s[strings.LastIndex(s, "/")+1:]
	return strings.Contains(last, ".") && !strings.HasSuffix(last, ".")
}

func pathLooksIntentional(s string) bool {
	if hasPathShape(s) {
		return true
	}
	for _, root := range []string{"work/", "reference/", "output/", "compiled/"} {
		if strings.HasPrefix(s, root) {
			return true
		}
	}
	return false
}
