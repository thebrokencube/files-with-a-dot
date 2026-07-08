package main

import (
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik/cmdtest"
)

// TestDendrikE2E execs the real dendrik binary to exercise main()'s router
// wiring (help, arity, unknown-subcommand, version) that in-process tests
// cannot reach.
func TestDendrikE2E(t *testing.T) {
	bin := cmdtest.Build(t, "dendrik")

	cases := []struct {
		name       string
		args       []string
		wantZero   bool
		wantSubstr string
	}{
		{"lint help", []string{"lint", "--help"}, true, "lint"},
		{"build help", []string{"build", "--help"}, true, "build"},
		{"unknown subcommand", []string{"nope"}, false, ""},
		{"bad flag", []string{"lint", "--bogus"}, false, ""},
		{"stray positional on lint", []string{"lint", "a", "b"}, false, ""},
		{"version", []string{"--version"}, true, "dendrik"},
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
