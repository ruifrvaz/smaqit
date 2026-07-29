#!/usr/bin/env bash
# smaqit.infrastructure-vault-loader — bootstrap an app's own SSH keypair onto a machine.
# Every app gets its own distinct keypair, authorized against the machine's base-ssh credential —
# no exceptions, including the project that originally provisioned the machine. Idempotent: no-ops
# if the app's keypair is already populated and still authenticates.
# All sensitive input/output uses hidden prompts and piped/file-based handling — nothing sensitive
# is ever passed as a command argument or written to shell history.
#
# Usage: bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/bootstrap-app-to-machine.sh <app-slug> <machine-slug>

set -euo pipefail

export VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
SSH_REMOTE_USER="${SSH_REMOTE_USER:-ubuntu}"

APP_SLUG="${1:-}"
MACHINE_SLUG="${2:-}"

if [ -z "$APP_SLUG" ] || [ -z "$MACHINE_SLUG" ]; then
  echo "Usage: $0 <app-slug> <machine-slug>"
  exit 1
fi

if [ "$APP_SLUG" = "machines" ]; then
  echo "ERROR: 'machines' is a reserved app-slug."
  exit 1
fi

if ! vault token lookup > /dev/null 2>&1; then
  echo "ERROR: No valid Vault token. Run load-credentials.sh first."
  exit 1
fi

APP_PATH="secret/apps/${APP_SLUG}"
MACHINE_PATH="secret/machines/${MACHINE_SLUG}"

# ── Helper: read a secret without echoing ──────────────────────────────────────
# Usage: read_secret VAR_NAME "Prompt label"
read_secret() {
  local _var="$1"
  local _prompt="$2"
  local _value
  IFS= read -rs -p "  ${_prompt}: " _value </dev/tty && echo >&2
  printf -v "$_var" '%s' "$_value"
  unset _value
}

# ── Idempotency check: already bootstrapped and still reachable? ──────────────

if vault kv get "${APP_PATH}/ssh" > /dev/null 2>&1; then
  echo "==> ${APP_PATH}/ssh already populated — verifying it still authenticates..."
  EXISTING_HOST=$(vault kv get -field=host "${MACHINE_PATH}/metadata" 2>/dev/null || true)
  if [ -z "$EXISTING_HOST" ]; then
    echo "    WARNING: ${MACHINE_PATH}/metadata missing 'host' — cannot verify connectivity."
    echo "    Treating existing key as already bootstrapped. No-op."
    exit 0
  fi
  TMPDIR_CHECK=$(mktemp -d)
  trap 'rm -rf "$TMPDIR_CHECK"' EXIT
  vault kv get -field=private_key "${APP_PATH}/ssh" > "${TMPDIR_CHECK}/app_key"
  chmod 600 "${TMPDIR_CHECK}/app_key"
  if ssh -i "${TMPDIR_CHECK}/app_key" -o BatchMode=yes -o StrictHostKeyChecking=accept-new \
      -o ConnectTimeout=5 "${SSH_REMOTE_USER}@${EXISTING_HOST}" 'true' 2>/dev/null; then
    echo "    OK — already bootstrapped and reachable. No-op."
    rm -rf "$TMPDIR_CHECK"
    trap - EXIT
    exit 0
  fi
  echo "    Key stored but connection failed — continuing to re-authorize."
  rm -rf "$TMPDIR_CHECK"
  trap - EXIT
fi

# ── Step 1: resolve the machine, registering it if this is the first time ─────

