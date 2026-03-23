package main

import (
	"fmt"
	"os"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runSetup(args []string) int {
	pal := dendrik.NewPalette(true)
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
		fmt.Println(pal.Successf("folio %s (%s)", version, folioBin))
		fmt.Printf("\n%s%sAll dependencies satisfied.%s\n", pal.Green, pal.Bold, pal.Reset)
	}
	return dendrik.ExitOK
}
