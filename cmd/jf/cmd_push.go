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

func runPush(args []string) int {
	fs := dendrik.NewFlagSet("push")
	plainText := fs.Bool('p', "plain-text", "Push as plain text if marklassian conversion fails")
	subtree := fs.String('s', "subtree", "", "Push node and all descendants")
	dir := fs.String('d', "dir", ".", "Directory to scan for forest.yml")
	dryRun := fs.Bool('n', "dry-run", "Preview what would be pushed without side effects")
	jsonOut := fs.Bool('j', "json", "Output plan as structured JSON")
	yes := fs.BoolLong("yes", "Proceed without confirmation in non-interactive mode")

	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	positional := fs.GetArgs()

	// Level 0: explicit key + file
	if len(positional) >= 2 {
		return pushSingle(positional[0], positional[1], *plainText)
	}

	// Forest mode: discover and push
	return pushForest(*dir, positional, *subtree, *plainText, *dryRun, *jsonOut, *yes)
}

func pushSingle(key, filePath string, plainText bool) int {
	source, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s: file not found\n", filePath)
		return dendrik.ExitUserError
	}

	stripped := pipeline.StripFrontmatter(source)

	// Construct synthetic node for engine.
	// node.File must be relative to forestDir; engine.Execute joins them.
	forestDir := resolveForestDir(filePath)
	relFile, _ := filepath.Rel(forestDir, filePath)
	if relFile == "" {
		relFile = filePath
	}
	node := &forest.Node{Key: key, File: relFile, Sync: "push"}
	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}

	// Load state if inside a forest directory
	var state *forest.State
	if _, err := os.Stat(filepath.Join(forestDir, "forest.yml")); err == nil {
		state, _ = forest.LoadState(forestDir)
		if state == nil {
			state = &forest.State{Nodes: make(map[string]forest.NodeState)}
		}
	}

	// Single-node Read (fetches remote state for this key)
	readings, err := engine.Read([]*forest.Node{node}, p, state, forestDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s: read failed: %s\n", key, err)
		return dendrik.ExitExternalErr
	}

	// Override local content from the already-read file (Read may re-read, but
	// for Level 0 the file path IS the absolute path, not relative to forestDir)
	readings[0].LocalContent = stripped
	readings[0].LocalHash = pipeline.ComputeLocalHash(stripped)
	readings[0].LocalErr = nil

	plan := engine.Plan(readings, engine.PlanOpts{Direction: "push", PlainText: plainText})

	displayPlan(plan)

	if len(plan) == 0 {
		return dendrik.ExitOK
	}

	// No batch gate needed — single node is never batch
	// Execute with state (loaded above if inside forest)
	results, execErr := engine.Execute(plan, p, state, forestDir)
	if execErr != nil {
		fmt.Fprintf(os.Stderr, "⚠ %s\n", execErr)
	}

	for _, r := range results {
		if r.Success && r.Kind == engine.ActionPush {
			fmt.Printf("✓ Pushed %s description (%d bytes)\n", key, len(source))
		} else if r.Error != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: %s\n", key, r.Error)
			return dendrik.ExitExternalErr
		}
	}

	return dendrik.ExitOK
}

func pushForest(dir string, positional []string, subtreeTarget string,
	plainText, dryRun, jsonOut, yes bool) int {

	f, roots, err := loadForest(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s\n", err)
		fmt.Fprintf(os.Stderr, "  For Level 0: jf push <KEY> <FILE>\n")
		return dendrik.ExitUserError
	}

	// Validate first
	issues := forest.Validate(roots, f)
	hasErrors := false
	for _, iss := range issues {
		if iss.Level == "error" {
			fmt.Fprintln(os.Stderr, iss.String())
			hasErrors = true
		}
	}
	if hasErrors {
		return dendrik.ExitUserError
	}

	// Resolve target
	var pushRoots []*forest.Node
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
		pushRoots = []*forest.Node{node}
	} else {
		pushRoots = roots
	}

	// Post-order traversal for push
	ordered := forest.PostOrder(pushRoots)

	// Filter to push-eligible nodes (sync:push or sync:both, skip TBD)
	var toPush []*forest.Node
	for _, n := range ordered {
		if forest.IsTBD(n.Key) {
			continue
		}
		if n.Sync == "pull" {
			continue
		}
		toPush = append(toPush, n)
	}

	if len(toPush) == 0 {
		fmt.Println("No push-mode nodes found.")
		return dendrik.ExitOK
	}

	state, stateErr := forest.LoadState(f.Dir)
	if stateErr != nil {
		state = &forest.State{Nodes: make(map[string]forest.NodeState)}
	}

	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}

	// Engine pipeline
	readings, err := engine.Read(toPush, p, state, f.Dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ read failed: %s\n", err)
		return dendrik.ExitExternalErr
	}

	plan := engine.Plan(readings, engine.PlanOpts{Direction: "push", PlainText: plainText})

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
		fmt.Fprintln(os.Stderr, "Non-interactive batch push. Run with --yes for non-interactive execution.")
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
