package touch

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
	"gopkg.in/yaml.v3"
)

// ItemState records the input digest for one complete batch item.
type ItemState struct {
	ID           string
	InputsSHA256 string
}

// State is the target composition state produced by Snapshot.
type State struct {
	ComposedAt   string
	InputsSHA256 string
	Items        []ItemState
	Final        bool
}

// Snapshot reads every local input and verifies every required local output.
func Snapshot(ctx config.Context, target *config.Target, final bool, now time.Time) (State, error) {
	if target == nil {
		return State{}, fmt.Errorf("target is nil")
	}
	if target.Forest != nil {
		return State{}, fmt.Errorf("forest targets are delegated to jf")
	}

	inputsSHA256, items, err := digestInputs(ctx, target)
	if err != nil {
		return State{}, err
	}
	for _, output := range target.Outputs {
		if output.Path == "" {
			continue
		}
		fullPath, err := ctx.ResolvePath(output.Path)
		if err != nil {
			return State{}, fmt.Errorf("resolving output %s: %w", output.Path, err)
		}
		info, err := os.Stat(fullPath)
		if err != nil {
			return State{}, fmt.Errorf("required output %s: %w", output.Path, err)
		}
		if !info.Mode().IsRegular() {
			return State{}, fmt.Errorf("required output %s is not a regular file", output.Path)
		}
	}

	if now.IsZero() {
		now = time.Now()
	}
	return State{
		ComposedAt:   now.UTC().Format(time.RFC3339),
		InputsSHA256: inputsSHA256,
		Items:        items,
		Final:        final,
	}, nil
}

// InputDigest returns the current aggregate digest for a target's local inputs.
func InputDigest(ctx config.Context, target *config.Target) (string, error) {
	if target == nil {
		return "", fmt.Errorf("target is nil")
	}
	digest, _, err := digestInputs(ctx, target)
	return digest, err
}

