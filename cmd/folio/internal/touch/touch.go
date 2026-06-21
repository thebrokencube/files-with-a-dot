package touch

import (
	"fmt"
	"os"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

// Target updates mtimes on all local output paths for the given target.
// Returns the number of files touched. folioDir is the directory containing folio.yml.
func Target(folioDir string, target *config.Target) (int, error) {
	now := time.Now()
	touched := 0

	reg, _ := config.LoadRegistry()

	for _, out := range target.Outputs {
		if out.Path == "" {
			continue
		}
		fullPath, err := config.ResolvePath(folioDir, out.Path, reg)
		if err != nil {
			return touched, fmt.Errorf("resolving %s: %w", out.Path, err)
		}
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return touched, fmt.Errorf("output file not found: %s", out.Path)
		}
		if err := os.Chtimes(fullPath, now, now); err != nil {
			return touched, fmt.Errorf("touching %s: %s", out.Path, err)
		}
		touched++
	}

	return touched, nil
}
