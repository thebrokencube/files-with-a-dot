package engine

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
)

var (
	pushOpts      = PlanOpts{Direction: "push"}
	pullOpts      = PlanOpts{Direction: "pull"}
	bothOpts      = PlanOpts{Direction: "both"}
	resolveLocal  = PlanOpts{Direction: "both", Resolve: "local"}
	resolveRemote = PlanOpts{Direction: "both", Resolve: "remote"}

	substantiveADF = json.RawMessage(`{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"text":"Hello"}]}]}`)
	emptyADF       = json.RawMessage(`{"content":[]}`)

	localContent = []byte("Real content here")
	emptyContent = []byte("")

	baselineLocal  = "localhash123"
	baselineRemote = "remotehash456"
)

func node(key, sync string) *forest.Node {
	return &forest.Node{Key: key, File: key + ".md", Sync: sync}
}

func TestPlan(t *testing.T) {
	tests := []struct {
		name      string
		reading   NodeReading
		opts      PlanOpts
		wantKind  ActionKind
		wantBlock BlockReason
	}{
		// TBD
		{"tbd_key_skip", NodeReading{
			Node: node("TBD", "push"),
		}, pushOpts, ActionSkip, BlockNone},

		// Emptiness (Tier 3)
		{"push_empty_blocked", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: emptyContent,
		}, pushOpts, ActionBlocked, BlockEmpty},

		{"pull_empty_remote_skip", NodeReading{
			Node: node("KEY-1", "pull"), RemoteADF: emptyADF,
		}, pullOpts, ActionSkip, BlockNone},

		{"both_empty_local_remote_has_content_pulls", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: emptyContent,
			RemoteADF: substantiveADF, RemoteHash: "rh",
		}, bothOpts, ActionPull, BlockNone},

		{"both_empty_local_remote_empty_blocked", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: emptyContent,
			RemoteADF: emptyADF,
		}, bothOpts, ActionBlocked, BlockEmpty},

		{"both_empty_local_remote_unreachable_blocked", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: emptyContent,
			RemoteErr: errMock,
		}, bothOpts, ActionBlocked, BlockEmpty},

		{"both_empty_remote_falls_to_push", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: localContent, LocalHash: "lh",
			RemoteADF: emptyADF, Mutable: true,
		}, bothOpts, ActionPush, BlockNone},

		// Remote unreachable (Tier 3 for push, Skip for pull)
		{"push_remote_err_blocked", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: localContent, LocalHash: "lh",
			RemoteErr: errMock, Mutable: true,
		}, pushOpts, ActionBlocked, BlockRemoteUnknown},

		{"pull_remote_err_skip", NodeReading{
			Node: node("KEY-1", "pull"), RemoteErr: errMock,
		}, pullOpts, ActionSkip, BlockNone},

		// First sync — no content match (RemoteMarkdown nil or content differs)
		{"first_push_remote_has_content_blocked", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: localContent, LocalHash: "lh",
			RemoteADF: substantiveADF, RemoteHash: "rh", Mutable: true,
		}, pushOpts, ActionBlocked, BlockFirstPush},

		{"first_push_remote_empty_push", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: localContent, LocalHash: "lh",
			RemoteADF: emptyADF, Mutable: true,
		}, pushOpts, ActionPush, BlockNone},

		{"first_pull_local_has_content_blocked", NodeReading{
			Node: node("KEY-1", "pull"), LocalContent: localContent,
			RemoteADF: substantiveADF, RemoteHash: "rh",
		}, pullOpts, ActionBlocked, BlockFirstPull},

		{"first_pull_local_empty_pull", NodeReading{
			Node: node("KEY-1", "pull"), LocalContent: emptyContent,
			RemoteADF: substantiveADF, RemoteHash: "rh",
		}, pullOpts, ActionPull, BlockNone},

		{"both_first_sync_remote_has_content_blocked", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: localContent, LocalHash: "lh",
			RemoteADF: substantiveADF, RemoteHash: "rh", Mutable: true,
		}, bothOpts, ActionBlocked, BlockFirstPush},

		{"both_first_sync_local_has_content_remote_empty_push", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: localContent, LocalHash: "lh",
			RemoteADF: emptyADF, Mutable: true,
		}, bothOpts, ActionPush, BlockNone},

		// First sync — content match via RemoteMarkdown (auto-push to establish baseline)
		{"first_push_content_match_auto_push", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: []byte("Hello world"),
			LocalHash: "lh", RemoteADF: substantiveADF, RemoteHash: "rh",
			RemoteMarkdown: []byte("Hello world"), Mutable: true,
		}, pushOpts, ActionPush, BlockNone},

		{"first_push_content_match_trailing_whitespace", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: []byte("Hello world"),
			LocalHash: "lh", RemoteADF: substantiveADF, RemoteHash: "rh",
			RemoteMarkdown: []byte("Hello world  \n"), Mutable: true,
		}, pushOpts, ActionPush, BlockNone},

		{"first_push_content_match_extra_blank_lines", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: []byte("# Title\n\nParagraph"),
			LocalHash: "lh", RemoteADF: substantiveADF, RemoteHash: "rh",
			RemoteMarkdown: []byte("# Title\n\n\n\nParagraph\n"), Mutable: true,
		}, pushOpts, ActionPush, BlockNone},

		{"first_push_content_differs_blocked", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: []byte("Local version"),
			LocalHash: "lh", RemoteADF: substantiveADF, RemoteHash: "rh",
			RemoteMarkdown: []byte("Remote version"), Mutable: true,
		}, pushOpts, ActionBlocked, BlockFirstPush},

		{"first_pull_content_match_auto_pull", NodeReading{
			Node: node("KEY-1", "pull"), LocalContent: []byte("Same content"),
			RemoteADF: substantiveADF, RemoteHash: "rh",
			RemoteMarkdown: []byte("Same content  \n"),
		}, pullOpts, ActionPull, BlockNone},

		{"first_pull_content_differs_blocked", NodeReading{
			Node: node("KEY-1", "pull"), LocalContent: []byte("Local text"),
			RemoteADF: substantiveADF, RemoteHash: "rh",
			RemoteMarkdown: []byte("Different remote text"),
		}, pullOpts, ActionBlocked, BlockFirstPull},

		{"both_first_sync_content_match_auto_push", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: []byte("Matching"),
			LocalHash: "lh", RemoteADF: substantiveADF, RemoteHash: "rh",
			RemoteMarkdown: []byte("Matching\n"), Mutable: true,
		}, bothOpts, ActionPush, BlockNone},

		{"both_first_sync_content_differs_blocked", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: []byte("Local"),
			LocalHash: "lh", RemoteADF: substantiveADF, RemoteHash: "rh",
			RemoteMarkdown: []byte("Remote"), Mutable: true,
		}, bothOpts, ActionBlocked, BlockFirstPush},

		// First sync — RemoteMarkdown nil (ADF conversion failed) still blocks
		{"first_push_no_remote_md_still_blocks", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: localContent, LocalHash: "lh",
			RemoteADF: substantiveADF, RemoteHash: "rh",
			RemoteMarkdown: nil, Mutable: true,
		}, pushOpts, ActionBlocked, BlockFirstPush},

		// With baseline — individual directions
		{"baseline_local_changed_push", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: localContent, LocalHash: "newhash",
			Baseline:   &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
			RemoteHash: baselineRemote, Mutable: true,
		}, pushOpts, ActionPush, BlockNone},

		{"baseline_remote_changed_pull", NodeReading{
			Node: node("KEY-1", "pull"), LocalHash: baselineLocal,
			RemoteADF: substantiveADF, RemoteHash: "newhash",
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, pullOpts, ActionPull, BlockNone},

		{"push_remote_changed_blocked", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: localContent, LocalHash: "newhash",
			RemoteHash: "newremote", Mutable: true,
			Baseline:   &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, pushOpts, ActionBlocked, BlockOverwrite},

		{"pull_local_changed_blocked", NodeReading{
			Node: node("KEY-1", "pull"), LocalContent: localContent, LocalHash: "newhash",
			RemoteADF: substantiveADF, RemoteHash: baselineRemote,
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, pullOpts, ActionBlocked, BlockOverwrite},

		{"baseline_neither_changed_skip_push", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: localContent, LocalHash: baselineLocal,
			RemoteHash: baselineRemote, Mutable: true,
			Baseline:   &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, pushOpts, ActionSkip, BlockNone},

		{"baseline_neither_changed_skip_pull", NodeReading{
			Node: node("KEY-1", "pull"), LocalHash: baselineLocal,
			RemoteHash: baselineRemote,
			Baseline:   &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, pullOpts, ActionSkip, BlockNone},

		// With baseline — both direction
		{"both_local_only_push", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: localContent, LocalHash: "newhash",
			RemoteHash: baselineRemote, Mutable: true,
			Baseline:   &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, bothOpts, ActionPush, BlockNone},

		{"both_remote_only_pull", NodeReading{
			Node: node("KEY-1", "both"), LocalHash: baselineLocal,
			RemoteADF: substantiveADF, RemoteHash: "newhash",
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, bothOpts, ActionPull, BlockNone},

		{"both_neither_skip", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: localContent, LocalHash: baselineLocal,
			RemoteADF: substantiveADF, RemoteHash: baselineRemote, Mutable: true,
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, bothOpts, ActionSkip, BlockNone},

		{"both_changed_conflict", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: localContent, LocalHash: "newhash",
			RemoteADF: substantiveADF, RemoteHash: "newremote", Mutable: true,
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, bothOpts, ActionBlocked, BlockConflict},

		{"both_changed_resolve_local", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: localContent, LocalHash: "newhash",
			RemoteADF: substantiveADF, RemoteHash: "newremote", Mutable: true,
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, resolveLocal, ActionPush, BlockNone},

		{"both_changed_resolve_remote", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: localContent, LocalHash: "newhash",
			RemoteADF: substantiveADF, RemoteHash: "newremote", Mutable: true,
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, resolveRemote, ActionPull, BlockNone},

		// Empty remoteHash baseline (Track 1 transition)
		{"baseline_empty_remote_hash_no_false_conflict", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: localContent, LocalHash: "newhash",
			RemoteHash: "anyhash", Mutable: true,
			Baseline:   &forest.NodeState{LocalHash: baselineLocal, RemoteHash: ""},
		}, pushOpts, ActionPush, BlockNone},

		// Direction resolution: "both" opts respects node sync mode
		{"both_opts_push_node_uses_push_rules", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: emptyContent,
		}, bothOpts, ActionBlocked, BlockEmpty},

		{"both_opts_pull_node_remote_empty_skips", NodeReading{
			Node: node("KEY-1", "pull"), RemoteADF: emptyADF,
		}, bothOpts, ActionSkip, BlockNone},

		// Mutability guard
		{"read_only_push_skipped", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: localContent, LocalHash: "lh",
			Mutable: false, LintIssues: []pipeline.LintIssue{{Line: 1, Message: "tables not supported"}},
		}, pushOpts, ActionSkip, BlockNone},

		{"read_only_both_demoted_to_pull", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: localContent, LocalHash: baselineLocal,
			RemoteADF: substantiveADF, RemoteHash: "newremote",
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
			Mutable: false, LintIssues: []pipeline.LintIssue{{Line: 1, Message: "tables not supported"}},
		}, bothOpts, ActionPull, BlockNone},

		{"roundtrip_only_failure_skipped", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: localContent, LocalHash: "lh",
			Mutable: false,
		}, pushOpts, ActionSkip, BlockNone},

		{"plain_text_bypasses_mutability", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: localContent, LocalHash: "lh",
			RemoteADF: emptyADF, Mutable: false,
		}, PlanOpts{Direction: "push", PlainText: true}, ActionPush, BlockNone},

		{"mutable_push_proceeds", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: localContent, LocalHash: "lh",
			RemoteADF: emptyADF, Mutable: true,
		}, pushOpts, ActionPush, BlockNone},

		// Demotion: read-only both→pull skips first-sync local content guard
		{"read_only_both_first_sync_demoted_pulls", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: localContent, LocalHash: "lh",
			RemoteADF: substantiveADF, RemoteHash: "rh",
			Mutable: false, LintIssues: []pipeline.LintIssue{{Line: 1, Message: "tables not supported"}},
		}, bothOpts, ActionPull, BlockNone},

		// Demotion: read-only both→pull skips overwrite guard with baseline
		{"read_only_both_baseline_local_changed_demoted_pulls", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: localContent, LocalHash: "newhash",
			RemoteADF: substantiveADF, RemoteHash: "newremote",
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
			Mutable: false, LintIssues: []pipeline.LintIssue{{Line: 1, Message: "tables not supported"}},
		}, bothOpts, ActionPull, BlockNone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actions := Plan([]NodeReading{tt.reading}, tt.opts)
			if len(actions) != 1 {
				t.Fatalf("Plan returned %d actions, want 1", len(actions))
			}
			a := actions[0]
			if a.Kind != tt.wantKind {
				t.Errorf("Kind = %v, want %v (reason: %s)", a.Kind, tt.wantKind, a.Reason)
			}
			if a.Block != tt.wantBlock {
				t.Errorf("Block = %v, want %v (reason: %s)", a.Block, tt.wantBlock, a.Reason)
			}
		})
	}
}

