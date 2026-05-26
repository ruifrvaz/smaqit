#!/usr/bin/env bash
# scripts/install.sh
# Installs pre-commit.sh as a git hook for this repository.
# Run from the repository root: bash .github/skills/smaqit.infrastructure-hook-pre-commit-validate/scripts/install.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HOOK_SOURCE="$SCRIPT_DIR/pre-commit.sh"
HOOK_TARGET=".git/hooks/pre-commit"

if [ ! -d ".git" ]; then
  echo "Error: run this script from the repository root." >&2
  exit 1
fi

cp "$HOOK_SOURCE" "$HOOK_TARGET"
chmod +x "$HOOK_TARGET"
echo "pre-commit hook installed. Run \`git commit\` to test."
