package main

import (
	"flag"
	"fmt"
	"jf/internal/pipeline"
	"os"
)

func runPush(args []string) int {
	fs := flag.NewFlagSet("push", flag.ContinueOnError)
	force := fs.Bool("force", false, "Push as plain text if marklassian conversion fails")

	if err := fs.Parse(args); err != nil {
		return 1
	}

	positional := fs.Args()
	if len(positional) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: jf push <KEY> <FILE> [--force]\n")
		return 1
	}

	key := positional[0]
	filePath := positional[1]

	source, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s: file not found\n", filePath)
		return 1
	}

	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}

	compiled, err := p.Compile(key, source)
	if err != nil {
		if *force {
			fmt.Fprintf(os.Stderr, "⚠ %s: conversion failed, pushing as plain text\n", key)
			compiled = buildPlainTextPayload(key, source)
		} else {
			fmt.Fprintf(os.Stderr, "✗ %s: marklassian conversion failed\n  %s\n", key, err)
			return 1
		}
	}

	if err := p.Push(compiled); err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s: acli error\n  %s\n", key, err)
		return 2
	}

	fmt.Printf("✓ Pushed %s description (%d bytes)\n", key, len(source))
	return 0
}

func buildPlainTextPayload(key string, source []byte) []byte {
	stripped := pipeline.StripFrontmatter(source)
	return []byte(fmt.Sprintf(`{"issues":[%q],"description":{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":%q}]}]}}`,
		key, string(stripped)))
}
