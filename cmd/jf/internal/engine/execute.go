package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
	"golang.org/x/term"
)

// Execute processes actions from Plan. It never makes independent decisions
// about what to push/pull — it only processes the Plan output.
// Saves state.json after each successful mutation (per-node persistence).
// Commands do NOT call SaveState after Execute — Execute owns all state persistence.
func Execute(actions []Action, p *pipeline.Pipeline,
	state *forest.State, forestDir string) ([]Result, error) {

	var saveFailed bool
	var results []Result
	for _, a := range actions {
		switch a.Kind {
		case ActionSkip:
			results = append(results, Result{Node: a.Node, Kind: ActionSkip, Success: true})
		case ActionBlocked:
			if isTier3(a.Block) {
				results = append(results, Result{Node: a.Node, Kind: ActionBlocked, Success: false})
				continue
			}
			if a.Block == BlockConflict {
				results = append(results, Result{Node: a.Node, Kind: ActionBlocked, Success: false})
				continue
			}
			// Tier 2: interactive override possible (TTY only)
			if confirmOverride(a) {
				result := executeAction(promoteAction(a), p, state, forestDir)
				results = append(results, result)
				saveStatePerNode(state, forestDir, result, &saveFailed)
			} else {
				results = append(results, Result{Node: a.Node, Kind: ActionBlocked, Success: false})
			}
		case ActionPush:
			result := executePush(a, p, state, forestDir)
			results = append(results, result)
			saveStatePerNode(state, forestDir, result, &saveFailed)
		case ActionPull:
			result := executePull(a, p, state, forestDir)
			results = append(results, result)
			saveStatePerNode(state, forestDir, result, &saveFailed)
		}
	}
	if saveFailed {
		return results, fmt.Errorf("one or more state saves failed — state.json may be inconsistent")
	}
	return results, nil
}

// executeAction dispatches a promoted action to push or pull.
func executeAction(a Action, p *pipeline.Pipeline, state *forest.State, forestDir string) Result {
	switch a.Kind {
	case ActionPush:
		return executePush(a, p, state, forestDir)
	case ActionPull:
		return executePull(a, p, state, forestDir)
	default:
		return Result{Node: a.Node, Kind: a.Kind, Success: false,
			Error: fmt.Errorf("cannot execute action kind: %v", a.Kind)}
	}
}

// executePush compiles and pushes local content to Jira.
func executePush(a Action, p *pipeline.Pipeline, state *forest.State, forestDir string) Result {
	// Re-read file from disk for compilation (Compile needs original source with frontmatter)
	filePath := filepath.Join(forestDir, a.Node.File)
	source, err := os.ReadFile(filePath)
	if err != nil {
		return Result{Node: a.Node, Kind: ActionPush, Success: false, Error: err}
	}

	// Guard: file must not have changed between Read and Execute.
	// Plan approved content with a.LocalHash — if the file changed, abort.
	if a.LocalHash != "" {
		currentHash := pipeline.ComputeLocalHash(pipeline.StripFrontmatter(source))
		if currentHash != a.LocalHash {
			return Result{Node: a.Node, Kind: ActionPush, Success: false,
				Error: fmt.Errorf("file changed after plan — re-run sync")}
		}
	}

	compiled, err := p.Compile(a.Node.Key, source, a.Node.Label)
	if err != nil {
		if a.PlainText {
			fmt.Fprintf(os.Stderr, "⚠ %s: conversion failed, pushing as plain text\n", a.Node.Key)
			compiled = BuildPlainTextPayload(a.Node.Key, source)
		} else {
			return Result{Node: a.Node, Kind: ActionPush, Success: false, Error: err}
		}
	}

	if err := p.Push(compiled); err != nil {
		return Result{Node: a.Node, Kind: ActionPush, Success: false, Error: err}
	}

	// Record sync — re-read remote to get post-push ADF hash.
	// Using a.RemoteHash (pre-push) would cause phantom BlockOverwrite on next sync.
	if state != nil {
		localHash := pipeline.ComputeLocalHash(pipeline.StripFrontmatter(source))
		remoteHash := postPushRemoteHash(a.Node.Key, p, compiled, a.RemoteHash)
		state.RecordSync(a.Node.Key, "push", localHash, remoteHash)
	}

	return Result{Node: a.Node, Kind: ActionPush, Success: true}
}

