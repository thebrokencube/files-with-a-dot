package main

import (
	"encoding/json"
	"fmt"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
	"os"
	"path/filepath"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runSync(args []string) int {
	fs := dendrik.NewFlagSet("sync")
	dir := fs.String('d', "dir", ".", "Directory to scan for forest.yml")
	failFast := fs.BoolLong("fail-fast", "Stop on first error")
	resolve := fs.String('r', "resolve", "", "Conflict resolution: local|remote (default: skip)")
	dryRun := fs.Bool('n', "dry-run", "Preview what would be synced without side effects")
	scaffold := fs.BoolLong("scaffold", "Create stub files for new Jira children")

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
		pushCode := pushForest(*dir, nil, "", false, *failFast, *dryRun, nil, nil, nil)
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

	// First-push safety: for nodes never synced, check if Jira has content
	// that would be overwritten. Always block — no bypass flag.
	blockedKeys, blocked := checkFirstPushSafety(f, roots, state, p, *dryRun)

	// Pass pre-loaded forest to push and pull (skip blocked keys)
	fmt.Println("── Push ──")
	pushCode := pushForest(*dir, nil, "", false, *failFast, *dryRun, f, roots, blockedKeys)

	fmt.Println()
	fmt.Println("── Pull ──")
	pullCode := pullForest(*dir, nil, *failFast, false, *dryRun, f, roots)

	// Completeness check: discover new Jira children not in local forest
	fmt.Println()
	fmt.Println("── Completeness ──")
	newChildren := checkCompleteness(f, roots, p, *dryRun, *scaffold)

	// Summary
	fmt.Println()
	fmt.Println("── Summary ──")
	if blocked > 0 {
		fmt.Printf("Blocked: %d node(s) — first sync with remote content (pull first to establish baseline)\n", blocked)
	}
	if conflicts > 0 {
		fmt.Printf("Conflicts: %d\n", conflicts)
	}
	if newChildren > 0 {
		fmt.Printf("New children: %d", newChildren)
		if *scaffold {
			fmt.Print(" (scaffolded)")
		} else {
			fmt.Print(" (use --scaffold to create stubs)")
		}
		fmt.Println()
	}
	if pushCode == 0 && pullCode == 0 && newChildren == 0 && conflicts == 0 && blocked == 0 {
		fmt.Println("Everything up to date.")
	}

	if pushCode != 0 || pullCode != 0 || (conflicts > 0 && *resolve == "") {
		return dendrik.ExitUserError
	}
	return dendrik.ExitOK
}

// checkFirstPushSafety checks push-eligible nodes that have never been synced.
// For each, it fetches the Jira description. If Jira has content, the node is
// blocked from pushing to prevent overwriting remote content the user hasn't seen.
// Returns the set of blocked keys (uppercased) and the count.
func checkFirstPushSafety(f *forest.Forest, roots []*forest.Node, state *forest.State, p *pipeline.Pipeline, dryRun bool) (map[string]bool, int) {
	all := forest.Flatten(roots)
	blockedKeys := make(map[string]bool)

	for _, n := range all {
		if forest.IsTBD(n.Key) || n.Sync == "pull" {
			continue
		}

		// Only check nodes with no state (never synced)
		if _, hasState := state.Nodes[n.Key]; hasState {
			continue
		}

		// Fetch Jira description to see if remote has content
		viewJSON, err := p.View(n.Key, "description", true)
		if err != nil {
			blockedKeys[strings.ToUpper(n.Key)] = true
			fmt.Fprintf(os.Stderr, "⚠ %s: BLOCKED — cannot reach Jira\n", n.Key)
			continue
		}

		adf, _ := pipeline.ExtractDescriptionADF(viewJSON)
		if adf == nil || string(adf) == "null" {
			continue // Jira description is empty, safe to push
		}

		// Jira has content — block this node
		blockedKeys[strings.ToUpper(n.Key)] = true
		fmt.Fprintf(os.Stderr, "⚠ %s: BLOCKED — first sync, Jira has content (pull first to establish baseline)\n", n.Key)
	}

	return blockedKeys, len(blockedKeys)
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

	if len(parents) == 0 {
		fmt.Println("No parent nodes to check.")
		return 0
	}

	totalNew := 0

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
