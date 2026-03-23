package main

import (
	"fmt"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/output"
	"os"
	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runStatus(args []string) int {
	fs := dendrik.NewFlagSet("status")
	dir := fs.String('d', "dir", ".", "Directory to scan for forest.yml")
	jsonOut := fs.Bool('j', "json", "Output as JSON")

	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	f, roots, code := loadForestOrFail(*dir, *jsonOut)
	if code != 0 {
		return code
	}

	all := forest.Flatten(roots)

	state, err := forest.LoadState(f.Dir)
	if err != nil {
		state = &forest.State{Nodes: make(map[string]forest.NodeState)}
	}

	var pushTotal, pushStale, pullTotal, pullStale, tbdTotal int

	for _, n := range all {
		if forest.IsTBD(n.Key) {
			tbdTotal++
			continue
		}

		switch n.Sync {
		case "pull":
			pullTotal++
			if state.IsPullStale(n.Key) {
				pullStale++
			}
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
		dendrik.WriteResult(os.Stdout, output.StatusResult{
			Forest:    f.Dir,
			Total:     len(all),
			TBD:       tbdTotal,
			PushTotal: pushTotal,
			PushStale: pushStale,
			PullTotal: pullTotal,
			PullStale: pullStale,
		})
		return dendrik.ExitOK
	}

	fmt.Printf("Forest: %s\n", f.Dir)
	fmt.Printf("Nodes:  %d total", len(all))
	if tbdTotal > 0 {
		fmt.Printf(" (%d TBD)", tbdTotal)
	}
	fmt.Println()
	fmt.Printf("Push:   %d nodes, %d stale\n", pushTotal, pushStale)
	fmt.Printf("Pull:   %d nodes, %d stale\n", pullTotal, pullStale)

	return dendrik.ExitOK
}
