package sync

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// CopySkill copies a skill directory from src to dst.
// Removes dst first to ensure clean state (no stale files).
func CopySkill(src, dst string) error {
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("removing old skill: %w", err)
	}

	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		return copyFile(path, target)
	})
}

// RemoveSkill removes a skill directory.
func RemoveSkill(skillDir string) error {
	return os.RemoveAll(skillDir)
}

// All syncs all skills from repoDir/skills/ to each target path.
// Returns the number of skills synced.
func All(repoDir string, targets map[string]string) (int, error) {
	skillsDir := filepath.Join(repoDir, "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return 0, fmt.Errorf("reading skills dir: %w", err)
	}

	count := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		src := filepath.Join(skillsDir, e.Name())
		if _, err := os.Stat(filepath.Join(src, "SKILL.md")); err != nil {
			continue
		}

		for _, targetPath := range targets {
			dst := filepath.Join(targetPath, e.Name())
			if err := CopySkill(src, dst); err != nil {
				return count, fmt.Errorf("syncing %s: %w", e.Name(), err)
			}
		}
		count++
	}
	return count, nil
}

// One syncs a single named skill to all targets.
func One(repoDir, name string, targets map[string]string) error {
	src := filepath.Join(repoDir, "skills", name)
	if _, err := os.Stat(filepath.Join(src, "SKILL.md")); err != nil {
		return fmt.Errorf("skill %q not found in repo", name)
	}

	for _, targetPath := range targets {
		dst := filepath.Join(targetPath, name)
		if err := CopySkill(src, dst); err != nil {
			return fmt.Errorf("syncing %s: %w", name, err)
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
