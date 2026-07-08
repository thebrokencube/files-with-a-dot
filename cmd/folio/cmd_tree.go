package main

import (
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

// buildRoot constructs the folio command tree. It is called fresh per process
// (and per test), so every leaf's flag vars are locals captured by its
// Flags/Run closures — no package-level flag state, preserving reentrancy.
func buildRoot() dendrik.Command {
	return dendrik.Command{
		Name:    "folio",
		Short:   "folio CLI",
		Version: version,
		Sub: []dendrik.Command{
			cmdValidate(),
			cmdStatus(),
			cmdStale(),
			cmdDag(),
			cmdHealth(),
			cmdNew(),
			cmdGather(),
			cmdTouch(),
			cmdObserve(),
			cmdArchive(),
			cmdInit(),
			cmdStores(),
			cmdHome(),
			cmdFleet(),
			cmdSetup(),
			cmdProject(),
		},
	}
}

// cmdProject is a back-compat alias: `folio project <cmd>` routes to the same
// top-level handlers. These commands are also available at the top level.
func cmdProject() dendrik.Command {
	return dendrik.Command{
		Name: "project", Short: "Compat alias for validate/status/init/observe",
		Sub: []dendrik.Command{
			cmdValidate(),
			cmdStatus(),
			cmdInit(),
			cmdObserve(),
		},
	}
}

func cmdValidate() dendrik.Command {
	var folioPath *string
	var jsonMode, noColor *bool
	return dendrik.Command{
		Name: "validate", Short: "Validate folio.yml structure", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) {
			folioPath = fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
			jsonMode = fs.Bool('j', "json", "Machine-readable JSON output")
			noColor = fs.BoolLong("no-color", "Disable colored output")
		},
		Run: func(_ *dendrik.FlagSet, _ []string) int { return runValidate(*folioPath, *jsonMode, *noColor) },
	}
}

func cmdStatus() dendrik.Command {
	var folioPath *string
	var jsonMode, noColor *bool
	return dendrik.Command{
		Name: "status", Short: "Derive and display target state", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) {
			folioPath = fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
			jsonMode = fs.Bool('j', "json", "Machine-readable JSON output")
			noColor = fs.BoolLong("no-color", "Disable colored output")
		},
		Run: func(_ *dendrik.FlagSet, _ []string) int { return runStatus(*folioPath, *jsonMode, *noColor) },
	}
}

func cmdStale() dendrik.Command {
	var folioPath *string
	var jsonMode, noColor *bool
	return dendrik.Command{
		Name: "stale", Short: "List stale/missing/unknown targets", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) {
			folioPath = fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
			jsonMode = fs.Bool('j', "json", "Machine-readable JSON output")
			noColor = fs.BoolLong("no-color", "Disable colored output")
		},
		Run: func(_ *dendrik.FlagSet, _ []string) int { return runStale(*folioPath, *jsonMode, *noColor) },
	}
}

func cmdDag() dendrik.Command {
	var folioPath *string
	var jsonMode, noColor *bool
	return dendrik.Command{
		Name: "dag", Short: "Show target dependency graph", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) {
			folioPath = fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
			jsonMode = fs.Bool('j', "json", "Machine-readable JSON output")
			noColor = fs.BoolLong("no-color", "Disable colored output")
		},
		Run: func(_ *dendrik.FlagSet, _ []string) int { return runDag(*folioPath, *jsonMode, *noColor) },
	}
}

func cmdHealth() dendrik.Command {
	var folioPath *string
	var noColor *bool
	return dendrik.Command{
		Name: "health", Short: "Project health report", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) {
			folioPath = fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
			noColor = fs.BoolLong("no-color", "Disable colored output")
		},
		Run: func(_ *dendrik.FlagSet, _ []string) int { return runHealth(*folioPath, *noColor) },
	}
}

