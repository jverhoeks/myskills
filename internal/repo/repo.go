package repo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Standard directories where skills can live in a repo,
// matching the skills.sh / agentskills.io discovery order.
var skillSearchDirs = []string{
	"skills",
	".agents/skills",
	".claude/skills",
	".cursor/skills",
	".copilot/skills",
	".github/skills",
}

var ownerRepoPattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+$`)

// ResolveURL expands shorthand repo references to full URLs.
//   - "owner/repo" → "https://github.com/owner/repo.git"
//   - full URLs pass through unchanged
func ResolveURL(input string) string {
	if ownerRepoPattern.MatchString(input) {
		return "https://github.com/" + input + ".git"
	}
	return input
}

// NameFromURL derives a short name from a repo URL or owner/repo shorthand.
//   - "owner/repo" → "repo"
//   - "https://github.com/owner/repo.git" → "repo"
func NameFromURL(input string) string {
	// Handle owner/repo shorthand
	if ownerRepoPattern.MatchString(input) {
		parts := strings.SplitN(input, "/", 2)
		return parts[1]
	}
	// Extract from URL
	base := filepath.Base(input)
	base = strings.TrimSuffix(base, ".git")
	return base
}

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

// FindSkillsDir returns the first directory in the repo that contains skills.
// Searches standard locations in order. Returns "" if none found.
func FindSkillsDir(repoDir string) string {
	// Check if root itself is a single skill (SKILL.md at root)
	if _, err := os.Stat(filepath.Join(repoDir, "SKILL.md")); err == nil {
		return "" // Special case: root is a skill, not a directory of skills
	}

	for _, dir := range skillSearchDirs {
		candidate := filepath.Join(repoDir, dir)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			// Check it actually contains at least one skill
			entries, err := os.ReadDir(candidate)
			if err != nil {
				continue
			}
			for _, e := range entries {
				if e.IsDir() {
					if _, err := os.Stat(filepath.Join(candidate, e.Name(), "SKILL.md")); err == nil {
						return candidate
					}
				}
			}
		}
	}
	return ""
}

// IsRootSkill returns true if the repo root itself is a single skill (SKILL.md at root).
func IsRootSkill(repoDir string) bool {
	_, err := os.Stat(filepath.Join(repoDir, "SKILL.md"))
	return err == nil
}

// ListSkills returns the names of all skill directories found in the repo.
// Searches standard skill directories (skills/, .agents/skills/, .claude/skills/, etc.).
// Also handles repos where the root itself is a single skill.
func ListSkills(repoDir string) ([]string, error) {
	// Special case: root is a single skill
	if IsRootSkill(repoDir) {
		name := filepath.Base(repoDir)
		return []string{name}, nil
	}

	skillsDir := FindSkillsDir(repoDir)
	if skillsDir == "" {
		return nil, nil
	}

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
	// Root skill
	if IsRootSkill(repoDir) && filepath.Base(repoDir) == name {
		return repoDir
	}

	skillsDir := FindSkillsDir(repoDir)
	if skillsDir == "" {
		// Fallback to the old default
		return filepath.Join(repoDir, "skills", name)
	}
	return filepath.Join(skillsDir, name)
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
