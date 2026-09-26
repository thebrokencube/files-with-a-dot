package lease

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

const (
	readyPoll   = 250 * time.Millisecond
	releaseWait = 10 * time.Second
	tailLines   = 20
)

// Record is <umbrella>/.fleet/leases/<store>.json. It names the env keys
// present at launch and their digest, never their values.
type Record struct {
	Store      string         `json:"store"`
	Checkout   string         `json:"checkout"`
	PGID       int            `json:"pgid"`
	Ports      map[string]int `json:"ports"`
	EnvPresent []string       `json:"env_present"`
	EnvDigest  string         `json:"env_digest"`
	Started    string         `json:"started"`
	Log        string         `json:"log"`
}

// Report is the data payload of every lease command.
type Report struct {
	Store    string `json:"store"`
	Declared bool   `json:"declared"`
	State    string `json:"state,omitempty"`
	Outcome  string `json:"outcome,omitempty"`
	Message  string `json:"message,omitempty"`
	URL      string `json:"url,omitempty"`
	Checkout string `json:"checkout,omitempty"`
	PGID     int    `json:"pgid,omitempty"`
	Since    string `json:"since,omitempty"`
	Env      string `json:"env,omitempty"`
	Log      string `json:"log,omitempty"`
	LogTail  string `json:"log_tail,omitempty"`
}

// Target is the store a lease command acts on and the caller's working directory.
type Target struct {
	Umbrella string
	Registry *config.Registry
	Store    config.Store
	Cwd      string
}

func (t Target) file(ext string) string {
	return filepath.Join(t.Umbrella, ".fleet", "leases", t.Store.Name+ext)
}

// Run launches the declared server when decide says the slot is free, and
// returns once every declared port is owned inside the new process group.
// dryRun reports the same decision and exit code without the lock, spawn, log,
// or lease write.
func Run(t Target, timeout time.Duration, dryRun bool) (Report, int, error) {
	s, declared, err := Decode(t.Store)
	if err != nil {
		return Report{}, dendrik.ExitUserError, err
	}
	if !declared {
		f := Facts{Store: t.Store.Name}
		v := decide(f)
		return report(f, v), v.Code, nil
	}
	unlock := func() {}
	if !dryRun {
		if err := os.MkdirAll(filepath.Dir(t.file(".json")), 0o755); err != nil {
			return Report{}, dendrik.ExitExternalErr, err
		}
		if unlock, err = lock(t.file(".lock")); err != nil {
			return Report{}, dendrik.ExitExternalErr, err
		}
	}
	f, err := gather(t, s)
	if err != nil {
		unlock()
		return Report{}, dendrik.ExitExternalErr, err
	}
	v := decide(f)
	if v.Outcome != actionStart {
		unlock()
		return report(f, v), v.Code, nil
	}
	if dryRun {
		r := report(f, v)
		r.Outcome = outcomeWouldStart
		r.Message = fmt.Sprintf("would start %s from %s: sh -c %q", f.Store, f.Path, script(s))
		return r, dendrik.ExitOK, nil
	}
	cmd, rec, err := spawn(t, f)
	unlock()
	if err != nil {
		return Report{}, dendrik.ExitExternalErr, err
	}
	return awaitReady(t, s, cmd, rec, timeout)
}