func cmdNew() dendrik.Command {
	var folioPath *string
	var noRegister, dryRun *bool
	return dendrik.Command{
		Name: "new", Short: "Scaffold a typed artifact (spike, design, plan, ...)", Args: dendrik.ArgsExactly(2),
		Flags: func(fs *dendrik.FlagSet) {
			folioPath = fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
			noRegister = fs.BoolLong("no-register", "Skip adding source entry to folio.yml")
			dryRun = fs.Bool('n', "dry-run", "Print what would be created, no side effects")
		},
		Run: func(_ *dendrik.FlagSet, pos []string) int {
			return runNew(*folioPath, *noRegister, *dryRun, pos[0], pos[1])
		},
	}
}

func cmdGather() dendrik.Command {
	var folioPath, typeFlag, name *string
	var materialize, read *bool
	return dendrik.Command{
		Name: "gather", Short: "Add source entry from URL", Args: dendrik.ArgsExactly(1),
		Flags: func(fs *dendrik.FlagSet) {
			folioPath = fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
			materialize = fs.Bool('m', "materialize", "Create reference file stub and wire path")
			typeFlag = fs.String('t', "type", "", "Reference type (spike, survey, design, ...)")
			name = fs.String('n', "name", "", "Reference file name (default: derived from URL)")
			read = fs.Bool('r', "read", "Read and summarize URL (requires Claude skill)")
		},
		Run: func(_ *dendrik.FlagSet, pos []string) int {
			return runGather(*folioPath, *materialize, *typeFlag, *name, *read, pos[0])
		},
	}
}

func cmdTouch() dendrik.Command {
	var folioPath *string
	return dendrik.Command{
		Name: "touch", Short: "Mark a target as current", Args: dendrik.ArgsExactly(1),
		Flags: func(fs *dendrik.FlagSet) {
			folioPath = fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
		},
		Run: func(_ *dendrik.FlagSet, pos []string) int { return runTouch(*folioPath, pos[0]) },
	}
}

func cmdArchive() dendrik.Command {
	var folioPath *string
	var dryRun, noPush *bool
	return dendrik.Command{
		Name: "archive", Short: "Move work track from active to archive", Args: dendrik.ArgsExactly(1),
		Flags: func(fs *dendrik.FlagSet) {
			folioPath = fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
			dryRun = fs.Bool('n', "dry-run", "Print what would happen, no side effects")
			noPush = fs.BoolLong("no-push", "Skip auto-commit")
		},
		Run: func(_ *dendrik.FlagSet, pos []string) int {
			return runArchive(*folioPath, *dryRun, *noPush, pos[0])
		},
	}
}

func cmdInit() dendrik.Command {
	var name, pathFlag *string
	return dendrik.Command{
		Name: "init", Short: "Bootstrap a new folio.yml", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) {
			name = fs.String('n', "name", "", "Project name")
			pathFlag = fs.String('p', "path", "", "Relative path under active/ (e.g., ret/kafka)")
		},
		Run: func(_ *dendrik.FlagSet, _ []string) int { return runInit(*name, *pathFlag) },
	}
}

func cmdSetup() dendrik.Command {
	var checkMode *bool
	return dendrik.Command{
		Name: "setup", Short: "Check folio dependencies", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) {
			checkMode = fs.Bool('c', "check", "Silent mode: exit 0 if OK, exit 1 if missing")
		},
		Run: func(_ *dendrik.FlagSet, _ []string) int { return runSetup(*checkMode) },
	}
}

// cmdObserve is a RunRaw leaf: `folio observe "fix(x): thing"` records a
// free-text observation, so this is a text fallback, not strict dispatch.
func cmdObserve() dendrik.Command {
	return dendrik.Command{
		Name: "observe", Short: "Observation management (add, list, resolve, lint, types)",
		//nolint:dispatch-router // free-text fallback: `observe "fix(x): thing"` is not strict dispatch
		RunRaw: runObserve,
	}
}

