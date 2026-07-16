#!/usr/bin/env bash
# smaqit.infrastructure-vault-loader — interactive credential loader
# Handles Vault start, unseal, and login — then prompts for all project secrets.
# All sensitive input uses hidden prompts (read -s). Nothing sensitive is ever
# passed as a command argument or written to shell history.
# Skips credential paths that already exist. Generates SSH keypair automatically.
#
# Usage: bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh

set -euo pipefail

# ── Environment ───────────────────────────────────────────────────────────────

export VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"

# ── Helper: read a secret without echoing, then export to named variable ──────
# Usage: read_secret VAR_NAME "Prompt label"
read_secret() {
  local _var="$1"
  local _prompt="$2"
  local _value
  IFS= read -rs -p "  ${_prompt}: " _value </dev/tty && echo >&2
  # Assign to the caller's variable name without eval
  printf -v "$_var" '%s' "$_value"
  unset _value
}

# ── Step 0: Start Vault if not reachable ──────────────────────────────────────

echo "==> Checking Vault status..."
# vault status exit codes: 0=running+unsealed, 1=error/unreachable, 2=running+sealed
# Use set +e to safely capture the non-zero exit without aborting
set +e
vault status > /dev/null 2>&1
VAULT_STATUS=$?
set -e
if [ "$VAULT_STATUS" -eq 1 ]; then
  echo "    Vault not running at $VAULT_ADDR — attempting to start..."
  VAULT_CONFIG="${HOME}/.vault/config.hcl"
  if [ ! -f "$VAULT_CONFIG" ]; then
    echo "ERROR: Config not found at $VAULT_CONFIG"
    echo "       Follow one-time setup in the skill README, then re-run."
    exit 1
  fi
  vault server -config="$VAULT_CONFIG" > /tmp/vault-server.log 2>&1 &
  VAULT_PID=$!
  sleep 4
  set +e
  vault status > /dev/null 2>&1
  VAULT_STATUS=$?
  set -e
  if [ "$VAULT_STATUS" -eq 1 ]; then
    echo "ERROR: Failed to start Vault. Check /tmp/vault-server.log"
    exit 1
  fi
  echo "    Started Vault (PID $VAULT_PID)"
fi

# ── Step 0b: Unseal if sealed ─────────────────────────────────────────────────

set +e
SEALED=$(vault status -format=json 2>/dev/null | python3 -c "import sys,json; print(json.load(sys.stdin)['sealed'])" 2>/dev/null)
[ -z "$SEALED" ] && SEALED="true"
set -e
if [ "$SEALED" = "True" ] || [ "$SEALED" = "true" ]; then
  echo "    Vault is sealed — prompting for unseal key (input hidden)"
  read_secret _UNSEAL_KEY "Unseal Key"
  vault operator unseal "$_UNSEAL_KEY" > /dev/null
  unset _UNSEAL_KEY
  echo "    Unsealed"
fi

# ── Step 0c: Authenticate if no valid token ───────────────────────────────────

if ! vault token lookup > /dev/null 2>&1; then
  echo "    No valid token — prompting for root/scoped token (input hidden)"
  read_secret _VAULT_TOKEN "Vault Token"
  export VAULT_TOKEN="$_VAULT_TOKEN"
  unset _VAULT_TOKEN
  if ! vault token lookup > /dev/null 2>&1; then
    echo "ERROR: Token rejected by Vault."
    exit 1
  fi
  echo "    Authenticated"
fi

echo "    Vault: running, unsealed, authenticated"

# ── Derive project slug ───────────────────────────────────────────────────────

# CLAUDE.md takes precedence when present; falls back to Copilot's instructions file.
INSTRUCTIONS_FILE="CLAUDE.md"
if [ ! -f "$INSTRUCTIONS_FILE" ]; then
  INSTRUCTIONS_FILE=".github/copilot-instructions.md"
fi
if [ ! -f "$INSTRUCTIONS_FILE" ]; then
  echo "ERROR: Neither CLAUDE.md nor .github/copilot-instructions.md found. Run from repo root."
  exit 1
fi