// Status reports the slot without writing anything.
func Status(t Target) (Report, int, error) {
	s, declared, err := Decode(t.Store)
	if err != nil {
		return Report{}, dendrik.ExitUserError, err
	}
	if !declared {
		return Report{Store: t.Store.Name, Message: fmt.Sprintf("store %q declares no serve block; folio lease does not apply", t.Store.Name)}, dendrik.ExitOK, nil
	}
	f, err := gather(t, s)
	if err != nil {
		return Report{}, dendrik.ExitExternalErr, err
	}
	if f.Conflict != nil {
		v := decide(f)
		return report(f, v), v.Code, nil
	}
	st := stateOf(f)
	r := report(f, verdict{})
	r.State = st.State
	r.Env = "n/a"
	switch {
	case f.live() && len(f.Missing) > 0:
		r.Env = "missing"
	case f.live() && f.Digest == f.Lease.EnvDigest:
		r.Env = "match"
	case f.live():
		r.Env = "differs"
	}
	switch st.State {
	case StateFree:
		r.Message = fmt.Sprintf("%s is free; launch it with `folio lease run %s` from %s", f.Store, f.Store, f.Path)
	case StateHeld:
		r.Message = fmt.Sprintf("port %d is held by unleased owner pid %d", st.Port, st.PID)
	case StateStarting:
		r.Message = fmt.Sprintf("%s is starting (pgid %d, since %s)", f.Store, f.Lease.PGID, f.Lease.Started)
	case StateRunning:
		r.Message = fmt.Sprintf("%s is running at %s from %s (pgid %d, since %s)", f.Store, f.Serve.where(), f.Lease.Checkout, f.Lease.PGID, f.Lease.Started)
	}
	return r, dendrik.ExitOK, nil
}

// Release stops the leased server — only while a declared port's owner is in
// the lease pgid — and deletes the lease. dryRun reports the same decision and
// exit code without the lock, signal, or lease delete.
func Release(t Target, dryRun bool) (Report, int, error) {
	s, declared, err := Decode(t.Store)
	if err != nil {
		return Report{}, dendrik.ExitUserError, err
	}
	if !declared {
		f := Facts{Store: t.Store.Name}
		v := decideRelease(f)
		return report(f, v), v.Code, nil
	}
	path := t.file(".json")
	if !dryRun {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return Report{}, dendrik.ExitExternalErr, err
		}
		unlock, err := lock(t.file(".lock"))
		if err != nil {
			return Report{}, dendrik.ExitExternalErr, err
		}
		defer unlock()
	}
	f, err := gather(t, s)
	if err != nil {
		return Report{}, dendrik.ExitExternalErr, err
	}
	v := decideRelease(f)
	r := report(f, v)
	if v.Code != dendrik.ExitOK {
		return r, v.Code, nil
	}
	r.Outcome = outcomeReleased
	if dryRun {
		r.Outcome = outcomeWouldRelease
	}
	if v.Outcome == actionNone {
		return r, dendrik.ExitOK, nil
	}
	r.Checkout, r.PGID, r.Since, r.Log = f.Lease.Checkout, f.Lease.PGID, f.Lease.Started, f.Lease.Log
	if v.Outcome == actionRemove {
		r.Message = fmt.Sprintf("removed %s lease; pgid %d owns no declared port, nothing signalled", f.Store, f.Lease.PGID)
		if dryRun {
			r.Message = fmt.Sprintf("would remove %s lease; pgid %d owns no declared port, nothing would be signalled", f.Store, f.Lease.PGID)
			return r, dendrik.ExitOK, nil
		}
		if err := os.Remove(path); err != nil {
			return Report{}, dendrik.ExitExternalErr, err
		}
		return r, dendrik.ExitOK, nil
	}
	if dryRun {
		r.Message = fmt.Sprintf("would stop %s (TERM pgid %d) and remove its lease", f.Store, f.Lease.PGID)
		return r, dendrik.ExitOK, nil
	}
	if err := signalGroup(f.Lease.PGID, syscall.SIGTERM); err != nil && !errors.Is(err, syscall.ESRCH) {
		return Report{}, dendrik.ExitExternalErr, err
	}
	freed, err := awaitFree(s, *f.Lease)
	if err != nil {
		return Report{}, dendrik.ExitExternalErr, err
	}
	if !freed {
		r.Outcome = outcomeFailed
		r.Message = fmt.Sprintf("sent TERM to %s pgid %d, but it still owns a declared port after %s; lease kept so a retry can signal again; log %s", f.Store, f.Lease.PGID, releaseWait, f.Lease.Log)
		return r, dendrik.ExitExternalErr, nil
	}
	if err := os.Remove(path); err != nil {
		return Report{}, dendrik.ExitExternalErr, err
	}
	r.Message = fmt.Sprintf("stopped %s (TERM pgid %d) and removed its lease", f.Store, f.Lease.PGID)
	return r, dendrik.ExitOK, nil
}

