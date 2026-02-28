package config

import "testing"

func TestResolveItemOutput(t *testing.T) {
	b := &Batch{
		System: "gdocs",
		Field:  "body",
		Items: []BatchItem{
			{ID: "tab-1", Output: Output{ID: "doc-tab-1"}},
		},
	}

	out := b.ResolveItemOutput(b.Items[0])
	if out.External != "gdocs" {
		t.Errorf("external = %q, want gdocs (inherited from batch)", out.External)
	}
	if out.Field != "body" {
		t.Errorf("field = %q, want body (inherited from batch)", out.Field)
	}
	if out.ID != "doc-tab-1" {
		t.Errorf("id = %q, want doc-tab-1 (from item)", out.ID)
	}
}

func TestResolveItemOutputNoDefaults(t *testing.T) {
	b := &Batch{
		System: "gdocs",
		Field:  "body",
		Items: []BatchItem{
			{ID: "tab-1", Output: Output{External: "jira", ID: "PROJ-99", Field: "description"}},
		},
	}

	out := b.ResolveItemOutput(b.Items[0])
	if out.External != "jira" {
		t.Errorf("external = %q, want jira (explicit override)", out.External)
	}
	if out.Field != "description" {
		t.Errorf("field = %q, want description (explicit override)", out.Field)
	}
	if out.ID != "PROJ-99" {
		t.Errorf("id = %q, want PROJ-99 (from item)", out.ID)
	}
}
