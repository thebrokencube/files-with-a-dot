package pipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
)

// Runner executes an external command and returns its combined output.
type Runner func(name string, args ...string) ([]byte, error)

// DefaultRunner shells out via os/exec.
func DefaultRunner(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Stderr = os.Stderr
	return cmd.Output()
}

// Pipeline wraps acli operations with an injectable Runner for testability.
type Pipeline struct {
	Run Runner
}

// Compile converts markdown source into an acli-edit JSON payload.
// Shells out to Node via embedded marklassian bundle.
func (p *Pipeline) Compile(id string, source []byte) ([]byte, error) {
	adf, err := CompileMarkdown(source)
	if err != nil {
		return nil, fmt.Errorf("compile: %w", err)
	}

	payload := map[string]any{
		"issues":      []string{id},
		"description": json.RawMessage(adf),
	}

	return json.MarshalIndent(payload, "", "  ")
}

// Push sends a compiled JSON payload to Jira via acli edit.
func (p *Pipeline) Push(compiled []byte) error {
	tmpFile, err := os.CreateTemp("", "jf-push-*.json")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(compiled); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	tmpFile.Close()

	out, err := p.Run("acli", "jira", "workitem", "edit", "--from-json", tmpFile.Name(), "--yes")
	if err != nil {
		return fmt.Errorf("acli edit: %s\n%s", err, string(out))
	}
	return nil
}

var reJiraKey = regexp.MustCompile(`[A-Z]+-\d+`)

// Create creates a Jira ticket from a JSON payload and returns the new key.
func (p *Pipeline) Create(payload []byte) (string, error) {
	tmpFile, err := os.CreateTemp("", "jf-create-*.json")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(payload); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("write temp file: %w", err)
	}
	tmpFile.Close()

	out, err := p.Run("acli", "jira", "workitem", "create", "--from-json", tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("acli create: %s\n%s", err, string(out))
	}

	key := reJiraKey.FindString(string(out))
	if key == "" {
		return "", fmt.Errorf("no Jira key found in acli output: %s", string(out))
	}
	return key, nil
}

// View fetches issue details via acli.
func (p *Pipeline) View(id string, fields string, jsonOut bool) ([]byte, error) {
	args := []string{"jira", "workitem", "view", id}
	if fields != "" {
		args = append(args, "--fields", fields)
	}
	if jsonOut {
		args = append(args, "--json")
	}

	out, err := p.Run("acli", args...)
	if err != nil {
		return nil, fmt.Errorf("acli view: %s\n%s", err, string(out))
	}
	return out, nil
}

// Search queries Jira issues by JQL via acli.
func (p *Pipeline) Search(jql string, fields string, limit int) ([]byte, error) {
	args := []string{"jira", "workitem", "search", "--jql", jql}
	if fields != "" {
		args = append(args, "--fields", fields)
	}
	if limit > 0 {
		args = append(args, "--limit", strconv.Itoa(limit))
	}

	out, err := p.Run("acli", args...)
	if err != nil {
		return nil, fmt.Errorf("acli search: %s\n%s", err, string(out))
	}
	return out, nil
}
