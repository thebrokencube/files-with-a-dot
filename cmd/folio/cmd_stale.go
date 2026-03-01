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
	noColor := fs.Bool("no-color", false, "Disable colored output")
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
	ps, causedBy := status.DeriveWithDAG(f, folioDir)

	// Build enriched entries for stale/missing/unknown targets
	var entries []output.StaleEntry
	for _, tid := range maputil.SortedKeys(ps.Targets) {
		ts := ps.Targets[tid]
		worst := "clean"
		for _, out := range ts.Outputs {
			if status.StatusRank(out.Status) > status.StatusRank(worst) {
				worst = out.Status
			}
		}
		if worst == "clean" {
			continue
		}

		target := f.Targets[tid]
		entry := output.StaleEntry{
			ID:     tid,
			Status: worst,
			Branch: target.Branch,
		}

		// Collect output labels
		for _, o := range target.Outputs {
			entry.Outputs = append(entry.Outputs, output.OutputLabel(o))
		}

		// Determine cause
		if cause, ok := causedBy[tid]; ok {
			entry.Cause = fmt.Sprintf("blocked by stale target %s", cause)
		} else {
			// Direct cause: check local outputs against sources
			var sourcePaths []string
			for _, s := range target.Sources {
				if s.Path != "" {
					sourcePaths = append(sourcePaths, s.Path)
				}
			}
			for _, out := range target.Outputs {
				if out.Path != "" {
					if c := status.DeriveLocalCause(folioDir, out.Path, sourcePaths); c != "" {
						entry.Cause = c
						break
					}
				} else if out.External != "" {
					entry.Cause = "external output status unknown"
					break
				}
			}
		}

		if entry.Outputs == nil {
			entry.Outputs = []string{}
		}
		entries = append(entries, entry)
	}

	if *jsonMode {
		if entries == nil {
			entries = []output.StaleEntry{}
		}
		data, err := json.Marshal(struct {
			Stale []output.StaleEntry `json:"stale"`
		}{Stale: entries})
		if err != nil {
			fmt.Fprintf(os.Stderr, `{"error":"json marshal error: %s"}`, err)
			fmt.Fprintln(os.Stderr)
			return 1
		}
		fmt.Println(string(data))
	} else {
		output.PrintStaleTerminal(os.Stdout, entries, !*noColor)
	}

	if len(entries) > 0 {
		return 1
	}
	return 0
}
