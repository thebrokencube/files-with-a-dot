package fleet

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

// Tier is the isolation mechanism folio picks per repo — a two-line switch, not
// a framework (design OQ6). The cooperative-lease "shared" tier is deferred:
// a repo that is neither jj nor worktree-capable is refused, not shared.
type Tier string

const (
	TierJJ          Tier = "jj-workspace" // jj repo (incl. colocated) → jj workspace
	TierGitWorktree Tier = "git-worktree" // git-only repo → git worktree
	TierNone        Tier = "none"         // neither → refuse-with-message
)

// DetectTier picks the best available isolation for a repo root. jj wins when
// present (colocated code repos isolate via jj workspaces); else git worktree;
// else none. Never assumes — probes the repo on disk.
func DetectTier(root string) Tier {
	if pathExists(filepath.Join(root, ".jj")) {
		return TierJJ
	}
	if pathExists(filepath.Join(root, ".git")) {
		return TierGitWorktree
	}
	return TierNone
}

// Slug turns a branch name into a filesystem-safe leaf (a/b -> a-b).
func Slug(branch string) string {
	s := strings.ReplaceAll(branch, "/", "-")
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

// CanonicalDir is the one placement: <umbrella>/.worktrees/<store>/<slug>.
func CanonicalDir(umbrella, store, branch string) string {
	return filepath.Join(WorktreesRoot(umbrella), store, Slug(branch))
}

// Open creates an isolated work area for a code/dot store on branch, positioned
// off `base` (default: store.DefaultBranch or "main"). It refuses a repo that
// cannot isolate (TierNone) and refuses to clobber an existing directory.
func Open(umbrella string, store config.Store, branch, base, session string) (WorkArea, error) {
	root := store.Path
	if !pathExists(root) {
		return WorkArea{}, fmt.Errorf("store %q path does not exist: %s", store.Name, root)
	}
	tier := DetectTier(root)
	if tier == TierNone {
		return WorkArea{}, fmt.Errorf("store %q can't isolate (neither jj nor git) — work in place", store.Name)
	}
	if branch == "" {
		return WorkArea{}, fmt.Errorf("a branch name is required")
	}
	if base == "" {
		base = store.DefaultBranch
	}
	if base == "" {
		base = "main"
	}
	dir := CanonicalDir(umbrella, store.Name, branch)
	if pathExists(dir) {
		return WorkArea{}, fmt.Errorf("work area already exists: %s", dir)
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return WorkArea{}, err
	}

	switch tier {
	case TierGitWorktree:
		if err := gitWorktreeAdd(root, dir, branch, base); err != nil {
			return WorkArea{}, err
		}
	case TierJJ:
		if err := jjWorkspaceAdd(root, dir, "fleet-"+Slug(branch)); err != nil {
			return WorkArea{}, err
		}
	}

	wa := WorkArea{
		Store:   store.Name,
		Kind:    store.Kind,
		Tier:    string(tier),
		Dir:     dir,
		Root:    root,
		Branch:  branch,
		Base:    base,
		Session: session,
		Created: time.Now().UTC().Format(time.RFC3339),
	}
	if err := AppendLedger(umbrella, wa); err != nil {
		// The worktree exists but the ledger append failed — surface it; the
		// work area is still usable and List will pick it up as an orphan.
		return wa, fmt.Errorf("work area created at %s but ledger append failed: %w", dir, err)
	}
	return wa, nil
}

// Reconciled is a ledger row cross-checked against VCS truth + disk.
type Reconciled struct {
	WorkArea
	State string // "ok" | "severed" (on disk, VCS forgot) | "gone" (disk absent)
}

// Unledgered states. These read from the VCS's side of the join, so they are
// deliberately distinct from Reconciled.State: "severed" there means the ledger
// and disk agree but git forgot, whereas "dangling" here is the mirror image —
// the VCS still remembers an area whose directory is gone.
const (
	StateStray    = "stray"    // on disk, outside .worktrees, absent from the ledger
	StateDangling = "dangling" // still registered with the VCS; its directory is gone
)

// Unledgered is an isolated checkout the VCS itself knows about but the ledger
// does not — an area made by hand, wherever on disk it happens to sit. Folio did
// not place it and cannot know what it holds, so these are surfaced and never
// auto-reaped.
type Unledgered struct {
	Store string // registry store the area belongs to
	Tier  string // TierJJ | TierGitWorktree
	Name  string // jj workspace name; "" for a git worktree
	Dir   string // resolved path; "" when the registration resolves nowhere
	State string // StateStray | StateDangling
}

// ScanVCS asks each store's own VCS which isolated checkouts it knows about and
// returns the ones the ledger does not cover. This is the truth half of
// reconciliation: Reconcile walks only <umbrella>/.worktrees, so an area created
// by hand beside a repo — or a registration whose directory was deleted — is
// invisible without this. Read-only and best-effort; a store that cannot be
// probed is skipped rather than failing the scan.
func ScanVCS(umbrella string, stores []config.Store) []Unledgered {
	var ledgered []string
	if areas, err := ReadLedger(umbrella); err == nil {
		for _, wa := range areas {
			ledgered = append(ledgered, wa.Dir)
		}
	}

	var out []Unledgered
	for _, s := range stores {
		if s.Path == "" || !pathExists(s.Path) {
			continue
		}
		// Probed independently rather than through DetectTier: a colocated repo
		// can carry both jj workspaces and git worktrees, and DetectTier answers
		// "which tier would folio pick", which is a different question.
		if pathExists(filepath.Join(s.Path, ".jj")) {
			out = append(out, scanJJWorkspaces(s, ledgered)...)
		}
		if pathExists(filepath.Join(s.Path, ".git")) {
			out = append(out, scanGitWorktrees(s, ledgered)...)
		}
	}
	return out
}

// scanJJWorkspaces lists a jj repo's workspaces and resolves each root. jj keeps
// a workspace registered after its directory is deleted, which is the residue
// that accumulates when sessions exit without `workspace forget`.
func scanJJWorkspaces(s config.Store, ledgered []string) []Unledgered {
	// --ignore-working-copy so a stale working copy in ANY workspace doesn't
	// fail the listing, and so a read-only scan never snapshots.
	out, err := run(s.Path, "jj", "--no-pager", "--ignore-working-copy",
		"workspace", "list", "-T", `name ++ "\t" ++ root ++ "\n"`)
	if err != nil {
		return nil
	}
	var found []Unledgered
	for line := range strings.SplitSeq(out, "\n") {
		name, root, ok := strings.Cut(line, "\t")
		if !ok || name == "" {
			continue
		}
		area := Unledgered{Store: s.Name, Tier: string(TierJJ), Name: name}
		// jj renders an unresolvable root as an inline "<Error: …>" instead of
		// failing the listing, so the test is absolute-and-present. The error
		// text itself is not a stable interface and is never parsed.
		if !filepath.IsAbs(root) || !pathExists(root) {
			area.State = StateDangling
			found = append(found, area)
			continue
		}
		if sameDir(root, s.Path) || covered(ledgered, root) {
			continue
		}
		area.Dir, area.State = root, StateStray
		found = append(found, area)
	}
	return found
}

// scanGitWorktrees lists a git repo's worktrees. One porcelain entry is the main
// working copy — the repo itself, never an area.
func scanGitWorktrees(s config.Store, ledgered []string) []Unledgered {
	out, err := run(s.Path, "git", "worktree", "list", "--porcelain")
	if err != nil {
		return nil
	}
	var found []Unledgered
	for line := range strings.SplitSeq(out, "\n") {
		dir, ok := strings.CutPrefix(strings.TrimSpace(line), "worktree ")
		if !ok || dir == "" {
			continue
		}
		if sameDir(dir, s.Path) || covered(ledgered, dir) {
			continue
		}
		area := Unledgered{Store: s.Name, Tier: string(TierGitWorktree), Dir: dir, State: StateStray}
		if !pathExists(dir) {
			area.State = StateDangling
		}
		found = append(found, area)
	}
	return found
}

// covered reports whether the ledger already accounts for dir.
func covered(ledgered []string, dir string) bool {
	for _, l := range ledgered {
		if sameDir(l, dir) {
			return true
		}
	}
	return false
}

// sameDir compares paths after resolving symlinks, so macOS's /tmp vs
// /private/tmp — or a symlinked store path — doesn't read as two directories.
func sameDir(a, b string) bool {
	if a == b {
		return true
	}
	ra, ea := filepath.EvalSymlinks(a)
	rb, eb := filepath.EvalSymlinks(b)
	return ea == nil && eb == nil && ra == rb
}

// Reconcile intersects the ledger with on-disk reality. Orphans (dirs under
// .worktrees not in the ledger) are returned separately — listed, never trusted.
func Reconcile(umbrella string) (rows []Reconciled, orphans []string, err error) {
	areas, err := ReadLedger(umbrella)
	if err != nil {
		return nil, nil, err
	}
	ledgered := map[string]bool{}
	for _, wa := range areas {
		ledgered[wa.Dir] = true
		r := Reconciled{WorkArea: wa}
		switch {
		case !pathExists(wa.Dir):
			r.State = "gone"
		case Tier(wa.Tier) == TierGitWorktree && !gitKnowsWorktree(wa):
			r.State = "severed"
		default:
			r.State = "ok"
		}
		rows = append(rows, r)
	}

	// Scan for orphan directories present under .worktrees but not in the ledger.
	base := WorktreesRoot(umbrella)
	storeDirs, _ := os.ReadDir(base)
	for _, sd := range storeDirs {
		if !sd.IsDir() {
			continue
		}
		leaves, _ := os.ReadDir(filepath.Join(base, sd.Name()))
		for _, lf := range leaves {
			if !lf.IsDir() {
				continue
			}
			p := filepath.Join(base, sd.Name(), lf.Name())
			if !ledgered[p] {
				orphans = append(orphans, p)
			}
		}
	}
	return rows, orphans, nil
}

// Reap removes work areas. It is tier-correct and safe:
//   - git worktree: `git worktree remove` + `prune` (NEVER os.RemoveAll — a stale
//     .git/worktrees/<name> would block re-checkout elsewhere)
//   - jj workspace: `jj workspace forget` + os.RemoveAll
//   - "gone" rows: just pruned from the ledger
//   - dirty/unpushed: skipped unless force
//   - severed/orphan: reported, never auto-removed unless force
//
// It returns human-readable action lines. `only` (a branch or dir) limits reap
// to one area; empty means all eligible.
func Reap(umbrella string, only string, force bool) ([]string, error) {
	rows, orphans, err := Reconcile(umbrella)
	if err != nil {
		return nil, err
	}
	var actions []string
	var keep []WorkArea

	for _, r := range rows {
		if only != "" && r.Branch != only && r.Dir != only {
			keep = append(keep, r.WorkArea)
			continue
		}
		switch r.State {
		case "gone":
			actions = append(actions, fmt.Sprintf("pruned ledger row (dir gone): %s", r.Dir))
			// dropped from keep → pruned
		case "severed":
			if !force {
				actions = append(actions, fmt.Sprintf("SEVERED, kept (use --force): %s", r.Dir))
				keep = append(keep, r.WorkArea)
				continue
			}
			pruneSevered(r.WorkArea)
			actions = append(actions, fmt.Sprintf("force-removed severed (pruned parent): %s", r.Dir))
		case "ok":
			if !force && isDirtyOrUnpushed(r.WorkArea) {
				actions = append(actions, fmt.Sprintf("dirty/unpushed, kept (use --force): %s", r.Dir))
				keep = append(keep, r.WorkArea)
				continue
			}
			if err := removeArea(r.WorkArea, force); err != nil {
				actions = append(actions, fmt.Sprintf("FAILED to remove %s: %v", r.Dir, err))
				keep = append(keep, r.WorkArea)
				continue
			}
			actions = append(actions, fmt.Sprintf("removed (%s): %s", r.Tier, r.Dir))
		}
	}
	if err := WriteLedger(umbrella, keep); err != nil {
		return actions, err
	}
	for _, o := range orphans {
		actions = append(actions, fmt.Sprintf("orphan (unledgered, not touched): %s", o))
	}
	return actions, nil
}

// removeArea dispatches tier-correct removal. force passes through to
// `git worktree remove --force` (needed to remove a dirty/locked worktree).
func removeArea(wa WorkArea, force bool) error {
	switch Tier(wa.Tier) {
	case TierGitWorktree:
		// `git worktree remove` deletes the dir AND deregisters it, leaving no
		// stale .git/worktrees/<name>. Run from the parent repo (Root) so it
		// works even if the worktree's own gitdir link is flaky; prune after to
		// clean any residual admin entry. NEVER os.RemoveAll a git worktree.
		from := gitOpDir(wa)
		args := []string{"worktree", "remove"}
		if force {
			args = append(args, "--force")
		}
		args = append(args, wa.Dir)
		if _, err := run(from, "git", args...); err != nil {
			return err
		}
		_, _ = run(from, "git", "worktree", "prune")
		return nil
	case TierJJ:
		from := gitOpDir(wa)
		_, _ = run(from, "jj", "--no-pager", "workspace", "forget", "fleet-"+Slug(wa.Branch))
		return os.RemoveAll(wa.Dir)
	default:
		return os.RemoveAll(wa.Dir)
	}
}

// gitOpDir returns the best directory to run a VCS command from: the parent repo
// Root when known (survives a severed worktree), else the work area dir.
func gitOpDir(wa WorkArea) string {
	if wa.Root != "" && pathExists(wa.Root) {
		return wa.Root
	}
	return wa.Dir
}

// pruneSevered removes a severed git worktree safely: the dir is gone-from-git
// but present on disk, so os.RemoveAll the dir THEN prune the parent repo's
// admin entry so no dangling .git/worktrees/<name> remains.
func pruneSevered(wa WorkArea) {
	_ = os.RemoveAll(wa.Dir)
	if Tier(wa.Tier) == TierGitWorktree && wa.Root != "" && pathExists(wa.Root) {
		_, _ = run(wa.Root, "git", "worktree", "prune")
	}
}

// isDirtyOrUnpushed guards reap: true if the work area has uncommitted changes
// or commits not on its upstream (git). Best-effort; on probe failure returns
// true (safer to keep than to destroy).
func isDirtyOrUnpushed(wa WorkArea) bool {
	if Tier(wa.Tier) == TierGitWorktree {
		out, err := run(wa.Dir, "git", "status", "--porcelain")
		if err != nil {
			return true
		}
		if strings.TrimSpace(out) != "" {
			return true // uncommitted changes
		}
		// Unpushed relative to the upstream, if one is configured.
		if _, uerr := run(wa.Dir, "git", "rev-parse", "--abbrev-ref", "@{u}"); uerr == nil {
			counts, cerr := run(wa.Dir, "git", "rev-list", "--count", "@{u}..HEAD")
			if cerr != nil {
				return true // can't tell → keep
			}
			return strings.TrimSpace(counts) != "0"
		}
		// No upstream: the branch was never pushed. Committed work here would be
		// LOST on reap, so treat any commits beyond the base branch as unpushed.
		base := wa.Base
		if base == "" {
			base = "main"
		}
		counts, cerr := run(wa.Dir, "git", "rev-list", "--count", base+"..HEAD")
		if cerr != nil {
			return true // base ref missing / can't tell → keep, don't destroy
		}
		return strings.TrimSpace(counts) != "0"
	}
	// jj: treat non-empty @ as dirty.
	out, err := run(wa.Dir, "jj", "--no-pager", "log", "-r", "@", "--no-graph", "-T", `if(empty,"e","c")`)
	if err != nil {
		return true
	}
	return strings.TrimSpace(out) == "c"
}

// --- low-level helpers ---

func gitWorktreeAdd(root, dir, branch, base string) error {
	// Reuse an existing branch if it exists; otherwise create it off base.
	if _, err := run(root, "git", "rev-parse", "--verify", "--quiet", "refs/heads/"+branch); err == nil {
		_, e := run(root, "git", "worktree", "add", dir, branch)
		return e
	}
	_, err := run(root, "git", "worktree", "add", dir, "-b", branch, base)
	return err
}

func jjWorkspaceAdd(root, dir, name string) error {
	_, err := run(root, "jj", "--no-pager", "workspace", "add", "--name", name, dir)
	return err
}

func gitKnowsWorktree(wa WorkArea) bool {
	// Ask the PARENT repo (Root) — a severed worktree's own dir may not resolve
	// git, which would falsely report "known". Root is authoritative.
	out, err := run(gitOpDir(wa), "git", "worktree", "list", "--porcelain")
	if err != nil {
		return false
	}
	return strings.Contains(out, wa.Dir)
}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func run(dir, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return string(out), err
}
