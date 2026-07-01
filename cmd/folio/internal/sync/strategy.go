// Package sync selects a per-kind synchronization strategy for a store.
//
// It is an EXTRACTION, not an invention: folio's KB push/pull mechanic already
// lived in internal/repo (jjPush's rebase-@→main+bookmark dance), frozen to one
// kind. This package lifts that behind a SyncStrategy interface keyed by
// config.Store.Kind so the fleet can carry code/dotfiles repos alongside KB
// stores, each synced the right way.
//
// P0 wires only the KB (folio) and external kinds — kbStrategy wraps repo.Push/
// Pull VERBATIM, externalStrategy refuses push. code/dot are stubs until P3/P4.
// The whole point of P0 is ZERO observable behavior change on existing flows.
package sync

import (
	"errors"
	"fmt"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/repo"
)

// ErrReadOnly is returned by a read-only strategy's Push (external stores).
var ErrReadOnly = errors.New("store is external (read-only) — folio never pushes it; contribute via its own PR flow")

// ErrNotImplemented is returned by kinds whose sync is not yet wired (code/dot in P0).
var ErrNotImplemented = errors.New("sync not yet implemented for this store kind")

// PushOpts carries push variants. Scoped, when non-empty, limits a KB push to
// those tree paths (the -f scoped-push case); empty means whole-tree.
type PushOpts struct {
	Scoped []string
}

// PushResult is the (currently empty) structured result of a push. It exists so
// the interface can grow — e.g. code push will return a delegate directive here.
type PushResult struct{}

// SyncStrategy is the per-kind sync mechanic. Methods take the already-resolved
// repo root `dir` (NOT s.Path) because the legacy single-home path resolves to
// the umbrella with a zero-value Store; dir is authoritative in every case.
type SyncStrategy interface {
	// ReadOnly reports whether folio must never push to this kind.
	ReadOnly() bool
	// Pull refreshes the repo (read-only, safe for every kind).
	Pull(dir string, s config.Store) error
	// Push commits+pushes; read-only kinds return ErrReadOnly.
	Push(dir string, s config.Store, msg string, o PushOpts) (PushResult, error)
}

// For selects the strategy for a store's kind.
//
// The empty-kind case maps to the KB strategy on purpose: a zero-value
// config.Store (Kind=="") only ever comes from the legacy single-home path
// (no stores.yml), which has always been a folio KB — so mapping it to
// kbStrategy preserves pre-registry behavior exactly. A *registered* store that
// omits kind never reaches here as "": parseRegistry already defaulted it to
// external. Any genuinely unknown kind falls through to external (read-only
// safe), never the writable KB push.
func For(s config.Store) SyncStrategy {
	switch s.Kind {
	case config.KindFolio, "":
		return kbStrategy{}
	case config.KindExternal:
		return externalStrategy{}
	case config.KindCode:
		return stubStrategy{kind: config.KindCode}
	case config.KindDot:
		return stubStrategy{kind: config.KindDot}
	default:
		return externalStrategy{}
	}
}

// CanPush reports whether folio may push to this store. It is the single
// read-only refusal predicate, replacing the two open-coded store.IsExternal()
// guards (cmd_home.go push + helpers.go write-target resolution).
func CanPush(s config.Store) bool { return !For(s).ReadOnly() }

// kbStrategy is the folio KB mechanic, wrapping internal/repo verbatim.
type kbStrategy struct{}

func (kbStrategy) ReadOnly() bool                        { return false }
func (kbStrategy) Pull(dir string, _ config.Store) error { return repo.Pull(dir) }
func (kbStrategy) Push(dir string, _ config.Store, msg string, o PushOpts) (PushResult, error) {
	if len(o.Scoped) > 0 {
		return PushResult{}, repo.PushScoped(dir, msg, o.Scoped)
	}
	return PushResult{}, repo.Push(dir, msg)
}

// externalStrategy is a read-only KB: pullable (git pull) but never pushed.
type externalStrategy struct{}

func (externalStrategy) ReadOnly() bool                        { return true }
func (externalStrategy) Pull(dir string, _ config.Store) error { return repo.Pull(dir) }
func (externalStrategy) Push(_ string, _ config.Store, _ string, _ PushOpts) (PushResult, error) {
	return PushResult{}, ErrReadOnly
}

// stubStrategy is a placeholder for kinds wired in later phases (code=P3, dot=P4).
// It is NOT read-only (those kinds do push, just via their own mechanic), so it
// does not silently masquerade as external; it simply errors until implemented.
type stubStrategy struct{ kind string }

func (stubStrategy) ReadOnly() bool { return false }
func (s stubStrategy) Pull(_ string, _ config.Store) error {
	return fmt.Errorf("%w: %s", ErrNotImplemented, s.kind)
}
func (s stubStrategy) Push(_ string, _ config.Store, _ string, _ PushOpts) (PushResult, error) {
	return PushResult{}, fmt.Errorf("%w: %s", ErrNotImplemented, s.kind)
}
