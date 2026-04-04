package detect

import (
	"os"
	"path/filepath"
)

// toolDirs maps tool names to the directory path relative to home that indicates installation.
var toolDirs = map[string]string{
	"claude":   ".claude",
	"copilot":  ".copilot",
	"codex":    ".codex",
	"opencode": filepath.Join(".config", "opencode"),
}

// Detect checks which AI tools are installed by looking for config directories in the user's home.
func Detect() map[string]bool {
	home, _ := os.UserHomeDir()
	return DetectWithHome(home)
}

// DetectWithHome checks for tool directories under the given home path.
func DetectWithHome(home string) map[string]bool {
	results := make(map[string]bool, len(toolDirs))
	for name, rel := range toolDirs {
		dir := filepath.Join(home, rel)
		info, err := os.Stat(dir)
		results[name] = err == nil && info.IsDir()
	}
	return results
}
