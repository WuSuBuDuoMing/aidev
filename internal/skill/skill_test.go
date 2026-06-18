package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()
	skillFile := filepath.Join(tmpDir, "test-skill.md")

	content := `---
name: test-skill
description: A test skill
trigger: /test
tools: [readFile, grep]
---

You are a test assistant.

Check {{file}} for issues.`

	if err := os.WriteFile(skillFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	skill, err := Load(skillFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if skill.Name != "test-skill" {
		t.Errorf("Name = %q, want %q", skill.Name, "test-skill")
	}
	if skill.Description != "A test skill" {
		t.Errorf("Description = %q, want %q", skill.Description, "A test skill")
	}
	if skill.Trigger != "/test" {
		t.Errorf("Trigger = %q, want %q", skill.Trigger, "/test")
	}
	if len(skill.Tools) != 2 {
		t.Errorf("Tools count = %d, want 2", len(skill.Tools))
	}
	if skill.Prompt == "" {
		t.Error("Prompt should not be empty")
	}
}

func TestLoad_NoFrontmatter(t *testing.T) {
	tmpDir := t.TempDir()
	skillFile := filepath.Join(tmpDir, "bad-skill.md")

	if err := os.WriteFile(skillFile, []byte("no frontmatter here"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(skillFile)
	if err == nil {
		t.Error("Load should fail for skill without frontmatter")
	}
}

func TestLoad_DefaultName(t *testing.T) {
	tmpDir := t.TempDir()
	skillFile := filepath.Join(tmpDir, "auto-name.md")

	content := `---
description: A skill with auto name
---

Content here.`

	if err := os.WriteFile(skillFile, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	skill, err := Load(skillFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if skill.Name != "auto-name" {
		t.Errorf("Name = %q, want %q (from filename)", skill.Name, "auto-name")
	}
}

func TestDiscover(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a skill file
	skillContent := `---
name: discovered
description: Found by discover
---

Content.`

	if err := os.WriteFile(filepath.Join(tmpDir, "discovered.md"), []byte(skillContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create a non-skill file
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("not a skill"), 0o644); err != nil {
		t.Fatal(err)
	}

	skills := Discover(tmpDir)
	if len(skills) != 1 {
		t.Errorf("Discovered %d skills, want 1", len(skills))
	}
	if len(skills) > 0 && skills[0].Name != "discovered" {
		t.Errorf("Skill name = %q, want %q", skills[0].Name, "discovered")
	}
}

func TestDiscover_NonexistentDir(t *testing.T) {
	skills := Discover("/nonexistent/path")
	if len(skills) != 0 {
		t.Error("Should return empty for nonexistent directory")
	}
}

func TestDiscover_ProjectOverridesGlobal(t *testing.T) {
	globalDir := t.TempDir()
	projectDir := t.TempDir()

	globalContent := `---
name: my-skill
description: Global version
---
Global.`
	projectContent := `---
name: my-skill
description: Project version
---
Project.`

	os.WriteFile(filepath.Join(globalDir, "my-skill.md"), []byte(globalContent), 0o644)
	os.WriteFile(filepath.Join(projectDir, "my-skill.md"), []byte(projectContent), 0o644)

	skills := Discover(globalDir, projectDir)
	if len(skills) != 1 {
		t.Errorf("Expected 1 skill, got %d", len(skills))
	}
	if len(skills) > 0 && skills[0].Description != "Global version" {
		t.Errorf("First discovered wins; got %q", skills[0].Description)
	}
}

func TestResolve(t *testing.T) {
	skills := []Skill{
		{Name: "review", Trigger: "/review"},
		{Name: "test"},
	}

	// Match by trigger
	s := Resolve(skills, "/review")
	if s == nil || s.Name != "review" {
		t.Error("Should resolve by trigger")
	}

	// Match by name
	s = Resolve(skills, "/test")
	if s == nil || s.Name != "test" {
		t.Error("Should resolve by /name")
	}

	// No match
	s = Resolve(skills, "/nonexistent")
	if s != nil {
		t.Error("Should not resolve nonexistent")
	}
}

func TestFormatList(t *testing.T) {
	// Empty
	result := FormatList(nil)
	if result == "" {
		t.Error("FormatList(nil) should not be empty")
	}

	// With skills
	skills := []Skill{
		{Name: "review", Description: "Code review", Trigger: "/review"},
	}
	result = FormatList(skills)
	if result == "" {
		t.Error("FormatList should not be empty with skills")
	}
}
