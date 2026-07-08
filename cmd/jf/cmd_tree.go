package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/output"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runTree(dir string, jsonOut, verbose bool) int {
	f, roots, code := loadForestOrFail(dir, jsonOut)
	if code != 0 {
		return code
	}

	if jsonOut {
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
	fmt.Printf("Forest: %s (%d nodes)\n\n", filepath.Dir(f.Dir), len(forest.Flatten(roots)))

	printTree(roots, "", verbose)
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
			line = fmt.Sprintf("%s%s %s  %-20s %s  (%s)", indent, connector, syncIcon, n.Key, n.Label, displayPath(n.File))
		} else {
			line = fmt.Sprintf("%s%s %-12s %s", indent, connector, n.Key, n.Label)
		}
		fmt.Println(strings.TrimRight(line, " "))

		if len(n.Children) > 0 {
			printTree(n.Children, childIndent, verbose)
		}
	}
}
