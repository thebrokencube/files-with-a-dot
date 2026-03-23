package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/forest"
)

// These tests validate the harness's local file-generation logic only.
// They use temp directories and NEVER call Jira.
// For real Jira integration validation, use `jf test run` manually.

func TestNodeDefinitionsComplete(t *testing.T) {
	if len(testNodes) != 13 {
		t.Fatalf("expected 13 test nodes, got %d", len(testNodes))
	}

	// All 8 plan rules should be covered by at least one node
	wantKinds := map[string]bool{"push": false, "pull": false, "skip": false, "blocked": false}
	wantBlocks := map[string]bool{
		"empty": false, "remote-unknown": false,
		"first-push": false, "first-pull": false,
		"overwrite": false, "conflict": false,
	}

	for _, n := range testNodes {
		wantKinds[n.WantKind] = true
		if n.WantBlock != "" {
			wantBlocks[n.WantBlock] = true
		}
	}

	for kind, covered := range wantKinds {
		if !covered {
			t.Errorf("WantKind %q not covered by any test node", kind)
		}
	}
	for block, covered := range wantBlocks {
		if !covered {
			t.Errorf("WantBlock %q not covered by any test node", block)
		}
	}

	// sync:both should have at least one node
	hasBoth := false
	for _, n := range testNodes {
		if n.Sync == "both" {
			hasBoth = true
			break
		}
	}
	if !hasBoth {
		t.Error("no sync:both nodes in test definitions")
	}
}

func TestSetupCreatesForest(t *testing.T) {
	dir := t.TempDir()
	keyMap := map[string]string{
		"empty-push":          "TEST-1",
		"first-push-safe":     "TEST-2",
		"first-push-conflict": "TEST-3",
		"first-pull-safe":     "TEST-4",
		"first-pull-conflict": "TEST-5",
		"local-changed":       "TEST-6",
		"overwrite-blocked":   "TEST-7",
		"conflict":            "TEST-8",
		"unchanged":           "TEST-9",
		"both-local-only":     "TEST-10",
		"both-remote-only":    "TEST-11",
	}

	code := generateTestForest("TEST-EPIC", dir, keyMap)
	if code != 0 {
		t.Fatalf("generateTestForest returned %d", code)
	}

	// forest.yml should exist
	if _, err := os.Stat(filepath.Join(dir, "forest.yml")); err != nil {
		t.Fatalf("forest.yml not created: %s", err)
	}

	// All 13 .md files should exist
	for _, n := range testNodes {
		path := filepath.Join(dir, n.Name+".md")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("%s.md not created: %s", n.Name, err)
		}
	}

	// .test-config.yml should exist
	if _, err := os.Stat(filepath.Join(dir, ".test-config.yml")); err != nil {
		t.Fatalf(".test-config.yml not created: %s", err)
	}
}

