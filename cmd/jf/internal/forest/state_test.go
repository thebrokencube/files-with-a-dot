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

func TestRecordPush(t *testing.T) {
	s := &State{Nodes: make(map[string]NodeState)}
	s.RecordPush("BEN-1", "testhash", "")

	ns, ok := s.Nodes["BEN-1"]
	if !ok {
		t.Fatal("expected node state after RecordPush")
	}
	if time.Since(ns.LastPush) > time.Second {
		t.Error("expected LastPush to be recent")
	}
	if ns.LocalHash != "testhash" {
		t.Errorf("expected LocalHash 'testhash', got %q", ns.LocalHash)
	}
	if ns.RemoteHash != "" {
		t.Errorf("expected empty RemoteHash, got %q", ns.RemoteHash)
	}
}

func TestRecordPushPreservesFields(t *testing.T) {
	s := &State{Nodes: map[string]NodeState{
		"BEN-1": {
			LastPull:   time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC),
			RemoteHash: "abc123",
			LocalHash:  "def456",
		},
	}}
	s.RecordPush("BEN-1", "newhash", "")

	ns := s.Nodes["BEN-1"]
	if ns.LastPull.IsZero() {
		t.Error("RecordPush should preserve LastPull")
	}
	if ns.LocalHash != "newhash" {
		t.Error("RecordPush should set LocalHash to new value")
	}
	if ns.RemoteHash != "" {
		t.Error("RecordPush should set RemoteHash to provided value")
	}
	if time.Since(ns.LastPush) > time.Second {
		t.Error("expected LastPush to be recent")
	}
}

func TestRecordPull(t *testing.T) {
	s := &State{Nodes: make(map[string]NodeState)}
	s.RecordPull("BEN-1", "localhash", "remotehash")

	ns, ok := s.Nodes["BEN-1"]
	if !ok {
		t.Fatal("expected node state after RecordPull")
	}
	if time.Since(ns.LastPull) > time.Second {
		t.Error("expected LastPull to be recent")
	}
	if ns.LocalHash != "localhash" {
		t.Errorf("expected LocalHash 'localhash', got %q", ns.LocalHash)
	}
	if ns.RemoteHash != "remotehash" {
		t.Errorf("expected RemoteHash 'remotehash', got %q", ns.RemoteHash)
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

func TestDetectConflictNone(t *testing.T) {
	local := []byte("content")
	remote := []byte("remote")
	s := &State{Nodes: map[string]NodeState{
		"BEN-1": {LocalHash: ComputeHash(local), RemoteHash: ComputeHash(remote)},
	}}
	if s.DetectConflict("BEN-1", local, remote) != ConflictNone {
		t.Error("expected ConflictNone")
	}
}

func TestDetectConflictLocalOnly(t *testing.T) {
	remote := []byte("remote")
	s := &State{Nodes: map[string]NodeState{
		"BEN-1": {LocalHash: ComputeHash([]byte("old")), RemoteHash: ComputeHash(remote)},
	}}
	if s.DetectConflict("BEN-1", []byte("new"), remote) != ConflictLocalOnly {
		t.Error("expected ConflictLocalOnly")
	}
}

func TestDetectConflictRemoteOnly(t *testing.T) {
	local := []byte("content")
	s := &State{Nodes: map[string]NodeState{
		"BEN-1": {LocalHash: ComputeHash(local), RemoteHash: ComputeHash([]byte("old"))},
	}}
	if s.DetectConflict("BEN-1", local, []byte("new")) != ConflictRemoteOnly {
		t.Error("expected ConflictRemoteOnly")
	}
}

func TestDetectConflictBoth(t *testing.T) {
	s := &State{Nodes: map[string]NodeState{
		"BEN-1": {LocalHash: ComputeHash([]byte("old-local")), RemoteHash: ComputeHash([]byte("old-remote"))},
	}}
	if s.DetectConflict("BEN-1", []byte("new-local"), []byte("new-remote")) != ConflictBoth {
		t.Error("expected ConflictBoth")
	}
}

func TestDetectConflictNeverSynced(t *testing.T) {
	s := &State{Nodes: make(map[string]NodeState)}
	if s.DetectConflict("BEN-1", []byte("local"), []byte("remote")) != ConflictNone {
		t.Error("expected ConflictNone for never-synced node")
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
