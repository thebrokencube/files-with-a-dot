package touch

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/thebrokencube/files-with-a-dot/cmd/folio/internal/config"
)

func TestSnapshotStreamsDeclarationOrderedPathBoundaries(t *testing.T) {
	root := t.TempDir()
	writeTouchFile(t, filepath.Join(root, "a"), "ab")
	writeTouchFile(t, filepath.Join(root, "bc"), "c")
	writeTouchFile(t, filepath.Join(root, "ab"), "a")
	writeTouchFile(t, filepath.Join(root, "c"), "bc")
	ctx := config.Context{FolioPath: filepath.Join(root, "folio.yml")}
	now := time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC)

	first, err := Snapshot(ctx, &config.Target{
		Sources: []config.Source{{Path: "a"}, {Path: "bc"}},
	}, false, now)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Snapshot(ctx, &config.Target{
		Sources: []config.Source{{Path: "ab"}, {Path: "c"}},
	}, false, now)
	if err != nil {
		t.Fatal(err)
	}
	if first.InputsSHA256 == second.InputsSHA256 {
		t.Fatal("distinct logical path boundaries produced the same digest")
	}
	if first.ComposedAt != "2026-09-12T10:00:00Z" {
		t.Fatalf("composed_at = %q", first.ComposedAt)
	}
}

func TestSnapshotRequiresAllLocalOutputsWithoutMtimeMutation(t *testing.T) {
	root := t.TempDir()
	writeTouchFile(t, filepath.Join(root, "source.md"), "source")
	output := filepath.Join(root, "compiled", "output.md")
	writeTouchFile(t, output, "output")
	manifest := filepath.Join(root, "folio.yml")
	manifestBefore := []byte("schema: 1\nproject: test\n")
	writeTouchFile(t, manifest, string(manifestBefore))
	past := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(output, past, past); err != nil {
		t.Fatal(err)
	}
	beforeInfo, err := os.Stat(output)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Snapshot(config.Context{FolioPath: manifest}, &config.Target{
		Sources: []config.Source{{Path: "source.md"}},
		Outputs: []config.Output{{Path: "compiled/output.md"}, {Path: "compiled/missing.md"}},
	}, false, time.Now())
	if err == nil {
		t.Fatal("expected missing output error")
	}
	afterInfo, err := os.Stat(output)
	if err != nil {
		t.Fatal(err)
	}
	if !afterInfo.ModTime().Equal(beforeInfo.ModTime()) {
		t.Fatal("snapshot changed an output mtime")
	}
	manifestAfter, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(manifestAfter, manifestBefore) {
		t.Fatal("snapshot changed the manifest")
	}
}

func TestSnapshotCompleteBatchWritesOnlySourceItemDigests(t *testing.T) {
	root := t.TempDir()
	writeTouchFile(t, filepath.Join(root, "source.md"), "source")
	writeTouchFile(t, filepath.Join(root, "item.md"), "item")
	writeTouchFile(t, filepath.Join(root, "compiled", "batch.md"), "batch")
	target := &config.Target{
		Sources: []config.Source{{Path: "source.md"}},
		Outputs: []config.Output{{Path: "compiled/batch.md"}},
		Batch: &config.Batch{Items: []config.BatchItem{
			{ID: "one", Source: "item.md"},
			{ID: "manual", Source: ""},
		}},
	}

	state, err := Snapshot(config.Context{FolioPath: filepath.Join(root, "folio.yml")}, target, false, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(state.Items) != 1 || state.Items[0].ID != "one" || len(state.Items[0].InputsSHA256) != 64 {
		t.Fatalf("batch state = %+v", state.Items)
	}
}

func TestWriteStatePatchesSelectedTargetAtomically(t *testing.T) {
	root := t.TempDir()
	writeTouchFile(t, filepath.Join(root, "source.md"), "source")
	writeTouchFile(t, filepath.Join(root, "item.md"), "item")
	writeTouchFile(t, filepath.Join(root, "compiled", "summary.md"), "summary")
	writeTouchFile(t, filepath.Join(root, "compiled", "other.md"), "other")
	manifest := filepath.Join(root, "folio.yml")
	before := []byte(`schema: 1
project: test
# preserve this comment
targets:
  summary:
    how: compose
    sources:
      - path: source.md
    outputs:
      - path: compiled/summary.md
    batch:
      items:
        - id: one
          source: item.md
  other:
    how: compose
    outputs:
      - path: compiled/other.md
`)
	writeTouchFile(t, manifest, string(before))
	summaryBefore, err := os.ReadFile(filepath.Join(root, "compiled", "summary.md"))
	if err != nil {
		t.Fatal(err)
	}
	summaryInfoBefore, err := os.Stat(filepath.Join(root, "compiled", "summary.md"))
	if err != nil {
		t.Fatal(err)
	}
	otherInfoBefore, err := os.Stat(filepath.Join(root, "compiled", "other.md"))
	if err != nil {
		t.Fatal(err)
	}
	folio, err := config.Load(manifest)
	if err != nil {
		t.Fatal(err)
	}
	target := folio.Targets["summary"]
	state, err := Snapshot(config.Context{FolioPath: manifest}, &target, false, time.Date(2026, 9, 12, 10, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if err := WriteState(manifest, "summary", state); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	summaryAfter, err := os.ReadFile(filepath.Join(root, "compiled", "summary.md"))
	if err != nil {
		t.Fatal(err)
	}
	summaryInfoAfter, err := os.Stat(filepath.Join(root, "compiled", "summary.md"))
	if err != nil {
		t.Fatal(err)
	}
	otherInfoAfter, err := os.Stat(filepath.Join(root, "compiled", "other.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(summaryAfter, summaryBefore) || !summaryInfoAfter.ModTime().Equal(summaryInfoBefore.ModTime()) {
		t.Fatal("state write changed selected output bytes or mtime")
	}
	if !otherInfoAfter.ModTime().Equal(otherInfoBefore.ModTime()) {
		t.Fatal("state write changed unrelated output mtime")
	}
	if !bytes.Contains(after, []byte("composed_at: 2026-09-12T10:00:00Z")) || !bytes.Contains(after, []byte("inputs_sha256: "+state.InputsSHA256)) {
		t.Fatalf("snapshot fields missing:\n%s", after)
	}
	if !bytes.Contains(after, []byte("inputs_sha256: "+state.Items[0].InputsSHA256)) {
		t.Fatalf("batch item digest missing:\n%s", after)
	}
	if !bytes.Contains(after, []byte("# preserve this comment\n")) || !bytes.Contains(after, []byte("  other:\n    how: compose\n")) {
		t.Fatal("unrelated manifest content changed")
	}
}

func TestWriteStateFailurePreservesManifest(t *testing.T) {
	root := t.TempDir()
	manifest := filepath.Join(root, "folio.yml")
	before := []byte("schema: 1\nproject: test\ntargets:\n  summary:\n    how: compose\n")
	writeTouchFile(t, manifest, string(before))

	err := WriteState(manifest, "summary", State{
		ComposedAt:   "2026-09-12T10:00:00Z",
		InputsSHA256: "not-a-digest",
	})
	if err == nil {
		t.Fatal("expected invalid state error")
	}
	after, readErr := os.ReadFile(manifest)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !bytes.Equal(after, before) {
		t.Fatal("failed state write changed the manifest")
	}
}

func writeTouchFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
