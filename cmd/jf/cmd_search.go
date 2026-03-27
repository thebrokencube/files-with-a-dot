package main

import (
	"fmt"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
	"os"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runSearch(args []string) int {
	fs := dendrik.NewFlagSet("search")
	project := fs.String('p', "project", "", "Filter by Jira project key")
	issueType := fs.String('t', "type", "", "Filter by issue type (Epic, Story, Task, etc.)")
	limit := fs.Int('l', "limit", 50, "Maximum results")
	jsonOut := fs.Bool('j', "json", "Output raw JSON from Jira")

	if done, code := dendrik.ParseCheck(fs, args); done {
		return code
	}

	query := fs.GetArgs()
	if len(query) == 0 && *project == "" {
		fmt.Fprintln(os.Stderr, "Usage: jf search <text> [--project KEY] [--type TYPE]")
		return dendrik.ExitUserError
	}

	jql := buildSearchJQL(query, *project, *issueType)

	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}
	out, err := p.Search(jql, "summary,issuetype,status", *limit, *jsonOut)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ search failed: %s\n", err)
		return dendrik.ExitExternalErr
	}

	fmt.Print(string(out))
	return dendrik.ExitOK
}

func buildSearchJQL(text []string, project, issueType string) string {
	var parts []string
	if len(text) > 0 {
		parts = append(parts, fmt.Sprintf("text ~ %q", strings.Join(text, " ")))
	}
	if project != "" {
		parts = append(parts, fmt.Sprintf("project = %q", project))
	}
	if issueType != "" {
		parts = append(parts, fmt.Sprintf("issuetype = %q", issueType))
	}
	return strings.Join(parts, " AND ")
}
