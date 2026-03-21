package forest

import (
	"strings"
	"testing"
)

func TestValidateClean(t *testing.T) {
	roots := []*Node{
		{Key: "BEN-1", Label: "Root", Type: "Epic", Sync: "push", File: "README.md"},
	}
	issues := Validate(roots, &Forest{})
	if len(issues) != 0 {
		t.Errorf("expected no issues, got %d: %v", len(issues), issues)
	}
}

func TestValidateDuplicateKey(t *testing.T) {
	roots := []*Node{
		{Key: "BEN-1", Label: "First", Type: "Story", Sync: "push", File: "first.md"},
		{Key: "BEN-1", Label: "Second", Type: "Story", Sync: "push", File: "second.md"},
	}
	issues := Validate(roots, &Forest{})
	hasError := false
	for _, iss := range issues {
		if iss.Level == "error" && strings.Contains(iss.Message, "duplicate key") {
			hasError = true
		}
	}
	if !hasError {
		t.Error("expected duplicate key error")
	}
}

func TestValidateTBDAllowsDuplicates(t *testing.T) {
	roots := []*Node{
		{Key: "TBD", Label: "First", Type: "Story", Sync: "push", File: "first.md"},
		{Key: "TBD", Label: "Second", Type: "Story", Sync: "push", File: "second.md"},
	}
	issues := Validate(roots, &Forest{})
	for _, iss := range issues {
		if iss.Level == "error" && strings.Contains(iss.Message, "duplicate key") {
			t.Error("TBD keys should not trigger duplicate error")
		}
	}
}

func TestValidateTBDMissingType(t *testing.T) {
	roots := []*Node{
		{Key: "TBD", Label: "No Type", Type: "", Sync: "push", File: "tbd.md"},
	}
	issues := Validate(roots, &Forest{})
	hasError := false
	for _, iss := range issues {
		if iss.Level == "error" && strings.Contains(iss.Message, "missing type") {
			hasError = true
		}
	}
	if !hasError {
		t.Error("expected TBD missing type error")
	}
}

func TestValidateTBDMissingLabel(t *testing.T) {
	roots := []*Node{
		{Key: "TBD", Label: "", Type: "Story", Sync: "push", File: "tbd.md"},
	}
	issues := Validate(roots, &Forest{})
	hasError := false
	for _, iss := range issues {
		if iss.Level == "error" && strings.Contains(iss.Message, "missing label") {
			hasError = true
		}
	}
	if !hasError {
		t.Error("expected TBD missing label error")
	}
}

func TestValidateInvalidSync(t *testing.T) {
	roots := []*Node{
		{Key: "BEN-1", Label: "Bad", Type: "Story", Sync: "bogus", File: "bad.md"},
	}
	issues := Validate(roots, &Forest{})
	hasError := false
	for _, iss := range issues {
		if iss.Level == "error" && strings.Contains(iss.Message, "invalid sync") {
			hasError = true
		}
	}
	if !hasError {
		t.Error("expected invalid sync error")
	}
}

func TestValidateSyncBoth(t *testing.T) {
	roots := []*Node{
		{Key: "BEN-1", Label: "Both", Type: "Story", Sync: "both", File: "both.md"},
	}
	issues := Validate(roots, &Forest{})
	for _, iss := range issues {
		if strings.Contains(iss.Message, "invalid sync") {
			t.Errorf("sync 'both' should be valid, got: %s", iss.Message)
		}
	}
}

func TestValidateValidSync(t *testing.T) {
	roots := []*Node{
		{Key: "BEN-1", Label: "Push", Type: "Story", Sync: "push", File: "a.md"},
		{Key: "BEN-2", Label: "Pull", Type: "Story", Sync: "pull", File: "b.md"},
		{Key: "BEN-3", Label: "Empty", Type: "Story", Sync: "", File: "c.md"},
	}
	issues := Validate(roots, &Forest{})
	for _, iss := range issues {
		if strings.Contains(iss.Message, "invalid sync") {
			t.Errorf("unexpected sync error: %s", iss.Message)
		}
	}
}

func TestValidateNestedDuplicate(t *testing.T) {
	child := &Node{Key: "BEN-2", Label: "Child", Type: "Story", Sync: "push", File: "epics/child.md"}
	roots := []*Node{
		{
			Key: "BEN-1", Label: "Root", Type: "Epic", Sync: "push", File: "README.md",
			Children: []*Node{
				child,
				{Key: "BEN-2", Label: "Dup", Type: "Story", Sync: "push", File: "epics/dup.md"},
			},
		},
	}
	issues := Validate(roots, &Forest{})
	hasError := false
	for _, iss := range issues {
		if iss.Level == "error" && strings.Contains(iss.Message, "duplicate key BEN-2") {
			hasError = true
		}
	}
	if !hasError {
		t.Error("expected duplicate key error for nested nodes")
	}
}

func TestValidateNegativeOrder(t *testing.T) {
	roots := []*Node{
		{Key: "BEN-1", Label: "Bad", Type: "Story", Sync: "push", Order: -1, File: "bad.md"},
	}
	issues := Validate(roots, &Forest{})
	hasError := false
	for _, iss := range issues {
		if iss.Level == "error" && strings.Contains(iss.Message, "invalid order") {
			hasError = true
		}
	}
	if !hasError {
		t.Error("expected invalid order error for negative order")
	}
}

func TestValidateZeroOrderValid(t *testing.T) {
	roots := []*Node{
		{Key: "BEN-1", Label: "Ok", Type: "Story", Sync: "push", Order: 0, File: "ok.md"},
	}
	issues := Validate(roots, &Forest{})
	for _, iss := range issues {
		if strings.Contains(iss.Message, "invalid order") {
			t.Errorf("unexpected order error for Order=0: %s", iss.Message)
		}
	}
}

func TestValidateIssueString(t *testing.T) {
	iss := ValidationIssue{Level: "error", File: "test.md", Message: "bad thing"}
	s := iss.String()
	if !strings.Contains(s, "✗") {
		t.Errorf("error should use ✗, got %q", s)
	}

	warn := ValidationIssue{Level: "warning", File: "", Message: "minor thing"}
	s = warn.String()
	if !strings.Contains(s, "⚠") {
		t.Errorf("warning should use ⚠, got %q", s)
	}
}
