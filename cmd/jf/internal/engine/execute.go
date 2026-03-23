package engine

import (
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

	compiled, err := p.Compile(a.Node.Key, source, a.Node.Label)
	if err != nil {
		if a.PlainText {
			fmt.Fprintf(os.Stderr, "⚠ %s: conversion failed, pushing as plain text\n", a.Node.Key)
			compiled = buildPlainTextPayload(a.Node.Key, source)
		} else {
			return Result{Node: a.Node, Kind: ActionPush, Success: false, Error: err}
		}
	}

	if err := p.Push(compiled); err != nil {
		return Result{Node: a.Node, Kind: ActionPush, Success: false, Error: err}
	}

	// Record sync
	if state != nil {
		localHash := forest.ComputeHash(pipeline.StripFrontmatter(source))
		state.RecordSync(a.Node.Key, "push", localHash, a.RemoteHash)
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
	content, err := mergeWithFrontmatter(filePath, md)
	if err != nil {
		return Result{Node: a.Node, Kind: ActionPull, Success: false, Error: err}
	}

	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return Result{Node: a.Node, Kind: ActionPull, Success: false, Error: err}
	}

	// Record sync
	if state != nil {
		localHash := forest.ComputeHash(pipeline.StripFrontmatter(content))
		remoteHash := forest.ComputeHash(a.RemoteADF)
		state.RecordSync(a.Node.Key, "pull", localHash, remoteHash)
	}

	return Result{Node: a.Node, Kind: ActionPull, Success: true}
}

// mergeWithFrontmatter preserves YAML frontmatter from the existing file
// and replaces the content below the closing fence with pulled content.
func mergeWithFrontmatter(filePath string, pulled []byte) ([]byte, error) {
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

// buildPlainTextPayload creates a plain-text ADF payload for push fallback.
func buildPlainTextPayload(key string, source []byte) []byte {
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

func isTier3(b BlockReason) bool {
	return b == BlockEmpty || b == BlockRemoteUnknown
}

// confirmOverride prompts for interactive TTY override of Tier 2 blocks.
func confirmOverride(a Action) bool {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false // agents and pipes: always blocked
	}
	prompt := overridePrompt(a)
	fmt.Fprintf(os.Stderr, "%s\nOverride? [y/N]: ", prompt)
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(strings.TrimSpace(response)) == "y"
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
