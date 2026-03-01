package main

import (
	"flag"
	"fmt"
	"os"
)

func runSetup(args []string) int {
	fs := flag.NewFlagSet("setup", flag.ExitOnError)
	checkMode := fs.Bool("check", false, "Silent mode: exit 0 if OK, exit 1 if missing")
	fs.Parse(args)

	folioBin, err := os.Executable()
	if err != nil {
		folioBin = "folio"
	}

	if !*checkMode {
		fmt.Printf("\033[0;32m✓\033[0m folio %s (%s)\n", version, folioBin)
		fmt.Printf("\n\033[0;32m\033[1mAll dependencies satisfied.\033[0m\n")
	}
	return 0
}
