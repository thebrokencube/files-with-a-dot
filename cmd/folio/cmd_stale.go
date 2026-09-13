package main

import (
	"fmt"
	"os"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/graph"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/maputil"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/status"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runStale(folioPath string, jsonMode, noColor bool) int {
	pal := dendrik.NewPalette(true)

	ctx, err := resolveContext(folioPath, contextReadOnly)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}
	folioPath = ctx.FolioPath

	if _, err := os.Stat(folioPath); os.IsNotExist(err) {
		if jsonMode {
			dendrik.WriteError(os.Stdout, fmt.Sprintf("folio.yml not found at %s", folioPath), "")
		} else {
			fmt.Fprintln(os.Stderr, pal.Errf("folio.yml not found at %s", folioPath))
		}
		return dendrik.ExitUserError
	}

	f, err := config.Load(folioPath)
	if err != nil {
		if jsonMode {
			dendrik.WriteError(os.Stdout, fmt.Sprintf("%s", err), "")
		} else {
			fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		}
		return dendrik.ExitUserError
	}

	var entries []output.StaleEntry
	ps, causedBy, causedStatus := status.DeriveWithDAGWithContext(f, ctx)
	blocking := false
	for _, tid := range maputil.SortedKeys(ps.Targets) {
		ts := ps.Targets[tid]
		target := f.Targets[tid]
		worst, hasLocal := targetOutputStatus(ts)
		if target.Forest != nil || (ts.Final && hasLocal && worst == "clean") {
			continue
		}
		if !hasLocal {
			worst = "unknown"
		}
		if worst == "clean" {
			continue
		}
		if worst == "stale" || worst == "missing" {
			blocking = true
		}

		entry := output.StaleEntry{ID: tid, Status: worst}
		for _, item := range target.Outputs {
			entry.Outputs = append(entry.Outputs, output.OutputLabel(item))
		}
		if cause, ok := causedBy[tid]; ok {
			entry.Cause = fmt.Sprintf("blocked by %s target %s", causedStatus[tid], cause)
		} else {
			entry.Cause = directTargetCause(ts, target)
		}
		if entry.Cause == "" {
			entry.Cause = "composition not recorded"
		}
		if entry.Outputs == nil {
			entry.Outputs = []string{}
		}
		entries = append(entries, entry)
	}

	code := dendrik.ExitOK
	if blocking {
		code = dendrik.ExitUserError
	}

	if jsonMode {
		if entries == nil {
			entries = []output.StaleEntry{}
		}
		dendrik.WriteResultWithExit(os.Stdout, struct {
			Stale []output.StaleEntry `json:"stale"`
		}{Stale: entries}, code)
	} else {
		output.PrintStaleTerminal(os.Stdout, entries, !noColor)
	}

	return code
}

func targetOutputStatus(target status.TargetStatus) (string, bool) {
	worst := "clean"
	hasLocal := false
	for _, item := range target.Outputs {
		if item.Type != "local" {
			continue
		}
		hasLocal = true
		if graph.StatusRank(item.Status) > graph.StatusRank(worst) {
			worst = item.Status
		}
	}
	return worst, hasLocal
}

func directTargetCause(target status.TargetStatus, cfg config.Target) string {
	for _, item := range target.Outputs {
		if item.Type == "local" && item.Cause != "" {
			return item.Cause
		}
	}
	for _, source := range cfg.Sources {
		if source.External != "" {
			return "external source freshness unverified"
		}
	}
	for _, item := range target.Outputs {
		if item.Type == "external" {
			return "external output status unknown"
		}
	}
	return ""
}
