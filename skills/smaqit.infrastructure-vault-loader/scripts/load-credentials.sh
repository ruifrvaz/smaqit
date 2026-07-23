#!/usr/bin/env bash
# smaqit.infrastructure-vault-loader — interactive credential loader
# Handles Vault start, unseal, and login — then prompts for all project secrets.
# All sensitive input uses hidden prompts (read -s). Nothing sensitive is ever
# passed as a command argument or written to shell history.
# Skips credential paths that already exist. Generates SSH keypair automatically.
#
# Usage: bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh
#
# Scheme (new apps/+machines/ vs legacy flat) is auto-detected — see "New scheme vs legacy flat
# scheme" below. For a new-scheme app's first load, pass MACHINE_SLUG explicitly:
#   MACHINE_SLUG=<machine-slug> bash .../load-credentials.sh
#
# PROVISIONING_MODE controls which paths are populated — legacy flat scheme only:
#   provision | existing-owned (default) — all four paths (cyso, ssh, tfstate, github)
#   existing-shared — only ssh + github; cyso/tfstate are never prompted for since
#     this project never provisions Terraform for a VM it doesn't own.
#
#   PROVISIONING_MODE=existing-shared bash .../load-credentials.sh

set -euo pipefail

# ── Environment ───────────────────────────────────────────────────────────────

export VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
PROVISIONING_MODE="${PROVISIONING_MODE:-provision}"

# If a previous `vault login` already wrote a token to disk, reuse it instead of prompting —
# harmless no-op if the file doesn't exist or VAULT_TOKEN is already exported.
export VAULT_TOKEN="${VAULT_TOKEN:-$(cat ~/.vault-token 2>/dev/null || true)}"

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

# AGENTS.md is shared by Codex and GitHub Copilot and is the canonical installed file.
# Fall back to legacy/platform-specific instruction files for older projects.
INSTRUCTIONS_FILE="AGENTS.md"
if [ ! -f "$INSTRUCTIONS_FILE" ]; then
  INSTRUCTIONS_FILE="CLAUDE.md"
fi
if [ ! -f "$INSTRUCTIONS_FILE" ]; then
  INSTRUCTIONS_FILE=".github/copilot-instructions.md"
fi
if [ ! -f "$INSTRUCTIONS_FILE" ]; then
  echo "ERROR: No AGENTS.md, CLAUDE.md, or .github/copilot-instructions.md found. Run from repo root."
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

# ── New scheme (apps/+machines/) vs legacy flat scheme ─────────────────────────
#
# An app is on the new scheme if it has already bootstrapped (secret/apps/<slug>/machine exists)
# or MACHINE_SLUG is passed explicitly (first-time load for an app that will bootstrap shortly via
# bootstrap-app-to-machine.sh). New-scheme apps never get ssh/cyso/tfstate from this script — ssh is
# exclusively bootstrap-app-to-machine.sh's job, and cyso/tfstate live at
# secret/machines/<machine-slug>/* now, populated once at machine registration. Legacy flat-scheme
# projects (no machine pointer, no MACHINE_SLUG) fall through to the unchanged behavior below.

APP_SLUG="$PROJECT_SLUG"
NEW_SCHEME=false
if [ -n "${MACHINE_SLUG:-}" ]; then
  NEW_SCHEME=true
elif vault kv get "secret/apps/${APP_SLUG}/machine" > /dev/null 2>&1; then
  NEW_SCHEME=true
  MACHINE_SLUG=$(vault kv get -field=machine-slug "secret/apps/${APP_SLUG}/machine")
fi

if [ "$NEW_SCHEME" = "true" ]; then
  echo "==> Scheme: new (secret/apps/${APP_SLUG}/*, secret/machines/${MACHINE_SLUG}/*)"
  echo ""
  echo "--- [1/1] GitHub fine-grained PAT (secret/apps/${APP_SLUG}/github) ---"
  if vault kv get "secret/apps/${APP_SLUG}/github" > /dev/null 2>&1; then
    echo "    SKIP — path already populated"
  else
    echo "  Required scopes: variables:write on the target repository"
    read -s -p "  github_token: " GH_TOKEN && echo
    vault kv put "secret/apps/${APP_SLUG}/github" token="$GH_TOKEN" > /dev/null
    unset GH_TOKEN
    echo "    DONE"
  fi
  echo ""
  if vault kv get "secret/apps/${APP_SLUG}/ssh" > /dev/null 2>&1; then
    echo "==> secret/apps/${APP_SLUG}/ssh already populated. Vault ready for local deployment."
  else
    echo "==> secret/apps/${APP_SLUG}/ssh not yet populated — run"
    echo "    bootstrap-app-to-machine.sh ${APP_SLUG} ${MACHINE_SLUG:-<machine-slug>} to get it."
  fi
  exit 0
fi

echo "==> Scheme: legacy (secret/${PROJECT_SLUG}/*)"
echo "==> Provisioning mode: $PROVISIONING_MODE"
echo ""

# ── Helper: check if path already populated ───────────────────────────────────

path_exists() {
  vault kv get "secret/${PROJECT_SLUG}/$1" > /dev/null 2>&1
}

if [ "$PROVISIONING_MODE" = "existing-shared" ]; then
  REQUIRED_PATHS=(ssh github)
  TOTAL_STEPS=2
else
  REQUIRED_PATHS=(cyso ssh tfstate github)
  TOTAL_STEPS=4
fi

# ── Step: Cyso app credentials ────────────────────────────────────────────────

