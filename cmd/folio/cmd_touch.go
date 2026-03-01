package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
)

func runTouch(args []string) int {
	fs := flag.NewFlagSet("touch", flag.ExitOnError)
	folioPath := fs.String("folio", "./folio.yml", "Path to folio.yml")
	fs.Parse(args)

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
	now := time.Now()
	touched := 0

	for _, out := range target.Outputs {
		if out.Path == "" {
			continue
		}
		fullPath := filepath.Join(folioDir, out.Path)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			fmt.Fprintln(os.Stderr, output.Errf("output file not found: %s", out.Path))
			return 1
		}
		if err := os.Chtimes(fullPath, now, now); err != nil {
			fmt.Fprintln(os.Stderr, output.Errf("touching %s: %s", out.Path, err))
			return 1
		}
		touched++
	}

	if touched == 0 {
		fmt.Fprintln(os.Stderr, output.Errf("target '%s' has no local output paths", targetID))
		return 1
	}

	fmt.Println(output.Successf("Touched %d output(s) for %s", touched, targetID))
	return 0
}
