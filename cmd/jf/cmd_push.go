package main

import (
	"flag"
	"fmt"
	"jf/internal/forest"
	"jf/internal/pipeline"
	"os"
	"path/filepath"
)

func runPush(args []string) int {
	fs := flag.NewFlagSet("push", flag.ContinueOnError)
	force := fs.Bool("force", false, "Push as plain text if marklassian conversion fails")
	subtree := fs.String("subtree", "", "Push node and all descendants")
	failFast := fs.Bool("fail-fast", false, "Stop on first error")
	dir := fs.String("dir", ".", "Directory to scan for forest.yml")

	if err := parseFlags(fs, args); err != nil {
		return 1
	}

	positional := fs.Args()

	// Level 0: explicit key + file
	if len(positional) >= 2 {
		return pushSingle(positional[0], positional[1], *force)
	}

	// Forest mode: discover and push
	return pushForest(*dir, positional, *subtree, *force, *failFast)
}

func pushSingle(key, filePath string, force bool) int {
	source, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s: file not found\n", filePath)
		return 1
	}

	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}

	compiled, err := p.Compile(key, source, "")
	if err != nil {
		if force {
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

func pushForest(dir string, positional []string, subtreeTarget string, force, failFast bool) int {
	f, roots, err := loadForest(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s\n", err)
		fmt.Fprintf(os.Stderr, "  For Level 0: jf push <KEY> <FILE>\n")
		return 1
	}

	// Validate first
	issues := forest.Validate(roots, f)
	hasErrors := false
	for _, iss := range issues {
		if iss.Level == "error" {
			fmt.Fprintln(os.Stderr, iss.String())
			hasErrors = true
		}
	}
	if hasErrors {
		return 1
	}

	// Resolve target
	var pushRoots []*forest.Node
	target := subtreeTarget
	if target == "" && len(positional) == 1 {
		target = positional[0]
	}

	if target != "" {
		node, err := forest.Subtree(roots, target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s\n", err)
			return 1
		}
		pushRoots = []*forest.Node{node}
	} else {
		pushRoots = roots
	}

	// Post-order traversal for push
	ordered := forest.PostOrder(pushRoots)

	// Filter to sync:push nodes (skip TBD and sync:pull)
	var toPush []*forest.Node
	for _, n := range ordered {
		if forest.IsTBD(n.Key) {
			continue
		}
		if n.Sync == "pull" {
			continue
		}
		toPush = append(toPush, n)
	}

	if len(toPush) == 0 {
		fmt.Println("No push-mode nodes found.")
		return 0
	}

	// Load state for tracking
	state, err := forest.LoadState(f.Dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠ Could not load state: %s\n", err)
		state = &forest.State{Nodes: make(map[string]forest.NodeState)}
	}

	p := &pipeline.Pipeline{Run: pipeline.DefaultRunner}
	succeeded := 0
	failed := 0

	for _, n := range toPush {
		filePath := filepath.Join(f.Dir, n.File)
		source, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: %s\n", n.Key, err)
			failed++
			if failFast {
				break
			}
			continue
		}

		compiled, err := p.Compile(n.Key, source, n.Label)
		if err != nil {
			if force {
				fmt.Fprintf(os.Stderr, "⚠ %s: conversion failed, pushing as plain text\n", n.Key)
				compiled = buildPlainTextPayload(n.Key, source)
			} else {
				fmt.Fprintf(os.Stderr, "✗ %s: marklassian conversion failed\n  %s\n", n.Key, err)
				failed++
				if failFast {
					break
				}
				continue
			}
		}

		if err := p.Push(compiled); err != nil {
			fmt.Fprintf(os.Stderr, "✗ %s: acli error\n  %s\n", n.Key, err)
			failed++
			if failFast {
				break
			}
			continue
		}

		state.RecordPush(n.Key)
		fmt.Printf("✓ %s (%s)\n", n.Key, n.File)
		succeeded++
	}

	// Save state
	if err := forest.SaveState(f.Dir, state); err != nil {
		fmt.Fprintf(os.Stderr, "⚠ Could not save state: %s\n", err)
	}

	fmt.Printf("\nPushed %d/%d nodes", succeeded, succeeded+failed)
	if failed > 0 {
		fmt.Printf(" (%d failed)", failed)
	}
	fmt.Println()

	if failed > 0 {
		return 1
	}
	return 0
}

func buildPlainTextPayload(key string, source []byte) []byte {
	stripped := pipeline.StripFrontmatter(source)
	return []byte(fmt.Sprintf(`{"issues":[%q],"description":{"type":"doc","version":1,"content":[{"type":"paragraph","content":[{"type":"text","text":%q}]}]}}`,
		key, string(stripped)))
}
