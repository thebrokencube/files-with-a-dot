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

func TestValidateFileNameMismatch(t *testing.T) {
	roots := []*Node{
		{Key: "BEN-123", Label: "Misnamed", Type: "Story", Sync: "push", File: "descriptive-name.md"},
	}
	issues := Validate(roots, &Forest{})
	hasWarning := false
	for _, iss := range issues {
		if iss.Level == "warning" && strings.Contains(iss.Message, "rename to BEN-123.md") {
			hasWarning = true
		}
	}
	if !hasWarning {
		t.Error("expected warning for filename not matching key")
	}
}

func TestValidateFileNameMatchesKey(t *testing.T) {
	roots := []*Node{
		{Key: "BEN-123", Label: "Correct", Type: "Story", Sync: "push", File: "BEN-123.md"},
	}
	issues := Validate(roots, &Forest{})
	for _, iss := range issues {
		if strings.Contains(iss.Message, "rename to") {
			t.Errorf("unexpected rename warning for correctly named file: %s", iss.Message)
		}
	}
}

func TestValidateFileNameMatchesKeyCaseInsensitive(t *testing.T) {
	roots := []*Node{
		{Key: "BEN-123", Label: "Lowercase", Type: "Story", Sync: "push", File: "ben-123.md"},
	}
	issues := Validate(roots, &Forest{})
	for _, iss := range issues {
		if strings.Contains(iss.Message, "rename to") {
			t.Errorf("unexpected rename warning for case-insensitive match: %s", iss.Message)
		}
	}
}

func TestValidateREADMEExemptFromNameCheck(t *testing.T) {
	roots := []*Node{
		{Key: "BEN-1", Label: "Root", Type: "Epic", Sync: "push", File: "README.md"},
	}
	issues := Validate(roots, &Forest{})
	for _, iss := range issues {
		if strings.Contains(iss.Message, "rename to") {
			t.Errorf("README.md should be exempt from name check: %s", iss.Message)
		}
	}
}

func TestValidateNestedREADMEExempt(t *testing.T) {
	roots := []*Node{
		{Key: "BEN-1", Label: "Root", Type: "Epic", Sync: "push", File: "epics/README.md"},
	}
	issues := Validate(roots, &Forest{})
	for _, iss := range issues {
		if strings.Contains(iss.Message, "rename to") {
			t.Errorf("nested README.md should be exempt from name check: %s", iss.Message)
		}
	}
}

func TestValidateTBDExemptFromNameCheck(t *testing.T) {
	roots := []*Node{
		{Key: "TBD", Label: "Pending", Type: "Story", Sync: "push", File: "my-feature.md"},
	}
	issues := Validate(roots, &Forest{})
	for _, iss := range issues {
		if strings.Contains(iss.Message, "rename to") {
			t.Errorf("TBD nodes should be exempt from name check: %s", iss.Message)
		}
	}
}

func TestValidateFileNameInSubdir(t *testing.T) {
	roots := []*Node{
		{Key: "BEN-456", Label: "Nested", Type: "Story", Sync: "push", File: "epics/old-name.md"},
	}
	issues := Validate(roots, &Forest{})
	hasWarning := false
	for _, iss := range issues {
		if iss.Level == "warning" && strings.Contains(iss.Message, "rename to BEN-456.md") {
			hasWarning = true
		}
	}
	if !hasWarning {
		t.Error("expected warning for misnamed file in subdirectory")
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
