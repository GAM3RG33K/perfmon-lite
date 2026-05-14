#!/bin/bash
# perfmon — install.sh
# Downloads the latest release binary for your OS/arch and installs it.
# Usage: curl -sfL https://perfmon.qzz.io | bash
#   or:  ./scripts/install.sh [--prefix /usr/local]

set -euo pipefail

REPO="GAM3RG33K/perfmon-lite"
BIN_NAME="perfmon"

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

# ── Resolve install directory ──────────────────────────────────────────
PREFIX="${1:-}"
if [ -z "$PREFIX" ]; then
  if [ -w "/usr/local/bin" ]; then
    PREFIX="/usr/local"
  else
    PREFIX="${HOME}/.local"
    mkdir -p "$PREFIX/bin"
  fi
fi
INSTALL_DIR="${PREFIX}/bin"

# ── Fetch latest release tag from GitHub ───────────────────────────────
echo "  Checking latest release..."
LATEST="$(
  curl -sfL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | sed 's/.*"tag_name": *"\(.*\)",*/\1/'
)"

if [ -z "$LATEST" ]; then
  echo "ERROR: could not fetch latest release from GitHub"
  exit 1
fi
echo "  latest release: $LATEST"

# ── Download binary ────────────────────────────────────────────────────
ASSET="perfmon_${LATEST#v}_${GOOS}_${GOARCH}"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${ASSET}"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "  downloading: ${URL}"
curl -sfL "$URL" -o "${TMPDIR}/${BIN_NAME}"
echo ""

# ── Install ────────────────────────────────────────────────────────────
install -m 755 "${TMPDIR}/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"

echo "  installed: ${INSTALL_DIR}/${BIN_NAME} ($(ls -lh "${INSTALL_DIR}/${BIN_NAME}" | awk '{print $5}'))"
echo ""

# ── Verify ─────────────────────────────────────────────────────────────
if ! "${INSTALL_DIR}/${BIN_NAME}" --version > /dev/null 2>&1; then
  echo "WARNING: installed binary failed --version check"
  exit 1
fi

echo "  ─────────────────────────────────────"
echo "   perfmon ${LATEST} installed!"
echo "   Binary: ${INSTALL_DIR}/${BIN_NAME}"
echo ""
echo "   Make sure ${INSTALL_DIR} is in your PATH."
echo "   Run:    ${BIN_NAME} --mock"
echo "  ─────────────────────────────────────"
