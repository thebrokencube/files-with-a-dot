package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/maputil"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/status"
)

func runStale(args []string) int {
	fs := flag.NewFlagSet("stale", flag.ExitOnError)
	folioPath := fs.String("folio", "./folio.yml", "Path to folio.yml")
	jsonMode := fs.Bool("json", false, "Machine-readable JSON output")
	fs.Parse(args)

	if _, err := os.Stat(*folioPath); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, output.Errf("folio.yml not found at %s", *folioPath))
		return 1
	}

	f, err := config.Load(*folioPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	folioDir := filepath.Dir(*folioPath)
	ps, _ := status.DeriveWithDAG(f, folioDir)

	// Collect stale/missing/unknown target IDs
	var stale []string
	for _, tid := range maputil.SortedKeys(ps.Targets) {
		ts := ps.Targets[tid]
		for _, out := range ts.Outputs {
			if out.Status == "stale" || out.Status == "missing" || out.Status == "unknown" {
				stale = append(stale, tid)
				break
			}
		}
	}

	if *jsonMode {
		if stale == nil {
			stale = []string{}
		}
		data, err := json.Marshal(map[string][]string{"stale": stale})
		if err != nil {
			fmt.Fprintf(os.Stderr, `{"error":"json marshal error: %s"}`, err)
			fmt.Fprintln(os.Stderr)
			return 1
		}
		fmt.Println(string(data))
	} else {
		for _, tid := range stale {
			fmt.Println(tid)
		}
	}

	if len(stale) > 0 {
		return 1
	}
	return 0
}
