package sync

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopySkill(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	skillSrc := filepath.Join(src, "deploy")
	os.MkdirAll(filepath.Join(skillSrc, "scripts"), 0o755)
	os.WriteFile(filepath.Join(skillSrc, "SKILL.md"), []byte("---\nname: deploy\n---\n"), 0o644)
	os.WriteFile(filepath.Join(skillSrc, "scripts", "run.sh"), []byte("#!/bin/bash\necho hi"), 0o644)

	skillDst := filepath.Join(dst, "deploy")

	if err := CopySkill(skillSrc, skillDst); err != nil {
		t.Fatalf("copy: %v", err)
	}

	if _, err := os.Stat(filepath.Join(skillDst, "SKILL.md")); err != nil {
		t.Error("expected SKILL.md in dest")
	}
	if _, err := os.Stat(filepath.Join(skillDst, "scripts", "run.sh")); err != nil {
		t.Error("expected scripts/run.sh in dest")
	}
}

func TestCopySkillOverwrites(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	skillSrc := filepath.Join(src, "deploy")
	os.MkdirAll(skillSrc, 0o755)
	os.WriteFile(filepath.Join(skillSrc, "SKILL.md"), []byte("new content"), 0o644)

	skillDst := filepath.Join(dst, "deploy")
	os.MkdirAll(skillDst, 0o755)
	os.WriteFile(filepath.Join(skillDst, "SKILL.md"), []byte("old content"), 0o644)
	os.WriteFile(filepath.Join(skillDst, "stale.md"), []byte("stale"), 0o644)

	if err := CopySkill(skillSrc, skillDst); err != nil {
		t.Fatalf("copy: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(skillDst, "SKILL.md"))
	if string(data) != "new content" {
		t.Errorf("expected overwritten content, got %q", data)
	}
	if _, err := os.Stat(filepath.Join(skillDst, "stale.md")); err == nil {
		t.Error("expected stale.md to be removed")
	}
}

func TestRemoveSkill(t *testing.T) {
	dir := t.TempDir()
	skillDir := filepath.Join(dir, "deploy")
	os.MkdirAll(skillDir, 0o755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("x"), 0o644)

	if err := RemoveSkill(skillDir); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if _, err := os.Stat(skillDir); err == nil {
		t.Error("expected skill dir to be removed")
	}
}

func TestSyncAll(t *testing.T) {
	repoDir := t.TempDir()
	target1 := t.TempDir()
	target2 := t.TempDir()

	for _, name := range []string{"alpha", "beta"} {
		d := filepath.Join(repoDir, "skills", name)
		os.MkdirAll(d, 0o755)
		os.WriteFile(filepath.Join(d, "SKILL.md"), []byte("---\nname: "+name+"\n---\n"), 0o644)
	}

	targets := map[string]string{
		"claude":  filepath.Join(target1, "skills"),
		"copilot": filepath.Join(target2, "skills"),
	}

	count, err := All(repoDir, targets)
	if err != nil {
		t.Fatalf("sync all: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 skills synced, got %d", count)
	}

	for tName, tPath := range targets {
		for _, sName := range []string{"alpha", "beta"} {
			if _, err := os.Stat(filepath.Join(tPath, sName, "SKILL.md")); err != nil {
				t.Errorf("expected %s/%s/SKILL.md in %s", tPath, sName, tName)
			}
		}
	}
}
