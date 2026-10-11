#!/usr/bin/env bash
# smaqit.infrastructure-vault-loader — register a machine's credentials in Vault.
# Populates secret/machines/<machine-slug>/{base-ssh,cyso,tfstate,metadata}. This is the one
# entrypoint for first-time machine population; rotate-credential.sh only rotates paths that
# already exist. Idempotent: every path that already exists is kept, so a re-run only fills gaps
# or, with --host, backfills metadata.host after the machine's IP is known.
# The machine slug is a required argument — never derived from the working directory.
# All sensitive input uses hidden prompts on /dev/tty and is handed to Vault on stdin, so no
# secret is ever passed as a command argument, printed, or written to shell history.
#
# Usage:
#   bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/register-machine.sh \
#     <machine-slug> [--host <ip-or-hostname>] [--provider <name>] [--owner-project <slug>]
#
#   Before provisioning: register-machine.sh <machine-slug>
#     (base-ssh generated if missing; Terraform installs its public half at apply time)
#   After apply:         register-machine.sh <machine-slug> --host <fixed_ip>

set -euo pipefail

export VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# shellcheck source=lib-project-slug.sh
source "${SCRIPT_DIR}/lib-project-slug.sh"

usage() {
  echo "Usage: $0 <machine-slug> [--host <ip-or-hostname>] [--provider <name>] [--owner-project <slug>]" >&2
  exit 1
}

MACHINE_SLUG="${1:-}"
[ -n "$MACHINE_SLUG" ] || usage
case "$MACHINE_SLUG" in -*) usage ;; esac
shift

HOST_ARG=""
PROVIDER_ARG=""
OWNER_ARG=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    --host)          [ "$#" -ge 2 ] || usage; HOST_ARG="$2"; shift 2 ;;
    --provider)      [ "$#" -ge 2 ] || usage; PROVIDER_ARG="$2"; shift 2 ;;
    --owner-project) [ "$#" -ge 2 ] || usage; OWNER_ARG="$2"; shift 2 ;;
    *) usage ;;
  esac
done

# Machine slugs also name GitHub Environments (smaqit.infrastructure-repo-config sync-machine),
# so keep them to characters that need no URL encoding anywhere.
if ! printf '%s' "$MACHINE_SLUG" | grep -Eq '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$'; then
  echo "ERROR: machine slug '$MACHINE_SLUG' must be lowercase letters, digits, and inner hyphens." >&2
  exit 1
fi

if ! vault token lookup > /dev/null 2>&1; then
  echo "ERROR: No valid Vault token. Run load-credentials.sh first." >&2
  exit 1
fi

MACHINE_PATH="secret/machines/${MACHINE_SLUG}"

# ── Helpers ────────────────────────────────────────────────────────────────────
# read_secret VAR_NAME "Prompt label" — hidden read from the terminal, never from stdin.
read_secret() {
  local _var="$1"
  local _prompt="$2"
  local _value
  IFS= read -rs -p "  ${_prompt}: " _value </dev/tty && echo >&2
  printf -v "$_var" '%s' "$_value"
  unset _value
}

# read_plain VAR_NAME "Prompt label" ["default"] — visible read from the terminal.
read_plain() {
  local _var="$1"
  local _prompt="$2"
  local _default="${3:-}"
  local _value
  if [ -n "$_default" ]; then
    IFS= read -r -p "  ${_prompt} [${_default}]: " _value </dev/tty
  else
    IFS= read -r -p "  ${_prompt}: " _value </dev/tty
  fi
  printf -v "$_var" '%s' "${_value:-$_default}"
  unset _value
}

path_exists() {
  vault kv get "$1" > /dev/null 2>&1
}

# ── base-ssh ───────────────────────────────────────────────────────────────────

if path_exists "${MACHINE_PATH}/base-ssh"; then
  echo "==> ${MACHINE_PATH}/base-ssh present — kept."
else
  echo "==> Generating ${MACHINE_PATH}/base-ssh..."
  TMPDIR_BASE=$(mktemp -d)
  trap 'rm -rf "$TMPDIR_BASE"' EXIT
  ssh-keygen -t ed25519 -f "${TMPDIR_BASE}/base_key" -N "" -q -C "${MACHINE_SLUG}-base"
  vault kv put "${MACHINE_PATH}/base-ssh" \
    private_key=@"${TMPDIR_BASE}/base_key" \
    public_key="$(tr -d '\n' < "${TMPDIR_BASE}/base_key.pub")" > /dev/null
  rm -rf "$TMPDIR_BASE"
  trap - EXIT
  echo "    DONE — bootstrap-only key. Terraform's keypair resource installs its public half at"
  echo "    apply time; for a machine no Terraform manages, install it in authorized_keys yourself."
