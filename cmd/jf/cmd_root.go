package main

import (
	"fmt"
	"os"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/setup"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

// buildRoot constructs the jf command tree. It is called fresh per process (and
// per test), so every leaf's flag vars are locals captured by its Flags/Run
// closures — no package-level flag state, preserving reentrancy.
func buildRoot() dendrik.Command {
	return dendrik.Command{
		Name:    "jf",
		Short:   "jf CLI",
		Version: version,
		Sub: []dendrik.Command{
			cmdSetup(),
			cmdInit(),
			cmdSchema(),
			cmdTree(),
			cmdList(),
			cmdValidate(),
			cmdStatus(),
			cmdShow(),
			cmdRm(),
			cmdURL(),
			cmdPush(),
			cmdPull(),
			cmdSync(),
			cmdCreateMissing(),
			cmdSearch(),
			cmdClone(),
			cmdView(),
		},
	}
}

// requireJira runs the prereq guard (node, acli, auth) that every Jira-touching
// leaf must clear before it runs. It replaces the subset-guard that used to sit
// between the local-only and Jira-touching switch arms in main(): the local-only
// leaves never call it, the Jira-touching leaves call it at the top of their Run.
func requireJira() int {
	if msg := setup.QuickCheck(setup.DefaultChecker); msg != "" {
		fmt.Fprintln(os.Stderr, msg)
		return dendrik.ExitUserError
	}
	return dendrik.ExitOK
}

// --- local-only leaves (no Jira auth needed) ---

func cmdSetup() dendrik.Command {
	var checkOnly, jsonOut, discover *bool
	return dendrik.Command{
		Name: "setup", Short: "Check prerequisites (node, acli, auth)", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) {
			checkOnly = fs.Bool('c', "check", "Non-interactive check only")
			jsonOut = fs.Bool('j', "json", "Output as JSON (with --check)")
			discover = fs.BoolLong("discover", "Discover and save Jira site from acli auth")
		},
		Run: func(_ *dendrik.FlagSet, _ []string) int { return runSetup(*checkOnly, *jsonOut, *discover) },
	}
}

func cmdInit() dendrik.Command {
	var project, dir *string
	return dendrik.Command{
		Name: "init", Short: "Create forest.yml in current directory", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) {
			project = fs.String('p', "project", "BEN", "Jira project key")
			dir = fs.String('d', "dir", ".", "Directory to create .jf/forest.yml in")
		},
		Run: func(_ *dendrik.FlagSet, _ []string) int { return runInit(*project, *dir) },
	}
}

func cmdSchema() dendrik.Command {
	var outputSchema *bool
	return dendrik.Command{
		Name: "schema", Short: "Emit JSON Schema for forest.yml and frontmatter", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) {
			outputSchema = fs.Bool('o', "output", "Emit command output schemas instead of input schemas")
		},
		Run: func(_ *dendrik.FlagSet, _ []string) int { return runSchema(*outputSchema) },
	}
}

func cmdTree() dendrik.Command {
	var dir *string
	var jsonOut, verbose *bool
	return dendrik.Command{
		Name: "tree", Short: "Show forest hierarchy", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) {
			dir = fs.String('d', "dir", ".", "Directory to scan (default: current directory)")
			jsonOut = fs.Bool('j', "json", "Output as JSON")
			verbose = fs.Bool('v', "verbose", "Show sync direction and file paths")
		},
		Run: func(_ *dendrik.FlagSet, _ []string) int { return runTree(*dir, *jsonOut, *verbose) },
	}
}

func cmdList() dendrik.Command {
	var dir *string
	var jsonOut *bool
	return dendrik.Command{
		Name: "list", Short: "Flat list of all nodes", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) {
			dir = fs.String('d', "dir", ".", "Directory to scan (default: current directory)")
			jsonOut = fs.Bool('j', "json", "Output as JSON")
		},
		Run: func(_ *dendrik.FlagSet, _ []string) int { return runList(*dir, *jsonOut) },
	}
}

func cmdValidate() dendrik.Command {
	var dir *string
	var jsonOut *bool
	return dendrik.Command{
		Name: "validate", Short: "Check forest integrity", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) {
			dir = fs.String('d', "dir", ".", "Directory to scan (default: current directory)")
			jsonOut = fs.Bool('j', "json", "Output as JSON")
		},
		Run: func(_ *dendrik.FlagSet, _ []string) int { return runValidate(*dir, *jsonOut) },
	}
}

