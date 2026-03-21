package pipeline

import (
	"strings"
	"testing"
)

func TestConvertADF(t *testing.T) {
	requireNode(t)
	adf := `{"type":"doc","version":1,"content":[{"type":"heading","attrs":{"level":2},"content":[{"type":"text","text":"Hello"}]},{"type":"paragraph","content":[{"type":"text","text":"Some "},{"type":"text","marks":[{"type":"strong"}],"text":"bold"},{"type":"text","text":" text"}]}]}`

	md, err := ConvertADF([]byte(adf))
	if err != nil {
		t.Fatal(err)
	}

	out := string(md)
	if !strings.Contains(out, "Hello") {
		t.Errorf("expected markdown to contain 'Hello', got %q", out)
	}
	if !strings.Contains(out, "**bold**") {
		t.Errorf("expected markdown to contain '**bold**', got %q", out)
	}
}

func TestConvertADFInvalidJSON(t *testing.T) {
	requireNode(t)
	_, err := ConvertADF([]byte("{invalid"))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestExtractDescriptionADF(t *testing.T) {
	t.Run("valid description", func(t *testing.T) {
		input := `{"fields":{"description":{"type":"doc","version":1,"content":[]}}}`
		adf, err := ExtractDescriptionADF([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		if adf == nil {
			t.Fatal("expected non-nil ADF")
		}
		if !strings.Contains(string(adf), `"type":"doc"`) {
			t.Errorf("expected ADF document, got %s", string(adf))
		}
	})

	t.Run("null description", func(t *testing.T) {
		input := `{"fields":{"description":null}}`
		adf, err := ExtractDescriptionADF([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		if adf != nil {
			t.Errorf("expected nil for null description, got %s", string(adf))
		}
	})

	t.Run("missing description", func(t *testing.T) {
		input := `{"fields":{}}`
		adf, err := ExtractDescriptionADF([]byte(input))
		if err != nil {
			t.Fatal(err)
		}
		if adf != nil {
			t.Errorf("expected nil for missing description, got %s", string(adf))
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := ExtractDescriptionADF([]byte("{bad"))
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}
