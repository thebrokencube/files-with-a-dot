package main

import (
	"fmt"
	"os"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

// runStores handles `folio stores <subcommand>`. The registry's CLI surface —
// the skill's `find` fan-out enumerates stores through `stores list --json`.
func runStores(args []string) int {
	if len(args) == 0 {
		args = []string{"list"}
	}
	switch args[0] {
	case "list":
		return runStoresList(args[1:])
	case "--help", "-h", "help":
		printStoresUsage()
		return dendrik.ExitOK
	default:
		fmt.Fprintf(os.Stderr, "Unknown stores command: %s\n", args[0])
		printStoresUsage()
		return dendrik.ExitUserError
	}
}

type storeJSON struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Kind string `json:"kind"`
}

func runStoresList(args []string) int {
	pal := dendrik.NewPalette(true)
	fs := dendrik.NewFlagSet("stores list")
	jsonMode := fs.Bool('j', "json", "Machine-readable JSON output")
	noColor := fs.BoolLong("no-color", "Disable colored output")
	if done, code := dendrik.ParseCheck(fs, args); done {
		return code
	}
	if *noColor {
		pal = dendrik.NewPalette(false)
	}

	ctx, err := resolveContext("", contextFleet)
	if err != nil {
		if *jsonMode {
			dendrik.WriteError(os.Stdout, fmt.Sprintf("%s", err), "")
		} else {
			fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		}
		return dendrik.ExitUserError
	}
	reg := ctx.Registry

	if *jsonMode {
		out := make([]storeJSON, 0, len(reg.Order))
		for _, name := range reg.Order {
			s := reg.Stores[name]
			out = append(out, storeJSON{Name: s.Name, Path: s.Path, Kind: s.Kind})
		}
		if err := dendrik.WriteResult(os.Stdout, out); err != nil {
			return dendrik.ExitExternalErr
		}
		return dendrik.ExitOK
	}

	if len(reg.Order) == 0 {
		fmt.Println("No stores registered (legacy isolated-home). `vault:` resolves under the selected FOLIO_HOME content root.")
		fmt.Println("Add a registry at $FOLIO_UMBRELLA/stores.yml.")
		return dendrik.ExitOK
	}
	for _, name := range reg.Order {
		s := reg.Stores[name]
		fmt.Printf("%-12s %-9s %s\n", s.Name, s.Kind, s.Path)
	}
	return dendrik.ExitOK
}

func printStoresUsage() {
	fmt.Fprintf(os.Stderr, `Usage: folio stores <command> [flags]

Commands:
  list       List registered stores (name, kind, path). --json for machine output.

The registry lives at $FOLIO_UMBRELLA/stores.yml. Without stores.yml, the
implicit legacy root is selected through FOLIO_HOME (or ~/.folio).
`)
}
