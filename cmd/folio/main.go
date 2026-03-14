package main

import (
	"fmt"
	"os"
)

const version = "0.6.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	// Top-level project commands (canonical)
	case "validate":
		os.Exit(runValidate(os.Args[2:]))
	case "status":
		os.Exit(runStatus(os.Args[2:]))
	case "init":
		os.Exit(runInit(os.Args[2:]))
	case "observe":
		os.Exit(runObserve(os.Args[2:]))
	case "stale":
		os.Exit(runStale(os.Args[2:]))
	case "touch":
		os.Exit(runTouch(os.Args[2:]))
	case "dag":
		os.Exit(runDag(os.Args[2:]))
	case "gather":
		os.Exit(runGather(os.Args[2:]))
	case "new":
		os.Exit(runNew(os.Args[2:]))
	case "health":
		os.Exit(runHealth(os.Args[2:]))
	case "archive":
		os.Exit(runArchive(os.Args[2:]))

	// Top-level utility commands
	case "pbcopy":
		os.Exit(runPbcopy(os.Args[2:]))
	case "setup":
		os.Exit(runSetup(os.Args[2:]))
	case "version":
		fmt.Printf("folio %s\n", version)
		os.Exit(0)
	case "--help", "-h", "help":
		printUsage()
		os.Exit(0)

	// Compat: folio project <cmd> routes to the same functions
	case "project":
		if len(os.Args) < 3 {
			printProjectUsage()
			os.Exit(1)
		}
		switch os.Args[2] {
		case "validate":
			os.Exit(runValidate(os.Args[3:]))
		case "status":
			os.Exit(runStatus(os.Args[3:]))
		case "init":
			os.Exit(runInit(os.Args[3:]))
		case "observe":
			os.Exit(runObserve(os.Args[3:]))
		case "--help", "-h", "help":
			printProjectUsage()
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown project command: %s\n", os.Args[2])
			printProjectUsage()
			os.Exit(1)
		}

	case "jira":
		if len(os.Args) < 3 {
			printJiraUsage()
			os.Exit(1)
		}
		switch os.Args[2] {
		case "lint":
			os.Exit(runJiraLint(os.Args[3:]))
		case "compile":
			os.Exit(runJiraCompile(os.Args[3:]))
		case "push":
			os.Exit(runJiraPush(os.Args[3:]))
		case "create":
			os.Exit(runJiraCreate(os.Args[3:]))
		case "view":
			os.Exit(runJiraView(os.Args[3:]))
		case "search":
			os.Exit(runJiraSearch(os.Args[3:]))
		case "--help", "-h", "help":
			printJiraUsage()
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown jira command: %s\n", os.Args[2])
			printJiraUsage()
			os.Exit(1)
		}

	case "home":
		if len(os.Args) < 3 {
			printHomeUsage()
			os.Exit(1)
		}
		switch os.Args[2] {
		case "init":
			os.Exit(runHomeInit(os.Args[3:]))
		case "validate":
			os.Exit(runHomeValidate(os.Args[3:]))
		case "list":
			os.Exit(runHomeList(os.Args[3:]))
		case "push":
			os.Exit(runHomePush(os.Args[3:]))
		case "pull":
			os.Exit(runHomePull(os.Args[3:]))
		case "archive":
			os.Exit(runHomeArchive(os.Args[3:]))
		case "activate":
			os.Exit(runHomeActivate(os.Args[3:]))
		case "health":
			os.Exit(runHomeHealth(os.Args[3:]))
		case "--help", "-h", "help":
			printHomeUsage()
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown home command: %s\n", os.Args[2])
			printHomeUsage()
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: folio <command> [flags]

Data commands:
  validate     Validate folio.yml structure
  status       Derive and display target state
  stale        List stale/missing/unknown targets
  dag          Show target dependency graph
  health       Project health report

Composition:
  new          Scaffold a typed artifact (spike, design, plan, ...)
  gather       Add source entry from URL
  touch        Mark a target as current
  observe      Append an observation
  archive      Move work track from active to archive
  pbcopy       Copy target output to clipboard

Integrations:
  jira         Jira pipeline (lint, compile, push, create, view, search)

Management:
  home         FOLIO_HOME commands (list, push, pull, archive, activate, health)
  init         Bootstrap a new folio.yml
  setup        Check folio dependencies
  version      Show version

Run 'folio <command> --help' for details.
`)
}

func printProjectUsage() {
	fmt.Fprintf(os.Stderr, `Usage: folio project <command> [flags]

Compat alias — these commands are also available at top level (folio validate, etc.).

Commands:
  validate     Validate folio.yml structure
  status       Derive and display target state
  init         Bootstrap a new folio.yml
  observe      Append an observation
`)
}

func printHomeUsage() {
	fmt.Fprintf(os.Stderr, `Usage: folio home <command> [flags]

Commands:
  init       Scaffold FOLIO_HOME directory
  validate   Structural checks (folio.yml in leaves, date prefixes)
  list       Show grouped summary of all folios
  push       git add + commit (+ push if remote) — requires -m
  pull       git pull
  archive    Move active path to archive with date prefix
  activate   Move archive path to active, strip date prefix
  health     Aggregate health report across all active projects
`)
}
