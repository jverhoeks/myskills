package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sbp/myskills/internal/skill"
	"gopkg.in/yaml.v3"
)

// OrgRules defines organization-specific validation rules.
type OrgRules struct {
	DescriptionMinLength int      `yaml:"description_min_length"`
	RequiredMetadata     []string `yaml:"required_metadata"`
	AllowedTeams         []string `yaml:"allowed_teams"`
	NamePrefix           string   `yaml:"name_prefix"`
	MaxSkillMDLines      int      `yaml:"max_skill_md_lines"`
}

type orgRulesFile struct {
	Org        string   `yaml:"org"`
	Validation OrgRules `yaml:"validation"`
}

// LoadOrgRules reads org validation rules from a .myskills.yaml file.
func LoadOrgRules(path string) (OrgRules, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return OrgRules{}, fmt.Errorf("reading org rules: %w", err)
	}
	var f orgRulesFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return OrgRules{}, fmt.Errorf("parsing org rules: %w", err)
	}
	return f.Validation, nil
}

// Org validates a skill directory against organization rules.
// Returns a list of error descriptions (empty means valid).
func Org(skillDir string, rules OrgRules) []string {
	var errs []string

	skillFile := skillDir
	if info, err := os.Stat(skillDir); err == nil && info.IsDir() {
		skillFile = filepath.Join(skillDir, "SKILL.md")
	}

	s, err := skill.Parse(skillFile)
	if err != nil {
		return []string{err.Error()}
	}

	if rules.DescriptionMinLength > 0 && len(s.Description) < rules.DescriptionMinLength {
		errs = append(errs, fmt.Sprintf("description: must be at least %d characters (got %d)", rules.DescriptionMinLength, len(s.Description)))
	}

	for _, key := range rules.RequiredMetadata {
		if s.Metadata == nil || s.Metadata[key] == "" {
			errs = append(errs, fmt.Sprintf("metadata.%s: required by org rules", key))
		}
	}

	if len(rules.AllowedTeams) > 0 && s.Metadata != nil && s.Metadata["team"] != "" {
		team := s.Metadata["team"]
		found := false
		for _, t := range rules.AllowedTeams {
			if t == team {
				found = true
				break
			}
		}
		if !found {
			errs = append(errs, fmt.Sprintf("metadata.team: %q is not in allowed teams %v", team, rules.AllowedTeams))
		}
	}

	if rules.NamePrefix != "" && !strings.HasPrefix(s.Name, rules.NamePrefix) {
		errs = append(errs, fmt.Sprintf("name: must start with prefix %q", rules.NamePrefix))
	}

	if rules.MaxSkillMDLines > 0 {
		data, _ := os.ReadFile(skillFile)
		lines := strings.Count(string(data), "\n") + 1
		if lines > rules.MaxSkillMDLines {
			errs = append(errs, fmt.Sprintf("SKILL.md: %d lines exceeds max %d lines", lines, rules.MaxSkillMDLines))
		}
	}

	return errs
}
