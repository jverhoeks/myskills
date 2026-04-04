package repo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Clone clones a git repository to dest.
func Clone(url, dest string) error {
	cmd := exec.Command("git", "clone", url, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone: %w", err)
	}
	return nil
}

// Pull runs git pull in the given repo directory.
func Pull(repoDir string) error {
	cmd := exec.Command("git", "pull", "--ff-only")
	cmd.Dir = repoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git pull: %w", err)
	}
	return nil
}

// ListSkills returns the names of all skill directories (those containing SKILL.md)
// under <repoDir>/skills/.
func ListSkills(repoDir string) ([]string, error) {
	skillsDir := filepath.Join(repoDir, "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("reading skills dir: %w", err)
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		skillFile := filepath.Join(skillsDir, e.Name(), "SKILL.md")
		if _, err := os.Stat(skillFile); err == nil {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// SkillDir returns the path to a skill directory in the repo.
func SkillDir(repoDir, name string) string {
	return filepath.Join(repoDir, "skills", name)
}

// CommitHashForPath returns the latest git commit hash that touched a given path.
func CommitHashForPath(repoDir, path string) (string, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%H", "--", path)
	cmd.Dir = repoDir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log for %s: %w", path, err)
	}
	return strings.TrimSpace(string(out)), nil
}
