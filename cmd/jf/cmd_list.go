package main

import (
	"flag"
	"fmt"
	"jf/internal/forest"
	"jf/internal/output"
	"os"
)

func runList(args []string) int {
	fs := flag.NewFlagSet("list", flag.ContinueOnError)
	dir := fs.String("dir", ".", "Directory to scan (default: current directory)")
	jsonOut := fs.Bool("json", false, "Output as JSON")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	f, err := forest.FindForest(*dir)
	if err != nil {
		if *jsonOut {
			output.Error(err.Error(), "")
		} else {
			fmt.Fprintf(os.Stderr, "✗ %s\n", err)
		}
		return 1
	}
	if f == nil {
		if *jsonOut {
			output.Error("No forest.yml found", *dir)
		} else {
			fmt.Fprintf(os.Stderr, "✗ No forest.yml found (searched up from %s)\n", *dir)
		}
		return 1
	}

	roots, err := forest.Discover(f)
	if err != nil {
		if *jsonOut {
			output.Error("Discovery failed", err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "✗ Discovery failed: %s\n", err)
		}
		return 1
	}

	all := forest.Flatten(roots)

	if *jsonOut {
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

	for _, n := range all {
		sync := "push"
		if n.Sync == "pull" {
			sync = "pull"
		}
		fmt.Printf("%-12s %-5s %s\n", n.Key, sync, n.File)
	}
	return 0
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
