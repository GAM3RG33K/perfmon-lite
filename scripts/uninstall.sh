#!/bin/bash
# perfmon — uninstall.sh
# Removes perfmon-tool binary from common install locations.
set -euo pipefail

REMOVED=false

echo "Uninstalling perfmon-tool..."

# Common locations — check both old name (perfmon) and new name (perfmon-tool)
for dir in /usr/local/bin "${HOME}/.local/bin"; do
  for bin in perfmon-tool perfmon; do
    path="${dir}/${bin}"
    if [ -f "$path" ]; then
      rm -f "$path"
      echo "  Removed ${path}"
      REMOVED=true
    fi
  done
done

if [ "$REMOVED" = false ]; then
  echo "  perfmon-tool not found in common locations."
fi

echo ""
echo "  ─────────────────────────────────────"
echo "  Goodbye! Thanks for trying perfmon-tool."
echo ""
echo "  To reinstall:"
echo "    curl -sfL https://get.perfmon.qzz.io | bash"
echo "  ─────────────────────────────────────"
