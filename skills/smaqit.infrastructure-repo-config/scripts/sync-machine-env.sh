#!/usr/bin/env bash
# smaqit.infrastructure-repo-config — sync-machine operation: configure one machine's GitHub
# Environment in a repository that owns machines (a platform repository).
#
# Creates or verifies the Environment named after the machine slug, preflights the token with
# check-pat-permissions.sh, then syncs secret/machines/<machine-slug>/* into that Environment:
#   secrets   BASE_SSH_PUBLIC_KEY, BASE_SSH_PRIVATE_KEY            ← base-ssh
#             CYSO_APPLICATION_CREDENTIAL_ID/_SECRET               ← cyso
#             CYSO_S3_ACCESS_KEY, CYSO_S3_SECRET_KEY               ← tfstate
#   variables RECONCILE_SSH_USER                                   ← --ssh-user (default ubuntu)
#             FIXED_IP                                             ← metadata.host, once known
# Secret values travel from Vault to gh on stdin — never as an argument, never printed.
# Vault is the source of truth: re-run after any rotation. Idempotent.
# Runs before provisioning (the environment-config slot precedes provision), when the machine has
# no address yet: FIXED_IP is then skipped with a notice. Re-run after apply, once
# register-machine.sh --host has backfilled metadata.host, to set it.
#
# Usage: bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-repo-config/scripts/sync-machine-env.sh \
#          <owner>/<repo> <machine-slug> [--ssh-user <user>]

set -euo pipefail

export VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

usage() {
  echo "Usage: $0 <owner>/<repo> <machine-slug> [--ssh-user <user>]" >&2
  exit 1
}

REPO="${1:-}"
MACHINE_SLUG="${2:-}"
[ -n "$REPO" ] && [ -n "$MACHINE_SLUG" ] || usage
shift 2
SSH_USER="ubuntu"
while [ "$#" -gt 0 ]; do
  case "$1" in
    --ssh-user) [ "$#" -ge 2 ] || usage; SSH_USER="$2"; shift 2 ;;
    *) usage ;;
  esac
done

