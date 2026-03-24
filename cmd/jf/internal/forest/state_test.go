package forest

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadStateEmpty(t *testing.T) {
	dir := t.TempDir()
	s, err := LoadState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Nodes) != 0 {
		t.Errorf("expected empty nodes, got %d", len(s.Nodes))
	}
}

func TestSaveAndLoadState(t *testing.T) {
	dir := t.TempDir()
	s := &State{Nodes: map[string]NodeState{
		"BEN-1": {LastPush: time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC)},
	}}

	if err := SaveState(dir, s); err != nil {
		t.Fatal(err)
	}

	// Verify file exists
	path := filepath.Join(dir, ".jf", "state.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("state.json not created: %v", err)
	}

	loaded, err := LoadState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(loaded.Nodes))
	}
	ns := loaded.Nodes["BEN-1"]
	if ns.LastPush.Year() != 2026 {
		t.Errorf("expected 2026, got %d", ns.LastPush.Year())
	}
}

func TestIsStaleNeverPushed(t *testing.T) {
	s := &State{Nodes: make(map[string]NodeState)}
	if !s.IsStale("BEN-1", time.Now()) {
		t.Error("expected stale for never-pushed node")
	}
}

func TestIsStaleAfterModification(t *testing.T) {
	pushTime := time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC)
	s := &State{Nodes: map[string]NodeState{
		"BEN-1": {LastPush: pushTime},
	}}

	// File modified after push
	modTime := pushTime.Add(time.Hour)
	if !s.IsStale("BEN-1", modTime) {
		t.Error("expected stale when file modified after push")
	}

	// File not modified after push
	oldTime := pushTime.Add(-time.Hour)
	if s.IsStale("BEN-1", oldTime) {
		t.Error("expected not stale when file older than push")
	}
}

func TestIsStaleUsesLastSync(t *testing.T) {
	syncTime := time.Date(2026, 3, 23, 12, 0, 0, 0, time.UTC)
	s := &State{Nodes: map[string]NodeState{
		"BEN-1": {LastSync: syncTime},
	}}

	// File modified after sync
	if !s.IsStale("BEN-1", syncTime.Add(time.Hour)) {
		t.Error("expected stale when file modified after sync")
	}

	// File not modified after sync
	if s.IsStale("BEN-1", syncTime.Add(-time.Hour)) {
		t.Error("expected not stale when file older than sync")
	}
}

func TestIsStaleLastSyncOverridesLastPush(t *testing.T) {
	pushTime := time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC)
	syncTime := time.Date(2026, 3, 23, 12, 0, 0, 0, time.UTC)
	s := &State{Nodes: map[string]NodeState{
		"BEN-1": {LastPush: pushTime, LastSync: syncTime},
	}}

	// Between push and sync — should be clean (LastSync is the reference)
	between := pushTime.Add(time.Hour)
	if s.IsStale("BEN-1", between) {
		t.Error("expected not stale: file older than LastSync even though newer than LastPush")
	}
}

func TestLoadStateCorrupt(t *testing.T) {
	dir := t.TempDir()
	stateDir := filepath.Join(dir, ".jf")
	os.MkdirAll(stateDir, 0755)
	os.WriteFile(filepath.Join(stateDir, "state.json"), []byte("{invalid json"), 0644)

	s, err := LoadState(dir)
	if err != nil {
		t.Fatalf("expected no error for corrupt state, got %v", err)
	}
	if len(s.Nodes) != 0 {
		t.Errorf("expected empty nodes for corrupt state, got %d", len(s.Nodes))
	}
}

func TestSaveStateAtomic(t *testing.T) {
	dir := t.TempDir()

	s := &State{Nodes: map[string]NodeState{
		"BEN-1": {LastPush: time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC)},
	}}
	if err := SaveState(dir, s); err != nil {
		t.Fatal(err)
	}

	// Verify no temp files left behind
	entries, _ := os.ReadDir(filepath.Join(dir, ".jf"))
	for _, e := range entries {
		if e.Name() != "state.json" {
			t.Errorf("unexpected file in .jf/: %s", e.Name())
		}
	}

	// Verify content is valid
	loaded, err := LoadState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(loaded.Nodes))
	}
}

