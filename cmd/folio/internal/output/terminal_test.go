package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/status"
)

func TestPrintStatusTerminalNoTree(t *testing.T) {
	ps := &status.ProjectStatus{
		Project: "Simple Project",
		Targets: map[string]status.TargetStatus{
			"summary": {
				Outputs: []status.OutputStatus{
					{Type: "local", Path: "compiled/summary.md", Status: "clean"},
				},
			},
		},
	}

	var buf bytes.Buffer
	PrintStatusTerminal(&buf, ps, nil, false)
	out := buf.String()

	if !strings.Contains(out, "Simple Project") {
		t.Error("expected project name in output")
	}
	if !strings.Contains(out, "clean") {
		t.Error("expected clean status in output")
	}
	// Should not have tree connectors
	if strings.Contains(out, "├──") || strings.Contains(out, "└──") {
		t.Error("unexpected tree connectors in non-tree output")
	}
}

func TestPrintStatusTerminalWithBatch(t *testing.T) {
	ps := &status.ProjectStatus{
		Project: "Batch Project",
		Targets: map[string]status.TargetStatus{
			"doc-tabs": {
				Outputs: []status.OutputStatus{
					{Type: "local", Path: "compiled/manifest.md", Status: "stale"},
				},
				BatchItems: []status.BatchItemStatus{
					{ID: "tab-1", Source: "tab1.md", System: "gdocs", ExtID: "doc-tab-1", Status: "clean"},
					{ID: "tab-2", Source: "tab2.md", System: "gdocs", ExtID: "doc-tab-2", Status: "stale"},
					{ID: "tab-3", Source: "tab3.md", System: "gdocs", ExtID: "doc-tab-3", Status: "missing"},
				},
			},
		},
	}

	var buf bytes.Buffer
	PrintStatusTerminal(&buf, ps, nil, false)
	out := buf.String()

	// Check batch items are rendered
	if !strings.Contains(out, "tab-1") {
		t.Error("expected tab-1 in output")
	}
	if !strings.Contains(out, "tab-2") {
		t.Error("expected tab-2 in output")
	}
	if !strings.Contains(out, "tab-3") {
		t.Error("expected tab-3 in output")
	}

	// Check system:ext_id format
	if !strings.Contains(out, "gdocs:doc-tab-1") {
		t.Error("expected gdocs:doc-tab-1 in output")
	}

	// Check connectors
	if !strings.Contains(out, "├──") {
		t.Error("expected ├── connector in batch output")
	}
	if !strings.Contains(out, "└──") {
		t.Error("expected └── connector for last batch item")
	}
}

func TestPrintStatusTerminalLifecycleHeader(t *testing.T) {
	ps := &status.ProjectStatus{
		Project: "Lifecycle Test",
		Lifecycle: status.LifecycleSummary{
			Observations: 3,
			Spikes:       2,
			Designs:      1,
			Plans:        1,
			Retros:       4,
			References:   5,
		},
		Targets: map[string]status.TargetStatus{},
	}

	var buf bytes.Buffer
	PrintStatusTerminal(&buf, ps, nil, false)
	out := buf.String()

	if !strings.Contains(out, "Lifecycle:") {
		t.Error("expected Lifecycle: header in output")
	}
	if !strings.Contains(out, "3 observations") {
		t.Errorf("expected '3 observations' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "2 spikes") {
		t.Errorf("expected '2 spikes' in output")
	}
}

func TestPrintStatusTerminalColor(t *testing.T) {
	ps := &status.ProjectStatus{
		Project: "Color Test",
		Targets: map[string]status.TargetStatus{
			"target": {
				Outputs: []status.OutputStatus{
					{Type: "local", Path: "out.md", Status: "stale"},
				},
			},
		},
	}

	var buf bytes.Buffer
	PrintStatusTerminal(&buf, ps, nil, true)
	out := buf.String()

	// Check ANSI codes are present when color=true
	if !strings.Contains(out, "\033[") {
		t.Error("expected ANSI color codes when color=true")
	}
}

func TestPrintStatusTerminalShowsFinalAndCause(t *testing.T) {
	ps := &status.ProjectStatus{
		Project: "Terminal Project",
		Targets: map[string]status.TargetStatus{
			"final-target": {
				Final: true,
				Outputs: []status.OutputStatus{
					{Type: "local", Path: "compiled/final.md", Status: "unknown", Cause: "freshness delegated to jf"},
				},
			},
		},
	}

	var buf bytes.Buffer
	PrintStatusTerminal(&buf, ps, nil, false)
	out := buf.String()
	if !strings.Contains(out, "final (terminal target)") {
		t.Errorf("expected terminal marker, got:\n%s", out)
	}
	if !strings.Contains(out, "cause: freshness delegated to jf") {
		t.Errorf("expected delegated cause, got:\n%s", out)
	}
}

func TestStatusJSONCarriesTargetFinalNotFifthOutputStatus(t *testing.T) {
	ps := &status.ProjectStatus{
		Project: "JSON Project",
		Targets: map[string]status.TargetStatus{
			"publish": {
				Final: true,
				Outputs: []status.OutputStatus{
					{Type: "local", Path: "compiled/review.md", Status: "clean"},
					{Type: "external", System: "jira", ID: "PROJ-1", Status: "unknown"},
				},
			},
		},
	}
	var buf bytes.Buffer
	PrintStatusJSON(&buf, ps)
	var envelope struct {
		Data struct {
			Targets map[string]struct {
				Final   bool             `json:"final"`
				Outputs []map[string]any `json:"outputs"`
			} `json:"targets"`
		} `json:"data"`
	}
	if err := json.Unmarshal(buf.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	target := envelope.Data.Targets["publish"]
	if !target.Final {
		t.Fatal("target final flag missing from JSON")
	}
	if len(target.Outputs) != 2 {
		t.Fatalf("output count = %d, want two", len(target.Outputs))
	}
	for _, output := range target.Outputs {
		if _, ok := output["final"]; ok {
			t.Fatal("final flag incorrectly serialized on output")
		}
	}
}
