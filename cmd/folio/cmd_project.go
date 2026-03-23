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
		return dendrik.ExitUserError
	}

	color := dendrik.ColorEnabled(*noColor)
	pal := dendrik.NewPalette(color)

	if _, err := os.Stat(*folioPath); os.IsNotExist(err) {
		if *jsonMode {
			dendrik.WriteError(os.Stdout, fmt.Sprintf("folio.yml not found at %s", *folioPath), "")
		} else {
			fmt.Fprintln(os.Stderr, pal.Errf("folio.yml not found at %s", *folioPath))
		}
		return dendrik.ExitExternalErr
	}

	f, err := config.Load(*folioPath)
	if err != nil {
		if *jsonMode {
			dendrik.WriteError(os.Stdout, "folio.yml is not valid YAML", "")
		} else {
			fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		}
		return dendrik.ExitExternalErr
	}

	folioDir := filepath.Dir(*folioPath)
	result := validate.Validate(f, folioDir)

	if *jsonMode {
		output.PrintValidateJSON(os.Stdout, result)
	} else {
		output.PrintValidateTerminal(os.Stdout, result, *folioPath, color)
	}

	if !result.Valid {
		return dendrik.ExitExternalErr
	}
	return dendrik.ExitOK
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
		return dendrik.ExitUserError
	}

	color := dendrik.ColorEnabled(*noColor)
	pal := dendrik.NewPalette(color)

	if _, err := os.Stat(*folioPath); os.IsNotExist(err) {
		if *jsonMode {
			dendrik.WriteError(os.Stdout, fmt.Sprintf("folio.yml not found at %s", *folioPath), "")
		} else {
			fmt.Fprintln(os.Stderr, pal.Errf("folio.yml not found at %s", *folioPath))
		}
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

	folioDir := filepath.Dir(*folioPath)
	ps, causedBy := status.DeriveWithDAG(f, folioDir)

	if *jsonMode {
		output.PrintStatusJSON(os.Stdout, ps)
	} else {
		output.PrintStatusTerminal(os.Stdout, ps, causedBy, color)
	}

	return dendrik.ExitOK
}

func runInit(args []string) int {
	pal := dendrik.NewPalette(true)
	fs := dendrik.NewFlagSet("init")
	name := fs.String('n', "name", "", "Project name")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if _, err := os.Stat("folio.yml"); err == nil {
		fmt.Fprintln(os.Stderr, pal.Errf("folio.yml already exists in %s", mustGetwd()))
		return dendrik.ExitUserError
	}

	if *name == "" {
		fmt.Fprintln(os.Stderr, pal.Errf("--name is required"))
		return dendrik.ExitUserError
	}

	content := fmt.Sprintf(`schema: 2
project: "%s"

sources: []

targets: {}

observations: []
`, *name)

	if err := os.WriteFile("folio.yml", []byte(content), 0644); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}

	fmt.Println(pal.Successf("Created folio.yml for %s%s%s", pal.Bold, *name, pal.Reset))
	return dendrik.ExitOK
}
