package list

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

// Entry represents a discovered folio in the repository.
type Entry struct {
	Section string // "active" or "archive"
	Path    string // relative path within section (e.g., "ben/state-retirement-mandates")
	Project string // project name from folio.yml
	Targets int    // number of targets
	Pending int    // number of pending items
}

// Scan walks active/ and archive/ under home, finds all folio.yml files,
// and returns a sorted list of entries.
func Scan(home string) ([]Entry, error) {
	var entries []Entry

	for _, section := range []string{"active", "archive"} {
		sectionDir := filepath.Join(home, section)
		if _, err := os.Stat(sectionDir); os.IsNotExist(err) {
			continue
		}

		err := filepath.WalkDir(sectionDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil // skip errors
			}
			if d.IsDir() || d.Name() != "folio.yml" {
				return nil
			}

			f, err := config.Load(path)
			if err != nil {
				return nil // skip unparseable files
			}

			rel, _ := filepath.Rel(sectionDir, filepath.Dir(path))

			entries = append(entries, Entry{
				Section: section,
				Path:    rel,
				Project: f.Project,
				Targets: len(f.Targets),
				Pending: len(f.Pending),
			})

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Section != entries[j].Section {
			return entries[i].Section < entries[j].Section // active before archive
		}
		return entries[i].Path < entries[j].Path
	})

	return entries, nil
}
