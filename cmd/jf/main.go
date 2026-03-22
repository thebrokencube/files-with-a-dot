package main

import (
	"fmt"
	"jf/internal/setup"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	// Commands that skip prereq checks
	switch cmd {
	case "setup":
		os.Exit(runSetup(os.Args[2:]))
	case "init":
		os.Exit(runInit(os.Args[2:]))
	case "schema":
		os.Exit(runSchema(os.Args[2:]))
	case "version":
		fmt.Printf("jf %s\n", version)
		os.Exit(0)
	case "--help", "-h", "help":
		printUsage()
		os.Exit(0)
	}

	// Local-only commands (no Jira auth needed)
	switch cmd {
	case "tree":
		os.Exit(runTree(os.Args[2:]))
	case "list":
		os.Exit(runList(os.Args[2:]))
	case "validate":
		os.Exit(runValidate(os.Args[2:]))
	case "status":
		os.Exit(runStatus(os.Args[2:]))
	case "show":
		os.Exit(runShow(os.Args[2:]))
	case "rm":
		os.Exit(runRm(os.Args[2:]))
	}

	// Prereq guard for Jira-touching commands
	if msg := setup.QuickCheck(setup.DefaultChecker); msg != "" {
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	switch cmd {
	case "push":
		os.Exit(runPush(os.Args[2:]))
	case "pull":
		os.Exit(runPull(os.Args[2:]))
	case "sync":
		os.Exit(runSync(os.Args[2:]))
	case "create-missing":
		os.Exit(runCreateMissing(os.Args[2:]))
	case "search":
		os.Exit(runSearch(os.Args[2:]))
	case "clone":
		os.Exit(runClone(os.Args[2:]))
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: jf <command> [flags]

Getting started:
  clone <KEY>          Scaffold local forest from Jira hierarchy
  init                 Create forest.yml in current directory
  setup                Check prerequisites (node, acli, auth)

Sync workflow:
  push <KEY> <FILE>    Compile markdown and push to Jira description
  pull <KEY> <FILE>    Pull Jira description to local file
  sync                 Push all stale + pull all pull-mode nodes

Forest inspection:
  tree                 Show forest hierarchy
  list                 Flat list of all nodes
  show <target>        Single-node detail view
  status               Forest summary with staleness
  validate             Check forest integrity

Ticket management:
  create-missing       Create Jira tickets for TBD nodes
  search <text>        Find Jira tickets by text/project/type
  rm <KEY>...          Remove node files from forest

Other:
  schema               Emit JSON Schema for forest.yml and frontmatter
  version              Show version

Flags must come before positional arguments.
Run 'jf <command> --help' for details.
`)
}
