#!/bin/bash
# perfmon — update.sh
# Checks the currently installed version against the latest GitHub release
# and upgrades if a newer version is available.
# Usage: ./scripts/update.sh
#   or:  curl -sfL https://get.perfmon.qzz.io/update | bash

set -euo pipefail

REPO="GAM3RG33K/perfmon-lite"
BIN_NAME="perfmon-tool"

# ── Locate installed binary ────────────────────────────────────────────
BIN_PATH="$(command -v "$BIN_NAME" 2>/dev/null || true)"
if [ -z "$BIN_PATH" ]; then
  # Check common install locations
  for dir in /usr/local/bin "${HOME}/.local/bin"; do
    if [ -x "${dir}/${BIN_NAME}" ]; then
      BIN_PATH="${dir}/${BIN_NAME}"
      break
    fi
  done
fi

if [ -z "$BIN_PATH" ]; then
  echo "ERROR: perfmon not found in PATH or common locations."
  echo "       Run scripts/install.sh first."
  exit 1
fi

# ── Detect current version ─────────────────────────────────────────────
CURRENT="$("${BIN_PATH}" --version 2>/dev/null | sed 's/.*v//' || true)"
if [ -z "$CURRENT" ]; then
  echo "ERROR: could not detect installed version from ${BIN_PATH}"
  exit 1
fi
echo "  installed:  v${CURRENT}"

# ── Fetch latest release from GitHub ───────────────────────────────────
echo "  checking GitHub..."
LATEST="$(
  curl -sfL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | sed 's/.*"tag_name": *"\(.*\)",*/\1/'
)"

if [ -z "$LATEST" ]; then
  echo "ERROR: could not fetch latest release from GitHub"
  exit 1
fi
echo "  latest:     ${LATEST}"

# ── Compare versions ───────────────────────────────────────────────────
LATEST_STR="${LATEST#v}"
if [ "$CURRENT" = "$LATEST_STR" ]; then
  echo ""
  echo "  ✓ perfmon is already up to date (v${CURRENT})"
  exit 0
fi

echo ""
echo "  New version available: ${LATEST} (current: v${CURRENT})"

# ── Detect platform ────────────────────────────────────────────────────
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Linux)  GOOS="linux"  ;;
  Darwin) GOOS="darwin" ;;
  *)      echo "ERROR: unsupported OS: $OS"; exit 1 ;;
esac

case "$ARCH" in
  x86_64|amd64) GOARCH="amd64" ;;
  aarch64|arm64) GOARCH="arm64" ;;
  *)            echo "ERROR: unsupported arch: $ARCH"; exit 1 ;;
esac

# ── Download and replace ───────────────────────────────────────────────
VER="${LATEST_STR}"
if [ "$GOOS" = "darwin" ]; then
  ASSET="perfmon-tool-${VER}-darwin-${GOARCH}"
elif [ "$GOOS" = "windows" ]; then
  ASSET="perfmon-tool-${VER}-windows-${GOARCH}.exe"
else
  ASSET="perfmon-tool-${VER}-linux-${GOARCH}"
fi
URL="https://github.com/${REPO}/releases/download/${LATEST}/${ASSET}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "  downloading: ${URL}"
curl -sfL "$URL" -o "${TMPDIR}/${BIN_NAME}"

# Preserve the original binary location and permissions
ORIG_MODE=$(stat -f "%Lp" "$BIN_PATH" 2>/dev/null || echo "755")
echo "  upgrading:  ${BIN_PATH}"

if [ -w "$BIN_PATH" ]; then
  cp "${TMPDIR}/${BIN_NAME}" "$BIN_PATH"
  chmod "$ORIG_MODE" "$BIN_PATH"
else
  # Need sudo for system directories
  echo "  (escalating with sudo...)"
  sudo cp "${TMPDIR}/${BIN_NAME}" "$BIN_PATH"
  sudo chmod "$ORIG_MODE" "$BIN_PATH"
fi

# ── Verify ─────────────────────────────────────────────────────────────
if ! "$BIN_PATH" --version > /dev/null 2>&1; then
  echo "ERROR: updated binary failed --version check"
  exit 1
fi

NEW_VER="$("${BIN_PATH}" --version 2>/dev/null | sed 's/.*v//')"
echo ""
echo "  ─────────────────────────────────────"
echo "   perfmon updated: v${CURRENT} → v${NEW_VER}"
echo "   Binary: ${BIN_PATH}"
echo "  ─────────────────────────────────────"
