package graph

import (
	"testing"
)

func TestDetectCycleNone(t *testing.T) {
	adj := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {},
	}
	cycle := DetectCycle(adj)
	if cycle != nil {
		t.Errorf("expected no cycle, got %v", cycle)
	}
}

func TestDetectCycleSimple(t *testing.T) {
	adj := map[string][]string{
		"a": {"b"},
		"b": {"a"},
	}
	cycle := DetectCycle(adj)
	if cycle == nil {
		t.Fatal("expected cycle")
	}
	if len(cycle) < 2 {
		t.Fatalf("cycle too short: %v", cycle)
	}
	if cycle[0] != cycle[len(cycle)-1] {
		t.Errorf("cycle should start and end with same node: %v", cycle)
	}
}

func TestDetectCycleTriangle(t *testing.T) {
	adj := map[string][]string{
		"a": {"b"},
		"b": {"c"},
		"c": {"a"},
	}
	cycle := DetectCycle(adj)
	if cycle == nil {
		t.Fatal("expected cycle")
	}
	if len(cycle) < 3 {
		t.Fatalf("triangle cycle should have at least 3 nodes: %v", cycle)
	}
	if cycle[0] != cycle[len(cycle)-1] {
		t.Errorf("cycle should start and end with same node: %v", cycle)
	}
}

func TestDetectCycleSelfLoop(t *testing.T) {
	adj := map[string][]string{
		"a": {"a"},
	}
	cycle := DetectCycle(adj)
	if cycle == nil {
		t.Fatal("expected cycle for self-loop")
	}
	if len(cycle) != 2 || cycle[0] != "a" || cycle[1] != "a" {
		t.Errorf("expected [a a], got %v", cycle)
	}
}

func TestDetectCycleDisconnected(t *testing.T) {
	adj := map[string][]string{
		"a": {"b"},
		"c": {"d"},
	}
	cycle := DetectCycle(adj)
	if cycle != nil {
		t.Errorf("expected no cycle in disconnected graph, got %v", cycle)
	}
}

func TestDetectCycleEmpty(t *testing.T) {
	adj := map[string][]string{}
	cycle := DetectCycle(adj)
	if cycle != nil {
		t.Errorf("expected no cycle in empty graph, got %v", cycle)
	}
}

func TestDetectCycleDiamond(t *testing.T) {
	// Diamond shape: a -> b, a -> c, b -> d, c -> d (no cycle)
	adj := map[string][]string{
		"a": {"b", "c"},
		"b": {"d"},
		"c": {"d"},
	}
	cycle := DetectCycle(adj)
	if cycle != nil {
		t.Errorf("expected no cycle in diamond graph, got %v", cycle)
	}
}
