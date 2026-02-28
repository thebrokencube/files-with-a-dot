package graph

import (
	"testing"
)

func TestPropagateStalenessClean(t *testing.T) {
	statuses := map[string]string{
		"a": "clean",
		"b": "clean",
	}
	adj := map[string][]string{
		"b": {"a"},
	}
	result, causedBy := PropagateStaleness(statuses, adj)
	if result["a"] != "clean" {
		t.Errorf("a = %q, want clean", result["a"])
	}
	if result["b"] != "clean" {
		t.Errorf("b = %q, want clean", result["b"])
	}
	if len(causedBy) != 0 {
		t.Errorf("causedBy should be empty, got %v", causedBy)
	}
}

func TestPropagateStalenessUpstreamStale(t *testing.T) {
	statuses := map[string]string{
		"a": "stale",
		"b": "clean",
	}
	adj := map[string][]string{
		"b": {"a"},
	}
	result, causedBy := PropagateStaleness(statuses, adj)
	if result["b"] != "stale" {
		t.Errorf("b = %q, want stale (propagated from a)", result["b"])
	}
	if causedBy["b"] != "a" {
		t.Errorf("causedBy[b] = %q, want a", causedBy["b"])
	}
}

func TestPropagateStalenessUpstreamMissing(t *testing.T) {
	statuses := map[string]string{
		"a": "missing",
		"b": "clean",
	}
	adj := map[string][]string{
		"b": {"a"},
	}
	result, _ := PropagateStaleness(statuses, adj)
	if result["b"] != "stale" {
		t.Errorf("b = %q, want stale (propagated from missing a)", result["b"])
	}
}

func TestPropagateStalenessTransitive(t *testing.T) {
	// a -> b -> c, a is stale, both b and c should become stale
	statuses := map[string]string{
		"a": "stale",
		"b": "clean",
		"c": "clean",
	}
	adj := map[string][]string{
		"b": {"a"},
		"c": {"b"},
	}
	result, _ := PropagateStaleness(statuses, adj)
	if result["b"] != "stale" {
		t.Errorf("b = %q, want stale", result["b"])
	}
	if result["c"] != "stale" {
		t.Errorf("c = %q, want stale (transitive)", result["c"])
	}
}

func TestPropagateStalenessUnknown(t *testing.T) {
	statuses := map[string]string{
		"a": "unknown",
		"b": "clean",
	}
	adj := map[string][]string{
		"b": {"a"},
	}
	result, _ := PropagateStaleness(statuses, adj)
	if result["b"] != "stale" {
		t.Errorf("b = %q, want stale (upstream unknown)", result["b"])
	}
}

func TestPropagateStalenessAlreadyStale(t *testing.T) {
	statuses := map[string]string{
		"a": "stale",
		"b": "stale",
	}
	adj := map[string][]string{
		"b": {"a"},
	}
	result, _ := PropagateStaleness(statuses, adj)
	if result["b"] != "stale" {
		t.Errorf("b = %q, want stale", result["b"])
	}
}
