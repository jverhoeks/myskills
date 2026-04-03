# myskills — Org-Wide AI Agent Skill Distribution Tool

## Overview

A Go CLI tool (`myskills`) that distributes AI agent skills from a central private GitHub repository to developer machines. Skills follow the [Agent Skills](https://agentskills.io/specification) open standard and are copied into tool-specific directories for Claude Code, Copilot CLI, Codex CLI, and OpenCode.

## Goals

- **Standardize** coding conventions, review patterns, and deployment workflows across the org
- **Distribute** domain-specific knowledge and playbooks to all developers
- **Accelerate** common tasks with org-specific automation
- Start with skills, extensible later to full `.claude/` config (hooks, subagents, CLAUDE.md snippets)

## Non-Goals

- Not a public registry or marketplace
- Not a package manager (no dependency resolution, no lock files, no semver)
- No automatic updates — user controls when to sync

## Central Repository Structure

Private GitHub repo, org-wide read access. All changes go through PRs.

```
sbp-skills/                          # or whatever the org repo is named
├── skills/
│   ├── deploy/
│   │   ├── SKILL.md
│   │   └── scripts/deploy.sh
│   ├── code-review/
│   │   ├── SKILL.md
│   │   └── references/checklist.md
│   ├── api-conventions/
│   │   └── SKILL.md
│   └── new-service/
│       ├── SKILL.md
│       └── assets/template/
├── config/                          # Future: shared CLAUDE.md snippets, hooks, subagents
│   └── .gitkeep
├── .myskills.yaml                   # Repo-level metadata and validation rules
├── .github/
│   └── workflows/
│       └── validate.yaml            # CI validation on PRs
└── README.md
```

- Flat `skills/` directory, no nesting by team/category
- Team ownership expressed via `metadata.team` in frontmatter
- Skill directory name must match the `name` field in SKILL.md

## CLI Commands

Binary: `myskills`. Config: `~/.config/myskills/config.yaml`.

```
myskills init                        # Interactive setup: repo URL, detect tools, write config
myskills sync                        # Pull latest, copy all skills to enabled tool directories
myskills sync <name>                 # Sync a single skill
myskills list                        # List installed skills with status (current / outdated / not installed)
myskills info <name>                 # Show skill details (description, metadata, files)
myskills validate <path>             # Validate a skill against agentskills.io spec + org rules
myskills dev <name>                  # Scaffold a new skill in local dev workspace
myskills submit <name>              # Validate + create branch + open PR (via gh CLI or token)
myskills remove <name>               # Remove a skill from all enabled tool directories
myskills doctor                      # Health check: repo reachable, tools detected, config valid
myskills config set <key> <value>    # Set a config value
myskills config list                 # Show current config
```

## Init Flow

Interactive wizard that:

1. Prompts for the central repo URL
2. Clones the repo to `~/.config/myskills/repo/`
3. Detects installed AI tools by checking for config directories
4. Asks which tools to enable
5. Writes `~/.config/myskills/config.yaml`

## Config File

```yaml
repo: git@github.com:sbp/skills.git
cache_dir: ~/.config/myskills/repo

github:
  method: gh          # "gh" (tries gh CLI) or "token"
  token: ""           # Personal access token, used when method is "token"

targets:
  claude:
    enabled: true
    skill_path: ~/.claude/skills
    # Future: config_path: ~/.claude/
  copilot:
    enabled: true
    skill_path: ~/.copilot/skills
  codex:
    enabled: false
    skill_path: ~/.codex/skills
  opencode:
    enabled: false
    skill_path: ~/.config/opencode/skills
```

- Target definitions (names + default paths) are built into the binary
- Init toggles `enabled` and detects defaults
- Custom paths override defaults for non-standard installs
- Config is plain YAML, editable by hand or via `myskills config set`

## Sync Mechanism

Copy-based. On `myskills sync`:

1. Run `git pull` on the cached clone at `~/.config/myskills/repo/`
2. For each skill in `skills/`:
   - For each enabled target: copy the skill directory to `<target.skill_path>/<skill-name>/`
   - Overwrites existing files, removes files no longer in source
3. Print summary: `12 skills synced to claude, copilot`

On `myskills sync <name>`: same, but only for the named skill.

On `myskills remove <name>`: delete `<target.skill_path>/<name>/` from all enabled targets.

## Staleness Detection

`myskills list` compares the git commit hash of each skill directory in the cached clone against a manifest stored at `~/.config/myskills/manifest.json` (written after each sync):

```json
{
  "last_sync": "2026-04-03T10:00:00Z",
  "skills": {
    "deploy": { "commit": "abc123", "synced_at": "2026-04-03T10:00:00Z" },
    "code-review": { "commit": "def456", "synced_at": "2026-04-03T10:00:00Z" }
  }
}
```

After `git pull`, compare current commits to manifest. Show: `current`, `outdated`, or `not installed`.

## Validation

Two layers, both run by `myskills validate`:

### Layer 1 — Agent Skills spec

- `SKILL.md` exists
- Valid YAML frontmatter with `name` and `description`
- `name` matches directory name
- `name` follows spec (lowercase, hyphens, 1-64 chars, no consecutive hyphens)
- `description` is 1-1024 chars

### Layer 2 — Org rules (from `.myskills.yaml`)

```yaml
org: sbp
validation:
  description_min_length: 50
  required_metadata:
    - team
  allowed_teams:
    - platform
    - infra
    - frontend
    - data
  name_prefix: ""                    # Optional: enforce prefix on all skills
  max_skill_md_lines: 500
```

### When validation runs

- `myskills validate <path>` — explicitly
- `myskills submit <name>` — automatically before PR creation
- CI on the central repo — GitHub Action on PRs touching `skills/`

## Dev & Submit Workflow

### Scaffold

`myskills dev <name>` creates:

```
~/.config/myskills/dev/<name>/
├── SKILL.md          # Template with required frontmatter
└── references/
    └── .gitkeep
```

Template:
```yaml
---
name: <name>
description: <describe what this skill does and when to use it>
metadata:
  team: <your-team>
---

# Instructions

<your skill content here>
```

### Submit

`myskills submit <name>`:

1. Validate the skill
2. Copy to the cached repo clone at `skills/<name>/`
3. Create branch `skill/<name>`
4. Commit changes
5. Push and open PR (via `gh` CLI first, falls back to token, otherwise prints manual instructions)

## CI Validation

GitHub Action on the central repo:

```yaml
on:
  pull_request:
    paths: ['skills/**']

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install myskills
        run: curl -sSL .../install.sh | bash
      - name: Validate changed skills
        run: |
          git diff --name-only origin/main... \
            | grep '^skills/' \
            | cut -d/ -f1-2 \
            | sort -u \
            | while read dir; do myskills validate "$dir"; done
```

## CLI Distribution

- Go binary built with GoReleaser
- Published to GitHub Releases
- Install script for `darwin/arm64`, `darwin/amd64`, `linux/amd64`
- `curl -sSL <repo>/install.sh | bash`

## Tool Detection Heuristics

Built into the binary:

| Tool | Detect by | Default skill_path |
|------|-----------|-------------------|
| Claude Code | `~/.claude/` exists | `~/.claude/skills` |
| Copilot CLI | `~/.copilot/` exists | `~/.copilot/skills` |
| Codex CLI | `~/.codex/` exists | `~/.codex/skills` |
| OpenCode | `~/.config/opencode/` exists | `~/.config/opencode/skills` |

## Future Extensions

Planned but not in v1:

- **Config sync**: CLAUDE.md snippets, hooks, subagent definitions via `config/` directory
- **Per-target skill filtering**: some skills only relevant for certain tools
- **Skill categories/tags**: filtering `myskills list --team infra`
- **Update notifications**: opt-in staleness warnings on shell startup

## Project Structure (Go)

```
myskills/
├── cmd/
│   └── myskills/
│       └── main.go                  # CLI entrypoint (cobra)
├── internal/
│   ├── config/                      # Config loading, writing, defaults
│   ├── repo/                        # Git operations (clone, pull)
│   ├── sync/                        # Copy skills to targets
│   ├── validate/                    # Spec + org rule validation
│   ├── dev/                         # Scaffold and submit workflows
│   ├── detect/                      # Tool detection
│   └── manifest/                    # Manifest read/write for staleness
├── .goreleaser.yaml
├── go.mod
├── go.sum
└── README.md
```

Dependencies: `cobra` for CLI, `gopkg.in/yaml.v3` for YAML parsing, standard library for the rest. Shell out to `git` and `gh` via `os/exec`.
