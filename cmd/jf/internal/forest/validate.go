package forest

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ValidationIssue represents a single validation problem.
type ValidationIssue struct {
	Level   string // "error" | "warning"
	File    string
	Message string
}

func (v ValidationIssue) String() string {
	icon := "⚠"
	if v.Level == "error" {
		icon = "✗"
	}
	if v.File != "" {
		return fmt.Sprintf("%s %s: %s", icon, v.File, v.Message)
	}
	return fmt.Sprintf("%s %s", icon, v.Message)
}

// Validate checks forest integrity. Returns a list of issues.
func Validate(roots []*Node, forest *Forest) []ValidationIssue {
	var issues []ValidationIssue

	all := Flatten(roots)

	issues = append(issues, checkKeyUniqueness(all)...)
	issues = append(issues, checkTBDNodes(all)...)
	issues = append(issues, checkFieldValues(all)...)
	issues = append(issues, checkStemUniqueness(all)...)

	return issues
}

// checkKeyUniqueness ensures no two nodes share a jira: key (except TBD).
func checkKeyUniqueness(nodes []*Node) []ValidationIssue {
	var issues []ValidationIssue
	seen := make(map[string]string) // key -> first file

	for _, n := range nodes {
		if IsTBD(n.Key) {
			continue
		}
		upper := strings.ToUpper(n.Key)
		if first, ok := seen[upper]; ok {
			issues = append(issues, ValidationIssue{
				Level:   "error",
				File:    n.File,
				Message: fmt.Sprintf("duplicate key %s (also in %s)", n.Key, first),
			})
		} else {
			seen[upper] = n.File
		}
	}

	return issues
}

// checkTBDNodes ensures TBD nodes have a type and a label source.
func checkTBDNodes(nodes []*Node) []ValidationIssue {
	var issues []ValidationIssue

	for _, n := range nodes {
		if !IsTBD(n.Key) {
			continue
		}
		if n.Type == "" {
			issues = append(issues, ValidationIssue{
				Level:   "error",
				File:    n.File,
				Message: "TBD node missing type (set type: in frontmatter or forest.yml defaults)",
			})
		}
		if n.Label == "" {
			issues = append(issues, ValidationIssue{
				Level:   "error",
				File:    n.File,
				Message: "TBD node missing label (add label:, a # heading, or use a descriptive filename)",
			})
		}
	}

	return issues
}

// checkFieldValues validates sync and type field values.
func checkFieldValues(nodes []*Node) []ValidationIssue {
	var issues []ValidationIssue

	for _, n := range nodes {
		if n.Sync != "" && n.Sync != "push" && n.Sync != "pull" {
			issues = append(issues, ValidationIssue{
				Level:   "error",
				File:    n.File,
				Message: fmt.Sprintf("invalid sync value %q (must be push or pull)", n.Sync),
			})
		}
	}

	return issues
}

// checkStemUniqueness warns when multiple files share the same filename stem
// within a directory (ambiguous for humans even if keys differ).
func checkStemUniqueness(nodes []*Node) []ValidationIssue {
	var issues []ValidationIssue

	// Group by directory
	dirStems := make(map[string]map[string][]string) // dir -> stem -> []files

	for _, n := range nodes {
		dir := filepath.Dir(n.File)
		stem := filepath.Base(n.File)
		stem = strings.TrimSuffix(stem, filepath.Ext(stem))

		if dirStems[dir] == nil {
			dirStems[dir] = make(map[string][]string)
		}
		dirStems[dir][stem] = append(dirStems[dir][stem], n.File)
	}

	for _, stems := range dirStems {
		for stem, files := range stems {
			if len(files) > 1 {
				issues = append(issues, ValidationIssue{
					Level:   "warning",
					File:    files[1],
					Message: fmt.Sprintf("duplicate stem %q in directory (also %s)", stem, files[0]),
				})
			}
		}
	}

	return issues
}
