package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Repo     string            `yaml:"repo"`
	CacheDir string            `yaml:"cache_dir"`
	GitHub   GitHubConfig      `yaml:"github"`
	Targets  map[string]Target `yaml:"targets"`
}

type GitHubConfig struct {
	Method string `yaml:"method"`
	Token  string `yaml:"token,omitempty"`
}

type Target struct {
	Enabled   bool   `yaml:"enabled"`
	SkillPath string `yaml:"skill_path"`
}

// Dir returns the myskills config directory.
func Dir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "myskills")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "myskills")
}

// Path returns the default config file path.
func Path() string {
	return filepath.Join(Dir(), "config.yaml")
}

// Default returns a Config with all known targets disabled and default paths.
func Default() Config {
	home, _ := os.UserHomeDir()
	return Config{
		CacheDir: filepath.Join(Dir(), "repo"),
		GitHub:   GitHubConfig{Method: "gh"},
		Targets: map[string]Target{
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
		},
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
