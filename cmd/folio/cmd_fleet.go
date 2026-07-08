package main

import (
	"fmt"
	"os"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/fleet"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/home"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/sync"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

// runFleetStatus fans a cheap, read-only probe across every registered store
// (all kinds) and prints a grouped view. It NEVER mutates anything and never
// blocks on one bad repo: each probe is timeout-bounded and errors render "?".
func runFleetStatus(dirtyOnly, jsonMode, noColor bool) int {
	color := dendrik.ColorEnabled(noColor)
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
		if dirtyOnly && !st.Dirty && st.Err == "" {
			continue
		}
		statuses = append(statuses, st)
	}

	if jsonMode {
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

// resolveUmbrellaOrFail resolves the FOLIO_HOME umbrella the workarea
// subcommands act on. Each workarea leaf calls this itself (help/arity already
// resolved by the router before Run).
func resolveUmbrellaOrFail(pal dendrik.Palette) (string, int) {
	umbrella, err := home.Dir()
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return "", dendrik.ExitUserError
	}
	return umbrella, dendrik.ExitOK
}

func runWorkareaOpen(base, storeName, branch string) int {
	pal := dendrik.NewPalette(true)
	umbrella, code := resolveUmbrellaOrFail(pal)
	if code != dendrik.ExitOK {
		return code
	}
	reg, err := config.LoadRegistry()
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}
	store, ok := reg.Lookup(storeName)
	if !ok {
		fmt.Fprintln(os.Stderr, pal.Errf("store %q is not registered in stores.yml", storeName))
		return dendrik.ExitUserError
	}
	if store.Kind != config.KindCode {
		fmt.Fprintln(os.Stderr, pal.Errf("workarea is only for code stores; %q is %s (folio KBs use `folio home workspace`; dotfiles is worked in place)", storeName, store.Kind))
		return dendrik.ExitUserError
	}
	wa, err := fleet.Open(umbrella, store, branch, base, os.Getenv("FOLIO_SESSION"))
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}
	fmt.Println(pal.Successf("Opened %s work area at %s (branch %s, off %s)", wa.Tier, wa.Dir, wa.Branch, wa.Base))
	fmt.Printf("  cd %s\n", wa.Dir)
	return dendrik.ExitOK
}

func runWorkareaList() int {
	pal := dendrik.NewPalette(true)
	umbrella, code := resolveUmbrellaOrFail(pal)
	if code != dendrik.ExitOK {
		return code
	}
	rows, orphans, err := fleet.Reconcile(umbrella)
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}
	if len(rows) == 0 && len(orphans) == 0 {
		fmt.Println("No work areas.")
		return dendrik.ExitOK
	}
	for _, r := range rows {
		fmt.Printf("  %-10s  %-24s  %-14s  %s\n", r.State, r.Branch, r.Tier, r.Dir)
	}
	for _, o := range orphans {
		fmt.Printf("  %-10s  %-24s  %-14s  %s\n", "orphan", "-", "-", o)
	}
	return dendrik.ExitOK
}

func runWorkareaReap(all, force bool, only string) int {
	pal := dendrik.NewPalette(true)
	umbrella, code := resolveUmbrellaOrFail(pal)
	if code != dendrik.ExitOK {
		return code
	}
	if only == "" && !all {
		// Default reap prunes dead rows + removes clean areas; make intent explicit.
		all = true
	}
	_ = all // preserved: the --all flag is a no-op sentinel in the current fleet.Reap contract
	actions, err := fleet.Reap(umbrella, only, force)
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}
	if len(actions) == 0 {
		fmt.Println("Nothing to reap.")
		return dendrik.ExitOK
	}
	for _, a := range actions {
		fmt.Printf("  %s\n", a)
	}
	return dendrik.ExitOK
}
