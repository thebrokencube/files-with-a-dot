package setup

import (
	"fmt"
	"testing"
)

func mockChecker(responses map[string]string) Checker {
	return func(name string, args ...string) (string, error) {
		key := name
		if out, ok := responses[key]; ok {
			return out, nil
		}
		return "", fmt.Errorf("command not found: %s", name)
	}
}

func TestCheckAllPasses(t *testing.T) {
	check := mockChecker(map[string]string{
		"node": "v20.11.0",
		"acli": "acli version 2.7.0",
	})

	t.Setenv("JIRA_API_TOKEN", "test-token")

	results, ok := CheckAll(check)
	if !ok {
		t.Fatal("expected all checks to pass")
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Status != "ok" {
			t.Errorf("check %s: expected ok, got %s", r.Name, r.Status)
		}
	}
}

func TestCheckAllNodeMissing(t *testing.T) {
	check := mockChecker(map[string]string{
		"acli": "acli version 2.7.0",
	})

	t.Setenv("JIRA_API_TOKEN", "test-token")

	results, ok := CheckAll(check)
	if ok {
		t.Fatal("expected check to fail")
	}

	var nodeResult CheckResult
	for _, r := range results {
		if r.Name == "node" {
			nodeResult = r
		}
	}
	if nodeResult.Status != "missing" {
		t.Errorf("node status: got %q, want %q", nodeResult.Status, "missing")
	}
	if nodeResult.Fix != "brew install node" {
		t.Errorf("node fix: got %q, want %q", nodeResult.Fix, "brew install node")
	}
}

func TestCheckAllAcliMissing(t *testing.T) {
	check := mockChecker(map[string]string{
		"node": "v20.11.0",
	})

	t.Setenv("JIRA_API_TOKEN", "test-token")

	results, ok := CheckAll(check)
	if ok {
		t.Fatal("expected check to fail")
	}

	var acliResult CheckResult
	for _, r := range results {
		if r.Name == "acli" {
			acliResult = r
		}
	}
	if acliResult.Status != "missing" {
		t.Errorf("acli status: got %q, want %q", acliResult.Status, "missing")
	}
}

func TestCheckAllJiraAuthMissing(t *testing.T) {
	check := mockChecker(map[string]string{
		"node": "v20.11.0",
		"acli": "acli version 2.7.0",
	})

	t.Setenv("JIRA_API_TOKEN", "")
	t.Setenv("JIRA_TOKEN", "")
	t.Setenv("ATLASSIAN_API_TOKEN", "")

	results, ok := CheckAll(check)
	if ok {
		t.Fatal("expected check to fail")
	}

	var authResult CheckResult
	for _, r := range results {
		if r.Name == "jira-auth" {
			authResult = r
		}
	}
	if authResult.Status != "missing" {
		t.Errorf("jira-auth status: got %q, want %q", authResult.Status, "missing")
	}
	if authResult.Fix != "Set JIRA_API_TOKEN in ~/.env.local" {
		t.Errorf("jira-auth fix: got %q, want %q", authResult.Fix, "Set JIRA_API_TOKEN in ~/.env.local")
	}
}

func TestCheckAllAlternativeToken(t *testing.T) {
	check := mockChecker(map[string]string{
		"node": "v20.11.0",
		"acli": "acli version 2.7.0",
	})

	t.Setenv("JIRA_API_TOKEN", "")
	t.Setenv("JIRA_TOKEN", "alt-token")

	results, ok := CheckAll(check)
	if !ok {
		t.Fatal("expected all checks to pass with alternative token")
	}

	var authResult CheckResult
	for _, r := range results {
		if r.Name == "jira-auth" {
			authResult = r
		}
	}
	if authResult.Status != "ok" {
		t.Errorf("jira-auth status: got %q, want %q", authResult.Status, "ok")
	}
}

func TestQuickCheckAllOk(t *testing.T) {
	check := mockChecker(map[string]string{
		"node": "v20.11.0",
		"acli": "acli version 2.7.0",
	})

	t.Setenv("JIRA_API_TOKEN", "test-token")

	msg := QuickCheck(check)
	if msg != "" {
		t.Errorf("expected empty string, got %q", msg)
	}
}

func TestQuickCheckFails(t *testing.T) {
	check := mockChecker(map[string]string{})

	t.Setenv("JIRA_API_TOKEN", "test-token")

	msg := QuickCheck(check)
	if msg == "" {
		t.Fatal("expected error message")
	}
	want := "✗ Node.js not found. Run: jf setup"
	if msg != want {
		t.Errorf("got %q, want %q", msg, want)
	}
}
