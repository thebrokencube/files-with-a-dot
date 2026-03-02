package taxonomy

import (
	"strings"
	"testing"
)

func TestValidTypesIncludesAllReferenceTypes(t *testing.T) {
	for _, rt := range ReferenceTypes {
		if !ValidTypes[rt] {
			t.Errorf("ReferenceType %q not in ValidTypes", rt)
		}
	}
}

func TestValidTypesIncludesBrief(t *testing.T) {
	if !ValidTypes["brief"] {
		t.Error("brief not in ValidTypes")
	}
}

func TestIsReferenceType(t *testing.T) {
	for _, rt := range ReferenceTypes {
		if !IsReferenceType(rt) {
			t.Errorf("IsReferenceType(%q) = false, want true", rt)
		}
	}
	if IsReferenceType("brief") {
		t.Error("IsReferenceType(brief) = true, want false")
	}
	if IsReferenceType("unknown") {
		t.Error("IsReferenceType(unknown) = true, want false")
	}
}

func TestIsWorkType(t *testing.T) {
	if !IsWorkType("brief") {
		t.Error("IsWorkType(brief) = false, want true")
	}
	if IsWorkType("spike") {
		t.Error("IsWorkType(spike) = true, want false")
	}
}

func TestTypePathReference(t *testing.T) {
	path := TypePath("spike", "test-topic")
	if !strings.HasPrefix(path, "reference/spike/") {
		t.Errorf("TypePath(spike) = %s, want prefix reference/spike/", path)
	}
	if !strings.HasSuffix(path, "-test-topic.md") {
		t.Errorf("TypePath(spike) = %s, want suffix -test-topic.md", path)
	}
}

func TestTypePathBrief(t *testing.T) {
	path := TypePath("brief", "my-project")
	if !strings.HasPrefix(path, "work/active/") {
		t.Errorf("TypePath(brief) = %s, want prefix work/active/", path)
	}
	if !strings.HasSuffix(path, "-my-project/README.md") {
		t.Errorf("TypePath(brief) = %s, want suffix -my-project/README.md", path)
	}
}

func TestTypePathInvalid(t *testing.T) {
	path := TypePath("invalid", "topic")
	if path != "" {
		t.Errorf("TypePath(invalid) = %s, want empty", path)
	}
}

func TestTemplateReturnsNonEmpty(t *testing.T) {
	types := append([]string{"brief"}, ReferenceTypes...)
	for _, typ := range types {
		tmpl := Template(typ, "test-topic")
		if tmpl == "" {
			t.Errorf("Template(%q) returned empty", typ)
		}
		if !strings.HasPrefix(tmpl, "# ") {
			t.Errorf("Template(%q) doesn't start with '# '", typ)
		}
	}
}

func TestTemplateDesignHasExpectedSections(t *testing.T) {
	tmpl := Template("design", "test")
	for _, section := range []string{"Problem", "Architecture", "Divergence Decisions", "What's NOT Included", "Open Questions"} {
		if !strings.Contains(tmpl, section) {
			t.Errorf("design template missing section %q", section)
		}
	}
}

func TestTemplateSpikeHasExpectedSections(t *testing.T) {
	tmpl := Template("spike", "test")
	for _, section := range []string{"Purpose", "Findings", "Gaps and Ambiguities", "Summary"} {
		if !strings.Contains(tmpl, section) {
			t.Errorf("spike template missing section %q", section)
		}
	}
}

func TestTemplateSurveyHasExpectedSections(t *testing.T) {
	tmpl := Template("survey", "test")
	for _, section := range []string{"Overview", "Key Features", "Comparison", "Relevance"} {
		if !strings.Contains(tmpl, section) {
			t.Errorf("survey template missing section %q", section)
		}
	}
}

func TestTemplateBriefHasExpectedSections(t *testing.T) {
	tmpl := Template("brief", "test")
	for _, section := range []string{"Objective", "Tracks", "Execution Conventions", "Open Questions"} {
		if !strings.Contains(tmpl, section) {
			t.Errorf("brief template missing section %q", section)
		}
	}
}
