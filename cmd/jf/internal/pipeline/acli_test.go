package pipeline

import (
	"encoding/json"
	"fmt"
	"os/exec"
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

func TestCompileProducesValidJSON(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not found, skipping integration test")
	}
	p := &Pipeline{Run: func(string, ...string) ([]byte, error) { return nil, nil }}
	out, err := p.Compile("TEST-1", []byte("## Hello"))
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	issues := parsed["issues"].([]any)
	if len(issues) != 1 || issues[0] != "TEST-1" {
		t.Errorf("expected issues [TEST-1], got %v", issues)
	}
	if parsed["description"] == nil {
		t.Error("expected description field")
	}
}

func TestCompileProducesADF(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not found, skipping integration test")
	}
	p := &Pipeline{Run: func(string, ...string) ([]byte, error) { return nil, nil }}
	out, err := p.Compile("KEY-1", []byte("## Heading\n\nParagraph"))
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatal(err)
	}
	desc := parsed["description"].(map[string]any)
	content := desc["content"].([]any)
	if len(content) != 2 {
		t.Fatalf("expected 2 content nodes, got %d", len(content))
	}
	first := content[0].(map[string]any)
	if first["type"] != "heading" {
		t.Errorf("expected heading, got %s", first["type"])
	}
}

func TestCompileStripsFrontmatter(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not found, skipping integration test")
	}
	p := &Pipeline{Run: func(string, ...string) ([]byte, error) { return nil, nil }}
	input := "---\ntitle: test\n---\n## Content"
	out, err := p.Compile("KEY-1", []byte(input))
	if err != nil {
		t.Fatal(err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatal(err)
	}
	desc := parsed["description"].(map[string]any)
	content := desc["content"].([]any)
	if len(content) != 1 {
		t.Fatalf("expected 1 content node after frontmatter strip, got %d", len(content))
	}
}

func TestPushCallsAcli(t *testing.T) {
	m := &mockRunner{output: []byte("ok")}
	p := &Pipeline{Run: m.run}

	err := p.Push([]byte(`{"issues":["TEST-1"]}`))
	if err != nil {
		t.Fatal(err)
	}

	if len(m.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(m.calls))
	}
	call := m.calls[0]
	if call[0] != "acli" {
		t.Errorf("expected acli, got %s", call[0])
	}
	joined := strings.Join(call[1:], " ")
	if !strings.Contains(joined, "jira workitem edit --from-json") {
		t.Errorf("unexpected args: %s", joined)
	}
	if !strings.Contains(joined, "--yes") {
		t.Errorf("expected --yes flag: %s", joined)
	}
}

func TestPushReturnsError(t *testing.T) {
	m := &mockRunner{output: []byte("permission denied"), err: fmt.Errorf("exit status 1")}
	p := &Pipeline{Run: m.run}

	err := p.Push([]byte(`{}`))
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "permission denied") {
		t.Errorf("expected acli output in error, got: %s", err)
	}
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
		jsonOut bool
		wantAll []string
	}{
		{"basic", "project=BEN", "", 0, false, []string{"--jql", "project=BEN"}},
		{"with fields", "project=BEN", "key,summary", 0, false, []string{"--fields", "key,summary"}},
		{"with limit", "project=BEN", "", 25, false, []string{"--limit", "25"}},
		{"all flags", "status=Open", "key", 10, true, []string{"--jql", "status=Open", "--fields", "key", "--limit", "10", "--json"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockRunner{output: []byte("result")}
			p := &Pipeline{Run: m.run}

			_, err := p.Search(tt.jql, tt.fields, tt.limit, tt.jsonOut)
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
