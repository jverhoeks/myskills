package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := Default()

	if cfg.Repo != "" {
		t.Errorf("expected empty repo, got %q", cfg.Repo)
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
	if claude.SkillPath == "" {
		t.Error("expected default skill path for claude")
	}
}

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg := Default()
	cfg.Repo = "git@github.com:sbp/skills.git"
	cfg.Targets["claude"] = Target{Enabled: true, SkillPath: "~/.claude/skills"}

	if err := Save(cfg, path); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if loaded.Repo != cfg.Repo {
		t.Errorf("repo: got %q, want %q", loaded.Repo, cfg.Repo)
	}
	if !loaded.Targets["claude"].Enabled {
		t.Error("expected claude enabled after load")
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
