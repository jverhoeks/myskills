package dev

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffold(t *testing.T) {
	dir := t.TempDir()

	if err := Scaffold(dir, "my-skill"); err != nil {
		t.Fatalf("scaffold: %v", err)
	}

	skillDir := filepath.Join(dir, "my-skill")
	skillFile := filepath.Join(skillDir, "SKILL.md")

	if _, err := os.Stat(skillFile); err != nil {
		t.Fatal("expected SKILL.md")
	}

	data, _ := os.ReadFile(skillFile)
	content := string(data)

	if !strings.Contains(content, "name: my-skill") {
		t.Error("expected name in template")
	}
	if !strings.Contains(content, "description:") {
		t.Error("expected description placeholder in template")
	}
	if !strings.Contains(content, "team:") {
		t.Error("expected team placeholder in template")
	}
}

func TestScaffoldAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "my-skill"), 0o755)

	err := Scaffold(dir, "my-skill")
	if err == nil {
		t.Error("expected error when skill dir already exists")
	}
}
