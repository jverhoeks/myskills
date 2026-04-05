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
	Method string `yaml:"method"` // "gh" (default) or "token" (reads GITHUB_TOKEN env)
}

type Target struct {
	Enabled   bool   `yaml:"enabled"`
	SkillPath string `yaml:"skill_path"`
}

// SkillsConfig tracks which skills are enabled.
// Skills are disabled by default — only entries in this list get synced.
// Entries use "repo:skill" format (e.g., "myskills:deploy").
type SkillsConfig struct {
	Enabled []string `yaml:"enabled,omitempty"`
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

// IsSkillEnabled returns true if the skill is in the enabled list.
// Checks for "repo:skill" qualified entries and bare "skill" entries.
func (c Config) IsSkillEnabled(name string) bool {
	for _, e := range c.Skills.Enabled {
		if e == name {
			return true
		}
		// Match "repo:skill" → "skill"
		if parts := strings.SplitN(e, ":", 2); len(parts) == 2 && parts[1] == name {
			return true
		}
	}
	return false
}

// IsSkillEnabledInRepo returns true if the skill from a specific repo is enabled.
func (c Config) IsSkillEnabledInRepo(repoName, skillName string) bool {
	qualified := repoName + ":" + skillName
	for _, e := range c.Skills.Enabled {
		if e == qualified || e == skillName {
			return true
		}
	}
	return false
}

// SetSkillEnabled adds or removes a skill from the enabled list.
func (c *Config) SetSkillEnabled(name string, enabled bool) {
	if enabled {
		if !c.IsSkillEnabled(name) {
			c.Skills.Enabled = append(c.Skills.Enabled, name)
		}
	} else {
		var kept []string
		for _, e := range c.Skills.Enabled {
			if e != name {
				// Also remove "repo:skill" entries that match
				if parts := strings.SplitN(e, ":", 2); len(parts) == 2 && parts[1] == name {
					continue
				}
				kept = append(kept, e)
			}
		}
		c.Skills.Enabled = kept
	}
}

// EnableSkills adds multiple qualified "repo:skill" entries to the enabled list.
func (c *Config) EnableSkills(repoName string, skillNames []string) {
	for _, name := range skillNames {
		qualified := repoName + ":" + name
		c.SetSkillEnabled(qualified, true)
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
