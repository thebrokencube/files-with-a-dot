// Package cmdtest is a layer-4 end-to-end harness: it builds a consumer CLI and
// execs the real binary, exercising main() wiring that in-process Execute tests
// cannot reach. Each cmd is a separate module, so it shells out to `go build`
// rather than importing a sibling main.
package cmdtest

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// repoRoot walks up from the test's working directory to the dir holding go.work.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("cmdtest: getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("cmdtest: go.work not found above working dir")
		}
		dir = parent
	}
}

// Build compiles ./cmd/<tool> into a temp binary and returns its path. The build
// is unstamped (no -ldflags), so main.version is its default ("dev").
func Build(t *testing.T, tool string) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), tool)
	cmd := exec.Command("go", "build", "-o", bin, "./cmd/"+tool)
	cmd.Dir = repoRoot(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("cmdtest: build %s: %v\n%s", tool, err, out)
	}
	return bin
}

// Run execs bin with args and returns (exitCode, combinedStdoutStderr).
func Run(t *testing.T, bin string, args ...string) (int, string) {
	t.Helper()
	out, err := exec.Command(bin, args...).CombinedOutput()
	if err == nil {
		return 0, string(out)
	}
	var ee *exec.ExitError
	if ok := asExitError(err, &ee); ok {
		return ee.ExitCode(), string(out)
	}
	t.Fatalf("cmdtest: run %v: %v\n%s", args, err, out)
	return -1, ""
}

func asExitError(err error, target **exec.ExitError) bool {
	if ee, ok := err.(*exec.ExitError); ok {
		*target = ee
		return true
	}
	return false
}
