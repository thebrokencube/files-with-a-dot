package engine

import (
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
)

// Plan is a pure function (zero I/O). It takes readings and options,
// returns one Action per reading. Output length always equals input length.
func Plan(readings []NodeReading, opts PlanOpts) []Action {
	actions := make([]Action, len(readings))
	for i, r := range readings {
		actions[i] = planNode(r, opts)
	}
	return actions
}

func planNode(r NodeReading, opts PlanOpts) Action {
	// Rule 1: TBD key — skip
	if forest.IsTBD(r.Node.Key) {
		return Action{Node: r.Node, Kind: ActionSkip, Reason: "TBD key"}
	}

	direction := resolveDirection(r.Node, opts)

	if direction == "push" || direction == "both" {
		// Rule 2: Emptiness — Tier 3 block (no override ever)
		if !IsSubstantiveLocal(r.LocalContent) {
			if direction == "push" {
				return Action{Node: r.Node, Kind: ActionBlocked,
					Block: BlockEmpty, Reason: "empty local content"}
			}
			// direction == "both" with empty local: fall through to pull check below.
		}

		// Rule 3: Remote unreachable — Tier 3 block (only if push-eligible)
		if r.RemoteErr != nil && (direction == "push" || IsSubstantiveLocal(r.LocalContent)) {
			return Action{Node: r.Node, Kind: ActionBlocked,
				Block: BlockRemoteUnknown, Reason: "cannot reach Jira"}
		}
	}

	if direction == "pull" || direction == "both" {
		// Symmetric: pull-direction remote unreachable -> skip (nothing to pull)
		if r.RemoteErr != nil {
			if direction == "pull" {
				return Action{Node: r.Node, Kind: ActionSkip,
					Reason: "remote unreachable"}
			}
			// direction == "both" with empty local + remote unreachable: block
			if !IsSubstantiveLocal(r.LocalContent) {
				return Action{Node: r.Node, Kind: ActionBlocked,
					Block: BlockEmpty, Reason: "empty local content, remote unreachable"}
			}
		}

		// Symmetric emptiness for pull: skip (nothing to pull)
		if !IsSubstantiveADF(r.RemoteADF) {
			if direction == "pull" {
				return Action{Node: r.Node, Kind: ActionSkip,
					Reason: "remote is empty"}
			}
			// direction == "both" with empty local + empty remote: block on empty
			if !IsSubstantiveLocal(r.LocalContent) {
				return Action{Node: r.Node, Kind: ActionBlocked,
					Block: BlockEmpty, Reason: "empty local content"}
			}
			// direction == "both" with substantive local + empty remote: fall through to push
		}
	}

	// direction == "both" with empty local + substantive remote: pull instead of blocking
	if direction == "both" && !IsSubstantiveLocal(r.LocalContent) && IsSubstantiveADF(r.RemoteADF) {
		if r.Baseline == nil {
			return Action{Node: r.Node, Kind: ActionPull, Reason: "first sync, pulling remote content",
				RemoteADF: r.RemoteADF, RemoteHash: r.RemoteHash}
		}
		return Action{Node: r.Node, Kind: ActionPull, Reason: "local empty, pulling remote content",
			RemoteADF: r.RemoteADF, RemoteHash: r.RemoteHash}
	}

	// Rule 2 nil-guard: baseline == nil means first sync
	if r.Baseline == nil {
		return planFirstSync(r, direction)
	}

	return planWithBaseline(r, direction, opts)
}

