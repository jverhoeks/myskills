package manifest

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")

	m := Manifest{
		LastSync: time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC),
		Skills: map[string]SkillEntry{
			"deploy": {Commit: "abc123", SyncedAt: time.Date(2026, 4, 3, 10, 0, 0, 0, time.UTC)},
		},
	}

	if err := Save(m, path); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if loaded.Skills["deploy"].Commit != "abc123" {
		t.Errorf("commit: got %q", loaded.Skills["deploy"].Commit)
	}
}

func TestLoadMissing(t *testing.T) {
	m, err := Load("/nonexistent/manifest.json")
	if err != nil {
		t.Fatalf("should not error on missing file: %v", err)
	}
	if m.Skills == nil {
		t.Error("expected initialized skills map")
	}
}

func TestIsOutdated(t *testing.T) {
	m := Manifest{
		Skills: map[string]SkillEntry{
			"deploy": {Commit: "abc123"},
		},
	}

	if m.IsOutdated("deploy", "abc123") {
		t.Error("same commit should not be outdated")
	}
	if !m.IsOutdated("deploy", "def456") {
		t.Error("different commit should be outdated")
	}
	if !m.IsOutdated("new-skill", "abc123") {
		t.Error("unknown skill should be outdated")
	}
}
