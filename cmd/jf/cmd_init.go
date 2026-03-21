package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const defaultForestYml = `schema: 1

defaults:
  sync: push
  type: Story
  project: %s
`

func runInit(args []string) int {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)
	project := fs.String("project", "BEN", "Jira project key")
	dir := fs.String("dir", ".", "Directory to create forest.yml in")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	path := filepath.Join(*dir, "forest.yml")

	if _, err := os.Stat(path); err == nil {
		fmt.Printf("⚠ forest.yml already exists at %s\n", path)
		return 0
	}

	content := fmt.Sprintf(defaultForestYml, *project)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Failed to create forest.yml: %s\n", err)
		return 1
	}

	fmt.Printf("✓ Created forest.yml (project: %s)\n", *project)
	return 0
}
