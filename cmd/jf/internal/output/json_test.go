package output

import (
	"encoding/json"
	"testing"
)

func TestNodeInfoMarshal(t *testing.T) {
	info := NodeInfo{
		Key:      "TEST-1",
		Label:    "My Task",
		Type:     "Story",
		Sync:     "push",
		File:     "task.md",
		Children: 0,
		Status:   "stale",
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatal(err)
	}

	// Round-trip
	var got NodeInfo
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}

	if got.Key != "TEST-1" {
		t.Errorf("key: got %q, want %q", got.Key, "TEST-1")
	}
	if got.Status != "stale" {
		t.Errorf("status: got %q, want %q", got.Status, "stale")
	}
}

func TestNodeInfoOmitsEmpty(t *testing.T) {
	info := NodeInfo{
		Key:  "TEST-1",
		File: "task.md",
	}

	data, _ := json.Marshal(info)
	s := string(data)

	// parent should be omitted (empty string)
	if containsStr(s, `"parent"`) {
		t.Errorf("expected parent to be omitted, got %s", s)
	}
}

func TestStatusResultMarshal(t *testing.T) {
	r := StatusResult{
		Forest:    "/tmp/test",
		Total:     5,
		TBD:       1,
		PushTotal: 3,
		PushStale: 2,
		PullTotal: 1,
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}

	var got StatusResult
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}

	if got.Total != 5 {
		t.Errorf("total: got %d, want %d", got.Total, 5)
	}
	if got.PushStale != 2 {
		t.Errorf("push_stale: got %d, want %d", got.PushStale, 2)
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
