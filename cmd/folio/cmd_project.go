package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/status"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/validate"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runValidate(folioPath string, jsonMode, noColor bool) int {
	ctx, err := resolveContext(folioPath, contextReadOnly)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}
	folioPath = ctx.FolioPath

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

	result := validate.Validate(f, ctx, validate.ReadOnly)

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
	ctx, err := resolveContext(folioPath, contextReadOnly)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}
	folioPath = ctx.FolioPath

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

	ps, causedBy, _ := status.DeriveWithDAGWithContext(f, ctx)

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

	// Prefer the selected Folio work root when it has an active/ directory;
	// otherwise preserve the legacy current-directory fallback.
	targetPath := "folio.yml"
	workRoot, err := resolveFolioRootForInit()
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}
	activeDir := filepath.Join(workRoot, "active")
	if fi, statErr := os.Stat(activeDir); statErr == nil && fi.IsDir() {
		slug := pathFlag
		if slug == "" {
			slug = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
		}
		targetPath = filepath.Join(activeDir, slug, "folio.yml")
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
