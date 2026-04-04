package dev

import (
	"fmt"
	"os"
	"path/filepath"
)

const skillTemplate = `---
name: %s
description: <describe what this skill does and when to use it>
metadata:
  team: <your-team>
---

# Instructions

<your skill content here>
`

// Scaffold creates a new skill directory with a template SKILL.md.
func Scaffold(baseDir, name string) error {
	skillDir := filepath.Join(baseDir, name)

	if _, err := os.Stat(skillDir); err == nil {
		return fmt.Errorf("skill %q already exists at %s", name, skillDir)
	}

	if err := os.MkdirAll(filepath.Join(skillDir, "references"), 0o755); err != nil {
		return fmt.Errorf("creating skill dir: %w", err)
	}

	content := fmt.Sprintf(skillTemplate, name)
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing SKILL.md: %w", err)
	}

	if err := os.WriteFile(filepath.Join(skillDir, "references", ".gitkeep"), []byte{}, 0o644); err != nil {
		return fmt.Errorf("writing .gitkeep: %w", err)
	}

	return nil
}
