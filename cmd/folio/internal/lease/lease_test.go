package lease

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

func TestScriptIsolatesProcesses(t *testing.T) {
	dir, err := config.CanonicalPath(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	s := Serve{Processes: []Process{
		{Port: 1, Run: "cd / # moves only this process"},
		{Port: 2, Run: "true &"},
		{Port: 3, Run: "pwd > where.txt"},
	}}
	cmd := exec.Command("sh", "-c", script(s))
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("sh -c %q: %v\n%s", script(s), err, out)
	}
	got, err := os.ReadFile(filepath.Join(dir, "where.txt"))
	if err != nil {
		t.Fatalf("third process never ran: %v", err)
	}
	if where := strings.TrimSpace(string(got)); where != dir {
		t.Fatalf("third process ran in %q, want the store path %q", where, dir)
	}
}
