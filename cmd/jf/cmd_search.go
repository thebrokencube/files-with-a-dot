package main

import (
	"flag"
	"fmt"
	"jf/internal/pipeline"
	"os"
	"strings"
)

func runSearch(args []string) int {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	project := fs.String("project", "", "Filter by Jira project key")
	issueType := fs.String("type", "", "Filter by issue type (Epic, Story, Task, etc.)")
	limit := fs.Int("limit", 50, "Maximum results")

	if err := parseFlags(fs, args); err != nil {
		return 1
	}

	query := fs.Args()
	if len(query) == 0 && *project == "" {
		fmt.Fprintln(os.Stderr, "Usage: jf search <text> [--project KEY] [--type TYPE]")
		return 1
	}

	jql := buildSearchJQL(query, *project, *issueType)

	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}
	out, err := p.Search(jql, "summary,issuetype,status", *limit, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ search failed: %s\n", err)
		return 2
	}

	fmt.Print(string(out))
	return 0
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
