#!/usr/bin/env bash
# smaqit.infrastructure-repo-config — preflight the gh token before writing GitHub Environment
# secrets or variables. Probes read-only endpoints for each named Environment and maps the result
# to the fix, so a missing grant surfaces before any write instead of as a bare 403 halfway through
# a sync.
#
# A fine-grained PAT needs more than repository-level Environments + Secrets + Variables grants to
# write an Environment's secrets/variables: writes return 403 until organization-level Secrets and
# Variables read/write is also enabled on the token. Known signatures:
#   Environment create          → 403 "Resource not accessible by personal access token"
#   gh variable set --env <env> → 404 (the Environment does not exist yet)
#   gh variable set --env <env> → 403 with repository-only grants (org-level grants missing)
#
# Usage: bash [SMAQIT_SKILLS_DIR]/smaqit.infrastructure-repo-config/scripts/check-pat-permissions.sh <owner>/<repo> <environment>...
# Exit:  0 all probes pass · 2 an Environment does not exist · 3 a permission is missing · 1 usage/auth

set -uo pipefail

REPO="${1:-}"
[ -n "$REPO" ] && [ "$#" -ge 2 ] || {
  echo "Usage: $0 <owner>/<repo> <environment>..." >&2
  exit 1
}
shift

if ! gh auth status > /dev/null 2>&1; then
  echo "ERROR: gh is not authenticated. Run 'gh auth login' (or export GH_TOKEN) first." >&2
  exit 1
fi

PAT_FIX="Enable organization-level Secrets and Variables read/write on the fine-grained PAT (repository-level Environments + Secrets + Variables grants are not enough for Environment writes), then re-authenticate gh."

# probe <path> — prints the HTTP status of a GET, or 000 when no status could be read.
probe() {
  local out status
  out=$(gh api -X GET -i "$1" 2>&1)
  status=$(printf '%s\n' "$out" | sed -n 's#^HTTP/[0-9.]* \([0-9][0-9][0-9]\).*#\1#p' | head -n1)
  if [ -z "$status" ]; then
    status=$(printf '%s\n' "$out" | sed -n 's/.*(HTTP \([0-9][0-9][0-9]\)).*/\1/p' | head -n1)
  fi
  printf '%s' "${status:-000}"
}

RESULT=0
for ENV_NAME in "$@"; do
  echo "==> ${REPO} Environment '${ENV_NAME}'"

  status=$(probe "repos/${REPO}/environments/${ENV_NAME}")
  case "$status" in
    200) echo "    exists" ;;
    404)
      echo "    MISSING — the Environment does not exist (a 'gh variable set --env' would return 404)."
      echo "    Fix: create it first: gh api repos/${REPO}/environments/${ENV_NAME} -X PUT"
      [ "$RESULT" -eq 0 ] && RESULT=2
      continue
      ;;
    403)
      echo "    FORBIDDEN — 403 'Resource not accessible by personal access token' reading the Environment."
      echo "    Fix: grant the token repository Environments read/write. ${PAT_FIX}"
      RESULT=3
      continue
      ;;
    *)
      echo "    UNKNOWN — reading the Environment returned '${status}'. Check the repository name and network."
      RESULT=3
      continue
      ;;
  esac

  status=$(probe "repos/${REPO}/environments/${ENV_NAME}/secrets/public-key")
  if [ "$status" = "200" ]; then
    echo "    secrets: accessible"
  else
    echo "    secrets: FORBIDDEN (${status}). ${PAT_FIX}"
    RESULT=3
  fi

  status=$(probe "repos/${REPO}/environments/${ENV_NAME}/variables")
  if [ "$status" = "200" ]; then
    echo "    variables: accessible"
  else
    echo "    variables: FORBIDDEN (${status}) — a 'gh variable set --env' would return 403. ${PAT_FIX}"
    RESULT=3
  fi
done

if [ "$RESULT" -eq 0 ]; then
  echo "OK — the token can read every named Environment's secrets and variables."
fi
exit "$RESULT"
