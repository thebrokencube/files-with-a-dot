package main

import (
	"fmt"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
	"os"
	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runSync(args []string) int {
	fs := dendrik.NewFlagSet("sync")
	dir := fs.String('d', "dir", ".", "Directory to scan for forest.yml")
	force := fs.Bool('f', "force", "Push as plain text if marklassian conversion fails")
	failFast := fs.BoolLong("fail-fast", "Stop on first error")
	resolve := fs.String('r', "resolve", "", "Conflict resolution: local|remote (default: skip)")
	dryRun := fs.Bool('n', "dry-run", "Preview what would be synced without side effects")

	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if *resolve != "" && *resolve != "local" && *resolve != "remote" {
		fmt.Fprintf(os.Stderr, "✗ --resolve must be 'local' or 'remote'\n")
		return dendrik.ExitUserError
	}

	// Load forest once for all phases
	f, roots, err := loadForest(*dir)
	if err != nil {
		// Fall through to push/pull which will print their own errors
		fmt.Println("── Push ──")
		pushCode := pushForest(*dir, nil, "", *force, *failFast, *dryRun, nil, nil)
		fmt.Println()
		fmt.Println("── Pull ──")
		pullCode := pullForest(*dir, nil, *failFast, false, *dryRun, nil, nil)
		if pushCode != 0 || pullCode != 0 {
			return dendrik.ExitUserError
		}
		return dendrik.ExitOK
	}

	state, stateErr := forest.LoadState(f.Dir)
	if stateErr != nil {
		state = &forest.State{Nodes: make(map[string]forest.NodeState)}
	}

	// Pre-scan for conflicts on sync:both nodes
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

	// Pass pre-loaded forest to push and pull
	fmt.Println("── Push ──")
	pushCode := pushForest(*dir, nil, "", *force, *failFast, *dryRun, f, roots)

	fmt.Println()
	fmt.Println("── Pull ──")
	pullCode := pullForest(*dir, nil, *failFast, false, *dryRun, f, roots)

	if pushCode != 0 || pullCode != 0 || (conflicts > 0 && *resolve == "") {
		return dendrik.ExitUserError
	}
	return dendrik.ExitOK
}
