package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/status"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/validate"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runValidate(args []string) int {
	fs := dendrik.NewFlagSet("validate")
	folioPath := fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
	jsonMode := fs.Bool('j', "json", "Machine-readable JSON output")
	noColor := fs.BoolLong("no-color", "Disable colored output")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if !resolveOrDie(folioPath) {
		return 1
	}

	color := dendrik.ColorEnabled(*noColor)

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
	fs := dendrik.NewFlagSet("status")
	folioPath := fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
	jsonMode := fs.Bool('j', "json", "Machine-readable JSON output")
	noColor := fs.BoolLong("no-color", "Disable colored output")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if !resolveOrDie(folioPath) {
		return 1
	}

	color := dendrik.ColorEnabled(*noColor)

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
	fs := dendrik.NewFlagSet("init")
	name := fs.String('n', "name", "", "Project name")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if _, err := os.Stat("folio.yml"); err == nil {
		fmt.Fprintln(os.Stderr, output.Errf("folio.yml already exists in %s", mustGetwd()))
		return 1
	}

	if *name == "" {
		fmt.Fprintln(os.Stderr, output.Errf("--name is required"))
		return 1
	}

	content := fmt.Sprintf(`schema: 2
project: "%s"

sources: []

targets: {}

observations: []
`, *name)

	if err := os.WriteFile("folio.yml", []byte(content), 0644); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	fmt.Println(output.Successf("Created folio.yml for %s%s%s", output.Bold, *name, output.Reset))
	return 0
}
