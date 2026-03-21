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

Utility:
  version              Show version
  help                 Show this help

Run 'jf <command> --help' for details.
`)
}
