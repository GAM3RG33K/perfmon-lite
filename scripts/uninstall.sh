#!/bin/bash
# perfmon — uninstall.sh
# Removes perfmon binary from common install locations.
set -euo pipefail

BIN_NAME="perfmon-tool"
REMOVED=false

echo "Uninstalling perfmon..."

# Common locations
for dir in /usr/local/bin "${HOME}/.local/bin"; do
  path="${dir}/${BIN_NAME}"
  if [ -f "$path" ]; then
    rm -f "$path"
    echo "  Removed ${path}"
    REMOVED=true
  fi
done

if [ "$REMOVED" = false ]; then
  echo "  perfmon not found in common locations."
  echo "  You may have installed it in a custom path — delete it manually."
  exit 0
fi

echo "  perfmon uninstalled successfully!"