// executePull fetches remote content and writes to local file.
func executePull(a Action, p *pipeline.Pipeline, state *forest.State, forestDir string) Result {
	if a.RemoteADF == nil {
		return Result{Node: a.Node, Kind: ActionPull, Success: false,
			Error: fmt.Errorf("no remote ADF to pull")}
	}

	md, err := pipeline.ConvertADF(a.RemoteADF)
	if err != nil {
		return Result{Node: a.Node, Kind: ActionPull, Success: false, Error: err}
	}

	filePath := filepath.Join(forestDir, a.Node.File)

	// Preserve frontmatter from existing file
	content, err := MergeWithFrontmatter(filePath, md)
	if err != nil {
		return Result{Node: a.Node, Kind: ActionPull, Success: false, Error: err}
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return Result{Node: a.Node, Kind: ActionPull, Success: false, Error: err}
	}

	// Record sync
	if state != nil {
		localHash := pipeline.ComputeLocalHash(pipeline.StripFrontmatter(content))
		remoteHash := forest.ComputeHash(a.RemoteADF)
		state.RecordSync(a.Node.Key, "pull", localHash, remoteHash)
	}

	return Result{Node: a.Node, Kind: ActionPull, Success: true}
}

// MergeWithFrontmatter preserves YAML frontmatter from the existing file
// and replaces the content below the closing fence with pulled content.
func MergeWithFrontmatter(filePath string, pulled []byte) ([]byte, error) {
	existing, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return pulled, nil
		}
		return nil, err
	}

	fm := extractExistingFrontmatter(existing)
	if fm == nil {
		return pulled, nil
	}

	var buf []byte
	buf = append(buf, fm...)
	buf = append(buf, "---\n"...)
	buf = append(buf, pulled...)
	return buf, nil
}

// extractExistingFrontmatter returns the opening fence and YAML content
// lines (excluding closing fence), or nil if no frontmatter found.
func extractExistingFrontmatter(content []byte) []byte {
	lines := strings.SplitN(string(content), "\n", -1)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil
	}

	for i := 1; i < len(lines) && i < forest.MaxFrontmatterLines; i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			var result string
			for j := 0; j < i; j++ {
				result += lines[j] + "\n"
			}
			return []byte(result)
		}
	}

	return nil
}

// BuildPlainTextPayload creates a plain-text ADF payload for push fallback.
func BuildPlainTextPayload(key string, source []byte) []byte {
	stripped := pipeline.StripFrontmatter(source)
	return []byte(fmt.Sprintf(`{"issues":[%q],"description":{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":%q}]}]}}`,
		key, string(stripped)))
}

// saveStatePerNode persists state.json after each successful mutation.
// Isolates partial failures: if node 5 of 10 fails, nodes 1-4 are recorded.
func saveStatePerNode(state *forest.State, forestDir string, result Result, saveFailed *bool) {
	if state == nil || !result.Success {
		return
	}
	if err := forest.SaveState(forestDir, state); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to persist state for %s: %v\n", result.Node.Key, err)
		*saveFailed = true
	}
}

// fetchRemoteHash fetches the current remote ADF for a key and returns its hash.
func fetchRemoteHash(key string, p *pipeline.Pipeline) (json.RawMessage, string, error) {
	viewJSON, err := p.View(key, "description", true)
	if err != nil {
		return nil, "", fmt.Errorf("view %s: %w", key, err)
	}
	adf, err := pipeline.ExtractDescriptionADF(viewJSON)
	if err != nil {
		return nil, "", fmt.Errorf("extract ADF %s: %w", key, err)
	}
	if adf == nil {
		return nil, "", nil
	}
	return adf, forest.ComputeHash(adf), nil
}

// postPushRemoteHash returns the best available remote hash after a push.
// Tries re-reading from Jira first; falls back to hashing the compiled ADF;
// last resort is the pre-push hash.
func postPushRemoteHash(key string, p *pipeline.Pipeline, compiled []byte, prePushHash string) string {
	_, hash, err := fetchRemoteHash(key, p)
	if err == nil {
		return hash
	}
	fmt.Fprintf(os.Stderr, "WARNING: post-push re-read failed for %s: %v\n", key, err)

	// Fallback: extract description ADF from compiled payload
	var payload struct {
		Description json.RawMessage `json:"description"`
	}
	if json.Unmarshal(compiled, &payload) == nil && len(payload.Description) > 0 {
		return forest.ComputeHash(payload.Description)
	}

	return prePushHash
}

func isTier3(b BlockReason) bool {
	return b == BlockEmpty || b == BlockRemoteUnknown
}

