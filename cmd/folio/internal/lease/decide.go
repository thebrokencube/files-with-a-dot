package lease

import (
	"fmt"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

// Status states, reported by `folio lease status` as data.state.
const (
	StateFree     = "free"
	StateStarting = "starting"
	StateRunning  = "running"
	StateHeld     = "held"
)

const (
	outcomeStarted      = "started"
	outcomeWouldStart   = "would-start"
	outcomeRunning      = "running"
	outcomeHeld         = "held"
	outcomeRefused      = "refused"
	outcomeFailed       = "failed"
	outcomeReleased     = "released"
	outcomeWouldRelease = "would-release"
)

// Actions a decision asks the command to take; reported as outcomes above.
const (
	actionStart  = "start"
	actionStop   = "stop"
	actionRemove = "remove"
	actionNone   = "none"
)

// Owner is one process listening on a declared port.
type Owner struct {
	PID  int
	PGID int
}

// Facts is everything decide needs, gathered before the decision so the
// decision itself is pure.
type Facts struct {
	Store      string
	Path       string
	Declared   bool
	Conflict   *PortConflict
	Serve      Serve
	Lease      *Record
	Alive      bool
	Owners     map[int][]Owner
	InCheckout bool
	Present    []string
	Missing    []string
	Digest     string
}

type verdict struct {
	Outcome string
	Code    int
	Message string
}

type standing struct {
	State string
	PID   int
	Port  int
}

func (f Facts) live() bool { return f.Lease != nil && f.Alive }

// stateOf classifies the declared ports against the lease pgid. Any listener
// outside a live lease pgid makes the slot held, whatever else is true.
func stateOf(f Facts) standing {
	bound := 0
	for _, p := range f.Serve.Processes {
		owners := f.Owners[p.Port]
		for _, o := range owners {
			if !f.live() || o.PGID != f.Lease.PGID {
				return standing{State: StateHeld, PID: o.PID, Port: p.Port}
			}
		}
		if len(owners) > 0 {
			bound++
		}
	}
	switch {
	case !f.live():
		return standing{State: StateFree}
	case bound == len(f.Serve.Processes):
		return standing{State: StateRunning}
	default:
		return standing{State: StateStarting}
	}
}

// ownerInLease reports whether a declared port's owner is inside the live
// lease pgid — the only condition under which folio may signal that pgid.
func ownerInLease(f Facts) bool {
	if !f.live() {
		return false
	}
	for _, p := range f.Serve.Processes {
		for _, o := range f.Owners[p.Port] {
			if o.PGID == f.Lease.PGID {
				return true
			}
		}
	}
	return false
}

func decide(f Facts) verdict {
	if !f.Declared {
		return verdict{outcomeRefused, dendrik.ExitUserError, fmt.Sprintf("store %q has no serve block; folio lease does not apply", f.Store)}
	}
	if c := f.Conflict; c != nil {
		return verdict{outcomeRefused, dendrik.ExitUserError, fmt.Sprintf("port %d is declared by both %q and %q; serve ports must be unique across stores", c.Port, f.Store, c.Other)}
	}
	st := stateOf(f)
	switch st.State {
	case StateHeld:
		return verdict{outcomeHeld, dendrik.ExitConflict, fmt.Sprintf("port %d is held by unleased owner pid %d; folio never signals it", st.Port, st.PID)}
	case StateStarting:
		return verdict{outcomeHeld, dendrik.ExitConflict, fmt.Sprintf("%s is starting (pgid %d, since %s) — do not relaunch; poll `folio lease status %s`", f.Store, f.Lease.PGID, f.Lease.Started, f.Store)}
	case StateRunning:
		holder := fmt.Sprintf("%s is running at %s from %s (pgid %d, since %s)", f.Store, f.Serve.where(), f.Lease.Checkout, f.Lease.PGID, f.Lease.Started)
		if !f.InCheckout {
			msg := holder + "; only " + f.Path + " may launch it"
			if f.Serve.Alternative != "" {
				msg += "; alternative: " + f.Serve.Alternative
			}
			return verdict{outcomeHeld, dendrik.ExitConflict, msg}
		}
		if len(f.Missing) > 0 {
			return verdict{outcomeRefused, dendrik.ExitUserError, fmt.Sprintf("%s; this shell is missing required env %s, so it cannot be compared — load the store's env and rerun, or keep using the running server", holder, strings.Join(f.Missing, ", "))}
		}
		if f.Digest != f.Lease.EnvDigest {
			return verdict{outcomeHeld, dendrik.ExitConflict, fmt.Sprintf("%s, but env differs from this shell's require_env; run `folio lease release %s` then `folio lease run %s`", holder, f.Store, f.Store)}
		}
		return verdict{outcomeRunning, dendrik.ExitOK, fmt.Sprintf("%s already running at %s (pgid %d, since %s)", f.Store, f.Serve.where(), f.Lease.PGID, f.Lease.Started)}
	}
	if !f.InCheckout {
		return verdict{outcomeRefused, dendrik.ExitUserError, fmt.Sprintf("launch %s from its main checkout %s", f.Store, f.Path)}
	}
	if len(f.Missing) > 0 {
		return verdict{outcomeRefused, dendrik.ExitUserError, fmt.Sprintf("missing required env for %s: %s", f.Store, strings.Join(f.Missing, ", "))}
	}
	return verdict{actionStart, dendrik.ExitOK, ""}
}

// decideRelease keeps a lease whose group is alive but owns no declared port
// yet, so a booting server never becomes an unleased orphan.
func decideRelease(f Facts) verdict {
	switch {
	case !f.Declared:
		return decide(f)
	case !f.InCheckout:
		return verdict{outcomeRefused, dendrik.ExitUserError, fmt.Sprintf("release %s only from its main checkout %s", f.Store, f.Path)}
	case f.Lease == nil:
		return verdict{actionNone, dendrik.ExitOK, fmt.Sprintf("%s has no lease; nothing to release", f.Store)}
	case ownerInLease(f):
		return verdict{actionStop, dendrik.ExitOK, ""}
	case f.live():
		return verdict{outcomeHeld, dendrik.ExitConflict, fmt.Sprintf("%s is starting (pgid %d, since %s) — retry once it is ready; to stop it by hand, check `pgrep -l -g %d`; if those are %s's processes, `kill -TERM -%d` (log %s)", f.Store, f.Lease.PGID, f.Lease.Started, f.Lease.PGID, f.Store, f.Lease.PGID, f.Lease.Log)}
	default:
		return verdict{actionRemove, dendrik.ExitOK, ""}
	}
}
