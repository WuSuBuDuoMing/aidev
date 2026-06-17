// Package skill provides the skill discovery and loading system.
package skill

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Skill represents a loaded skill.
type Skill struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Trigger     string   `json:"trigger"`
	Prompt      string   `json:"prompt"`
	Tools       []string `json:"tools,omitempty"`
	FilePath    string   `json:"filePath"`
}

// Discover finds all skills in the given directories.
func Discover(dirs ...string) []Skill {
	var skills []Skill
	seen := make(map[string]bool)

	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
				continue
			}
			path := filepath.Join(dir, entry.Name())
			skill, err := Load(path)
			if err != nil {
				continue
			}
			// Project-level skills override global skills with the same name
			if !seen[skill.Name] {
				skills = append(skills, *skill)
				seen[skill.Name] = true
			}
		}
	}

	return skills
}

// Load reads a skill file (YAML frontmatter + markdown body).
func Load(path string) (*Skill, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	skill := &Skill{FilePath: path}
	scanner := bufio.NewScanner(file)
	inFrontmatter := false
	frontmatterDone := false
	var bodyLines []string

	for scanner.Scan() {
		line := scanner.Text()

		// YAML frontmatter delimiter
		if line == "---" {
			if !inFrontmatter && !frontmatterDone {
				inFrontmatter = true
				continue
			}
			if inFrontmatter {
				inFrontmatter = false
				frontmatterDone = true
				continue
			}
		}

		if inFrontmatter {
			// Parse simple YAML key: value
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.TrimSpace(parts[1])
				switch key {
				case "name":
					skill.Name = val
				case "description":
					skill.Description = val
				case "trigger":
					skill.Trigger = val
				case "tools":
					// Parse [tool1, tool2] format
					val = strings.Trim(val, "[]")
					for _, t := range strings.Split(val, ",") {
						skill.Tools = append(skill.Tools, strings.TrimSpace(t))
					}
				}
			}
			continue
		}

		if frontmatterDone {
			bodyLines = append(bodyLines, line)
		}
	}

	if !frontmatterDone {
		return nil, fmt.Errorf("invalid skill file (missing frontmatter): %s", path)
	}

	skill.Prompt = strings.TrimSpace(strings.Join(bodyLines, "\n"))

	// Default name from filename
	if skill.Name == "" {
		skill.Name = strings.TrimSuffix(filepath.Base(path), ".md")
	}

	return skill, nil
}

// Resolve finds a skill by trigger command or name.
func Resolve(skills []Skill, input string) *Skill {
	input = strings.TrimSpace(input)

	for _, s := range skills {
		if s.Trigger != "" && input == s.Trigger {
			return &s
		}
		if input == "/"+s.Name {
			return &s
		}
	}

	return nil
}

// FormatList returns a formatted string listing all skills.
func FormatList(skills []Skill) string {
	if len(skills) == 0 {
		return "  No skills loaded."
	}

	var lines []string
	lines = append(lines, "  Available Skills:")
	lines = append(lines, "  ─────────────────────────────────────────")
	for _, s := range skills {
		trigger := s.Trigger
		if trigger == "" {
			trigger = "/" + s.Name
		}
		lines = append(lines, fmt.Sprintf("  %-20s %s", trigger, s.Description))
	}
	return strings.Join(lines, "\n")
}
