package main

import (
	"fmt"
	"os"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runSetup(args []string) int {
	fs := dendrik.NewFlagSet("setup")
	checkMode := fs.Bool('c', "check", "Silent mode: exit 0 if OK, exit 1 if missing")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	folioBin, err := os.Executable()
	if err != nil {
		folioBin = "folio"
	}

	if !*checkMode {
		fmt.Println(output.Successf("folio %s (%s)", version, folioBin))
		fmt.Printf("\n%s%sAll dependencies satisfied.%s\n", output.Green, output.Bold, output.Reset)
	}
	return 0
}
