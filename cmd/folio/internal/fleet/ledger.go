// Package fleet manages code/dotfiles work areas (isolated checkouts) that folio
// hovers over: a canonical off-repo placement, an advisory JSONL ledger, and a
// tier-correct reaper that never destroys a git worktree with os.RemoveAll.
//
// The ledger is ADVISORY. The VCS's own list (git worktree list / jj workspace
// list) is TRUTH; reconciliation (List/Reap) always intersects the two so a
// hand-made or severed worktree is surfaced, never silently trusted or reaped.
package fleet

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
)

// WorkArea is one ledger row: an isolated checkout of a store on a branch.
type WorkArea struct {
	Store   string `json:"store"`
	Kind    string `json:"kind"`
	Tier    string `json:"tier"`
	Dir     string `json:"dir"`
	Root    string `json:"root,omitempty"` // parent repo path — lets the reaper prune from the source repo
	Branch  string `json:"branch"`
	Base    string `json:"base,omitempty"`
	Session string `json:"session,omitempty"`
	Created string `json:"created,omitempty"`
	// Lease is reserved for the deferred cooperative-lease tier (design OQ6);
	// unused in P2. Kept so the ledger schema does not churn when it lands.
	Lease string `json:"lease,omitempty"`
}

// LedgerPath is the gitignored advisory ledger under the umbrella.
func LedgerPath(umbrella string) string {
	return filepath.Join(umbrella, ".fleet", "workareas.jsonl")
}

// WorktreesRoot is the canonical off-repo home for code/dot work areas.
func WorktreesRoot(umbrella string) string {
	return filepath.Join(umbrella, ".worktrees")
}

// ReadLedger parses the JSONL ledger. A missing file is not an error (empty).
// A malformed line is skipped rather than failing the whole read — the ledger
// is advisory and must never block a status/reap.
func ReadLedger(umbrella string) ([]WorkArea, error) {
	f, err := os.Open(LedgerPath(umbrella))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var out []WorkArea
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var wa WorkArea
		if err := json.Unmarshal(line, &wa); err != nil {
			continue // skip malformed advisory row
		}
		out = append(out, wa)
	}
	return out, sc.Err()
}

// AppendLedger appends one work area, creating the .fleet dir if needed.
func AppendLedger(umbrella string, wa WorkArea) error {
	if err := os.MkdirAll(filepath.Dir(LedgerPath(umbrella)), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(LedgerPath(umbrella), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(wa)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(b, '\n')); err != nil {
		return err
	}
	return nil
}

// WriteLedger rewrites the whole ledger (used by reap to prune dead rows).
func WriteLedger(umbrella string, areas []WorkArea) error {
	if err := os.MkdirAll(filepath.Dir(LedgerPath(umbrella)), 0755); err != nil {
		return err
	}
	tmp := LedgerPath(umbrella) + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	for _, wa := range areas {
		b, err := json.Marshal(wa)
		if err != nil {
			f.Close()
			return err
		}
		if _, err := w.Write(append(b, '\n')); err != nil {
			f.Close()
			return err
		}
	}
	if err := w.Flush(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, LedgerPath(umbrella))
}
