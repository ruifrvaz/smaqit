#!/usr/bin/env bash
# Post-rollout exec check for smaqit.infrastructure-deploy-k3s-app.
#
# namespace-guard.sh asks the API whether the deployer may exec into pods. This script proves it
# against the running app: it attempts a real `kubectl exec` into the Deployment and requires the
# API to refuse it as Forbidden. It runs after the rollout, when a pod exists to exec into.
#
# Usage: exec-refused.sh <namespace> <deployment> [--deployer-exec denied|allowed]
#   Uses the kubeconfig in $KUBECONFIG. --deployer-exec is the spec's `Deployer Exec` Constraints
#   row (default: denied).
#   Exit 0 — the exec was refused as Forbidden, or it was permitted and the spec allows it (warning).
#   Exit 1 — the exec was permitted under denied, or the result was anything else (NotFound, a
#            timeout, a connection error): fails closed under both settings.
#   Exit 2 — usage error.
#
# A non-zero exit is the expected answer here, so the attempt runs under `set +e` and the verdict
# comes from the captured output. Under `bash -e` (GitHub Actions' default shell) capturing it
# directly would end the step on the correct answer before it is evaluated.

set -euo pipefail

usage() {
  echo "Usage: $0 <namespace> <deployment> [--deployer-exec denied|allowed]" >&2
  exit 2
}

NAMESPACE="${1:-}"
DEPLOYMENT="${2:-}"
[ -n "$NAMESPACE" ] && [ -n "$DEPLOYMENT" ] || usage
shift 2
DEPLOYER_EXEC="denied"
while [ "$#" -gt 0 ]; do
  case "$1" in
    --deployer-exec)
      [ "$#" -ge 2 ] || usage
      DEPLOYER_EXEC="$2"; shift 2 ;;
    *) usage ;;
  esac
done
case "$DEPLOYER_EXEC" in
  denied|allowed) ;;
  *) usage ;;
esac

command -v kubectl >/dev/null 2>&1 || { echo "exec-refused: kubectl not found on PATH" >&2; exit 1; }

TIMEOUT="${EXEC_REFUSED_TIMEOUT:-30}"
RUN=(kubectl exec -n "$NAMESPACE" "deployment/$DEPLOYMENT" --request-timeout=15s -- true)
if command -v timeout >/dev/null 2>&1; then
  RUN=(timeout "$TIMEOUT" "${RUN[@]}")
fi

set +e
OUTPUT="$("${RUN[@]}" 2>&1)"
RC=$?
set -e

permitted() {
  if [ "$DEPLOYER_EXEC" = "allowed" ]; then
    echo "exec-refused: WARNING — the deployer could exec into deployment/$DEPLOYMENT ($1); allowed by the spec's Deployer Exec row." >&2
    [ "${GITHUB_ACTIONS:-}" != "true" ] || echo "::warning::exec-refused: the deployer could exec into deployment/$DEPLOYMENT; allowed by the spec's Deployer Exec row"
    exit 0
  fi
  echo "exec-refused: FAIL — the deployer could exec into deployment/$DEPLOYMENT ($1)." >&2
  echo "              Ask the platform side to remove pods/exec from the deployer's Role, or, if the" >&2
  echo "              grant is deliberate, record it with the spec's \`Deployer Exec: allowed\` row." >&2
  exit 1
}

if [ "$RC" -eq 0 ]; then
  permitted "the command ran"
fi
case "$OUTPUT" in
  *Forbidden*|*forbidden*)
    echo "exec-refused: PASS — exec into deployment/$DEPLOYMENT was refused: $(printf '%s' "$OUTPUT" | head -1)"
    exit 0 ;;
  # The API let the exec through and only the command inside the container failed (for
  # example a distroless image without `true`): exec itself was permitted.
  *"command terminated with exit code"*|*"OCI runtime exec failed"*|*"executable file not found"*)
    permitted "the exec was accepted; the command inside the container failed" ;;
esac
if [ "$RC" -eq 124 ]; then
  echo "exec-refused: FAIL — the exec attempt timed out after ${TIMEOUT}s; the result is unproven." >&2
else
  echo "exec-refused: FAIL — the exec attempt was neither refused as Forbidden nor permitted (exit $RC):" >&2
  printf '%s\n' "$OUTPUT" | head -5 | sed 's/^/              /' >&2
fi
exit 1
