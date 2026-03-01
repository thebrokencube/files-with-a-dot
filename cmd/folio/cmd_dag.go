package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/graph"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/maputil"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
)

type dagNode struct {
	ID      string   `json:"id"`
	Sources []string `json:"sources"`
	Outputs []string `json:"outputs"`
}

type dagJSON struct {
	Nodes []dagNode `json:"nodes"`
	Edges []dagEdge `json:"edges"`
}

type dagEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func runDag(args []string) int {
	fs := flag.NewFlagSet("dag", flag.ExitOnError)
	folioPath := fs.String("folio", "./folio.yml", "Path to folio.yml")
	jsonMode := fs.Bool("json", false, "Machine-readable JSON output")
	noColor := fs.Bool("no-color", false, "Disable colored output")
	fs.Parse(args)

	if _, err := os.Stat(*folioPath); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, output.Errf("folio.yml not found at %s", *folioPath))
		return 1
	}

	f, err := config.Load(*folioPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	outputMap := graph.BuildOutputMap(f)
	producerMap := graph.SingleProducerMap(outputMap)
	inferred := graph.InferEdges(f, producerMap)
	merged := graph.MergeEdges(f, inferred)

	allTargets := maputil.SortedKeys(f.Targets)

	if *jsonMode {
		dj := dagJSON{}
		for _, tid := range allTargets {
			node := dagNode{ID: tid}
			target := f.Targets[tid]
			for _, s := range target.Sources {
				node.Sources = append(node.Sources, output.SourceLabel(s))
			}
			for _, o := range target.Outputs {
				node.Outputs = append(node.Outputs, output.OutputLabel(o))
			}
			if node.Sources == nil {
				node.Sources = []string{}
			}
			if node.Outputs == nil {
				node.Outputs = []string{}
			}
			dj.Nodes = append(dj.Nodes, node)
		}
		if dj.Nodes == nil {
			dj.Nodes = []dagNode{}
		}
		for _, tid := range allTargets {
			for _, dep := range merged[tid] {
				dj.Edges = append(dj.Edges, dagEdge{From: tid, To: dep})
			}
		}
		if dj.Edges == nil {
			dj.Edges = []dagEdge{}
		}
		data, err := json.Marshal(dj)
		if err != nil {
			fmt.Fprintf(os.Stderr, `{"error":"json marshal error: %s"}`, err)
			fmt.Fprintln(os.Stderr)
			return 1
		}
		fmt.Println(string(data))
	} else {
		output.PrintDAGTerminal(os.Stdout, f.Targets, merged, allTargets, !*noColor)
	}

	return 0
}
