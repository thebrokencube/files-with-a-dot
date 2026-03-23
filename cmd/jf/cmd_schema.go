package main

import (
	"encoding/json"
	"fmt"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func runSchema(args []string) int {
	fs := dendrik.NewFlagSet("schema")
	outputSchema := fs.Bool('o', "output", "Emit command output schemas instead of input schemas")

	if err := dendrik.Parse(fs, args); err != nil {
		return dendrik.ExitUserError
	}

	if *outputSchema {
		return printOutputSchemas()
	}
	return printInputSchemas()
}

func printInputSchemas() int {
	schema := map[string]any{
		"$schema":     "https://json-schema.org/draft/2020-12/schema",
		"title":       "jf schemas",
		"description": "Input schemas for jf CLI",
		"definitions": map[string]any{
			"forest.yml": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"schema": map[string]any{
						"type":        "integer",
						"description": "Schema version (currently 1)",
						"const":       1,
					},
					"defaults": map[string]any{
						"type": "object",
						"properties": map[string]any{
							"sync":    map[string]any{"type": "string", "enum": []string{"push", "pull", "both"}},
							"type":    map[string]any{"type": "string", "description": "Default Jira issue type"},
							"field":   map[string]any{"type": "string", "description": "Default field (description or comment)"},
							"project": map[string]any{"type": "string", "description": "Jira project key"},
						},
					},
					"acli": map[string]any{
						"type":        "string",
						"description": "acli version constraint (optional)",
					},
				},
				"required": []string{"schema"},
			},
			"frontmatter": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"jira":  map[string]any{"type": "string", "description": "Jira key (e.g. BEN-123) or TBD"},
					"label": map[string]any{"type": "string", "description": "Display label (overrides heading/filename)"},
					"type":  map[string]any{"type": "string", "description": "Jira issue type"},
					"sync":  map[string]any{"type": "string", "enum": []string{"push", "pull", "both"}},
					"order": map[string]any{"type": "integer", "description": "Sibling sort order (lower first)"},
				},
				"required": []string{"jira"},
			},
		},
	}

	out, _ := json.MarshalIndent(schema, "", "  ")
	fmt.Println(string(out))
	return dendrik.ExitOK
}

func printOutputSchemas() int {
	schema := map[string]any{
		"$schema":     "https://json-schema.org/draft/2020-12/schema",
		"title":       "jf output schemas",
		"description": "JSON output schemas for jf commands",
		"definitions": map[string]any{
			"NodeInfo": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"key":      map[string]any{"type": "string"},
					"label":    map[string]any{"type": "string"},
					"type":     map[string]any{"type": "string"},
					"sync":     map[string]any{"type": "string", "enum": []string{"push", "pull", "both"}},
					"file":     map[string]any{"type": "string"},
					"parent":   map[string]any{"type": "string"},
					"children": map[string]any{"type": "integer"},
					"status":   map[string]any{"type": "string", "enum": []string{"stale", "clean", "unknown"}},
				},
			},
			"StatusResult": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"forest":     map[string]any{"type": "string"},
					"total":      map[string]any{"type": "integer"},
					"tbd":        map[string]any{"type": "integer"},
					"push_total": map[string]any{"type": "integer"},
					"push_stale": map[string]any{"type": "integer"},
					"pull_total": map[string]any{"type": "integer"},
					"pull_stale": map[string]any{"type": "integer"},
				},
			},
			"ValidateResult": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"valid": map[string]any{"type": "boolean"},
					"nodes": map[string]any{"type": "integer"},
					"issues": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "object",
							"properties": map[string]any{
								"level":   map[string]any{"type": "string", "enum": []string{"error", "warning"}},
								"message": map[string]any{"type": "string"},
							},
						},
					},
				},
			},
			"ErrorResult": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"error":  map[string]any{"type": "string"},
					"detail": map[string]any{"type": "string"},
				},
				"required": []string{"error"},
			},
		},
	}

	out, _ := json.MarshalIndent(schema, "", "  ")
	fmt.Println(string(out))
	return dendrik.ExitOK
}
