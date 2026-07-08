package main

import (
	"fmt"
	"os"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runSetup(checkMode bool) int {
	pal := dendrik.NewPalette(true)

	folioBin, err := os.Executable()
	if err != nil {
		folioBin = "folio"
	}

	if !checkMode {
		fmt.Println(pal.Successf("folio %s (%s)", version, folioBin))
		fmt.Printf("\n%s%sAll dependencies satisfied.%s\n", pal.Green, pal.Bold, pal.Reset)
	}
	return dendrik.ExitOK
}
