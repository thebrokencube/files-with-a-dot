package main

import (
	"flag"
	"fmt"
	"jf/internal/forest"
	"os"
	"strings"
)

func runDiscover(args []string) int {
	fs := flag.NewFlagSet("discover", flag.ContinueOnError)
	dir := fs.String("dir", ".", "Directory to scan (default: current directory)")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	f, err := forest.FindForest(*dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s\n", err)
		return 1
	}
	if f == nil {
		fmt.Fprintf(os.Stderr, "✗ No forest.yml found (searched up from %s)\n", *dir)
		return 1
	}

	roots, err := forest.Discover(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ Discovery failed: %s\n", err)
		return 1
	}

	if len(roots) == 0 {
		fmt.Println("No jira: nodes found.")
		return 0
	}

	printDiscoverTree(roots, "")
	return 0
}

func printDiscoverTree(nodes []*forest.Node, indent string) {
	for i, n := range nodes {
		connector := "├─"
		childIndent := indent + "│  "
		if i == len(nodes)-1 {
			connector = "└─"
			childIndent = indent + "   "
		}

		syncIcon := "↑"
		if n.Sync == "pull" {
			syncIcon = "↓"
		}

		line := fmt.Sprintf("%s%s %s  %-20s %s  (%s)", indent, connector, syncIcon, n.Key, n.Label, n.File)
		fmt.Println(strings.TrimRight(line, " "))

		if len(n.Children) > 0 {
			printDiscoverTree(n.Children, childIndent)
		}
	}
}
