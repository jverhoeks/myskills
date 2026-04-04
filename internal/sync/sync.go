package sync

import (
	"fmt"
	"os"
	"path/filepath"
)

// LinkSkill creates a symlink from dst pointing to src.
// Removes any existing dst (file, dir, or old symlink) first.
func LinkSkill(src, dst string) error {
	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("creating parent dir: %w", err)
	}

	// Remove existing (could be old symlink, dir, or file)
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("removing old skill at %s: %w", dst, err)
	}

	// Resolve src to absolute path for the symlink
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("resolving absolute path: %w", err)
	}

	if err := os.Symlink(absSrc, dst); err != nil {
		return fmt.Errorf("creating symlink %s -> %s: %w", dst, absSrc, err)
	}
	return nil
}

// RemoveSkill removes a skill (symlink or directory).
func RemoveSkill(skillPath string) error {
	// os.Remove works for symlinks, os.RemoveAll for dirs
	info, err := os.Lstat(skillPath)
	if err != nil {
		return fmt.Errorf("stat %s: %w", skillPath, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return os.Remove(skillPath)
	}
	return os.RemoveAll(skillPath)
}

// All syncs all skills from repoDir/skills/ to each target path using symlinks.
// enabledFilter: if non-nil, only skills where enabledFilter(name) returns true are synced.
// Returns the number of skills synced.
func All(repoDir string, targets map[string]string, enabledFilter func(string) bool) (int, error) {
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
		name := e.Name()
		src := filepath.Join(skillsDir, name)
		if _, err := os.Stat(filepath.Join(src, "SKILL.md")); err != nil {
			continue
		}

		if enabledFilter != nil && !enabledFilter(name) {
			// Skill is disabled — remove any existing symlinks
			for _, targetPath := range targets {
				dst := filepath.Join(targetPath, name)
				if _, err := os.Lstat(dst); err == nil {
					os.Remove(dst)
				}
			}
			continue
		}

		for _, targetPath := range targets {
			dst := filepath.Join(targetPath, name)
			if err := LinkSkill(src, dst); err != nil {
				return count, fmt.Errorf("linking %s: %w", name, err)
			}
		}
		count++
	}
	return count, nil
}

// One syncs a single named skill to all targets using symlinks.
func One(repoDir, name string, targets map[string]string) error {
	src := filepath.Join(repoDir, "skills", name)
	if _, err := os.Stat(filepath.Join(src, "SKILL.md")); err != nil {
		return fmt.Errorf("skill %q not found in repo", name)
	}

	for _, targetPath := range targets {
		dst := filepath.Join(targetPath, name)
		if err := LinkSkill(src, dst); err != nil {
			return fmt.Errorf("linking %s: %w", name, err)
		}
	}
	return nil
}

// CopySkill copies a skill directory from src to dst (used by submit workflow).
// Removes dst first to ensure clean state.
func CopySkill(src, dst string) error {
	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("removing old skill: %w", err)
	}

	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
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

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}
