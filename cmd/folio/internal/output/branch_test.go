package output

import (
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

func TestDeriveBaseFallback(t *testing.T) {
	targets := map[string]config.Target{
		"standalone": {Branch: "feat-standalone"},
	}
	base := DeriveBase("standalone", targets)
	if base != "main" {
		t.Errorf("DeriveBase fallback = %q, want main", base)
	}
}

func TestDeriveBaseFromBlockedBy(t *testing.T) {
	targets := map[string]config.Target{
		"parent": {Branch: "feat-parent"},
		"child":  {Branch: "feat-child", BlockedBy: []string{"parent"}},
	}
	base := DeriveBase("child", targets)
	if base != "feat-parent" {
		t.Errorf("DeriveBase = %q, want feat-parent", base)
	}
}

func TestDeriveBaseSkipsNoBranch(t *testing.T) {
	targets := map[string]config.Target{
		"no-branch": {Transform: "distill"},
		"parent":    {Branch: "feat-parent"},
		"child":     {Branch: "feat-child", BlockedBy: []string{"no-branch", "parent"}},
	}
	base := DeriveBase("child", targets)
	if base != "feat-parent" {
		t.Errorf("DeriveBase = %q, want feat-parent (should skip no-branch)", base)
	}
}

func TestBuildBranchTopologyEmpty(t *testing.T) {
	targets := map[string]config.Target{
		"no-branch": {Transform: "distill"},
	}
	bt := BuildBranchTopology(targets)
	if len(bt.Roots) != 0 {
		t.Errorf("expected 0 roots, got %d", len(bt.Roots))
	}
}

func TestBuildBranchTopologyStacked(t *testing.T) {
	targets := map[string]config.Target{
		"docs-tooling": {
			Branch: "feat-tooling",
			PR:     "#100",
		},
		"docs-proposal": {
			Branch:    "feat-proposal",
			PR:        "#200",
			BlockedBy: []string{"docs-tooling"},
		},
	}

	bt := BuildBranchTopology(targets)

	if len(bt.Roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(bt.Roots))
	}
	root := bt.Roots[0]
	if root.Base != "main" {
		t.Errorf("root base = %q, want main", root.Base)
	}
	if len(root.Children) != 1 {
		t.Fatalf("expected 1 child of main, got %d", len(root.Children))
	}

	tooling := root.Children[0]
	if tooling.ID != "docs-tooling" {
		t.Errorf("first child ID = %q, want docs-tooling", tooling.ID)
	}
	if tooling.Branch != "feat-tooling" {
		t.Errorf("branch = %q, want feat-tooling", tooling.Branch)
	}
	if tooling.Base != "main" {
		t.Errorf("base = %q, want main", tooling.Base)
	}
	if tooling.PR != "#100" {
		t.Errorf("PR = %q, want #100", tooling.PR)
	}

	if len(tooling.Children) != 1 {
		t.Fatalf("expected 1 child of docs-tooling, got %d", len(tooling.Children))
	}

	proposal := tooling.Children[0]
	if proposal.ID != "docs-proposal" {
		t.Errorf("child ID = %q, want docs-proposal", proposal.ID)
	}
	if proposal.Base != "feat-tooling" {
		t.Errorf("child base = %q, want feat-tooling", proposal.Base)
	}
	if len(proposal.Children) != 0 {
		t.Errorf("expected 0 children on leaf, got %d", len(proposal.Children))
	}
}

func TestBuildBranchTopologyUnstacked(t *testing.T) {
	targets := map[string]config.Target{
		"alpha": {Branch: "feat-alpha"},
		"beta":  {Branch: "feat-beta"},
	}

	bt := BuildBranchTopology(targets)

	if len(bt.Roots) != 1 {
		t.Fatalf("expected 1 root (main), got %d", len(bt.Roots))
	}
	if bt.Roots[0].Base != "main" {
		t.Errorf("root base = %q, want main", bt.Roots[0].Base)
	}
	if len(bt.Roots[0].Children) != 2 {
		t.Errorf("expected 2 children of main, got %d", len(bt.Roots[0].Children))
	}
}

func TestBuildBranchTopologyChildrenNeverNil(t *testing.T) {
	targets := map[string]config.Target{
		"leaf": {Branch: "feat-leaf"},
	}

	bt := BuildBranchTopology(targets)
	if bt.Roots[0].Children[0].Children == nil {
		t.Error("leaf Children should be [], not nil")
	}
}

func TestPropagationOrder(t *testing.T) {
	targets := map[string]config.Target{
		"root": {Branch: "feat-root"},
		"child-a": {
			Branch:    "feat-a",
			BlockedBy: []string{"root"},
		},
		"child-b": {
			Branch:    "feat-b",
			BlockedBy: []string{"root"},
		},
		"grandchild": {
			Branch:    "feat-gc",
			BlockedBy: []string{"child-a"},
		},
	}

	bt := BuildBranchTopology(targets)
	order := PropagationOrder(bt)

	if len(order) != 4 {
		t.Fatalf("expected 4 targets in order, got %d: %v", len(order), order)
	}

	// Root must come first
	if order[0] != "root" {
		t.Errorf("first = %q, want root", order[0])
	}

	// child-a must come before grandchild
	aIdx, gcIdx := -1, -1
	for i, id := range order {
		if id == "child-a" {
			aIdx = i
		}
		if id == "grandchild" {
			gcIdx = i
		}
	}
	if aIdx >= gcIdx {
		t.Errorf("child-a (idx %d) should come before grandchild (idx %d)", aIdx, gcIdx)
	}
}

func TestPropagationOrderEmpty(t *testing.T) {
	bt := &BranchTopology{Roots: []*BranchRoot{}}
	order := PropagationOrder(bt)
	if len(order) != 0 {
		t.Errorf("expected empty order, got %v", order)
	}
}