// cmdStores is a RunRaw leaf: bare `folio stores` defaults to `stores list`, a
// non-strict convenience the router's empty-group error would break.
func cmdStores() dendrik.Command {
	return dendrik.Command{
		Name: "stores", Short: "List registered stores (multi-store registry)",
		//nolint:dispatch-router // bare `stores` defaults to `stores list` (not strict dispatch)
		RunRaw: runStores,
	}
}

func cmdHome() dendrik.Command {
	return dendrik.Command{
		Name: "home", Short: "FOLIO_HOME commands (list, push, pull, archive, activate, health)",
		Sub: []dendrik.Command{
			cmdHomeInit(),
			cmdHomeValidate(),
			cmdHomeList(),
			cmdHomePush(),
			cmdHomePull(),
			cmdHomeArchive(),
			cmdHomeActivate(),
			cmdHomeHealth(),
			cmdHomeStats(),
			cmdHomeWorkspace(),
		},
	}
}

func cmdHomeInit() dendrik.Command {
	return dendrik.Command{
		Name: "init", Short: "Scaffold FOLIO_HOME directory", Args: dendrik.ArgsNone,
		Run: func(_ *dendrik.FlagSet, _ []string) int { return runHomeInit() },
	}
}

func cmdHomeValidate() dendrik.Command {
	var noColor *bool
	return dendrik.Command{
		Name: "validate", Short: "Structural checks (folio.yml in leaves, date prefixes)", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) { noColor = fs.BoolLong("no-color", "Disable colored output") },
		Run:   func(_ *dendrik.FlagSet, _ []string) int { return runHomeValidate(*noColor) },
	}
}

func cmdHomeList() dendrik.Command {
	var jsonMode, noColor *bool
	return dendrik.Command{
		Name: "list", Short: "Show grouped summary of all folios", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) {
			jsonMode = fs.Bool('j', "json", "Machine-readable JSON output")
			noColor = fs.BoolLong("no-color", "Disable colored output")
		},
		Run: func(_ *dendrik.FlagSet, _ []string) int { return runHomeList(*jsonMode, *noColor) },
	}
}

// cmdHomePush is a RunRaw leaf: an optional [<store>] positional whose presence
// is flag-conditional (push requires -m; the positional names a store, never
// the message), so arity is expressed inside the handler.
func cmdHomePush() dendrik.Command {
	return dendrik.Command{
		Name: "push", Short: "git add + commit (+ push if remote) — requires -m",
		//nolint:dispatch-router // optional [<store>] positional with flag-conditional semantics
		RunRaw: runHomePush,
	}
}

// cmdHomePull is a RunRaw leaf: optional [<store>] positional, self-validated.
func cmdHomePull() dendrik.Command {
	return dendrik.Command{
		Name: "pull", Short: "git pull",
		//nolint:dispatch-router // optional [<store>] positional, self-validated
		RunRaw: runHomePull,
	}
}

func cmdHomeArchive() dendrik.Command {
	return dendrik.Command{
		Name: "archive", Short: "Move active path to archive with date prefix", Args: dendrik.ArgsExactly(1),
		Run: func(_ *dendrik.FlagSet, pos []string) int { return runHomeArchive(pos[0]) },
	}
}

func cmdHomeActivate() dendrik.Command {
	return dendrik.Command{
		Name: "activate", Short: "Move archive path to active, strip date prefix", Args: dendrik.ArgsExactly(1),
		Run: func(_ *dendrik.FlagSet, pos []string) int { return runHomeActivate(pos[0]) },
	}
}

func cmdHomeHealth() dendrik.Command {
	var noColor *bool
	return dendrik.Command{
		Name: "health", Short: "Aggregate health report across all active projects", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) { noColor = fs.BoolLong("no-color", "Disable colored output") },
		Run:   func(_ *dendrik.FlagSet, _ []string) int { return runHomeHealth(*noColor) },
	}
}

