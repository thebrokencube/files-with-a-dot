package main

import (
	"flag"
	"fmt"
	"jf/internal/forest"
	"jf/internal/output"
	"strings"
)

func runTree(args []string) int {
	fs := flag.NewFlagSet("tree", flag.ContinueOnError)
	dir := fs.String("dir", ".", "Directory to scan (default: current directory)")
	jsonOut := fs.Bool("json", false, "Output as JSON")
	verbose := fs.Bool("verbose", false, "Show sync direction and file paths")

	if err := parseFlags(fs, args); err != nil {
		return 1
	}

	f, roots, code := loadForestOrFail(*dir, *jsonOut)
	if code != 0 {
		return code
	}

	if *jsonOut {
		all := forest.Flatten(roots)
		var items []output.NodeInfo
		for _, n := range all {
			items = append(items, nodeToInfo(n))
		}
		output.Result(items)
		return 0
	}

	if len(roots) == 0 {
		fmt.Println("No jira: nodes found.")
		return 0
	}

	// Print header
	fmt.Printf("Forest: %s (%d nodes)\n\n", f.Dir, len(forest.Flatten(roots)))

	printTree(roots, "", *verbose)
	return 0
}

func printTree(nodes []*forest.Node, indent string, verbose bool) {
	for i, n := range nodes {
		connector := "├─"
		childIndent := indent + "│  "
		if i == len(nodes)-1 {
			connector = "└─"
			childIndent = indent + "   "
		}

		var line string
		if verbose {
			syncIcon := "↑"
			if n.Sync == "pull" {
				syncIcon = "↓"
			}
			line = fmt.Sprintf("%s%s %s  %-20s %s  (%s)", indent, connector, syncIcon, n.Key, n.Label, n.File)
		} else {
			line = fmt.Sprintf("%s%s %-12s %s", indent, connector, n.Key, n.Label)
		}
		fmt.Println(strings.TrimRight(line, " "))

		if len(n.Children) > 0 {
			printTree(n.Children, childIndent, verbose)
		}
	}
}