fi

# ── cyso (OpenStack application credential) ────────────────────────────────────

if path_exists "${MACHINE_PATH}/cyso"; then
  echo "==> ${MACHINE_PATH}/cyso present — kept (use rotate-credential.sh to replace)."
else
  echo "==> ${MACHINE_PATH}/cyso — Cyso application credential:"
  read_plain CYSO_ID "app_credential_id"
  read_secret CYSO_SECRET "app_credential_secret"
  if [ -z "$CYSO_ID" ] || [ -z "$CYSO_SECRET" ]; then
    echo "ERROR: app_credential_id or app_credential_secret is empty — refusing to write a placeholder secret." >&2
    exit 1
  fi
  printf '%s' "$CYSO_SECRET" | vault kv put "${MACHINE_PATH}/cyso" \
    app_credential_id="$CYSO_ID" app_credential_secret=- > /dev/null
  unset CYSO_ID CYSO_SECRET
  echo "    DONE"
fi

# ── tfstate (Object Storage / S3 credentials for the Terraform backend) ────────

if path_exists "${MACHINE_PATH}/tfstate"; then
  echo "==> ${MACHINE_PATH}/tfstate present — kept (use rotate-credential.sh to replace)."
else
  echo "==> ${MACHINE_PATH}/tfstate — Object Storage (S3) credentials:"
  read_plain TF_KEY "access_key"
  read_secret TF_SECRET "secret_key"
  if [ -z "$TF_KEY" ] || [ -z "$TF_SECRET" ]; then
    echo "ERROR: access_key or secret_key is empty — refusing to write a placeholder secret." >&2
    exit 1
  fi
  printf '%s' "$TF_SECRET" | vault kv put "${MACHINE_PATH}/tfstate" \
    access_key="$TF_KEY" secret_key=- > /dev/null
  unset TF_KEY TF_SECRET
  echo "    DONE"
fi

# ── metadata (non-secret: host, provider, owner_project) ───────────────────────

if path_exists "${MACHINE_PATH}/metadata"; then
  PATCH_ARGS=()
  [ -n "$HOST_ARG" ] && PATCH_ARGS+=("host=${HOST_ARG}")
  [ -n "$PROVIDER_ARG" ] && PATCH_ARGS+=("provider=${PROVIDER_ARG}")
  [ -n "$OWNER_ARG" ] && PATCH_ARGS+=("owner_project=${OWNER_ARG}")
  if [ "${#PATCH_ARGS[@]}" -gt 0 ]; then
    vault kv patch "${MACHINE_PATH}/metadata" "${PATCH_ARGS[@]}" > /dev/null
    echo "==> ${MACHINE_PATH}/metadata updated (${PATCH_ARGS[*]%%=*})."
  else
    echo "==> ${MACHINE_PATH}/metadata present — kept."
  fi
else
  echo "==> ${MACHINE_PATH}/metadata:"
  PROVIDER="$PROVIDER_ARG"
  [ -n "$PROVIDER" ] || read_plain PROVIDER "provider" "cyso"
  OWNER="$OWNER_ARG"
  [ -n "$OWNER" ] || read_plain OWNER "owner_project (repository whose Terraform state owns this machine)" "$(derive_project_slug)"
  if [ -z "$PROVIDER" ] || [ -z "$OWNER" ]; then
    echo "ERROR: provider and owner_project are required for ${MACHINE_PATH}/metadata." >&2
    exit 1
  fi
  META_ARGS=("provider=${PROVIDER}" "owner_project=${OWNER}")
  [ -n "$HOST_ARG" ] && META_ARGS+=("host=${HOST_ARG}")
  vault kv put "${MACHINE_PATH}/metadata" "${META_ARGS[@]}" > /dev/null
  echo "    DONE"
fi

# ── Summary ────────────────────────────────────────────────────────────────────

HOST_NOW=$(vault kv get -field=host "${MACHINE_PATH}/metadata" 2>/dev/null || true)
echo ""
echo "==> Machine '${MACHINE_SLUG}' registered: base-ssh, cyso, tfstate, metadata present."
if [ -z "$HOST_NOW" ]; then
  echo "    metadata.host is not set yet. After apply, run:"
  echo "      $0 ${MACHINE_SLUG} --host <fixed_ip>"
  echo "    (rotate-credential.sh machines/${MACHINE_SLUG}/base-ssh requires it.)"
else
  echo "    metadata.host = ${HOST_NOW}"
fi
