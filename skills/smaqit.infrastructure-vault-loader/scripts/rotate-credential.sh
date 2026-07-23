#!/usr/bin/env bash
# Rotate a single Vault credential path for a smaqit project, app, or machine.
# All sensitive input is via hidden prompts — nothing enters shell history.
#
# Usage: bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/rotate-credential.sh <path>
#
#   Legacy flat scheme (secret/<project-slug>/*):
#     cyso | ssh | tfstate | github
#
#   New scheme:
#     apps/<app-slug>/ssh | apps/<app-slug>/github
#     machines/<machine-slug>/cyso | machines/<machine-slug>/tfstate | machines/<machine-slug>/base-ssh
#
# For everything except machines/<slug>/base-ssh, the path is deleted and re-populated (via
# load-credentials.sh for the legacy scheme; directly here for the new scheme, since
# load-credentials.sh never writes ssh under the new scheme). base-ssh is different: it generates a
# new base keypair, uses the OLD one one last time to install the new public key on the machine,
# then retires the old one — it's the one credential type that has to install itself onto the
# remote machine, not just be replaced in Vault. This never touches any already-bootstrapped app's
# secret/apps/<app-slug>/ssh.
#
# Example:
#   bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/rotate-credential.sh cyso
#   bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/rotate-credential.sh apps/<app-slug>/ssh
#   bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/rotate-credential.sh machines/<machine-slug>/base-ssh

set -euo pipefail

export VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
SSH_REMOTE_USER="${SSH_REMOTE_USER:-ubuntu}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

CREDENTIAL_PATH="${1:-}"
if [ -z "$CREDENTIAL_PATH" ]; then
  echo "Usage: $0 <path>"
  echo "  Legacy: cyso | ssh | tfstate | github"
  echo "  New:    apps/<app-slug>/{ssh,github} | machines/<machine-slug>/{cyso,tfstate,base-ssh}"
  exit 1
fi

# ── Classify the target: legacy | apps | machines ──────────────────────────────

case "$CREDENTIAL_PATH" in
  cyso|ssh|tfstate|github)
    SCHEME="legacy"
    ;;
  apps/*/ssh|apps/*/github)
    SCHEME="apps"
    APP_SLUG="$(echo "$CREDENTIAL_PATH" | cut -d/ -f2)"
    CRED_TYPE="$(echo "$CREDENTIAL_PATH" | cut -d/ -f3)"
    ;;
  machines/*/cyso|machines/*/tfstate|machines/*/base-ssh)
    SCHEME="machines"
    MACHINE_SLUG="$(echo "$CREDENTIAL_PATH" | cut -d/ -f2)"
    CRED_TYPE="$(echo "$CREDENTIAL_PATH" | cut -d/ -f3)"
    ;;
  *)
    echo "ERROR: Unrecognized path '$CREDENTIAL_PATH'."
    echo "  Legacy: cyso | ssh | tfstate | github"
    echo "  New:    apps/<app-slug>/{ssh,github} | machines/<machine-slug>/{cyso,tfstate,base-ssh}"
    exit 1
    ;;
esac

# ── Legacy scheme: unchanged behavior ───────────────────────────────────────────

if [ "$SCHEME" = "legacy" ]; then
  INSTRUCTIONS_FILE="AGENTS.md"
  if [ ! -f "$INSTRUCTIONS_FILE" ]; then
    INSTRUCTIONS_FILE="CLAUDE.md"
  fi
  if [ ! -f "$INSTRUCTIONS_FILE" ]; then
    INSTRUCTIONS_FILE=".github/copilot-instructions.md"
  fi
  if [ ! -f "$INSTRUCTIONS_FILE" ]; then
    echo "ERROR: Run from repo root (no AGENTS.md, CLAUDE.md, or .github/copilot-instructions.md found)"
    exit 1
  fi

  PROJECT_SLUG=$(grep -i "project name" "$INSTRUCTIONS_FILE" | grep -i ": " | head -1 \
    | sed 's/.*: *//' | sed 's/[^a-zA-Z0-9 -].*$//' \
    | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | tr -s '-' || true)

  if [ -z "$PROJECT_SLUG" ]; then
    read -r -p "Enter project slug manually: " PROJECT_SLUG
  fi

  FULL_PATH="secret/${PROJECT_SLUG}/${CREDENTIAL_PATH}"

  echo "==> About to delete and re-populate: $FULL_PATH"
  read -r -p "    Continue? [y/N] " CONFIRM
  case "$CONFIRM" in
    y|Y) ;;
    *) echo "Aborted."; exit 0 ;;
  esac

  if vault kv get "$FULL_PATH" > /dev/null 2>&1; then
    vault kv delete "$FULL_PATH"
    echo "    Deleted $FULL_PATH"
  else
    echo "    Path does not exist — will be created fresh"
  fi

  echo "==> Running credential loader to repopulate $CREDENTIAL_PATH..."
  bash "${SCRIPT_DIR}/load-credentials.sh"

  echo ""
  echo "==> Rotation complete for $FULL_PATH"
  echo "    Re-run smaqit.infrastructure-repo-config to sync new value to GitHub Secrets."
  exit 0
