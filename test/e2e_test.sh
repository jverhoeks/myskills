#!/bin/bash
set -euo pipefail

echo "=== myskills e2e test ==="

TESTDIR=$(mktemp -d)
trap "rm -rf $TESTDIR" EXIT

export HOME="$TESTDIR/home"
export XDG_CONFIG_HOME="$TESTDIR/home/.config"
mkdir -p "$HOME"

# Build
echo "Building..."
go build -o "$TESTDIR/myskills" ./cmd/myskills

BIN="$TESTDIR/myskills"

# Create a fake remote repo
REMOTE="$TESTDIR/remote"
mkdir -p "$REMOTE"
git -C "$REMOTE" init --bare 2>/dev/null

WORK="$TESTDIR/work"
mkdir -p "$WORK"
git -C "$WORK" init 2>/dev/null
git -C "$WORK" config user.email "test@test.com"
git -C "$WORK" config user.name "Test"

# Create a skill in the work repo
mkdir -p "$WORK/skills/hello"
cat > "$WORK/skills/hello/SKILL.md" << 'SKILLEOF'
---
name: hello
description: A test skill that greets the user with helpful instructions for getting started.
metadata:
  team: platform
---

# Hello

Say hello to the user.
SKILLEOF

# Create .myskills.yaml
cat > "$WORK/.myskills.yaml" << 'RULESEOF'
org: test
validation:
  description_min_length: 20
  required_metadata:
    - team
  allowed_teams:
    - platform
    - infra
  max_skill_md_lines: 500
RULESEOF

git -C "$WORK" add .
git -C "$WORK" commit -m "init" 2>/dev/null
git -C "$WORK" remote add origin "$REMOTE"
git -C "$WORK" push origin HEAD:main 2>/dev/null

# Create fake .claude directory so tool detection works
mkdir -p "$HOME/.claude"

# Test: validate
echo "Testing validate..."
$BIN validate "$WORK/skills/hello"
echo "  ✓ validate works"

# Test: config set (create config manually since init is interactive)
mkdir -p "$XDG_CONFIG_HOME/myskills"
cat > "$XDG_CONFIG_HOME/myskills/config.yaml" << EOF
repo: $REMOTE
cache_dir: $XDG_CONFIG_HOME/myskills/repo
github:
  method: gh
targets:
  claude:
    enabled: true
    skill_path: $HOME/.claude/skills
  copilot:
    enabled: false
    skill_path: $HOME/.copilot/skills
  codex:
    enabled: false
    skill_path: $HOME/.codex/skills
  opencode:
    enabled: false
    skill_path: $HOME/.config/opencode/skills
EOF

# Clone the repo cache
git clone "$REMOTE" "$XDG_CONFIG_HOME/myskills/repo" 2>/dev/null

# Test: sync
echo "Testing sync..."
$BIN sync
echo "  ✓ sync works"

# Verify files were copied
if [ -f "$HOME/.claude/skills/hello/SKILL.md" ]; then
  echo "  ✓ skill copied to claude"
else
  echo "  ✗ skill NOT found in claude dir" >&2
  exit 1
fi

# Test: list
echo "Testing list..."
$BIN list
echo "  ✓ list works"

# Test: info
echo "Testing info..."
$BIN info hello
echo "  ✓ info works"

# Test: doctor
echo "Testing doctor..."
$BIN doctor
echo "  ✓ doctor works"

# Test: config list
echo "Testing config list..."
$BIN config list
echo "  ✓ config list works"

# Test: remove
echo "Testing remove..."
$BIN remove hello
if [ -d "$HOME/.claude/skills/hello" ]; then
  echo "  ✗ skill NOT removed" >&2
  exit 1
fi
echo "  ✓ remove works"

# Test: dev
echo "Testing dev..."
$BIN dev test-skill
if [ -f "$XDG_CONFIG_HOME/myskills/dev/test-skill/SKILL.md" ]; then
  echo "  ✓ dev scaffold works"
else
  echo "  ✗ dev scaffold failed" >&2
  exit 1
fi

echo ""
echo "=== All e2e tests passed ==="
