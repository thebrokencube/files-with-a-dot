package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

func runPbcopy(args []string) int {
	fs := flag.NewFlagSet("pbcopy", flag.ExitOnError)
	folioPath := fs.String("folio", "./folio.yml", "Path to folio.yml")
	fs.Parse(args)

	if fs.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Usage: folio pbcopy <target-id> [--folio PATH]\n")
		fmt.Fprintf(os.Stderr, "  Copies the first local output of a target to the clipboard.\n")
		return 1
	}

	targetID := fs.Arg(0)

	f, err := config.Load(*folioPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[0;31mError:\033[0m %s\n", err)
		return 1
	}

	target, ok := f.Targets[targetID]
	if !ok {
		fmt.Fprintf(os.Stderr, "\033[0;31mError:\033[0m target '%s' not found\n", targetID)
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
			outputPath = filepath.Join(folioDir, out.Path)
			break
		}
	}

	if outputPath == "" {
		fmt.Fprintf(os.Stderr, "\033[0;31mError:\033[0m target '%s' has no local output path\n", targetID)
		return 1
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "\033[0;31mError:\033[0m output file not found: %s\n", outputPath)
		return 1
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\033[0;31mError:\033[0m reading %s: %s\n", outputPath, err)
		return 1
	}

	cmd := exec.Command("pbcopy")
	cmd.Stdin = bytes.NewReader(data)
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "\033[0;31mError:\033[0m pbcopy: %s\n", err)
		return 1
	}

	fmt.Printf("\033[0;32m✓\033[0m Copied %s to clipboard (%d bytes)\n", outputPath, len(data))
	return 0
}