if ! vault kv get "${MACHINE_PATH}/base-ssh" > /dev/null 2>&1; then
  echo "==> ${MACHINE_PATH}/base-ssh not found — registering a new machine record."
  read -r -p "  Machine host (IP or hostname): " MACHINE_HOST_INPUT
  read -r -p "  Cloud provider (e.g. cyso): " MACHINE_PROVIDER_INPUT
  read -r -p "  Owner project slug (whose Terraform state provisions this machine, or the requesting project's own slug if no Terraform manages it — existing-unmanaged): " MACHINE_OWNER_INPUT

  TMPDIR_BASE=$(mktemp -d)
  trap 'rm -rf "$TMPDIR_BASE"' EXIT
  ssh-keygen -t ed25519 -f "${TMPDIR_BASE}/base_key" -N "" -q
  vault kv put "${MACHINE_PATH}/base-ssh" \
    private_key=@"${TMPDIR_BASE}/base_key" \
    public_key="$(cat "${TMPDIR_BASE}/base_key.pub" | tr -d '\n')" > /dev/null
  vault kv put "${MACHINE_PATH}/metadata" \
    host="$MACHINE_HOST_INPUT" \
    provider="$MACHINE_PROVIDER_INPUT" \
    owner_project="$MACHINE_OWNER_INPUT" > /dev/null

  echo "    DONE — machine registered."
  echo ""
  echo "    NOTE: this base-ssh keypair was generated here, not installed by Terraform at"
  echo "    provision time — it is NOT YET authorized on ${MACHINE_HOST_INPUT}. Install the"
  echo "    following public key on that host's ~/.ssh/authorized_keys before continuing"
  echo "    (this script has no prior access to the machine and cannot do that part):"
  echo ""
  cat "${TMPDIR_BASE}/base_key.pub"
  echo ""
  rm -rf "$TMPDIR_BASE"
  trap - EXIT
  echo "==> Re-run this script once the base key is confirmed authorized on the machine."
  exit 0
fi

MACHINE_HOST=$(vault kv get -field=host "${MACHINE_PATH}/metadata")
if [ -z "$MACHINE_HOST" ]; then
  echo "ERROR: ${MACHINE_PATH}/metadata has no 'host' field. Cannot bootstrap."
  exit 1
fi

# ── Step 2: generate the app's own distinct keypair ────────────────────────────

echo "==> Generating a new keypair for '${APP_SLUG}'..."
TMPDIR_APP=$(mktemp -d)
trap 'rm -rf "$TMPDIR_APP" "${TMPDIR_BASE:-}"' EXIT
ssh-keygen -t ed25519 -f "${TMPDIR_APP}/app_key" -N "" -q

# ── Step 3: use the machine's base credential to authorize the new app key ────

echo "==> Authorizing new key on ${MACHINE_HOST} via the machine's base credential..."
TMPDIR_BASE=$(mktemp -d)
vault kv get -field=private_key "${MACHINE_PATH}/base-ssh" > "${TMPDIR_BASE}/base_key"
chmod 600 "${TMPDIR_BASE}/base_key"

cat "${TMPDIR_APP}/app_key.pub" | ssh -i "${TMPDIR_BASE}/base_key" \
  -o StrictHostKeyChecking=accept-new -o ConnectTimeout=10 \
  "${SSH_REMOTE_USER}@${MACHINE_HOST}" \
  'mkdir -p ~/.ssh && chmod 700 ~/.ssh && cat >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys'

rm -f "${TMPDIR_BASE}/base_key"

# ── Step 4: store the app's new keypair ─────────────────────────────────────────

vault kv put "${APP_PATH}/ssh" \
  private_key=@"${TMPDIR_APP}/app_key" \
  public_key="$(cat "${TMPDIR_APP}/app_key.pub" | tr -d '\n')" > /dev/null
vault kv put "${APP_PATH}/machine" \
  machine-slug="$MACHINE_SLUG" > /dev/null

# ── Step 5: verify the new key actually authenticates ──────────────────────────

echo "==> Verifying new key authenticates..."
if ssh -i "${TMPDIR_APP}/app_key" -o BatchMode=yes -o StrictHostKeyChecking=accept-new \
    -o ConnectTimeout=5 "${SSH_REMOTE_USER}@${MACHINE_HOST}" 'true' 2>/dev/null; then
  echo "    OK — '${APP_SLUG}' bootstrapped onto '${MACHINE_SLUG}'."
else
  echo "ERROR: New key was installed but failed to authenticate. Check ${MACHINE_HOST}'s"
  echo "       authorized_keys and sshd configuration."
  exit 1
fi

rm -rf "$TMPDIR_APP" "$TMPDIR_BASE"
trap - EXIT
