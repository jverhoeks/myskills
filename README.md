# 🧠 myskills

Distribute AI agent skills across your team from a central Git repo. Works with **Claude Code**, **GitHub Copilot CLI**, **OpenAI Codex CLI**, and **OpenCode**.

Skills follow the open [Agent Skills](https://agentskills.io) standard — write once, use everywhere. Compatible with [skills.sh](https://skills.sh).

---

## 📚 Skills

Browse the [`skills/`](skills/) directory or run `myskills list` after installing.

| Skill | Description |
|-------|-------------|
| 🧹 [dependency-bloat-reduction](skills/dependency-bloat-reduction/) | Analyze imports, identify trivial/outdated packages, inline or replace with native code |

### Using a skill

Skills are automatically available in your AI coding tool after syncing. Invoke them by name:

```
/dependency-bloat-reduction
```

Or let your AI agent pick them up automatically when the task matches the skill's description.

### ✍️ Contributing a skill via CLI

```bash
# Scaffold a new skill
myskills dev my-new-skill

# Edit the generated SKILL.md
$EDITOR ~/.config/myskills/dev/my-new-skill/SKILL.md

# Validate it
myskills validate ~/.config/myskills/dev/my-new-skill

# Open a PR
myskills submit my-new-skill
```

### ✍️ Adding a skill manually to your repo

You can also add skills directly to your repo without the CLI. Just create a directory under `skills/` with a `SKILL.md` file:

```
skills/
└── my-new-skill/
    ├── SKILL.md           # Required: frontmatter + instructions
    ├── references/        # Optional: extra docs loaded on demand
    └── scripts/           # Optional: scripts the agent can run
```

The `SKILL.md` file needs YAML frontmatter followed by markdown instructions:

```yaml
---
name: my-new-skill
description: What this skill does and when to use it. Be specific so agents
  know when to activate it. At least 50 characters.
metadata:
  team: platform
---

# Instructions for the agent

Step-by-step instructions, examples, rules, templates — whatever helps
the agent do the job well.

## Optional sections

- Reference supporting files: [see details](references/api-docs.md)
- Run scripts: `scripts/helper.sh`
- Include examples, edge cases, constraints
```

**Rules for the frontmatter:**
- `name` — must match the directory name, lowercase with hyphens, max 64 chars
- `description` — what and when, 50–1024 chars (this is what agents see to decide if the skill is relevant)
- `metadata.team` — required by org validation rules

**Optional frontmatter fields** (from the [Agent Skills spec](https://agentskills.io/specification)):
- `license` — e.g., `Apache-2.0`
- `compatibility` — e.g., `Requires Python 3.10+ and Docker`
- `metadata` — any extra key-value pairs

Then commit and push (or open a PR):

```bash
git add skills/my-new-skill/
git commit -m "feat: add my-new-skill"
git push
```

After pushing, anyone on the team runs `myskills sync` to get the new skill.

---

## 🔧 myskills CLI

A Go CLI tool that syncs skills from Git repos into your AI coding tools via symlinks. Single binary, no runtime dependencies.

### ✨ Features

- 🔗 **Symlinks** — skills point to the cached repo, always up to date after `git pull`
- 📦 **Multi-repo** — pull skills from multiple Git repos (org, community, public)
- 🌐 **skills.sh compatible** — `owner/repo` shorthand, same repos work everywhere
- ✅ **Validation** — enforces the [Agent Skills](https://agentskills.io) spec + org-specific rules
- 🎛️ **Per-skill toggle** — interactive TUI to enable/disable individual skills
- 🔍 **Auto-detection** — finds which AI tools you have installed
- 🖥️ **Cross-platform** — macOS, Linux, Windows (amd64 + arm64)

### 📥 Install

**From GitHub Releases (recommended):**

```bash
curl -sSL https://raw.githubusercontent.com/jverhoeks/myskills/main/install.sh | bash
```

**Or download manually** from [Releases](https://github.com/jverhoeks/myskills/releases/latest).

**From source:**

```bash
go install github.com/jverhoeks/myskills/cmd/myskills@latest
```

### 🚀 Quick start

```bash
# 1. Set up with your org's skills repo
myskills init jverhoeks/myskills

# 2. Add public skills from skills.sh
myskills add-skill vercel-labs/agent-skills
myskills add-skill microsoft/skills

# 3. Sync all enabled skills to your AI tools
myskills sync

# 4. See what's installed
myskills list
# REPO           SKILL                        ENABLED  STATUS   SYNCED
# myskills       dependency-bloat-reduction    yes      current  2026-04-04 14:30
# agent-skills   react-best-practices         yes      current  2026-04-04 14:30
# skills         azure-development            yes      current  2026-04-04 14:30

# 5. Toggle individual skills on/off
myskills enable
```

### 📋 Commands

| Command | Description |
|---------|-------------|
| `myskills init <url\|owner/repo> [name]` | 🚀 Set up with a repo, detect tools, write config |
| `myskills add-repo <url\|owner/repo> [name]` | ➕ Add another skill repository |
| `myskills add-skill <owner/repo>` | 🌐 Add skills from [skills.sh](https://skills.sh) / GitHub |
| `myskills sync [skill]` | 🔄 Pull latest and symlink skills to tool directories |
| `myskills list` | 📋 List skills with enabled/synced status |
| `myskills info <name>` | ℹ️ Show skill details and files |
| `myskills search <query>` | 🔎 Search skills by name or description |
| `myskills enable` | 🎛️ Interactive TUI to toggle skills on/off |
| `myskills validate <path>` | ✅ Validate a skill against spec + org rules |
| `myskills dev <name>` | 🆕 Scaffold a new skill |
| `myskills submit <name> [--repo name]` | 📤 Validate and open a PR |
| `myskills remove <name>` | 🗑️ Remove a skill from all tool directories |
| `myskills doctor` | 🩺 Health check: repos, tools, config |
| `myskills config list` | ⚙️ Show current configuration |
| `myskills config set <key> <val>` | ⚙️ Update a config value |

### 🌐 skills.sh compatibility

Add any skill from [skills.sh](https://skills.sh) using `owner/repo` shorthand — the same repos that work with `npx skills add` work here:

```bash
myskills add-skill vercel-labs/agent-skills
myskills add-skill microsoft/skills
myskills add-skill anthropics/courses
myskills sync
```

Skills are auto-discovered in these directories (matching the agentskills.io convention):
- `skills/` · `.agents/skills/` · `.claude/skills/` · `.github/skills/` · `.copilot/skills/` · `.cursor/skills/`
- Or a single `SKILL.md` at the repo root

### 📦 Multi-repo setup

```bash
# Your org's private skills
myskills init git@github.com:my-org/skills.git org

# Add public skill collections
myskills add-skill vercel-labs/agent-skills
myskills add-repo my-org/team-specific-skills

# Sync pulls from all repos
myskills sync
```

### 🎛️ Enable/disable skills

**Skills are disabled by default** — you explicitly choose which ones to activate. This keeps your AI tools lean and avoids unwanted skills from large repos.

```bash
# Search for what you need
myskills search deploy

# Interactive picker to toggle on/off
myskills enable
```

```
  Select skills to enable

> [x] myskills:dependency-bloat-reduction — Analyze imports, identify trivial...
  [ ] agent-skills:react-best-practices — React patterns and best practices
  [ ] agent-skills:web-design-guidelines — Web design system guidelines

  space: toggle  a: toggle all  enter/q: save  esc: cancel
```

After toggling, run `myskills sync` to apply — enabled skills get symlinked, disabled ones get removed.

### 🔌 Supported tools

| Tool | Detected by | Skills synced to |
|------|-------------|-----------------|
| 🟣 Claude Code | `~/.claude/` | `~/.claude/skills/` |
| 🐙 GitHub Copilot CLI | `~/.copilot/` | `~/.copilot/skills/` |
| 🟢 OpenAI Codex CLI | `~/.codex/` | `~/.codex/skills/` |
| 🔵 OpenCode | `~/.config/opencode/` | `~/.config/opencode/skills/` |

### ⚙️ Configuration

| What | macOS / Linux | Windows |
|------|---------------|---------|
| Config | `~/.config/myskills/config.yaml` | `%APPDATA%\myskills\config.yaml` |
| Cache | `~/.cache/myskills/repos/` | `%LOCALAPPDATA%\myskills\cache\repos\` |

Respects `XDG_CONFIG_HOME` and `XDG_CACHE_HOME`.

### 🔐 Authentication

For private repos, use SSH keys or set environment variables — git and `gh` pick them up automatically:

```bash
export GITHUB_TOKEN=ghp_...    # GitHub
export GITLAB_TOKEN=glpat-...  # GitLab
```

No tokens are stored in config files.

---

## 📄 License

MIT
