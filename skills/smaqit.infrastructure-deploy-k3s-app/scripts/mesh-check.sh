#!/usr/bin/env bash
# platform-prereqs check for smaqit.infrastructure-deploy-k3s-app (operation: mesh-check).
#
# A private-mesh machine's Kubernetes API is reachable only from mesh peers. Before credentials,
# onboarding, or any kubectl call, confirm this machine can reach it: the environment's Mesh
# Wait-For Host (the API's mesh name) must resolve, and the API port must accept a TCP connection.
# A public environment needs no mesh, so the check passes without probing anything. In CI the
# generated deploy.yml joins the mesh itself; this check is for local runs, and the slot plan skips it
# on the existing-cicd-pr and ci-push paths.
#
# Usage: mesh-check.sh <environments.json> <environment> [--port <api-port>]
#   <environments.json>  output of `smaqit spec environments --json` ("-" reads stdin)
#   <environment>        the Environments row to check (e.g. test)
#   --port               the API port (default 6443)
#   Exit 0 — public environment, or the mesh name resolves and the API port answers.
#   Exit 1 — the row is missing, the name does not resolve, or the port does not answer.
#   Exit 2 — usage.

set -euo pipefail

usage() {
  echo "Usage: mesh-check.sh <environments.json> <environment> [--port <api-port>]" >&2
  exit 2
}

PORT=6443
POSITIONAL=()
while [ "$#" -gt 0 ]; do
  case "$1" in
    --port)
      [ "$#" -ge 2 ] && [[ "$2" =~ ^[0-9]+$ ]] || usage
      PORT="$2"; shift 2 ;;
    -) POSITIONAL+=("$1"); shift ;;
    -*) usage ;;
    *) POSITIONAL+=("$1"); shift ;;
  esac
done
[ "${#POSITIONAL[@]}" -eq 2 ] || usage
INPUT="${POSITIONAL[0]}"
ENVIRONMENT="${POSITIONAL[1]}"

command -v jq >/dev/null 2>&1 || { echo "mesh-check: jq not found on PATH" >&2; exit 1; }
if [ "$INPUT" = "-" ]; then
  SPEC="$(cat)"
else
  [ -f "$INPUT" ] || { echo "mesh-check: input not found: ${INPUT}" >&2; exit 1; }
  SPEC="$(cat "$INPUT")"
fi

ROW="$(jq -c --arg e "$ENVIRONMENT" '.environments[]? | select(.environment == $e)' <<<"$SPEC" 2>/dev/null || true)"
if [ -z "$ROW" ]; then
  echo "mesh-check: FAIL — no Environments row for '${ENVIRONMENT}' in the input (expected 'smaqit spec environments --json' output)" >&2
  exit 1
fi
REACHABILITY="$(jq -r '.reachability' <<<"$ROW")"
if [ "$REACHABILITY" != "private-mesh" ]; then
  echo "mesh-check: PASS — '${ENVIRONMENT}' is ${REACHABILITY}; no mesh access needed"
  exit 0
fi
HOST="$(jq -r '.meshWaitForHost // ""' <<<"$ROW")"
[ -n "$HOST" ] || { echo "mesh-check: FAIL — '${ENVIRONMENT}' is private-mesh but declares no Mesh Wait-For Host" >&2; exit 1; }

resolve() {
  if command -v getent >/dev/null 2>&1; then
    getent hosts "$1" | awk '{print $1; exit}' || true
  elif command -v python3 >/dev/null 2>&1; then
    python3 -c 'import socket, sys; print(socket.gethostbyname(sys.argv[1]))' "$1" 2>/dev/null || true
  fi
}

instructions() {
  echo "                Join the platform's mesh on this machine (its mesh client, with your own" >&2
  echo "                login or a setup key the platform issues), confirm the mesh name resolves," >&2
  echo "                then run this slot again. CI joins the mesh itself in deploy.yml." >&2
}

ADDRESS="$(resolve "$HOST")"
if [ -z "$ADDRESS" ]; then
  echo "mesh-check: FAIL — '${HOST}' does not resolve from this machine; it is not on the mesh, or the mesh's DNS is not in use." >&2
  instructions
  exit 1
fi
if ! timeout 5 bash -c 'exec 3<>"/dev/tcp/$1/$2"' _ "$HOST" "$PORT" 2>/dev/null; then
  echo "mesh-check: FAIL — '${HOST}' resolves to ${ADDRESS}, but port ${PORT} does not answer from this machine." >&2
  instructions
  exit 1
fi
echo "mesh-check: PASS — '${HOST}' (${ADDRESS}) answers on port ${PORT}"
