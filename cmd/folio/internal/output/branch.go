package output

import (
	"sort"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

// BranchNode represents a target with an associated branch in the branch topology.
type BranchNode struct {
	ID       string        `json:"id"`
	Branch   string        `json:"branch"`
	Base     string        `json:"base"`
	PR       string        `json:"pr,omitempty"`
	Status   string        `json:"status,omitempty"`
	StaleVia string        `json:"stale_via,omitempty"`
	Children []*BranchNode `json:"children"`
}

// BranchRoot represents a root branch in the topology (e.g., "main").
type BranchRoot struct {
	Base     string        `json:"base"`
	Children []*BranchNode `json:"children"`
}

// BranchTopology is the complete branch tree.
type BranchTopology struct {
	Roots []*BranchRoot `json:"roots"`
}

// DeriveBase returns the base branch for a target by looking at its first
// blocked_by dependency that has a Branch set. Falls back to "main".
func DeriveBase(tid string, targets map[string]config.Target) string {
	t := targets[tid]
	for _, dep := range t.BlockedBy {
		if dt, ok := targets[dep]; ok && dt.Branch != "" {
			return dt.Branch
		}
	}
	return "main"
}

// BuildBranchTopology constructs the branch tree from targets with Branch set.
// Children arrays are always non-nil (empty slice, never null).
func BuildBranchTopology(targets map[string]config.Target) *BranchTopology {
	// Collect target IDs with branch set
	var withBranch []string
	for tid, t := range targets {
		if t.Branch != "" {
			withBranch = append(withBranch, tid)
		}
	}
	sort.Strings(withBranch)

	if len(withBranch) == 0 {
		return &BranchTopology{Roots: []*BranchRoot{}}
	}

	// Build child map: base branch name → list of target IDs branching from it
	children := make(map[string][]string)
	for _, tid := range withBranch {
		base := DeriveBase(tid, targets)
		children[base] = append(children[base], tid)
	}

	// Identify roots: base branches that are not themselves a target's branch name
	branchToTid := make(map[string]string)
	for _, tid := range withBranch {
		branchToTid[targets[tid].Branch] = tid
	}

	var rootBases []string
	for base := range children {
		if _, isBranch := branchToTid[base]; !isBranch {
			rootBases = append(rootBases, base)
		}
	}
	sort.Strings(rootBases)

	// Recursive node builder
	var buildNode func(tid string) *BranchNode
	buildNode = func(tid string) *BranchNode {
		t := targets[tid]
		node := &BranchNode{
			ID:       tid,
			Branch:   t.Branch,
			Base:     DeriveBase(tid, targets),
			PR:       t.PR,
			Children: []*BranchNode{},
		}
		for _, kid := range children[t.Branch] {
			node.Children = append(node.Children, buildNode(kid))
		}
		return node
	}

	var roots []*BranchRoot
	for _, base := range rootBases {
		root := &BranchRoot{
			Base:     base,
			Children: []*BranchNode{},
		}
		for _, kid := range children[base] {
			root.Children = append(root.Children, buildNode(kid))
		}
		roots = append(roots, root)
	}

	return &BranchTopology{Roots: roots}
}

// PropagationOrder returns target IDs in pre-order traversal of the topology
// tree — root-to-leaf order suitable for stacked-pr propagation.
func PropagationOrder(bt *BranchTopology) []string {
	var order []string
	var walk func(node *BranchNode)
	walk = func(node *BranchNode) {
		order = append(order, node.ID)
		for _, child := range node.Children {
			walk(child)
		}
	}
	for _, root := range bt.Roots {
		for _, child := range root.Children {
			walk(child)
		}
	}
	return order
}
