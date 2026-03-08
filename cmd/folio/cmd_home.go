package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/home"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/list"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/move"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/repo"
)

func mustResolveHome() string {
	dir, err := home.Dir()
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		os.Exit(1)
	}
	return dir
}

func runHomeInit(args []string) int {
	dir := mustResolveHome()

	if err := home.Init(dir); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	fmt.Println(output.Successf("Initialized FOLIO_HOME at %s", dir))
	return 0
}

func runHomeValidate(args []string) int {
	fs := flag.NewFlagSet("home validate", flag.ExitOnError)
	noColor := fs.Bool("no-color", false, "Disable colored output")
	fs.Parse(args)

	color := colorEnabled(*noColor)
	dir := mustResolveHome()

	errs := home.Validate(dir)

	if len(errs) == 0 {
		if color {
			fmt.Println(output.Successf("FOLIO_HOME structure is valid (%s)", dir))
		} else {
			fmt.Printf("FOLIO_HOME structure is valid (%s)\n", dir)
		}
		return 0
	}

	if color {
		fmt.Fprintf(os.Stderr, "%sErrors:%s\n", output.Red, output.Reset)
	} else {
		fmt.Fprintf(os.Stderr, "Errors:\n")
	}
	for _, e := range errs {
		fmt.Fprintf(os.Stderr, "  - %s\n", e)
	}
	return 2
}

func runHomeList(args []string) int {
	fs := flag.NewFlagSet("home list", flag.ExitOnError)
	jsonMode := fs.Bool("json", false, "Machine-readable JSON output")
	noColor := fs.Bool("no-color", false, "Disable colored output")
	fs.Parse(args)

	color := colorEnabled(*noColor)
	dir := mustResolveHome()

	entries, err := list.Scan(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	if *jsonMode {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(entries)
		return 0
	}

	if len(entries) == 0 {
		fmt.Println("No folios found.")
		return 0
	}

	// Group by section
	active := filterEntries(entries, "active")
	archived := filterEntries(entries, "archive")

	if len(active) > 0 {
		if color {
			fmt.Printf("%sActive%s (%d)\n", output.Bold, output.Reset, len(active))
		} else {
			fmt.Printf("Active (%d)\n", len(active))
		}
		printEntryTable(active, color)
	}

	if len(archived) > 0 {
		if len(active) > 0 {
			fmt.Println()
		}
		if color {
			fmt.Printf("%sArchived%s (%d)\n", output.Bold, output.Reset, len(archived))
		} else {
			fmt.Printf("Archived (%d)\n", len(archived))
		}
		printEntryTable(archived, color)
	}

	return 0
}

func filterEntries(entries []list.Entry, section string) []list.Entry {
	var out []list.Entry
	for _, e := range entries {
		if e.Section == section {
			out = append(out, e)
		}
	}
	return out
}

func printEntryTable(entries []list.Entry, color bool) {
	// Calculate column widths
	pathW, projW := 4, 7 // "Path", "Project" minimums
	for _, e := range entries {
		if len(e.Path) > pathW {
			pathW = len(e.Path)
		}
		if len(e.Project) > projW {
			projW = len(e.Project)
		}
	}

	header := fmt.Sprintf("  %-*s  %-*s  %s  %s", pathW, "Path", projW, "Project", "Targets", "Pending")
	sep := fmt.Sprintf("  %s  %s  %s  %s", strings.Repeat("-", pathW), strings.Repeat("-", projW), "-------", "-------")

	if color {
		fmt.Printf("%s%s%s\n", output.Dim, header, output.Reset)
		fmt.Printf("%s%s%s\n", output.Dim, sep, output.Reset)
	} else {
		fmt.Println(header)
		fmt.Println(sep)
	}

	for _, e := range entries {
		fmt.Printf("  %-*s  %-*s  %7d  %7d\n", pathW, e.Path, projW, e.Project, e.Targets, e.Pending)
	}
}

func runHomePush(args []string) int {
	fs := flag.NewFlagSet("home push", flag.ExitOnError)
	msg := fs.String("m", "", "Commit message: type(scope): description")
	folioName := fs.String("folio", "", "Scope commit to a single folio (shortname or path)")
	all := fs.Bool("all", false, "Stage all changes (current behavior, default)")
	fs.Parse(args)

	// Allow positional args as message for convenience: folio home push "my message"
	if fs.NArg() > 0 {
		*msg = strings.Join(fs.Args(), " ")
	}

	if *msg == "" {
		fmt.Fprintln(os.Stderr, output.Errf("commit message required (-m or positional arg)"))
		fmt.Fprintf(os.Stderr, "  Format: type(scope): description\n")
		fmt.Fprintf(os.Stderr, "  Types:  feat fix docs refactor test chore style perf auto\n")
		return 1
	}

	if *folioName != "" && *all {
		fmt.Fprintln(os.Stderr, output.Errf("--folio and --all are mutually exclusive"))
		return 1
	}

	dir := mustResolveHome()

	var pushErr error
	if *folioName != "" {
		entries, err := list.Scan(dir)
		if err != nil {
			fmt.Fprintln(os.Stderr, output.Errf("%s", err))
			return 1
		}
		var match *list.Entry
		for i, e := range entries {
			if e.Path == *folioName || e.Project == *folioName {
				match = &entries[i]
				break
			}
		}
		if match == nil {
			fmt.Fprintln(os.Stderr, output.Errf("folio %q not found", *folioName))
			return 1
		}
		pushErr = repo.PushScoped(dir, *msg, []string{match.Section + "/" + match.Path})
	} else {
		pushErr = repo.Push(dir, *msg)
	}

	if pushErr != nil {
		if errors.Is(pushErr, repo.ErrNothingToCommit) {
			fmt.Println("Nothing to commit (working tree clean)")
			return 0
		}
		if errors.Is(pushErr, repo.ErrInvalidCommitMessage) {
			fmt.Fprintln(os.Stderr, output.Errf("%s", pushErr))
			return 1
		}
		fmt.Fprintln(os.Stderr, output.Errf("%s", pushErr))
		return 1
	}

	fmt.Println(output.Successf("Committed and pushed"))
	return 0
}

func runHomePull(args []string) int {
	dir := mustResolveHome()

	if err := repo.Pull(dir); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	fmt.Println(output.Successf("Pulled latest"))
	return 0
}

func runHomeArchive(args []string) int {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: folio home archive <path>\n")
		fmt.Fprintf(os.Stderr, "  Path is relative to active/, e.g., 'ben/my-project'\n")
		return 1
	}

	dir := mustResolveHome()
	relPath := args[0]

	if err := move.Archive(dir, relPath); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	fmt.Println(output.Successf("Archived active/%s", relPath))
	return 0
}

func runHomeActivate(args []string) int {
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: folio home activate <path>\n")
		fmt.Fprintf(os.Stderr, "  Path is relative to archive/, e.g., 'ben/2026-02-20-my-project'\n")
		return 1
	}

	dir := mustResolveHome()
	relPath := args[0]

	if err := move.Activate(dir, relPath); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	fmt.Println(output.Successf("Activated archive/%s", relPath))
	return 0
}
