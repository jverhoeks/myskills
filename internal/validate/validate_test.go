package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeSkill(t *testing.T, dir, name, content string) string {
	t.Helper()
	skillDir := filepath.Join(dir, name)
	os.MkdirAll(skillDir, 0o755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644)
	return skillDir
}

func TestSpecValid(t *testing.T) {
	dir := t.TempDir()
	path := writeSkill(t, dir, "my-skill", `---
name: my-skill
description: A valid skill that does something useful for the team.
---

Instructions here.
`)

	errs := Spec(path)
	if len(errs) > 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestSpecMissingSKILLmd(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "bad-skill"), 0o755)

	errs := Spec(filepath.Join(dir, "bad-skill"))
	if len(errs) == 0 {
		t.Error("expected error for missing SKILL.md")
	}
}

func TestSpecNameMismatch(t *testing.T) {
	dir := t.TempDir()
	path := writeSkill(t, dir, "my-skill", `---
name: different-name
description: A valid description that is long enough.
---

Body.
`)

	errs := Spec(path)
	assertContains(t, errs, "name")
}

func TestSpecInvalidName(t *testing.T) {
	dir := t.TempDir()

	cases := []struct {
		dirName string
		name    string
	}{
		{"UPPER", "UPPER"},
		{"-leading", "-leading"},
		{"trailing-", "trailing-"},
		{"double--hyphen", "double--hyphen"},
	}

	for _, tc := range cases {
		path := writeSkill(t, dir, tc.dirName, "---\nname: "+tc.name+"\ndescription: A valid description for testing.\n---\nBody.\n")
		errs := Spec(path)
		if len(errs) == 0 {
			t.Errorf("expected error for name %q", tc.name)
		}
	}
}

func TestSpecMissingDescription(t *testing.T) {
	dir := t.TempDir()
	path := writeSkill(t, dir, "no-desc", `---
name: no-desc
---

Body.
`)

	errs := Spec(path)
	assertContains(t, errs, "description")
}

// --- Org rules tests ---

func TestOrgRulesValid(t *testing.T) {
	dir := t.TempDir()
	path := writeSkill(t, dir, "my-skill", `---
name: my-skill
description: A valid skill that does something useful for the team and is long enough.
metadata:
  team: infra
---

Instructions here.
`)

	rules := OrgRules{
		DescriptionMinLength: 20,
		RequiredMetadata:     []string{"team"},
		AllowedTeams:         []string{"infra", "platform"},
		MaxSkillMDLines:      500,
	}

	errs := Org(path, rules)
	if len(errs) > 0 {
		t.Errorf("expected no errors, got: %v", errs)
	}
}

func TestOrgRulesShortDescription(t *testing.T) {
	dir := t.TempDir()
	path := writeSkill(t, dir, "short", `---
name: short
description: Too short
metadata:
  team: infra
---

Body.
`)

	rules := OrgRules{DescriptionMinLength: 50}
	errs := Org(path, rules)
	assertContains(t, errs, "description")
}

func TestOrgRulesMissingMetadata(t *testing.T) {
	dir := t.TempDir()
	path := writeSkill(t, dir, "no-team", `---
name: no-team
description: A valid description that is long enough for the org rules.
---

Body.
`)

	rules := OrgRules{RequiredMetadata: []string{"team"}}
	errs := Org(path, rules)
	assertContains(t, errs, "team")
}

func TestOrgRulesInvalidTeam(t *testing.T) {
	dir := t.TempDir()
	path := writeSkill(t, dir, "bad-team", `---
name: bad-team
description: A valid description that is long enough for the org rules.
metadata:
  team: unknown
---

Body.
`)

	rules := OrgRules{
		RequiredMetadata: []string{"team"},
		AllowedTeams:     []string{"infra", "platform"},
	}
	errs := Org(path, rules)
	assertContains(t, errs, "team")
}

func TestOrgRulesTooManyLines(t *testing.T) {
	dir := t.TempDir()
	longBody := strings.Repeat("line\n", 600)
	path := writeSkill(t, dir, "long", "---\nname: long\ndescription: A valid description that is long enough.\nmetadata:\n  team: infra\n---\n"+longBody)

	rules := OrgRules{MaxSkillMDLines: 500, AllowedTeams: []string{"infra"}}
	errs := Org(path, rules)
	assertContains(t, errs, "lines")
}

func TestOrgRulesNamePrefix(t *testing.T) {
	dir := t.TempDir()
	path := writeSkill(t, dir, "my-skill", `---
name: my-skill
description: A valid description that is long enough for the org rules.
metadata:
  team: infra
---

Body.
`)

	rules := OrgRules{NamePrefix: "sbp-", AllowedTeams: []string{"infra"}}
	errs := Org(path, rules)
	assertContains(t, errs, "prefix")
}

func TestLoadOrgRules(t *testing.T) {
	dir := t.TempDir()
	content := `org: sbp
validation:
  description_min_length: 50
  required_metadata:
    - team
  allowed_teams:
    - infra
    - platform
  name_prefix: ""
  max_skill_md_lines: 500
`
	os.WriteFile(filepath.Join(dir, ".myskills.yaml"), []byte(content), 0o644)

	rules, err := LoadOrgRules(filepath.Join(dir, ".myskills.yaml"))
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if rules.DescriptionMinLength != 50 {
		t.Errorf("description_min_length: got %d", rules.DescriptionMinLength)
	}
	if len(rules.AllowedTeams) != 2 {
		t.Errorf("allowed_teams: got %d", len(rules.AllowedTeams))
	}
}

func assertContains(t *testing.T, errs []string, substr string) {
	t.Helper()
	for _, e := range errs {
		if strings.Contains(e, substr) {
			return
		}
	}
	t.Errorf("expected an error containing %q, got: %v", substr, errs)
}
