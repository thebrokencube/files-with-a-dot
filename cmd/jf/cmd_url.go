package main

import (
	"fmt"
	"os"

	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/config"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runURL(key string) int {
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
