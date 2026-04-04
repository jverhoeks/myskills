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

	// Verify it's a symlink
	info, err := os.Lstat(linkDst)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected symlink")
	}

	// Verify files are accessible through the symlink
	if _, err := os.Stat(filepath.Join(linkDst, "SKILL.md")); err != nil {
		t.Error("expected SKILL.md accessible via symlink")
	}
	if _, err := os.Stat(filepath.Join(linkDst, "scripts", "run.sh")); err != nil {
		t.Error("expected scripts/run.sh accessible via symlink")
	}

	// Verify symlink target
	target, err := os.Readlink(linkDst)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	absSkillSrc, _ := filepath.Abs(skillSrc)
	if target != absSkillSrc {
		t.Errorf("symlink target: got %q, want %q", target, absSkillSrc)
	}
}

func TestLinkSkillOverwrites(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	skillSrc := filepath.Join(src, "deploy")
	os.MkdirAll(skillSrc, 0o755)
	os.WriteFile(filepath.Join(skillSrc, "SKILL.md"), []byte("new content"), 0o644)

	// Create an existing directory (not a symlink) at destination
	linkDst := filepath.Join(dst, "deploy")
	os.MkdirAll(linkDst, 0o755)
	os.WriteFile(filepath.Join(linkDst, "old.md"), []byte("old"), 0o644)

	if err := LinkSkill(skillSrc, linkDst); err != nil {
		t.Fatalf("link: %v", err)
	}

	// Should now be a symlink
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

func TestRemoveSkillDir(t *testing.T) {
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

	count, err := All(repoDir, targets, nil)
	if err != nil {
		t.Fatalf("sync all: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 skills synced, got %d", count)
	}

	// Verify symlinks exist and point to right place
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
			if _, err := os.Stat(filepath.Join(linkPath, "SKILL.md")); err != nil {
				t.Errorf("expected SKILL.md accessible via %s", linkPath)
			}
		}
	}
}

func TestSyncAllWithFilter(t *testing.T) {
	repoDir := t.TempDir()
	target := t.TempDir()

	for _, name := range []string{"alpha", "beta", "gamma"} {
		d := filepath.Join(repoDir, "skills", name)
		os.MkdirAll(d, 0o755)
		os.WriteFile(filepath.Join(d, "SKILL.md"), []byte("---\nname: "+name+"\n---\n"), 0o644)
	}

	targets := map[string]string{"claude": filepath.Join(target, "skills")}

	// Only enable alpha and gamma
	filter := func(name string) bool {
		return name == "alpha" || name == "gamma"
	}

	count, err := All(repoDir, targets, filter)
	if err != nil {
		t.Fatalf("sync all: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 skills synced, got %d", count)
	}

	// alpha should exist
	if _, err := os.Lstat(filepath.Join(target, "skills", "alpha")); err != nil {
		t.Error("expected alpha synced")
	}
	// beta should not exist
	if _, err := os.Lstat(filepath.Join(target, "skills", "beta")); err == nil {
		t.Error("expected beta NOT synced")
	}
	// gamma should exist
	if _, err := os.Lstat(filepath.Join(target, "skills", "gamma")); err != nil {
		t.Error("expected gamma synced")
	}
}

func TestSyncAllRemovesDisabled(t *testing.T) {
	repoDir := t.TempDir()
	target := t.TempDir()

	for _, name := range []string{"alpha", "beta"} {
		d := filepath.Join(repoDir, "skills", name)
		os.MkdirAll(d, 0o755)
		os.WriteFile(filepath.Join(d, "SKILL.md"), []byte("---\nname: "+name+"\n---\n"), 0o644)
	}

	targets := map[string]string{"claude": filepath.Join(target, "skills")}

	// First sync: both enabled
	All(repoDir, targets, nil)

	// Second sync: beta disabled — should remove its symlink
	filter := func(name string) bool { return name == "alpha" }
	All(repoDir, targets, filter)

	if _, err := os.Lstat(filepath.Join(target, "skills", "alpha")); err != nil {
		t.Error("expected alpha still synced")
	}
	if _, err := os.Lstat(filepath.Join(target, "skills", "beta")); err == nil {
		t.Error("expected beta removed after disable")
	}
}
