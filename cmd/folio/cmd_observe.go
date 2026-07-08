package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/home"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/observe"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/repo"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runObserve(args []string) int {
	if len(args) > 0 && dendrik.IsHelpArg(args[0]) {
		printObserveUsage()
		return dendrik.ExitOK
	}
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		switch args[0] {
		case "list":
			return runObserveList(args[1:])
		case "resolve":
			return runObserveResolve(args[1:])
		case "types":
			return runObserveTypes(args[1:])
		case "lint":
			return runObserveLint(args[1:])
		}
	}
	return runObserveAppend(args)
}

func printObserveUsage() {
	fmt.Fprintf(os.Stderr, "Usage: folio observe <item text> [--folio PATH]\n")
	fmt.Fprintf(os.Stderr, "       folio observe list [--json] [--scope X] [--type X]\n")
	fmt.Fprintf(os.Stderr, "       folio observe resolve <match> [match...]\n")
	fmt.Fprintf(os.Stderr, "       folio observe types\n")
	fmt.Fprintf(os.Stderr, "       folio observe lint\n")
}

func runObserveAppend(args []string) int {
	pal := dendrik.NewPalette(true)
	fs := dendrik.NewFlagSet("observe")
	folioPath := fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
	noSync := fs.BoolLong("no-sync", "Skip pull-before/push-after (default syncs automatically)")
	if done, code := dendrik.ParseCheck(fs, args); done {
		return code
	}

	if !resolveOrDie(folioPath) {
		return dendrik.ExitUserError
	}

	item := strings.TrimSpace(strings.Join(fs.GetArgs(), " "))
	if item == "" {
		fmt.Fprintf(os.Stderr, "Usage: folio observe <item text> [--folio PATH]\n")
		fmt.Fprintf(os.Stderr, "       folio observe list [--json] [--scope X] [--type X]\n")
		fmt.Fprintf(os.Stderr, "       folio observe resolve <match> [match...]\n")
		fmt.Fprintf(os.Stderr, "       folio observe types\n")
		fmt.Fprintf(os.Stderr, "       folio observe lint\n")
		return dendrik.ExitUserError
	}

	if _, err := config.Load(*folioPath); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}

	homeDir, homeErr := home.Dir()
	sync := !*noSync && homeDir != ""
	if !*noSync && homeDir == "" && homeErr != nil {
		fmt.Fprintf(os.Stderr, "%ssync disabled: %s%s\n", pal.Dim, homeErr, pal.Reset)
	}

	if sync {
		if err := repo.Pull(homeDir); err != nil {
			fmt.Fprintln(os.Stderr, pal.Errf("sync pull: %s", err))
			return dendrik.ExitUserError
		}
	}

	if err := observe.Append(*folioPath, item); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}

	fmt.Println(pal.Successf("Added: %s", item))

	if sync {
		typ, scope, _, _ := observe.ParseObservation(item)
		msg := fmt.Sprintf("auto(observe): add %s(%s)", typ, scope)
		if err := repo.Push(homeDir, msg); err != nil {
			if errors.Is(err, repo.ErrNothingToCommit) {
				return dendrik.ExitOK
			}
			fmt.Fprintln(os.Stderr, pal.Errf("sync push: %s — run 'folio home push' to retry", err))
			return dendrik.ExitUserError
		}
		fmt.Println(pal.Successf("Synced"))
	}

	return dendrik.ExitOK
}

type listEntry struct {
	Index       int    `json:"index"`
	Type        string `json:"type"`
	Scope       string `json:"scope"`
	Description string `json:"description"`
}

