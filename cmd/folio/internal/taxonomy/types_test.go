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
	if !IsWorkType("spike") {
		t.Error("IsWorkType(spike) = false, want true")
	}
	if !IsWorkType("retro") {
		t.Error("IsWorkType(retro) = false, want true")
	}
}

func TestTypePathSpike(t *testing.T) {
	path := TypePath("spike", "test-topic")
	if !strings.HasPrefix(path, "work/active/") {
		t.Errorf("TypePath(spike) = %s, want prefix work/active/", path)
	}
	if !strings.Contains(path, "/spike/") {
		t.Errorf("TypePath(spike) = %s, want to contain /spike/", path)
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
	// spike and retro are no longer reference dirs — they belong in work/
	if IsReferenceDir("spike") {
		t.Error("IsReferenceDir(spike) = true, want false")
	}
	if IsReferenceDir("retro") {
		t.Error("IsReferenceDir(retro) = true, want false")
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

func TestTypePathRetro(t *testing.T) {
	path := TypePath("retro", "session-findings")
	if !strings.HasPrefix(path, "work/active/") {
		t.Errorf("TypePath(retro) = %s, want prefix work/active/", path)
	}
	if !strings.Contains(path, "/retro/") {
		t.Errorf("TypePath(retro) = %s, want to contain /retro/", path)
	}
	if !strings.HasSuffix(path, "-session-findings.md") {
		t.Errorf("TypePath(retro) = %s, want suffix -session-findings.md", path)
	}
}

func TestTypePathDesignUnchanged(t *testing.T) {
	path := TypePath("design", "topic")
	if !strings.HasPrefix(path, "reference/design/") {
		t.Errorf("TypePath(design) = %s, want prefix reference/design/ (regression)", path)
	}
}

func TestTypePathPlanUnchanged(t *testing.T) {
	path := TypePath("plan", "topic")
	if !strings.HasPrefix(path, "work/active/") {
		t.Errorf("TypePath(plan) = %s, want prefix work/active/ (regression)", path)
	}
	if !strings.HasSuffix(path, "/README.md") {
		t.Errorf("TypePath(plan) = %s, want suffix /README.md (regression)", path)
	}
}

func TestTypePathSurveyUnchanged(t *testing.T) {
	path := TypePath("survey", "topic")
	if !strings.HasPrefix(path, "reference/survey/") {
		t.Errorf("TypePath(survey) = %s, want prefix reference/survey/ (regression)", path)
	}
}

func TestIsReferenceTypeSpikeFalse(t *testing.T) {
	if IsReferenceType("spike") {
		t.Error("IsReferenceType(spike) = true, want false")
	}
}

func TestIsReferenceTypeRetroFalse(t *testing.T) {
	if IsReferenceType("retro") {
		t.Error("IsReferenceType(retro) = true, want false")
	}
}

func TestIsReferenceTypeSurveyTrue(t *testing.T) {
	if !IsReferenceType("survey") {
		t.Error("IsReferenceType(survey) = false, want true (regression)")
	}
}

func TestIsReferenceTypeGuideTrue(t *testing.T) {
	if !IsReferenceType("guide") {
		t.Error("IsReferenceType(guide) = false, want true (regression)")
	}
}

func TestValidTypesIncludesSpike(t *testing.T) {
	if !ValidTypes["spike"] {
		t.Error("ValidTypes[spike] = false, want true")
	}
}

func TestValidTypesIncludesRetro(t *testing.T) {
	if !ValidTypes["retro"] {
		t.Error("ValidTypes[retro] = false, want true")
	}
}

func TestInferTypeFlatSubdirectories(t *testing.T) {
	cases := map[string]string{
		"work/active/2026-01-01-foo/spike/2026-01-01-foo.md":  "spike",
		"work/active/2026-01-01-foo/retro/2026-01-01-foo.md":  "retro",
		"work/active/2026-01-01-foo/design/2026-01-01-foo.md": "design",
		"work/archive/2026-01-01-foo/spike/2026-01-01-foo.md": "spike",
		// Old patterns still work
		"work/active/2026-01-01-bar/retro.md": "retro",
	}
	for path, want := range cases {
		if got := InferType(path); got != want {
			t.Errorf("InferType(%q) = %q, want %q", path, got, want)
		}
	}
}

func TestIsReferenceDirSurveyStillTrue(t *testing.T) {
	if !IsReferenceDir("survey") {
		t.Error("IsReferenceDir(survey) = false, want true (regression)")
	}
}

func TestColocatableTypesRetroStillTrue(t *testing.T) {
	if !ColocatableTypes["retro"] {
		t.Error("ColocatableTypes[retro] = false, want true (kept for old single-file pattern)")
	}
}

func TestStageForTypeSketch(t *testing.T) {
	if StageForType("sketch") != StageIdea {
		t.Error("StageForType(sketch) != StageIdea")
	}
	if !(StageSpike < StageIdea && StageIdea < StageDesign) {
		t.Error("StageIdea should be ordered between StageSpike and StageDesign")
	}
}

func TestSketchIsColocatableNotReference(t *testing.T) {
	if !ColocatableTypes["sketch"] {
		t.Error("ColocatableTypes[sketch] = false, want true")
	}
	if !ValidTypes["sketch"] {
		t.Error("ValidTypes[sketch] = false, want true")
	}
	if IsReferenceType("sketch") {
		t.Error("IsReferenceType(sketch) = true, want false")
	}
	if !IsReferenceDir("sketch") {
		t.Error("IsReferenceDir(sketch) = false, want true (colocates under reference/sketch/)")
	}
}

func TestInferTypeSketch(t *testing.T) {
	got := InferType("work/active/2026-07-10-x/reference/sketch/index.html")
	if got != "sketch" {
		t.Errorf("InferType(...reference/sketch/index.html) = %q, want %q", got, "sketch")
	}
}

// regression: schema-2 labels must be valid types
func TestValidTypesIncludesReferenceLabels(t *testing.T) {
	for _, l := range []string{"research", "insight"} {
		if !ValidTypes[l] {
			t.Errorf("ValidTypes[%q] = false, want true", l)
		}
	}
}