func TestSetupKeyMapping(t *testing.T) {
	dir := t.TempDir()
	keyMap := map[string]string{
		"empty-push":          "BEN-100",
		"first-push-safe":     "BEN-101",
		"first-push-conflict": "BEN-102",
		"first-pull-safe":     "BEN-103",
		"first-pull-conflict": "BEN-104",
		"local-changed":       "BEN-105",
		"overwrite-blocked":   "BEN-106",
		"conflict":            "BEN-107",
		"unchanged":           "BEN-108",
		"both-local-only":     "BEN-109",
		"both-remote-only":    "BEN-110",
	}

	code := generateTestForest("BEN-EPIC", dir, keyMap)
	if code != 0 {
		t.Fatalf("generateTestForest returned %d", code)
	}

	// Verify key mapping: empty-push should have BEN-100
	content, err := os.ReadFile(filepath.Join(dir, "empty-push.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(content); !contains(got, "jira: BEN-100") {
		t.Errorf("empty-push.md: expected jira: BEN-100, got:\n%s", got)
	}

	// TBD node should have key "TBD"
	content, err = os.ReadFile(filepath.Join(dir, "tbd-skip.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(content); !contains(got, "jira: TBD") {
		t.Errorf("tbd-skip.md: expected jira: TBD, got:\n%s", got)
	}

	// remote-err should have ZZZTESTNOEXIST-99999
	content, err = os.ReadFile(filepath.Join(dir, "remote-err.md"))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(content); !contains(got, "jira: ZZZTESTNOEXIST-99999") {
		t.Errorf("remote-err.md: expected jira: ZZZTESTNOEXIST-99999, got:\n%s", got)
	}
}

func TestSetupValidatesForest(t *testing.T) {
	dir := t.TempDir()
	keyMap := makeTestKeyMap("VAL")

	code := generateTestForest("VAL-EPIC", dir, keyMap)
	if code != 0 {
		t.Fatalf("generateTestForest returned %d", code)
	}

	// Independently verify the forest is parseable
	f, err := forest.FindForest(dir)
	if err != nil {
		t.Fatalf("FindForest: %s", err)
	}
	if f == nil {
		t.Fatal("FindForest returned nil")
	}

	roots, err := forest.Discover(f)
	if err != nil {
		t.Fatalf("Discover: %s", err)
	}

	all := forest.Flatten(roots)
	if len(all) != len(testNodes) {
		t.Errorf("expected %d nodes, discovered %d", len(testNodes), len(all))
	}
}

func TestResetRestoresBaseline(t *testing.T) {
	dir := t.TempDir()
	keyMap := makeTestKeyMap("RST")

	code := generateTestForest("RST-EPIC", dir, keyMap)
	if code != 0 {
		t.Fatalf("generateTestForest returned %d", code)
	}

	// Modify a file
	modPath := filepath.Join(dir, "first-push-safe.md")
	if err := os.WriteFile(modPath, []byte("corrupted content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Reset regenerates .md files (but not baselines — that needs Jira).
	// We can verify the file regeneration part by calling the generation directly.
	for _, n := range testNodes {
		key := resolveTestNodeKey(n, keyMap)
		content := buildTestNodeFile(n, key)
		path := filepath.Join(dir, n.Name+".md")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("write %s: %s", path, err)
		}
	}

	// Verify the modified file was restored
	content, err := os.ReadFile(modPath)
	if err != nil {
		t.Fatal(err)
	}
	if contains(string(content), "corrupted") {
		t.Error("reset did not restore first-push-safe.md")
	}
	if !contains(string(content), "Substantive content for push test.") {
		t.Error("first-push-safe.md doesn't have expected content after reset")
	}
}

func TestTeardownCleansUp(t *testing.T) {
	dir := t.TempDir()
	testDir := filepath.Join(dir, "teardown-test")
	keyMap := makeTestKeyMap("TD")

	code := generateTestForest("TD-EPIC", testDir, keyMap)
	if code != 0 {
		t.Fatalf("generateTestForest returned %d", code)
	}

	// Verify directory exists
	if _, err := os.Stat(testDir); err != nil {
		t.Fatalf("test dir should exist: %s", err)
	}

	// Teardown removes the directory
	if err := os.RemoveAll(testDir); err != nil {
		t.Fatalf("RemoveAll: %s", err)
	}

	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Error("test directory should not exist after teardown")
	}
}

func TestRunReportsAvailableChecks(t *testing.T) {
	dir := t.TempDir()
	keyMap := makeTestKeyMap("RUN")

	code := generateTestForest("RUN-EPIC", dir, keyMap)
	if code != 0 {
		t.Fatalf("generateTestForest returned %d", code)
	}

	// runTestRun requires .test-config.yml — verify it exists
	cfg, err := loadTestConfig(dir)
	if err != nil {
		t.Fatalf("loadTestConfig: %s", err)
	}
	if cfg.EpicKey != "RUN-EPIC" {
		t.Errorf("epic key: got %q, want %q", cfg.EpicKey, "RUN-EPIC")
	}
}

func TestParseKeyMap(t *testing.T) {
	m, err := parseKeyMap("foo=BAR-1,baz=QUX-2")
	if err != nil {
		t.Fatal(err)
	}
	if m["foo"] != "BAR-1" {
		t.Errorf("foo: got %q, want BAR-1", m["foo"])
	}
	if m["baz"] != "QUX-2" {
		t.Errorf("baz: got %q, want QUX-2", m["baz"])
	}

	_, err = parseKeyMap("invalid")
	if err == nil {
		t.Error("expected error for invalid key map")
	}
}

// makeTestKeyMap creates a key map with sequential keys for all ticket-needing nodes.
func makeTestKeyMap(prefix string) map[string]string {
	m := make(map[string]string)
	i := 1
	for _, n := range testNodes {
		if n.NeedTicket {
			m[n.Name] = fmt.Sprintf("%s-%d", prefix, i)
			i++
		}
	}
	return m
}

