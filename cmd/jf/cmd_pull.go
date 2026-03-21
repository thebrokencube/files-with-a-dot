package main

import (
	"bytes"
	"flag"
	"fmt"
	"jf/internal/forest"
	"jf/internal/pipeline"
	"os"
	"path/filepath"
	"strings"
)

func runPull(args []string) int {
	fs := flag.NewFlagSet("pull", flag.ContinueOnError)
	dir := fs.String("dir", ".", "Directory to scan for forest.yml")
	failFast := fs.Bool("fail-fast", false, "Stop on first error")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	positional := fs.Args()

	// Level 0: explicit key + file
	if len(positional) >= 2 {
		return pullSingle(positional[0], positional[1])
	}

	// Forest mode
	return pullForest(*dir, positional, *failFast)
}

func pullSingle(key, filePath string) int {
	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}

	out, err := p.View(key, "description", false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s: acli error\n  %s\n", key, err)
		return 2
	}

	if err := os.WriteFile(filePath, out, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s: write error\n  %s\n", filePath, err)
		return 1
	}

	fmt.Printf("✓ Pulled %s description -> %s\n", key, filePath)
	return 0
}

func pullForest(dir string, positional []string, failFast bool) int {
	f, err := forest.FindForest(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s\n", err)
		return 1
	}
	if f == nil {
		fmt.Fprintf(os.Stderr, "✗ No forest.yml found\n")
		fmt.Fprintf(os.Stderr, "  For Level 0: jf pull <KEY> <FILE>\n")
		return 1
	}

	roots, err := forest.Discover(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ Discovery failed: %s\n", err)
		return 1
	}

	// Resolve target if specified
	var toPull []*forest.Node
	if len(positional) == 1 {
		node, err := forest.Resolve(roots, positional[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s\n", err)
			return 1
		}
		toPull = []*forest.Node{node}
	} else {
		// Collect all sync:pull nodes
		all := forest.Flatten(roots)
		for _, n := range all {
			if n.Sync == "pull" && strings.ToUpper(n.Key) != "TBD" {
				toPull = append(toPull, n)
			}
		}
	}

	if len(toPull) == 0 {
		fmt.Println("No pull-mode nodes found.")
		return 0
	}

	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}
	succeeded := 0
	failed := 0

	for _, n := range toPull {
		filePath := filepath.Join(f.Dir, n.File)

		out, err := p.View(n.Key, "description", false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: acli error\n  %s\n", n.Key, err)
			failed++
			if failFast {
				break
			}
			continue
		}

		// Preserve frontmatter if file exists
		content, err := mergeWithFrontmatter(filePath, out)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: %s\n", n.Key, err)
			failed++
			if failFast {
				break
			}
			continue
		}

		if err := os.WriteFile(filePath, content, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: write error\n  %s\n", n.Key, err)
			failed++
			if failFast {
				break
			}
			continue
		}

		fmt.Printf("✓ %s -> %s\n", n.Key, n.File)
		succeeded++
	}

	fmt.Printf("\nPulled %d/%d nodes", succeeded, succeeded+failed)
	if failed > 0 {
		fmt.Printf(" (%d failed)", failed)
	}
	fmt.Println()

	if failed > 0 {
		return 1
	}
	return 0
}

// mergeWithFrontmatter preserves YAML frontmatter from the existing file
// and replaces the content below the closing fence with pulled content.
func mergeWithFrontmatter(filePath string, pulled []byte) ([]byte, error) {
	existing, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// New file — no frontmatter to preserve
			return pulled, nil
		}
		return nil, err
	}

	// Extract frontmatter from existing file
	fm := extractExistingFrontmatter(existing)
	if fm == nil {
		return pulled, nil
	}

	// Combine: existing frontmatter + pulled content
	var buf bytes.Buffer
	buf.Write(fm)
	buf.WriteString("---\n")
	buf.Write(pulled)
	return buf.Bytes(), nil
}

// extractExistingFrontmatter returns the opening fence and YAML content
// lines (excluding closing fence), or nil if no frontmatter found.
func extractExistingFrontmatter(content []byte) []byte {
	lines := bytes.SplitN(content, []byte("\n"), -1)
	if len(lines) == 0 || strings.TrimSpace(string(lines[0])) != "---" {
		return nil
	}

	for i := 1; i < len(lines) && i < 50; i++ {
		if strings.TrimSpace(string(lines[i])) == "---" {
			var buf bytes.Buffer
			for j := 0; j < i; j++ {
				buf.Write(lines[j])
				buf.WriteByte('\n')
			}
			return buf.Bytes()
		}
	}

	return nil
}
