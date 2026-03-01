package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/home"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/list"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/move"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/repo"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/status"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/validate"
)

const version = "0.2.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "project":
		if len(os.Args) < 3 {
			printProjectUsage()
			os.Exit(1)
		}
		switch os.Args[2] {
		case "validate":
			os.Exit(runValidate(os.Args[3:]))
		case "status":
			os.Exit(runStatus(os.Args[3:]))
		case "init":
			os.Exit(runInit(os.Args[3:]))
		case "--help", "-h", "help":
			printProjectUsage()
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown project command: %s\n", os.Args[2])
			printProjectUsage()
			os.Exit(1)
		}

	case "home":
		if len(os.Args) < 3 {
			printHomeUsage()
			os.Exit(1)
		}
		switch os.Args[2] {
		case "init":
			os.Exit(runHomeInit(os.Args[3:]))
		case "validate":
			os.Exit(runHomeValidate(os.Args[3:]))
		case "list":
			os.Exit(runHomeList(os.Args[3:]))
		case "push":
			os.Exit(runHomePush(os.Args[3:]))
		case "pull":
			os.Exit(runHomePull(os.Args[3:]))
		case "archive":
			os.Exit(runHomeArchive(os.Args[3:]))
		case "activate":
			os.Exit(runHomeActivate(os.Args[3:]))
		case "--help", "-h", "help":
			printHomeUsage()
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "Unknown home command: %s\n", os.Args[2])
			printHomeUsage()
			os.Exit(1)
		}

	// Top-level commands
	case "setup":
		os.Exit(runSetup(os.Args[2:]))
	case "version":
		fmt.Printf("folio %s\n", version)
		os.Exit(0)
	case "--help", "-h", "help":
		printUsage()
		os.Exit(0)

	// Backward compatibility: bare validate/status/init route to project subcommands
	case "validate":
		os.Exit(runValidate(os.Args[2:]))
	case "status":
		os.Exit(runStatus(os.Args[2:]))
	case "init":
		os.Exit(runInit(os.Args[2:]))

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: folio <command> [flags]

Command groups:
  project    Per-project commands (cwd-based)
  home       Repository-level commands (FOLIO_HOME)

Top-level commands:
  setup      Check folio dependencies
  version    Show version

Run 'folio <group> --help' for details.
`)
}

func printProjectUsage() {
	fmt.Fprintf(os.Stderr, `Usage: folio project <command> [flags]

Commands:
  validate   Validate folio.yml structure
  status     Derive and display target state
  init       Bootstrap a new folio.yml
`)
}

func printHomeUsage() {
	fmt.Fprintf(os.Stderr, `Usage: folio home <command> [flags]

Commands:
  init       Scaffold FOLIO_HOME directory
  validate   Structural checks (folio.yml in leaves, date prefixes)
  list       Show grouped summary of all folios
  push       git add + commit (+ push if remote) — requires -m
  pull       git pull
  archive    Move active path to archive with date prefix
  activate   Move archive path to active, strip date prefix
`)
}

// =============================================================================
// Project commands (cwd-based)
// =============================================================================

func runValidate(args []string) int {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	folioPath := fs.String("folio", "./folio.yml", "Path to folio.yml")
	jsonMode := fs.Bool("json", false, "Machine-readable JSON output")
	noColor := fs.Bool("no-color", false, "Disable colored output")
	fs.Parse(args)

	color := colorEnabled(*noColor)

	if _, err := os.Stat(*folioPath); os.IsNotExist(err) {
		if *jsonMode {
			fmt.Printf(`{"valid":false,"errors":["folio.yml not found at %s"],"warnings":[]}`, *folioPath)
			fmt.Println()
		} else {
			fmt.Fprintf(os.Stderr, "\033[0;31mError:\033[0m folio.yml not found at %s\n", *folioPath)
		}
		return 2
	}

	f, err := config.Load(*folioPath)
	if err != nil {
		if *jsonMode {
			fmt.Println(`{"valid":false,"errors":["folio.yml is not valid YAML"],"warnings":[]}`)
		} else {
			fmt.Fprintf(os.Stderr, "\033[0;31mError:\033[0m %s\n", err)
		}
		return 2
	}

	folioDir := filepath.Dir(*folioPath)
	result := validate.Validate(f, folioDir)

	if *jsonMode {
		output.PrintValidateJSON(os.Stdout, result)
	} else {
		output.PrintValidateTerminal(os.Stdout, result, *folioPath, color)
	}

	if !result.Valid {
		return 2
	}
	return 0
}

func runStatus(args []string) int {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	folioPath := fs.String("folio", "./folio.yml", "Path to folio.yml")
	jsonMode := fs.Bool("json", false, "Machine-readable JSON output")
	noColor := fs.Bool("no-color", false, "Disable colored output")
	fs.Parse(args)

	color := colorEnabled(*noColor)

	if _, err := os.Stat(*folioPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: folio.yml not found at %s\n", *folioPath)
		return 1
	}

	f, err := config.Load(*folioPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return 1
	}

	folioDir := filepath.Dir(*folioPath)
	ps, causedBy := status.DeriveWithDAG(f, folioDir)

	if *jsonMode {
		output.PrintStatusJSON(os.Stdout, ps)
	} else {
		output.PrintStatusTerminal(os.Stdout, ps, causedBy, color)
	}

	return 0
}

func runInit(args []string) int {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	name := fs.String("name", "", "Project name")
	fs.Parse(args)

	if _, err := os.Stat("folio.yml"); err == nil {
		fmt.Fprintf(os.Stderr, "\033[0;31mError:\033[0m folio.yml already exists in %s\n", mustGetwd())
		return 1
	}

	if *name == "" {
		fmt.Fprintf(os.Stderr, "\033[0;31mError:\033[0m --name is required\n")
		return 1
	}

	content := fmt.Sprintf(`schema: 1
