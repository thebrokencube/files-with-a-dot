// Package build holds the pure functional core of the dendrik `build` verb:
// the release matrix, version parsing, artifact naming, and linker flags. It is
// I/O-free — the imperative shell (cmd/dendrik) reads files and runs `go build`.
package build

import (
	"fmt"
	"runtime"
	"strings"
)

// Target is one GOOS/GOARCH pair to compile for.
type Target struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

// Artifact is one produced binary.
type Artifact struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
	Path string `json:"path"`
}

// Result is the structured outcome of a build run.
type Result struct {
	Tool      string     `json:"tool"`
	Version   string     `json:"version"`
	Artifacts []Artifact `json:"artifacts"`
}

// ReleaseMatrix is the standard set of platforms produced for a release.
var ReleaseMatrix = []Target{
	{OS: "darwin", Arch: "arm64"},
	{OS: "darwin", Arch: "amd64"},
	{OS: "linux", Arch: "arm64"},
	{OS: "linux", Arch: "amd64"},
}

// ParseVersion returns the override if set, else the trimmed VERSION file
// content. It is the pure half of version resolution: the shell reads the file
// (and reports a missing file); this validates and normalizes the value.
func ParseVersion(content, override string) (string, error) {
	if override != "" {
		return override, nil
	}
	v := strings.TrimSpace(content)
	if v == "" {
		return "", fmt.Errorf("VERSION is empty")
	}
	return v, nil
}

// Targets returns the release matrix when matrix is set, else the host platform only.
func Targets(matrix bool) []Target {
	if matrix {
		return ReleaseMatrix
	}
	return []Target{{OS: runtime.GOOS, Arch: runtime.GOARCH}}
}

// ArtifactName is the released binary's filename: <tool>-<os>-<arch>.
func ArtifactName(tool, goos, goarch string) string {
	return fmt.Sprintf("%s-%s-%s", tool, goos, goarch)
}

// LDFlags are the reproducible linker flags with the version stamped in.
func LDFlags(version string) string {
	return "-buildid= -X main.version=" + version
}
