#!/bin/bash
# Post-build hook for GoReleaser — verifies binary is valid
# Only executes a smoke test (--version) on binaries matching the host platform.
set -euo pipefail

BINARY_PATH="$1"
BINARY_NAME="$2"

if [ ! -f "$BINARY_PATH" ]; then
  echo "ERROR: Binary not found at $BINARY_PATH"
  exit 1
fi

chmod +x "$BINARY_PATH" 2>/dev/null || true

SIZE=$(ls -lh "$BINARY_PATH" | awk '{print $5}')

# Detect host platform (Go naming conventions)
HOST_OS="$(uname -s)"
HOST_ARCH="$(uname -m)"
case "$HOST_OS" in
  Linux)  HOST_OS="linux"  ;;
  Darwin) HOST_OS="darwin" ;;
  *)      HOST_OS=""       ;;
esac
case "$HOST_ARCH" in
  x86_64|amd64) HOST_ARCH="amd64" ;;
  aarch64|arm64) HOST_ARCH="arm64" ;;
  *)            HOST_ARCH=""       ;;
esac

# Check if this binary matches the host platform
BINARY_DIR="$(dirname "$BINARY_PATH")"
DIR_NAME="$(basename "$BINARY_DIR")"

if [ -n "$HOST_OS" ] && [ -n "$HOST_ARCH" ] && \
   echo "$DIR_NAME" | grep -qi "${HOST_OS}.*${HOST_ARCH}" 2>/dev/null; then
  # Native binary — run smoke test
  if ! "$BINARY_PATH" --version > /dev/null 2>&1; then
    echo "ERROR: $BINARY_NAME ($SIZE) — smoke test failed"
    exit 1
  fi
  echo "  ✓ $BINARY_NAME ($SIZE) — smoke test passed"
else
  # Cross-compiled binary — just report size
  echo "  ~ $BINARY_NAME ($SIZE) — cross-compiled, skipping smoke test"
fi
