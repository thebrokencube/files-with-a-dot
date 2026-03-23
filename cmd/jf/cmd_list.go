package main

import (
	"fmt"
	"jf/internal/forest"
	"jf/internal/output"
	"os"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runList(args []string) int {
	fs := dendrik.NewFlagSet("list")
	dir := fs.String('d', "dir", ".", "Directory to scan (default: current directory)")
	jsonOut := fs.Bool('j', "json", "Output as JSON")

	if err := dendrik.Parse(fs, args); err != nil {
		return dendrik.ExitUserError
	}

	_, roots, code := loadForestOrFail(*dir, *jsonOut)
	if code != 0 {
		return code
	}

	all := forest.Flatten(roots)

	if *jsonOut {
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

	for _, n := range all {
		sync := "push"
		if n.Sync == "pull" {
			sync = "pull"
		}
		fmt.Printf("%-12s %-5s %s\n", n.Key, sync, n.File)
	}
	return dendrik.ExitOK
}

func nodeToInfo(n *forest.Node) output.NodeInfo {
	parent := ""
	if n.Parent != nil {
		parent = n.Parent.Key
	}
	return output.NodeInfo{
		Key:      n.Key,
		Label:    n.Label,
		Type:     n.Type,
		Sync:     n.Sync,
		File:     n.File,
		Parent:   parent,
		Children: len(n.Children),
	}
}
