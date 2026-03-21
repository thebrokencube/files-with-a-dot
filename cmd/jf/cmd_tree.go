package main

import (
	"flag"
	"fmt"
	"jf/internal/forest"
	"os"
	"strings"
)

func runTree(args []string) int {
	fs := flag.NewFlagSet("tree", flag.ContinueOnError)
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

	// Print header
	fmt.Printf("Forest: %s (%d nodes)\n\n", f.Dir, countNodes(roots))

	printTree(roots, "")
	return 0
}

func printTree(nodes []*forest.Node, indent string) {
	for i, n := range nodes {
		connector := "├─"
		childIndent := indent + "│  "
		if i == len(nodes)-1 {
			connector = "└─"
			childIndent = indent + "   "
		}

		line := fmt.Sprintf("%s%s %-12s %s", indent, connector, n.Key, n.Label)
		fmt.Println(strings.TrimRight(line, " "))

		if len(n.Children) > 0 {
			printTree(n.Children, childIndent)
		}
	}
}

func countNodes(nodes []*forest.Node) int {
	count := len(nodes)
	for _, n := range nodes {
		count += countNodes(n.Children)
	}
	return count
}
