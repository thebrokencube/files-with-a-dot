package main

import (
	"fmt"
	"os"
)

var version = "dev" // overridden at build time via -ldflags -X main.version (see Makefile/VERSION)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "lint":
		os.Exit(runLint(os.Args[2:]))
	case "build":
		os.Exit(runBuild(os.Args[2:]))
	case "version", "--version", "-V":
		fmt.Printf("dendrik %s\n", version)
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
	fmt.Fprintf(os.Stderr, `Usage: dendrik <command> [flags]

Commands:
  lint <path>    Run tool contract validation
  build [dir]    Build a tool's release artifacts (reproducible, version-stamped)
  version        Show version

Run 'dendrik <command> --help' for details.
`)
}
