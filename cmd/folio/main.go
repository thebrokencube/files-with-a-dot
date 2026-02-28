package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/status"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/validate"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "validate":
		os.Exit(runValidate(os.Args[2:]))
	case "status":
		os.Exit(runStatus(os.Args[2:]))
	case "init":
		os.Exit(runInit(os.Args[2:]))
	case "setup":
		os.Exit(runSetup(os.Args[2:]))
	case "version":
		fmt.Printf("folio %s\n", version)
		os.Exit(0)
	case "--help", "-h", "help":
		printUsage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: folio <command> [flags]

Commands:
  validate   Validate folio.yml structure
  status     Derive and display target state
  init       Bootstrap a new folio.yml
  setup      Check folio dependencies
  version    Show version

Run 'folio <command> --help' for details.
`)
}

// --- validate ---

func runValidate(args []string) int {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	folioPath := fs.String("folio", "./folio.yml", "Path to folio.yml")
	jsonMode := fs.Bool("json", false, "Machine-readable JSON output")
	noColor := fs.Bool("no-color", false, "Disable colored output")
	fs.Parse(args)

	color := colorEnabled(*noColor)

	// Check file exists
	if _, err := os.Stat(*folioPath); os.IsNotExist(err) {
		if *jsonMode {
			fmt.Printf(`{"valid":false,"errors":["folio.yml not found at %s"],"warnings":[]}`, *folioPath)
			fmt.Println()
		} else {
			fmt.Fprintf(os.Stderr, "\033[0;31mError:\033[0m folio.yml not found at %s\n", *folioPath)
		}
		return 2
	}

	// Parse
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

// --- status ---

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

// --- init ---

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

// --- setup ---

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
