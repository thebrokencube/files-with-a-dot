package conventions

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
)

// contractChecksRef is the one derived enumeration of the contract. The
// count-assert below keeps it honest against the canonical Contract slice so
// the documented catalog cannot silently re-drift (e.g. "29 vs 30").
//
// Path is resolved relative to this test file so it works regardless of the
// working directory the test runner uses.
func contractChecksRefPath(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not resolve test file path")
	}
	// pkg/dendrik/conventions/ -> repo root -> plugins/dendrik/skills/dendrik/references/
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
	return filepath.Join(root, "plugins", "dendrik", "skills", "dendrik", "references", "contract-checks.md")
}

var refHeadingRe = regexp.MustCompile(`(?m)^### ([a-z][a-z0-9-]*)`)

// enumeratedCheckIDs returns the check IDs documented as `### <id>` headings in
// the derived contract-checks.md reference.
func enumeratedCheckIDs(t *testing.T) map[string]bool {
	t.Helper()
	data, err := os.ReadFile(contractChecksRefPath(t))
	if err != nil {
		t.Fatalf("read contract-checks.md: %v", err)
	}
	ids := map[string]bool{}
	for _, m := range refHeadingRe.FindAllStringSubmatch(string(data), -1) {
		ids[m[1]] = true
	}
	return ids
}

// TestContractCheckCount asserts the enumerated reference covers exactly the
// canonical Contract slice — every check, no extras. Derives the expected count
// from len(Contract); never hardcodes a literal.
func TestContractCheckCount(t *testing.T) {
	enumerated := enumeratedCheckIDs(t)

	if got, want := len(enumerated), len(Contract); got != want {
		t.Errorf("contract-checks.md enumerates %d checks; Contract has %d. The derived reference has drifted from the canonical slice.", got, want)
	}

	canonical := map[string]bool{}
	for _, c := range Contract {
		canonical[c.ID] = true
	}

	for _, c := range Contract {
		if !enumerated[c.ID] {
			t.Errorf("check %q is in Contract but missing from contract-checks.md", c.ID)
		}
	}
	for id := range enumerated {
		if !canonical[id] {
			t.Errorf("check %q is enumerated in contract-checks.md but not in Contract", id)
		}
	}
}

// TestContractChecksByLayer asserts every contract entry belongs to a known
// layer and that ChecksByLayer partitions the slice exactly — so per-layer
// counts also derive from the canonical slice, not from documented numbers.
func TestContractChecksByLayer(t *testing.T) {
	layers := []Layer{LayerGo, LayerSkill, LayerBridge}

	total := 0
	for _, layer := range layers {
		total += len(ChecksByLayer(layer))
	}
	if total != len(Contract) {
		t.Errorf("ChecksByLayer over %v covers %d entries; Contract has %d (an entry has an unknown layer)", layers, total, len(Contract))
	}
}
