#!/usr/bin/env bash
# smaqit.infrastructure-repo-config — runtime operation: put an existing-k3s app's runtime secrets
# for one environment into that environment's machine-slug GitHub Environment.
#
# The secrets come from the Infrastructure spec's `## Runtime Secrets` table, read through
# `smaqit spec environments --env=<environment> --json`; each one's value lives in Vault at
# secret/apps/<app-slug>/<machine-slug>/runtime, field <NAME> (smaqit.infrastructure-vault-loader's
# load-credentials.sh puts it there). The generated deploy workflow builds the app's Kubernetes
# Secret from these Environment secrets.
#
#   no runtime secrets       nothing to sync; exits 0 before any Vault or gh call
#   Vault Access: direct     each value travels from Vault to `gh secret set` on stdin — never as
#                            an argument, never printed — then every name is verified listed
#   Vault Access:            no Vault call at all: the Environment is created if needed, and each
#     operator-scripts       name not yet listed is handed off to the operator (Vault path, field,
#                            Environment, secret name, command shape for their own script); exits 4
#                            until `gh secret list` shows every name, then 0
#
# The app slug defaults to the repository name, normalised like vault-loader's derivation; export
# PROJECT_SLUG to override it. Idempotent: re-run after any rotation.
#
# Usage: bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-repo-config/scripts/sync-runtime-env.sh \
#          <owner>/<repo> <environment>
# Exit:  0 done · 1 error · 3 Environment create refused · 4 handoff pending (operator-scripts)

set -euo pipefail

export VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

usage() {
  echo "Usage: $0 <owner>/<repo> <environment>" >&2
  exit 1
}

[ "$#" -eq 2 ] || usage
REPO="$1"
ENVIRONMENT="$2"
case "$REPO" in */*) ;; *) usage ;; esac

if ! SPEC="$(smaqit spec environments --env="$ENVIRONMENT" --json)"; then
  echo "ERROR: smaqit spec environments --env=${ENVIRONMENT} failed; fix the SLOTS-SPEC-* lines above." >&2
  exit 1
fi
if [ "$(jq '.runtimeSecrets | length' <<<"$SPEC")" -eq 0 ]; then
  echo "==> No runtime secrets for ${ENVIRONMENT}; nothing to sync."
  exit 0
fi

MACHINE_SLUG="$(jq -r '.environments[0].machineSlug' <<<"$SPEC")"
VAULT_ACCESS="$(jq -r '.vaultAccess' <<<"$SPEC")"
mapfile -t NAMES < <(jq -r '.runtimeSecrets[].name' <<<"$SPEC")
# The slug is used verbatim as the Environment name in API paths and --env; gh does not
# URL-encode Environment names, so allow only characters that never need encoding.
if ! printf '%s' "$MACHINE_SLUG" | grep -Eq '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$'; then
  echo "ERROR: machine slug '$MACHINE_SLUG' must be lowercase letters, digits, and inner hyphens." >&2
  exit 1
fi
APP_SLUG="${PROJECT_SLUG:-$(printf '%s' "${REPO#*/}" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9-]/-/g' | tr -s '-' | sed 's/^-//; s/-$//')}"
RUNTIME_PATH="secret/apps/${APP_SLUG}/${MACHINE_SLUG}/runtime"

if ! gh auth status > /dev/null 2>&1; then
  echo "ERROR: gh is not authenticated. Run 'gh auth login' (or export GH_TOKEN) first." >&2
  exit 1
fi

if [ "$VAULT_ACCESS" = "direct" ]; then
  if ! vault token lookup > /dev/null 2>&1; then
    echo "ERROR: No valid Vault token. Run smaqit.infrastructure-vault-loader first." >&2
    exit 1
  fi
  # Every value must exist before anything is written.
  MISSING=0
  for name in "${NAMES[@]}"; do
    if ! vault kv get -field="$name" "$RUNTIME_PATH" > /dev/null 2>&1; then
      echo "ERROR: ${RUNTIME_PATH} has no field ${name}." >&2
      MISSING=1
    fi
  done
  if [ "$MISSING" -ne 0 ]; then
    echo "       Load them first: PROVISIONING_MODE=existing-k3s ENVIRONMENT=${ENVIRONMENT} \\" >&2
    echo "         bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh" >&2
    exit 1
  fi
fi

# ── Create or verify the Environment (gh has no Environment command) ───────────

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

listed() {
  gh secret list --env "$MACHINE_SLUG" -R "$REPO" --json name --jq '.[].name'
}

# ── operator-scripts: hand off, then verify through gh alone ───────────────────

if [ "$VAULT_ACCESS" = "operator-scripts" ]; then
  LISTED="$(listed)"
  PENDING=0
  for name in "${NAMES[@]}"; do
    printf '%s\n' "$LISTED" | grep -qx "$name" && { echo "    secret   ${name} (listed)"; continue; }
    SOURCE="$(jq -r --arg n "$name" '.runtimeSecrets[] | select(.name == $n) | .source' <<<"$SPEC")"
    PENDING=$((PENDING + 1))
    echo ""
    echo "HANDOFF ${name} (${SOURCE}) — run with your own Vault scripts, in your own terminal:"
    echo "  Vault path:   ${RUNTIME_PATH}"
    echo "  Field:        ${name}"
    if [ "$SOURCE" = "generated" ]; then
      echo "  Value:        a random value you generate (at least 32 random bytes, base64, no trailing newline)"
    else
      echo "  Value:        the value the operator provides"
    fi
    echo "  Environment:  ${MACHINE_SLUG} on ${REPO}"
    echo "  Secret name:  ${name}"
    echo "  Command:      <your script that prints ${RUNTIME_PATH} field ${name}> \\"
    echo "                  | gh secret set ${name} --env ${MACHINE_SLUG} -R ${REPO}"
    echo "                (the value goes to gh on stdin — never as an argument, never echoed)"
  done
  echo ""
  if [ "$PENDING" -ne 0 ]; then
    echo "==> ${PENDING} runtime secret(s) handed off for '${MACHINE_SLUG}'. Re-run this script once they are set;"
    echo "    it completes when gh secret list shows every name."
    exit 4
  fi
  echo "==> Environment '${MACHINE_SLUG}' on ${REPO}: all ${#NAMES[@]} runtime secret(s) listed (Vault Access: operator-scripts)."
  exit 0
fi

# ── direct: Vault → gh on stdin ────────────────────────────────────────────────

bash "${SCRIPT_DIR}/check-pat-permissions.sh" "$REPO" "$MACHINE_SLUG"

echo "==> Syncing ${RUNTIME_PATH} into Environment '${MACHINE_SLUG}'..."
for name in "${NAMES[@]}"; do
  vault kv get -field="$name" "$RUNTIME_PATH" \
    | gh secret set "$name" --env "$MACHINE_SLUG" -R "$REPO" > /dev/null
  echo "    secret   ${name}"
done

LISTED="$(listed)"
ABSENT=0
for name in "${NAMES[@]}"; do
  printf '%s\n' "$LISTED" | grep -qx "$name" || { echo "ERROR: secret ${name} not listed on '${MACHINE_SLUG}'." >&2; ABSENT=1; }
done
[ "$ABSENT" -eq 0 ] || exit 1

echo ""
echo "==> Environment '${MACHINE_SLUG}' on ${REPO}: ${#NAMES[@]} runtime secret(s) in place."
