#!/usr/bin/env bash
# Rotate a single Vault credential path for a smaqit project.
# Deletes the named path and re-runs load-credentials.sh to repopulate it.
# All sensitive input is via hidden prompts — nothing enters shell history.
#
# Usage: bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/rotate-credential.sh <path>
#   where <path> is one of: cyso | ssh | tfstate | github
#
# Example:
#   bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/rotate-credential.sh cyso

set -euo pipefail

export VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

CREDENTIAL_PATH="${1:-}"
if [ -z "$CREDENTIAL_PATH" ]; then
  echo "Usage: $0 <path>   (one of: cyso | ssh | tfstate | github)"
  exit 1
fi

case "$CREDENTIAL_PATH" in
  cyso|ssh|tfstate|github) ;;
  *) echo "ERROR: Unknown path '$CREDENTIAL_PATH'. Must be one of: cyso ssh tfstate github"; exit 1 ;;
esac

# ── Derive project slug ───────────────────────────────────────────────────────

# CLAUDE.md takes precedence when present; falls back to Copilot's instructions file.
INSTRUCTIONS_FILE="CLAUDE.md"
if [ ! -f "$INSTRUCTIONS_FILE" ]; then
  INSTRUCTIONS_FILE=".github/copilot-instructions.md"
fi
if [ ! -f "$INSTRUCTIONS_FILE" ]; then
  echo "ERROR: Run from repo root (neither CLAUDE.md nor .github/copilot-instructions.md found)"
  exit 1
fi

PROJECT_SLUG=$(grep -i "project name" "$INSTRUCTIONS_FILE" | grep -i ": " | head -1 \
  | sed 's/.*: *//' | sed 's/[^a-zA-Z0-9 -].*$//' \
  | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | tr -s '-' || true)

if [ -z "$PROJECT_SLUG" ]; then
  read -r -p "Enter project slug manually: " PROJECT_SLUG
fi

FULL_PATH="secret/${PROJECT_SLUG}/${CREDENTIAL_PATH}"

# ── Confirm ───────────────────────────────────────────────────────────────────

echo "==> About to delete and re-populate: $FULL_PATH"
read -r -p "    Continue? [y/N] " CONFIRM
case "$CONFIRM" in
  y|Y) ;;
  *) echo "Aborted."; exit 0 ;;
esac

# ── Delete existing path ──────────────────────────────────────────────────────

if vault kv get "$FULL_PATH" > /dev/null 2>&1; then
  vault kv delete "$FULL_PATH"
  echo "    Deleted $FULL_PATH"
else
  echo "    Path does not exist — will be created fresh"
fi

# ── Re-run credential loader (it will prompt only for the deleted path) ───────

echo "==> Running credential loader to repopulate $CREDENTIAL_PATH..."
bash "${SCRIPT_DIR}/load-credentials.sh"

echo ""
echo "==> Rotation complete for $FULL_PATH"
echo "    Re-run smaqit.infrastructure-repo-config to sync new value to GitHub Secrets."
