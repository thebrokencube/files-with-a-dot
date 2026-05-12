package github

import (
	"encoding/json"
	"os/exec"
	"strings"
)

// PR represents a GitHub pull request returned by gh pr list.
type PR struct {
	Number            int              `json:"number"`
	State             string           `json:"state"` // OPEN, CLOSED, MERGED
	IsDraft           bool             `json:"isDraft"`
	HeadRefName       string           `json:"headRefName"`
	Title             string           `json:"title"`
	StatusCheckRollup []StatusCheckRun `json:"statusCheckRollup"`
}

// StatusCheckRun is a single CI check within a PR's statusCheckRollup.
type StatusCheckRun struct {
	Conclusion *string `json:"conclusion"` // SUCCESS, FAILURE, TIMED_OUT, CANCELLED, or null
	State      string  `json:"state"`      // COMPLETED, IN_PROGRESS, QUEUED
}

// Fetcher abstracts gh calls for testing. Searches across multiple repos.
type Fetcher func(repos []string, keys []string) ([]PR, error)

// DefaultFetcher shells out to gh pr list for each repo with a search query
// built from keys. Returns nil, nil if gh is not on PATH.
func DefaultFetcher(repos []string, keys []string) ([]PR, error) {
	if _, err := exec.LookPath("gh"); err != nil {
		return nil, nil
	}
	if len(keys) == 0 || len(repos) == 0 {
		return nil, nil
	}

	search := strings.Join(keys, " OR ")

	var allPRs []PR
	for _, repo := range repos {
		cmd := exec.Command("gh", "pr", "list",
			"--repo", repo,
			"--state", "all",
			"--search", search,
			"--json", "number,state,isDraft,headRefName,title,statusCheckRollup",
			"--limit", "200",
		)
		out, err := cmd.Output()
		if err != nil {
			continue // skip repos that fail (auth, not found, etc.)
		}

		var prs []PR
		if err := json.Unmarshal(out, &prs); err != nil {
			continue
		}
		allPRs = append(allPRs, prs...)
	}
	return allPRs, nil
}

// MatchPRs filters prs to those matching key via title prefix [KEY] or
// branch prefix key-lower-. Returns empty slice on no match.
func MatchPRs(prs []PR, key string) []PR {
	var matched []PR
	keyUpper := strings.ToUpper(key)
	branchPrefix := strings.ToLower(key) + "-"
	titlePrefix := "[" + keyUpper + "]"

	for _, pr := range prs {
		titleUpper := strings.ToUpper(pr.Title)
		if strings.HasPrefix(titleUpper, titlePrefix) {
			matched = append(matched, pr)
			continue
		}
		branch := strings.ToLower(pr.HeadRefName)
		if strings.HasPrefix(branch, branchPrefix) {
			matched = append(matched, pr)
		}
	}
	return matched
}

// DeriveCIStatus computes an aggregate CI status from a PR's statusCheckRollup.
//
//   - Any FAILURE/TIMED_OUT/CANCELLED → "fail"
//   - Any null conclusion or non-COMPLETED state → "pending"
//   - All SUCCESS → "pass"
//   - Empty rollup → "none"
func DeriveCIStatus(checks []StatusCheckRun) string {
	if len(checks) == 0 {
		return "none"
	}

	for _, c := range checks {
		if c.Conclusion != nil {
			switch *c.Conclusion {
			case "FAILURE", "TIMED_OUT", "CANCELLED":
				return "fail"
			}
		}
	}

	for _, c := range checks {
		if c.Conclusion == nil || c.State != "COMPLETED" {
			return "pending"
		}
	}

	return "pass"
}

// DeriveState maps gh's PR state + isDraft to a display state.
func DeriveState(pr PR) string {
	switch pr.State {
	case "MERGED":
		return "merged"
	case "CLOSED":
		return "closed"
	default: // OPEN
		if pr.IsDraft {
			return "draft"
		}
		return "open"
	}
}
