package main

import (
	"flag"
	"fmt"
	"jf/internal/forest"
	"jf/internal/pipeline"
	"os"
	"path/filepath"
)

func runSync(args []string) int {
	fs := flag.NewFlagSet("sync", flag.ContinueOnError)
	dir := fs.String("dir", ".", "Directory to scan for forest.yml")
	force := fs.Bool("force", false, "Push as plain text if marklassian conversion fails")
	failFast := fs.Bool("fail-fast", false, "Stop on first error")
	resolve := fs.String("resolve", "", "Conflict resolution: local|remote (default: skip)")

	if err := parseFlags(fs, args); err != nil {
		return 1
	}

	if *resolve != "" && *resolve != "local" && *resolve != "remote" {
		fmt.Fprintf(os.Stderr, "✗ --resolve must be 'local' or 'remote'\n")
		return 1
	}

	// Pre-scan for conflicts on sync:both nodes
	f, roots, err := loadForest(*dir)
	if err != nil {
		// Fall through to push/pull which will print their own errors
		fmt.Println("── Push ──")
		pushCode := pushForest(*dir, nil, "", *force, *failFast)
		fmt.Println()
		fmt.Println("── Pull ──")
		pullCode := pullForest(*dir, nil, *failFast, false)
		if pushCode != 0 || pullCode != 0 {
			return 1
		}
		return 0
	}

	state, stateErr := forest.LoadState(f.Dir)
	if stateErr != nil {
		state = &forest.State{Nodes: make(map[string]forest.NodeState)}
	}

	all := forest.Flatten(roots)
	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}
	conflicts := 0

	for _, n := range all {
		if n.Sync != "both" || forest.IsTBD(n.Key) {
			continue
		}

		viewJSON, err := p.View(n.Key, "description", true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠ %s: cannot check conflict (%s)\n", n.Key, err)
			continue
		}

		adf, _ := pipeline.ExtractDescriptionADF(viewJSON)
		if adf == nil {
			continue
		}

		filePath := filepath.Join(f.Dir, n.File)
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		localContent := pipeline.StripFrontmatter(content)

		status := state.DetectConflict(n.Key, localContent, adf)
		if status == forest.ConflictBoth {
			if *resolve == "" {
				fmt.Fprintf(os.Stderr, "⚠ %s: CONFLICT (both local and remote changed) — skipping\n", n.Key)
				conflicts++
				continue
			}
			fmt.Fprintf(os.Stderr, "⚠ %s: CONFLICT — resolving with --%s\n", n.Key, *resolve)
		}
	}

	if conflicts > 0 && *resolve == "" {
		fmt.Fprintf(os.Stderr, "\n%d conflict(s) detected. Use --resolve local|remote to resolve.\n\n", conflicts)
	}

	fmt.Println("── Push ──")
	pushCode := pushForest(*dir, nil, "", *force, *failFast)

	fmt.Println()
	fmt.Println("── Pull ──")
	pullCode := pullForest(*dir, nil, *failFast, false)

	if pushCode != 0 || pullCode != 0 || (conflicts > 0 && *resolve == "") {
		return 1
	}
	return 0
}
