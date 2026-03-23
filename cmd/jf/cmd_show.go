package main

import (
	"fmt"
	"jf/internal/forest"
	"os"
	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runShow(args []string) int {
	fs := dendrik.NewFlagSet("show")
	dir := fs.String('d', "dir", ".", "Directory to scan for forest.yml")
	jsonOut := fs.Bool('j', "json", "Output as JSON")

	if err := dendrik.Parse(fs, args); err != nil {
		return dendrik.ExitUserError
	}

	positional := fs.GetArgs()
	if len(positional) == 0 {
		fmt.Fprintf(os.Stderr, "Usage: jf show <target>\n")
		return dendrik.ExitUserError
	}

	f, roots, code := loadForestOrFail(*dir, false)
	if code != 0 {
		return code
	}

	node, err := forest.Resolve(roots, positional[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %s\n", err)
		return dendrik.ExitUserError
	}

	state, err := forest.LoadState(f.Dir)
	if err != nil {
		state = &forest.State{Nodes: make(map[string]forest.NodeState)}
	}

	staleStr := nodeStatus(node, f, state)

	if *jsonOut {
		info := nodeToInfo(node)
		info.Status = staleStr
		dendrik.WriteResult(os.Stdout, info)
		return dendrik.ExitOK
	}

	parentStr := "(root)"
	if node.Parent != nil {
		parentStr = node.Parent.Key
	}

	fmt.Printf("Key:      %s\n", node.Key)
	fmt.Printf("Label:    %s\n", node.Label)
	fmt.Printf("Type:     %s\n", node.Type)
	fmt.Printf("Sync:     %s\n", syncDisplay(node.Sync))
	fmt.Printf("File:     %s\n", node.File)
	fmt.Printf("Parent:   %s\n", parentStr)
	fmt.Printf("Children: %d\n", len(node.Children))
	fmt.Printf("Status:   %s\n", staleStr)

	if ns, ok := state.Nodes[node.Key]; ok && !ns.LastPush.IsZero() {
		fmt.Printf("Pushed:   %s\n", ns.LastPush.Format("2006-01-02 15:04:05"))
	}

	return dendrik.ExitOK
}

func syncDisplay(sync string) string {
	if sync == "pull" {
		return "pull ↓"
	}
	return "push ↑"
}

func nodeStatus(node *forest.Node, f *forest.Forest, state *forest.State) string {
	if forest.IsTBD(node.Key) {
		return "unknown"
	}
	filePath := filepath.Join(f.Dir, node.File)
	info, err := os.Stat(filePath)
	if err != nil {
		return "unknown"
	}
	if state.IsStale(node.Key, info.ModTime()) {
		return "stale"
	}
	return "clean"
}
