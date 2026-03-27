package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/maputil"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/status"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runStale(args []string) int {
	pal := dendrik.NewPalette(true)
	fs := dendrik.NewFlagSet("stale")
	folioPath := fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
	jsonMode := fs.Bool('j', "json", "Machine-readable JSON output")
	noColor := fs.BoolLong("no-color", "Disable colored output")
	if done, code := dendrik.ParseCheck(fs, args); done {
		return code
	}

	if !resolveOrDie(folioPath) {
		return dendrik.ExitUserError
	}

	if _, err := os.Stat(*folioPath); os.IsNotExist(err) {
		if *jsonMode {
			dendrik.WriteError(os.Stdout, fmt.Sprintf("folio.yml not found at %s", *folioPath), "")
		} else {
			fmt.Fprintln(os.Stderr, pal.Errf("folio.yml not found at %s", *folioPath))
		}
		return dendrik.ExitUserError
	}

	f, err := config.Load(*folioPath)
	if err != nil {
		if *jsonMode {
			dendrik.WriteError(os.Stdout, fmt.Sprintf("%s", err), "")
		} else {
			fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		}
		return dendrik.ExitUserError
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
		dendrik.WriteResult(os.Stdout, struct {
			Stale []output.StaleEntry `json:"stale"`
		}{Stale: entries})
	} else {
		output.PrintStaleTerminal(os.Stdout, entries, !*noColor)
	}

	if len(entries) > 0 {
		return dendrik.ExitUserError
	}
	return dendrik.ExitOK
}
