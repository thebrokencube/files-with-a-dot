package main

import (
	"fmt"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/output"
	"os"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runTree(args []string) int {
	fs := dendrik.NewFlagSet("tree")
	dir := fs.String('d', "dir", ".", "Directory to scan (default: current directory)")
	jsonOut := fs.Bool('j', "json", "Output as JSON")
	verbose := fs.Bool('v', "verbose", "Show sync direction and file paths")

	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
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
		dendrik.WriteResult(os.Stdout, items)
		return dendrik.ExitOK
	}

	if len(roots) == 0 {
		fmt.Println("No jira: nodes found.")
		return dendrik.ExitOK
	}

	// Print header
	fmt.Printf("Forest: %s (%d nodes)\n\n", f.Dir, len(forest.Flatten(roots)))

	printTree(roots, "", *verbose)
	return dendrik.ExitOK
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
