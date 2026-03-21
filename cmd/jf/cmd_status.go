package main

import (
	"flag"
	"fmt"
	"jf/internal/forest"
	"jf/internal/output"
	"os"
	"path/filepath"
	"strings"
)

func runStatus(args []string) int {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	dir := fs.String("dir", ".", "Directory to scan for forest.yml")
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

	state, err := forest.LoadState(f.Dir)
	if err != nil {
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
				pushStale++
				continue
			}
			if state.IsStale(n.Key, info.ModTime()) {
				pushStale++
			}
		}
	}

	if *jsonOut {
		output.Result(output.StatusResult{
			Forest:    f.Dir,
			Total:     len(all),
			TBD:       tbdTotal,
			PushTotal: pushTotal,
			PushStale: pushStale,
			PullTotal: pullTotal,
		})
		return 0
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
