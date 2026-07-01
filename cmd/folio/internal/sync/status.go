package sync

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/repo"
)

// currentBranch resolves the branch a push would target, PER VCS:
//   - jj (incl. colocated code): the bookmark on @ — git HEAD is detached under
//     jj, so the git ref is meaningless; the bookmark is the branch. Refuses
//     (with jj-appropriate guidance) when @ has no bookmark or several.
//   - git-only: the checked-out branch via symbolic-ref; refuses on detached HEAD.
//
// Refusing on ambiguity is deliberate: never guess a ref to push.
func currentBranch(dir string) (string, error) {
	if repo.IsJJ(dir) {
		return currentJJBookmark(dir)
	}
	out, err := run(dir, "git", "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil || strings.TrimSpace(out) == "" {
		return "", fmt.Errorf("cannot determine current branch (detached HEAD or not a git repo)")
	}
	return strings.TrimSpace(out), nil
}

// currentJJBookmark returns the single bookmark on @, or a helpful error. It
// reads clean bookmark names (no sync markers) via a template map.
func currentJJBookmark(dir string) (string, error) {
	out, err := run(dir, "jj", "--no-pager", "log", "-r", "@", "--no-graph", "-T", `bookmarks.map(|b| b.name()).join("\n")`)
	if err != nil {
		return "", fmt.Errorf("cannot read jj bookmarks: %w", err)
	}
	var names []string
	for _, l := range strings.Split(strings.TrimSpace(out), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			names = append(names, l)
		}
	}
	switch len(names) {
	case 0:
		return "", fmt.Errorf("no bookmark on @ — set one first: jj bookmark set <name> -r @")
	case 1:
		return names[0], nil
	default:
		return "", fmt.Errorf("multiple bookmarks on @ (%s) — ambiguous; leave one on @", strings.Join(names, ", "))
	}
}

// statusTimeout bounds every per-repo probe so one hung repo (auth prompt,
// network stall) never blocks the whole fleet view. fleet status is read-only;
// on any failure the StoreStatus carries Err and the repo renders as "?".
const statusTimeout = 8 * time.Second

// StoreStatus is a cheap, read-only snapshot of one store for `fleet status`.
// A per-repo failure sets Err (rendered as "?"); it is never a hard error.
type StoreStatus struct {
	Name   string `json:"name"`
	Kind   string `json:"kind"`
	Path   string `json:"path"`
	Exists bool   `json:"exists"`
	Branch string `json:"branch,omitempty"`
	Dirty  bool   `json:"dirty"`
	Ahead  int    `json:"ahead,omitempty"`
	Behind int    `json:"behind,omitempty"`
	Detail string `json:"detail,omitempty"`
	Err    string `json:"error,omitempty"`
}

// base builds a StoreStatus pre-filled from the store, checking existence first.
func base(dir string, s config.Store) StoreStatus {
	st := StoreStatus{Name: s.Name, Kind: s.Kind, Path: dir}
	if dir == "" {
		st.Err = "no path"
		return st
	}
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		st.Detail = "missing"
		return st
	}
	st.Exists = true
	return st
}

// run executes name+args in dir with a timeout, returning trimmed stdout.
// stderr is discarded — probes are best-effort and never interactive.
func run(dir, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), statusTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}

// gitStatus reports branch, dirty, and ahead/behind vs upstream for a git repo
// (external and code kinds). Missing upstream leaves ahead/behind at 0.
func gitStatus(dir string, s config.Store) StoreStatus {
	st := base(dir, s)
	if !st.Exists {
		return st
	}
	branch, err := run(dir, "git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		st.Err = "not a git repo (or git unavailable)"
		return st
	}
	st.Branch = branch

	porcelain, err := run(dir, "git", "status", "--porcelain")
	if err != nil {
		st.Err = "git status failed"
		return st
	}
	st.Dirty = porcelain != ""

	// ahead/behind vs upstream, only if one is configured.
	if _, uerr := run(dir, "git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"); uerr == nil {
		if counts, cerr := run(dir, "git", "rev-list", "--left-right", "--count", "@{u}...HEAD"); cerr == nil {
			fields := strings.Fields(counts)
			if len(fields) == 2 {
				st.Behind, _ = strconv.Atoi(fields[0])
				st.Ahead, _ = strconv.Atoi(fields[1])
			}
		}
	}
	st.Detail = summarize(st)
	return st
}

// jjStatus reports the state of a folio KB store (jj). Dirty = @ is non-empty
// (uncommitted working-copy changes); ahead = local main not yet on the remote.
func jjStatus(dir string, s config.Store) StoreStatus {
	st := base(dir, s)
	if !st.Exists {
		return st
	}
	empty, err := run(dir, "jj", "--no-pager", "log", "-r", "@", "--no-graph", "-T", `if(empty, "empty", "changed")`)
	if err != nil {
		st.Err = "not a jj repo (or jj unavailable)"
		return st
	}
	st.Branch = "main"
	st.Dirty = empty == "changed"

	// Unpushed: commits on local main not on main@origin.
	if out, aerr := run(dir, "jj", "--no-pager", "log", "-r", "main@origin..main", "--no-graph", "-T", `"x"`); aerr == nil {
		st.Ahead = len(out) // one "x" per unpushed commit
	}
	st.Detail = summarize(st)
	return st
}

// jjWorkStatus reports a jj (or jj-colocated) code/dot repo: branch = the
// bookmark on @ (git HEAD is detached under jj, so the git ref is stale/wrong),
// dirty = @ has uncommitted changes. This is why code/dot Status is per-VCS —
// git status on a colocated repo mislabels branch and dirtiness.
func jjWorkStatus(dir string, s config.Store) StoreStatus {
	st := base(dir, s)
	if !st.Exists {
		return st
	}
	empty, err := run(dir, "jj", "--no-pager", "log", "-r", "@", "--no-graph", "-T", `if(empty, "empty", "changed")`)
	if err != nil {
		st.Err = "jj status failed (jj unavailable?)"
		return st
	}
	st.Dirty = empty == "changed"
	if bm, berr := run(dir, "jj", "--no-pager", "log", "-r", "@", "--no-graph", "-T", `bookmarks.map(|b| b.name()).join(",")`); berr == nil && strings.TrimSpace(bm) != "" {
		st.Branch = strings.TrimSpace(bm)
	} else {
		st.Branch = "(no bookmark)"
	}
	st.Detail = summarize(st)
	return st
}

// gitFetchOnly refreshes remote refs WITHOUT touching the working tree or
// merging — the only pull a code repo gets (folio never checks out/merges code).
func gitFetchOnly(dir string) error {
	_, err := run(dir, "git", "fetch", "--quiet")
	return err
}

// dotRun shells a `dot` subcommand (pull, etc.) against the dotfiles repo. It is
// timeout-bounded (like every other exec in this package) so an unattended run
// can't hang indefinitely on a network stall or credential prompt. Longer than
// the status probes: `dot pull` does real work (git pull + re-apply symlinks).
func dotRun(dir, sub string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, "dot", sub)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func summarize(st StoreStatus) string {
	var parts []string
	if st.Dirty {
		parts = append(parts, "dirty")
	} else {
		parts = append(parts, "clean")
	}
	if st.Ahead > 0 {
		parts = append(parts, strconv.Itoa(st.Ahead)+" ahead")
	}
	if st.Behind > 0 {
		parts = append(parts, strconv.Itoa(st.Behind)+" behind")
	}
	return strings.Join(parts, ", ")
}
