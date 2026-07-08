package main

import (
	"fmt"
	"github.com/thebrokencube/files-with-a-dot/cmd/jf/internal/pipeline"
	"os"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runView(fields string, jsonOut bool, key string) int {
	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}
	out, err := p.View(key, fields, jsonOut)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ view failed: %s\n", err)
		return dendrik.ExitExternalErr
	}

	fmt.Print(string(out))
	return dendrik.ExitOK
}