func runObserveList(args []string) int {
	fs := dendrik.NewFlagSet("observe list")
	folioPath := fs.String('f', "folio", "./folio.yml", "Path or shortname")
	jsonMode := fs.Bool('j', "json", "Machine-readable JSON output")
	scopeFilter := fs.String('s', "scope", "", "Filter by scope")
	typeFilter := fs.String('t', "type", "", "Filter by type")
	noColor := fs.BoolLong("no-color", "Disable colored output")
	if done, code := dendrik.ParseCheck(fs, args); done {
		return code
	}

	pal := dendrik.NewPalette(true)
	_ = noColor // reserved for future use

	if !resolveOrDie(folioPath) {
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

	// Parse all observations
	var entries []listEntry
	for i, item := range f.Observations {
		typ, scope, desc, err := observe.ParseObservation(item)
		if err != nil {
			// Skip unparseable in display (lint catches those)
			continue
		}
		if *scopeFilter != "" && scope != *scopeFilter {
			continue
		}
		if *typeFilter != "" && typ != *typeFilter {
			continue
		}
		entries = append(entries, listEntry{
			Index:       i + 1,
			Type:        typ,
			Scope:       scope,
			Description: desc,
		})
	}

	if *jsonMode {
		if entries == nil {
			entries = []listEntry{}
		}
		dendrik.WriteResult(os.Stdout, entries)
		return dendrik.ExitOK
	}

	if len(entries) == 0 {
		fmt.Println("No observations.")
		return dendrik.ExitOK
	}

	// Group by scope
	grouped := map[string][]listEntry{}
	for _, e := range entries {
		grouped[e.Scope] = append(grouped[e.Scope], e)
	}
	scopes := make([]string, 0, len(grouped))
	for s := range grouped {
		scopes = append(scopes, s)
	}
	sort.Strings(scopes)

	for i, scope := range scopes {
		fmt.Printf("%s:\n", scope)
		for _, e := range grouped[scope] {
			fmt.Printf("  #%-3d %s: %s\n", e.Index, e.Type, e.Description)
		}
		if i < len(scopes)-1 {
			fmt.Println()
		}
	}

	return dendrik.ExitOK
}

func runObserveResolve(args []string) int {
	pal := dendrik.NewPalette(true)
	fs := dendrik.NewFlagSet("observe resolve")
	folioPath := fs.String('f', "folio", "./folio.yml", "Path or shortname")
	noSync := fs.BoolLong("no-sync", "Skip pull-before/push-after (default syncs automatically)")
	if done, code := dendrik.ParseCheck(fs, args); done {
		return code
	}

	if !resolveOrDie(folioPath) {
		return dendrik.ExitUserError
	}

	matches := fs.GetArgs()
	if len(matches) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: folio observe resolve <#N|substring> [match...]\n")
		return dendrik.ExitUserError
	}

	homeDir, homeErr := home.Dir()
	sync := !*noSync && homeDir != ""
	if !*noSync && homeDir == "" && homeErr != nil {
		fmt.Fprintf(os.Stderr, "%ssync disabled: %s%s\n", pal.Dim, homeErr, pal.Reset)
	}

	if sync {
		if err := repo.Pull(homeDir); err != nil {
			fmt.Fprintln(os.Stderr, pal.Errf("sync pull: %s", err))
			return dendrik.ExitUserError
		}
	}

	removed, err := observe.Remove(*folioPath, matches)
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}

	for _, item := range removed {
		fmt.Println(pal.Successf("Resolved: %s", item))
	}

	if sync {
		msg := fmt.Sprintf("auto(observe): resolve %d observations", len(removed))
		if err := repo.Push(homeDir, msg); err != nil {
			if errors.Is(err, repo.ErrNothingToCommit) {
				return dendrik.ExitOK
			}
			fmt.Fprintln(os.Stderr, pal.Errf("sync push: %s — run 'folio home push' to retry", err))
			return dendrik.ExitUserError
		}
		fmt.Println(pal.Successf("Synced"))
	}

	return dendrik.ExitOK
}

func runObserveTypes(args []string) int {
	types := observe.ValidTypes()
	for _, t := range types {
		fmt.Printf("  %-6s %s\n", t.Name, t.Description)
	}
	return dendrik.ExitOK
}

func runObserveLint(args []string) int {
	pal := dendrik.NewPalette(true)
	fs := dendrik.NewFlagSet("observe lint")
	folioPath := fs.String('f', "folio", "./folio.yml", "Path or shortname")
	if done, code := dendrik.ParseCheck(fs, args); done {
		return code
	}

	if !resolveOrDie(folioPath) {
		return dendrik.ExitUserError
	}

	f, err := config.Load(*folioPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}

	folioDir := filepath.Dir(*folioPath)
	issues := observe.Lint(folioDir, f.Observations)

	if len(issues) == 0 {
		fmt.Println(pal.Successf("All observations valid"))
		return dendrik.ExitOK
	}

	for _, issue := range issues {
		fmt.Fprintf(os.Stderr, "  #%d: %s\n", issue.Index, issue.Reason)
	}
	return dendrik.ExitUserError
}
