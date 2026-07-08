package main

import (
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

// buildRoot constructs the dendrik command tree. It is called fresh per process
// (and per test), so every leaf's flag vars are locals captured by its
// Flags/Run closures — no package-level flag state, preserving reentrancy.
func buildRoot() dendrik.Command {
	return dendrik.Command{
		Name:    "dendrik",
		Short:   "dendrik CLI",
		Version: version,
		Sub: []dendrik.Command{
			cmdLint(),
			cmdBuild(),
		},
	}
}

type lintOpts struct {
	json, plain, strict, fix, noColor bool
	explain                           string
}

// cmdLint uses ArgsBetween(0,1) rather than ArgsExactly(1): the --explain mode
// takes zero positionals, while a normal lint takes exactly one. The router
// still rejects extras (>1); the zero-vs-one guard lives in runLint.
func cmdLint() dendrik.Command {
	var jsonFlag, plainFlag, strictFlag, fixFlag, noColor *bool
	var explainFlag *string
	return dendrik.Command{
		Name: "lint", Short: "Run tool contract validation", Args: dendrik.ArgsBetween(0, 1),
		Flags: func(fs *dendrik.FlagSet) {
			jsonFlag = fs.BoolLong("json", "JSON output")
			plainFlag = fs.BoolLong("plain", "Undecorated text output (no color, no JSON)")
			strictFlag = fs.BoolLong("strict", "Promote warnings to errors")
			fixFlag = fs.BoolLong("fix", "Apply mechanical fixes for auto-fixable checks, then re-lint")
			explainFlag = fs.StringLong("explain", "", "Show rationale for a check ID")
			noColor = fs.BoolLong("no-color", "Disable color output")
		},
		Run: func(_ *dendrik.FlagSet, pos []string) int {
			return runLint(lintOpts{
				json:    *jsonFlag,
				plain:   *plainFlag,
				strict:  *strictFlag,
				fix:     *fixFlag,
				noColor: *noColor,
				explain: *explainFlag,
			}, pos)
		},
	}
}

type buildOpts struct {
	matrix, json, plain, noColor bool
	outDir, versionOverride      string
}

func cmdBuild() dendrik.Command {
	var matrix, jsonFlag, plainFlag, noColor *bool
	var outDir, versionOverride *string
	return dendrik.Command{
		Name: "build", Short: "Build a tool's release artifacts (reproducible, version-stamped)", Args: dendrik.ArgsBetween(0, 1),
		Flags: func(fs *dendrik.FlagSet) {
			matrix = fs.BoolLong("matrix", "Build the release matrix (darwin/arm64, linux/amd64) instead of the host platform")
			outDir = fs.StringLong("out", "dist", "Output directory for artifacts")
			versionOverride = fs.StringLong("version", "", "Override the version (default: read <dir>/VERSION)")
			jsonFlag = fs.BoolLong("json", "JSON output")
			plainFlag = fs.BoolLong("plain", "Undecorated text output (no color, no JSON)")
			noColor = fs.BoolLong("no-color", "Disable color output")
		},
		Run: func(_ *dendrik.FlagSet, pos []string) int {
			return runBuild(buildOpts{
				matrix:          *matrix,
				json:            *jsonFlag,
				plain:           *plainFlag,
				noColor:         *noColor,
				outDir:          *outDir,
				versionOverride: *versionOverride,
			}, pos)
		},
	}
}
