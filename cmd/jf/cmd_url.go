package main

import (
	"fmt"
	"os"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/config"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runURL(args []string) int {
	fs := dendrik.NewFlagSet("url")

	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	posArgs := fs.GetArgs()
	if len(posArgs) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: jf url <KEY>")
		return dendrik.ExitUserError
	}

	key := posArgs[0]

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ Failed to load config: %s\n", err)
		return dendrik.ExitUserError
	}

	url := cfg.BrowseURL(key)
	if url == "" {
		fmt.Fprintln(os.Stderr, "✗ Site not configured. Run: jf setup --discover")
		return dendrik.ExitUserError
	}

	fmt.Println(url)
	return dendrik.ExitOK
}