func cmdHomeStats() dendrik.Command {
	var noColor *bool
	return dendrik.Command{
		Name: "stats", Short: "Commit statistics for the home repository", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) { noColor = fs.BoolLong("no-color", "Disable colored output") },
		Run:   func(_ *dendrik.FlagSet, _ []string) int { return runHomeStats(*noColor) },
	}
}

func cmdHomeWorkspace() dendrik.Command {
	return dendrik.Command{
		Name: "workspace", Short: "Manage jj workspaces (create|list|cleanup)",
		Sub: []dendrik.Command{
			{Name: "create", Short: "Create a jj workspace for this session (prints its path)", Args: dendrik.ArgsNone,
				Run: func(_ *dendrik.FlagSet, _ []string) int { return runWorkspaceCreate() }},
			{Name: "list", Short: "List all jj workspaces", Args: dendrik.ArgsNone,
				Run: func(_ *dendrik.FlagSet, _ []string) int { return runWorkspaceList() }},
			{Name: "cleanup", Short: "Remove a workspace (path arg, or FOLIO_HOME if it is a workspace)", Args: dendrik.ArgsBetween(0, 1),
				Run: func(_ *dendrik.FlagSet, pos []string) int { return runWorkspaceCleanup(pos) }},
		},
	}
}

func cmdFleet() dendrik.Command {
	return dendrik.Command{
		Name: "fleet", Short: "Multi-store status and work areas",
		Sub: []dendrik.Command{
			cmdFleetStatus(),
			cmdFleetWorkarea(),
		},
	}
}

func cmdFleetStatus() dendrik.Command {
	var dirtyOnly, jsonMode, noColor *bool
	return dendrik.Command{
		Name: "status", Short: "Read-only status across every registered store", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) {
			dirtyOnly = fs.BoolLong("dirty", "Show only stores with uncommitted changes")
			jsonMode = fs.Bool('j', "json", "Machine-readable JSON output")
			noColor = fs.BoolLong("no-color", "Disable colored output")
		},
		Run: func(_ *dendrik.FlagSet, _ []string) int { return runFleetStatus(*dirtyOnly, *jsonMode, *noColor) },
	}
}

func cmdFleetWorkarea() dendrik.Command {
	return dendrik.Command{
		Name: "workarea", Short: "Isolated checkouts (worktree/jj workspace)",
		Sub: []dendrik.Command{
			cmdWorkareaOpen(),
			{Name: "list", Short: "Reconcile ledger vs VCS truth vs disk", Args: dendrik.ArgsNone,
				Run: func(_ *dendrik.FlagSet, _ []string) int { return runWorkareaList() }},
			cmdWorkareaReap(),
		},
	}
}

func cmdWorkareaOpen() dendrik.Command {
	var base *string
	return dendrik.Command{
		Name: "open", Short: "Create an isolated checkout <store> <branch>", Args: dendrik.ArgsExactly(2),
		Flags: func(fs *dendrik.FlagSet) {
			base = fs.String('b', "base", "", "Base branch to fork from (default: store default_branch or main)")
		},
		Run: func(_ *dendrik.FlagSet, pos []string) int { return runWorkareaOpen(*base, pos[0], pos[1]) },
	}
}

func cmdWorkareaReap() dendrik.Command {
	var all, force *bool
	return dendrik.Command{
		Name: "reap", Short: "Remove stale work areas (tier-correct, dirty-guarded)", Args: dendrik.ArgsBetween(0, 1),
		Flags: func(fs *dendrik.FlagSet) {
			all = fs.BoolLong("all", "Reap every eligible work area (default without a branch)")
			force = fs.BoolLong("force", "Remove even dirty/unpushed or severed areas")
		},
		Run: func(_ *dendrik.FlagSet, pos []string) int {
			only := ""
			if len(pos) > 0 {
				only = pos[0]
			}
			return runWorkareaReap(*all, *force, only)
		},
	}
}