// ItemInputDigest returns the current digest for one local batch input.
func ItemInputDigest(ctx config.Context, source string) (string, error) {
	if source == "" {
		return "", fmt.Errorf("batch item source is empty")
	}
	digest := sha256.New()
	if err := streamInput(ctx, source, digest, nil); err != nil {
		return "", err
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func digestInputs(ctx config.Context, target *config.Target) (string, []ItemState, error) {
	aggregate := sha256.New()
	var items []ItemState
	for _, source := range target.Sources {
		if source.Path == "" || source.External != "" {
			continue
		}
		if err := streamInput(ctx, source.Path, aggregate, nil); err != nil {
			return "", nil, err
		}
	}
	if target.Batch != nil {
		for _, item := range target.Batch.Items {
			if item.Source == "" {
				continue
			}
			itemDigest := sha256.New()
			if err := streamInput(ctx, item.Source, aggregate, itemDigest); err != nil {
				return "", nil, fmt.Errorf("batch item %q: %w", item.ID, err)
			}
			items = append(items, ItemState{
				ID:           item.ID,
				InputsSHA256: hex.EncodeToString(itemDigest.Sum(nil)),
			})
		}
	}
	return hex.EncodeToString(aggregate.Sum(nil)), items, nil
}

func streamInput(ctx config.Context, logicalPath string, aggregate hash.Hash, item hash.Hash) error {
	fullPath, err := ctx.ResolvePath(logicalPath)
	if err != nil {
		return fmt.Errorf("resolving input %s: %w", logicalPath, err)
	}
	file, err := os.Open(fullPath)
	if err != nil {
		return fmt.Errorf("reading input %s: %w", logicalPath, err)
	}
	defer file.Close()

	if err := writePathBoundary(aggregate, logicalPath); err != nil {
		return fmt.Errorf("framing input %s: %w", logicalPath, err)
	}
	writer := io.Writer(aggregate)
	if item != nil {
		if err := writePathBoundary(item, logicalPath); err != nil {
			return fmt.Errorf("framing input %s: %w", logicalPath, err)
		}
		writer = io.MultiWriter(aggregate, item)
	}
	if _, err := io.Copy(writer, file); err != nil {
		return fmt.Errorf("reading input %s: %w", logicalPath, err)
	}
	return nil
}

func writePathBoundary(writer io.Writer, logicalPath string) error {
	if _, err := io.WriteString(writer, fmt.Sprintf("%d:", len(logicalPath))); err != nil {
		return err
	}
	if _, err := io.WriteString(writer, logicalPath); err != nil {
		return err
	}
	_, err := io.WriteString(writer, "\x00")
	return err
}

// WriteState atomically patches only the selected target's snapshot fields.
func WriteState(folioPath, targetID string, state State) error {
	raw, err := os.ReadFile(folioPath)
	if err != nil {
		return fmt.Errorf("reading folio.yml: %w", err)
	}
	folio, err := config.Parse(raw)
	if err != nil {
		return err
	}
	if _, ok := folio.Targets[targetID]; !ok {
		return fmt.Errorf("target %q not found", targetID)
	}
	if err := validateState(state); err != nil {
		return err
	}
	patched, err := patchTargetState(raw, targetID, state)
	if err != nil {
		return err
	}

	info, err := os.Stat(folioPath)
	if err != nil {
		return fmt.Errorf("stating folio.yml: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(folioPath), ".folio-state-*")
	if err != nil {
		return fmt.Errorf("creating temporary folio.yml: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Chmod(info.Mode().Perm()); err != nil {
		tmp.Close()
		return fmt.Errorf("setting temporary folio.yml mode: %w", err)
	}
	if _, err := tmp.Write(patched); err != nil {
		tmp.Close()
		return fmt.Errorf("writing temporary folio.yml: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("syncing temporary folio.yml: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temporary folio.yml: %w", err)
	}
	if err := os.Rename(tmpPath, folioPath); err != nil {
		return fmt.Errorf("replacing folio.yml: %w", err)
	}
	return nil
}

func validateState(state State) error {
	if state.ComposedAt == "" || !strings.HasSuffix(state.ComposedAt, "Z") {
		return fmt.Errorf("composition timestamp must be a UTC RFC 3339 value")
	}
	if _, err := time.Parse(time.RFC3339, state.ComposedAt); err != nil {
		return fmt.Errorf("invalid composition timestamp: %w", err)
	}
	if len(state.InputsSHA256) != sha256.Size*2 {
		return fmt.Errorf("inputs digest must be a 64-character SHA-256 value")
	}
	if _, err := hex.DecodeString(state.InputsSHA256); err != nil || state.InputsSHA256 != strings.ToLower(state.InputsSHA256) {
		return fmt.Errorf("inputs digest must be lowercase hexadecimal")
	}
	seen := make(map[string]bool, len(state.Items))
	for _, item := range state.Items {
		if item.ID == "" || seen[item.ID] {
			return fmt.Errorf("batch item state IDs must be non-empty and unique")
		}
		if len(item.InputsSHA256) != sha256.Size*2 {
			return fmt.Errorf("batch item %q digest must be a 64-character SHA-256 value", item.ID)
		}
		if _, err := hex.DecodeString(item.InputsSHA256); err != nil || item.InputsSHA256 != strings.ToLower(item.InputsSHA256) {
			return fmt.Errorf("batch item %q digest must be lowercase hexadecimal", item.ID)
		}
		seen[item.ID] = true
	}
	return nil
}

type manifestLine struct {
	indent int
	key    string
	value  string
}

func patchTargetState(raw []byte, targetID string, state State) ([]byte, error) {
	lines, lineEnding := splitManifestLines(string(raw))
	targetStart, targetEnd, err := findTargetLines(lines, targetID)
	if err != nil {
		return nil, err
	}

	replacements := make(map[int]string)
	removed := make(map[int]bool)
	inserts := make(map[int][]string)
	fields := map[string]string{
		"composed_at":   state.ComposedAt,
		"inputs_sha256": state.InputsSHA256,
	}
	if state.Final {
		fields["final"] = "true"
	}
	for i := targetStart + 1; i < targetEnd; i++ {
		line, ok := parseManifestLine(lines[i])
		if !ok || line.indent != 4 {
			continue
		}
		value, wanted := fields[line.key]
		if !wanted && line.key != "final" {
			continue
		}
		if line.key == "final" && !state.Final {
			removed[i] = true
			continue
		}
		replacements[i] = replaceManifestScalar(lines[i], value, lineEnding)
		delete(fields, line.key)
	}
	var newFields []string
	for _, key := range []string{"composed_at", "inputs_sha256", "final"} {
		if value, ok := fields[key]; ok {
			newFields = append(newFields, "    "+key+": "+value+lineEnding)
		}
	}
	if len(newFields) > 0 {
		inserts[targetStart+1] = append(inserts[targetStart+1], newFields...)
	}

	itemRanges, err := findBatchItems(lines, targetStart, targetEnd)
	if err != nil {
		return nil, err
	}
	itemStates := make(map[string]ItemState, len(state.Items))
	for _, item := range state.Items {
		itemStates[item.ID] = item
	}
	for _, itemRange := range itemRanges {
		itemID, idLine, sourceLine, digestLine := batchItemLines(lines, itemRange)
		if itemID == "" {
			return nil, fmt.Errorf("batch item is missing id")
		}
		itemState, hasState := itemStates[itemID]
		if !hasState {
			if digestLine >= 0 {
				removed[digestLine] = true
			}
			continue
		}
		if sourceLine < 0 {
			return nil, fmt.Errorf("batch item %q has digest state but no source", itemID)
		}
		if digestLine >= 0 {
			replacements[digestLine] = replaceManifestScalar(lines[digestLine], itemState.InputsSHA256, lineEnding)
		} else {
			inserts[sourceLine+1] = append(inserts[sourceLine+1], "          inputs_sha256: "+itemState.InputsSHA256+lineEnding)
		}
		delete(itemStates, itemID)
		_ = idLine
	}
	if len(itemStates) > 0 {
		return nil, fmt.Errorf("batch state contains an unknown item")
	}

	var output strings.Builder
	for i, line := range lines {
		for _, inserted := range inserts[i] {
			output.WriteString(inserted)
		}
		if removed[i] {
			continue
		}
		if replacement, ok := replacements[i]; ok {
			output.WriteString(replacement)
		} else {
			output.WriteString(line)
		}
	}
	for _, inserted := range inserts[len(lines)] {
		output.WriteString(inserted)
	}
	return []byte(output.String()), nil
}

func findTargetLines(lines []string, targetID string) (int, int, error) {
	targetsLine := -1
	for i, raw := range lines {
		line, ok := parseManifestLine(raw)
		if ok && line.indent == 0 && line.key == "targets" {
			targetsLine = i
			break
		}
	}
	if targetsLine < 0 {
		return 0, 0, fmt.Errorf("manifest has no targets mapping")
	}
	start := -1
	for i := targetsLine + 1; i < len(lines); i++ {
		line, ok := parseManifestLine(lines[i])
		if !ok {
			continue
		}
		if line.indent == 0 {
			break
		}
		if line.indent == 2 && line.key == targetID {
			start = i
			break
		}
	}
	if start < 0 {
		return 0, 0, fmt.Errorf("target %q mapping not found", targetID)
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		line, ok := parseManifestLine(lines[i])
		if ok && (line.indent == 2 || line.indent == 0) {
			end = i
			break
		}
	}
	return start, end, nil
}

func findBatchItems(lines []string, targetStart, targetEnd int) ([][2]int, error) {
	itemsLine := -1
	for i := targetStart + 1; i < targetEnd; i++ {
		line, ok := parseManifestLine(lines[i])
		if ok && line.indent == 6 && line.key == "items" {
			itemsLine = i
			break
		}
	}
	if itemsLine < 0 {
		return nil, nil
	}
	var ranges [][2]int
	start := -1
	for i := itemsLine + 1; i < targetEnd; i++ {
		body := manifestLineBody(lines[i])
		indent := leadingSpaces(body)
		trimmed := strings.TrimSpace(body)
		if indent <= 6 && trimmed != "" {
			break
		}
		if indent == 8 && strings.HasPrefix(trimmed, "-") {
			if start >= 0 {
				ranges = append(ranges, [2]int{start, i})
			}
			start = i
		}
	}
	if start >= 0 {
		ranges = append(ranges, [2]int{start, targetEnd})
	}
	return ranges, nil
}

func batchItemLines(lines []string, itemRange [2]int) (string, int, int, int) {
	id := ""
	idLine, sourceLine, digestLine := -1, -1, -1
	for i := itemRange[0]; i < itemRange[1]; i++ {
		body := manifestLineBody(lines[i])
		trimmed := strings.TrimSpace(body)
		if i == itemRange[0] && strings.HasPrefix(trimmed, "-") {
			if inline, ok := parseManifestLine(strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))); ok && inline.key == "id" {
				id = parseYAMLString(inline.value)
				idLine = i
			}
		}
		line, ok := parseManifestLine(lines[i])
		if !ok || line.indent < 10 {
			continue
		}
		switch line.key {
		case "id":
			id = parseYAMLString(line.value)
			idLine = i
		case "source":
			sourceLine = i
		case "inputs_sha256":
			digestLine = i
		}
	}
	return id, idLine, sourceLine, digestLine
}

func parseManifestLine(raw string) (manifestLine, bool) {
	body := manifestLineBody(raw)
	indent := leadingSpaces(body)
	trimmed := strings.TrimSpace(body)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "-") {
		return manifestLine{}, false
	}
	colon := strings.IndexByte(trimmed, ':')
	if colon <= 0 {
		return manifestLine{}, false
	}
	key := strings.TrimSpace(trimmed[:colon])
	if strings.HasPrefix(key, "\"") || strings.HasPrefix(key, "'") {
		key = parseYAMLString(key)
	}
	return manifestLine{indent: indent, key: key, value: strings.TrimSpace(trimmed[colon+1:])}, true
}

func parseYAMLString(value string) string {
	var decoded string
	if err := yaml.Unmarshal([]byte(value), &decoded); err == nil {
		return decoded
	}
	return strings.Trim(value, "\"'")
}

func replaceManifestScalar(raw, value, lineEnding string) string {
	body := manifestLineBody(raw)
	colon := strings.IndexByte(body, ':')
	return body[:colon+1] + " " + value + lineEnding
}

func splitManifestLines(text string) ([]string, string) {
	if text == "" {
		return nil, "\n"
	}
	lines := strings.SplitAfter(text, "\n")
	lineEnding := "\n"
	if strings.Contains(text, "\r\n") {
		lineEnding = "\r\n"
	}
	return lines, lineEnding
}

func manifestLineBody(raw string) string {
	return strings.TrimSuffix(strings.TrimSuffix(raw, "\n"), "\r")
}

func leadingSpaces(value string) int {
	for i, r := range value {
		if r != ' ' {
			return i
		}
	}
	return len(value)
}
