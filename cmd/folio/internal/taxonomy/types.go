package taxonomy

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// ArtifactType describes a typed artifact in the taxonomy.
type ArtifactType struct {
	Name  string
	Layer string // "reference" or "work"
}

// LifecycleStage represents a stage in the knowledge lifecycle.
type LifecycleStage int

const (
	StageObservation LifecycleStage = iota
	StageSpike
	StageDesign
	StagePlan
	StageImplementation
	StageRetro
	StageReference // not a lifecycle stage — accumulates over time
)

// ReferenceLabels lists valid labels for the "reference" meta-type.
var ReferenceLabels = map[string]bool{
	"research": true,
	"insight":  true,
	"guide":    true,
	"domain":   true,
	"review":   true,
}

// TypeAlias maps old type names to (canonical, label) pairs.
// Used for backward-compatible type resolution.
var TypeAlias = map[string][2]string{
	"survey":    {"reference", "research"},
	"synthesis": {"reference", "research"},
	"pattern":   {"reference", "insight"},
	"domain":    {"reference", "domain"},
	"guide":     {"reference", "guide"},
	"review":    {"reference", "review"},
}

// ResolveAlias resolves an old type name to its canonical form and label.
// Returns (canonical, label, true) if an alias exists, or ("", "", false).
func ResolveAlias(t string) (string, string, bool) {
	if pair, ok := TypeAlias[t]; ok {
		return pair[0], pair[1], true
	}
	return "", "", false
}

// StageForType returns the lifecycle stage for a given type.
func StageForType(t string) LifecycleStage {
	switch t {
	case "observation":
		return StageObservation
	case "spike":
		return StageSpike
	case "design":
		return StageDesign
	case "plan", "brief":
		return StagePlan
	case "track":
		return StageImplementation
	case "retro":
		return StageRetro
	default:
		return StageReference
	}
}

// InferType extracts the artifact type from a source path's directory component.
// Examples: "reference/spike/foo.md" -> "spike", "work/active/bar/README.md" -> "brief"
func InferType(path string) string {
	parts := strings.Split(filepath.ToSlash(path), "/")
	if len(parts) >= 2 && parts[0] == "reference" {
		return parts[1]
	}
	if len(parts) >= 2 && parts[0] == "work" {
		return "brief"
	}
	return ""
}

// ReferenceTypes lists all valid reference-layer type names.
var ReferenceTypes = []string{
	"spike", "survey",
	"synthesis", "domain",
	"pattern", "guide", "review", "retro",
}

// ValidTypes maps every valid type name to true.
var ValidTypes map[string]bool

func init() {
	ValidTypes = make(map[string]bool, len(ReferenceTypes)+1)
	for _, t := range ReferenceTypes {
		ValidTypes[t] = true
	}
	ValidTypes["brief"] = true
	ValidTypes["design"] = true
	ValidTypes["plan"] = true
}

// IsReferenceType returns true if t is a recognized reference-layer type.
func IsReferenceType(t string) bool {
	if t == "brief" || t == "design" {
		return false
	}
	return ValidTypes[t]
}

// IsReferenceDir returns true if t should have a directory under reference/.
// This includes all ReferenceTypes plus "design" (which was removed from
// ReferenceTypes but still lives in reference/).
func IsReferenceDir(t string) bool {
	return IsReferenceType(t) || t == "design" || t == "research" || t == "insight"
}

// ColocatableTypes lists types that colocate with a matching work directory.
var ColocatableTypes = map[string]bool{
	"design": true,
	"retro":  true,
}

// IsWorkType returns true if t is a recognized work-layer type.
func IsWorkType(t string) bool {
	return t == "brief"
}

// FindWorkDir returns the path to a work directory matching the given topic,
// searching active/ then archive/ under folioDir/work/.
// Returns "" if no match is found.
func FindWorkDir(folioDir, topic string) string {
	if folioDir == "" {
		return ""
	}
	for _, layer := range []string{"active", "archive"} {
		pattern := filepath.Join(folioDir, "work", layer, "*-"+topic)
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			return matches[0]
		}
	}
	return ""
}

// TypePath returns the relative path for a new artifact of the given type and topic.
// The path is relative to the folio directory (where folio.yml lives).
func TypePath(artifactType, topic string) string {
	date := time.Now().Format("2006-01-02")
	if artifactType == "design" {
		return filepath.Join("reference", "design", fmt.Sprintf("%s-%s.md", date, topic))
	}
	if IsReferenceType(artifactType) {
		return filepath.Join("reference", artifactType, fmt.Sprintf("%s-%s.md", date, topic))
	}
	if IsWorkType(artifactType) {
		return filepath.Join("work", "active", fmt.Sprintf("%s-%s", date, topic), "README.md")
	}
	return ""
}

// Template returns a markdown template for the given artifact type and topic.
func Template(artifactType, topic string) string {
	title := strings.ReplaceAll(topic, "-", " ")
	title = strings.Title(title) //nolint:staticcheck

	switch artifactType {
	case "design":
		return fmt.Sprintf(`# %s

## Problem
<!-- What problem does this solve and why? -->

## Architecture
<!-- Key decisions, type definitions, function signatures -->

## Divergence Decisions
<!-- Table: Aspect | Option A | Option B | Chosen | Rationale -->

## What's NOT Included
<!-- Explicit scope boundary -->

## Open Questions
<!-- What remains unresolved? -->
`, title)

	case "spike":
		return fmt.Sprintf(`# %s

## Purpose
<!-- What are we investigating and why? Time-boxed to: -->

## Findings
<!-- What did we learn? -->

## Gaps and Ambiguities
<!-- What couldn't be resolved? -->

## Summary
<!-- Key takeaway in 2-3 sentences -->
`, title)

	case "survey":
		return fmt.Sprintf(`# %s

## Overview
<!-- What external system/spec/tool is being surveyed? -->

## Key Features
<!-- Notable capabilities and design choices -->

## Comparison
<!-- How does this compare to alternatives or existing approaches? -->

## Relevance
<!-- How does this apply to our project? -->
`, title)

	case "brief":
		return fmt.Sprintf(`# %s

## Objective
<!-- What are we trying to accomplish? -->

## Context
<!-- Distill from design doc: what this work is, key design decisions (non-negotiable),
     scope boundary (what's NOT included, framed as stop signals).
     Target: 10-15 lines. Every line should change how the execution agent behaves. -->

## Agent Setup
<!-- Skill loading, repo mapping, escalation triggers.
     1. Which skills to invoke (/folio status, /commit, etc.) and key rules
     2. Which tracks operate in which repos
     3. When should the agent stop and ask instead of improvising? -->

## Tracks
<!-- Parallel work tracks with their scope -->

## Execution Conventions
<!-- Commit format, validation commands, scope target, folio integration -->

## Open Questions
<!-- What remains unresolved? -->
`, title)

	case "retro":
		return fmt.Sprintf(`# %s

## Context
<!-- What was done and in what setting? -->

## What Happened
<!-- Factual account of events and outcomes -->

## What Worked
<!-- Practices, tools, or decisions that went well -->

## What Didn't
<!-- Pain points, surprises, or failures -->

## Action Items
<!-- Concrete next steps with owners if applicable -->
`, title)

	default:
		return fmt.Sprintf(`# %s

## Problem
<!-- What problem does this solve and why? -->

## Approach
<!-- Key decisions, type definitions, function signatures -->

## Alternatives
<!-- What was considered and rejected? -->

## Open Questions
<!-- What remains unresolved? -->
`, title)
	}
}
