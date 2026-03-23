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
	case "lint":
		os.Exit(runLint(os.Args[2:]))
	case "version":
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
  lint <path>    Run 25-check tool contract validation
  version        Show version

Run 'dendrik <command> --help' for details.
`)
}
