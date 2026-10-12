#!/usr/bin/env bash
# smaqit.infrastructure-vault-loader — write one credential to a private temporary file.
# For tools that read a credential from a file (ssh -i, KUBECONFIG, gh's token): the value goes
# from Vault straight into a 0600 file inside a fresh 0700 directory, and only the file's PATH is
# printed — the value never reaches stdout, an argument, or shell history. The caller deletes the
# directory when done: rm -rf "$(dirname "$FILE")".
#
# Usage:
#   FILE="$(bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-vault-loader/scripts/credential-file.sh <path>)"
#   trap 'rm -rf "$(dirname "$FILE")"' EXIT
#
# <path> is relative to secret/ and must be one of (the field is implied by the path):
#   apps/<app-slug>/ssh                      private_key — an app's own deploy key
#   apps/<app-slug>/<machine-slug>/kubeconfig value       — an app's scoped kubeconfig
#   apps/<app-slug>/platform-repo            token        — PAT for the platform repository
#   <project-slug>/ssh                       private_key — legacy flat scheme deploy key
# Anything else is refused: this is not a general-purpose Vault reader.

set -euo pipefail

export VAULT_ADDR="${VAULT_ADDR:-http://127.0.0.1:8200}"

usage() {
  echo "Usage: $0 <apps/<app>/ssh | apps/<app>/<machine>/kubeconfig | apps/<app>/platform-repo | <project>/ssh>" >&2
  exit 2
}

[ "$#" -eq 1 ] || usage
CRED_PATH="$1"
SLUG='[a-z0-9]([a-z0-9-]*[a-z0-9])?'

if [[ "$CRED_PATH" =~ ^apps/${SLUG}/ssh$ ]]; then
  FIELD="private_key"
elif [[ "$CRED_PATH" =~ ^apps/${SLUG}/${SLUG}/kubeconfig$ ]]; then
  FIELD="value"
elif [[ "$CRED_PATH" =~ ^apps/${SLUG}/platform-repo$ ]]; then
  FIELD="token"
elif [[ "$CRED_PATH" =~ ^${SLUG}/ssh$ ]] && [[ ! "$CRED_PATH" =~ ^(apps|machines|organizations)/ ]]; then
  FIELD="private_key"
else
  echo "ERROR: '$CRED_PATH' is not a credential this script hands out (see its header)." >&2
  usage
fi

if ! vault token lookup > /dev/null 2>&1; then
  echo "ERROR: No valid Vault token. Run load-credentials.sh first." >&2
  exit 1
fi

umask 077
DIR="$(mktemp -d)"
FILE="${DIR}/credential"
if ! vault kv get -field="$FIELD" "secret/${CRED_PATH}" > "$FILE" 2>/dev/null || [ ! -s "$FILE" ]; then
  rm -rf "$DIR"
  echo "ERROR: secret/${CRED_PATH} (field ${FIELD}) is missing or empty." >&2
  case "$CRED_PATH" in
    apps/*/*/kubeconfig)
      echo "       The platform has not handed off this machine's kubeconfig yet, or it was never" >&2
      echo "       loaded: finish the platform's onboarding hand-off, then run" >&2
      echo "       PROVISIONING_MODE=existing-k3s MACHINE_SLUG=<machine-slug> load-credentials.sh." >&2
      ;;
  esac
  exit 1
fi
chmod 600 "$FILE"
printf '%s\n' "$FILE"