func gather(t Target, s Serve) (Facts, error) {
	f := Facts{Store: t.Store.Name, Path: t.Store.Path, Declared: true, Serve: s}
	f.Conflict = findConflict(t.Registry, t.Store.Name, s)
	rec, err := readRecord(t.file(".json"))
	if err != nil {
		return f, err
	}
	f.Lease = rec
	f.Alive = rec != nil && groupAlive(rec.PGID)
	if f.Owners, err = portOwners(s); err != nil {
		return f, err
	}
	f.InCheckout = inCheckout(t.Cwd, t.Store.Path)
	f.Present, f.Missing, f.Digest = envFacts(s.RequireEnv)
	return f, nil
}

func report(f Facts, v verdict) Report {
	r := Report{Store: f.Store, Declared: f.Declared, Outcome: v.Outcome, Message: v.Message}
	if !f.Declared {
		return r
	}
	r.URL, r.Checkout = f.Serve.URL, f.Path
	if f.live() {
		r.Checkout, r.PGID, r.Since, r.Log = f.Lease.Checkout, f.Lease.PGID, f.Lease.Started, f.Lease.Log
	}
	return r
}

// script runs each declared process in its own backgrounded subshell, so one
// run string's `cd`, comment, or trailing `&` cannot reach the others, then
// waits for all of them.
func script(s Serve) string {
	var b strings.Builder
	for _, p := range s.Processes {
		b.WriteString("(" + p.Run + "\n) &\n")
	}
	b.WriteString("wait")
	return b.String()
}

// spawn starts every declared process under one `sh` in a new session, so the
// shell's pid is the pgid and the caller's own group is never part of it.
func spawn(t Target, f Facts) (*exec.Cmd, Record, error) {
	logPath := t.file(".log")
	logf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return nil, Record{}, err
	}
	defer logf.Close()
	cmd := exec.Command("sh", "-c", script(f.Serve))
	cmd.Dir = t.Store.Path
	cmd.Stdout, cmd.Stderr = logf, logf
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return nil, Record{}, fmt.Errorf("launching %s: %w", t.Store.Name, err)
	}
	rec := Record{
		Store:      t.Store.Name,
		Checkout:   t.Store.Path,
		PGID:       cmd.Process.Pid,
		Ports:      map[string]int{},
		EnvPresent: f.Present,
		EnvDigest:  f.Digest,
		Started:    time.Now().UTC().Format(time.RFC3339),
		Log:        logPath,
	}
	if err := writeRecord(t.file(".json"), rec); err != nil {
		reaped := make(chan struct{})
		go func() {
			_ = cmd.Wait()
			close(reaped)
		}()
		stopGroup(rec.PGID)
		<-reaped
		return nil, Record{}, err
	}
	return cmd, rec, nil
}

