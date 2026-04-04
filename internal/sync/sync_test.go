package sync

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLinkSkill(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	skillSrc := filepath.Join(src, "deploy")
	os.MkdirAll(filepath.Join(skillSrc, "scripts"), 0o755)
	os.WriteFile(filepath.Join(skillSrc, "SKILL.md"), []byte("---\nname: deploy\n---\n"), 0o644)
	os.WriteFile(filepath.Join(skillSrc, "scripts", "run.sh"), []byte("#!/bin/bash\necho hi"), 0o644)

	linkDst := filepath.Join(dst, "deploy")

	if err := LinkSkill(skillSrc, linkDst); err != nil {
		t.Fatalf("link: %v", err)
	}

	info, err := os.Lstat(linkDst)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected symlink")
	}

	if _, err := os.Stat(filepath.Join(linkDst, "SKILL.md")); err != nil {
		t.Error("expected SKILL.md accessible via symlink")
	}
	if _, err := os.Stat(filepath.Join(linkDst, "scripts", "run.sh")); err != nil {
		t.Error("expected scripts/run.sh accessible via symlink")
	}
}

func TestLinkSkillOverwrites(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	skillSrc := filepath.Join(src, "deploy")
	os.MkdirAll(skillSrc, 0o755)
	os.WriteFile(filepath.Join(skillSrc, "SKILL.md"), []byte("new content"), 0o644)

	linkDst := filepath.Join(dst, "deploy")
	os.MkdirAll(linkDst, 0o755)
	os.WriteFile(filepath.Join(linkDst, "old.md"), []byte("old"), 0o644)

	if err := LinkSkill(skillSrc, linkDst); err != nil {
		t.Fatalf("link: %v", err)
	}

	info, err := os.Lstat(linkDst)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected symlink after overwrite")
	}
}

func TestRemoveSkillSymlink(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	os.MkdirAll(filepath.Join(src, "deploy"), 0o755)
	linkDst := filepath.Join(dst, "deploy")
	os.Symlink(filepath.Join(src, "deploy"), linkDst)

	if err := RemoveSkill(linkDst); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if _, err := os.Lstat(linkDst); err == nil {
		t.Error("expected symlink removed")
	}
}

func TestAllFromMap(t *testing.T) {
	src := t.TempDir()
	target1 := t.TempDir()
	target2 := t.TempDir()

	for _, name := range []string{"alpha", "beta"} {
		d := filepath.Join(src, name)
		os.MkdirAll(d, 0o755)
		os.WriteFile(filepath.Join(d, "SKILL.md"), []byte("---\nname: "+name+"\n---\n"), 0o644)
	}

	skillMap := map[string]string{
		"alpha": filepath.Join(src, "alpha"),
		"beta":  filepath.Join(src, "beta"),
	}
	targets := map[string]string{
		"claude":  filepath.Join(target1, "skills"),
		"copilot": filepath.Join(target2, "skills"),
	}

	count, err := AllFromMap(skillMap, targets, nil)
	if err != nil {
		t.Fatalf("sync all: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 skills synced, got %d", count)
	}

	for _, tPath := range targets {
		for _, sName := range []string{"alpha", "beta"} {
			linkPath := filepath.Join(tPath, sName)
			info, err := os.Lstat(linkPath)
			if err != nil {
				t.Errorf("expected symlink at %s", linkPath)
				continue
			}
			if info.Mode()&os.ModeSymlink == 0 {
				t.Errorf("expected %s to be a symlink", linkPath)
			}
		}
	}
}

func TestAllFromMapWithFilter(t *testing.T) {
	src := t.TempDir()
	target := t.TempDir()

	for _, name := range []string{"alpha", "beta", "gamma"} {
		d := filepath.Join(src, name)
		os.MkdirAll(d, 0o755)
		os.WriteFile(filepath.Join(d, "SKILL.md"), []byte("x"), 0o644)
	}

	skillMap := map[string]string{
		"alpha": filepath.Join(src, "alpha"),
		"beta":  filepath.Join(src, "beta"),
		"gamma": filepath.Join(src, "gamma"),
	}
	targets := map[string]string{"claude": filepath.Join(target, "skills")}

	filter := func(name string) bool {
		return name == "alpha" || name == "gamma"
	}

	count, err := AllFromMap(skillMap, targets, filter)
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 skills synced, got %d", count)
	}

	if _, err := os.Lstat(filepath.Join(target, "skills", "alpha")); err != nil {
		t.Error("expected alpha synced")
	}
	if _, err := os.Lstat(filepath.Join(target, "skills", "beta")); err == nil {
		t.Error("expected beta NOT synced")
	}
}