// confirmOverride prompts for interactive TTY override of Tier 2 blocks.
// Shows a diff of local vs remote content before the prompt.
func confirmOverride(a Action) bool {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false // agents and pipes: always blocked
	}
	prompt := overridePrompt(a)
	fmt.Fprintf(os.Stderr, "%s\n", prompt)

	showOverrideDiff(a)

	fmt.Fprintf(os.Stderr, "Override? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

// showOverrideDiff displays a unified diff of local vs remote content
// to help the user understand what an override will change.
func showOverrideDiff(a Action) {
	local := strings.TrimSpace(string(a.LocalContent))
	remote := RemoteAsMarkdown(a.RemoteADF)

	if local == "" && remote == "" {
		return
	}

	localLines := strings.Split(local, "\n")
	remoteLines := strings.Split(remote, "\n")

	// For push: local replaces remote. For pull: remote replaces local.
	var fromLabel, toLabel string
	var fromLines, toLines []string
	switch a.Block {
	case BlockFirstPull:
		fromLabel = "local: " + a.Node.File
		toLabel = "remote: " + a.Node.Key
		fromLines = localLines
		toLines = remoteLines
	default: // BlockFirstPush, BlockOverwrite
		fromLabel = "remote: " + a.Node.Key
		toLabel = "local: " + a.Node.File
		fromLines = remoteLines
		toLines = localLines
	}

	diff := SimpleDiff(fromLines, toLines)
	if len(diff) == 0 {
		fmt.Fprintf(os.Stderr, "\n  (content is identical)\n\n")
		return
	}

	fmt.Fprintf(os.Stderr, "\n--- %s\n+++ %s\n", fromLabel, toLabel)
	for _, line := range diff {
		fmt.Fprintf(os.Stderr, "%s\n", line)
	}
	fmt.Fprintln(os.Stderr)
}

// RemoteAsMarkdown converts RemoteADF to markdown for diffing.
// Returns empty string on nil ADF or conversion failure.
func RemoteAsMarkdown(adf []byte) string {
	if adf == nil || string(adf) == "null" {
		return ""
	}
	md, err := pipeline.ConvertADF(adf)
	if err != nil {
		return "(ADF conversion failed)"
	}
	return strings.TrimSpace(string(md))
}

// SimpleDiff produces a minimal line-level diff output.
// Uses a basic LCS approach suitable for small documents.
func SimpleDiff(from, to []string) []string {
	// Handle trivial cases
	if len(from) == 0 && len(to) == 0 {
		return nil
	}
	if len(from) == 0 {
		var out []string
		for _, l := range to {
			out = append(out, "+"+l)
		}
		return out
	}
	if len(to) == 0 {
		var out []string
		for _, l := range from {
			out = append(out, "-"+l)
		}
		return out
	}

	// LCS table
	m, n := len(from), len(to)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if from[i-1] == to[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	// Backtrack to produce diff
	var diff []string
	i, j := m, n
	for i > 0 || j > 0 {
		if i > 0 && j > 0 && from[i-1] == to[j-1] {
			diff = append(diff, " "+from[i-1])
			i--
			j--
		} else if j > 0 && (i == 0 || dp[i][j-1] >= dp[i-1][j]) {
			diff = append(diff, "+"+to[j-1])
			j--
		} else {
			diff = append(diff, "-"+from[i-1])
			i--
		}
	}

	// Reverse
	for l, r := 0, len(diff)-1; l < r; l, r = l+1, r-1 {
		diff[l], diff[r] = diff[r], diff[l]
	}

	return diff
}

// overridePrompt returns a block-reason-specific message explaining what override will do.
func overridePrompt(a Action) string {
	switch a.Block {
	case BlockFirstPush:
		return fmt.Sprintf("BLOCKED %s: Remote has existing content (first sync — no baseline).\n"+
			"  Override will PUSH local content, replacing the Jira description.", a.Node.Key)
	case BlockFirstPull:
		return fmt.Sprintf("BLOCKED %s: Local has existing content (first sync — no baseline).\n"+
			"  Override will PULL remote content, replacing your local file.", a.Node.Key)
	case BlockOverwrite:
		return fmt.Sprintf("BLOCKED %s: Remote description changed since last sync.\n"+
			"  Override will PUSH local content, discarding those remote changes.", a.Node.Key)
	default:
		return fmt.Sprintf("BLOCKED %s: %s", a.Node.Key, a.Reason)
	}
}

// promoteAction converts a blocked action to push/pull if override is allowed.
func promoteAction(a Action) Action {
	switch a.Block {
	case BlockFirstPush, BlockOverwrite:
		return Action{Node: a.Node, Kind: ActionPush,
			LocalContent: a.LocalContent, LocalHash: a.LocalHash,
			RemoteADF: a.RemoteADF, RemoteHash: a.RemoteHash,
			PlainText: a.PlainText}
	case BlockFirstPull:
		return Action{Node: a.Node, Kind: ActionPull,
			RemoteADF: a.RemoteADF, RemoteHash: a.RemoteHash}
	case BlockConflict:
		return a // conflicts require --resolve, not interactive override
	default:
		return a // Tier 3: stay blocked
	}
}
