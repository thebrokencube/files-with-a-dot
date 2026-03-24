package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/engine"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
	"golang.org/x/term"
)

func runSync(args []string) int {
	fs := dendrik.NewFlagSet("sync")
	dir := fs.String('d', "dir", ".", "Directory to scan for forest.yml")
	resolve := fs.String('r', "resolve", "", "Conflict resolution: local|remote (default: block)")
	dryRun := fs.Bool('n', "dry-run", "Preview what would be synced without side effects")
	scaffold := fs.BoolLong("scaffold", "Create stub files for new Jira children")
	plainText := fs.Bool('p', "plain-text", "Push as plain text if marklassian conversion fails")
	jsonOut := fs.Bool('j', "json", "Output plan as structured JSON")
	yes := fs.BoolLong("yes", "Proceed without confirmation in non-interactive mode")

	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if *resolve != "" && *resolve != "local" && *resolve != "remote" {
		fmt.Fprintf(os.Stderr, "✗ --resolve must be 'local' or 'remote'\n")
		return dendrik.ExitUserError
	}

	f, roots, err := loadForest(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s\n", err)
		return dendrik.ExitUserError
	}

	state, stateErr := forest.LoadState(f.Dir)
	if stateErr != nil {
		state = &forest.State{Nodes: make(map[string]forest.NodeState)}
	}

	all := forest.Flatten(roots)
	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}

	// Engine pipeline: Read → Plan → Execute
	readings, err := engine.Read(all, p, state, f.Dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ read failed: %s\n", err)
		return dendrik.ExitExternalErr
	}

	opts := engine.PlanOpts{Direction: "both", Resolve: *resolve, PlainText: *plainText}
	plan := engine.Plan(readings, opts)

	if *jsonOut {
		writePlanJSON(plan)
		return dendrik.ExitOK
	}

	displayPlan(plan)

	if *dryRun {
		return dendrik.ExitOK
	}

	// Batch safety gate
	if isBatch(plan) && !term.IsTerminal(int(os.Stdin.Fd())) && !*yes {
		fmt.Fprintln(os.Stderr, "Non-interactive batch sync. Run with --yes for non-interactive execution.")
		return dendrik.ExitUserError
	}

	// Execute (owns all state persistence)
	results, execErr := engine.Execute(plan, p, state, f.Dir)
	if execErr != nil {
		fmt.Fprintf(os.Stderr, "⚠ %s\n", execErr)
	}

	// Completeness check
	fmt.Println()
	fmt.Println("── Completeness ──")
	newChildren := checkCompleteness(f, roots, p, *dryRun, *scaffold)

	// Summary
	displayResults(results, newChildren, *scaffold)

	if hasFailures(results) {
		return dendrik.ExitUserError
	}
	return dendrik.ExitOK
}

// isBatch returns true if the plan contains more than 1 mutation action (Push or Pull).
func isBatch(actions []engine.Action) bool {
	mutations := 0
	for _, a := range actions {
		if a.Kind == engine.ActionPush || a.Kind == engine.ActionPull {
			mutations++
			if mutations > 1 {
				return true
			}
		}
	}
	return false
}

// displayPlan prints the plan sorted: BLOCKED first, then PUSH/PULL, then SKIP.
func displayPlan(actions []engine.Action) {
	if len(actions) == 0 {
		fmt.Println("No nodes to process.")
		return
	}

	sorted := make([]engine.Action, len(actions))
	copy(sorted, actions)
	sort.SliceStable(sorted, func(i, j int) bool {
		return actionSortKey(sorted[i]) < actionSortKey(sorted[j])
	})

	fmt.Println("── Plan ──────────────────────────────────────────")
	for _, a := range sorted {
		label := strings.ToUpper(a.Kind.String())
		hint := ""
		if a.Kind == engine.ActionBlocked {
			hint = " (" + blockHint(a) + ")"
		} else if a.Reason != "" {
			hint = " (" + a.Reason + ")"
		}
		fmt.Printf("  %-7s %-10s %s%s\n", label, a.Node.Key, a.Node.File, hint)
	}

	// Summary line
	counts := countActions(actions)
	var parts []string
	if counts[engine.ActionPush] > 0 {
		parts = append(parts, fmt.Sprintf("%d push", counts[engine.ActionPush]))
	}
	if counts[engine.ActionPull] > 0 {
		parts = append(parts, fmt.Sprintf("%d pull", counts[engine.ActionPull]))
	}
	if counts[engine.ActionBlocked] > 0 {
		parts = append(parts, fmt.Sprintf("%d blocked", counts[engine.ActionBlocked]))
	}
	if counts[engine.ActionSkip] > 0 {
		parts = append(parts, fmt.Sprintf("%d skip", counts[engine.ActionSkip]))
	}
	fmt.Printf("── %s ──\n", strings.Join(parts, ", "))
}

