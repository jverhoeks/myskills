package repo

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestResolveURL(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{"vercel-labs/agent-skills", "https://github.com/vercel-labs/agent-skills.git"},
		{"jverhoeks/myskills", "https://github.com/jverhoeks/myskills.git"},
		{"https://github.com/foo/bar.git", "https://github.com/foo/bar.git"},
		{"git@github.com:foo/bar.git", "git@github.com:foo/bar.git"},
	}
	for _, tc := range cases {
		got := ResolveURL(tc.input)
		if got != tc.want {
			t.Errorf("ResolveURL(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestNameFromURL(t *testing.T) {
	cases := []struct {
		input, want string
	}{
		{"vercel-labs/agent-skills", "agent-skills"},
		{"https://github.com/foo/bar.git", "bar"},
		{"git@github.com:org/my-skills.git", "my-skills"},
	}
	for _, tc := range cases {
		got := NameFromURL(tc.input)
		if got != tc.want {
			t.Errorf("NameFromURL(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestCloneAndPull(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found")
	}

	remote := t.TempDir()
	run(t, remote, "git", "init", "--bare", "--initial-branch=main")

	work := t.TempDir()
	run(t, work, "git", "init", "--initial-branch=main")
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

func TestListSkillsStandardDir(t *testing.T) {
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

func TestListSkillsAgentsDir(t *testing.T) {
	dir := t.TempDir()
	// Skills in .agents/skills/ (skills.sh pattern)
	agentsDir := filepath.Join(dir, ".agents", "skills")
	os.MkdirAll(filepath.Join(agentsDir, "deploy"), 0o755)
	os.WriteFile(filepath.Join(agentsDir, "deploy", "SKILL.md"), []byte("x"), 0o644)

	skills, err := ListSkills(dir)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(skills) != 1 || skills[0] != "deploy" {
		t.Fatalf("expected [deploy], got %v", skills)
	}
}

func TestListSkillsClaudeDir(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude", "skills")
	os.MkdirAll(filepath.Join(claudeDir, "review"), 0o755)
	os.WriteFile(filepath.Join(claudeDir, "review", "SKILL.md"), []byte("x"), 0o644)

	skills, err := ListSkills(dir)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(skills) != 1 || skills[0] != "review" {
		t.Fatalf("expected [review], got %v", skills)
	}
}

func TestListSkillsRootSkill(t *testing.T) {
	dir := t.TempDir()
	// SKILL.md at root — repo is a single skill
	os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: solo\n---\n"), 0o644)

	skills, err := ListSkills(dir)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
}

func TestFindSkillsDir(t *testing.T) {
	// Test priority: skills/ wins over .agents/skills/
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "skills", "a"), 0o755)
	os.WriteFile(filepath.Join(dir, "skills", "a", "SKILL.md"), []byte("x"), 0o644)
	os.MkdirAll(filepath.Join(dir, ".agents", "skills", "b"), 0o755)
	os.WriteFile(filepath.Join(dir, ".agents", "skills", "b", "SKILL.md"), []byte("x"), 0o644)

	got := FindSkillsDir(dir)
	want := filepath.Join(dir, "skills")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestSkillDirRootSkill(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "solo-skill")
	os.MkdirAll(dir, 0o755)
	os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("x"), 0o644)

	got := SkillDir(dir, "solo-skill")
	if got != dir {
		t.Errorf("got %q, want %q", got, dir)
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
