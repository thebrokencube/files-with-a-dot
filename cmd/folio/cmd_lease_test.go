package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/lease"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik/cmdtest"
)

const listenerPortEnv = "FOLIO_LEASE_TEST_PORT"

// TestLeaseListenerHelper is the declared server in TestLeaseLifecycle: the
// test binary re-execs itself with listenerPortEnv set and listens until signalled.
func TestLeaseListenerHelper(t *testing.T) {
	port := os.Getenv(listenerPortEnv)
	if port == "" {
		return
	}
	l, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		t.Fatal(err)
	}
	for {
		c, err := l.Accept()
		if err != nil {
			return
		}
		c.Close()
	}
}

func TestLeaseLifecycle(t *testing.T) {
	if _, err := exec.LookPath("lsof"); err != nil {
		t.Skip("lsof not on PATH")
	}
	bin := cmdtest.Build(t, "folio")
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	umbrella := t.TempDir()
	storeDir := filepath.Join(t.TempDir(), "app")
	elsewhere := t.TempDir()
	if err := os.MkdirAll(storeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	port := freePort(t)
	run := fmt.Sprintf("%s=%d '%s' '-test.run=^TestLeaseListenerHelper$' -test.timeout=5m", listenerPortEnv, port, self)
	stores := fmt.Sprintf(`schema: 3
stores:
  app:
    path: %s
    kind: code
    location: referenced
    serve:
      url: http://127.0.0.1:%d
      processes:
        - { name: web, port: %d, run: %q }
      require_env: [FOLIO_LEASE_TEST_TOKEN]
      alternative: use the running server
`, storeDir, port, port, run)
	if err := os.WriteFile(filepath.Join(umbrella, "stores.yml"), []byte(stores), 0o644); err != nil {
		t.Fatal(err)
	}
	leaseFile := filepath.Join(umbrella, ".fleet", "leases", "app.json")

	folio := func(dir, verb string, flags ...string) (int, lease.Report) {
		t.Helper()
		cmd := exec.Command(bin, append([]string{"lease", verb, "app", "--json"}, flags...)...)
		cmd.Dir = dir
		for _, kv := range os.Environ() {
			if !strings.HasPrefix(kv, "FOLIO_") {
				cmd.Env = append(cmd.Env, kv)
			}
		}
		cmd.Env = append(cmd.Env, "FOLIO_UMBRELLA="+umbrella, "FOLIO_LEASE_TEST_TOKEN=s3cret")
		out, err := cmd.Output()
		code := 0
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			code = ee.ExitCode()
		} else if err != nil {
			t.Fatalf("folio lease %s: %v", verb, err)
		}
		var env struct{ Data lease.Report }
		if err := json.Unmarshal(out, &env); err != nil {
			var stderr []byte
			if ee != nil {
				stderr = ee.Stderr
			}
			t.Fatalf("folio lease %s: exit %d, decoding %q: %v\n%s", verb, code, out, err, stderr)
		}
		return code, env.Data
	}
	live := map[int]bool{}
	t.Cleanup(func() {
		for pgid := range live {
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		}
	})
	started := func(r lease.Report) {
		t.Helper()
		if r.PGID <= 1 {
			t.Fatalf("started report has no pgid: %+v", r)
		}
		live[r.PGID] = true
	}
	gone := func(pgid int) {
		t.Helper()
		waitFor(t, fmt.Sprintf("pgid %d to exit", pgid), func() bool {
			return errors.Is(syscall.Kill(-pgid, 0), syscall.ESRCH)
		})
		delete(live, pgid)
	}

	code, dry := folio(storeDir, "run", "--dry-run")
	if code != dendrik.ExitOK || dry.Outcome != "would-start" {
		t.Fatalf("dry-run run = %d %+v, want 0 would-start", code, dry)
	}
	if listening(port) {
		t.Fatalf("port %d listening after a dry run", port)
	}
	if _, err := os.Stat(filepath.Dir(leaseFile)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("dry run touched the lease dir: %v", err)
	}

	code, first := folio(storeDir, "run")
	if code != dendrik.ExitOK || first.Outcome != "started" {
		t.Fatalf("first run = %d %+v, want 0 started", code, first)
	}
	started(first)
	if !listening(port) {
		t.Fatalf("port %d not listening after started", port)
	}
	recBefore, err := os.ReadFile(leaseFile)
	if err != nil {
		t.Fatal(err)
	}
	var rec lease.Record
	if err := json.Unmarshal(recBefore, &rec); err != nil {
		t.Fatal(err)
	}
	if rec.PGID != first.PGID || rec.Ports[strconv.Itoa(port)] == 0 || !strings.HasPrefix(rec.EnvDigest, "sha256:") {
		t.Fatalf("lease record = %+v, want pgid %d owning port %d with a digest", rec, first.PGID, port)
	}
	if strings.Contains(string(recBefore), "s3cret") {
		t.Fatal("lease record contains an env value")
	}

	code, again := folio(storeDir, "run")
	if code != dendrik.ExitOK || again.Outcome != "running" || again.PGID != first.PGID {
		t.Fatalf("repeat run = %d %+v, want 0 running pgid %d", code, again, first.PGID)
	}
	if recAfter, _ := os.ReadFile(leaseFile); string(recAfter) != string(recBefore) {
		t.Fatal("repeat run rewrote the lease")
	}

	code, held := folio(elsewhere, "run")
	if code != dendrik.ExitConflict || held.Outcome != "held" || !strings.Contains(held.Message, "use the running server") {
		t.Fatalf("run from another dir = %d %+v, want 3 held with alternative", code, held)
	}

	code, st := folio(elsewhere, "status")
	if code != dendrik.ExitOK || st.State != lease.StateRunning || st.Env != "match" || st.PGID != first.PGID {
		t.Fatalf("status = %d %+v, want 0 running env=match", code, st)
	}

	code, dryRel := folio(storeDir, "release", "--dry-run")
	if code != dendrik.ExitOK || dryRel.Outcome != "would-release" || dryRel.PGID != first.PGID {
		t.Fatalf("dry-run release = %d %+v, want 0 would-release pgid %d", code, dryRel, first.PGID)
	}
	if !listening(port) {
		t.Fatalf("port %d stopped by a dry-run release", port)
	}
	if recAfter, _ := os.ReadFile(leaseFile); string(recAfter) != string(recBefore) {
		t.Fatal("dry-run release changed the lease")
	}

	if code, r := folio(elsewhere, "release"); code != dendrik.ExitUserError {
		t.Fatalf("release from another dir = %d %+v, want 1", code, r)
	}
	code, rel := folio(storeDir, "release")
	if code != dendrik.ExitOK || rel.Outcome != "released" {
		t.Fatalf("release = %d %+v, want 0 released", code, rel)
	}
	gone(first.PGID)
	if listening(port) {
		t.Fatalf("port %d still listening after release", port)
	}
	if _, err := os.Stat(leaseFile); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("lease file after release: %v, want not exist", err)
	}

	code, second := folio(storeDir, "run")
	if code != dendrik.ExitOK || second.Outcome != "started" {
		t.Fatalf("run after release = %d %+v, want 0 started", code, second)
	}
	started(second)
	if err := syscall.Kill(-second.PGID, syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	gone(second.PGID)
	waitFor(t, "port to free", func() bool { return !listening(port) })
	if code, st := folio(storeDir, "status"); code != dendrik.ExitOK || st.State != lease.StateFree {
		t.Fatalf("status with a stale lease = %d %+v, want 0 free", code, st)
	}

	code, third := folio(storeDir, "run")
	if code != dendrik.ExitOK || third.Outcome != "started" || third.PGID == second.PGID {
		t.Fatalf("run over a stale lease = %d %+v, want 0 started with a new pgid", code, third)
	}
	started(third)
	if code, r := folio(storeDir, "release"); code != dendrik.ExitOK {
		t.Fatalf("final release = %d %+v", code, r)
	}
	gone(third.PGID)
}

func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func listening(port int) bool {
	c, err := net.DialTimeout("tcp", "127.0.0.1:"+strconv.Itoa(port), time.Second)
	if err != nil {
		return false
	}
	c.Close()
	return true
}

func waitFor(t *testing.T, what string, ok func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !ok() {
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", what)
		}
		time.Sleep(100 * time.Millisecond)
	}
}
