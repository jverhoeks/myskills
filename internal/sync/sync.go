package sync

import (
	"fmt"
	"os"
	"path/filepath"
)

// LinkSkill creates a symlink from dst pointing to src.
// Removes any existing dst (file, dir, or old symlink) first.
func LinkSkill(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("creating parent dir: %w", err)
	}

	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("removing old skill at %s: %w", dst, err)
	}

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
	info, err := os.Lstat(skillPath)
	if err != nil {
		return fmt.Errorf("stat %s: %w", skillPath, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return os.Remove(skillPath)
	}
	return os.RemoveAll(skillPath)
}

// AllFromMap syncs skills given as name→srcPath to each target using symlinks.
// enabledFilter: if non-nil, only skills where enabledFilter(name) returns true are synced.
// Returns the number of skills synced.
func AllFromMap(skills map[string]string, targets map[string]string, enabledFilter func(string) bool) (int, error) {
	count := 0
	for name, src := range skills {
		if enabledFilter != nil && !enabledFilter(name) {
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
func One(srcDir, name string, targets map[string]string) error {
	if _, err := os.Stat(filepath.Join(srcDir, "SKILL.md")); err != nil {
		return fmt.Errorf("skill %q not found at %s", name, srcDir)
	}

	for _, targetPath := range targets {
		dst := filepath.Join(targetPath, name)
		if err := LinkSkill(srcDir, dst); err != nil {
			return fmt.Errorf("linking %s: %w", name, err)
		}
	}
	return nil
}

// CopySkill copies a skill directory from src to dst (used by submit workflow).
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