func cmdStatus() dendrik.Command {
	var dir *string
	var jsonOut *bool
	return dendrik.Command{
		Name: "status", Short: "Forest summary with staleness", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) {
			dir = fs.String('d', "dir", ".", "Directory to scan for forest.yml")
			jsonOut = fs.Bool('j', "json", "Output as JSON")
		},
		Run: func(_ *dendrik.FlagSet, _ []string) int { return runStatus(*dir, *jsonOut) },
	}
}

func cmdShow() dendrik.Command {
	var dir *string
	var jsonOut *bool
	return dendrik.Command{
		Name: "show", Short: "Single-node detail view", Args: dendrik.ArgsExactly(1),
		Flags: func(fs *dendrik.FlagSet) {
			dir = fs.String('d', "dir", ".", "Directory to scan for forest.yml")
			jsonOut = fs.Bool('j', "json", "Output as JSON")
		},
		Run: func(_ *dendrik.FlagSet, pos []string) int { return runShow(*dir, *jsonOut, pos[0]) },
	}
}

func cmdRm() dendrik.Command {
	var dir *string
	return dendrik.Command{
		Name: "rm", Short: "Remove node files from forest", Args: dendrik.ArgsBetween(1, -1),
		Flags: func(fs *dendrik.FlagSet) {
			dir = fs.String('d', "dir", ".", "Directory to scan for forest.yml")
		},
		Run: func(_ *dendrik.FlagSet, pos []string) int { return runRm(*dir, pos) },
	}
}

func cmdURL() dendrik.Command {
	return dendrik.Command{
		Name: "url", Short: "Print browse URL for an issue", Args: dendrik.ArgsExactly(1),
		Run: func(_ *dendrik.FlagSet, pos []string) int { return runURL(pos[0]) },
	}
}

// --- Jira-touching leaves (guarded by requireJira) ---

func cmdPush() dendrik.Command {
	var plainText, dryRun, jsonOut, yes *bool
	var subtree, dir *string
	return dendrik.Command{
		Name: "push", Short: "Compile markdown and push to Jira description", Args: dendrik.ArgsBetween(0, 2),
		Flags: func(fs *dendrik.FlagSet) {
			plainText = fs.Bool('p', "plain-text", "Push as plain text if marklassian conversion fails")
			subtree = fs.String('s', "subtree", "", "Push node and all descendants")
			dir = fs.String('d', "dir", ".", "Directory to scan for forest.yml")
			dryRun = fs.Bool('n', "dry-run", "Preview what would be pushed without side effects")
			jsonOut = fs.Bool('j', "json", "Output plan as structured JSON")
			yes = fs.BoolLong("yes", "Proceed without confirmation in non-interactive mode")
		},
		Run: func(_ *dendrik.FlagSet, pos []string) int {
			if code := requireJira(); code != dendrik.ExitOK {
				return code
			}
			return runPush(*plainText, *subtree, *dir, *dryRun, *jsonOut, *yes, pos)
		},
	}
}

func cmdPull() dendrik.Command {
	var dryRun, jsonOut, yes *bool
	var subtree, dir *string
	return dendrik.Command{
		Name: "pull", Short: "Pull Jira description to local file", Args: dendrik.ArgsBetween(0, 2),
		Flags: func(fs *dendrik.FlagSet) {
			subtree = fs.String('s', "subtree", "", "Pull node and all descendants")
			dir = fs.String('d', "dir", ".", "Directory to scan for forest.yml")
			dryRun = fs.Bool('n', "dry-run", "Preview what would be pulled without side effects")
			jsonOut = fs.Bool('j', "json", "Output plan as structured JSON")
			yes = fs.BoolLong("yes", "Proceed without confirmation in non-interactive mode")
		},
		Run: func(_ *dendrik.FlagSet, pos []string) int {
			if code := requireJira(); code != dendrik.ExitOK {
				return code
			}
			return runPull(*subtree, *dir, *dryRun, *jsonOut, *yes, pos)
		},
	}
}