fi

# ── New scheme: apps/<app-slug>/{ssh,github} ────────────────────────────────────

if [ "$SCHEME" = "apps" ]; then
  FULL_PATH="secret/apps/${APP_SLUG}/${CRED_TYPE}"
  echo "==> About to delete and re-populate: $FULL_PATH"
  read -r -p "    Continue? [y/N] " CONFIRM
  case "$CONFIRM" in
    y|Y) ;;
    *) echo "Aborted."; exit 0 ;;
  esac

  if vault kv get "$FULL_PATH" > /dev/null 2>&1; then
    vault kv delete "$FULL_PATH"
    echo "    Deleted $FULL_PATH"
  fi

  if [ "$CRED_TYPE" = "ssh" ]; then
    MACHINE_SLUG=$(vault kv get -field=machine-slug "secret/apps/${APP_SLUG}/machine" 2>/dev/null || true)
    if [ -z "$MACHINE_SLUG" ]; then
      echo "ERROR: secret/apps/${APP_SLUG}/machine not found — cannot determine which machine to"
      echo "       re-authorize against. Run bootstrap-app-to-machine.sh directly instead."
      exit 1
    fi
    echo "==> Re-running bootstrap against machine '${MACHINE_SLUG}'..."
    bash "${SCRIPT_DIR}/bootstrap-app-to-machine.sh" "$APP_SLUG" "$MACHINE_SLUG"
  else
    read -s -p "  github_token: " GH_TOKEN && echo
    vault kv put "$FULL_PATH" token="$GH_TOKEN" > /dev/null
    unset GH_TOKEN
    echo "    DONE"
  fi

  echo ""
  echo "==> Rotation complete for $FULL_PATH"
  echo "    Re-run smaqit.infrastructure-repo-config to sync new value to GitHub Secrets."
  exit 0
fi

# ── New scheme: machines/<machine-slug>/{cyso,tfstate,base-ssh} ────────────────

if [ "$CRED_TYPE" = "cyso" ] || [ "$CRED_TYPE" = "tfstate" ]; then
  FULL_PATH="secret/machines/${MACHINE_SLUG}/${CRED_TYPE}"
  echo "==> About to delete and re-populate: $FULL_PATH"
  read -r -p "    Continue? [y/N] " CONFIRM
  case "$CONFIRM" in
    y|Y) ;;
    *) echo "Aborted."; exit 0 ;;
  esac

  if vault kv get "$FULL_PATH" > /dev/null 2>&1; then
    vault kv delete "$FULL_PATH"
    echo "    Deleted $FULL_PATH"
  fi

  if [ "$CRED_TYPE" = "cyso" ]; then
    read -r -p "  app_credential_id: " CYSO_ID
    read -rs -p "  app_credential_secret: " CYSO_SECRET && echo
    vault kv put "$FULL_PATH" \
      app_credential_id="$CYSO_ID" \
      app_credential_secret="$CYSO_SECRET" > /dev/null
    unset CYSO_ID CYSO_SECRET
  else
    read -r -p "  access_key: " TF_KEY
    read -rs -p "  secret_key: " TF_SECRET && echo
    vault kv put "$FULL_PATH" \
      access_key="$TF_KEY" \
      secret_key="$TF_SECRET" > /dev/null
    unset TF_KEY TF_SECRET
  fi
  echo "    DONE"
  echo ""
  echo "==> Rotation complete for $FULL_PATH"
  exit 0
