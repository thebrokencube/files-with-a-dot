package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/engine"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
	"golang.org/x/term"
)

func runPull(args []string) int {
	fs := dendrik.NewFlagSet("pull")
	subtree := fs.String('s', "subtree", "", "Pull node and all descendants")
	dir := fs.String('d', "dir", ".", "Directory to scan for forest.yml")
	dryRun := fs.Bool('n', "dry-run", "Preview what would be pulled without side effects")
	jsonOut := fs.Bool('j', "json", "Output plan as structured JSON")
	yes := fs.BoolLong("yes", "Proceed without confirmation in non-interactive mode")

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
	return pullForest(*dir, positional, *subtree, *dryRun, *jsonOut, *yes)
}

func pullSingle(key, filePath string) int {
	// Construct synthetic node for engine.
	// node.File must be relative to forestDir; engine.Execute joins them.
	forestDir := resolveForestDir(filePath)
	relFile, _ := filepath.Rel(forestDir, filePath)
	if relFile == "" {
		relFile = filePath
	}
	node := &forest.Node{Key: key, File: relFile, Sync: "pull"}
	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}

	// Single-node Read
	readings, err := engine.Read([]*forest.Node{node}, p, nil, forestDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s: read failed: %s\n", key, err)
		return dendrik.ExitExternalErr
	}

	plan := engine.Plan(readings, engine.PlanOpts{Direction: "pull"})

	displayPlan(plan)

	if len(plan) == 0 {
		return dendrik.ExitOK
	}

	// No batch gate — single node
	// Execute with nil state — Level 0 has no state tracking
	results, execErr := engine.Execute(plan, p, nil, forestDir)
	if execErr != nil {
		fmt.Fprintf(os.Stderr, "⚠ %s\n", execErr)
	}

	for _, r := range results {
		if r.Success && r.Kind == engine.ActionPull {
			fmt.Printf("✓ Pulled %s description -> %s\n", key, filePath)
		} else if r.Error != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: %s\n", key, r.Error)
			return dendrik.ExitExternalErr
		}
	}

	return dendrik.ExitOK
}

func pullForest(dir string, positional []string, subtreeTarget string,
	dryRun, jsonOut, yes bool) int {

	f, roots, err := loadForest(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s\n", err)
		fmt.Fprintf(os.Stderr, "  For Level 0: jf pull <KEY> <FILE>\n")
		return dendrik.ExitUserError
	}

	// Resolve target
	var pullRoots []*forest.Node
	target := subtreeTarget
	if target == "" && len(positional) == 1 {
		target = positional[0]
	}

	if target != "" {
		node, err := forest.Subtree(roots, target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s\n", err)
			return dendrik.ExitUserError
		}
		pullRoots = []*forest.Node{node}
	} else {
		pullRoots = roots
	}

	// Collect pull-eligible nodes from resolved roots
	ordered := forest.Flatten(pullRoots)
	var toPull []*forest.Node
	for _, n := range ordered {
		if (n.Sync == "pull" || n.Sync == "both") && !forest.IsTBD(n.Key) {
			toPull = append(toPull, n)
		}
	}

	if len(toPull) == 0 {
		fmt.Println("No pull-mode nodes found.")
		return dendrik.ExitOK
	}

	state, stateErr := forest.LoadState(f.Dir)
	if stateErr != nil {
		state = &forest.State{Nodes: make(map[string]forest.NodeState)}
	}

	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}

	// Engine pipeline
	readings, err := engine.Read(toPull, p, state, f.Dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ read failed: %s\n", err)
		return dendrik.ExitExternalErr
	}

	plan := engine.Plan(readings, engine.PlanOpts{Direction: "pull"})

	if jsonOut {
		writePlanJSON(plan)
		return dendrik.ExitOK
	}

	displayPlan(plan)

	if dryRun {
		return dendrik.ExitOK
	}

	// Batch safety gate
	if isBatch(plan) && !term.IsTerminal(int(os.Stdin.Fd())) && !yes {
		fmt.Fprintln(os.Stderr, "Non-interactive batch pull. Run with --yes for non-interactive execution.")
		return dendrik.ExitUserError
	}

	// Execute owns all state persistence
	results, execErr := engine.Execute(plan, p, state, f.Dir)
	if execErr != nil {
		fmt.Fprintf(os.Stderr, "⚠ %s\n", execErr)
	}

	displayResults(results, 0, false)

	if hasFailures(results) {
		return dendrik.ExitUserError
	}
	return dendrik.ExitOK
}
