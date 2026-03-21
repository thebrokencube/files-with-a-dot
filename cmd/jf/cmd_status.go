package main

import (
	"flag"
	"fmt"
	"jf/internal/forest"
	"os"
	"path/filepath"
	"strings"
)

func runStatus(args []string) int {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	dir := fs.String("dir", ".", "Directory to scan for forest.yml")

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

	all := forest.Flatten(roots)
	if len(all) == 0 {
		fmt.Println("No jira: nodes found.")
		return 0
	}

	state, err := forest.LoadState(f.Dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠ Could not load state: %s\n", err)
		state = &forest.State{Nodes: make(map[string]forest.NodeState)}
	}

	var pushTotal, pushStale, pullTotal, tbdTotal int

	for _, n := range all {
		if strings.ToUpper(n.Key) == "TBD" {
			tbdTotal++
			continue
		}

		switch n.Sync {
		case "pull":
			pullTotal++
		default:
			pushTotal++
			filePath := filepath.Join(f.Dir, n.File)
			info, err := os.Stat(filePath)
			if err != nil {
				pushStale++ // can't stat = treat as stale
				continue
			}
			if state.IsStale(n.Key, info.ModTime()) {
				pushStale++
			}
		}
	}

	fmt.Printf("Forest: %s\n", f.Dir)
	fmt.Printf("Nodes:  %d total", len(all))
	if tbdTotal > 0 {
		fmt.Printf(" (%d TBD)", tbdTotal)
	}
	fmt.Println()
	fmt.Printf("Push:   %d nodes, %d stale\n", pushTotal, pushStale)
	fmt.Printf("Pull:   %d nodes\n", pullTotal)

	return 0
}
