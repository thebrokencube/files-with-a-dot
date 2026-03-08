package health

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/taxonomy"
)

// Report contains the health analysis for a single folio project.
type Report struct {
	Project      string
	Reference    map[string]int // type → file count
	Unrecognized []string       // dirs in reference/ not in taxonomy
	Untyped      []string       // files in flat reference/ (not in type subdir)
	Work         WorkReport
	Pending      PendingReport
	Retro        RetroReport
	Naming       []string // files without date prefix
	Grade        string   // "Good", "Needs Attention", "Stale"
}

// RetroReport summarizes retro file colocation status.
type RetroReport struct {
	Total     int // total retro files
	Colocated int // retros inside a work dir
	Orphaned  int // retros in reference/retro/ with a matching work dir they could move to
}

// WorkReport summarizes the work layer.
type WorkReport struct {
	Active   int
	Archived int
}

// PendingReport summarizes pending items.
type PendingReport struct {
	Active   int
	Terminal int
}

var datePrefix = regexp.MustCompile(`^\d{4}-`)

// Analyze performs a health analysis of a folio project.
func Analyze(f *config.Folio, folioDir string) *Report {
	r := &Report{
		Project:   f.Project,
		Reference: make(map[string]int),
	}

	analyzeReference(r, folioDir)
	analyzeWork(r, folioDir)
	analyzePending(r, f)
	analyzeRetro(r, folioDir)
	analyzeNaming(r, folioDir)
	r.Grade = computeGrade(r)

	return r
}

func analyzeReference(r *Report, folioDir string) {
	refDir := filepath.Join(folioDir, "reference")
	if _, err := os.Stat(refDir); os.IsNotExist(err) {
		return
	}

	entries, err := os.ReadDir(refDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			if taxonomy.IsReferenceType(name) {
				// Count files in this type directory
				typeDir := filepath.Join(refDir, name)
				count := countFiles(typeDir)
				if count > 0 {
					r.Reference[name] = count
				}
			} else {
				r.Unrecognized = append(r.Unrecognized, name)
			}
		} else {
			// File directly in reference/ — untyped
			r.Untyped = append(r.Untyped, entry.Name())
		}
	}
}

func countFiles(dir string) int {
	count := 0
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	return count
}

func analyzeWork(r *Report, folioDir string) {
	workDir := filepath.Join(folioDir, "work")
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		return
	}

	activeDir := filepath.Join(workDir, "active")
	if entries, err := os.ReadDir(activeDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				r.Work.Active++
			}
		}
	}

	archiveDir := filepath.Join(workDir, "archive")
	if entries, err := os.ReadDir(archiveDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				r.Work.Archived++
			}
		}
	}
}

func analyzePending(r *Report, f *config.Folio) {
	for _, item := range f.Pending {
		if IsPendingTerminal(item) {
			r.Pending.Terminal++
		} else {
			r.Pending.Active++
		}
	}
}

func analyzeRetro(r *Report, folioDir string) {
	// Count retros in reference/retro/
	retroDir := filepath.Join(folioDir, "reference", "retro")
	if entries, err := os.ReadDir(retroDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			r.Retro.Total++
			// Extract topic from YYYY-MM-DD-<topic>.md
			name := strings.TrimSuffix(e.Name(), ".md")
			parts := strings.SplitN(name, "-", 4)
			if len(parts) >= 4 {
				topic := parts[3]
				if taxonomy.FindWorkDir(folioDir, topic) != "" {
					r.Retro.Orphaned++
				}
			}
		}
	}

	// Count colocated retros in work dirs
	for _, layer := range []string{"active", "archive"} {
		layerDir := filepath.Join(folioDir, "work", layer)
		workDirs, err := os.ReadDir(layerDir)
		if err != nil {
			continue
		}
		for _, wd := range workDirs {
			if !wd.IsDir() {
				continue
			}
			retroSubDir := filepath.Join(layerDir, wd.Name(), "reference", "retro")
			if entries, err := os.ReadDir(retroSubDir); err == nil {
				for _, e := range entries {
					if !e.IsDir() {
						r.Retro.Total++
						r.Retro.Colocated++
					}
				}
			}
			// Also check for retro.md directly in work dir
			retroFile := filepath.Join(layerDir, wd.Name(), "retro.md")
			if _, err := os.Stat(retroFile); err == nil {
				r.Retro.Total++
				r.Retro.Colocated++
			}
		}
	}
}

func analyzeNaming(r *Report, folioDir string) {
	refDir := filepath.Join(folioDir, "reference")
	if _, err := os.Stat(refDir); os.IsNotExist(err) {
		return
	}

	// Check files in type subdirectories for date prefix
	for _, t := range taxonomy.ReferenceTypes {
		typeDir := filepath.Join(refDir, t)
		if _, err := os.Stat(typeDir); os.IsNotExist(err) {
			continue
		}
		entries, err := os.ReadDir(typeDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if !datePrefix.MatchString(e.Name()) {
				r.Naming = append(r.Naming, filepath.Join(t, e.Name()))
			}
		}
	}
}

// IsPendingTerminal returns true if the pending item has a terminal-state prefix.
func IsPendingTerminal(item string) bool {
	return strings.HasPrefix(item, "[DONE:") ||
		strings.HasPrefix(item, "[SPLIT→") ||
		strings.HasPrefix(item, "[SPLIT->") ||
		strings.HasPrefix(item, "[DROPPED:")
}

func computeGrade(r *Report) string {
	if len(r.Untyped) > 0 || len(r.Unrecognized) > 0 {
		return "Needs Attention"
	}
	if len(r.Naming) > 0 {
		return "Needs Attention"
	}
	totalRef := 0
	for _, c := range r.Reference {
		totalRef += c
	}
	if totalRef == 0 && r.Work.Active == 0 && r.Pending.Active == 0 {
		return "Good" // empty project is fine
	}
	return "Good"
}

// TotalReferenceFiles returns the total count of typed reference files.
func (r *Report) TotalReferenceFiles() int {
	total := 0
	for _, c := range r.Reference {
		total += c
	}
	return total
}
