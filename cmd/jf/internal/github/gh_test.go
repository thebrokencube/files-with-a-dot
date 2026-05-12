package github

import (
	"testing"
)

func strPtr(s string) *string { return &s }

func TestMatchPRs(t *testing.T) {
	tests := []struct {
		name     string
		prs      []PR
		key      string
		wantNums []int
	}{
		{
			name:     "title match [KEY]",
			prs:      []PR{{Number: 1, Title: "[BEN-123] fix stuff", HeadRefName: "unrelated"}},
			key:      "BEN-123",
			wantNums: []int{1},
		},
		{
			name:     "title match case-insensitive",
			prs:      []PR{{Number: 2, Title: "[ben-123] fix stuff", HeadRefName: "unrelated"}},
			key:      "BEN-123",
			wantNums: []int{2},
		},
		{
			name:     "branch prefix match",
			prs:      []PR{{Number: 3, Title: "fix stuff", HeadRefName: "ben-123-fix-stuff"}},
			key:      "BEN-123",
			wantNums: []int{3},
		},
		{
			name:     "hyphen collision guard BEN-1 vs BEN-10",
			prs:      []PR{{Number: 4, Title: "feature", HeadRefName: "ben-10-feature"}},
			key:      "BEN-1",
			wantNums: nil,
		},
		{
			name:     "hyphen collision correct BEN-10",
			prs:      []PR{{Number: 5, Title: "feature", HeadRefName: "ben-10-feature"}},
			key:      "BEN-10",
			wantNums: []int{5},
		},
		{
			name:     "no match",
			prs:      []PR{{Number: 6, Title: "unrelated PR", HeadRefName: "other-branch"}},
			key:      "BEN-123",
			wantNums: nil,
		},
		{
			name:     "empty PR list",
			prs:      nil,
			key:      "BEN-123",
			wantNums: nil,
		},
		{
			name: "multiple PRs for one key",
			prs: []PR{
				{Number: 7, Title: "[BEN-50] part 1", HeadRefName: "unrelated"},
				{Number: 8, Title: "part 2", HeadRefName: "ben-50-part-2"},
			},
			key:      "BEN-50",
			wantNums: []int{7, 8},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MatchPRs(tt.prs, tt.key)
			if len(got) != len(tt.wantNums) {
				t.Fatalf("got %d PRs, want %d", len(got), len(tt.wantNums))
			}
			for i, pr := range got {
				if pr.Number != tt.wantNums[i] {
					t.Errorf("got PR #%d at index %d, want #%d", pr.Number, i, tt.wantNums[i])
				}
			}
		})
	}
}

func TestDeriveCIStatus(t *testing.T) {
	tests := []struct {
		name   string
		checks []StatusCheckRun
		want   string
	}{
		{
			name:   "all success",
			checks: []StatusCheckRun{{Conclusion: strPtr("SUCCESS"), State: "COMPLETED"}},
			want:   "pass",
		},
		{
			name:   "any failure",
			checks: []StatusCheckRun{{Conclusion: strPtr("FAILURE"), State: "COMPLETED"}},
			want:   "fail",
		},
		{
			name:   "timed out",
			checks: []StatusCheckRun{{Conclusion: strPtr("TIMED_OUT"), State: "COMPLETED"}},
			want:   "fail",
		},
		{
			name:   "cancelled",
			checks: []StatusCheckRun{{Conclusion: strPtr("CANCELLED"), State: "COMPLETED"}},
			want:   "fail",
		},
		{
			name:   "null conclusion",
			checks: []StatusCheckRun{{Conclusion: nil, State: "COMPLETED"}},
			want:   "pending",
		},
		{
			name:   "in progress",
			checks: []StatusCheckRun{{Conclusion: nil, State: "IN_PROGRESS"}},
			want:   "pending",
		},
		{
			name:   "queued",
			checks: []StatusCheckRun{{Conclusion: nil, State: "QUEUED"}},
			want:   "pending",
		},
		{
			name: "mixed pass and pending",
			checks: []StatusCheckRun{
				{Conclusion: strPtr("SUCCESS"), State: "COMPLETED"},
				{Conclusion: nil, State: "IN_PROGRESS"},
			},
			want: "pending",
		},
		{
			name: "mixed pass and fail",
			checks: []StatusCheckRun{
				{Conclusion: strPtr("SUCCESS"), State: "COMPLETED"},
				{Conclusion: strPtr("FAILURE"), State: "COMPLETED"},
			},
			want: "fail",
		},
		{
			name:   "empty rollup",
			checks: nil,
			want:   "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeriveCIStatus(tt.checks)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDeriveState(t *testing.T) {
	tests := []struct {
		name string
		pr   PR
		want string
	}{
		{"open", PR{State: "OPEN"}, "open"},
		{"draft", PR{State: "OPEN", IsDraft: true}, "draft"},
		{"merged", PR{State: "MERGED"}, "merged"},
		{"closed", PR{State: "CLOSED"}, "closed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeriveState(tt.pr)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