case "$REPO" in */*) ;; *) usage ;; esac
# The slug is used verbatim as the Environment name in API paths and --env; gh does not
# URL-encode Environment names, so allow only characters that never need encoding.
if ! printf '%s' "$MACHINE_SLUG" | grep -Eq '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$'; then
  echo "ERROR: machine slug '$MACHINE_SLUG' must be lowercase letters, digits, and inner hyphens." >&2
  exit 1
fi

if ! vault token lookup > /dev/null 2>&1; then
  echo "ERROR: No valid Vault token. Run smaqit.infrastructure-vault-loader first." >&2
  exit 1
fi
if ! gh auth status > /dev/null 2>&1; then
  echo "ERROR: gh is not authenticated. Run 'gh auth login' (or export GH_TOKEN) first." >&2
  exit 1
fi

MACHINE_PATH="secret/machines/${MACHINE_SLUG}"
REGISTER="smaqit.infrastructure-vault-loader/scripts/register-machine.sh ${MACHINE_SLUG}"

# ── 1. Every source path must exist before anything is written ─────────────────

MISSING=0
for sub in base-ssh cyso tfstate metadata; do
  if ! vault kv get "${MACHINE_PATH}/${sub}" > /dev/null 2>&1; then
    echo "ERROR: ${MACHINE_PATH}/${sub} is missing." >&2
    MISSING=1
  fi
done
if [ "$MISSING" -ne 0 ]; then
  echo "       Register the machine first: ${REGISTER}" >&2
  exit 1
fi
FIXED_IP=$(vault kv get -field=host "${MACHINE_PATH}/metadata" 2>/dev/null || true)

# ── 2. Create or verify the Environment ────────────────────────────────────────

echo "==> Ensuring Environment '${MACHINE_SLUG}' exists on ${REPO}..."
if ! CREATE_OUT=$(gh api -X PUT "repos/${REPO}/environments/${MACHINE_SLUG}" --silent 2>&1); then
  echo "ERROR: could not create or update Environment '${MACHINE_SLUG}': ${CREATE_OUT}" >&2
  case "$CREATE_OUT" in
    *403*|*"Resource not accessible"*)
      echo "       403 on Environment create: the token lacks Environments write, or organization-level" >&2
      echo "       Secrets and Variables read/write. Fix the token, re-authenticate gh, and re-run." >&2
      ;;
  esac
  exit 3
fi

# ── 3. Token preflight (read-only probes of the Environment's secrets and variables) ─

bash "${SCRIPT_DIR}/check-pat-permissions.sh" "$REPO" "$MACHINE_SLUG"

# ── 4. Secrets: Vault → gh on stdin ────────────────────────────────────────────

sync_secret() {
  local name="$1" path="$2" field="$3"
  vault kv get -field="$field" "${MACHINE_PATH}/${path}" \
    | gh secret set "$name" --env "$MACHINE_SLUG" -R "$REPO" > /dev/null
  echo "    secret   ${name}"
}

echo "==> Syncing ${MACHINE_PATH}/* into Environment '${MACHINE_SLUG}'..."
sync_secret BASE_SSH_PUBLIC_KEY                base-ssh public_key
sync_secret BASE_SSH_PRIVATE_KEY               base-ssh private_key
sync_secret CYSO_APPLICATION_CREDENTIAL_ID     cyso     app_credential_id
sync_secret CYSO_APPLICATION_CREDENTIAL_SECRET cyso     app_credential_secret
sync_secret CYSO_S3_ACCESS_KEY                 tfstate  access_key
sync_secret CYSO_S3_SECRET_KEY                 tfstate  secret_key

# ── 5. Variables (non-secret) ──────────────────────────────────────────────────

gh variable set RECONCILE_SSH_USER --env "$MACHINE_SLUG" -R "$REPO" --body "$SSH_USER" > /dev/null
echo "    variable RECONCILE_SSH_USER"
if [ -n "$FIXED_IP" ]; then
  gh variable set FIXED_IP --env "$MACHINE_SLUG" -R "$REPO" --body "$FIXED_IP" > /dev/null
  echo "    variable FIXED_IP"
else
  echo "    variable FIXED_IP skipped — ${MACHINE_PATH}/metadata has no 'host' yet (machine not"
  echo "             provisioned). After apply: ${REGISTER} --host <fixed_ip>, then re-run this script."
fi

# ── 6. Verify every name is present ────────────────────────────────────────────

SECRETS=$(gh secret list --env "$MACHINE_SLUG" -R "$REPO" --json name --jq '.[].name')
VARIABLES=$(gh variable list --env "$MACHINE_SLUG" -R "$REPO" --json name --jq '.[].name')
ABSENT=0
for name in BASE_SSH_PUBLIC_KEY BASE_SSH_PRIVATE_KEY CYSO_APPLICATION_CREDENTIAL_ID \
            CYSO_APPLICATION_CREDENTIAL_SECRET CYSO_S3_ACCESS_KEY CYSO_S3_SECRET_KEY; do
  printf '%s\n' "$SECRETS" | grep -qx "$name" || { echo "ERROR: secret ${name} not listed on '${MACHINE_SLUG}'." >&2; ABSENT=1; }
done
EXPECTED_VARIABLES="RECONCILE_SSH_USER"
[ -n "$FIXED_IP" ] && EXPECTED_VARIABLES="$EXPECTED_VARIABLES FIXED_IP"
for name in $EXPECTED_VARIABLES; do
  printf '%s\n' "$VARIABLES" | grep -qx "$name" || { echo "ERROR: variable ${name} not listed on '${MACHINE_SLUG}'." >&2; ABSENT=1; }
done
[ "$ABSENT" -eq 0 ] || exit 1

echo ""
if [ -n "$FIXED_IP" ]; then
  echo "==> Environment '${MACHINE_SLUG}' on ${REPO}: 6 secrets and 2 variables in place."
else
  echo "==> Environment '${MACHINE_SLUG}' on ${REPO}: 6 secrets and RECONCILE_SSH_USER in place; FIXED_IP pending."
fi
