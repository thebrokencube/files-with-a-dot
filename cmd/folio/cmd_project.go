package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/home"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/status"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/validate"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runValidate(folioPath string, jsonMode, noColor bool) int {
	if !resolveOrDie(&folioPath) {
		return dendrik.ExitUserError
	}

	color := dendrik.ColorEnabled(noColor)
	pal := dendrik.NewPalette(color)

	if _, err := os.Stat(folioPath); os.IsNotExist(err) {
		if jsonMode {
			dendrik.WriteError(os.Stdout, fmt.Sprintf("folio.yml not found at %s", folioPath), "")
		} else {
			fmt.Fprintln(os.Stderr, pal.Errf("folio.yml not found at %s", folioPath))
		}
		return dendrik.ExitExternalErr
	}

	f, err := config.Load(folioPath)
	if err != nil {
		if jsonMode {
			dendrik.WriteError(os.Stdout, "folio.yml is not valid YAML", "")
		} else {
			fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		}
		return dendrik.ExitExternalErr
	}

	folioDir := filepath.Dir(folioPath)
	result := validate.Validate(f, folioDir)

	if jsonMode {
		output.PrintValidateJSON(os.Stdout, result)
	} else {
		output.PrintValidateTerminal(os.Stdout, result, folioPath, color)
	}

	if !result.Valid {
		return dendrik.ExitExternalErr
	}
	return dendrik.ExitOK
}

func runStatus(folioPath string, jsonMode, noColor bool) int {
	if !resolveOrDie(&folioPath) {
		return dendrik.ExitUserError
	}

	color := dendrik.ColorEnabled(noColor)
	pal := dendrik.NewPalette(color)

	if _, err := os.Stat(folioPath); os.IsNotExist(err) {
		if jsonMode {
			dendrik.WriteError(os.Stdout, fmt.Sprintf("folio.yml not found at %s", folioPath), "")
		} else {
			fmt.Fprintln(os.Stderr, pal.Errf("folio.yml not found at %s", folioPath))
		}
		return dendrik.ExitUserError
	}

	f, err := config.Load(folioPath)
	if err != nil {
		if jsonMode {
			dendrik.WriteError(os.Stdout, fmt.Sprintf("%s", err), "")
		} else {
			fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		}
		return dendrik.ExitUserError
	}

	folioDir := filepath.Dir(folioPath)
	ps, causedBy := status.DeriveWithDAG(f, folioDir)

	if jsonMode {
		output.PrintStatusJSON(os.Stdout, ps)
	} else {
		output.PrintStatusTerminal(os.Stdout, ps, causedBy, color)
	}

	return dendrik.ExitOK
}

func runInit(name, pathFlag string) int {
	pal := dendrik.NewPalette(true)

	if name == "" {
		fmt.Fprintln(os.Stderr, pal.Errf("--name is required"))
		return dendrik.ExitUserError
	}

	// Determine target path: prefer FOLIO_HOME/active/<slug>/ if initialized,
	// otherwise fall back to current working directory.
	targetPath := "folio.yml"
	if homeDir, err := home.Dir(); err == nil {
		activeDir := filepath.Join(homeDir, "active")
		if fi, err := os.Stat(activeDir); err == nil && fi.IsDir() {
			slug := pathFlag
			if slug == "" {
				slug = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
			}
			targetPath = filepath.Join(activeDir, slug, "folio.yml")
		}
	}

	if _, err := os.Stat(targetPath); err == nil {
		fmt.Fprintln(os.Stderr, pal.Errf("folio.yml already exists at %s", targetPath))
		return dendrik.ExitUserError
	}

	content := fmt.Sprintf(`schema: 3
project: "%s"

sources: []

targets: {}

observations: []
`, name)

	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}

	if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}

	fmt.Println(pal.Successf("Created folio.yml for %s%s%s", pal.Bold, name, pal.Reset))
	fmt.Printf("  %s\n", targetPath)
	return dendrik.ExitOK
}
