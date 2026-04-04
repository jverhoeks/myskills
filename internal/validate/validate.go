package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sbp/myskills/internal/skill"
)

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// Spec validates a skill directory against the Agent Skills specification.
// Returns a list of error descriptions (empty means valid).
func Spec(skillDir string) []string {
	var errs []string

	skillFile := filepath.Join(skillDir, "SKILL.md")
	if _, err := os.Stat(skillFile); err != nil {
		return []string{"SKILL.md not found in " + skillDir}
	}

	s, err := skill.Parse(skillFile)
	if err != nil {
		return []string{err.Error()}
	}

	dirName := filepath.Base(skillDir)

	if s.Name == "" {
		errs = append(errs, "name: required field is missing")
	} else {
		if s.Name != dirName {
			errs = append(errs, fmt.Sprintf("name: %q does not match directory name %q", s.Name, dirName))
		}
		if len(s.Name) > 64 {
			errs = append(errs, "name: exceeds 64 characters")
		}
		if !namePattern.MatchString(s.Name) {
			errs = append(errs, fmt.Sprintf("name: %q must be lowercase letters, numbers, and single hyphens", s.Name))
		}
		if strings.Contains(s.Name, "--") {
			errs = append(errs, "name: must not contain consecutive hyphens")
		}
	}

	if s.Description == "" {
		errs = append(errs, "description: required field is missing")
	} else if len(s.Description) > 1024 {
		errs = append(errs, "description: exceeds 1024 characters")
	}

	return errs
}
