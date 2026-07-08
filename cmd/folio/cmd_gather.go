package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/taxonomy"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runGather(folioPath string, materialize bool, typeFlag, name string, read bool, rawURL string) int {
	pal := dendrik.NewPalette(true)

	if !resolveOrDie(&folioPath) {
		return dendrik.ExitUserError
	}

	if read {
		fmt.Fprintln(os.Stderr, "The --read flag requires /folio gather (Claude skill).")
		fmt.Fprintln(os.Stderr, "Use the skill invocation to read, summarize, and materialize URLs.")
		return dendrik.ExitUserError
	}

	// Validate URL
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		fmt.Fprintln(os.Stderr, pal.Errf("invalid URL: %s", rawURL))
		return dendrik.ExitUserError
	}

	// Validate the file parses before modifying
	if _, err := config.Load(folioPath); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}

	// Derive name from URL if not specified
	refName := name
	if refName == "" {
		refName = deriveNameFromURL(parsed)
	}

	today := time.Now().Format("2006-01-02")
	folioDir := filepath.Dir(folioPath)

	if materialize {
		// --type is required with --materialize
		if typeFlag == "" {
			fmt.Fprintln(os.Stderr, pal.Errf("--materialize requires --type <type> (spike, survey, design, ...)"))
			fmt.Fprintf(os.Stderr, "  Valid types: %s\n", strings.Join(taxonomy.ReferenceTypes, ", "))
			return dendrik.ExitUserError
		}
		if !taxonomy.IsReferenceDir(typeFlag) {
			fmt.Fprintln(os.Stderr, pal.Errf("unknown reference type %q", typeFlag))
			fmt.Fprintf(os.Stderr, "  Valid types: %s\n", strings.Join(taxonomy.ReferenceTypes, ", "))
			return dendrik.ExitUserError
		}

		// Create reference file stub in type directory
		refRelPath := filepath.Join("reference", typeFlag, today+"-"+refName+".md")
		refAbsPath := filepath.Join(folioDir, refRelPath)
		if err := os.MkdirAll(filepath.Dir(refAbsPath), 0755); err != nil {
			fmt.Fprintln(os.Stderr, pal.Errf("creating reference directory: %s", err))
			return dendrik.ExitUserError
		}
		if _, err := os.Stat(refAbsPath); err == nil {
			fmt.Fprintln(os.Stderr, pal.Errf("reference file already exists: %s", refRelPath))
			return dendrik.ExitUserError
		}

		stub := fmt.Sprintf("# %s\n\nSource: %s\nCached: %s\n\n<!-- TODO: add content -->\n", refName, rawURL, today)
		if err := os.WriteFile(refAbsPath, []byte(stub), 0644); err != nil {
			fmt.Fprintln(os.Stderr, pal.Errf("writing reference file: %s", err))
			return dendrik.ExitUserError
		}

		// Add materialized source entry (path + derived_from)
		if err := appendMaterializedSource(folioPath, refRelPath, rawURL, today); err != nil {
			fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
			return dendrik.ExitUserError
		}

		fmt.Println(pal.Successf("Gathered %s", rawURL))
		fmt.Printf("  Created %s\n", refRelPath)
		fmt.Printf("  Added source entry to folio.yml\n")
	} else {
		// Add URL-only source entry (external + derived_from, no path)
		if err := appendURLSource(folioPath, rawURL, today); err != nil {
			fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
			return dendrik.ExitUserError
		}

		fmt.Println(pal.Successf("Gathered %s", rawURL))
		fmt.Printf("  Added source entry to folio.yml\n")
	}

	return dendrik.ExitOK
}

