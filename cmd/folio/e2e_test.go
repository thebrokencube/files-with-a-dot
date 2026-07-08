package main

import (
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik/cmdtest"
)

// TestFolioE2E execs the real folio binary to exercise main()'s router wiring
// (help, arity, unknown-subcommand, version) that in-process tests cannot reach.
func TestFolioE2E(t *testing.T) {
	bin := cmdtest.Build(t, "folio")

	cases := []struct {
		name       string
		args       []string
		wantZero   bool
		wantSubstr string
	}{
		{"nested leaf help", []string{"home", "workspace", "create", "--help"}, true, "create"},
		{"group help", []string{"home", "--help"}, true, "workspace"},
		{"unknown flag", []string{"home", "workspace", "create", "--bogus"}, false, ""},
		{"stray positional", []string{"home", "workspace", "create", "extra"}, false, ""},
		{"unknown subcommand", []string{"nope"}, false, ""},
		{"version", []string{"--version"}, true, "folio"},
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
