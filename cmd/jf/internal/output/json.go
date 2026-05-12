package output

// NodeResult is the per-node output for push/pull/sync operations.
type NodeResult struct {
	Node   string `json:"node"`
	File   string `json:"file,omitempty"`
	Status string `json:"status"` // "ok" | "error" | "skipped"
	Error  string `json:"error,omitempty"`
	Detail string `json:"detail,omitempty"`
	Size   int    `json:"size,omitempty"`
}

// NodeInfo is the JSON representation of a forest node for list/show/status.
type NodeInfo struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Sync     string `json:"sync"`
	File     string `json:"file"`
	Parent   string `json:"parent,omitempty"`
	Children int    `json:"children"`
	Status   string `json:"status,omitempty"` // "stale" | "clean" | "unknown"
}

// PRBadge is the per-PR output for status enrichment.
type PRBadge struct {
	Key    string `json:"key"` // Jira key this PR is associated with
	Number int    `json:"number"`
	State  string `json:"state"`            // draft | open | merged | closed
	CI     string `json:"ci"`               // pass | fail | pending | none
	Branch string `json:"branch,omitempty"` // headRefName
}

// StatusResult is the JSON output for jf status.
type StatusResult struct {
	Forest    string    `json:"forest"`
	Total     int       `json:"total"`
	TBD       int       `json:"tbd"`
	PushTotal int       `json:"push_total"`
	PushStale int       `json:"push_stale"`
	PullTotal int       `json:"pull_total"`
	PullStale int       `json:"pull_stale"`
	Mutable   int       `json:"mutable"`
	ReadOnly  int       `json:"read_only"`
	Empty     int       `json:"empty"`
	Repos     []string  `json:"repos,omitempty"`
	PRs       []PRBadge `json:"prs,omitempty"`
}

// ValidateResult is the JSON output for jf validate.
type ValidateResult struct {
	Valid  bool            `json:"valid"`
	Nodes  int             `json:"nodes"`
	Issues []ValidateIssue `json:"issues,omitempty"`
}

// ValidateIssue is a single validation issue.
type ValidateIssue struct {
	Level   string `json:"level"` // "error" | "warning"
	Message string `json:"message"`
}
