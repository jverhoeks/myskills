package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectTools(t *testing.T) {
	home := t.TempDir()

	os.MkdirAll(filepath.Join(home, ".claude"), 0o755)
	os.MkdirAll(filepath.Join(home, ".copilot"), 0o755)

	results := DetectWithHome(home)

	if !results["claude"] {
		t.Error("expected claude detected")
	}
	if !results["copilot"] {
		t.Error("expected copilot detected")
	}
	if results["codex"] {
		t.Error("expected codex not detected")
	}
	if results["opencode"] {
		t.Error("expected opencode not detected")
	}
}

func TestDetectEmpty(t *testing.T) {
	home := t.TempDir()
	results := DetectWithHome(home)

	for name, found := range results {
		if found {
			t.Errorf("expected %s not detected in empty home", name)
		}
	}
}
