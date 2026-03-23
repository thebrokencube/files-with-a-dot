package main

import (
	"fmt"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
	"os"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runView(args []string) int {
	fs := dendrik.NewFlagSet("view")
	fields := fs.String('f', "fields", "summary,status,issuetype", "Comma-separated fields to display")
	jsonOut := fs.Bool('j', "json", "Output raw JSON from Jira")

	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	positional := fs.GetArgs()
	if len(positional) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: jf view <KEY> [--fields F] [--json]")
		return dendrik.ExitUserError
	}

	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}
	out, err := p.View(positional[0], *fields, *jsonOut)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ view failed: %s\n", err)
		return dendrik.ExitExternalErr
	}

	fmt.Print(string(out))
	return dendrik.ExitOK
}
