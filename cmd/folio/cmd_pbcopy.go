package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runPbcopy(args []string) int {
	fs := dendrik.NewFlagSet("pbcopy")
	folioPath := fs.StringLong("folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if !resolveOrDie(folioPath) {
		return 1
	}

	if len(fs.GetArgs()) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: folio pbcopy <target-id> [--folio PATH]\n")
		fmt.Fprintf(os.Stderr, "  Copies the first local output of a target to the clipboard.\n")
		return 1
	}

	targetID := fs.GetArgs()[0]

	f, err := config.Load(*folioPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	target, ok := f.Targets[targetID]
	if !ok {
		fmt.Fprintln(os.Stderr, output.Errf("target '%s' not found", targetID))
		fmt.Fprintf(os.Stderr, "Available targets:")
		for tid := range f.Targets {
			fmt.Fprintf(os.Stderr, " %s", tid)
		}
		fmt.Fprintln(os.Stderr)
		return 1
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
		fmt.Fprintln(os.Stderr, output.Errf("target '%s' has no local output path", targetID))
		return 1
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, output.Errf("output file not found: %s", outputPath))
		return 1
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("reading %s: %s", outputPath, err))
		return 1
	}

	cmd := exec.Command("pbcopy")
	cmd.Stdin = bytes.NewReader(data)
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("pbcopy: %s", err))
		return 1
	}

	fmt.Println(output.Successf("Copied %s to clipboard (%d bytes)", outputPath, len(data)))
	return 0
}
