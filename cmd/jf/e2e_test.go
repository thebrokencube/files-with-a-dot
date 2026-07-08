package main

import (
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik/cmdtest"
)

// TestJfE2E execs the real jf binary to exercise main()'s router wiring (help,
// arity, unknown-subcommand, version) that in-process tests cannot reach. All
// cases stay on the near side of the Jira prereq guard: --help is handled during
// flag parsing (before Run), and the arity/dispatch errors never reach a Run
// body, so none of these shell out to acli.
func TestJfE2E(t *testing.T) {
	bin := cmdtest.Build(t, "jf")

	cases := []struct {
		name       string
		args       []string
		wantZero   bool
		wantSubstr string
	}{
		{"leaf help", []string{"clone", "--help"}, true, "clone"},
		{"unknown subcommand", []string{"nope"}, false, ""},
		{"bad flag", []string{"list", "--bogus"}, false, ""},
		{"stray positional on no-arg command", []string{"status", "extra"}, false, ""},
		{"version", []string{"--version"}, true, "jf"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, out := cmdtest.Run(t, bin, tc.args...)
			if tc.wantZero && code != 0 {
				t.Fatalf("code = %d, want 0\n%s", code, out)
			}
			if !tc.wantZero && code == 0 {
				t.Fatalf("code = 0, want non-zero\n%s", out)
			}
			if tc.wantSubstr != "" && !strings.Contains(out, tc.wantSubstr) {
				t.Fatalf("output missing %q\n%s", tc.wantSubstr, out)
			}
		})
	}
}
