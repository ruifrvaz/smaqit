#!/usr/bin/env bash
# smaqit.infrastructure-repo-config — sync operation for provisioning_mode existing-k3s:
# configure the app's GitHub Environment for one target machine.
#
# The Environment is named after the machine slug — the same slug that keys the app's kubeconfig
# in Vault (secret/apps/<app-slug>/<machine-slug>/kubeconfig) — so each environment's distinct
# scoped kubeconfig has exactly one home. The script creates or verifies the Environment,
# preflights the token with check-pat-permissions.sh, then sets:
#   secret   KUBECONFIG                         ← secret/apps/<app>/<machine>/kubeconfig (value)
#   variable APP_HOST                           ← --host (the environment's Ingress hostname)
#   secrets  REGISTRY_USERNAME, REGISTRY_TOKEN  ← secret/organizations/<org>/github-package-read
#                                                 (only with --registry-org, for a private image)
# Secret values travel from Vault to gh on stdin — never as an argument, never printed.
# Idempotent: re-run after any rotation in Vault.
#
# Usage: bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-repo-config/scripts/sync-k3s-env.sh \
#          <owner>/<repo> <app-slug> <machine-slug> [--host <app-host>] [--registry-org <org>]

set -euo pipefail

export VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

usage() {
  echo "Usage: $0 <owner>/<repo> <app-slug> <machine-slug> [--host <app-host>] [--registry-org <org>]" >&2
  exit 1
}

REPO="${1:-}"
APP_SLUG="${2:-}"
MACHINE_SLUG="${3:-}"
[ -n "$REPO" ] && [ -n "$APP_SLUG" ] && [ -n "$MACHINE_SLUG" ] || usage
shift 3
APP_HOST=""
REGISTRY_ORG=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    --host)         [ "$#" -ge 2 ] || usage; APP_HOST="$2"; shift 2 ;;
    --registry-org) [ "$#" -ge 2 ] || usage; REGISTRY_ORG="$2"; shift 2 ;;
    *) usage ;;
  esac
done

case "$REPO" in */*) ;; *) usage ;; esac
# Both slugs appear verbatim in Vault paths and in the Environment name; gh does not URL-encode
# Environment names, so allow only characters that never need encoding.
for slug in "$APP_SLUG" "$MACHINE_SLUG" ${REGISTRY_ORG:+"$REGISTRY_ORG"}; do
  if ! printf '%s' "$slug" | grep -Eq '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$'; then
    echo "ERROR: '$slug' must be lowercase letters, digits, and inner hyphens." >&2
    exit 1
  fi
done

if ! vault token lookup > /dev/null 2>&1; then
  echo "ERROR: No valid Vault token. Run smaqit.infrastructure-vault-loader first." >&2
  exit 1
fi
if ! gh auth status > /dev/null 2>&1; then
  echo "ERROR: gh is not authenticated. Run 'gh auth login' (or export GH_TOKEN) first." >&2
  exit 1
fi

KUBECONFIG_PATH="secret/apps/${APP_SLUG}/${MACHINE_SLUG}/kubeconfig"
REGISTRY_PATH="secret/organizations/${REGISTRY_ORG}/github-package-read"

# ── 1. Every source path must exist before anything is written ─────────────────

if ! vault kv get -field=value "$KUBECONFIG_PATH" > /dev/null 2>&1; then
  echo "ERROR: ${KUBECONFIG_PATH} is missing. A merged onboarding request is not enough — the" >&2
  echo "       platform's own hand-off must deliver the kubeconfig. Then load it with:" >&2
  echo "       PROVISIONING_MODE=existing-k3s MACHINE_SLUG=${MACHINE_SLUG} load-credentials.sh" >&2
  exit 1
fi
if [ -n "$REGISTRY_ORG" ] && ! vault kv get "$REGISTRY_PATH" > /dev/null 2>&1; then
  echo "ERROR: ${REGISTRY_PATH} is missing. Load it with:" >&2
  echo "       PROVISIONING_MODE=existing-k3s CREDENTIAL_SCOPE=access REGISTRY_ORG=${REGISTRY_ORG} load-credentials.sh" >&2
  exit 1
fi

# ── 2. Create or verify the Environment, then preflight the token ──────────────

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

bash "${SCRIPT_DIR}/check-pat-permissions.sh" "$REPO" "$MACHINE_SLUG"

# ── 3. Secrets and variables ───────────────────────────────────────────────────

sync_secret() {
  local name="$1" path="$2" field="$3"
  vault kv get -field="$field" "$path" | gh secret set "$name" --env "$MACHINE_SLUG" -R "$REPO" > /dev/null
  echo "    secret   ${name}"
}

echo "==> Syncing into Environment '${MACHINE_SLUG}'..."
sync_secret KUBECONFIG "$KUBECONFIG_PATH" value
EXPECTED_SECRETS="KUBECONFIG"
if [ -n "$REGISTRY_ORG" ]; then
  sync_secret REGISTRY_USERNAME "$REGISTRY_PATH" username
  sync_secret REGISTRY_TOKEN "$REGISTRY_PATH" token
  EXPECTED_SECRETS="$EXPECTED_SECRETS REGISTRY_USERNAME REGISTRY_TOKEN"
fi
EXPECTED_VARIABLES=""
if [ -n "$APP_HOST" ]; then
  gh variable set APP_HOST --env "$MACHINE_SLUG" -R "$REPO" --body "$APP_HOST" > /dev/null
  echo "    variable APP_HOST"
  EXPECTED_VARIABLES="APP_HOST"
fi

# ── 4. Verify every name is present ────────────────────────────────────────────

SECRETS=$(gh secret list --env "$MACHINE_SLUG" -R "$REPO" --json name --jq '.[].name')
VARIABLES=$(gh variable list --env "$MACHINE_SLUG" -R "$REPO" --json name --jq '.[].name')
ABSENT=0
for name in $EXPECTED_SECRETS; do
  printf '%s\n' "$SECRETS" | grep -qx "$name" || { echo "ERROR: secret ${name} not listed on '${MACHINE_SLUG}'." >&2; ABSENT=1; }
done
for name in $EXPECTED_VARIABLES; do
  printf '%s\n' "$VARIABLES" | grep -qx "$name" || { echo "ERROR: variable ${name} not listed on '${MACHINE_SLUG}'." >&2; ABSENT=1; }
done
[ "$ABSENT" -eq 0 ] || exit 1

echo ""
echo "==> Environment '${MACHINE_SLUG}' on ${REPO}: ${EXPECTED_SECRETS}${EXPECTED_VARIABLES:+, $EXPECTED_VARIABLES} in place."
