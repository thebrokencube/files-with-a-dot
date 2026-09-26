package main

import (
	"fmt"
	"os"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/lease"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

// runLease runs one `folio lease` verb against a registered store. The exit
// code is the decision itself: 0 ok, 1 fix the invocation, 2 launch failed,
// 3 someone holds the slot.
func runLease(verb, storeName string, jsonMode, dryRun bool, timeout string) int {
	pal := dendrik.NewPalette(!jsonMode && dendrik.ColorEnabled(false))
	fail := func(code int, err error) int {
		if jsonMode {
			dendrik.WriteError(os.Stdout, err.Error(), "")
		} else {
			fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		}
		return code
	}

	ctx, err := resolveContext("", contextFleet)
	if err != nil {
		return fail(dendrik.ExitUserError, err)
	}
	umbrella := ctx.Umbrella
	if umbrella == "" {
		umbrella = ctx.WorkRoot
	}
	store, ok := ctx.Registry.Lookup(storeName)
	if !ok {
		return fail(dendrik.ExitUserError, fmt.Errorf("store %q is not registered in stores.yml", storeName))
	}
	t := lease.Target{Umbrella: umbrella, Registry: ctx.Registry, Store: store, Cwd: mustGetwd()}

	var rep lease.Report
	var code int
	switch verb {
	case "run":
		d, perr := time.ParseDuration(timeout)
		if perr != nil || d <= 0 {
			return fail(dendrik.ExitUserError, fmt.Errorf("--timeout %q: want a positive duration such as 180s", timeout))
		}
		rep, code, err = lease.Run(t, d, dryRun)
	case "status":
		rep, code, err = lease.Status(t)
	case "release":
		rep, code, err = lease.Release(t, dryRun)
	}
	if err != nil {
		return fail(code, err)
	}

	if jsonMode {
		if werr := dendrik.WriteResultWithExit(os.Stdout, rep, code); werr != nil {
			return dendrik.ExitExternalErr
		}
		return code
	}
	if code != dendrik.ExitOK {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", rep.Message))
		if rep.LogTail != "" {
			fmt.Fprintf(os.Stderr, "%s--- tail of %s%s\n%s\n", pal.Dim, rep.Log, pal.Reset, rep.LogTail)
		}
		return code
	}
	if rep.State == "" {
		fmt.Println(pal.Successf("%s", rep.Message))
		return code
	}
	fmt.Printf("%s%s%s  %s\n", pal.Bold, rep.State, pal.Reset, rep.Message)
	if rep.Env != "n/a" {
		fmt.Printf("  env  %s\n", rep.Env)
	}
	if rep.Log != "" {
		fmt.Printf("  log  %s\n", rep.Log)
	}
	return code
}