func cmdSync() dendrik.Command {
	var dryRun, scaffold, plainText, jsonOut, yes *bool
	var dir, resolve *string
	return dendrik.Command{
		Name: "sync", Short: "Push all stale + pull all pull-mode nodes", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) {
			dir = fs.String('d', "dir", ".", "Directory to scan for forest.yml")
			resolve = fs.String('r', "resolve", "", "Conflict resolution: local|remote (default: block)")
			dryRun = fs.Bool('n', "dry-run", "Preview what would be synced without side effects")
			scaffold = fs.BoolLong("scaffold", "Create stub files for new Jira children")
			plainText = fs.Bool('p', "plain-text", "Push as plain text if marklassian conversion fails")
			jsonOut = fs.Bool('j', "json", "Output plan as structured JSON")
			yes = fs.BoolLong("yes", "Proceed without confirmation in non-interactive mode")
		},
		Run: func(_ *dendrik.FlagSet, _ []string) int {
			if code := requireJira(); code != dendrik.ExitOK {
				return code
			}
			return runSync(*dir, *resolve, *dryRun, *scaffold, *plainText, *jsonOut, *yes)
		},
	}
}

func cmdCreateMissing() dendrik.Command {
	var dryRun, plainText *bool
	var dir *string
	return dendrik.Command{
		Name: "create-missing", Short: "Create Jira tickets for TBD nodes", Args: dendrik.ArgsNone,
		Flags: func(fs *dendrik.FlagSet) {
			dir = fs.String('d', "dir", ".", "Directory to scan for forest.yml")
			dryRun = fs.Bool('n', "dry-run", "Show what would be created without side effects")
			plainText = fs.Bool('p', "plain-text", "Push as plain text if marklassian conversion fails")
		},
		Run: func(_ *dendrik.FlagSet, _ []string) int {
			if code := requireJira(); code != dendrik.ExitOK {
				return code
			}
			return runCreateMissing(*dir, *dryRun, *plainText)
		},
	}
}

func cmdSearch() dendrik.Command {
	var project, issueType *string
	var limit *int
	var jsonOut *bool
	return dendrik.Command{
		Name: "search", Short: "Find Jira tickets by text/project/type", Args: dendrik.ArgsBetween(0, -1),
		Flags: func(fs *dendrik.FlagSet) {
			project = fs.String('p', "project", "", "Filter by Jira project key")
			issueType = fs.String('t', "type", "", "Filter by issue type (Epic, Story, Task, etc.)")
			limit = fs.Int('l', "limit", 50, "Maximum results")
			jsonOut = fs.Bool('j', "json", "Output raw JSON from Jira")
		},
		Run: func(_ *dendrik.FlagSet, pos []string) int {
			if code := requireJira(); code != dendrik.ExitOK {
				return code
			}
			return runSearch(*project, *issueType, *limit, *jsonOut, pos)
		},
	}
}

func cmdClone() dendrik.Command {
	var dir, syncMode, repo *string
	var depth *int
	return dendrik.Command{
		Name: "clone", Short: "Scaffold local forest from Jira hierarchy", Args: dendrik.ArgsExactly(1),
		Flags: func(fs *dendrik.FlagSet) {
			dir = fs.String('d', "dir", ".", "Parent directory for cloned forest")
			depth = fs.IntLong("depth", 0, "Max hierarchy depth (0 = unlimited)")
			syncMode = fs.StringLong("sync", "", "Sync direction override for scaffolded nodes: push|pull|both (default: omit, derives from mutability)")
			repo = fs.StringLong("repo", "", "GitHub org/repo for PR badge enrichment (e.g. Gusto/hawaiian-ice)")
		},
		Run: func(_ *dendrik.FlagSet, pos []string) int {
			if code := requireJira(); code != dendrik.ExitOK {
				return code
			}
			return runClone(*dir, *depth, *syncMode, *repo, pos[0])
		},
	}
}

func cmdView() dendrik.Command {
	var fields *string
	var jsonOut *bool
	return dendrik.Command{
		Name: "view", Short: "Fetch remote issue details from Jira", Args: dendrik.ArgsExactly(1),
		Flags: func(fs *dendrik.FlagSet) {
			fields = fs.String('f', "fields", "summary,status,issuetype", "Comma-separated fields to display")
			jsonOut = fs.Bool('j', "json", "Output raw JSON from Jira")
		},
		Run: func(_ *dendrik.FlagSet, pos []string) int {
			if code := requireJira(); code != dendrik.ExitOK {
				return code
			}
			return runView(*fields, *jsonOut, pos[0])
		},
	}
}
