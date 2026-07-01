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
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/repo"
)

// ErrReadOnly is returned by a read-only strategy's Push (external stores).
var ErrReadOnly = errors.New("store is external (read-only) — folio never pushes it; contribute via its own PR flow")

// ErrNotImplemented is returned by kinds whose sync is not yet wired.
var ErrNotImplemented = errors.New("sync not yet implemented for this store kind")

// ErrDelegate is the directive a `code` push emits: folio verified a non-main
// branch, and the driving AGENT must run the Next skills to compose the commit/PR.
// Folio never invokes skills. It is a plain typed Go struct on purpose (design
// OQ3) — NOT modeled on dendrik's ResultEnvelope; typed-not-stringly is free and
// premature-proof.
type ErrDelegate struct {
	Dir    string   // repo/worktree the agent should act in
	Branch string   // the non-main branch commits land on
	Next   []string // skills to run, in order (e.g. ["/commit"])
}

func (e *ErrDelegate) Error() string {
	return fmt.Sprintf("delegate to agent: in %s on branch %q, run %s",
		e.Dir, e.Branch, strings.Join(e.Next, " then "))
}

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
	// Status is cheap, read-only, and NEVER mutates; it powers `fleet status`.
	// A per-repo failure is reported in StoreStatus.Err, never returned as a
	// hard error, so one broken repo never blocks the fleet view.
	Status(dir string, s config.Store) StoreStatus
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
		return codeStrategy{}
	case config.KindDot:
		return dotStrategy{}
	default:
		return externalStrategy{}
	}
}

// workRepoStatus reports a code/dot repo's status PER VCS: jj bookmark + @-dirty
// for jj/colocated repos (git HEAD is detached there, so git would mislabel),
// git branch + porcelain for git-only repos.
func workRepoStatus(dir string, s config.Store) StoreStatus {
	if repo.IsJJ(dir) {
		return jjWorkStatus(dir, s)
	}
	return gitStatus(dir, s)
}

// CanPush reports whether folio may push to this store. It is the single
// read-only refusal predicate, replacing the two open-coded store.IsExternal()
// guards (cmd_home.go push + helpers.go write-target resolution).
func CanPush(s config.Store) bool { return !For(s).ReadOnly() }

// kbStrategy is the folio KB mechanic, wrapping internal/repo verbatim.
type kbStrategy struct{}

func (kbStrategy) ReadOnly() bool                                { return false }
func (kbStrategy) Status(dir string, s config.Store) StoreStatus { return jjStatus(dir, s) }
func (kbStrategy) Pull(dir string, _ config.Store) error         { return repo.Pull(dir) }
func (kbStrategy) Push(dir string, _ config.Store, msg string, o PushOpts) (PushResult, error) {
	if len(o.Scoped) > 0 {
		return PushResult{}, repo.PushScoped(dir, msg, o.Scoped)
	}
	return PushResult{}, repo.Push(dir, msg)
}

// externalStrategy is a read-only KB: pullable (git pull) but never pushed.
type externalStrategy struct{}

func (externalStrategy) ReadOnly() bool                                { return true }
func (externalStrategy) Status(dir string, s config.Store) StoreStatus { return gitStatus(dir, s) }
func (externalStrategy) Pull(dir string, _ config.Store) error         { return repo.Pull(dir) }
func (externalStrategy) Push(_ string, _ config.Store, _ string, _ PushOpts) (PushResult, error) {
	return PushResult{}, ErrReadOnly
}

// codeStrategy hovers over a code repo: real read-only Status + Pull (git fetch
// only, never checkout/merge). Push is deferred to P3 (position a non-main branch
// + emit a delegate directive); until then it returns ErrNotImplemented.
type codeStrategy struct{}

func (codeStrategy) ReadOnly() bool                                { return false }
func (codeStrategy) Status(dir string, s config.Store) StoreStatus { return workRepoStatus(dir, s) }
func (codeStrategy) Pull(dir string, _ config.Store) error         { return gitFetchOnly(dir) }

// Push never touches a shared main. It reads the current branch (per-VCS) and
// REFUSES on the default branch or on ambiguity (detached/empty/unknown) — "never
// main" is enforced defensively, not just by convention. On a safe non-main
// branch it composes no commit/PR itself; it returns an ErrDelegate directing the
// driving agent to run /commit (which owns commit format + stacks/PRs).
func (codeStrategy) Push(dir string, s config.Store, _ string, _ PushOpts) (PushResult, error) {
	def := s.DefaultBranch
	if def == "" {
		def = "main"
	}
	branch, err := currentBranch(dir)
	if err != nil {
		return PushResult{}, fmt.Errorf("refusing code push: %w", err)
	}
	if branch == def || branch == "main" || branch == "master" {
		return PushResult{}, fmt.Errorf("refusing to push code repo on shared branch %q — switch to a feature branch first", branch)
	}
	return PushResult{}, &ErrDelegate{Dir: dir, Branch: branch, Next: []string{"/commit"}}
}

// dotStrategy manages a dotfiles repo. Status is per-VCS (the dotfiles repo is
// jj-colocated, so jj gives the real branch+dirty; `dot status` is a verbose
// report, not a status cell). Pull shells `dot pull`. Push is deferred to P4.
type dotStrategy struct{}

func (dotStrategy) ReadOnly() bool                                { return false }
func (dotStrategy) Status(dir string, s config.Store) StoreStatus { return workRepoStatus(dir, s) }
func (dotStrategy) Pull(dir string, _ config.Store) error         { return dotRun(dir, "pull") }
func (dotStrategy) Push(_ string, _ config.Store, _ string, _ PushOpts) (PushResult, error) {
	return PushResult{}, fmt.Errorf("%w: dotfiles push lands in P4", ErrNotImplemented)
}