fi

# ── machines/<machine-slug>/base-ssh: bespoke generate/install-with-old/retire ──

MACHINE_PATH="secret/machines/${MACHINE_SLUG}"
if ! vault kv get "${MACHINE_PATH}/base-ssh" > /dev/null 2>&1; then
  echo "ERROR: ${MACHINE_PATH}/base-ssh not found — nothing to rotate. Use"
  echo "       bootstrap-app-to-machine.sh to register this machine first."
  exit 1
fi
MACHINE_HOST=$(vault kv get -field=host "${MACHINE_PATH}/metadata" 2>/dev/null || true)
if [ -z "$MACHINE_HOST" ]; then
  echo "ERROR: ${MACHINE_PATH}/metadata has no 'host' field. Cannot rotate base-ssh."
  exit 1
fi

echo "==> About to rotate ${MACHINE_PATH}/base-ssh on ${MACHINE_HOST}."
echo "    This does NOT require any already-bootstrapped app's key to change."
read -r -p "    Continue? [y/N] " CONFIRM
case "$CONFIRM" in
  y|Y) ;;
  *) echo "Aborted."; exit 0 ;;
esac

TMPDIR_ROT=$(mktemp -d)
trap 'rm -rf "$TMPDIR_ROT"' EXIT

# Fetch the OLD key — it's the only thing currently authorized on the machine.
vault kv get -field=private_key "${MACHINE_PATH}/base-ssh" > "${TMPDIR_ROT}/old_key"
chmod 600 "${TMPDIR_ROT}/old_key"

# Generate the NEW base keypair.
ssh-keygen -t ed25519 -f "${TMPDIR_ROT}/new_key" -N "" -q

echo "==> Installing new base key via the old one..."
cat "${TMPDIR_ROT}/new_key.pub" | ssh -i "${TMPDIR_ROT}/old_key" \
  -o StrictHostKeyChecking=accept-new -o ConnectTimeout=10 \
  "${SSH_REMOTE_USER}@${MACHINE_HOST}" \
  'mkdir -p ~/.ssh && chmod 700 ~/.ssh && cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys'

echo "==> Verifying new base key authenticates..."
if ! ssh -i "${TMPDIR_ROT}/new_key" -o BatchMode=yes -o StrictHostKeyChecking=accept-new \
    -o ConnectTimeout=5 "${SSH_REMOTE_USER}@${MACHINE_HOST}" 'true' 2>/dev/null; then
  echo "ERROR: New base key was installed but failed to authenticate. Old key left in place."
  exit 1
fi

vault kv put "${MACHINE_PATH}/base-ssh" \
  private_key=@"${TMPDIR_ROT}/new_key" \
  public_key="$(cat "${TMPDIR_ROT}/new_key.pub" | tr -d '\n')" > /dev/null

echo "==> Retiring old base key from authorized_keys..."
ssh-keygen -y -f "${TMPDIR_ROT}/old_key" > "${TMPDIR_ROT}/old_key.pub"
cat "${TMPDIR_ROT}/old_key.pub" | ssh -i "${TMPDIR_ROT}/new_key" \
  -o StrictHostKeyChecking=accept-new -o ConnectTimeout=10 \
  "${SSH_REMOTE_USER}@${MACHINE_HOST}" \
  'grep -vFf - ~/.ssh/authorized_keys > ~/.ssh/authorized_keys.tmp && mv ~/.ssh/authorized_keys.tmp ~/.ssh/authorized_keys'

rm -rf "$TMPDIR_ROT"
trap - EXIT

echo ""
echo "==> Rotation complete for ${MACHINE_PATH}/base-ssh."
echo "    No already-bootstrapped app's secret/apps/<app-slug>/ssh was touched."
