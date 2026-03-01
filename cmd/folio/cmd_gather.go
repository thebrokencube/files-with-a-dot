package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
)

func runGather(args []string) int {
	fs := flag.NewFlagSet("gather", flag.ExitOnError)
	folioPath := fs.String("folio", "./folio.yml", "Path to folio.yml")
	materialize := fs.Bool("materialize", false, "Create reference file stub and wire path")
	name := fs.String("name", "", "Reference file name (default: derived from URL)")
	read := fs.Bool("read", false, "Read and summarize URL (requires Claude skill)")
	fs.Parse(args)

	if fs.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Usage: folio gather <url> [--materialize] [--name <name>] [--folio PATH]\n")
		fmt.Fprintf(os.Stderr, "  Adds a source entry to folio.yml for the given URL.\n")
		return 1
	}

	if *read {
		fmt.Fprintln(os.Stderr, "The --read flag requires /folio gather (Claude skill).")
		fmt.Fprintln(os.Stderr, "Use the skill invocation to read, summarize, and materialize URLs.")
		return 1
	}

	rawURL := fs.Arg(0)

	// Validate URL
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		fmt.Fprintln(os.Stderr, output.Errf("invalid URL: %s", rawURL))
		return 1
	}

	// Validate the file parses before modifying
	if _, err := config.Load(*folioPath); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	// Derive name from URL if not specified
	refName := *name
	if refName == "" {
		refName = deriveNameFromURL(parsed)
	}

	today := time.Now().Format("2006-01-02")
	folioDir := filepath.Dir(*folioPath)

	if *materialize {
		// Create reference file stub
		refDir := filepath.Join(folioDir, "reference")
		if err := os.MkdirAll(refDir, 0755); err != nil {
			fmt.Fprintln(os.Stderr, output.Errf("creating reference directory: %s", err))
			return 1
		}
		refPath := filepath.Join(refDir, refName+".md")
		if _, err := os.Stat(refPath); err == nil {
			fmt.Fprintln(os.Stderr, output.Errf("reference file already exists: reference/%s.md", refName))
			return 1
		}

		stub := fmt.Sprintf("# %s\n\nSource: %s\nCached: %s\n\n<!-- TODO: add content -->\n", refName, rawURL, today)
		if err := os.WriteFile(refPath, []byte(stub), 0644); err != nil {
			fmt.Fprintln(os.Stderr, output.Errf("writing reference file: %s", err))
			return 1
		}

		// Add materialized source entry (path + derived_from)
		if err := appendMaterializedSource(*folioPath, refName, rawURL, today); err != nil {
			fmt.Fprintln(os.Stderr, output.Errf("%s", err))
			return 1
		}

		fmt.Println(output.Successf("Gathered %s", rawURL))
		fmt.Printf("  Created reference/%s.md\n", refName)
		fmt.Printf("  Added source entry to folio.yml\n")
	} else {
		// Add URL-only source entry (external + derived_from, no path)
		if err := appendURLSource(*folioPath, rawURL, today); err != nil {
			fmt.Fprintln(os.Stderr, output.Errf("%s", err))
			return 1
		}

		fmt.Println(output.Successf("Gathered %s", rawURL))
		fmt.Printf("  Added source entry to folio.yml\n")
	}

	return 0
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
func appendMaterializedSource(path string, name string, rawURL string, cached string) error {
	lines, sourcesIdx, insertIdx, indent, err := findSourcesInsertPoint(path)
	if err != nil {
		return err
	}

	_ = sourcesIdx

	refPath := fmt.Sprintf("reference/%s.md", name)
	newLines := []string{
		fmt.Sprintf("%s- path: %s", indent, refPath),
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
