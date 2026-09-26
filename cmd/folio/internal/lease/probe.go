package lease

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

const probeTimeout = 10 * time.Second

// listeners returns every process listening on TCP port, with its pgid.
func listeners(port int) ([]Owner, error) {
	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "lsof", "-nP", "-iTCP:"+strconv.Itoa(port), "-sTCP:LISTEN", "-Fp").Output()
	if err != nil {
		// lsof exits 1 when nothing matches the selection.
		var ee *exec.ExitError
		if !errors.As(err, &ee) || ee.ExitCode() != 1 {
			return nil, fmt.Errorf("probing port %d with lsof: %w", port, err)
		}
	}
	var owners []Owner
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.HasPrefix(line, "p") {
			continue
		}
		pid, err := strconv.Atoi(line[1:])
		if err != nil {
			continue
		}
		pgid, err := syscall.Getpgid(pid)
		if errors.Is(err, syscall.ESRCH) {
			continue
		}
		if err != nil {
			pgid = -1
		}
		owners = append(owners, Owner{PID: pid, PGID: pgid})
	}
	return owners, nil
}

// portOwners probes every declared port.
func portOwners(s Serve) (map[int][]Owner, error) {
	owners := make(map[int][]Owner, len(s.Processes))
	for _, p := range s.Processes {
		o, err := listeners(p.Port)
		if err != nil {
			return nil, err
		}
		owners[p.Port] = o
	}
	return owners, nil
}

// groupAlive reports whether any process is still in process group pgid.
func groupAlive(pgid int) bool {
	if pgid <= 1 {
		return false
	}
	err := syscall.Kill(-pgid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}

// signalGroup signals a lease pgid, refusing the caller's own group and the
// pgid values kill(2) would widen to every process.
func signalGroup(pgid int, sig syscall.Signal) error {
	if pgid <= 1 || pgid == syscall.Getpgrp() {
		return fmt.Errorf("refusing to signal process group %d", pgid)
	}
	return syscall.Kill(-pgid, sig)
}

// inCheckout reports whether cwd's realpath is exactly the store path; a jj
// workspace or git worktree nested inside the checkout must not match.
func inCheckout(cwd, storePath string) bool {
	dir, err := config.CanonicalPath(cwd)
	return err == nil && dir == storePath
}

// envFacts splits require_env into present and missing names and digests the
// present NAME=value lines. An empty value counts as missing.
func envFacts(keys []string) (present, missing []string, digest string) {
	present = []string{}
	h := sha256.New()
	for _, k := range slices.Sorted(slices.Values(keys)) {
		v := os.Getenv(k)
		if v == "" {
			missing = append(missing, k)
			continue
		}
		present = append(present, k)
		fmt.Fprintf(h, "%s=%s\n", k, v)
	}
	return present, missing, "sha256:" + hex.EncodeToString(h.Sum(nil))
}
