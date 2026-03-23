package main

import "testing"

func TestRunSetupCheck(t *testing.T) {
	// --check mode should return 0 or 1 based on prereqs, not panic
	code := runSetup([]string{"--check"})
	// We can't predict the result (depends on env), but it should not be > 1
	if code > 1 {
		t.Fatalf("expected exit 0 or 1, got %d", code)
	}
}

func TestRunSetupCheckJSON(t *testing.T) {
	code := runSetup([]string{"--check", "--json"})
	if code > 1 {
		t.Fatalf("expected exit 0 or 1, got %d", code)
	}
}
