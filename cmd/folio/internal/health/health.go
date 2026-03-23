package health

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/observe"
	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/taxonomy"
)

// Report contains the health analysis for a single folio project.
type Report struct {
	Project      string
	Reference    map[string]int // type → file count
	Unrecognized []string       // dirs in reference/ not in taxonomy
	Untyped      []string       // files in flat reference/ (not in type subdir)
	Work         WorkReport
	Observations ObservationReport
	Design       ColocationReport
	Retro        ColocationReport
	Naming       []string // files without date prefix
	Grade        string   // "Good", "Needs Attention", "Stale"
}

// ColocationReport summarizes colocation status for a colocatable type (design, retro).
type ColocationReport struct {
	Total     int // total files of this type
	Colocated int // files inside a work dir
	Orphaned  int // files in reference/<type>/ with a matching work dir they could move to
}

// WorkReport summarizes the work layer.
type WorkReport struct {
	Active   int
	Archived int
}

// ObservationReport summarizes observation items.
type ObservationReport struct {
	Active       int
	LintWarnings []string
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
	analyzeObservations(r, f, folioDir)
	analyzeDesign(r, folioDir)
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
			if taxonomy.IsReferenceDir(name) {
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

func analyzeObservations(r *Report, f *config.Folio, folioDir string) {
	r.Observations.Active = len(f.Observations)
	issues := observe.Lint(folioDir, f.Observations)
	for _, issue := range issues {
		r.Observations.LintWarnings = append(r.Observations.LintWarnings,
			fmt.Sprintf("#%d: %s", issue.Index, issue.Reason))
	}
}

func analyzeDesign(r *Report, folioDir string) {
	analyzeColocation(&r.Design, folioDir, "design")
}

func analyzeRetro(r *Report, folioDir string) {
	analyzeColocation(&r.Retro, folioDir, "retro")
}

// analyzeColocation checks for files of a colocatable type that sit in
// reference/<typeName>/ but have a matching work dir they could move into.
// It also counts colocated files already inside work dirs.
func analyzeColocation(cr *ColocationReport, folioDir string, typeName string) {
	// Count files in reference/<type>/ and detect orphans
	refTypeDir := filepath.Join(folioDir, "reference", typeName)
	if entries, err := os.ReadDir(refTypeDir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			cr.Total++
			// Extract topic from YYYY-MM-DD-<topic>.md
			name := strings.TrimSuffix(e.Name(), ".md")
			parts := strings.SplitN(name, "-", 4)
			if len(parts) >= 4 {
				topic := parts[3]
				if taxonomy.FindWorkDir(folioDir, topic) != "" {
					cr.Orphaned++
				}
			}
		}
	}

	// Count colocated files in work dirs
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
			typeSubDir := filepath.Join(layerDir, wd.Name(), "reference", typeName)
			if entries, err := os.ReadDir(typeSubDir); err == nil {
				for _, e := range entries {
					if !e.IsDir() {
						cr.Total++
						cr.Colocated++
					}
				}
			}
			// Also check for <type>.md directly in work dir (flat format)
			flatFile := filepath.Join(layerDir, wd.Name(), typeName+".md")
			if _, err := os.Stat(flatFile); err == nil {
				cr.Total++
				cr.Colocated++
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
	entries, err := os.ReadDir(refDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() || !taxonomy.IsReferenceDir(entry.Name()) {
			continue
		}
		t := entry.Name()
		typeDir := filepath.Join(refDir, t)
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
	if totalRef == 0 && r.Work.Active == 0 && r.Observations.Active == 0 {
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
