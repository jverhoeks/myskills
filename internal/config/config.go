package config

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Repos   []Repo            `yaml:"repos"`
	GitHub  GitHubConfig      `yaml:"github"`
	Targets map[string]Target `yaml:"targets"`
	Skills  SkillsConfig      `yaml:"skills"`
}

type Repo struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

type GitHubConfig struct {
	Method string `yaml:"method"`
	Token  string `yaml:"token,omitempty"`
}

type Target struct {
	Enabled   bool   `yaml:"enabled"`
	SkillPath string `yaml:"skill_path"`
}

// SkillsConfig tracks which skills are enabled/disabled.
// Disabled entries use "repo:skill" format for multi-repo disambiguation,
// or just "skill" if unambiguous.
type SkillsConfig struct {
	Disabled []string `yaml:"disabled,omitempty"`
}

// Dir returns the myskills config directory.
func Dir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "myskills")
	}
	home, _ := os.UserHomeDir()
	if runtime.GOOS == "windows" {
		if appdata := os.Getenv("APPDATA"); appdata != "" {
			return filepath.Join(appdata, "myskills")
		}
	}
	return filepath.Join(home, ".config", "myskills")
}

// CacheDir returns the OS-appropriate cache directory.
// macOS/Linux: ~/.cache/myskills/
// Windows: %LOCALAPPDATA%\myskills\cache
func CacheDir() string {
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, "myskills")
	}
	if runtime.GOOS == "windows" {
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, "myskills", "cache")
		}
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "myskills")
}

// RepoDir returns the cache directory for a specific repo.
// Uses the repo name as subdirectory under CacheDir()/repos/.
func RepoDir(repoName string) string {
	return filepath.Join(CacheDir(), "repos", repoName)
}

// RepoHash returns a short hash for a repo URL (fallback if no name given).
func RepoHash(url string) string {
	h := sha256.Sum256([]byte(url))
	return fmt.Sprintf("%x", h[:8])
}

// Path returns the default config file path.
func Path() string {
	return filepath.Join(Dir(), "config.yaml")
}

// Default returns a Config with all known targets disabled and default paths.
func Default() Config {
	home, _ := os.UserHomeDir()
	targets := map[string]Target{
		"claude": {
			SkillPath: filepath.Join(home, ".claude", "skills"),
		},
		"copilot": {
			SkillPath: filepath.Join(home, ".copilot", "skills"),
		},
		"codex": {
			SkillPath: filepath.Join(home, ".codex", "skills"),
		},
		"opencode": {
			SkillPath: filepath.Join(home, ".config", "opencode", "skills"),
		},
	}
	return Config{
		GitHub:  GitHubConfig{Method: "gh"},
		Targets: targets,
	}
}

// IsSkillDisabled returns true if the skill is explicitly disabled.
// Checks both "skill" and "repo:skill" formats.
func (c Config) IsSkillDisabled(name string) bool {
	for _, d := range c.Skills.Disabled {
		if d == name {
			return true
		}
		// Also match "repo:skill" format
		if parts := strings.SplitN(d, ":", 2); len(parts) == 2 && parts[1] == name {
			return true
		}
	}
	return false
}

// IsSkillDisabledInRepo returns true if the skill from a specific repo is disabled.
func (c Config) IsSkillDisabledInRepo(repoName, skillName string) bool {
	qualified := repoName + ":" + skillName
	for _, d := range c.Skills.Disabled {
		if d == qualified || d == skillName {
			return true
		}
	}
	return false
}

// SetSkillDisabled adds or removes a skill from the disabled list.
func (c *Config) SetSkillDisabled(name string, disabled bool) {
	if disabled {
		if !c.IsSkillDisabled(name) {
			c.Skills.Disabled = append(c.Skills.Disabled, name)
		}
	} else {
		var kept []string
		for _, d := range c.Skills.Disabled {
			if d != name {
				// Also remove "repo:skill" entries that match
				if parts := strings.SplitN(d, ":", 2); len(parts) == 2 && parts[1] == name {
					continue
				}
				kept = append(kept, d)
			}
		}
		c.Skills.Disabled = kept
	}
}

// Load reads a config from a YAML file.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("reading config: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}
	return cfg, nil
}

// Save writes a config to a YAML file, creating parent directories.
func Save(cfg Config, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// ExpandPath replaces a leading ~ with the user's home directory.
func ExpandPath(p string) string {
	if strings.HasPrefix(p, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, p[2:])
	}
	return p
}
