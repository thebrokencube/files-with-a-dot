package lint

import (
	"testing"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik/conventions"
)

func TestSkillLint_NoSkillMD(t *testing.T) {
	data := &ToolData{ToolName: "test", SkillMD: nil}
	results := SkillLint(data)
	assertCheckPresent(t, results, "skill-exists")
	if len(results) != 1 {
		t.Errorf("expected 1 result (structure gate), got %d", len(results))
	}
}

func TestSkillLint_ArgumentHint(t *testing.T) {
	t.Run("invocable without hint", func(t *testing.T) {
		data := skillToolData("test")
		data.SkillMD = []byte("---\nname: test\ndescription: \"Use when testing things\"\nuser_invocable: true\n---\n# Test\n")
		results := filterCheck(SkillLint(data), "argument-hint")
		assertCheckPresent(t, results, "argument-hint")
	})

	t.Run("invocable with hint", func(t *testing.T) {
		data := skillToolData("test")
		data.SkillMD = []byte("---\nname: test\ndescription: \"Use when testing things\"\nuser_invocable: true\nargument-hint: \"<cmd> [flags]\"\n---\n# Test\n")
		results := filterCheck(SkillLint(data), "argument-hint")
		if len(results) > 0 {
			t.Errorf("expected no argument-hint errors, got %v", results)
		}
	})

	t.Run("not invocable", func(t *testing.T) {
		data := skillToolData("test")
		data.SkillMD = []byte("---\nname: test\ndescription: \"Use when testing things\"\nuser_invocable: false\n---\n# Test\n")
		results := filterCheck(SkillLint(data), "argument-hint")
		if len(results) > 0 {
			t.Errorf("expected no argument-hint errors when not invocable, got %v", results)
		}
	})
}

func TestSkillLint_ArrowRefs(t *testing.T) {
	t.Run("valid arrow ref", func(t *testing.T) {
		data := skillToolData("test")
		data.SkillMD = []byte("---\nname: test\ndescription: \"Use when testing\"\n---\n-> Read references/guide.md\n")
		data.RefFiles = []string{"guide.md"}
		data.RefContents = map[string][]byte{"guide.md": []byte("# Guide")}
		results := filterCheck(SkillLint(data), "arrow-refs")
		if len(results) > 0 {
			t.Errorf("expected no arrow-refs errors, got %v", results)
		}
	})

	t.Run("broken arrow ref", func(t *testing.T) {
		data := skillToolData("test")
		data.SkillMD = []byte("---\nname: test\ndescription: \"Use when testing\"\n---\n-> Read references/missing.md\n")
		data.RefContents = map[string][]byte{}
		results := filterCheck(SkillLint(data), "arrow-refs")
		assertCheckPresent(t, results, "arrow-refs")
	})

	t.Run("arrow ref in reference file", func(t *testing.T) {
		data := skillToolData("test")
		data.SkillMD = []byte("---\nname: test\ndescription: \"Use when testing\"\n---\n# Body\n")
		data.RefFiles = []string{"guide.md"}
		data.RefContents = map[string][]byte{
			"guide.md": []byte("-> Read references/missing.md"),
		}
		results := filterCheck(SkillLint(data), "arrow-refs")
		assertCheckPresent(t, results, "arrow-refs")
	})
}

func TestSkillLint_ActivationGuidance(t *testing.T) {
	t.Run("has guidance", func(t *testing.T) {
		data := skillToolData("test")
		data.SkillMD = []byte("---\nname: test\ndescription: \"Use when you need to test things\"\n---\n# Test\n")
		results := filterCheck(SkillLint(data), "activation-guidance")
		if len(results) > 0 {
			t.Errorf("expected no activation-guidance warnings, got %v", results)
		}
	})

	t.Run("missing guidance", func(t *testing.T) {
		data := skillToolData("test")
		data.SkillMD = []byte("---\nname: test\ndescription: \"A tool for testing\"\n---\n# Test\n")
		results := filterCheck(SkillLint(data), "activation-guidance")
		assertCheckPresent(t, results, "activation-guidance")
		if results[0].Severity != conventions.SeverityWarning {
			t.Errorf("activation-guidance should be warning, got %s", results[0].Severity)
		}
	})
}

func TestSkillLint_ActivationMetadata(t *testing.T) {
	t.Run("valid trigger", func(t *testing.T) {
		data := skillToolData("test")
		data.SkillMD = []byte("---\nname: test\ndescription: \"Use when testing\"\ntrigger: \"test keyword\"\n---\n# Test\n")
		results := filterCheck(SkillLint(data), "activation-metadata")
		if len(results) > 0 {
			t.Errorf("expected no activation-metadata errors, got %v", results)
		}
	})

	t.Run("empty trigger", func(t *testing.T) {
		data := skillToolData("test")
		data.SkillMD = []byte("---\nname: test\ndescription: \"Use when testing\"\ntrigger: \"\"\n---\n# Test\n")
		results := filterCheck(SkillLint(data), "activation-metadata")
		assertCheckPresent(t, results, "activation-metadata")
	})

	t.Run("no trigger field", func(t *testing.T) {
		data := skillToolData("test")
		data.SkillMD = []byte("---\nname: test\ndescription: \"Use when testing\"\n---\n# Test\n")
		results := filterCheck(SkillLint(data), "activation-metadata")
		if len(results) > 0 {
			t.Errorf("expected no errors when trigger absent, got %v", results)
		}
	})
}

// --- helpers ---

func skillToolData(name string) *ToolData {
	return &ToolData{
		ToolName:    name,
		SkillMD:     []byte("---\nname: " + name + "\ndescription: \"Use when testing\"\n---\n# " + name + "\n"),
		SkillDir:    "/tmp/skill",
		RefFiles:    []string{},
		RefContents: map[string][]byte{},
	}
}
