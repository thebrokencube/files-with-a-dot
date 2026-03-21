package pipeline

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sync"
)

//go:embed adf2md.bundle.mjs
var adf2mdScript []byte

var (
	adf2mdOnce sync.Once
	adf2mdFile string
	adf2mdErr  error
)

// ensureADF2MDScript writes the embedded adf2md bundle to a temp file once.
func ensureADF2MDScript() (string, error) {
	adf2mdOnce.Do(func() {
		var f *os.File
		f, adf2mdErr = os.CreateTemp("", "jf-adf2md-*.mjs")
		if adf2mdErr != nil {
			return
		}
		if _, adf2mdErr = f.Write(adf2mdScript); adf2mdErr != nil {
			f.Close()
			return
		}
		f.Close()
		adf2mdFile = f.Name()
	})
	if adf2mdErr != nil {
		return "", fmt.Errorf("write adf2md script: %w", adf2mdErr)
	}
	return adf2mdFile, nil
}

// ConvertADF converts ADF JSON bytes to markdown via extended-markdown-adf-parser.
// Input: raw ADF document JSON (the .fields.description value from Jira).
// Output: markdown bytes.
func ConvertADF(adfJSON []byte) ([]byte, error) {
	script, err := ensureADF2MDScript()
	if err != nil {
		return nil, err
	}

	tmpFile, err := os.CreateTemp("", "jf-adf2md-*.json")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.Write(adfJSON); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	tmpFile.Close()

	cmd := exec.Command("node", script, tmpFile.Name())
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("adf2md: %w\n%s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// ExtractDescriptionADF extracts the .fields.description ADF document from
// acli's --json output. Returns nil if the field is absent or null.
func ExtractDescriptionADF(viewJSON []byte) (json.RawMessage, error) {
	var issue struct {
		Fields struct {
			Description json.RawMessage `json:"description"`
		} `json:"fields"`
	}
	if err := json.Unmarshal(viewJSON, &issue); err != nil {
		return nil, fmt.Errorf("parse view JSON: %w", err)
	}
	if len(issue.Fields.Description) == 0 || string(issue.Fields.Description) == "null" {
		return nil, nil
	}
	return issue.Fields.Description, nil
}
