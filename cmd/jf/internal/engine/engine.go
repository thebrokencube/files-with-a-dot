package engine

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
)

// NodeReading holds both local and remote state for a single forest node.
type NodeReading struct {
	Node         *forest.Node
	LocalContent []byte            // stripped of frontmatter
	LocalHash    string
	LocalErr     error
	RemoteADF    json.RawMessage
	RemoteHash   string
	RemoteErr    error
	Baseline     *forest.NodeState // nil if never synced
}

// ActionKind classifies the engine's decision for a node.
type ActionKind int

const (
	ActionPush    ActionKind = iota
	ActionPull
	ActionSkip
	ActionBlocked
)

// String returns the action kind name.
func (k ActionKind) String() string {
	switch k {
	case ActionPush:
		return "push"
	case ActionPull:
		return "pull"
	case ActionSkip:
		return "skip"
	case ActionBlocked:
		return "blocked"
	default:
		return "unknown"
	}
}

// BlockReason classifies why an action is blocked.
type BlockReason int

const (
	BlockNone          BlockReason = iota
	BlockEmpty                     // Tier 3: empty local content
	BlockRemoteUnknown             // Tier 3: cannot reach Jira
	BlockLocalUnknown              // reserved
	BlockConflict                  // Tier 2b: both sides changed
	BlockFirstPush                 // Tier 2: first sync, remote has content
	BlockFirstPull                 // Tier 2: first sync, local has content
	BlockOverwrite                 // Tier 2: remote changed since last sync
)

// String returns the block reason name.
func (b BlockReason) String() string {
	switch b {
	case BlockNone:
		return ""
	case BlockEmpty:
		return "empty"
	case BlockRemoteUnknown:
		return "remote-unknown"
	case BlockLocalUnknown:
		return "local-unknown"
	case BlockConflict:
		return "conflict"
	case BlockFirstPush:
		return "first-push"
	case BlockFirstPull:
		return "first-pull"
	case BlockOverwrite:
		return "overwrite"
	default:
		return "unknown"
	}
}

// Action is the engine's decision for a single node.
type Action struct {
	Node         *forest.Node
	Kind         ActionKind
	Reason       string
	Block        BlockReason
	LocalHash    string
	RemoteHash   string
	RemoteADF    json.RawMessage
	LocalContent []byte
	PlainText    bool // compile-failure fallback: push as plain text
}

// PlanOpts configures Plan behavior.
type PlanOpts struct {
	Direction string // "push" | "pull" | "both"
	Resolve   string // "" | "local" | "remote"
}

// Result records the outcome of executing a single action.
type Result struct {
	Node    *forest.Node
	Kind    ActionKind
	Success bool
	Error   error
}

const readConcurrency = 5

// resolveDirection maps command direction + node sync mode to per-node direction.
func resolveDirection(node *forest.Node, opts PlanOpts) string {
	if opts.Direction == "both" {
		return node.Sync // respect node's declared mode
	}
	return opts.Direction // command overrides node mode
}

// Read fetches both local and remote state for every node.
// Uses a bounded worker pool (concurrency 5) for Jira API calls.
// Side effect: writes snapshot to .jf/snapshots/latest.json after all fetches.
// Logs warning (does not error) if snapshot write fails.
func Read(nodes []*forest.Node, p *pipeline.Pipeline,
	state *forest.State, forestDir string) ([]NodeReading, error) {

	if len(nodes) == 0 {
		return nil, nil
	}

	readings := make([]NodeReading, len(nodes))

	// Bounded worker pool
	sem := make(chan struct{}, readConcurrency)
	var wg sync.WaitGroup

	for i, n := range nodes {
		wg.Add(1)
		go func(idx int, node *forest.Node) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			readings[idx] = readNode(node, p, state, forestDir)
		}(i, n)
	}

	wg.Wait()

	// Write snapshot
	writeSnapshot(readings, forestDir)

	return readings, nil
}

// readNode fetches local and remote state for a single node.
func readNode(node *forest.Node, p *pipeline.Pipeline,
	state *forest.State, forestDir string) NodeReading {

	r := NodeReading{Node: node}

	// Local content
	filePath := filepath.Join(forestDir, node.File)
	source, err := os.ReadFile(filePath)
	if err != nil {
		r.LocalErr = err
	} else {
		r.LocalContent = pipeline.StripFrontmatter(source)
		r.LocalHash = forest.ComputeHash(r.LocalContent)
	}

	// Remote content (skip TBD keys)
	if !forest.IsTBD(node.Key) {
		viewJSON, err := p.View(node.Key, "description", true)
		if err != nil {
			r.RemoteErr = err
		} else {
			adf, extractErr := pipeline.ExtractDescriptionADF(viewJSON)
			if extractErr != nil {
				r.RemoteErr = extractErr
			} else if adf != nil {
				r.RemoteADF = adf
				r.RemoteHash = forest.ComputeHash(adf)
			}
		}
	}

	// Baseline from state
	if state != nil {
		if ns, ok := state.Nodes[node.Key]; ok {
			r.Baseline = &ns
		}
	}

	return r
}

// snapshotEntry is the JSON format for each node in the snapshot file.
type snapshotEntry struct {
	Key          string          `json:"key"`
	File         string          `json:"file"`
	RemoteADF    json.RawMessage `json:"remote_adf,omitempty"`
	RemoteHash   string          `json:"remote_hash,omitempty"`
	LocalHash    string          `json:"local_hash,omitempty"`
	LocalContent string          `json:"local_content,omitempty"` // base64-encoded
	RemoteErr    string          `json:"remote_err,omitempty"`
}

// writeSnapshot persists a snapshot to .jf/snapshots/latest.json.
// Warns on failure; never returns an error.
func writeSnapshot(readings []NodeReading, forestDir string) {
	entries := make([]snapshotEntry, len(readings))
	for i, r := range readings {
		e := snapshotEntry{
			Key:        r.Node.Key,
			File:       r.Node.File,
			RemoteADF:  r.RemoteADF,
			RemoteHash: r.RemoteHash,
			LocalHash:  r.LocalHash,
		}
		if len(r.LocalContent) > 0 {
			e.LocalContent = base64.StdEncoding.EncodeToString(r.LocalContent)
		}
		if r.RemoteErr != nil {
			e.RemoteErr = r.RemoteErr.Error()
		}
		entries[i] = e
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to marshal snapshot: %v\n", err)
		return
	}
	data = append(data, '\n')

	snapDir := filepath.Join(forestDir, ".jf", "snapshots")
	if err := os.MkdirAll(snapDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to create snapshot dir: %v\n", err)
		return
	}

	snapPath := filepath.Join(snapDir, "latest.json")
	if err := os.WriteFile(snapPath, data, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: failed to write snapshot: %v\n", err)
	}
}
