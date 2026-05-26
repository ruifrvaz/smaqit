#!/usr/bin/env bash
# smaqit.infrastructure-vault-loader — interactive credential loader
# Prompts for all project secrets and writes them to a local Vault instance.
# Skips paths that already exist. Generates SSH keypair automatically.
# Usage: bash .github/skills/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh

set -euo pipefail

# ── Environment ───────────────────────────────────────────────────────────────

export VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
export VAULT_TOKEN="${VAULT_TOKEN:-$(cat ~/.vault-token 2>/dev/null || true)}"

# ── Pre-flight: Vault must be running and unsealed ────────────────────────────

echo "==> Checking Vault status..."
if ! vault status > /dev/null 2>&1; then
  echo "ERROR: Cannot reach Vault at $VAULT_ADDR"
  echo "Start it with: vault server -config=~/.vault/config.hcl &"
  exit 1
fi

SEALED=$(vault status -format=json | python3 -c "import sys,json; print(json.load(sys.stdin)['sealed'])")
if [ "$SEALED" = "True" ]; then
  echo "ERROR: Vault is sealed. Run: vault operator unseal"
  exit 1
fi

if ! vault token lookup > /dev/null 2>&1; then
  echo "ERROR: No valid Vault token. Run: vault login"
  exit 1
fi

echo "    Vault: running, unsealed, authenticated"

# ── Derive project slug ───────────────────────────────────────────────────────

INSTRUCTIONS_FILE=".github/copilot-instructions.md"
if [ ! -f "$INSTRUCTIONS_FILE" ]; then
  echo "ERROR: $INSTRUCTIONS_FILE not found. Run from repo root."
  exit 1
fi

# Handle both inline format ("Project Name: value") and heading+next-line format ("## Project Name\n\nvalue")
PROJECT_SLUG=$(grep -i "project name" "$INSTRUCTIONS_FILE" | grep -i ": " | head -1 | sed 's/.*: *//' | sed 's/[^a-zA-Z0-9 -].*$//' | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | tr -s '-' || true)
if [ -z "$PROJECT_SLUG" ]; then
  PROJECT_SLUG=$(awk '/^##? *[Pp]roject [Nn]ame/{found=1; next} found && /^[[:space:]]*$/{next} found{print; exit}' "$INSTRUCTIONS_FILE" \
    | sed 's/ .*$//' | tr '[:upper:]' '[:lower:]' | tr -s '-' || true)
fi

if [ -z "$PROJECT_SLUG" ]; then
  read -p "Could not derive project slug from copilot-instructions.md. Enter manually: " PROJECT_SLUG
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
