package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

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
	if strings.HasPrefix(filepath.Base(dir), "folio-ws-") {
		fmt.Fprintf(os.Stderr, "%sfolio workspace: %s%s\n", pal.Dim, dir, pal.Reset)
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

	var pushErr error
	if *folioName != "" {
		// Scoped push: resolve the target folio and validate only it. A scoped
		// push must isolate the caller from unrelated validation debt elsewhere
		// in the tree — validating everything would defeat the point of -f.
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
		if errs := validateProject(dir, *match); len(errs) > 0 {
			printValidationErrors(pal, errs)
			return dendrik.ExitUserError
		}
		pushErr = repo.PushScoped(dir, *msg, []string{match.Section + "/" + match.Path})
	} else {
		// Whole-tree push: validate every active folio plus home structure.
		if errs := validateActiveProjects(dir); len(errs) > 0 {
			printValidationErrors(pal, errs)
			return dendrik.ExitUserError
		}
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
		if errors.Is(pushErr, repo.ErrConflict) {
			fmt.Fprintln(os.Stderr, pal.Errf("%s", pushErr))
			fmt.Fprintf(os.Stderr, "  Resolve conflicts, then retry: folio home push -m \"...\"\n")
			return dendrik.ExitExternalErr
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

var statDatePrefixRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-`)

func runHomeStats(args []string) int {
	fs := dendrik.NewFlagSet("home stats")
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

	// Total commits
	out, err := repo.GitOutput(dir, "rev-list", "--count", "HEAD")
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("git rev-list: %s", err))
		return dendrik.ExitExternalErr
	}
	totalCommits, _ := strconv.Atoi(strings.TrimSpace(out))

	if totalCommits == 0 {
		fmt.Println("No commits yet.")
		return dendrik.ExitOK
	}

	// First commit date (via root commit)
	out, err = repo.GitOutput(dir, "rev-list", "--max-parents=0", "HEAD")
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("git rev-list: %s", err))
		return dendrik.ExitExternalErr
	}
	rootHash := strings.TrimSpace(out)
	out, err = repo.GitOutput(dir, "log", "--format=%aI", "-1", rootHash)
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("git log: %s", err))
		return dendrik.ExitExternalErr
	}
	firstDate, _ := time.Parse(time.RFC3339, strings.TrimSpace(out))

	// Latest commit date
	out, err = repo.GitOutput(dir, "log", "--format=%aI", "--max-count=1")
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("git log: %s", err))
		return dendrik.ExitExternalErr
	}
	latestDate, _ := time.Parse(time.RFC3339, strings.TrimSpace(out))

	elapsedDays := int(latestDate.Sub(firstDate).Hours() / 24)

	// Summary
	if color {
		fmt.Printf("%sRepository%s  %s\n", pal.Bold, pal.Reset, dir)
		fmt.Printf("%sCommits%s     %d  (%s → %s, %d days)\n", pal.Bold, pal.Reset,
			totalCommits, firstDate.Format("2006-01-02"), latestDate.Format("2006-01-02"), elapsedDays)
	} else {
		fmt.Printf("Repository  %s\n", dir)
		fmt.Printf("Commits     %d  (%s → %s, %d days)\n",
			totalCommits, firstDate.Format("2006-01-02"), latestDate.Format("2006-01-02"), elapsedDays)
	}

	if elapsedDays > 0 {
		perDay := float64(totalCommits) / float64(elapsedDays)
		perWeek := perDay * 7
		fmt.Printf("Rate        %.1f/day  %.1f/week\n", perDay, perWeek)
	} else {
		fmt.Printf("Rate        —\n")
	}

	// Per-folio breakdown
	entries, err := list.Scan(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("scan: %s", err))
		return dendrik.ExitExternalErr
	}

	if len(entries) == 0 {
		return dendrik.ExitOK
	}

	type folioStat struct {
		Path      string
		Section   string
		Commits   int
		FirstDate time.Time
		PerDay    float64
		PerWeek   float64
	}

	var stats []folioStat
	for _, e := range entries {
		// Build pathspecs for git
		var pathspecs []string
		sectionPath := e.Section + "/" + e.Path
		pathspecs = append(pathspecs, sectionPath)

		// For archived folios, also include the original active path
		if e.Section == "archive" {
			leaf := filepath.Base(e.Path)
			if statDatePrefixRe.MatchString(leaf) {
				origLeaf := leaf[11:]
				parent := filepath.Dir(e.Path)
				var origPath string
				if parent == "." {
					origPath = "active/" + origLeaf
				} else {
					origPath = "active/" + parent + "/" + origLeaf
				}
				pathspecs = append(pathspecs, origPath)
			}
		}

		// Commit count
		gitArgs := append([]string{"rev-list", "--count", "HEAD", "--"}, pathspecs...)
		out, err := repo.GitOutput(dir, gitArgs...)
		if err != nil {
			continue
		}
		count, _ := strconv.Atoi(strings.TrimSpace(out))

		// First commit date for this folio (last line of git log output = oldest)
		gitArgs = append([]string{"log", "--format=%aI", "--"}, pathspecs...)
		out, err = repo.GitOutput(dir, gitArgs...)
		if err != nil || strings.TrimSpace(out) == "" {
			stats = append(stats, folioStat{Path: e.Path, Section: e.Section, Commits: count})
			continue
		}
		lines := strings.Split(strings.TrimSpace(out), "\n")
		fd, _ := time.Parse(time.RFC3339, lines[len(lines)-1])

		fs := folioStat{Path: e.Path, Section: e.Section, Commits: count, FirstDate: fd}
		folioDays := int(latestDate.Sub(fd).Hours() / 24)
		if folioDays > 0 {
			fs.PerDay = float64(count) / float64(folioDays)
			fs.PerWeek = fs.PerDay * 7
		}
		stats = append(stats, fs)
	}

	// Sort by commit count descending
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Commits > stats[j].Commits
	})

	fmt.Printf("Folios      %d  (avg %.1f commits each)\n", len(stats), float64(totalCommits)/float64(len(stats)))
	fmt.Println()

	// Column widths
	pathW := 4
	for _, s := range stats {
		if len(s.Path) > pathW {
			pathW = len(s.Path)
		}
	}

	header := fmt.Sprintf("  %-*s  %-8s  %7s  %8s  %8s", pathW, "Path", "Section", "Commits", "/day", "/week")
	sep := fmt.Sprintf("  %s  %s  %s  %s  %s", strings.Repeat("-", pathW), "--------", "-------", "--------", "--------")
	if color {
		fmt.Printf("%s%s%s\n", pal.Dim, header, pal.Reset)
		fmt.Printf("%s%s%s\n", pal.Dim, sep, pal.Reset)
	} else {
		fmt.Println(header)
		fmt.Println(sep)
	}

	for _, s := range stats {
		rateDay := "—"
		rateWeek := "—"
		if s.PerDay > 0 {
			rateDay = fmt.Sprintf("%.1f", s.PerDay)
			rateWeek = fmt.Sprintf("%.1f", s.PerWeek)
		}
		fmt.Printf("  %-*s  %-8s  %7d  %8s  %8s\n", pathW, s.Path, s.Section, s.Commits, rateDay, rateWeek)
	}

	return dendrik.ExitOK
}

func runHomeWorkspace(args []string) int {
	pal := dendrik.NewPalette(true)

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: folio home workspace <create|list|cleanup>\n")
		return dendrik.ExitUserError
	}

	dir, code := resolveHomeOrFail()
	if code != dendrik.ExitOK {
		return code
	}

	// Workspace commands require jj
	if !repo.IsJJ(dir) {
		fmt.Fprintln(os.Stderr, pal.Errf("workspace requires jj — no .jj directory in %s", dir))
		return dendrik.ExitUserError
	}

	switch args[0] {
	case "create":
		return runWorkspaceCreate(dir, pal)
	case "list":
		return runWorkspaceList(dir, pal)
	case "cleanup":
		return runWorkspaceCleanup(dir, args[1:], pal)
	default:
		fmt.Fprintf(os.Stderr, "Unknown workspace command: %s\n", args[0])
		return dendrik.ExitUserError
	}
}

func runWorkspaceCreate(homeDir string, pal dendrik.Palette) int {
	wsID := fmt.Sprintf("folio-ws-%d-%d", time.Now().Unix(), os.Getpid())
	wsDir := filepath.Join("/tmp", wsID)

	cmd := exec.Command("jj", "--no-pager", "workspace", "add", wsDir, "-r", "main", "-R", homeDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("jj workspace add: %s", err))
		return dendrik.ExitExternalErr
	}

	fmt.Println(wsDir)
	return dendrik.ExitOK
}

func runWorkspaceList(homeDir string, pal dendrik.Palette) int {
	cmd := exec.Command("jj", "--no-pager", "workspace", "list", "-R", homeDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("jj workspace list: %s", err))
		return dendrik.ExitExternalErr
	}
	return dendrik.ExitOK
}

func runWorkspaceCleanup(homeDir string, args []string, pal dendrik.Palette) int {
	// Determine workspace path: from args, or from FOLIO_HOME if it's a workspace
	var wsDir string
	if len(args) > 0 {
		wsDir = args[0]
	} else {
		// Use current FOLIO_HOME if it looks like a workspace
		folio := os.Getenv("FOLIO_HOME")
		if folio == "" || !strings.HasPrefix(filepath.Base(folio), "folio-ws-") {
			fmt.Fprintln(os.Stderr, pal.Errf("specify workspace path or set FOLIO_HOME to a workspace"))
			return dendrik.ExitUserError
		}
		wsDir = folio
	}

	wsName := filepath.Base(wsDir)

	// Check for unpushed changes
	cmd := exec.Command("jj", "--no-pager", "log", "-r", "@", "--no-graph",
		"-T", `if(empty, "empty", "changed")`, "-R", wsDir)
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("cannot check workspace status: %s", err))
		return dendrik.ExitExternalErr
	}
	if strings.TrimSpace(string(out)) != "empty" {
		fmt.Fprintln(os.Stderr, pal.Errf("workspace has unpushed changes — run 'folio home push' first"))
		return dendrik.ExitUserError
	}

	// Forget the workspace
	forgetCmd := exec.Command("jj", "--no-pager", "--quiet", "workspace", "forget", wsName, "-R", homeDir)
	forgetCmd.Stdout = os.Stdout
	forgetCmd.Stderr = os.Stderr
	if err := forgetCmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("jj workspace forget: %s", err))
		return dendrik.ExitExternalErr
	}

	// Remove directory
	if err := os.RemoveAll(wsDir); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("rm workspace dir: %s", err))
		return dendrik.ExitExternalErr
	}

	fmt.Println(pal.Successf("Cleaned up workspace %s", wsName))
	return dendrik.ExitOK
}

// validateActiveProjects loads and validates every folio.yml in active/
// and the vault structure. Returns a list of human-readable errors (empty on success).
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
		errs = append(errs, validateProject(homeDir, e)...)
	}

	// Validate home + vault structure
	errs = append(errs, home.Validate(homeDir)...)

	return errs
}

// validateProject validates a single folio: its folio.yml structure and its
// observations. Returns human-readable errors (empty on success).
func validateProject(homeDir string, e list.Entry) []string {
	var errs []string
	ymlPath := filepath.Join(homeDir, e.Section, e.Path, "folio.yml")
	f, err := config.Load(ymlPath)
	if err != nil {
		return []string{fmt.Sprintf("%s: %s", e.Path, err)}
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
	return errs
}

// printValidationErrors writes a validation failure block to stderr.
func printValidationErrors(pal dendrik.Palette, errs []string) {
	fmt.Fprintln(os.Stderr, pal.Errf("validation failed — fix before pushing:"))
	for _, e := range errs {
		fmt.Fprintf(os.Stderr, "  - %s\n", e)
	}
}
