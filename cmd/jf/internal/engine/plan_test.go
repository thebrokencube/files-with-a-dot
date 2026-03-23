package engine

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
)

var (
	pushOpts     = PlanOpts{Direction: "push"}
	pullOpts     = PlanOpts{Direction: "pull"}
	bothOpts     = PlanOpts{Direction: "both"}
	resolveLocal = PlanOpts{Direction: "both", Resolve: "local"}
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
			RemoteADF: emptyADF,
		}, bothOpts, ActionPush, BlockNone},

		// Remote unreachable (Tier 3 for push, Skip for pull)
		{"push_remote_err_blocked", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: localContent, LocalHash: "lh",
			RemoteErr: errMock,
		}, pushOpts, ActionBlocked, BlockRemoteUnknown},

		{"pull_remote_err_skip", NodeReading{
			Node: node("KEY-1", "pull"), RemoteErr: errMock,
		}, pullOpts, ActionSkip, BlockNone},

		// First sync
		{"first_push_remote_has_content_blocked", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: localContent, LocalHash: "lh",
			RemoteADF: substantiveADF, RemoteHash: "rh",
		}, pushOpts, ActionBlocked, BlockFirstPush},

		{"first_push_remote_empty_push", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: localContent, LocalHash: "lh",
			RemoteADF: emptyADF,
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
			RemoteADF: substantiveADF, RemoteHash: "rh",
		}, bothOpts, ActionBlocked, BlockFirstPush},

		{"both_first_sync_local_has_content_remote_empty_push", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: localContent, LocalHash: "lh",
			RemoteADF: emptyADF,
		}, bothOpts, ActionPush, BlockNone},

		// With baseline — individual directions
		{"baseline_local_changed_push", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: localContent, LocalHash: "newhash",
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
			RemoteHash: baselineRemote,
		}, pushOpts, ActionPush, BlockNone},

		{"baseline_remote_changed_pull", NodeReading{
			Node: node("KEY-1", "pull"), LocalHash: baselineLocal,
			RemoteADF: substantiveADF, RemoteHash: "newhash",
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, pullOpts, ActionPull, BlockNone},

		{"push_remote_changed_blocked", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: localContent, LocalHash: "newhash",
			RemoteHash: "newremote",
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, pushOpts, ActionBlocked, BlockOverwrite},

		{"pull_local_changed_blocked", NodeReading{
			Node: node("KEY-1", "pull"), LocalContent: localContent, LocalHash: "newhash",
			RemoteADF: substantiveADF, RemoteHash: baselineRemote,
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, pullOpts, ActionBlocked, BlockOverwrite},

		{"baseline_neither_changed_skip_push", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: localContent, LocalHash: baselineLocal,
			RemoteHash: baselineRemote,
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, pushOpts, ActionSkip, BlockNone},

		{"baseline_neither_changed_skip_pull", NodeReading{
			Node: node("KEY-1", "pull"), LocalHash: baselineLocal,
			RemoteHash: baselineRemote,
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, pullOpts, ActionSkip, BlockNone},

		// With baseline — both direction
		{"both_local_only_push", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: localContent, LocalHash: "newhash",
			RemoteHash: baselineRemote,
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, bothOpts, ActionPush, BlockNone},

		{"both_remote_only_pull", NodeReading{
			Node: node("KEY-1", "both"), LocalHash: baselineLocal,
			RemoteADF: substantiveADF, RemoteHash: "newhash",
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, bothOpts, ActionPull, BlockNone},

		{"both_neither_skip", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: localContent, LocalHash: baselineLocal,
			RemoteADF: substantiveADF, RemoteHash: baselineRemote,
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, bothOpts, ActionSkip, BlockNone},

		{"both_changed_conflict", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: localContent, LocalHash: "newhash",
			RemoteADF: substantiveADF, RemoteHash: "newremote",
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, bothOpts, ActionBlocked, BlockConflict},

		{"both_changed_resolve_local", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: localContent, LocalHash: "newhash",
			RemoteADF: substantiveADF, RemoteHash: "newremote",
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, resolveLocal, ActionPush, BlockNone},

		{"both_changed_resolve_remote", NodeReading{
			Node: node("KEY-1", "both"), LocalContent: localContent, LocalHash: "newhash",
			RemoteADF: substantiveADF, RemoteHash: "newremote",
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: baselineRemote},
		}, resolveRemote, ActionPull, BlockNone},

		// Empty remoteHash baseline (Track 1 transition)
		{"baseline_empty_remote_hash_no_false_conflict", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: localContent, LocalHash: "newhash",
			RemoteHash: "anyhash",
			Baseline: &forest.NodeState{LocalHash: baselineLocal, RemoteHash: ""},
		}, pushOpts, ActionPush, BlockNone},

		// Direction resolution: "both" opts respects node sync mode
		{"both_opts_push_node_uses_push_rules", NodeReading{
			Node: node("KEY-1", "push"), LocalContent: emptyContent,
		}, bothOpts, ActionBlocked, BlockEmpty},

		{"both_opts_pull_node_remote_empty_skips", NodeReading{
			Node: node("KEY-1", "pull"), RemoteADF: emptyADF,
		}, bothOpts, ActionSkip, BlockNone},
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
