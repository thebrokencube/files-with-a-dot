package config

import (
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
func Parse(data []byte) (*Folio, error) {
	var f Folio
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parsing YAML: %w", err)
	}
	return &f, nil
}
