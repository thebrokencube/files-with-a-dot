package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/health"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/list"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runHealth(args []string) int {
	fs := dendrik.NewFlagSet("health")
	folioPath := fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
	noColor := fs.BoolLong("no-color", "Disable colored output")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if !resolveOrDie(folioPath) {
		return dendrik.ExitUserError
	}

	color := dendrik.ColorEnabled(*noColor)

	f, err := config.Load(*folioPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return dendrik.ExitUserError
	}

	folioDir := filepath.Dir(*folioPath)
	report := health.Analyze(f, folioDir)
	printHealthReport(report, color)

	return 0 // always exit 0 (advisory)
}

func runHomeHealth(args []string) int {
	fs := dendrik.NewFlagSet("home health")
	noColor := fs.BoolLong("no-color", "Disable colored output")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	color := dendrik.ColorEnabled(*noColor)
	homeDir := mustResolveHome()

	entries, err := list.Scan(homeDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return dendrik.ExitUserError
	}

	active := filterEntries(entries, "active")
	if len(active) == 0 {
		fmt.Println("No active folios found.")
		return 0
	}

	for i, entry := range active {
		folioYml := filepath.Join(homeDir, entry.Section, entry.Path, "folio.yml")
		f, err := config.Load(folioYml)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Skipping %s: %s\n", entry.Path, err)
			continue
		}
		folioDir := filepath.Dir(folioYml)
		report := health.Analyze(f, folioDir)
		printHealthReport(report, color)
		if i < len(active)-1 {
			fmt.Println()
		}
	}

	return 0
}

func printHealthReport(r *health.Report, color bool) {
	// Header line
	grade := r.Grade
	if color {
		gradeColor := output.Green
		if grade == "Needs Attention" {
			gradeColor = output.Yellow
		}
		fmt.Printf("%s%-40s%s Health: %s%s%s\n",
			output.Bold, r.Project, output.Reset,
			gradeColor, grade, output.Reset)
	} else {
		fmt.Printf("%-40s Health: %s\n", r.Project, grade)
	}

	// Observations line (most actionable)
	if r.Observations.Active > 0 {
		fmt.Printf("  Observations  %d active\n", r.Observations.Active)
	}
	for _, w := range r.Observations.LintWarnings {
		if color {
			fmt.Printf("  %s⚠ %s%s\n", output.Yellow, w, output.Reset)
		} else {
			fmt.Printf("  ! %s\n", w)
		}
	}

	// Work line
	if r.Work.Active > 0 || r.Work.Archived > 0 {
		fmt.Printf("  Work     %d active, %d archived\n", r.Work.Active, r.Work.Archived)
	}

	// Reference line
	totalRef := r.TotalReferenceFiles()
	untypedCount := len(r.Untyped)
	if totalRef > 0 || untypedCount > 0 {
		refLine := fmt.Sprintf("  Reference   %d files", totalRef+untypedCount)
		if untypedCount > 0 {
			refLine += fmt.Sprintf(" (%d untyped)", untypedCount)
		}
		fmt.Println(refLine)

		// Type breakdown
		if len(r.Reference) > 0 {
			printTypeBreakdown(r.Reference)
		}
	}

	// Design colocation line
	if r.Design.Total > 0 {
		fmt.Printf("  Designs   %d total, %d colocated, %d orphaned\n", r.Design.Total, r.Design.Colocated, r.Design.Orphaned)
	}

	// Retro line
	if r.Retro.Total > 0 {
		fmt.Printf("  Retros    %d total, %d colocated, %d orphaned\n", r.Retro.Total, r.Retro.Colocated, r.Retro.Orphaned)
	}

	// Naming issues
	if len(r.Naming) > 0 {
		fmt.Printf("  Naming   %d files without date prefix\n", len(r.Naming))
	}

	// Warnings
	if untypedCount > 0 {
		if color {
			fmt.Printf("  %s⚠ %d files in flat reference/ — need migration to type directories%s\n",
				output.Yellow, untypedCount, output.Reset)
		} else {
			fmt.Printf("  ! %d files in flat reference/ — need migration to type directories\n", untypedCount)
		}
	}
	if len(r.Unrecognized) > 0 {
		if color {
			fmt.Printf("  %s⚠ Unrecognized directories in reference/: %s (rename to a recognized type)%s\n",
				output.Yellow, strings.Join(r.Unrecognized, ", "), output.Reset)
		} else {
			fmt.Printf("  ! Unrecognized directories in reference/: %s (rename to a recognized type)\n",
				strings.Join(r.Unrecognized, ", "))
		}
	}
}

func printTypeBreakdown(ref map[string]int) {
	// Sort types by count descending, then name ascending
	type typeCount struct {
		name  string
		count int
	}
	var types []typeCount
	for name, count := range ref {
		types = append(types, typeCount{name, count})
	}
	sort.Slice(types, func(i, j int) bool {
		if types[i].count != types[j].count {
			return types[i].count > types[j].count
		}
		return types[i].name < types[j].name
	})

	// Print in rows of 3 columns
	line := "   "
	for i, tc := range types {
		line += fmt.Sprintf(" %-12s %d", tc.name+"/", tc.count)
		if (i+1)%3 == 0 && i+1 < len(types) {
			fmt.Println(line)
			line = "   "
		}
	}
	if strings.TrimSpace(line) != "" {
		fmt.Println(line)
	}
}
