package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	if len(cfg.Repos) != 0 {
		t.Errorf("expected no repos, got %d", len(cfg.Repos))
	}
	if cfg.GitHub.Method != "gh" {
		t.Errorf("expected github method 'gh', got %q", cfg.GitHub.Method)
	}
	if len(cfg.Targets) != 4 {
		t.Errorf("expected 4 targets, got %d", len(cfg.Targets))
	}
	if len(cfg.Skills.Enabled) != 0 {
		t.Errorf("expected no skills enabled by default, got %d", len(cfg.Skills.Enabled))
	}
}

func TestCacheDir(t *testing.T) {
	dir := CacheDir()
	if dir == "" {
		t.Error("expected non-empty cache dir")
	}
	if !strings.Contains(dir, "myskills") {
		t.Errorf("expected cache dir to contain 'myskills', got %q", dir)
	}
}

func TestCacheDirXDG(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", tmp)
	dir := CacheDir()
	want := filepath.Join(tmp, "myskills")
	if dir != want {
		t.Errorf("got %q, want %q", dir, want)
	}
}

func TestRepoDir(t *testing.T) {
	dir := RepoDir("org")
	if !strings.HasSuffix(dir, filepath.Join("repos", "org")) {
		t.Errorf("expected repo dir to end with repos/org, got %q", dir)
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := Default()
	cfg.Repos = []Repo{
		{Name: "org", URL: "git@github.com:sbp/skills.git"},
	}
	cfg.Targets["claude"] = Target{Enabled: true, SkillPath: "~/.claude/skills"}
	cfg.Skills.Enabled = []string{"org:deploy", "org:review"}

	if err := Save(cfg, path); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if len(loaded.Repos) != 1 {
		t.Errorf("repos: got %d, want 1", len(loaded.Repos))
	}
	if !loaded.Targets["claude"].Enabled {
		t.Error("expected claude enabled after load")
	}
	if len(loaded.Skills.Enabled) != 2 {
		t.Errorf("enabled skills: got %v", loaded.Skills.Enabled)
	}
}

func TestLoadMissing(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("expected error loading missing config")
	}
}

func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()
	got := ExpandPath("~/.claude/skills")
	want := filepath.Join(home, ".claude", "skills")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestConfigDir(t *testing.T) {
	dir := Dir()
	if dir == "" {
		t.Error("expected non-empty config dir")
	}
}

func TestIsSkillEnabled(t *testing.T) {
	cfg := Config{Skills: SkillsConfig{Enabled: []string{"org:deploy", "review"}}}

	if !cfg.IsSkillEnabled("deploy") {
		t.Error("expected deploy enabled (via org:deploy)")
	}
	if !cfg.IsSkillEnabled("review") {
		t.Error("expected review enabled (bare name)")
	}
	if cfg.IsSkillEnabled("unknown") {
		t.Error("expected unknown not enabled")
	}
}

func TestIsSkillEnabledInRepo(t *testing.T) {
	cfg := Config{Skills: SkillsConfig{Enabled: []string{"org:deploy", "review"}}}

	if !cfg.IsSkillEnabledInRepo("org", "deploy") {
		t.Error("expected org:deploy enabled")
	}
	if !cfg.IsSkillEnabledInRepo("any", "review") {
		t.Error("expected review enabled globally")
	}
	if cfg.IsSkillEnabledInRepo("org", "unknown") {
		t.Error("expected unknown not enabled")
	}
}

func TestSetSkillEnabled(t *testing.T) {
	cfg := Config{}

	cfg.SetSkillEnabled("org:deploy", true)
	if !cfg.IsSkillEnabled("deploy") {
		t.Error("expected deploy enabled after set")
	}

	// Duplicate should not add twice
	cfg.SetSkillEnabled("org:deploy", true)
	count := 0
	for _, e := range cfg.Skills.Enabled {
		if e == "org:deploy" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 entry, got %d", count)
	}

	cfg.SetSkillEnabled("org:deploy", false)
	if cfg.IsSkillEnabled("org:deploy") {
		t.Error("expected disabled after unset")
	}
}

func TestEnableSkills(t *testing.T) {
	cfg := Config{}
	cfg.EnableSkills("org", []string{"deploy", "review", "test"})

	if len(cfg.Skills.Enabled) != 3 {
		t.Errorf("expected 3 enabled, got %d", len(cfg.Skills.Enabled))
	}
	if !cfg.IsSkillEnabledInRepo("org", "deploy") {
		t.Error("expected org:deploy enabled")
	}
	if !cfg.IsSkillEnabledInRepo("org", "review") {
		t.Error("expected org:review enabled")
	}
}