// Property: Plan never produces ActionPush when local content is empty
func TestPlanNeverPushesEmpty(t *testing.T) {
	emptyVariants := [][]byte{
		nil,
		{},
		[]byte(""),
		[]byte("   \n  "),
		[]byte("TBD"),
		[]byte("todo"),
		[]byte("WIP"),
		[]byte("# Heading only"),
		[]byte("# H1\n## H2"),
	}

	allOpts := []PlanOpts{pushOpts, pullOpts, bothOpts, resolveLocal, resolveRemote}

	for _, content := range emptyVariants {
		for _, opts := range allOpts {
			readings := []NodeReading{{
				Node:         node("KEY-1", "both"),
				LocalContent: content,
				RemoteADF:    substantiveADF,
				RemoteHash:   "rh",
				Baseline:     &forest.NodeState{LocalHash: "old", RemoteHash: "rh"},
			}}

			actions := Plan(readings, opts)
			for _, a := range actions {
				if a.Kind == ActionPush {
					t.Errorf("Plan produced ActionPush with empty content %q opts=%v", string(content), opts)
				}
			}
		}
	}
}

// Property: Plan output length always equals input length
func TestPlanOutputLengthEqualsInput(t *testing.T) {
	sizes := []int{0, 1, 5, 20}
	for _, size := range sizes {
		readings := make([]NodeReading, size)
		for i := range readings {
			readings[i] = NodeReading{
				Node:         node("KEY-1", "push"),
				LocalContent: localContent,
				LocalHash:    "lh",
			}
		}
		actions := Plan(readings, pushOpts)
		if len(actions) != size {
			t.Errorf("len(Plan(%d readings)) = %d", size, len(actions))
		}
	}
}

