package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/home"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/list"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/move"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/observe"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/repo"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/validate"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func resolveHomeOrFail() (string, int) {
	pal := dendrik.NewPalette(true)
	dir, err := home.Dir()
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return "", dendrik.ExitUserError
	}
	return dir, dendrik.ExitOK
}

func runHomeInit(args []string) int {
	pal := dendrik.NewPalette(true)
	dir, code := resolveHomeOrFail()
	if code != dendrik.ExitOK {
		return code
	}

	if err := home.Init(dir); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}

	fmt.Println(pal.Successf("Initialized FOLIO_HOME at %s", dir))
	return dendrik.ExitOK
}

func runHomeValidate(args []string) int {
	fs := dendrik.NewFlagSet("home validate")
	noColor := fs.BoolLong("no-color", "Disable colored output")
	if done, code := dendrik.ParseCheck(fs, args); done {
		return code
	}

	color := dendrik.ColorEnabled(*noColor)
	pal := dendrik.NewPalette(color)
	dir, code := resolveHomeOrFail()
	if code != dendrik.ExitOK {
		return code
	}

	errs := home.Validate(dir)

	if len(errs) == 0 {
		if color {
			fmt.Println(pal.Successf("FOLIO_HOME structure is valid (%s)", dir))
		} else {
			fmt.Printf("FOLIO_HOME structure is valid (%s)\n", dir)
		}
		return dendrik.ExitOK
	}

	if color {
		fmt.Fprintf(os.Stderr, "%sErrors:%s\n", pal.Red, pal.Reset)
	} else {
		fmt.Fprintf(os.Stderr, "Errors:\n")
	}
	for _, e := range errs {
		fmt.Fprintf(os.Stderr, "  - %s\n", e)
	}
	return dendrik.ExitExternalErr
}

func runHomeList(args []string) int {
	fs := dendrik.NewFlagSet("home list")
	jsonMode := fs.Bool('j', "json", "Machine-readable JSON output")
	noColor := fs.BoolLong("no-color", "Disable colored output")
	if done, code := dendrik.ParseCheck(fs, args); done {
		return code
	}

	color := dendrik.ColorEnabled(*noColor)
	pal := dendrik.NewPalette(color)
	dir, code := resolveHomeOrFail()
	if code != dendrik.ExitOK {
		return code
	}

	entries, err := list.Scan(dir)
	if err != nil {
		if *jsonMode {
			dendrik.WriteError(os.Stdout, fmt.Sprintf("%s", err), "")
		} else {
			fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		}
		return dendrik.ExitUserError
	}

	if *jsonMode {
		dendrik.WriteResult(os.Stdout, entries)
		return dendrik.ExitOK
	}

	if len(entries) == 0 {
		fmt.Println("No folios found.")
		return dendrik.ExitOK
	}

	// Group by section
	active := filterEntries(entries, "active")
	archived := filterEntries(entries, "archive")

	if len(active) > 0 {
		if color {
			fmt.Printf("%sActive%s (%d)\n", pal.Bold, pal.Reset, len(active))
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
			fmt.Printf("%sArchived%s (%d)\n", pal.Bold, pal.Reset, len(archived))
		} else {
			fmt.Printf("Archived (%d)\n", len(archived))
		}
		printEntryTable(archived, color)
	}

	return dendrik.ExitOK
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
	pal := dendrik.NewPalette(color)
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

	header := fmt.Sprintf("  %-*s  %-*s  %s  %s", pathW, "Path", projW, "Project", "Targets", "Observations")
	sep := fmt.Sprintf("  %s  %s  %s  %s", strings.Repeat("-", pathW), strings.Repeat("-", projW), "-------", "------------")

	if color {
		fmt.Printf("%s%s%s\n", pal.Dim, header, pal.Reset)
		fmt.Printf("%s%s%s\n", pal.Dim, sep, pal.Reset)
	} else {
		fmt.Println(header)
		fmt.Println(sep)
	}

	for _, e := range entries {
		fmt.Printf("  %-*s  %-*s  %7d  %12d\n", pathW, e.Path, projW, e.Project, e.Targets, e.Observations)
	}
}

