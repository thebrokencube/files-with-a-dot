package setup

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/config"
)

// CheckResult represents the outcome of a single prerequisite check.
type CheckResult struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "ok" | "missing" | "outdated"
	Detail string `json:"detail"`
	Fix    string `json:"fix"`
}

// Checker runs an external command and returns its output.
// Injected for testability.
type Checker func(name string, args ...string) (string, error)

// DefaultChecker shells out via os/exec.
func DefaultChecker(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	return strings.TrimSpace(string(out)), err
}

// CheckAll runs all prerequisite checks and returns results + overall pass.
func CheckAll(check Checker) ([]CheckResult, bool) {
	var results []CheckResult
	allOk := true

	results = append(results, checkNode(check))
	results = append(results, checkAcli(check))
	results = append(results, checkJiraAuth(check))
	results = append(results, checkConfig())

	for _, r := range results {
		if r.Status != "ok" {
			allOk = false
		}
	}

	return results, allOk
}

func checkNode(check Checker) CheckResult {
	version, err := check("node", "--version")
	if err != nil {
		return CheckResult{
			Name:   "node",
			Status: "missing",
			Detail: "Node.js not found",
			Fix:    "brew install node",
		}
	}
	return CheckResult{
		Name:   "node",
		Status: "ok",
		Detail: version,
	}
}

func checkAcli(check Checker) CheckResult {
	version, err := check("acli", "--version")
	if err != nil {
		return CheckResult{
			Name:   "acli",
			Status: "missing",
			Detail: "acli not found",
			Fix:    "brew install acli",
		}
	}
	return CheckResult{
		Name:   "acli",
		Status: "ok",
		Detail: version,
	}
}

func checkJiraAuth(check Checker) CheckResult {
	// acli manages its own OAuth auth — verify it can reach Jira
	out, err := check("acli", "jira", "project", "list", "--limit", "1")
	if err != nil {
		return CheckResult{
			Name:   "jira-auth",
			Status: "missing",
			Detail: "acli cannot reach Jira",
			Fix:    "Run: acli auth login",
		}
	}
	if strings.Contains(out, "unauthorized") || strings.Contains(out, "Unauthorized") {
		return CheckResult{
			Name:   "jira-auth",
			Status: "missing",
			Detail: "acli auth expired",
			Fix:    "Run: acli auth login",
		}
	}
	return CheckResult{
		Name:   "jira-auth",
		Status: "ok",
		Detail: "acli authenticated",
	}
}

// QuickCheck runs checks and returns a one-line error if anything fails.
// Returns empty string if all checks pass.
func QuickCheck(check Checker) string {
	results, ok := CheckAll(check)
	if ok {
		return ""
	}

	for _, r := range results {
		if r.Status != "ok" {
			return fmt.Sprintf("✗ %s. Run: jf setup", r.Detail)
		}
	}
	return ""
}

func checkConfig() CheckResult {
	cfg, err := config.Load()
	if err != nil {
		return CheckResult{
			Name:   "config",
			Status: "missing",
			Detail: fmt.Sprintf("cannot read ~/.jf.yml: %s", err),
			Fix:    "Check file permissions",
		}
	}

	if cfg.Site == "" {
		return CheckResult{
			Name:   "config",
			Status: "missing",
			Detail: "site not configured in ~/.jf.yml",
			Fix:    "Run: jf setup --discover",
		}
	}

	return CheckResult{
		Name:   "config",
		Status: "ok",
		Detail: cfg.Site + ".atlassian.net",
	}
}