func awaitReady(t Target, s Serve, cmd *exec.Cmd, rec Record, timeout time.Duration) (Report, int, error) {
	exited := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(exited)
	}()
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()
	tick := time.NewTicker(readyPoll)
	defer tick.Stop()
	for {
		select {
		case <-exited:
			return failed(t, rec, fmt.Sprintf("%s exited (%s) before every declared port was listening", t.Store.Name, cmd.ProcessState))
		case <-deadline.C:
			return failed(t, rec, fmt.Sprintf("%s did not listen on every declared port within %s", t.Store.Name, timeout))
		case <-tick.C:
			owners, err := portOwners(s)
			if err != nil {
				return failed(t, rec, err.Error())
			}
			f := Facts{Serve: s, Lease: &rec, Alive: true, Owners: owners}
			st := stateOf(f)
			if st.State == StateHeld {
				return failed(t, rec, fmt.Sprintf("port %d owned by pid %d outside pgid %d; declared processes must stay in the launched process group", st.Port, st.PID, rec.PGID))
			}
			if st.State != StateRunning {
				continue
			}
			for _, p := range s.Processes {
				rec.Ports[strconv.Itoa(p.Port)] = owners[p.Port][0].PID
			}
			ran, err := t.ifLeased(rec.PGID, func(path string) error { return writeRecord(path, rec) })
			if err != nil {
				return Report{}, dendrik.ExitExternalErr, err
			}
			if !ran {
				return failed(t, rec, fmt.Sprintf("%s was released during startup; stopped pgid %d", t.Store.Name, rec.PGID))
			}
			return Report{
				Store: t.Store.Name, Declared: true, Outcome: outcomeStarted,
				Message: fmt.Sprintf("started %s at %s (pgid %d); log %s", t.Store.Name, s.where(), rec.PGID, rec.Log),
				URL:     s.URL, Checkout: rec.Checkout, PGID: rec.PGID, Since: rec.Started, Log: rec.Log,
			}, dendrik.ExitOK, nil
		}
	}
}

// failed stops the group this run launched and removes its lease, if the
// lease still names it.
func failed(t Target, rec Record, why string) (Report, int, error) {
	stopGroup(rec.PGID)
	if _, err := t.ifLeased(rec.PGID, os.Remove); err != nil {
		return Report{}, dendrik.ExitExternalErr, err
	}
	return Report{
		Store: t.Store.Name, Declared: true, Outcome: outcomeFailed, Message: why,
		Checkout: rec.Checkout, Log: rec.Log, LogTail: logTail(rec.Log),
	}, dendrik.ExitExternalErr, nil
}

// stopGroup TERMs pgid, waits up to releaseWait for it to empty, then KILLs
// whatever remains so a failed launch leaves no orphans.
func stopGroup(pgid int) {
	if signalGroup(pgid, syscall.SIGTERM) != nil {
		return
	}
	deadline := time.Now().Add(releaseWait)
	for groupAlive(pgid) {
		if time.Now().After(deadline) {
			_ = signalGroup(pgid, syscall.SIGKILL)
			return
		}
		time.Sleep(readyPoll)
	}
}

// awaitFree waits up to releaseWait for no declared port to be owned inside the lease pgid.
func awaitFree(s Serve, rec Record) (bool, error) {
	deadline := time.Now().Add(releaseWait)
	for {
		owners, err := portOwners(s)
		if err != nil {
			return false, err
		}
		if !ownerInLease(Facts{Serve: s, Lease: &rec, Alive: true, Owners: owners}) {
			return true, nil
		}
		if time.Now().After(deadline) {
			return false, nil
		}
		time.Sleep(readyPoll)
	}
}

// ifLeased runs fn on the lease path under the lock, only while the lease
// still names pgid — a concurrent release or relaunch owns it otherwise.
func (t Target) ifLeased(pgid int, fn func(path string) error) (bool, error) {
	unlock, err := lock(t.file(".lock"))
	if err != nil {
		return false, err
	}
	defer unlock()
	path := t.file(".json")
	cur, err := readRecord(path)
	if err != nil || cur == nil || cur.PGID != pgid {
		return false, err
	}
	return true, fn(path)
}

func lock(path string) (func(), error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, fmt.Errorf("locking %s: %w", path, err)
	}
	return func() { f.Close() }, nil
}

func readRecord(path string) (*Record, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var rec Record
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("reading lease %s: %w", path, err)
	}
	return &rec, nil
}

func writeRecord(path string, rec Record) error {
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

func logTail(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) > tailLines {
		lines = lines[len(lines)-tailLines:]
	}
	return strings.Join(lines, "\n")
}
