package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/touch"
	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runTouch(folioPath, targetID string) int {
	pal := dendrik.NewPalette(true)

	if !resolveOrDie(&folioPath) {
		return dendrik.ExitUserError
	}

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

	folioDir := filepath.Dir(folioPath)
	touched, err := touch.Target(folioDir, &target)
	if err != nil {
		fmt.Fprintln(os.Stderr, pal.Errf("%s", err))
		return dendrik.ExitUserError
	}

	if touched == 0 {
		fmt.Fprintln(os.Stderr, pal.Errf("target '%s' has no local output paths", targetID))
		return dendrik.ExitUserError
	}

	fmt.Println(pal.Successf("Touched %d output(s) for %s", touched, targetID))
	return dendrik.ExitOK
}
