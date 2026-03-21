package main

import (
	"strings"
	"testing"
)

func TestBuildSearchJQLTextOnly(t *testing.T) {
	jql := buildSearchJQL([]string{"state", "retirement"}, "", "")
	want := `text ~ "state retirement"`
	if jql != want {
		t.Errorf("expected %q, got %q", want, jql)
	}
}

func TestBuildSearchJQLWithFilters(t *testing.T) {
	jql := buildSearchJQL([]string{"compliance"}, "BEN", "Epic")
	if jql == "" {
		t.Fatal("expected non-empty JQL")
	}
	for _, want := range []string{"text ~", "project =", "issuetype ="} {
		if !strings.Contains(jql, want) {
			t.Errorf("expected %q in JQL: %s", want, jql)
		}
	}
}

func TestBuildSearchJQLProjectOnly(t *testing.T) {
	jql := buildSearchJQL(nil, "BEN", "")
	want := `project = "BEN"`
	if jql != want {
		t.Errorf("expected %q, got %q", want, jql)
	}
}