# Handle both inline format ("Project Name: value") and heading+next-line format ("## Project Name\n\nvalue")
PROJECT_SLUG=$(grep -i "project name" "$INSTRUCTIONS_FILE" | grep -i ": " | head -1 | sed 's/.*: *//' | sed 's/[^a-zA-Z0-9 -].*$//' | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | tr -s '-' || true)
if [ -z "$PROJECT_SLUG" ]; then
  PROJECT_SLUG=$(awk '/^##? *[Pp]roject [Nn]ame/{found=1; next} found && /^[[:space:]]*$/{next} found{print; exit}' "$INSTRUCTIONS_FILE" \
    | sed 's/ .*$//' | tr '[:upper:]' '[:lower:]' | tr -s '-' || true)
fi

if [ -z "$PROJECT_SLUG" ]; then
  read -p "Could not derive project slug from $INSTRUCTIONS_FILE. Enter manually: " PROJECT_SLUG
fi

echo "==> Project slug: $PROJECT_SLUG"
echo ""

# ── Helper: check if path already populated ───────────────────────────────────

path_exists() {
  vault kv get "secret/${PROJECT_SLUG}/$1" > /dev/null 2>&1
}

# ── Step 1: Cyso app credentials ──────────────────────────────────────────────

echo "--- [1/4] Cyso Cloud app credentials (secret/${PROJECT_SLUG}/cyso) ---"
if path_exists "cyso"; then
  echo "    SKIP — path already populated"
else
  read -p "  app_credential_id: " CYSO_ID
  read -s -p "  app_credential_secret: " CYSO_SECRET && echo
  vault kv put "secret/${PROJECT_SLUG}/cyso" \
    app_credential_id="$CYSO_ID" \
    app_credential_secret="$CYSO_SECRET" > /dev/null
  unset CYSO_ID CYSO_SECRET
  echo "    DONE"
fi

# ── Step 2: SSH deploy keypair ────────────────────────────────────────────────

echo "--- [2/4] SSH deploy keypair (secret/${PROJECT_SLUG}/ssh) ---"
if path_exists "ssh"; then
  echo "    SKIP — path already populated"
else
  TMPDIR_KEY=$(mktemp -d)
  SSH_KEY_PATH="${TMPDIR_KEY}/deploy_key"
  ssh-keygen -t ed25519 -f "$SSH_KEY_PATH" -N "" -q
  vault kv put "secret/${PROJECT_SLUG}/ssh" \
    private_key="$(cat "${SSH_KEY_PATH}")" \
    public_key="$(cat "${SSH_KEY_PATH}.pub" | tr -d '\n')" > /dev/null
  rm -rf "$TMPDIR_KEY"
  echo "    DONE — ed25519 keypair generated and stored"
fi

# ── Step 3: Terraform state S3 credentials ────────────────────────────────────

echo "--- [3/4] Terraform state S3 credentials (secret/${PROJECT_SLUG}/tfstate) ---"
if path_exists "tfstate"; then
  echo "    SKIP — path already populated"
else
  read -p "  s3_access_key: " S3_KEY
  read -s -p "  s3_secret_key: " S3_SECRET && echo
  vault kv put "secret/${PROJECT_SLUG}/tfstate" \
    access_key="$S3_KEY" \
    secret_key="$S3_SECRET" > /dev/null
  unset S3_KEY S3_SECRET
  echo "    DONE"
fi

# ── Step 4: GitHub token ──────────────────────────────────────────────────────

echo "--- [4/4] GitHub fine-grained PAT (secret/${PROJECT_SLUG}/github) ---"
if path_exists "github"; then
  echo "    SKIP — path already populated"
else
  echo "  Required scopes: variables:write on the target repository"
  read -s -p "  github_token: " GH_TOKEN && echo
  vault kv put "secret/${PROJECT_SLUG}/github" \
    token="$GH_TOKEN" > /dev/null
  unset GH_TOKEN
  echo "    DONE"
fi

# ── Verification ──────────────────────────────────────────────────────────────

echo ""
echo "==> Verifying all paths..."
ALL_OK=true
for PATH_NAME in cyso ssh tfstate github; do
  if path_exists "$PATH_NAME"; then
    echo "    secret/${PROJECT_SLUG}/${PATH_NAME} — OK"
  else
    echo "    secret/${PROJECT_SLUG}/${PATH_NAME} — MISSING"
    ALL_OK=false
  fi
done

echo ""
if [ "$ALL_OK" = "true" ]; then
  echo "==> All credential paths populated. Vault ready for local deployment."
else
  echo "==> WARNING: One or more paths are missing. Re-run this script to fill them."
  exit 1
fi
