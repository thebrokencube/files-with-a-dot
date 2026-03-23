package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runPbcopy(args []string) int {
	pal := dendrik.NewPalette(true)
	fs := dendrik.NewFlagSet("pbcopy")
	folioPath := fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if !resolveOrDie(folioPath) {
		return dendrik.ExitUserError
	}

	if len(fs.GetArgs()) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: folio pbcopy <target-id> [--folio PATH]\n")
		fmt.Fprintf(os.Stderr, "  Copies the first local output of a target to the clipboard.\n")
		return dendrik.ExitUserError
	}

	targetID := fs.GetArgs()[0]

	f, err := config.Load(*folioPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}

	target, ok := f.Targets[targetID]
	if !ok {
		fmt.Fprintln(os.Stderr, pal.Errf("target '%s' not found", targetID))
		fmt.Fprintf(os.Stderr, "Available targets:")
		for tid := range f.Targets {
			fmt.Fprintf(os.Stderr, " %s", tid)
		}
		fmt.Fprintln(os.Stderr)
		return dendrik.ExitUserError
	}

	// Find first local output path
	folioDir := filepath.Dir(*folioPath)
	var outputPath string
	for _, out := range target.Outputs {
		if out.Path != "" {
			outputPath = config.ResolvePath(folioDir, out.Path)
			break
		}
	}

	if outputPath == "" {
		fmt.Fprintln(os.Stderr, pal.Errf("target '%s' has no local output path", targetID))
		return dendrik.ExitUserError
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, pal.Errf("output file not found: %s", outputPath))
		return dendrik.ExitUserError
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("reading %s: %s", outputPath, err))
		return dendrik.ExitUserError
	}

	cmd := exec.Command("pbcopy")
	cmd.Stdin = bytes.NewReader(data)
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("pbcopy: %s", err))
		return dendrik.ExitUserError
	}

	fmt.Println(pal.Successf("Copied %s to clipboard (%d bytes)", outputPath, len(data)))
	return dendrik.ExitOK
}
