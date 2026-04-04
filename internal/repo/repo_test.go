package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCloneAndPull(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found")
	}

	remote := t.TempDir()
	run(t, remote, "git", "init", "--bare")

	work := t.TempDir()
	run(t, work, "git", "init")
	run(t, work, "git", "config", "user.email", "test@test.com")
	run(t, work, "git", "config", "user.name", "Test")

	skillDir := filepath.Join(work, "skills", "hello")
	os.MkdirAll(skillDir, 0o755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: hello\n---\n"), 0o644)
	run(t, work, "git", "add", ".")
	run(t, work, "git", "commit", "-m", "init")
	run(t, work, "git", "remote", "add", "origin", remote)
	run(t, work, "git", "push", "origin", "HEAD:main")

	dest := filepath.Join(t.TempDir(), "clone")
	if err := Clone(remote, dest); err != nil {
		t.Fatalf("clone: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, "skills", "hello", "SKILL.md")); err != nil {
		t.Fatal("expected SKILL.md in clone")
	}

	if err := Pull(dest); err != nil {
		t.Fatalf("pull: %v", err)
	}
}

func TestListSkills(t *testing.T) {
	dir := t.TempDir()
	skillsDir := filepath.Join(dir, "skills")
	os.MkdirAll(filepath.Join(skillsDir, "alpha"), 0o755)
	os.MkdirAll(filepath.Join(skillsDir, "beta"), 0o755)
	os.WriteFile(filepath.Join(skillsDir, "alpha", "SKILL.md"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(skillsDir, "beta", "SKILL.md"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(skillsDir, "README.md"), []byte("x"), 0o644)

	skills, err := ListSkills(dir)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}
}

func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}
