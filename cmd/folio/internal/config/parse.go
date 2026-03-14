package config

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads and parses a folio.yml file from the given path.
func Load(path string) (*Folio, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return Parse(data)
}

// Parse decodes YAML bytes into a Folio struct.
// Unknown top-level keys are rejected to catch typos that would silently drop data.
func Parse(data []byte) (*Folio, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var f Folio
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}
	Normalize(&f)
	return &f, nil
}

// Normalize initializes nil maps and slices after parsing so downstream code
// doesn't need repeated nil checks.
func Normalize(f *Folio) {
	if f.Sources == nil {
		f.Sources = []Source{}
	}
	if f.Targets == nil {
		f.Targets = make(map[string]Target)
	}
	if f.CrossReferences == nil {
		f.CrossReferences = []CrossReference{}
	}
	if f.Tasks == nil {
		f.Tasks = []string{}
	}
	if f.Pending == nil {
		f.Pending = []string{}
	}
	if f.Observations == nil {
		f.Observations = []string{}
		// Auto-upgrade: merge Tasks into Observations for schema 1 files
		if len(f.Tasks) > 0 {
			f.Observations = append(f.Observations, f.Tasks...)
		}
	}
	if f.Repositories == nil {
		f.Repositories = make(map[string]string)
	}
}
