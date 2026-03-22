package main

import (
	"flag"
	"fmt"
	"jf/internal/forest"
	"strings"
)

func runTree(args []string) int {
	fs := flag.NewFlagSet("tree", flag.ContinueOnError)
	dir := fs.String("dir", ".", "Directory to scan (default: current directory)")

	if err := parseFlags(fs, args); err != nil {
		return 1
	}

	f, roots, code := loadForestOrFail(*dir, false)
	if code != 0 {
		return code
	}

	if len(roots) == 0 {
		fmt.Println("No jira: nodes found.")
		return 0
	}

	// Print header
	fmt.Printf("Forest: %s (%d nodes)\n\n", f.Dir, len(forest.Flatten(roots)))

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
