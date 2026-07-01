package main

import (
	"fmt"
	"os"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/sync"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

// runFleet dispatches `folio fleet <subcommand>`. P1 ships `status` (read-only);
// `workarea` lands in P2.
func runFleet(args []string) int {
	if len(args) == 0 {
		printFleetUsage()
		return dendrik.ExitUserError
	}
	switch args[0] {
	case "status":
		return runFleetStatus(args[1:])
	case "--help", "-h", "help":
		printFleetUsage()
		return dendrik.ExitOK
	default:
		fmt.Fprintf(os.Stderr, "Unknown fleet command: %s\n", args[0])
		printFleetUsage()
		return dendrik.ExitUserError
	}
}

func printFleetUsage() {
	fmt.Fprintln(os.Stderr, "Usage: folio fleet <command>")
	fmt.Fprintln(os.Stderr, "  status [--dirty] [--json]   Read-only status across every registered store")
}

// runFleetStatus fans a cheap, read-only probe across every registered store
// (all kinds) and prints a grouped view. It NEVER mutates anything and never
// blocks on one bad repo: each probe is timeout-bounded and errors render "?".
func runFleetStatus(args []string) int {
	fs := dendrik.NewFlagSet("fleet status")
	dirtyOnly := fs.BoolLong("dirty", "Show only stores with uncommitted changes")
	jsonMode := fs.Bool('j', "json", "Machine-readable JSON output")
	noColor := fs.BoolLong("no-color", "Disable colored output")
	if done, code := dendrik.ParseCheck(fs, args); done {
		return code
	}
	color := dendrik.ColorEnabled(*noColor)
	pal := dendrik.NewPalette(color)

	reg, err := config.LoadRegistry()
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}

	stores := reg.AllStores()
	statuses := make([]sync.StoreStatus, 0, len(stores))
	for _, s := range stores {
		st := sync.For(s).Status(s.Path, s)
		if *dirtyOnly && !st.Dirty && st.Err == "" {
			continue
		}
		statuses = append(statuses, st)
	}

	if *jsonMode {
		dendrik.WriteResult(os.Stdout, statuses)
		return dendrik.ExitOK
	}

	if len(statuses) == 0 {
		fmt.Println("No stores registered (or none matched).")
		return dendrik.ExitOK
	}
	printFleetStatus(statuses, color)
	return dendrik.ExitOK
}

// printFleetStatus renders statuses grouped by kind, in a stable kind order.
func printFleetStatus(statuses []sync.StoreStatus, color bool) {
	pal := dendrik.NewPalette(color)
	kindOrder := []string{config.KindFolio, config.KindDot, config.KindCode, config.KindExternal}
	byKind := map[string][]sync.StoreStatus{}
	for _, st := range statuses {
		k := st.Kind
		if k == "" {
			k = config.KindFolio
		}
		byKind[k] = append(byKind[k], st)
	}
	// Any kind not in kindOrder (shouldn't happen) still prints, appended.
	for k := range byKind {
		if !contains(kindOrder, k) {
			kindOrder = append(kindOrder, k)
		}
	}

	first := true
	for _, k := range kindOrder {
		group := byKind[k]
		if len(group) == 0 {
			continue
		}
		if !first {
			fmt.Println()
		}
		first = false
		label := kindLabel(k)
		if color {
			fmt.Printf("%s%s%s (%d)\n", pal.Bold, label, pal.Reset, len(group))
		} else {
			fmt.Printf("%s (%d)\n", label, len(group))
		}
		nameW, branchW := 4, 6
		for _, st := range group {
			if len(st.Name) > nameW {
				nameW = len(st.Name)
			}
			if len(st.Branch) > branchW {
				branchW = len(st.Branch)
			}
		}
		if branchW > 40 { // cap so one long bookmark can't blow out the column
			branchW = 40
		}
		for _, st := range group {
			fmt.Printf("  %-*s  %s\n", nameW, st.Name, statusCell(st, branchW, pal, color))
		}
	}
}

// statusCell renders one store's state: an error "?", a "missing" note, or the
// branch + dirty/ahead/behind summary.
func statusCell(st sync.StoreStatus, branchW int, pal dendrik.Palette, color bool) string {
	if st.Err != "" {
		if color {
			return fmt.Sprintf("%s?%s %s", pal.Yellow, pal.Reset, st.Err)
		}
		return "? " + st.Err
	}
	if !st.Exists {
		return "missing"
	}
	detail := st.Detail
	if detail == "" {
		detail = "clean"
	}
	branch := st.Branch
	if branch == "" {
		branch = "-"
	}
	line := fmt.Sprintf("%-*s  %s", branchW, branch, detail)
	if color && st.Dirty {
		return pal.Yellow + line + pal.Reset
	}
	return line
}

func kindLabel(k string) string {
	switch k {
	case config.KindFolio:
		return "Folio KB"
	case config.KindDot:
		return "Dotfiles"
	case config.KindCode:
		return "Code"
	case config.KindExternal:
		return "External (read-only)"
	default:
		return k
	}
}

func contains(xs []string, x string) bool {
	for _, v := range xs {
		if v == x {
			return true
		}
	}
	return false
}
