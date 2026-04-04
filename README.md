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

### Contributing a skill

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

Every skill needs:
- A `name` matching the directory name (lowercase, hyphens)
- A `description` (50+ chars)
- A `metadata.team` field

See [agentskills.io/specification](https://agentskills.io/specification) for the full spec.

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
# Set up with this repo (owner/repo shorthand works everywhere)
myskills init jverhoeks/myskills

# Add skills from skills.sh
myskills add-skill vercel-labs/agent-skills

# Sync all enabled skills to your tools
myskills sync

# See what's installed
myskills list
# REPO           SKILL                        ENABLED  STATUS   SYNCED
# myskills       dependency-bloat-reduction    yes      current  2026-04-04 14:30
# agent-skills   react-best-practices         yes      current  2026-04-04 14:30
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

```bash
myskills enable
```

```
  Select skills to enable

> [x] dependency-bloat-reduction — Analyze imports, identify trivial/outdated...
  [x] react-best-practices — React patterns and best practices
  [ ] experimental-thing — Work in progress, not ready yet

  space: toggle  a: toggle all  enter/q: save  esc: cancel
```

Disabled skills are removed from your tool directories on the next `myskills sync`.

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
