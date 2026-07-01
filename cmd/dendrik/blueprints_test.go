package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestBlueprintComposesResolve asserts every per-system blueprint's `composes:` ids resolve to a
// real entry in the generated building-block map. A blueprint is only build-from-able if its
// concept references point at concepts that actually exist — this catches a stale or typo'd id.
func TestBlueprintComposesResolve(t *testing.T) {
	const refs = "skill/references"

	raw, err := os.ReadFile(filepath.Join(refs, "building-blocks.json"))
	if err != nil {
		t.Fatalf("read building-blocks.json (run scripts/generate-building-blocks): %v", err)
	}
	var m struct {
		Blocks []struct {
			ID string `json:"id"`
		} `json:"blocks"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("parse building-blocks.json: %v", err)
	}
	known := map[string]bool{}
	for _, b := range m.Blocks {
		known[b.ID] = true
	}

	blueprints, err := filepath.Glob(filepath.Join(refs, "*-blueprint.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(blueprints) == 0 {
		t.Fatal("no *-blueprint.md files found")
	}

	for _, bp := range blueprints {
		data, err := os.ReadFile(bp)
		if err != nil {
			t.Fatal(err)
		}
		name := filepath.Base(bp)
		found := false
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "-"))
			rest, ok := strings.CutPrefix(line, "composes:")
			if !ok {
				continue
			}
			found = true
			for _, id := range strings.Split(rest, ",") {
				id = strings.TrimSpace(id)
				if id != "" && !known[id] {
					t.Errorf("%s: composes id %q not in building-blocks.json", name, id)
				}
			}
			break
		}
		if !found {
			t.Errorf("%s: no `composes:` line", name)
		}
	}
}
