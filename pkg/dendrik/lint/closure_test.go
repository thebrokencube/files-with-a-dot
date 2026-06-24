package lint

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik/conventions"
)

// lintResultCallRe matches the CheckID literal of a lintResult("id", ...) call.
var lintResultCallRe = regexp.MustCompile(`lintResult\("([a-z0-9-]+)"`)

// TestContractClosure asserts every CheckID emitted by the linters is registered
// in conventions.Contract — so --explain resolves it and the check count is
// honest. It scans this package's own source for lintResult("id", ...) call sites
// rather than maintaining a second hand-written list (which would itself drift).
func TestContractClosure(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}

	emitted := map[string]bool{}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		for _, m := range lintResultCallRe.FindAllStringSubmatch(string(src), -1) {
			emitted[m[1]] = true
		}
	}

	if len(emitted) == 0 {
		t.Fatal("found no lintResult(...) call sites — the closure scan is vacuous")
	}

	for id := range emitted {
		if conventions.LookupCheck(id) == nil {
			t.Errorf("CheckID %q is emitted by a linter but absent from conventions.Contract", id)
		}
	}
}
