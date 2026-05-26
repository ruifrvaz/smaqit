#!/usr/bin/env bash
# check-amendments.sh
# Scans spec files for `amendment:` annotations left by implementation phases.
# Usage: bash check-amendments.sh <specs-dir>
# Exit 0 = no amendments found (Phase 8 review step can be skipped)
# Exit 1 = amendments found (Phase 8 review step must run)

set -euo pipefail

SPECS_DIR="${1:-specs}"

if [[ ! -d "$SPECS_DIR" ]]; then
  echo "ERROR: directory not found: $SPECS_DIR" >&2
  exit 2
fi

MATCHES=$(grep -rl "amendment:" "$SPECS_DIR" 2>/dev/null || true)

if [[ -z "$MATCHES" ]]; then
  echo "PASS: no amendment annotations found in $SPECS_DIR — review step skipped."
  exit 0
fi

echo "AMENDMENTS FOUND — review required before release:"
echo ""
while IFS= read -r file; do
  echo "  $file"
  grep -n "amendment:" "$file" | sed 's/^/    /'
  echo ""
done <<< "$MATCHES"

exit 1