// Sort priorities for plan display (not exit codes).
const (
	sortBlocked = iota
	sortPush
	sortPull
	sortSkip
	sortOther
)

func actionSortKey(a engine.Action) int {
	switch a.Kind {
	case engine.ActionBlocked:
		return sortBlocked
	case engine.ActionPush:
		return sortPush
	case engine.ActionPull:
		return sortPull
	case engine.ActionSkip:
		return sortSkip
	default:
		return sortOther
	}
}

func countActions(actions []engine.Action) map[engine.ActionKind]int {
	counts := make(map[engine.ActionKind]int)
	for _, a := range actions {
		counts[a.Kind]++
	}
	return counts
}

// blockHint returns a human-readable hint for blocked actions.
func blockHint(a engine.Action) string {
	switch a.Block {
	case engine.BlockEmpty:
		return "empty content — no override"
	case engine.BlockRemoteUnknown:
		return "remote unreachable — no override"
	case engine.BlockFirstPush:
		return "first sync, remote has content — resolve in terminal"
	case engine.BlockFirstPull:
		return "first sync, local has content — resolve in terminal"
	case engine.BlockOverwrite:
		return "remote changed — resolve in terminal"
	case engine.BlockConflict:
		return "conflict — use --resolve local|remote"
	default:
		return a.Reason
	}
}

// Block tier constants for JSON output.
const (
	tierNoOverride  = 3 // Tier 3: no override possible
	tierInteractive = 2 // Tier 2: interactive override or --resolve
)

// blockTier returns the tier number for a blocked action.
func blockTier(a engine.Action) int {
	switch a.Block {
	case engine.BlockEmpty, engine.BlockRemoteUnknown:
		return tierNoOverride
	default:
		return tierInteractive
	}
}

// planJSON is the JSON output structure for --json.
type planJSON struct {
	Plan    []planEntryJSON `json:"plan"`
	Summary planSummary     `json:"summary"`
}

type planEntryJSON struct {
	Action string `json:"action"`
	Key    string `json:"key"`
	File   string `json:"file"`
	Reason string `json:"reason"`
	Tier   int    `json:"tier,omitempty"`
	Hint   string `json:"hint,omitempty"`
}

type planSummary struct {
	Push    int `json:"push"`
	Pull    int `json:"pull"`
	Blocked int `json:"blocked"`
	Skip    int `json:"skip"`
}

func writePlanJSON(actions []engine.Action) {
	counts := countActions(actions)
	result := planJSON{
		Plan: make([]planEntryJSON, len(actions)),
		Summary: planSummary{
			Push:    counts[engine.ActionPush],
			Pull:    counts[engine.ActionPull],
			Blocked: counts[engine.ActionBlocked],
			Skip:    counts[engine.ActionSkip],
		},
	}

	for i, a := range actions {
		entry := planEntryJSON{
			Action: a.Kind.String(),
			Key:    a.Node.Key,
			File:   a.Node.File,
			Reason: a.Reason,
		}
		if a.Kind == engine.ActionBlocked {
			entry.Tier = blockTier(a)
			entry.Hint = blockHint(a)
		}
		result.Plan[i] = entry
	}

	dendrik.WriteResult(os.Stdout, result)
}

// displayResults prints the execution summary.
func displayResults(results []engine.Result, newChildren int, scaffold bool) {
	succeeded := 0
	failed := 0
	blocked := 0
	skipped := 0
	for _, r := range results {
		switch {
		case r.Kind == engine.ActionSkip:
			skipped++
		case r.Kind == engine.ActionBlocked:
			blocked++
		case r.Success:
			succeeded++
			fmt.Printf("✓ %s %s (%s)\n", r.Kind.String(), r.Node.Key, r.Node.File)
		default:
			failed++
			if r.Error != nil {
				fmt.Fprintf(os.Stderr, "✗ %s: %s\n", r.Node.Key, r.Error)
			}
		}
	}

	fmt.Println()
	fmt.Println("── Summary ──")
	if succeeded > 0 {
		fmt.Printf("Synced: %d node(s)\n", succeeded)
	}
	if blocked > 0 {
		fmt.Printf("Blocked: %d node(s)\n", blocked)
	}
	if failed > 0 {
		fmt.Printf("Failed: %d node(s)\n", failed)
	}
	if newChildren > 0 {
		fmt.Printf("New children: %d", newChildren)
		if scaffold {
			fmt.Print(" (scaffolded)")
		} else {
			fmt.Print(" (use --scaffold to create stubs)")
		}
		fmt.Println()
	}
	if succeeded == 0 && blocked == 0 && failed == 0 && newChildren == 0 {
		fmt.Println("Everything up to date.")
	}
}

