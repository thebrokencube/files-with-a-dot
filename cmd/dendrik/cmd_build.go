package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik/build"
)

// resolveVersion reads <dir>/VERSION (the I/O the pure core can't do) and
// delegates validation to build.ParseVersion. It is the single source of truth
// for the stamped version.
func resolveVersion(dir, override string) (string, error) {
	if override != "" {
		return override, nil
	}
	b, err := os.ReadFile(filepath.Join(dir, "VERSION"))
	if err != nil {
		return "", fmt.Errorf("no --version given and no VERSION file in %s: %w", dir, err)
	}
	return build.ParseVersion(string(b), "")
}

func runBuild(o buildOpts, pos []string) int {
	out := dendrik.NewOutput(o.json, o.plain, o.noColor)

	dir := "."
	if len(pos) > 0 {
		dir = pos[0]
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return buildFail(out, "resolving dir", err)
	}
	tool := filepath.Base(absDir)

	version, err := resolveVersion(absDir, o.versionOverride)
	if err != nil {
		return buildFail(out, "resolving version", err)
	}

	absOut, err := filepath.Abs(o.outDir)
	if err != nil {
		return buildFail(out, "resolving out dir", err)
	}
	if err := os.MkdirAll(absOut, 0o755); err != nil {
		return buildFail(out, "creating out dir", err)
	}

	var artifacts []build.Artifact
	for _, t := range build.Targets(o.matrix) {
		outPath := filepath.Join(absOut, build.ArtifactName(tool, t.OS, t.Arch))
		cmd := exec.Command("go", "build", "-C", absDir,
			"-trimpath", "-buildvcs=false",
			"-ldflags", build.LDFlags(version),
			"-o", outPath, ".")
		cmd.Env = append(os.Environ(), "GOOS="+t.OS, "GOARCH="+t.Arch)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return buildFail(out, fmt.Sprintf("building %s/%s", t.OS, t.Arch), err)
		}
		artifacts = append(artifacts, build.Artifact{OS: t.OS, Arch: t.Arch, Path: outPath})
	}

	res := build.Result{Tool: tool, Version: version, Artifacts: artifacts}
	if out.IsJSON() {
		fmt.Print(string(out.MustResult(res)))
		return dendrik.ExitOK
	}
	fmt.Println(out.Success("built %s %s (%d artifact(s))", tool, version, len(artifacts)))
	for _, a := range artifacts {
		fmt.Printf("  %s/%s -> %s\n", a.OS, a.Arch, a.Path)
	}
	return dendrik.ExitOK
}

// buildFail reports an error (JSON envelope or stderr) and returns the user-error exit code.
func buildFail(out dendrik.Output, what string, err error) int {
	if out.IsJSON() {
		_ = dendrik.WriteError(os.Stdout, what, err.Error())
	} else {
		fmt.Fprintf(os.Stderr, "Error %s: %v\n", what, err)
	}
	return dendrik.ExitUserError
}