if [ "$PROVISIONING_MODE" = "existing-shared" ]; then
  echo "--- Cyso Cloud app credentials (secret/${PROJECT_SLUG}/cyso) ---"
  echo "    SKIP — existing-shared mode never provisions Terraform for this project"
else
  echo "--- [1/${TOTAL_STEPS}] Cyso Cloud app credentials (secret/${PROJECT_SLUG}/cyso) ---"
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
fi

# ── Step: SSH deploy keypair ──────────────────────────────────────────────────

SSH_STEP_LABEL="[2/${TOTAL_STEPS}]"
[ "$PROVISIONING_MODE" = "existing-shared" ] && SSH_STEP_LABEL="[1/${TOTAL_STEPS}]"

echo "--- ${SSH_STEP_LABEL} SSH deploy keypair (secret/${PROJECT_SLUG}/ssh) ---"
if path_exists "ssh"; then
  echo "    SKIP — path already populated"
elif [ "$PROVISIONING_MODE" = "existing-shared" ]; then
  echo "  Targeting a VM this project doesn't own — choose how to get SSH access:"
  echo "    1) Copy the owning project's keypair into this project's Vault namespace"
  echo "    2) Generate a new keypair (you will manually append the public key to"
  echo "       the shared VM's authorized_keys yourself — this script cannot do that"
  echo "       part, since it has no access to the shared VM)"
  read -p "  Choice [1/2]: " SSH_CHOICE
  if [ "$SSH_CHOICE" = "1" ]; then
    read -p "  Source project slug (owns the keypair already in Vault): " SOURCE_SLUG
    if ! vault kv get "secret/${SOURCE_SLUG}/ssh" > /dev/null 2>&1; then
      echo "ERROR: secret/${SOURCE_SLUG}/ssh not found — cannot copy."
      exit 1
    fi
    SRC_PUB=$(vault kv get -field=public_key "secret/${SOURCE_SLUG}/ssh")
    # Fetch private_key to a file and re-append trailing newline before storing via @file —
    # command substitution ("$(...)") on either side of this copy strips the newline OpenSSH's
    # key parser requires (see "error in libcrypto" Gotcha).
    SRC_PRIV_TMP=$(mktemp)
    vault kv get -field=private_key "secret/${SOURCE_SLUG}/ssh" > "$SRC_PRIV_TMP"
    printf '\n' >> "$SRC_PRIV_TMP"
    vault kv put "secret/${PROJECT_SLUG}/ssh" \
      private_key=@"$SRC_PRIV_TMP" \
      public_key="$SRC_PUB" > /dev/null
    rm -f "$SRC_PRIV_TMP"
    unset SRC_PUB
    echo "    DONE — copied secret/${SOURCE_SLUG}/ssh into secret/${PROJECT_SLUG}/ssh"
  else
    TMPDIR_KEY=$(mktemp -d)
    SSH_KEY_PATH="${TMPDIR_KEY}/deploy_key"
    ssh-keygen -t ed25519 -f "$SSH_KEY_PATH" -N "" -q
    # Use @file syntax (not "$(cat ...)") for private_key — see "error in libcrypto" Gotcha.
    vault kv put "secret/${PROJECT_SLUG}/ssh" \
      private_key=@"${SSH_KEY_PATH}" \
      public_key="$(cat "${SSH_KEY_PATH}.pub" | tr -d '\n')" > /dev/null
    echo "    DONE — ed25519 keypair generated and stored"
    echo "    ACTION REQUIRED: append the following public key to the shared VM's"
    echo "    ~/.ssh/authorized_keys yourself — this script has no access to that VM:"
    echo ""
    cat "${SSH_KEY_PATH}.pub"
    echo ""
    rm -rf "$TMPDIR_KEY"
  fi
else
  TMPDIR_KEY=$(mktemp -d)
  SSH_KEY_PATH="${TMPDIR_KEY}/deploy_key"
  ssh-keygen -t ed25519 -f "$SSH_KEY_PATH" -N "" -q
  # Use @file syntax (not "$(cat ...)") for private_key — command substitution strips the
  # trailing newline that OpenSSH's key parser requires; without it, ssh/ssh-keygen fail with
  # "error in libcrypto" on every subsequent fetch from Vault. @file preserves exact file bytes.
  vault kv put "secret/${PROJECT_SLUG}/ssh" \
    private_key=@"${SSH_KEY_PATH}" \
    public_key="$(cat "${SSH_KEY_PATH}.pub" | tr -d '\n')" > /dev/null
  rm -rf "$TMPDIR_KEY"
  echo "    DONE — ed25519 keypair generated and stored"
fi

# ── Step: Terraform state S3 credentials ──────────────────────────────────────

if [ "$PROVISIONING_MODE" = "existing-shared" ]; then
  echo "--- Terraform state S3 credentials (secret/${PROJECT_SLUG}/tfstate) ---"
  echo "    SKIP — existing-shared mode has no Terraform state for this project"
else
  echo "--- [3/${TOTAL_STEPS}] Terraform state S3 credentials (secret/${PROJECT_SLUG}/tfstate) ---"
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
fi

# ── Step: GitHub token ─────────────────────────────────────────────────────────

echo "--- [${TOTAL_STEPS}/${TOTAL_STEPS}] GitHub fine-grained PAT (secret/${PROJECT_SLUG}/github) ---"
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
echo "==> Verifying required paths for mode '${PROVISIONING_MODE}'..."
ALL_OK=true
for PATH_NAME in "${REQUIRED_PATHS[@]}"; do
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
