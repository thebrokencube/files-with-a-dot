package main

import (
	"fmt"
	"os"
)

const version = "0.3.0"

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
	case "add-pending":
		os.Exit(runAddPending(os.Args[2:]))
	case "stale":
		os.Exit(runStale(os.Args[2:]))
	case "touch":
		os.Exit(runTouch(os.Args[2:]))

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
		case "add-pending":
			os.Exit(runAddPending(os.Args[3:]))
		case "--help", "-h", "help":
			printProjectUsage()
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown project command: %s\n", os.Args[2])
			printProjectUsage()
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

Project commands:
  validate     Validate folio.yml structure
  status       Derive and display target state
  stale        List stale/missing/unknown targets
  touch        Mark a target as current (update output mtime)
  init         Bootstrap a new folio.yml
  add-pending  Append an item to the pending list

Utility commands:
  pbcopy       Copy target output to clipboard
  setup        Check folio dependencies
  version      Show version

Command groups:
  home         Repository-level commands (FOLIO_HOME)
  project      Compat alias for project commands

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
  add-pending  Append an item to the pending list
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
`)
}
