package taxonomy

import (
	"os"
	"path/filepath"
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
	if IsReferenceType("design") {
		t.Error("IsReferenceType(design) = true, want false (design removed from ReferenceTypes)")
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
	for _, section := range []string{"Objective", "Context", "Agent Setup", "Tracks", "Execution Conventions", "Open Questions"} {
		if !strings.Contains(tmpl, section) {
			t.Errorf("brief template missing section %q", section)
		}
	}
}

func TestColocatableTypesContainsDesignAndRetro(t *testing.T) {
	if !ColocatableTypes["design"] {
		t.Error("ColocatableTypes[design] = false, want true")
	}
	if !ColocatableTypes["retro"] {
		t.Error("ColocatableTypes[retro] = false, want true")
	}
}

func TestColocatableTypesExcludesBrief(t *testing.T) {
	if ColocatableTypes["brief"] {
		t.Error("ColocatableTypes[brief] = true, want false")
	}
}

func TestIsReferenceDirDesign(t *testing.T) {
	if !IsReferenceDir("design") {
		t.Error("IsReferenceDir(design) = false, want true")
	}
	for _, rt := range ReferenceTypes {
		if !IsReferenceDir(rt) {
			t.Errorf("IsReferenceDir(%q) = false, want true", rt)
		}
	}
	if IsReferenceDir("brief") {
		t.Error("IsReferenceDir(brief) = true, want false")
	}
	if IsReferenceDir("unknown") {
		t.Error("IsReferenceDir(unknown) = true, want false")
	}
}

func TestFindWorkDirMatchesActive(t *testing.T) {
	dir := t.TempDir()
	workDir := filepath.Join(dir, "work", "active", "2026-01-01-foo")
	os.MkdirAll(workDir, 0755)

	got := FindWorkDir(dir, "foo")
	if got != workDir {
		t.Errorf("FindWorkDir = %q, want %q", got, workDir)
	}
}

func TestFindWorkDirReturnsEmptyOnNoMatch(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "work", "active"), 0755)

	got := FindWorkDir(dir, "nonexistent")
	if got != "" {
		t.Errorf("FindWorkDir = %q, want empty", got)
	}
}

func TestValidTypesIncludesDesign(t *testing.T) {
	if !ValidTypes["design"] {
		t.Error("ValidTypes[design] = false, want true (design still valid even though not a ReferenceType)")
	}
}

func TestTypePathDesign(t *testing.T) {
	path := TypePath("design", "test-topic")
	if !strings.HasPrefix(path, "reference/design/") {
		t.Errorf("TypePath(design) = %s, want prefix reference/design/", path)
	}
	if !strings.HasSuffix(path, "-test-topic.md") {
		t.Errorf("TypePath(design) = %s, want suffix -test-topic.md", path)
	}
}

func TestNoteNotInValidTypes(t *testing.T) {
	if ValidTypes["note"] {
		t.Error("note should not be in ValidTypes")
	}
}

func TestResolveAlias(t *testing.T) {
	canon, label, ok := ResolveAlias("survey")
	if !ok || canon != "reference" || label != "research" {
		t.Errorf("ResolveAlias(survey) = (%q, %q, %v), want (reference, research, true)", canon, label, ok)
	}
	_, _, ok = ResolveAlias("spike")
	if ok {
		t.Error("ResolveAlias(spike) should return false")
	}
}

func TestStageForType(t *testing.T) {
	cases := map[string]LifecycleStage{
		"spike":   StageSpike,
		"plan":    StagePlan,
		"brief":   StagePlan,
		"design":  StageDesign,
		"retro":   StageRetro,
		"survey":  StageReference,
		"unknown": StageReference,
	}
	for typ, want := range cases {
		if got := StageForType(typ); got != want {
			t.Errorf("StageForType(%q) = %d, want %d", typ, got, want)
		}
	}
}

func TestInferType(t *testing.T) {
	cases := map[string]string{
		"reference/spike/foo.md":               "spike",
		"work/active/2026-01-01-bar/README.md": "plan",
		"README.md":                            "",
		// Colocated types inside work dirs
		"work/active/2026-01-01-bar/reference/design/2026-01-01-bar.md": "design",
		"work/active/2026-01-01-bar/reference/retro/2026-01-01-bar.md":  "retro",
		"work/active/2026-01-01-bar/retro.md":                           "retro",
		"work/active/2026-01-01-bar/design.md":                          "design",
		// Session logs and track files default to plan
		"work/active/2026-01-01-bar/session-1.md": "plan",
		"work/active/2026-01-01-bar/track-1.md":   "plan",
		// Archive paths work the same
		"work/archive/2026-01-01-old/README.md":                          "plan",
		"work/archive/2026-01-01-old/reference/design/2026-01-01-old.md": "design",
	}
	for path, want := range cases {
		if got := InferType(path); got != want {
			t.Errorf("InferType(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestIsReferenceDirNewNames(t *testing.T) {
	if !IsReferenceDir("research") {
		t.Error("IsReferenceDir(research) = false, want true")
	}
	if !IsReferenceDir("insight") {
		t.Error("IsReferenceDir(insight) = false, want true")
	}
}

func TestValidTypesIncludesPlan(t *testing.T) {
	if !ValidTypes["plan"] {
		t.Error("plan not in ValidTypes")
	}
}

func TestIsWorkTypePlan(t *testing.T) {
	if !IsWorkType("plan") {
		t.Error("IsWorkType(plan) = false, want true")
	}
}

func TestTypePathPlan(t *testing.T) {
	p := TypePath("plan", "topic")
	if !strings.Contains(p, "work/active") || !strings.HasSuffix(p, "topic/README.md") {
		t.Errorf("TypePath(plan, topic) = %q, want work/active/...-topic/README.md", p)
	}
}

func TestTemplatePlanMatchesBrief(t *testing.T) {
	planTmpl := Template("plan", "test")
	briefTmpl := Template("brief", "test")
	if planTmpl != briefTmpl {
		t.Error("Template(plan) and Template(brief) should produce identical output")
	}
	if !strings.Contains(planTmpl, "Objective") {
		t.Error("plan template missing Objective section")
	}
}

func TestTemplateRetroHasExpectedSections(t *testing.T) {
	tmpl := Template("retro", "test")
	for _, section := range []string{"Context", "What Happened", "What Worked", "What Didn't", "Action Items"} {
		if !strings.Contains(tmpl, section) {
			t.Errorf("retro template missing section %q", section)
		}
	}
}
