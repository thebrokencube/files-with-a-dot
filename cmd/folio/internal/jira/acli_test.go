package jira

import (
	"fmt"
	"strings"
	"testing"
)

// mockRunner records calls and returns canned output.
type mockRunner struct {
	calls  [][]string
	output []byte
	err    error
}

func (m *mockRunner) run(name string, args ...string) ([]byte, error) {
	call := append([]string{name}, args...)
	m.calls = append(m.calls, call)
	return m.output, m.err
}

func TestCreateExtractsKey(t *testing.T) {
	m := &mockRunner{output: []byte("✓ Work item BEN-54321 created: https://jira.example.com/browse/BEN-54321")}
	p := &Pipeline{Run: m.run}

	key, err := p.Create([]byte(`{"projectKey":"BEN"}`))
	if err != nil {
		t.Fatal(err)
	}
	if key != "BEN-54321" {
		t.Errorf("expected BEN-54321, got %s", key)
	}
}

func TestCreateNonBenKey(t *testing.T) {
	m := &mockRunner{output: []byte("✓ Work item PROJ-123 created: https://jira.example.com")}
	p := &Pipeline{Run: m.run}

	key, err := p.Create([]byte(`{}`))
	if err != nil {
		t.Fatal(err)
	}
	if key != "PROJ-123" {
		t.Errorf("expected PROJ-123, got %s", key)
	}
}

func TestCreateNoKey(t *testing.T) {
	m := &mockRunner{output: []byte("unexpected output without a key")}
	p := &Pipeline{Run: m.run}

	_, err := p.Create([]byte(`{}`))
	if err == nil {
		t.Fatal("expected error when no key in output")
	}
	if !strings.Contains(err.Error(), "unexpected output") {
		t.Errorf("expected full stdout in error, got: %s", err)
	}
}

func TestCreateReturnsErrorOnFailure(t *testing.T) {
	m := &mockRunner{output: []byte("auth failed"), err: fmt.Errorf("exit status 1")}
	p := &Pipeline{Run: m.run}

	_, err := p.Create([]byte(`{}`))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "auth failed") {
		t.Errorf("expected acli output in error, got: %s", err)
	}
}

func TestViewBuildsArgs(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		fields  string
		jsonOut bool
		wantAll []string
	}{
		{"basic", "BEN-123", "", false, []string{"jira", "workitem", "view", "BEN-123"}},
		{"with fields", "BEN-123", "summary,status", false, []string{"--fields", "summary,status"}},
		{"with json", "BEN-123", "", true, []string{"--json"}},
		{"all flags", "BEN-123", "key", true, []string{"--fields", "key", "--json"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockRunner{output: []byte("result")}
			p := &Pipeline{Run: m.run}

			_, err := p.View(tt.id, tt.fields, tt.jsonOut)
			if err != nil {
				t.Fatal(err)
			}

			call := m.calls[0]
			joined := strings.Join(call, " ")
			for _, want := range tt.wantAll {
				if !strings.Contains(joined, want) {
					t.Errorf("expected %q in args: %s", want, joined)
				}
			}
		})
	}
}

func TestSearchBuildsArgs(t *testing.T) {
	tests := []struct {
		name    string
		jql     string
		fields  string
		limit   int
		wantAll []string
	}{
		{"basic", "project=BEN", "", 0, []string{"--jql", "project=BEN"}},
		{"with fields", "project=BEN", "key,summary", 0, []string{"--fields", "key,summary"}},
		{"with limit", "project=BEN", "", 25, []string{"--limit", "25"}},
		{"all flags", "status=Open", "key", 10, []string{"--jql", "status=Open", "--fields", "key", "--limit", "10"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockRunner{output: []byte("result")}
			p := &Pipeline{Run: m.run}

			_, err := p.Search(tt.jql, tt.fields, tt.limit)
			if err != nil {
				t.Fatal(err)
			}

			call := m.calls[0]
			joined := strings.Join(call, " ")
			for _, want := range tt.wantAll {
				if !strings.Contains(joined, want) {
					t.Errorf("expected %q in args: %s", want, joined)
				}
			}
		})
	}
}
