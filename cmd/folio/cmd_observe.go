package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/observe"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runObserve(args []string) int {
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

func runObserveAppend(args []string) int {
	fs := dendrik.NewFlagSet("observe")
	folioPath := fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if !resolveOrDie(folioPath) {
		return 1
	}

	item := strings.TrimSpace(strings.Join(fs.GetArgs(), " "))
	if item == "" {
		fmt.Fprintf(os.Stderr, "Usage: folio observe <item text> [--folio PATH]\n")
		fmt.Fprintf(os.Stderr, "       folio observe list [--json] [--scope X] [--type X]\n")
		fmt.Fprintf(os.Stderr, "       folio observe resolve <match> [match...]\n")
		fmt.Fprintf(os.Stderr, "       folio observe types\n")
		fmt.Fprintf(os.Stderr, "       folio observe lint\n")
		return 1
	}

	if _, err := config.Load(*folioPath); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	if err := observe.Append(*folioPath, item); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	fmt.Println(output.Successf("Added: %s", item))
	return 0
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
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	_ = noColor // reserved for future use

	if !resolveOrDie(folioPath) {
		return 1
	}

	f, err := config.Load(*folioPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
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
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if entries == nil {
			entries = []listEntry{}
		}
		enc.Encode(entries)
		return 0
	}

	if len(entries) == 0 {
		fmt.Println("No observations.")
		return 0
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

	return 0
}

func runObserveResolve(args []string) int {
	fs := dendrik.NewFlagSet("observe resolve")
	folioPath := fs.String('f', "folio", "./folio.yml", "Path or shortname")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if !resolveOrDie(folioPath) {
		return 1
	}

	matches := fs.GetArgs()
	if len(matches) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: folio observe resolve <#N|substring> [match...]\n")
		return 1
	}

	removed, err := observe.Remove(*folioPath, matches)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	for _, item := range removed {
		fmt.Println(output.Successf("Resolved: %s", item))
	}
	return 0
}

func runObserveTypes(args []string) int {
	types := observe.ValidTypes()
	for _, t := range types {
		fmt.Printf("  %-6s %s\n", t.Name, t.Description)
	}
	return 0
}

func runObserveLint(args []string) int {
	fs := dendrik.NewFlagSet("observe lint")
	folioPath := fs.String('f', "folio", "./folio.yml", "Path or shortname")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if !resolveOrDie(folioPath) {
		return 1
	}

	f, err := config.Load(*folioPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	folioDir := filepath.Dir(*folioPath)
	issues := observe.Lint(folioDir, f.Observations)

	if len(issues) == 0 {
		fmt.Println(output.Successf("All observations valid"))
		return 0
	}

	for _, issue := range issues {
		fmt.Fprintf(os.Stderr, "  #%d: %s\n", issue.Index, issue.Reason)
	}
	return 1
}
