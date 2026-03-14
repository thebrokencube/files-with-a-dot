package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/pending"
)

func runAddPending(args []string) int {
	fs := flag.NewFlagSet("add-pending", flag.ExitOnError)
	folioPath := fs.String("folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
	fs.Parse(args)

	if !resolveOrDie(folioPath) {
		return 1
	}

	item := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if item == "" {
		fmt.Fprintf(os.Stderr, "Usage: folio observe <item text> [--folio PATH]\n")
		fmt.Fprintf(os.Stderr, "       folio add-pending <item text> [--folio PATH]  (compat alias)\n")
		return 1
	}

	// Validate the file parses before modifying
	if _, err := config.Load(*folioPath); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	if err := pending.Append(*folioPath, item); err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	fmt.Println(output.Successf("Added: %s", item))
	return 0
}
