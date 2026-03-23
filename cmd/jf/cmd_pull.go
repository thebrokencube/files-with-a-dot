package main

import (
	"bytes"
	"fmt"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
	"os"
	"path/filepath"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runPull(args []string) int {
	fs := dendrik.NewFlagSet("pull")
	dir := fs.String('d', "dir", ".", "Directory to scan for forest.yml")
	failFast := fs.BoolLong("fail-fast", "Stop on first error")
	_ = fs.Bool('f', "force", "Overwrite local file even if conflict detected")
	dryRun := fs.Bool('n', "dry-run", "Preview what would be pulled without side effects")

	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	positional := fs.GetArgs()

	// Level 0: explicit key + file
	if len(positional) >= 2 {
		return pullSingle(positional[0], positional[1])
	}

	// Forest mode
	return pullForest(*dir, positional, *failFast, false, *dryRun, nil, nil)
}

// richPull extracts ADF from View JSON output and converts to markdown.
// Returns (markdown, rawADF, error). rawADF is needed for state tracking.
func richPull(viewJSON []byte) (md []byte, rawADF []byte, err error) {
	adf, err := pipeline.ExtractDescriptionADF(viewJSON)
	if err != nil {
		return nil, nil, err
	}
	if adf == nil {
		return []byte(""), nil, nil
	}
	converted, err := pipeline.ConvertADF(adf)
	if err != nil {
		return nil, nil, err
	}
	return converted, adf, nil
}

func pullSingle(key, filePath string) int {
	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}

	// Try rich pull: JSON -> extract ADF -> convert to markdown
	out, err := p.View(key, "description", true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s: acli error\n  %s\n", key, err)
		return dendrik.ExitExternalErr
	}

	md, _, err := richPull(out)
	if err != nil {
		// Fallback: plain text pull
		fmt.Fprintf(os.Stderr, "⚠ %s: ADF conversion failed (%s), falling back to plain text\n", key, err)
		out, err = p.View(key, "description", false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: fallback acli error\n  %s\n", key, err)
			return dendrik.ExitExternalErr
		}
		md = out
	}

	if err := os.WriteFile(filePath, md, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s: write error\n  %s\n", filePath, err)
		return dendrik.ExitUserError
	}

	fmt.Printf("✓ Pulled %s description -> %s\n", key, filePath)
	return dendrik.ExitOK
}

func pullForest(dir string, positional []string,
	failFast, skipState, dryRun bool,
	f *forest.Forest, roots []*forest.Node) int {

	if f == nil {
		var err error
		f, roots, err = loadForest(dir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s\n", err)
			fmt.Fprintf(os.Stderr, "  For Level 0: jf pull <KEY> <FILE>\n")
			return dendrik.ExitUserError
		}
	}

	// Resolve target if specified
	var toPull []*forest.Node
	if len(positional) == 1 {
		node, err := forest.Resolve(roots, positional[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s\n", err)
			return dendrik.ExitUserError
		}
		toPull = []*forest.Node{node}
	} else {
		// Collect all sync:pull nodes
		all := forest.Flatten(roots)
		for _, n := range all {
			if n.Sync == "pull" && !forest.IsTBD(n.Key) {
				toPull = append(toPull, n)
			}
		}
	}

	if len(toPull) == 0 {
		fmt.Println("No pull-mode nodes found.")
		return dendrik.ExitOK
	}

	if dryRun {
		for _, n := range toPull {
			fmt.Printf("[dry-run] would pull %s -> %s\n", n.Key, n.File)
		}
		return dendrik.ExitOK
	}

	state, err := forest.LoadState(f.Dir)
	if err != nil {
		state = &forest.State{Nodes: make(map[string]forest.NodeState)}
	}

	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}
	succeeded := 0
	failed := 0

	for _, n := range toPull {
		filePath := filepath.Join(f.Dir, n.File)

		out, err := p.View(n.Key, "description", true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: acli error\n  %s\n", n.Key, err)
			failed++
			if failFast {
				break
			}
			continue
		}

		md, adfJSON, richErr := richPull(out)
		if richErr != nil {
			fmt.Fprintf(os.Stderr, "⚠ %s: ADF conversion failed (%s), falling back to plain text\n", n.Key, richErr)
			out, err = p.View(n.Key, "description", false)
			if err != nil {
				fmt.Fprintf(os.Stderr, "✗ %s: fallback acli error\n  %s\n", n.Key, err)
				failed++
				if failFast {
					break
				}
				continue
			}
			md = out
			adfJSON = nil
		}

		// Preserve frontmatter if file exists
		content, err := mergeWithFrontmatter(filePath, md)
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

		// Record pull state (skip during clone so first sync detects no conflict)
		if !skipState && adfJSON != nil {
			localHash := forest.ComputeHash(pipeline.StripFrontmatter(content))
			remoteHash := forest.ComputeHash(adfJSON)
			state.RecordPull(n.Key, localHash, remoteHash)
		}

		fmt.Printf("✓ %s -> %s\n", n.Key, n.File)
		succeeded++
	}

	// Save state (skip during clone so first sync detects no conflict)
	if !skipState {
		if err := forest.SaveState(f.Dir, state); err != nil {
			fmt.Fprintf(os.Stderr, "⚠ state save failed: %s\n", err)
		}
	}

	fmt.Printf("\nPulled %d/%d nodes", succeeded, succeeded+failed)
	if failed > 0 {
		fmt.Printf(" (%d failed)", failed)
	}
	fmt.Println()

	if failed > 0 {
		return dendrik.ExitUserError
	}
	return dendrik.ExitOK
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

	for i := 1; i < len(lines) && i < forest.MaxFrontmatterLines; i++ {
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