func TestMutabilityCache(t *testing.T) {
	s := &State{Nodes: map[string]NodeState{}}

	// Miss on empty
	_, found := s.MutabilityCache("KEY-1", "abc123")
	if found {
		t.Error("expected miss on empty state")
	}

	// Set and hit
	s.SetMutability("KEY-1", "abc123", true)
	clean, found := s.MutabilityCache("KEY-1", "abc123")
	if !found || !clean {
		t.Errorf("expected cache hit clean=true, got found=%v clean=%v", found, clean)
	}

	// Miss on hash change
	_, found = s.MutabilityCache("KEY-1", "def456")
	if found {
		t.Error("expected miss on hash change")
	}

	// Dirty cache
	s.SetMutability("KEY-1", "def456", false)
	clean, found = s.MutabilityCache("KEY-1", "def456")
	if !found || clean {
		t.Errorf("expected cache hit clean=false, got found=%v clean=%v", found, clean)
	}
}

func TestComputeHash(t *testing.T) {
	h := ComputeHash([]byte("hello world"))
	if len(h) != 64 {
		t.Errorf("expected 64-char hex string, got %d chars", len(h))
	}
	// Deterministic
	if ComputeHash([]byte("hello world")) != h {
		t.Error("expected deterministic hash")
	}
	// Different content → different hash
	if ComputeHash([]byte("hello world!")) == h {
		t.Error("expected different hash for different content")
	}
}

func TestRecordSync(t *testing.T) {
	s := &State{Nodes: make(map[string]NodeState)}
	s.RecordSync("BEN-1", "push", "localhash", "remotehash")

	ns, ok := s.Nodes["BEN-1"]
	if !ok {
		t.Fatal("expected node state after RecordSync")
	}
	if time.Since(ns.LastSync) > time.Second {
		t.Error("expected LastSync to be recent")
	}
	if ns.Direction != "push" {
		t.Errorf("expected Direction 'push', got %q", ns.Direction)
	}
	if ns.LocalHash != "localhash" {
		t.Errorf("expected LocalHash 'localhash', got %q", ns.LocalHash)
	}
	if ns.RemoteHash != "remotehash" {
		t.Errorf("expected RemoteHash 'remotehash', got %q", ns.RemoteHash)
	}
}

func TestMigrationPushOnly(t *testing.T) {
	pushTime := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	s := &State{Nodes: map[string]NodeState{
		"KEY-1": {LastPush: pushTime, LocalHash: "lh", RemoteHash: "rh"},
	}}
	migrateState(s)
	ns := s.Nodes["KEY-1"]
	if !ns.LastSync.Equal(pushTime) {
		t.Errorf("LastSync = %v, want %v", ns.LastSync, pushTime)
	}
	if ns.Direction != "push" {
		t.Errorf("Direction = %q, want push", ns.Direction)
	}
	if ns.LocalHash != "lh" || ns.RemoteHash != "rh" {
		t.Error("hashes not preserved")
	}
}

func TestMigrationPullOnly(t *testing.T) {
	pullTime := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	s := &State{Nodes: map[string]NodeState{
		"KEY-1": {LastPull: pullTime, LocalHash: "lh"},
	}}
	migrateState(s)
	ns := s.Nodes["KEY-1"]
	if !ns.LastSync.Equal(pullTime) {
		t.Errorf("LastSync = %v, want %v", ns.LastSync, pullTime)
	}
	if ns.Direction != "pull" {
		t.Errorf("Direction = %q, want pull", ns.Direction)
	}
}

func TestMigrationBothTimestampsPushNewer(t *testing.T) {
	pushTime := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)
	pullTime := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	s := &State{Nodes: map[string]NodeState{
		"KEY-1": {LastPush: pushTime, LastPull: pullTime},
	}}
	migrateState(s)
	ns := s.Nodes["KEY-1"]
	if !ns.LastSync.Equal(pushTime) {
		t.Errorf("LastSync = %v, want %v", ns.LastSync, pushTime)
	}
	if ns.Direction != "push" {
		t.Errorf("Direction = %q, want push", ns.Direction)
	}
}

