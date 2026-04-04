package manifest

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Manifest struct {
	LastSync time.Time             `json:"last_sync"`
	Skills   map[string]SkillEntry `json:"skills"`
}

type SkillEntry struct {
	Commit   string    `json:"commit"`
	SyncedAt time.Time `json:"synced_at"`
}

// Load reads a manifest from a JSON file. Returns an empty manifest if the file doesn't exist.
func Load(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Manifest{Skills: make(map[string]SkillEntry)}, nil
		}
		return Manifest{}, fmt.Errorf("reading manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("parsing manifest: %w", err)
	}
	if m.Skills == nil {
		m.Skills = make(map[string]SkillEntry)
	}
	return m, nil
}

// Save writes a manifest to a JSON file.
func Save(m Manifest, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating manifest dir: %w", err)
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling manifest: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// IsOutdated returns true if the skill's commit differs from the manifest entry,
// or if the skill is not in the manifest at all.
func (m Manifest) IsOutdated(name, currentCommit string) bool {
	entry, ok := m.Skills[name]
	if !ok {
		return true
	}
	return entry.Commit != currentCommit
}
