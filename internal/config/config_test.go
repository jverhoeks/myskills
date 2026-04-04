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
	claude, ok := cfg.Targets["claude"]
	if !ok {
		t.Fatal("expected 'claude' target")
	}
	if claude.Enabled {
		t.Error("expected claude disabled by default")
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
		{Name: "community", URL: "git@github.com:sbp/community-skills.git"},
	}
	cfg.Targets["claude"] = Target{Enabled: true, SkillPath: "~/.claude/skills"}
	cfg.Skills.Disabled = []string{"experimental"}

	if err := Save(cfg, path); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if len(loaded.Repos) != 2 {
		t.Errorf("repos: got %d, want 2", len(loaded.Repos))
	}
	if loaded.Repos[0].Name != "org" {
		t.Errorf("repo name: got %q", loaded.Repos[0].Name)
	}
	if !loaded.Targets["claude"].Enabled {
		t.Error("expected claude enabled after load")
	}
	if len(loaded.Skills.Disabled) != 1 || loaded.Skills.Disabled[0] != "experimental" {
		t.Errorf("disabled skills: got %v", loaded.Skills.Disabled)
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

	got2 := ExpandPath("/absolute/path")
	if got2 != "/absolute/path" {
		t.Errorf("got %q, want %q", got2, "/absolute/path")
	}
}

func TestConfigDir(t *testing.T) {
	dir := Dir()
	if dir == "" {
		t.Error("expected non-empty config dir")
	}
}

func TestIsSkillDisabled(t *testing.T) {
	cfg := Config{Skills: SkillsConfig{Disabled: []string{"alpha", "org:beta"}}}

	if !cfg.IsSkillDisabled("alpha") {
		t.Error("expected alpha disabled")
	}
	if !cfg.IsSkillDisabled("beta") {
		t.Error("expected beta disabled (via org:beta)")
	}
	if cfg.IsSkillDisabled("gamma") {
		t.Error("expected gamma not disabled")
	}
}

func TestIsSkillDisabledInRepo(t *testing.T) {
	cfg := Config{Skills: SkillsConfig{Disabled: []string{"org:beta", "gamma"}}}

	if !cfg.IsSkillDisabledInRepo("org", "beta") {
		t.Error("expected org:beta disabled")
	}
	if !cfg.IsSkillDisabledInRepo("community", "gamma") {
		t.Error("expected gamma disabled globally")
	}
	if cfg.IsSkillDisabledInRepo("org", "alpha") {
		t.Error("expected alpha not disabled")
	}
}

func TestSetSkillDisabled(t *testing.T) {
	cfg := Config{}

	cfg.SetSkillDisabled("alpha", true)
	if !cfg.IsSkillDisabled("alpha") {
		t.Error("expected alpha disabled after set")
	}

	// Duplicate should not add twice
	cfg.SetSkillDisabled("alpha", true)
	count := 0
	for _, d := range cfg.Skills.Disabled {
		if d == "alpha" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 entry for alpha, got %d", count)
	}

	cfg.SetSkillDisabled("alpha", false)
	if cfg.IsSkillDisabled("alpha") {
		t.Error("expected alpha enabled after unset")
	}
}