func runHomePush(args []string) int {
	pal := dendrik.NewPalette(true)
	fs := dendrik.NewFlagSet("home push")
	msg := fs.String('m', "message", "", "Commit message: type(scope): description")
	folioName := fs.String('f', "folio", "", "Scope commit to a single folio (shortname or path)")
	all := fs.Bool('a', "all", "Stage all changes (current behavior, default)")
	if done, code := dendrik.ParseCheck(fs, args); done {
		return code
	}

	// Allow positional args as message for convenience: folio home push "my message"
	if len(fs.GetArgs()) > 0 {
		*msg = strings.Join(fs.GetArgs(), " ")
	}

	if *msg == "" {
		fmt.Fprintln(os.Stderr, pal.Errf("commit message required (-m or positional arg)"))
		fmt.Fprintf(os.Stderr, "  Format: type(scope): description\n")
		fmt.Fprintf(os.Stderr, "  Types:  feat fix docs refactor test chore style perf auto\n")
		return dendrik.ExitUserError
	}

	if *folioName != "" && *all {
		fmt.Fprintln(os.Stderr, pal.Errf("--folio and --all are mutually exclusive"))
		return dendrik.ExitUserError
	}

	dir, code := resolveHomeOrFail()
	if code != dendrik.ExitOK {
		return code
	}

	// Validate all active folio.yml files before committing
	if errs := validateActiveProjects(dir); len(errs) > 0 {
		fmt.Fprintln(os.Stderr, pal.Errf("validation failed — fix before pushing:"))
		for _, e := range errs {
			fmt.Fprintf(os.Stderr, "  - %s\n", e)
		}
		return dendrik.ExitUserError
	}

	var pushErr error
	if *folioName != "" {
		entries, err := list.Scan(dir)
		if err != nil {
			fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
			return dendrik.ExitUserError
		}
		var match *list.Entry
		for i, e := range entries {
			if e.Path == *folioName || e.Project == *folioName {
				match = &entries[i]
				break
			}
		}
		if match == nil {
			fmt.Fprintln(os.Stderr, pal.Errf("folio %q not found", *folioName))
			return dendrik.ExitUserError
		}
		pushErr = repo.PushScoped(dir, *msg, []string{match.Section + "/" + match.Path})
	} else {
		pushErr = repo.Push(dir, *msg)
	}

	if pushErr != nil {
		if errors.Is(pushErr, repo.ErrNothingToCommit) {
			fmt.Println("Nothing to commit (working tree clean)")
			return dendrik.ExitOK
		}
		if errors.Is(pushErr, repo.ErrInvalidCommitMessage) {
			fmt.Fprintln(os.Stderr, pal.Errf("%s", pushErr))
			return dendrik.ExitUserError
		}
		fmt.Fprintln(os.Stderr, pal.Errf("%s", pushErr))
		return dendrik.ExitUserError
	}

	fmt.Println(pal.Successf("Committed and pushed"))
	return dendrik.ExitOK
}

func runHomePull(args []string) int {
	pal := dendrik.NewPalette(true)
	dir, code := resolveHomeOrFail()
	if code != dendrik.ExitOK {
		return code
	}

	if err := repo.Pull(dir); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}

	fmt.Println(pal.Successf("Pulled latest"))
	return dendrik.ExitOK
}

func runHomeArchive(args []string) int {
	pal := dendrik.NewPalette(true)
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: folio home archive <path>\n")
		fmt.Fprintf(os.Stderr, "  Path is relative to active/, e.g., 'ben/my-project'\n")
		return dendrik.ExitUserError
	}

	dir, code := resolveHomeOrFail()
	if code != dendrik.ExitOK {
		return code
	}
	relPath := args[0]

	if err := move.Archive(dir, relPath); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}

	fmt.Println(pal.Successf("Archived active/%s", relPath))
	return dendrik.ExitOK
}

func runHomeActivate(args []string) int {
	pal := dendrik.NewPalette(true)
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: folio home activate <path>\n")
		fmt.Fprintf(os.Stderr, "  Path is relative to archive/, e.g., 'ben/2026-02-20-my-project'\n")
		return dendrik.ExitUserError
	}

	dir, code := resolveHomeOrFail()
	if code != dendrik.ExitOK {
		return code
	}
	relPath := args[0]

	if err := move.Activate(dir, relPath); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}

	fmt.Println(pal.Successf("Activated archive/%s", relPath))
	return dendrik.ExitOK
}

// validateActiveProjects loads and validates every folio.yml in active/.
// Returns a list of human-readable errors (empty on success).
func validateActiveProjects(homeDir string) []string {
	entries, err := list.Scan(homeDir)
	if err != nil {
		return []string{fmt.Sprintf("scan: %s", err)}
	}

	var errs []string
	for _, e := range entries {
		if e.Section != "active" {
			continue
		}
		ymlPath := filepath.Join(homeDir, "active", e.Path, "folio.yml")
		f, err := config.Load(ymlPath)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %s", e.Path, err))
			continue
		}
		folioDir := filepath.Dir(ymlPath)
		result := validate.Validate(f, folioDir)
		for _, ve := range result.Errors {
			errs = append(errs, fmt.Sprintf("%s: %s", e.Path, ve))
		}
		issues := observe.Lint(folioDir, f.Observations)
		for _, issue := range issues {
			errs = append(errs, fmt.Sprintf("%s: observation #%d: %s", e.Path, issue.Index, issue.Reason))
		}
	}
	return errs
}