project: "%s"

sources: []

targets: {}

tasks: []

pending: []
`, *name)

	if err := os.WriteFile("folio.yml", []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "\033[0;31mError:\033[0m %s\n", err)
		return 1
	}

	fmt.Printf("\033[0;32m✓\033[0m Created folio.yml for \033[1m%s\033[0m\n", *name)
	return 0
}

// =============================================================================
// Home commands (FOLIO_HOME)
// =============================================================================

func mustResolveHome() string {
	dir, err := home.Dir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[0;31mError:\033[0m %s\n", err)
		os.Exit(1)
	}
	return dir
}

func runHomeInit(args []string) int {
	dir := mustResolveHome()

	if err := home.Init(dir); err != nil {
		fmt.Fprintf(os.Stderr, "\033[0;31mError:\033[0m %s\n", err)
		return 1
	}

	fmt.Printf("\033[0;32m✓\033[0m Initialized FOLIO_HOME at %s\n", dir)
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
			fmt.Printf("\033[0;32m✓\033[0m FOLIO_HOME structure is valid (%s)\n", dir)
		} else {
			fmt.Printf("FOLIO_HOME structure is valid (%s)\n", dir)
		}
		return 0
	}

	if color {
		fmt.Fprintf(os.Stderr, "\033[0;31mErrors:\033[0m\n")
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
		fmt.Fprintf(os.Stderr, "\033[0;31mError:\033[0m %s\n", err)
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
			fmt.Printf("\033[1mActive\033[0m (%d)\n", len(active))
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
			fmt.Printf("\033[1mArchived\033[0m (%d)\n", len(archived))
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
		fmt.Printf("\033[2m%s\033[0m\n", header)
		fmt.Printf("\033[2m%s\033[0m\n", sep)
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
	fs.Parse(args)

	// Allow positional args as message for convenience: folio home push "my message"
	if fs.NArg() > 0 {
		*msg = strings.Join(fs.Args(), " ")
	}

	if *msg == "" {
		fmt.Fprintf(os.Stderr, "Error: commit message required (-m or positional arg)\n")
		fmt.Fprintf(os.Stderr, "  Format: type(scope): description\n")
		fmt.Fprintf(os.Stderr, "  Types:  feat fix docs refactor test chore style perf auto\n")
		return 1
	}

	dir := mustResolveHome()

	if err := repo.Push(dir, *msg); err != nil {
		if errors.Is(err, repo.ErrNothingToCommit) {
			fmt.Println("Nothing to commit (working tree clean)")
			return 0
		}
		if errors.Is(err, repo.ErrInvalidCommitMessage) {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err)
			return 1
		}
		fmt.Fprintf(os.Stderr, "\033[0;31mError:\033[0m %s\n", err)
		return 1
	}

	fmt.Printf("\033[0;32m✓\033[0m Committed and pushed\n")
	return 0
}

func runHomePull(args []string) int {
	dir := mustResolveHome()

	if err := repo.Pull(dir); err != nil {
		fmt.Fprintf(os.Stderr, "\033[0;31mError:\033[0m %s\n", err)
		return 1
	}

	fmt.Printf("\033[0;32m✓\033[0m Pulled latest\n")
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
		fmt.Fprintf(os.Stderr, "\033[0;31mError:\033[0m %s\n", err)
		return 1
	}

	fmt.Printf("\033[0;32m✓\033[0m Archived active/%s\n", relPath)
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
		fmt.Fprintf(os.Stderr, "\033[0;31mError:\033[0m %s\n", err)
		return 1
	}

	fmt.Printf("\033[0;32m✓\033[0m Activated archive/%s\n", relPath)
	return 0
}

// =============================================================================
// Top-level commands
// =============================================================================

func runSetup(args []string) int {
	fs := flag.NewFlagSet("setup", flag.ExitOnError)
	checkMode := fs.Bool("check", false, "Silent mode: exit 0 if OK, exit 1 if missing")
	fs.Parse(args)

	folioBin, err := os.Executable()
	if err != nil {
		folioBin = "folio"
	}

	if !*checkMode {
		fmt.Printf("\033[0;32m✓\033[0m folio %s (%s)\n", version, folioBin)
		fmt.Printf("\n\033[0;32m\033[1mAll dependencies satisfied.\033[0m\n")
	}
	return 0
}

// =============================================================================
// Helpers
// =============================================================================

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func colorEnabled(noColorFlag bool) bool {
	if noColorFlag {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return isTerminal()
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
