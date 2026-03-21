package main

import (
	"flag"
	"fmt"
	"jf/internal/forest"
	"os"
	"path/filepath"
	"strings"
)

func runShow(args []string) int {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	dir := fs.String("dir", ".", "Directory to scan for forest.yml")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	positional := fs.Args()
	if len(positional) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: jf show <target>\n")
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

	node, err := forest.Resolve(roots, positional[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s\n", err)
		return 1
	}

	state, err := forest.LoadState(f.Dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠ Could not load state: %s\n", err)
		state = &forest.State{Nodes: make(map[string]forest.NodeState)}
	}

	// Staleness
	staleStr := "unknown"
	if strings.ToUpper(node.Key) != "TBD" {
		filePath := filepath.Join(f.Dir, node.File)
		info, err := os.Stat(filePath)
		if err == nil {
			if state.IsStale(node.Key, info.ModTime()) {
				staleStr = "stale"
			} else {
				staleStr = "clean"
			}
		}
	}

	// Parent info
	parentStr := "(root)"
	if node.Parent != nil {
		parentStr = node.Parent.Key
	}

	fmt.Printf("Key:      %s\n", node.Key)
	fmt.Printf("Label:    %s\n", node.Label)
	fmt.Printf("Type:     %s\n", node.Type)
	fmt.Printf("Sync:     %s\n", syncDisplay(node.Sync))
	fmt.Printf("File:     %s\n", node.File)
	fmt.Printf("Parent:   %s\n", parentStr)
	fmt.Printf("Children: %d\n", len(node.Children))
	fmt.Printf("Status:   %s\n", staleStr)

	// Show last push time if available
	if ns, ok := state.Nodes[node.Key]; ok && !ns.LastPush.IsZero() {
		fmt.Printf("Pushed:   %s\n", ns.LastPush.Format("2006-01-02 15:04:05"))
	}

	return 0
}

func syncDisplay(sync string) string {
	if sync == "pull" {
		return "pull ↓"
	}
	return "push ↑"
}
