#!/bin/bash
# Post-build hook for GoReleaser — verifies binary and strips debug info
set -euo pipefail

BINARY_PATH="$1"
BINARY_NAME="$2"

if [ ! -f "$BINARY_PATH" ]; then
  echo "ERROR: Binary not found at $BINARY_PATH"
  exit 1
fi

# Check binary is executable
if [ ! -x "$BINARY_PATH" ]; then
  chmod +x "$BINARY_PATH"
fi

# Verify it can print its version
if ! "$BINARY_PATH" --version > /dev/null 2>&1; then
  echo "ERROR: Binary $BINARY_PATH failed --version check"
  exit 1
fi

# Print size info
SIZE=$(ls -lh "$BINARY_PATH" | awk '{print $5}')
echo "  ✓ $BINARY_NAME ($SIZE) — smoke test passed"
