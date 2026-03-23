package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/graph"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/maputil"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/output"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/status"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
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
	fs := dendrik.NewFlagSet("dag")
	folioPath := fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
	jsonMode := fs.Bool('j', "json", "Machine-readable JSON output")
	branches := fs.Bool('b', "branches", "Show branch topology")
	statusFlag := fs.Bool('s', "status", "Show staleness overlay (requires --branches)")
	noColor := fs.BoolLong("no-color", "Disable colored output")
	if err := dendrik.Parse(fs, args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}

	if !resolveOrDie(folioPath) {
		return 1
	}

	if *statusFlag && !*branches {
		fmt.Fprintln(os.Stderr, output.Errf("--status requires --branches"))
		return 1
	}

	if _, err := os.Stat(*folioPath); os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, output.Errf("folio.yml not found at %s", *folioPath))
		return 1
	}

	f, err := config.Load(*folioPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, output.Errf("%s", err))
		return 1
	}

	if *branches {
		bt := output.BuildBranchTopology(f.Targets)

		if *statusFlag {
			folioDir := filepath.Dir(*folioPath)
			_, causedBy := status.DeriveWithDAG(f, folioDir)
			output.AnnotateBranchStatus(bt, f, filepath.Dir(*folioPath), causedBy)
		}

		if *jsonMode {
			output.PrintBranchDAGJSON(os.Stdout, bt)
		} else {
			output.PrintBranchDAGFromTopology(os.Stdout, bt, !*noColor, *statusFlag)
		}
		return 0
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