// deriveNameFromURL extracts a reasonable file name from a URL.
func deriveNameFromURL(u *url.URL) string {
	// Use the last non-empty path segment
	parts := strings.Split(strings.TrimRight(u.Path, "/"), "/")
	for i := len(parts) - 1; i >= 0; i-- {
		seg := parts[i]
		if seg == "" {
			continue
		}
		// Strip common extensions
		seg = strings.TrimSuffix(seg, ".html")
		seg = strings.TrimSuffix(seg, ".htm")
		seg = strings.TrimSuffix(seg, ".md")
		// Replace non-alphanumeric with hyphens
		seg = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
				return r
			}
			return '-'
		}, seg)
		return strings.ToLower(seg)
	}
	// Fallback: use host
	return strings.ReplaceAll(u.Host, ".", "-")
}

// appendURLSource adds an external web source entry to folio.yml's sources list.
// Uses the standard external source pattern: external + id.
func appendURLSource(path string, rawURL string, cached string) error {
	lines, sourcesIdx, insertIdx, indent, err := findSourcesInsertPoint(path)
	if err != nil {
		return err
	}

	_ = sourcesIdx // used by findSourcesInsertPoint to locate the block

	newLines := []string{
		fmt.Sprintf("%s- external: web", indent),
		fmt.Sprintf("%s  id: %q", indent, rawURL),
		fmt.Sprintf("%s  notes: \"gathered %s, not yet materialized\"", indent, cached),
	}

	return insertLines(path, lines, insertIdx, newLines)
}

// appendMaterializedSource adds a path + derived_from source entry to folio.yml's sources list.
// refRelPath is the relative path to the reference file (e.g., "reference/spike/2026-03-01-topic.md").
func appendMaterializedSource(path string, refRelPath string, rawURL string, cached string) error {
	lines, sourcesIdx, insertIdx, indent, err := findSourcesInsertPoint(path)
	if err != nil {
		return err
	}

	_ = sourcesIdx

	newLines := []string{
		fmt.Sprintf("%s- path: %s", indent, refRelPath),
		fmt.Sprintf("%s  derived_from:", indent),
		fmt.Sprintf("%s    - external: web", indent),
		fmt.Sprintf("%s      url: %q", indent, rawURL),
		fmt.Sprintf("%s      cached: %q", indent, cached),
	}

	return insertLines(path, lines, insertIdx, newLines)
}

// findSourcesInsertPoint finds the sources: key and the correct insertion point.
// Returns: lines, sourcesIdx, insertIdx, indent, error
func findSourcesInsertPoint(path string) ([]string, int, int, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, -1, -1, "", fmt.Errorf("reading %s: %w", path, err)
	}

	lines := strings.Split(string(data), "\n")
	sourcesIdx := -1
	insertIdx := -1
	indent := "  "

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "sources:" || trimmed == "sources: []" {
			if trimmed == "sources: []" {
				lines[i] = "sources:"
			}
			sourcesIdx = i
			insertIdx = i + 1
			continue
		}

		// If we're inside the sources block, track the last item
		if sourcesIdx >= 0 && insertIdx > 0 && i >= insertIdx {
			if strings.HasPrefix(line, "  - ") || strings.HasPrefix(line, "  #") {
				insertIdx = i + 1
				indent = "  "
			} else if strings.HasPrefix(line, "    ") {
				// Continuation of a source entry (derived_from, etc.)
				insertIdx = i + 1
			} else if trimmed == "" {
				// Blank line within sources is ok, keep scanning
				continue
			} else {
				// Hit a new top-level key, stop
				break
			}
		}
	}

	if sourcesIdx < 0 {
		return nil, -1, -1, "", fmt.Errorf("no 'sources:' key found in %s", path)
	}

	return lines, sourcesIdx, insertIdx, indent, nil
}

// insertLines splices new lines at the given index and writes back.
func insertLines(path string, lines []string, idx int, newLines []string) error {
	result := make([]string, 0, len(lines)+len(newLines))
	result = append(result, lines[:idx]...)
	result = append(result, newLines...)
	result = append(result, lines[idx:]...)
	return os.WriteFile(path, []byte(strings.Join(result, "\n")), 0644)
}
