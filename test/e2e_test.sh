#!/bin/bash
set -euo pipefail

echo "=== myskills e2e test ==="

TESTDIR=$(mktemp -d)
trap "rm -rf $TESTDIR 2>/dev/null || true" EXIT

export HOME="$TESTDIR/home"
export XDG_CONFIG_HOME="$TESTDIR/home/.config"
export XDG_CACHE_HOME="$TESTDIR/home/.cache"
mkdir -p "$HOME"

# Build
echo "Building..."
go build -o "$TESTDIR/myskills" ./cmd/myskills

BIN="$TESTDIR/myskills"

# Create two fake remote repos
REMOTE1="$TESTDIR/remote1"
REMOTE2="$TESTDIR/remote2"
mkdir -p "$REMOTE1" "$REMOTE2"
git -C "$REMOTE1" init --bare 2>/dev/null
git -C "$REMOTE2" init --bare 2>/dev/null

# Populate first remote (org skills)
WORK1="$TESTDIR/work1"
mkdir -p "$WORK1"
git -C "$WORK1" init 2>/dev/null
git -C "$WORK1" config user.email "test@test.com"
git -C "$WORK1" config user.name "Test"

mkdir -p "$WORK1/skills/deploy"
cat > "$WORK1/skills/deploy/SKILL.md" << 'EOF'
---
name: deploy
description: Deploy the application to production with safety checks and rollback support.
metadata:
  team: platform
---

# Deploy

Run the deploy script.
EOF

mkdir -p "$WORK1/skills/review"
cat > "$WORK1/skills/review/SKILL.md" << 'EOF'
---
name: review
description: Code review workflow with automated checks and style enforcement.
metadata:
  team: platform
---

# Review

Review the code.
EOF

cat > "$WORK1/.myskills.yaml" << 'EOF'
org: test
validation:
  description_min_length: 20
  required_metadata:
    - team
  allowed_teams:
    - platform
    - infra
  max_skill_md_lines: 500
EOF

git -C "$WORK1" add .
git -C "$WORK1" commit -m "init" 2>/dev/null
git -C "$WORK1" remote add origin "$REMOTE1"
git -C "$WORK1" push origin HEAD:main 2>/dev/null

# Populate second remote (community skills)
WORK2="$TESTDIR/work2"
mkdir -p "$WORK2"
git -C "$WORK2" init 2>/dev/null
git -C "$WORK2" config user.email "test@test.com"
git -C "$WORK2" config user.name "Test"

mkdir -p "$WORK2/skills/debug"
cat > "$WORK2/skills/debug/SKILL.md" << 'EOF'
---
name: debug
description: Systematic debugging workflow for finding and fixing bugs quickly.
metadata:
  team: community
---

# Debug

Debug the issue.
EOF

git -C "$WORK2" add .
git -C "$WORK2" commit -m "init" 2>/dev/null
git -C "$WORK2" remote add origin "$REMOTE2"
git -C "$WORK2" push origin HEAD:main 2>/dev/null

# Create fake .claude directory
mkdir -p "$HOME/.claude"

# Test: validate
echo "Testing validate..."
$BIN validate "$WORK1/skills/deploy"
echo "  ✓ validate works"

# Manually set up config (since init is interactive for tool detection)
mkdir -p "$XDG_CONFIG_HOME/myskills"
cat > "$XDG_CONFIG_HOME/myskills/config.yaml" << EOF
repos:
  - name: org
    url: $REMOTE1
  - name: community
    url: $REMOTE2
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

# Clone both repos to cache
mkdir -p "$XDG_CACHE_HOME/myskills/repos"
git clone "$REMOTE1" "$XDG_CACHE_HOME/myskills/repos/org" 2>/dev/null
git clone "$REMOTE2" "$XDG_CACHE_HOME/myskills/repos/community" 2>/dev/null

# Test: sync (multi-repo)
echo "Testing sync (multi-repo)..."
$BIN sync
echo "  ✓ sync works"

# Verify symlinks exist
for skill in deploy review debug; do
  link="$HOME/.claude/skills/$skill"
  if [ -L "$link" ]; then
    echo "  ✓ $skill is a symlink"
  else
    echo "  ✗ $skill is NOT a symlink" >&2
    exit 1
  fi
  # Verify SKILL.md is accessible through symlink
  if [ -f "$link/SKILL.md" ]; then
    echo "  ✓ $skill/SKILL.md accessible"
  else
    echo "  ✗ $skill/SKILL.md NOT accessible" >&2
    exit 1
  fi
done

# Test: list (multi-repo)
echo "Testing list..."
$BIN list
echo "  ✓ list works"

# Test: info
echo "Testing info..."
$BIN info deploy
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
$BIN remove deploy
if [ -L "$HOME/.claude/skills/deploy" ] || [ -d "$HOME/.claude/skills/deploy" ]; then
  echo "  ✗ deploy NOT removed" >&2
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

# Test: sync single skill
echo "Testing sync single skill..."
$BIN sync debug
if [ -L "$HOME/.claude/skills/debug" ]; then
  echo "  ✓ single skill sync works"
else
  echo "  ✗ single skill sync failed" >&2
  exit 1
fi

echo ""
echo "=== All e2e tests passed ==="
