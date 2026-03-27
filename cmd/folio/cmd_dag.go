package main

import (
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
	pal := dendrik.NewPalette(true)
	fs := dendrik.NewFlagSet("dag")
	folioPath := fs.String('f', "folio", "./folio.yml", "Path or shortname (e.g., ben/my-project)")
	jsonMode := fs.Bool('j', "json", "Machine-readable JSON output")
	branches := fs.Bool('b', "branches", "Show branch topology")
	statusFlag := fs.Bool('s', "status", "Show staleness overlay (requires --branches)")
	noColor := fs.BoolLong("no-color", "Disable colored output")
	if done, code := dendrik.ParseCheck(fs, args); done {
		return code
	}

	if !resolveOrDie(folioPath) {
		return dendrik.ExitUserError
	}

	if *statusFlag && !*branches {
		if *jsonMode {
			dendrik.WriteError(os.Stdout, "--status requires --branches", "")
		} else {
			fmt.Fprintln(os.Stderr, pal.Errf("--status requires --branches"))
		}
		return dendrik.ExitUserError
	}

	if _, err := os.Stat(*folioPath); os.IsNotExist(err) {
		if *jsonMode {
			dendrik.WriteError(os.Stdout, fmt.Sprintf("folio.yml not found at %s", *folioPath), "")
		} else {
			fmt.Fprintln(os.Stderr, pal.Errf("folio.yml not found at %s", *folioPath))
		}
		return dendrik.ExitUserError
	}

	f, err := config.Load(*folioPath)
	if err != nil {
		if *jsonMode {
			dendrik.WriteError(os.Stdout, fmt.Sprintf("%s", err), "")
		} else {
			fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		}
		return dendrik.ExitUserError
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
		return dendrik.ExitOK
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
		dendrik.WriteResult(os.Stdout, dj)
	} else {
		output.PrintDAGTerminal(os.Stdout, f.Targets, merged, allTargets, !*noColor)
	}

	return dendrik.ExitOK
}
