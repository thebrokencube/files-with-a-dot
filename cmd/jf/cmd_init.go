package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

const defaultForestYml = `schema: 1

defaults:
  sync: push
  type: Story
  project: %s
`

func runInit(args []string) int {
	fs := dendrik.NewFlagSet("init")
	project := fs.String('p', "project", "BEN", "Jira project key")
	dir := fs.String('d', "dir", ".", "Directory to create forest.yml in")

	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	path := filepath.Join(*dir, "forest.yml")

	if _, err := os.Stat(path); err == nil {
		fmt.Printf("⚠ forest.yml already exists at %s\n", path)
		return dendrik.ExitOK
	}

	content := fmt.Sprintf(defaultForestYml, *project)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Failed to create forest.yml: %s\n", err)
		return dendrik.ExitUserError
	}

	fmt.Printf("✓ Created forest.yml (project: %s)\n", *project)
	return dendrik.ExitOK
}
