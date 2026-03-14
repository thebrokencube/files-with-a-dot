package output

import (
	"bytes"
	"strings"
	"testing"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/status"
)

func TestPrintStatusTerminalWithTree(t *testing.T) {
	ps := &status.ProjectStatus{
		Project: "Tree Project",
		Targets: map[string]status.TargetStatus{
			"initiative": {
				Outputs: []status.OutputStatus{
					{Type: "local", Path: "compiled/manifest.md", Status: "stale"},
				},
				TreeRoot: &status.TreeNodeStatus{
					ID:       "PROJ-1",
					Label:    "Initiative",
					Status:   "stale",
					CausedBy: "PROJ-10",
					Children: []status.TreeNodeStatus{
						{
							ID:     "PROJ-10",
							Label:  "Project A",
							Status: "stale",
							Children: []status.TreeNodeStatus{
								{ID: "PROJ-100", Label: "Epic 1", Status: "clean"},
							},
						},
						{ID: "PROJ-20", Label: "Project B", Status: "clean"},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	PrintStatusTerminal(&buf, ps, nil, false)
	out := buf.String()

	// Check tree nodes are rendered
	if !strings.Contains(out, "PROJ-1") {
		t.Error("expected PROJ-1 in output")
	}
	if !strings.Contains(out, "PROJ-10") {
		t.Error("expected PROJ-10 in output")
	}
	if !strings.Contains(out, "PROJ-100") {
		t.Error("expected PROJ-100 in output")
	}
	if !strings.Contains(out, "PROJ-20") {
		t.Error("expected PROJ-20 in output")
	}

	// Check labels are shown
	if !strings.Contains(out, "Initiative") {
		t.Error("expected Initiative label in output")
	}
	if !strings.Contains(out, "Project A") {
		t.Error("expected Project A label in output")
	}

	// Check tree connectors
	if !strings.Contains(out, "├──") && !strings.Contains(out, "└──") {
		t.Error("expected tree connectors in output")
	}

	// Check causedBy annotation
	if !strings.Contains(out, "<< PROJ-10") {
		t.Error("expected causedBy annotation '<< PROJ-10' in output")
	}
}

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
				TreeRoot: &status.TreeNodeStatus{
					ID:     "ROOT",
					Status: "stale",
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
