package main

import (
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "push":
		os.Exit(runPush(os.Args[2:]))
	case "pull":
		os.Exit(runPull(os.Args[2:]))
	case "discover":
		os.Exit(runDiscover(os.Args[2:]))
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
	case "sync":
		os.Exit(runSync(os.Args[2:]))
	case "create-missing":
		os.Exit(runCreateMissing(os.Args[2:]))
	case "version":
		fmt.Printf("jf %s\n", version)
		os.Exit(0)
	case "--help", "-h", "help":
		printUsage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: jf <command> [flags]

Level 0 (stateless):
  push <KEY> <FILE>    Compile markdown and push to Jira description
  pull <KEY> <FILE>    Pull Jira description to local file

Level 1 (forest):
  discover             Detect and display forest tree (dry run)
  tree                 Show forest hierarchy
  list                 Flat list of all nodes
  validate             Check forest integrity
  status               Forest summary with staleness
  show <target>        Single-node detail view
  sync                 Push all + pull all
  create-missing       Create Jira tickets for TBD nodes

Utility:
  version              Show version
  help                 Show this help

Run 'jf <command> --help' for details.
`)
}