// planFirstSync handles nodes that have never been synced before.
func planFirstSync(r NodeReading, direction string) Action {
	switch direction {
	case "push":
		// First push: if remote has content, block (Tier 2)
		if IsSubstantiveADF(r.RemoteADF) {
			return Action{Node: r.Node, Kind: ActionBlocked,
				Block: BlockFirstPush, Reason: "first sync — remote has content",
				LocalContent: r.LocalContent, LocalHash: r.LocalHash,
				RemoteADF: r.RemoteADF, RemoteHash: r.RemoteHash}
		}
		// Remote empty: safe to push
		return Action{Node: r.Node, Kind: ActionPush,
			Reason: "first sync, remote empty",
			LocalContent: r.LocalContent, LocalHash: r.LocalHash}

	case "pull":
		// First pull: if local has content, block (Tier 2)
		if IsSubstantiveLocal(r.LocalContent) {
			return Action{Node: r.Node, Kind: ActionBlocked,
				Block: BlockFirstPull, Reason: "first sync — local has content",
				LocalContent: r.LocalContent, LocalHash: r.LocalHash,
				RemoteADF: r.RemoteADF, RemoteHash: r.RemoteHash}
		}
		// Local empty: safe to pull
		return Action{Node: r.Node, Kind: ActionPull,
			Reason: "first sync, local empty",
			RemoteADF: r.RemoteADF, RemoteHash: r.RemoteHash}

	case "both":
		// Both direction, first sync:
		// If remote has content and local has content: block first-push
		if IsSubstantiveADF(r.RemoteADF) {
			return Action{Node: r.Node, Kind: ActionBlocked,
				Block: BlockFirstPush, Reason: "first sync — remote has content",
				LocalContent: r.LocalContent, LocalHash: r.LocalHash,
				RemoteADF: r.RemoteADF, RemoteHash: r.RemoteHash}
		}
		// Remote empty, local has content: push
		return Action{Node: r.Node, Kind: ActionPush,
			Reason: "first sync, remote empty",
			LocalContent: r.LocalContent, LocalHash: r.LocalHash}
	}

	return Action{Node: r.Node, Kind: ActionSkip, Reason: "unknown direction"}
}

// planWithBaseline handles nodes with an existing sync baseline.
func planWithBaseline(r NodeReading, direction string, opts PlanOpts) Action {
	localChanged := r.LocalHash != r.Baseline.LocalHash

	// Empty RemoteHash from Track 1 transition: treat as "no remote baseline"
	remoteChanged := false
	if r.Baseline.RemoteHash != "" {
		remoteChanged = r.RemoteHash != r.Baseline.RemoteHash
	}

	switch direction {
	case "push":
		if remoteChanged {
			return Action{Node: r.Node, Kind: ActionBlocked,
				Block: BlockOverwrite, Reason: "remote changed since last sync",
				LocalContent: r.LocalContent, LocalHash: r.LocalHash,
				RemoteADF: r.RemoteADF, RemoteHash: r.RemoteHash}
		}
		if localChanged {
			return Action{Node: r.Node, Kind: ActionPush,
				Reason: "local changed",
				LocalContent: r.LocalContent, LocalHash: r.LocalHash}
		}
		return Action{Node: r.Node, Kind: ActionSkip, Reason: "no changes"}

	case "pull":
		if localChanged {
			return Action{Node: r.Node, Kind: ActionBlocked,
				Block: BlockOverwrite, Reason: "local changed since last sync",
				LocalContent: r.LocalContent, LocalHash: r.LocalHash,
				RemoteADF: r.RemoteADF, RemoteHash: r.RemoteHash}
		}
		if remoteChanged {
			return Action{Node: r.Node, Kind: ActionPull,
				Reason: "remote changed",
				RemoteADF: r.RemoteADF, RemoteHash: r.RemoteHash}
		}
		return Action{Node: r.Node, Kind: ActionSkip, Reason: "no changes"}

	case "both":
		if localChanged && remoteChanged {
			// Conflict — check --resolve
			if opts.Resolve == "local" {
				return Action{Node: r.Node, Kind: ActionPush,
					Reason: "conflict resolved: local wins",
					LocalContent: r.LocalContent, LocalHash: r.LocalHash}
			}
			if opts.Resolve == "remote" {
				return Action{Node: r.Node, Kind: ActionPull,
					Reason: "conflict resolved: remote wins",
					RemoteADF: r.RemoteADF, RemoteHash: r.RemoteHash}
			}
			return Action{Node: r.Node, Kind: ActionBlocked,
				Block: BlockConflict, Reason: "both sides changed",
				LocalContent: r.LocalContent, LocalHash: r.LocalHash,
				RemoteADF: r.RemoteADF, RemoteHash: r.RemoteHash}
		}
		if localChanged {
			return Action{Node: r.Node, Kind: ActionPush,
				Reason: "local changed",
				LocalContent: r.LocalContent, LocalHash: r.LocalHash}
		}
		if remoteChanged {
			return Action{Node: r.Node, Kind: ActionPull,
				Reason: "remote changed",
				RemoteADF: r.RemoteADF, RemoteHash: r.RemoteHash}
		}
		return Action{Node: r.Node, Kind: ActionSkip, Reason: "no changes"}
	}

	return Action{Node: r.Node, Kind: ActionSkip, Reason: "unknown direction"}
}
