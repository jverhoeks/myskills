# 🧠 myskills

Distribute AI agent skills across your team from a central Git repo. Works with **Claude Code**, **GitHub Copilot CLI**, **OpenAI Codex CLI**, and **OpenCode**.

Skills follow the open [Agent Skills](https://agentskills.io) standard — write once, use everywhere.

---

## 📚 Skills

This repo contains shared skills for the team. Browse the [`skills/`](skills/) directory or use `myskills list` after installing.

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

A Go CLI tool that syncs skills from Git repos into your AI coding tools via symlinks.

### Features

- 🔗 **Symlinks** — skills point to the cached repo, always up to date after `git pull`
- 📦 **Multi-repo** — pull skills from multiple Git repos (org, community, public)
- ✅ **Validation** — enforces the Agent Skills spec + org-specific rules
- 🎛️ **Per-skill toggle** — interactive TUI to enable/disable individual skills
- 🔍 **Auto-detection** — finds which AI tools you have installed
- 🖥️ **Cross-platform** — macOS, Linux, Windows binaries

### Install

**From GitHub Releases (recommended):**

```bash
curl -sSL https://raw.githubusercontent.com/jverhoeks/myskills/main/install.sh | bash
```

**Or download manually** from [Releases](https://github.com/jverhoeks/myskills/releases/latest) — binaries for `darwin/arm64`, `darwin/amd64`, `linux/amd64`, `linux/arm64`, `windows/amd64`, `windows/arm64`.

**From source:**

```bash
go install github.com/jverhoeks/myskills/cmd/myskills@latest
```

### Quick start

```bash
# Set up with this repo
myskills init https://github.com/jverhoeks/myskills.git org

# Sync all enabled skills to your tools
myskills sync

# See what's installed
myskills list
```

### Commands

| Command | Description |
|---------|-------------|
| `myskills init <url\|owner/repo> [name]` | 🚀 Set up with a repo, detect tools, write config |
| `myskills add-repo <url\|owner/repo> [name]` | ➕ Add another skill repository |
| `myskills add-skill <owner/repo>` | 🌐 Add a skill from [skills.sh](https://skills.sh) / GitHub |
| `myskills sync [skill]` | 🔄 Pull latest and symlink skills to tool directories |
| `myskills list` | 📋 List skills with enabled/synced status |
| `myskills info <name>` | ℹ️ Show skill details |
| `myskills enable` | 🎛️ Interactive TUI to toggle skills on/off |
| `myskills validate <path>` | ✅ Validate a skill against spec + org rules |
| `myskills dev <name>` | 🆕 Scaffold a new skill |
| `myskills submit <name>` | 📤 Validate and open a PR |
| `myskills remove <name>` | 🗑️ Remove a skill from all tool directories |
| `myskills doctor` | 🩺 Health check: repos, tools, config |
| `myskills config list` | ⚙️ Show current config |
| `myskills config set <key> <val>` | ⚙️ Set a config value |

### 🌐 skills.sh compatibility

Add any skill from [skills.sh](https://skills.sh) using `owner/repo` shorthand — the same repos that work with `npx skills add` work here:

```bash
# Add Vercel's agent skills
myskills add-skill vercel-labs/agent-skills

# Add Microsoft's skills
myskills add-skill microsoft/skills

# Sync to install them
myskills sync
```

Skills are auto-discovered in `skills/`, `.agents/skills/`, `.claude/skills/`, `.github/skills/`, or as a root `SKILL.md`.

### Multi-repo setup

```bash
# All commands accept owner/repo shorthand
myskills add-repo my-org/community-skills

# Sync pulls from all repos
myskills sync

# Skills show their source repo
myskills list
# REPO        SKILL                        ENABLED  STATUS   SYNCED
# org         dependency-bloat-reduction    yes      current  2026-04-04 14:30
# community   frontend-patterns            yes      current  2026-04-04 14:30
```

### Enable/disable skills

```bash
# Interactive picker
myskills enable
```

```
  Select skills to enable

> [x] dependency-bloat-reduction — Analyze imports, identify trivial/outdated...
  [x] frontend-patterns — Common React patterns for the team
  [ ] experimental-thing — Work in progress, not ready yet

  space: toggle  a: toggle all  enter/q: save  esc: cancel
```

### Supported tools

| Tool | Detected by | Skills path |
|------|-------------|-------------|
| 🟣 Claude Code | `~/.claude/` | `~/.claude/skills/` |
| 🐙 GitHub Copilot CLI | `~/.copilot/` | `~/.copilot/skills/` |
| 🟢 OpenAI Codex CLI | `~/.codex/` | `~/.codex/skills/` |
| 🔵 OpenCode | `~/.config/opencode/` | `~/.config/opencode/skills/` |

### Config

Config lives at `~/.config/myskills/config.yaml`. Cache at `~/.cache/myskills/`.

On Windows: `%APPDATA%\myskills\` for config, `%LOCALAPPDATA%\myskills\cache\` for cache.

### Authentication

For private repos, use SSH keys or set `GITHUB_TOKEN` / `GITLAB_TOKEN` — git and `gh` pick them up automatically. No tokens stored in config files.

---

## 📄 License

MIT
