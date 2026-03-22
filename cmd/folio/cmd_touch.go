package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/touch"
)

func runTouch(args []string) int {
	fs := flag.NewFlagSet("touch", flag.ExitOnError)
	folioPath := fs.String("folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
	parseFlags(fs, args)

	if !resolveOrDie(folioPath) {
		return 1
	}

	if fs.NArg() == 0 {
		fmt.Fprintf(os.Stderr, "Usage: folio touch <target-id> [--folio PATH]\n")
		fmt.Fprintf(os.Stderr, "  Marks a target as current by updating output file mtimes.\n")
		return 1
	}

	targetID := fs.Arg(0)

	f, err := config.Load(*folioPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	target, ok := f.Targets[targetID]
	if !ok {
		fmt.Fprintln(os.Stderr, output.Errf("target '%s' not found", targetID))
		return 1
	}

	folioDir := filepath.Dir(*folioPath)
	touched, err := touch.Target(folioDir, &target)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	if touched == 0 {
		fmt.Fprintln(os.Stderr, output.Errf("target '%s' has no local output paths", targetID))
		return 1
	}

	fmt.Println(output.Successf("Touched %d output(s) for %s", touched, targetID))
	return 0
}
