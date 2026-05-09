package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
)

// NodeReading holds both local and remote state for a single forest node.
type NodeReading struct {
	Node           *forest.Node
	LocalContent   []byte // stripped of frontmatter
	LocalHash      string
	LocalLabel     string // from node.Label (frontmatter, heading, or filename)
	LocalErr       error
	RemoteADF      json.RawMessage
	RemoteHash     string
	RemoteMarkdown []byte // ADF converted to markdown (first-sync only)
	RemoteErr      error
	Baseline       *forest.NodeState // nil if never synced
	Mutable        bool              // lint + roundtrip passed
	RoundtripDiff  string            // first divergent line hint (empty if clean)
	LintIssues     []pipeline.LintIssue
}

// ActionKind classifies the engine's decision for a node.
type ActionKind int

const (
	ActionPush ActionKind = iota
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
	Node          *forest.Node
	Kind          ActionKind
	Reason        string
	Block         BlockReason
	RoundtripDiff string // first divergent line hint (empty if clean or not applicable)
	LocalHash     string
	RemoteHash    string
	RemoteADF     json.RawMessage
	LocalContent  []byte
	PlainText     bool // compile-failure fallback: push as plain text
}

// PlanOpts configures Plan behavior.
type PlanOpts struct {
	Direction string // "push" | "pull" | "both"
	Resolve   string // "" | "local" | "remote"
	PlainText bool   // compile-failure fallback: push as plain text
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
func Read(nodes []*forest.Node, p *pipeline.Pipeline,
	state *forest.State, forestDir string) ([]NodeReading, error) {

	if len(nodes) == 0 {
		return nil, nil
	}

	readings := make([]NodeReading, len(nodes))

	// Bounded worker pool
	sem := make(chan struct{}, readConcurrency)
	var wg sync.WaitGroup
	var stateMu sync.Mutex

	for i, n := range nodes {
		wg.Add(1)
		go func(idx int, node *forest.Node) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			readings[idx] = readNode(node, p, state, forestDir, &stateMu)
		}(i, n)
	}

	wg.Wait()

	return readings, nil
}

// readNode fetches local and remote state for a single node.
func readNode(node *forest.Node, p *pipeline.Pipeline,
	state *forest.State, forestDir string, stateMu *sync.Mutex) NodeReading {

	r := NodeReading{Node: node}
	r.LocalLabel = node.Label

	// Local content
	filePath := filepath.Join(forestDir, node.File)
	source, err := os.ReadFile(filePath)
	if err != nil {
		r.LocalErr = err
	} else {
		r.LocalContent = pipeline.StripFrontmatter(source)
		r.LocalHash = pipeline.ComputeLocalHash(r.LocalContent)
	}

	// Baseline from state (before mutability, which may add entries)
	if state != nil {
		stateMu.Lock()
		if ns, ok := state.Nodes[node.Key]; ok {
			r.Baseline = &ns
		}
		stateMu.Unlock()
	}

	// Mutability: lint + roundtrip for substantive content
	if IsSubstantiveLocal(r.LocalContent) {
		issues := pipeline.Lint(r.LocalContent, node.File)
		if len(issues) > 0 {
			r.LintIssues = issues
			r.Mutable = false
		} else if state != nil {
			stateMu.Lock()
			cached, ok := state.MutabilityCache(node.Key, r.LocalHash)
			stateMu.Unlock()
			if ok {
				r.Mutable = cached
				if !cached {
					r.RoundtripDiff = pipeline.FirstDivergence(r.LocalContent)
				}
			} else {
				clean, err := pipeline.CheckRoundtrip(r.LocalContent)
				r.Mutable = err == nil && clean
				if !r.Mutable {
					r.RoundtripDiff = pipeline.FirstDivergence(r.LocalContent)
				}
				stateMu.Lock()
				state.SetMutability(node.Key, r.LocalHash, r.Mutable)
				stateMu.Unlock()
			}
		} else {
			clean, err := pipeline.CheckRoundtrip(r.LocalContent)
			r.Mutable = err == nil && clean
			if !r.Mutable {
				r.RoundtripDiff = pipeline.FirstDivergence(r.LocalContent)
			}
		}
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

	// First-sync content comparison: convert remote ADF to markdown so Plan
	// can compare normalized content without I/O. Only when no baseline and
	// both sides have content (the first-sync blocked scenario).
	if r.Baseline == nil && r.RemoteADF != nil && r.LocalContent != nil {
		md, convErr := pipeline.ConvertADF(r.RemoteADF)
		if convErr == nil {
			r.RemoteMarkdown = md
		}
	}

	return r
}
