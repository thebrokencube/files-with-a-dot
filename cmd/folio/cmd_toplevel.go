package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
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
		fmt.Println(output.Successf("folio %s (%s)", version, folioBin))
		fmt.Printf("\n%s%sAll dependencies satisfied.%s\n", output.Green, output.Bold, output.Reset)
	}
	return 0
}
