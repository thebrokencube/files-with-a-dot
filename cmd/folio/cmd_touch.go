package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/touch"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runTouch(args []string) int {
	fs := dendrik.NewFlagSet("touch")
	folioPath := fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if !resolveOrDie(folioPath) {
		return dendrik.ExitUserError
	}

	if len(fs.GetArgs()) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: folio touch <target-id> [--folio PATH]\n")
		fmt.Fprintf(os.Stderr, "  Marks a target as current by updating output file mtimes.\n")
		return dendrik.ExitUserError
	}

	targetID := fs.GetArgs()[0]

	f, err := config.Load(*folioPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return dendrik.ExitUserError
	}

	target, ok := f.Targets[targetID]
	if !ok {
		fmt.Fprintln(os.Stderr, output.Errf("target '%s' not found", targetID))
		return dendrik.ExitUserError
	}

	folioDir := filepath.Dir(*folioPath)
	touched, err := touch.Target(folioDir, &target)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return dendrik.ExitUserError
	}

	if touched == 0 {
		fmt.Fprintln(os.Stderr, output.Errf("target '%s' has no local output paths", targetID))
		return dendrik.ExitUserError
	}

	fmt.Println(output.Successf("Touched %d output(s) for %s", touched, targetID))
	return 0
}
