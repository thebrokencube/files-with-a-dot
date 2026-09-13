package main

import (
	"fmt"
	"os"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/touch"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/validate"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runTouch(folioPath, targetID string, final bool) int {
	pal := dendrik.NewPalette(true)

	ctx, err := resolveContext(folioPath, contextMutation)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return dendrik.ExitUserError
	}
	folioPath = ctx.FolioPath

	f, err := config.Load(folioPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}

	target, ok := f.Targets[targetID]
	if !ok {
		fmt.Fprintln(os.Stderr, pal.Errf("target '%s' not found", targetID))
		return dendrik.ExitUserError
	}
	if target.Final {
		fmt.Fprintln(os.Stderr, pal.Errf("target '%s' is already final; remove final before touching it again", targetID))
		return dendrik.ExitUserError
	}
	if target.Forest != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("forest target '%s' freshness is delegated to jf; use jf status and jf push", targetID))
		return dendrik.ExitUserError
	}
	if !hasLocalOutput(target) {
		fmt.Fprintln(os.Stderr, pal.Errf("target '%s' has no local output paths", targetID))
		return dendrik.ExitUserError
	}

	before := validate.Validate(f, ctx, validate.Mutation)
	state, err := touch.Snapshot(ctx, &target, final, time.Now())
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}

	projected := projectedTouchFolio(f, targetID, state)
	after := validate.Validate(projected, ctx, validate.Mutation)
	delta := validate.Delta(before, after, nil)
	if !delta.Valid {
		fmt.Fprintln(os.Stderr, pal.Errf("touch refused: validation failed"))
		for _, validationErr := range delta.Errors {
			fmt.Fprintf(os.Stderr, "  - %s\n", validationErr)
		}
		return dendrik.ExitUserError
	}

	if err := touch.WriteState(folioPath, targetID, state); err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}

	if final {
		fmt.Println(pal.Successf("Finalized target %s (operator-confirmed local review copy)", targetID))
	} else {
		fmt.Println(pal.Successf("Recorded composition snapshot for %s", targetID))
	}
	return dendrik.ExitOK
}

func projectedTouchFolio(f *config.Folio, targetID string, state touch.State) *config.Folio {
	projected := *f
	projected.Targets = make(map[string]config.Target, len(f.Targets))
	for id, target := range f.Targets {
		projected.Targets[id] = target
	}
	target := projected.Targets[targetID]
	target.ComposedAt = state.ComposedAt
	target.InputsSHA256 = state.InputsSHA256
	target.Final = state.Final
	if target.Batch != nil {
		batch := *target.Batch
		batch.Items = append([]config.BatchItem(nil), target.Batch.Items...)
		for i := range batch.Items {
			batch.Items[i].InputsSHA256 = ""
		}
		for _, itemState := range state.Items {
			for i := range batch.Items {
				if batch.Items[i].ID == itemState.ID {
					batch.Items[i].InputsSHA256 = itemState.InputsSHA256
					break
				}
			}
		}
		target.Batch = &batch
	}
	projected.Targets[targetID] = target
	return &projected
}

func hasLocalOutput(target config.Target) bool {
	for _, output := range target.Outputs {
		if output.Path != "" {
			return true
		}
	}
	return false
}
