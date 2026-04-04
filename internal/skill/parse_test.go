package skill

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseValid(t *testing.T) {
	dir := t.TempDir()
	content := `---
name: deploy
description: Deploy the application to production
metadata:
  team: infra
---

# Instructions

Run the deploy script.
`
	os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644)

	s, err := Parse(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if s.Name != "deploy" {
		t.Errorf("name: got %q, want %q", s.Name, "deploy")
	}
	if s.Description != "Deploy the application to production" {
		t.Errorf("description: got %q", s.Description)
	}
	if s.Metadata["team"] != "infra" {
		t.Errorf("metadata.team: got %q", s.Metadata["team"])
	}
	if s.Body == "" {
		t.Error("expected non-empty body")
	}
}

func TestParseNoFrontmatter(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("just markdown"), 0o644)

	_, err := Parse(filepath.Join(dir, "SKILL.md"))
	if err == nil {
		t.Error("expected error for missing frontmatter")
	}
}

func TestParseMissingName(t *testing.T) {
	dir := t.TempDir()
	content := `---
description: Some skill
---

Body here.
`
	os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(content), 0o644)

	s, err := Parse(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if s.Name != "" {
		t.Errorf("expected empty name, got %q", s.Name)
	}
}
