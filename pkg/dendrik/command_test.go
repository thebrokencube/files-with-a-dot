package dendrik

import (
	"strings"
	"testing"
)

// buildTree constructs a fresh tree per call so leaf flag vars are local (no
// package-level flag state) — mirrors how consumers must build their trees.
func buildTree() (Command, *bool, *[]string) {
	var verbose bool
	var ran []string
	root := Command{
		Name:    "tool",
		Version: "1.2.3",
		Sub: []Command{
			{Name: "flat", Short: "a flat leaf", Args: ArgsNone,
				Flags: func(fs *FlagSet) { verbose = *fs.Bool('v', "verbose", "verbose") },
				Run:   func(_ *FlagSet, _ []string) int { ran = append(ran, "flat"); return ExitOK }},
			{Name: "one", Short: "needs one arg", Args: ArgsExactly(1),
				Run: func(_ *FlagSet, pos []string) int { ran = append(ran, "one:"+pos[0]); return ExitOK }},
			{Name: "many", Short: "variadic", Args: ArgsBetween(1, -1),
				Run: func(_ *FlagSet, pos []string) int { ran = append(ran, "many:"+strings.Join(pos, ",")); return ExitOK }},
			{Name: "grp", Short: "a nested group", Sub: []Command{
				{Name: "leaf", Short: "deep leaf", Args: ArgsNone,
					Run: func(_ *FlagSet, _ []string) int { ran = append(ran, "grp.leaf"); return ExitOK }},
			}},
			{Name: "raw", Short: "escape hatch", //nolint:dispatch-router
				RunRaw: func(args []string) int { ran = append(ran, "raw:"+strings.Join(args, ",")); return ExitOK }},
		},
	}
	return root, &verbose, &ran
}

func TestExecuteDispatch(t *testing.T) {
	cases := []struct {
		name     string
		args     []string
		wantCode int
		wantRan  string // last recorded run, "" if none
	}{
		{"leaf runs", []string{"flat"}, ExitOK, "flat"},
		{"leaf with flag", []string{"flat", "-v"}, ExitOK, "flat"},
		{"exact arity ok", []string{"one", "x"}, ExitOK, "one:x"},
		{"variadic ok", []string{"many", "a", "b", "c"}, ExitOK, "many:a,b,c"},
		{"nested leaf", []string{"grp", "leaf"}, ExitOK, "grp.leaf"},
		{"runraw passthrough", []string{"raw", "--anything", "z"}, ExitOK, "raw:--anything,z"},
		{"unknown subcommand", []string{"nope"}, ExitUserError, ""},
		{"unknown nested", []string{"grp", "nope"}, ExitUserError, ""},
		{"missing required arg", []string{"one"}, ExitUserError, ""},
		{"stray positional", []string{"flat", "extra"}, ExitUserError, ""},
		{"too many for exact", []string{"one", "a", "b"}, ExitUserError, ""},
		{"unknown flag", []string{"flat", "--bogus"}, ExitUserError, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, _, ran := buildTree()
			code := root.Execute(tc.args)
			if code != tc.wantCode {
				t.Errorf("code = %d, want %d", code, tc.wantCode)
			}
			got := ""
			if len(*ran) > 0 {
				got = (*ran)[len(*ran)-1]
			}
			if got != tc.wantRan {
				t.Errorf("ran = %q, want %q", got, tc.wantRan)
			}
		})
	}
}

func TestExecuteHelpAtEveryLevel(t *testing.T) {
	for _, args := range [][]string{
		{"--help"}, {"-h"}, {"help"}, // root
		{"grp", "--help"}, {"grp", "help"}, // nested group
		{"flat", "--help"}, {"grp", "leaf", "-h"}, // leaves via pflag
	} {
		root, _, ran := buildTree()
		if code := root.Execute(args); code != ExitOK {
			t.Errorf("%v: code = %d, want ExitOK", args, code)
		}
		if len(*ran) != 0 {
			t.Errorf("%v: a handler ran (%v); help must not reach Run", args, *ran)
		}
	}
}

func TestExecuteEmptyGroupIsUserError(t *testing.T) {
	root, _, _ := buildTree()
	if code := root.Execute(nil); code != ExitUserError {
		t.Errorf("empty root = %d, want ExitUserError", code)
	}
	if code := root.Execute([]string{"grp"}); code != ExitUserError {
		t.Errorf("empty nested group = %d, want ExitUserError", code)
	}
}

func TestExecuteVersionRootOnly(t *testing.T) {
	for _, a := range []string{"version", "--version", "-V"} {
		root, _, ran := buildTree()
		if code := root.Execute([]string{a}); code != ExitOK {
			t.Errorf("root %q = %d, want ExitOK", a, code)
		}
		if len(*ran) != 0 {
			t.Errorf("root %q ran a handler", a)
		}
	}
	// Nested group has no Version → version token is just an unknown subcommand.
	root, _, _ := buildTree()
	if code := root.Execute([]string{"grp", "--version"}); code != ExitUserError {
		t.Errorf("grp --version = %d, want ExitUserError (root-only)", code)
	}
}

func TestExecutePanicsOnMalformedNode(t *testing.T) {
	cases := map[string]Command{
		"no kind":        {Name: "x"},
		"group and run":  {Name: "x", Sub: []Command{{Name: "y", Run: func(*FlagSet, []string) int { return 0 }}}, Run: func(*FlagSet, []string) int { return 0 }},
		"run and runraw": {Name: "x", Run: func(*FlagSet, []string) int { return 0 }, RunRaw: func([]string) int { return 0 }},
	}
	for name, cmd := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Errorf("expected panic for malformed node")
				}
			}()
			cmd.Execute(nil)
		})
	}
}