func TestMigrationBothTimestampsPullNewer(t *testing.T) {
	pushTime := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	pullTime := time.Date(2026, 3, 21, 10, 0, 0, 0, time.UTC)
	s := &State{Nodes: map[string]NodeState{
		"KEY-1": {LastPush: pushTime, LastPull: pullTime},
	}}
	migrateState(s)
	ns := s.Nodes["KEY-1"]
	if !ns.LastSync.Equal(pullTime) {
		t.Errorf("LastSync = %v, want %v", ns.LastSync, pullTime)
	}
	if ns.Direction != "pull" {
		t.Errorf("Direction = %q, want pull", ns.Direction)
	}
}

func TestMigrationAlreadyMigrated(t *testing.T) {
	syncTime := time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC)
	s := &State{Nodes: map[string]NodeState{
		"KEY-1": {LastSync: syncTime, Direction: "push", LocalHash: "lh", RemoteHash: "rh"},
	}}
	migrateState(s)
	ns := s.Nodes["KEY-1"]
	if !ns.LastSync.Equal(syncTime) {
		t.Error("LastSync changed during re-migration")
	}
	if ns.Direction != "push" {
		t.Error("Direction changed during re-migration")
	}
}

func TestMigrationNeitherSet(t *testing.T) {
	s := &State{Nodes: map[string]NodeState{
		"KEY-1": {},
	}}
	migrateState(s)
	ns := s.Nodes["KEY-1"]
	if !ns.LastSync.IsZero() {
		t.Error("LastSync should remain zero")
	}
}

func TestMigrationEmptyRemoteHash(t *testing.T) {
	pushTime := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	s := &State{Nodes: map[string]NodeState{
		"KEY-1": {LastPush: pushTime, LocalHash: "lh", RemoteHash: ""},
	}}
	migrateState(s)
	ns := s.Nodes["KEY-1"]
	if ns.RemoteHash != "" {
		t.Error("empty RemoteHash should be preserved as-is")
	}
}

func TestMigrationHashesPreserved(t *testing.T) {
	pushTime := time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC)
	s := &State{Nodes: map[string]NodeState{
		"KEY-1": {LastPush: pushTime, LocalHash: "abc123", RemoteHash: "def456"},
	}}
	migrateState(s)
	ns := s.Nodes["KEY-1"]
	if ns.LocalHash != "abc123" || ns.RemoteHash != "def456" {
		t.Errorf("hashes changed: local=%q remote=%q", ns.LocalHash, ns.RemoteHash)
	}
}

func TestMigrationViaLoadState(t *testing.T) {
	dir := t.TempDir()
	// Save old-format state
	s := &State{Nodes: map[string]NodeState{
		"KEY-1": {
			LastPush:   time.Date(2026, 3, 20, 10, 0, 0, 0, time.UTC),
			LocalHash:  "lh",
			RemoteHash: "rh",
		},
	}}
	if err := SaveState(dir, s); err != nil {
		t.Fatal(err)
	}

	// Load should trigger migration
	loaded, err := LoadState(dir)
	if err != nil {
		t.Fatal(err)
	}
	ns := loaded.Nodes["KEY-1"]
	if ns.LastSync.IsZero() {
		t.Error("LoadState should have triggered migration")
	}
	if ns.Direction != "push" {
		t.Errorf("Direction = %q, want push", ns.Direction)
	}
}

func TestIsPullStale(t *testing.T) {
	s := &State{Nodes: map[string]NodeState{
		"BEN-1": {LastPull: time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC)},
		"BEN-2": {LastPush: time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC)},
	}}

	if s.IsPullStale("BEN-1") {
		t.Error("expected not stale for pulled node")
	}
	if !s.IsPullStale("BEN-2") {
		t.Error("expected stale for never-pulled node")
	}
	if !s.IsPullStale("BEN-3") {
		t.Error("expected stale for unknown node")
	}
}
