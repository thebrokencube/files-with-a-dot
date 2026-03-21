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

func TestRecordPush(t *testing.T) {
	s := &State{Nodes: make(map[string]NodeState)}
	s.RecordPush("BEN-1")

	ns, ok := s.Nodes["BEN-1"]
	if !ok {
		t.Fatal("expected node state after RecordPush")
	}
	if time.Since(ns.LastPush) > time.Second {
		t.Error("expected LastPush to be recent")
	}
}
