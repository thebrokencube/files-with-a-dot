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
			fmt.Fprintln(os.Stderr, output.Errf("folio.yml not found at %s", *folioPath))
		}
		return 2
	}

	f, err := config.Load(*folioPath)
	if err != nil {
		if *jsonMode {
			fmt.Println(`{"valid":false,"errors":["folio.yml is not valid YAML"],"warnings":[]}`)
		} else {
			fmt.Fprintln(os.Stderr, output.Errf("%s", err))
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
		fmt.Fprintln(os.Stderr, output.Errf("folio.yml not found at %s", *folioPath))
		return 1
	}

	f, err := config.Load(*folioPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
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
		fmt.Fprintln(os.Stderr, output.Errf("folio.yml already exists in %s", mustGetwd()))
		return 1
	}

	if *name == "" {
		fmt.Fprintln(os.Stderr, output.Errf("--name is required"))
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
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	fmt.Println(output.Successf("Created folio.yml for %s%s%s", output.Bold, *name, output.Reset))
	return 0
}
