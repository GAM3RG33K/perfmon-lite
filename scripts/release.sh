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

# ── Checks ──────────────────────────────────────────────────────────────
if [ -n "$(git status --porcelain)" ]; then
  echo "ERROR: working tree is dirty — commit or stash changes first"
  exit 1
fi

if git rev-parse "v$VERSION" >/dev/null 2>&1; then
  echo "ERROR: tag v$VERSION already exists locally"
  exit 1
fi

echo ""
echo "─── Release Checklist ───"
echo "  source version:  $VERSION"
echo "  tag to create:   v$VERSION"
echo "  target branch:   $(git rev-parse --abbrev-ref HEAD)"
echo ""
echo "Press Enter to create and push tag v$VERSION, or Ctrl+C to cancel."
read -r

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