var errMock = fmt.Errorf("mock error")

// TestNormalizedContentEqual verifies the comparison helper absorbs
// ADF roundtrip noise: trailing whitespace, blank line collapsing, trimming.
func TestNormalizedContentEqual(t *testing.T) {
	tests := []struct {
		name string
		a, b []byte
		want bool
	}{
		{"identical", []byte("hello"), []byte("hello"), true},
		{"trailing_whitespace", []byte("hello"), []byte("hello  \n"), true},
		{"trailing_tabs", []byte("line1\nline2"), []byte("line1\t\nline2\t\n"), true},
		{"extra_blank_lines", []byte("a\n\nb"), []byte("a\n\n\n\nb"), true},
		{"leading_trailing_newlines", []byte("content"), []byte("\n\ncontent\n\n"), true},
		{"different_content", []byte("hello"), []byte("world"), false},
		{"subset", []byte("hello world"), []byte("hello"), false},
		{"empty_both", []byte(""), []byte(""), true},
		{"empty_vs_whitespace", []byte(""), []byte("   \n\n"), true},
		{"mixed_noise", []byte("# Title\n\nBody"), []byte("# Title  \n\n\n\nBody  \n"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizedContentEqual(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("normalizedContentEqual(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestFirstSyncContentMatchReason verifies that auto-push/pull on content
// match includes the correct reason string for debugging.
func TestFirstSyncContentMatchReason(t *testing.T) {
	matchContent := []byte("Same content")
	matchRemoteMD := []byte("Same content  \n")

	tests := []struct {
		name       string
		opts       PlanOpts
		wantKind   ActionKind
		wantReason string
	}{
		{"push_match", pushOpts, ActionPush, "first sync, content matches — establishing baseline"},
		{"pull_match", pullOpts, ActionPull, "first sync, content matches — establishing baseline"},
		{"both_match", bothOpts, ActionPush, "first sync, content matches — establishing baseline"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reading := NodeReading{
				Node: node("KEY-1", tt.opts.Direction), LocalContent: matchContent,
				LocalHash: "lh", RemoteADF: substantiveADF, RemoteHash: "rh",
				RemoteMarkdown: matchRemoteMD, Mutable: true,
			}
			actions := Plan([]NodeReading{reading}, tt.opts)
			a := actions[0]
			if a.Kind != tt.wantKind {
				t.Errorf("Kind = %v, want %v", a.Kind, tt.wantKind)
			}
			if a.Reason != tt.wantReason {
				t.Errorf("Reason = %q, want %q", a.Reason, tt.wantReason)
			}
		})
	}
}

// TestFirstSyncContentMatchCarriesHashes verifies that auto-push on content
// match propagates both local and remote hashes for state recording.
func TestFirstSyncContentMatchCarriesHashes(t *testing.T) {
	reading := NodeReading{
		Node: node("KEY-1", "push"), LocalContent: []byte("content"),
		LocalHash: "local-hash-123", RemoteADF: substantiveADF, RemoteHash: "remote-hash-456",
		RemoteMarkdown: []byte("content\n"), Mutable: true,
	}
	actions := Plan([]NodeReading{reading}, pushOpts)
	a := actions[0]
	if a.Kind != ActionPush {
		t.Fatalf("expected ActionPush, got %v", a.Kind)
	}
	if a.LocalHash != "local-hash-123" {
		t.Errorf("LocalHash = %q, want %q", a.LocalHash, "local-hash-123")
	}
	if a.RemoteHash != "remote-hash-456" {
		t.Errorf("RemoteHash = %q, want %q", a.RemoteHash, "remote-hash-456")
	}
	if a.RemoteADF == nil {
		t.Error("RemoteADF should be carried through for post-push state recording")
	}
}
