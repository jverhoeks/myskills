package skill

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Skill represents a parsed SKILL.md file.
type Skill struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description"`
	License     string            `yaml:"license"`
	Metadata    map[string]string `yaml:"metadata"`
	Body        string
}

// Parse reads and parses a SKILL.md file.
func Parse(path string) (Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Skill{}, fmt.Errorf("reading %s: %w", path, err)
	}

	content := string(data)

	if !strings.HasPrefix(content, "---\n") {
		return Skill{}, fmt.Errorf("%s: missing YAML frontmatter (must start with ---)", path)
	}

	rest := content[4:]
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return Skill{}, fmt.Errorf("%s: missing closing --- for frontmatter", path)
	}

	frontmatter := rest[:idx]
	body := strings.TrimSpace(rest[idx+4:])

	var s Skill
	if err := yaml.Unmarshal([]byte(frontmatter), &s); err != nil {
		return Skill{}, fmt.Errorf("%s: invalid frontmatter YAML: %w", path, err)
	}
	s.Body = body

	return s, nil
}