func hasFailures(results []engine.Result) bool {
	for _, r := range results {
		if !r.Success && r.Kind != engine.ActionSkip && r.Kind != engine.ActionBlocked {
			return true
		}
	}
	return false
}

// checkCompleteness queries Jira for children of each parent node and reports
// any that don't exist in the local forest. Returns the count of new children found.
func checkCompleteness(f *forest.Forest, roots []*forest.Node, p *pipeline.Pipeline, dryRun, scaffold bool) int {
	all := forest.Flatten(roots)

	// Build set of all local keys for fast lookup
	localKeys := make(map[string]bool)
	for _, n := range all {
		localKeys[strings.ToUpper(n.Key)] = true
	}

	// Find parent nodes: nodes with children, plus root nodes (which may have
	// Jira children we haven't discovered yet)
	var parents []*forest.Node
	seen := make(map[string]bool)
	for _, n := range all {
		if forest.IsTBD(n.Key) || seen[n.Key] {
			continue
		}
		if len(n.Children) > 0 {
			parents = append(parents, n)
			seen[n.Key] = true
		}
	}
	// Also check root nodes even if they have no local children
	for _, r := range roots {
		if !forest.IsTBD(r.Key) && !seen[r.Key] {
			parents = append(parents, r)
			seen[r.Key] = true
		}
	}

	totalNew := 0

	if len(parents) == 0 {
		fmt.Println("No parent nodes to check.")
		return totalNew
	}

	for _, parent := range parents {
		jql := fmt.Sprintf("parent = %s ORDER BY rank", parent.Key)
		out, err := p.Search(jql, "summary,issuetype", 100, true)
		if err != nil {
			// No children or search failed — skip silently
			continue
		}

		children, err := parseCompletenessResults(out)
		if err != nil {
			continue
		}

		for _, child := range children {
			if localKeys[strings.ToUpper(child.Key)] {
				continue
			}

			totalNew++
			if dryRun {
				fmt.Printf("[dry-run] new: %s — %s (parent: %s)\n", child.Key, child.Summary, parent.Key)
			} else if scaffold {
				stubPath := scaffoldStub(f, parent, child)
				if stubPath != "" {
					fmt.Printf("+ %s — %s -> %s\n", child.Key, child.Summary, stubPath)
					localKeys[strings.ToUpper(child.Key)] = true
				}
			} else {
				fmt.Printf("  new: %s — %s (parent: %s)\n", child.Key, child.Summary, parent.Key)
			}
		}
	}

	if totalNew == 0 {
		fmt.Println("All children accounted for.")
	}

	return totalNew
}

type completenessChild struct {
	Key     string
	Summary string
	Type    string
}

func parseCompletenessResults(out []byte) ([]completenessChild, error) {
	var issues []struct {
		Key    string `json:"key"`
		Fields struct {
			Summary   string `json:"summary"`
			IssueType struct {
				Name string `json:"name"`
			} `json:"issuetype"`
		} `json:"fields"`
	}
	if err := json.Unmarshal(out, &issues); err != nil {
		return nil, err
	}

	var result []completenessChild
	for _, iss := range issues {
		result = append(result, completenessChild{
			Key:     iss.Key,
			Summary: iss.Fields.Summary,
			Type:    iss.Fields.IssueType.Name,
		})
	}
	return result, nil
}

// scaffoldStub creates a stub .md file for a new Jira child in the parent's directory.
func scaffoldStub(f *forest.Forest, parent *forest.Node, child completenessChild) string {
	parentDir := filepath.Dir(filepath.Join(f.Dir, parent.File))
	stubFile := filepath.Join(parentDir, child.Key+".md")

	// Don't overwrite existing files
	if _, err := os.Stat(stubFile); err == nil {
		return ""
	}

	syncMode := f.Defaults.Sync
	if syncMode == "" {
		syncMode = "push"
	}

	label := strings.ReplaceAll(child.Summary, "\"", "\\\"")
	fm := fmt.Sprintf("---\njira: %s\nlabel: \"%s\"\nsync: %s\n---\n", child.Key, label, syncMode)

	if err := os.WriteFile(stubFile, []byte(fm), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "⚠ %s: could not create stub: %s\n", child.Key, err)
		return ""
	}

	rel, err := filepath.Rel(f.Dir, stubFile)
	if err != nil {
		return stubFile
	}
	return rel
}
