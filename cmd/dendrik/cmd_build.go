package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

// buildTarget is one GOOS/GOARCH pair to compile for.
type buildTarget struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

// buildArtifact is one produced binary.
type buildArtifact struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
	Path string `json:"path"`
}

type buildResult struct {
	Tool      string          `json:"tool"`
	Version   string          `json:"version"`
	Artifacts []buildArtifact `json:"artifacts"`
}

// releaseMatrix is the standard set of platforms produced for a release.
var releaseMatrix = []buildTarget{
	{OS: "darwin", Arch: "arm64"},
	{OS: "linux", Arch: "amd64"},
}

// resolveVersion returns the override if set, else the trimmed contents of
// <dir>/VERSION. It is the single source of truth for the stamped version.
func resolveVersion(dir, override string) (string, error) {
	if override != "" {
		return override, nil
	}
	b, err := os.ReadFile(filepath.Join(dir, "VERSION"))
	if err != nil {
		return "", fmt.Errorf("no --version given and no VERSION file in %s: %w", dir, err)
	}
	v := strings.TrimSpace(string(b))
	if v == "" {
		return "", fmt.Errorf("VERSION file in %s is empty", dir)
	}
	return v, nil
}

// buildTargets returns the release matrix when matrix is set, else the host platform only.
func buildTargets(matrix bool) []buildTarget {
	if matrix {
		return releaseMatrix
	}
	return []buildTarget{{OS: runtime.GOOS, Arch: runtime.GOARCH}}
}

// artifactName is the released binary's filename: <tool>-<os>-<arch>.
func artifactName(tool, goos, goarch string) string {
	return fmt.Sprintf("%s-%s-%s", tool, goos, goarch)
}

// buildLDFlags are the reproducible linker flags with the version stamped in.
func buildLDFlags(version string) string {
	return "-buildid= -X main.version=" + version
}

func runBuild(args []string) int {
	fs := dendrik.NewFlagSet("dendrik build")
	matrix := fs.BoolLong("matrix", "Build the release matrix (darwin/arm64, linux/amd64) instead of the host platform")
	outDir := fs.StringLong("out", "dist", "Output directory for artifacts")
	versionOverride := fs.StringLong("version", "", "Override the version (default: read <dir>/VERSION)")
	jsonFlag := fs.BoolLong("json", "JSON output")
	plainFlag := fs.BoolLong("plain", "Undecorated text output (no color, no JSON)")
	noColor := fs.BoolLong("no-color", "Disable color output")

	if done, code := dendrik.ParseCheck(fs, args); done {
		return code
	}

	out := dendrik.NewOutput(*jsonFlag, *plainFlag, *noColor)

	dir := "."
	if rem := fs.GetArgs(); len(rem) > 0 {
		dir = rem[0]
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return buildFail(out, "resolving dir", err)
	}
	tool := filepath.Base(absDir)

	version, err := resolveVersion(absDir, *versionOverride)
	if err != nil {
		return buildFail(out, "resolving version", err)
	}

	absOut, err := filepath.Abs(*outDir)
	if err != nil {
		return buildFail(out, "resolving out dir", err)
	}
	if err := os.MkdirAll(absOut, 0o755); err != nil {
		return buildFail(out, "creating out dir", err)
	}

	var artifacts []buildArtifact
	for _, t := range buildTargets(*matrix) {
		outPath := filepath.Join(absOut, artifactName(tool, t.OS, t.Arch))
		cmd := exec.Command("go", "build", "-C", absDir,
			"-trimpath", "-buildvcs=false",
			"-ldflags", buildLDFlags(version),
			"-o", outPath, ".")
		cmd.Env = append(os.Environ(), "GOOS="+t.OS, "GOARCH="+t.Arch)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return buildFail(out, fmt.Sprintf("building %s/%s", t.OS, t.Arch), err)
		}
		artifacts = append(artifacts, buildArtifact{OS: t.OS, Arch: t.Arch, Path: outPath})
	}

	res := buildResult{Tool: tool, Version: version, Artifacts: artifacts}
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
