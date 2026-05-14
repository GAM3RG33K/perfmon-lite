#!/bin/bash
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
VERSION=$(grep -E '\bversion\s*=' "$ROOT/cmd/perfmon/main.go" | sed 's/.*"\(.*\)".*/\1/')

if [ -z "$VERSION" ]; then
  echo "ERROR: could not detect version from cmd/perfmon/main.go"
  exit 1
fi

echo "  detected version: $VERSION"
echo "  tag:              v$VERSION"

# ── Parse flags ─────────────────────────────────────────────────────────
RETAG=false
while [ $# -gt 0 ]; do
  case "$1" in
    --retag|-f) RETAG=true; shift ;;
    *) echo "ERROR: unknown flag $1 (usage: --retag|-f to re-tag)"; exit 1 ;;
  esac
done

# ── Checks ──────────────────────────────────────────────────────────────
if [ -n "$(git status --porcelain)" ]; then
  echo "ERROR: working tree is dirty — commit or stash changes first"
  exit 1
fi

TAG_EXISTS=false
if git rev-parse "v$VERSION" >/dev/null 2>&1; then
  TAG_EXISTS=true
  if [ "$RETAG" = false ]; then
    echo "ERROR: tag v$VERSION already exists locally (use --retag to replace)"
    exit 1
  fi
fi

# ── Confirm ─────────────────────────────────────────────────────────────
echo ""
echo "─── Release Checklist ───"
echo "  source version:  $VERSION"
echo "  tag:             v$VERSION"
echo "  target branch:   $(git rev-parse --abbrev-ref HEAD)"
if [ "$TAG_EXISTS" = true ] && [ "$RETAG" = true ]; then
  echo "  mode:            RETAG (delete existing + recreate)"
fi
echo ""
echo "Press Enter to continue, or Ctrl+C to cancel."
read -r

# ── Re-tag if needed ────────────────────────────────────────────────────
if [ "$TAG_EXISTS" = true ] && [ "$RETAG" = true ]; then
  echo ""
  echo "  Deleting existing local tag v$VERSION..."
  git tag -d "v$VERSION" > /dev/null 2>&1

  echo "  Deleting existing remote tag v$VERSION..."
  git push --delete origin "v$VERSION" 2>/dev/null || \
    echo "  (no remote tag to delete)"
fi

# ── Tag and push ─────────────────────────────────────────────────────────
echo ""
echo "  Creating tag v$VERSION..."
git tag -a "v$VERSION" -m "Release v$VERSION"

echo "  Pushing tag..."
git push origin "v$VERSION"

echo ""
echo "  ✓ Release v$VERSION triggered!"
echo "  View progress: https://github.com/GAM3RG33K/perfmon-lite/actions"
echo "  View release:  https://github.com/GAM3RG33K/perfmon-lite/releases"
