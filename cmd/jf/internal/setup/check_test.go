package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTestConfig points $HOME at a temp dir holding a minimal valid ~/.jf.yml so
// the config check is hermetic — independent of the host's real config (and CI,
// which has none). config.Load resolves the path via os.UserHomeDir() → $HOME.
func writeTestConfig(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	if err := os.WriteFile(filepath.Join(home, ".jf.yml"), []byte("site: example\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// mockChecker routes responses by command name. For "acli" calls, it also
// checks the first arg to distinguish version checks from jira auth checks.
func mockChecker(responses map[string]string) Checker {
	return func(name string, args ...string) (string, error) {
		// Try name + first arg (e.g., "acli jira" for auth check)
		if len(args) > 0 {
			compound := name + " " + args[0]
			if out, ok := responses[compound]; ok {
				return out, nil
			}
		}
		if out, ok := responses[name]; ok {
			return out, nil
		}
		return "", fmt.Errorf("command not found: %s", name)
	}
}

func TestCheckAllPasses(t *testing.T) {
	writeTestConfig(t)
	check := mockChecker(map[string]string{
		"node":      "v20.11.0",
		"acli":      "acli version 2.7.0",
		"acli jira": "KEY-1  Some issue",
	})

	results, ok := CheckAll(check)
	if !ok {
		for _, r := range results {
			if r.Status != "ok" {
				t.Logf("  %s: %s (%s)", r.Name, r.Status, r.Detail)
			}
		}
		t.Fatal("expected all checks to pass")
	}
	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Status != "ok" {
			t.Errorf("check %s: expected ok, got %s", r.Name, r.Status)
		}
	}
}

func TestCheckAllNodeMissing(t *testing.T) {
	check := mockChecker(map[string]string{
		"acli":      "acli version 2.7.0",
		"acli jira": "KEY-1  Some issue",
	})

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

func TestCheckAllJiraAuthFails(t *testing.T) {
	check := mockChecker(map[string]string{
		"node":           "v20.11.0",
		"acli --version": "acli version 2.7.0",
		// No "acli jira" entry — search call returns error
	})

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
	if !strings.Contains(authResult.Fix, "acli auth login") {
		t.Errorf("jira-auth fix: got %q, want contains %q", authResult.Fix, "acli auth login")
	}
}

func TestCheckAllJiraAuthUnauthorized(t *testing.T) {
	check := mockChecker(map[string]string{
		"node":      "v20.11.0",
		"acli":      "acli version 2.7.0",
		"acli jira": "unauthorized: use 'acli auth login' to authenticate",
	})

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
	if authResult.Detail != "acli auth expired" {
		t.Errorf("jira-auth detail: got %q, want %q", authResult.Detail, "acli auth expired")
	}
}

func TestQuickCheckAllOk(t *testing.T) {
	writeTestConfig(t)
	check := mockChecker(map[string]string{
		"node":      "v20.11.0",
		"acli":      "acli version 2.7.0",
		"acli jira": "KEY-1  Some issue",
	})

	msg := QuickCheck(check)
	if msg != "" {
		t.Errorf("expected empty string, got %q", msg)
	}
}

func TestQuickCheckFails(t *testing.T) {
	check := mockChecker(map[string]string{})

	msg := QuickCheck(check)
	if msg == "" {
		t.Fatal("expected error message")
	}
	want := "✗ Node.js not found. Run: jf setup"
	if msg != want {
		t.Errorf("got %q, want %q", msg, want)
	}
}
