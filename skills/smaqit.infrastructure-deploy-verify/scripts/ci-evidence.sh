#!/usr/bin/env bash
# verify check for smaqit.infrastructure-deploy-verify (operation: ci-evidence).
#
# An existing-k3s deploy on a CI path (existing-cicd-pr, ci-push) already verified the target from
# inside the deploy run: "Verify revision" proved the Deployment runs the run's commit, and "Verify in
# cluster" or "Verify over HTTPS" proved the app answers. This check decides from that run's record
# alone, so the operator's machine needs no mesh access, kubeconfig, or Vault. It reads
# `gh run view <run-id> --json headSha,status,conclusion,jobs` and, only when the run's head is not the
# expected commit itself, `gh api repos/{owner}/{repo}/compare/<expected>...<headSha>`.
#
# Usage: ci-evidence.sh <environments.json> <environment> --run-id <id> --expected-sha <sha>
#   <environments.json>  output of `smaqit spec environments --json` ("-" reads stdin)
#   <environment>        the Environments row the run deployed (e.g. prod)
#   --run-id             the deploy.yml run's database ID
#   --expected-sha       the commit the deploy was meant to ship: the merge SHA (existing-cicd-pr) or
#                        the pushed SHA (ci-push), never the run's own headSha
#   Exit 0 — the run finished successfully at the expected commit or a descendant, and its "Verify
#            revision" and applicable verify step succeeded.
#   Exit 1 — any check failed, or the evidence could not be read.
#   Exit 2 — usage.

set -euo pipefail

usage() {
  echo "Usage: ci-evidence.sh <environments.json> <environment> --run-id <id> --expected-sha <sha>" >&2
  exit 2
}

RUN_ID=""
EXPECTED=""
POSITIONAL=()
while [ "$#" -gt 0 ]; do
  case "$1" in
    --run-id)
      [ "$#" -ge 2 ] && [[ "$2" =~ ^[0-9]+$ ]] || usage
      RUN_ID="$2"; shift 2 ;;
    --expected-sha)
      [ "$#" -ge 2 ] && [[ "$2" =~ ^[0-9a-f]{7,40}$ ]] || usage
      EXPECTED="$2"; shift 2 ;;
    -) POSITIONAL+=("$1"); shift ;;
    -*) usage ;;
    *) POSITIONAL+=("$1"); shift ;;
  esac
done
[ "${#POSITIONAL[@]}" -eq 2 ] && [ -n "$RUN_ID" ] && [ -n "$EXPECTED" ] || usage
INPUT="${POSITIONAL[0]}"
ENVIRONMENT="${POSITIONAL[1]}"

fail() {
  echo "ci-evidence: FAIL — $*" >&2
  exit 1
}

command -v jq >/dev/null 2>&1 || { echo "ci-evidence: jq not found on PATH" >&2; exit 1; }
command -v gh >/dev/null 2>&1 || { echo "ci-evidence: gh not found on PATH" >&2; exit 1; }
if [ "$INPUT" = "-" ]; then
  SPEC="$(cat)"
else
  [ -f "$INPUT" ] || { echo "ci-evidence: input not found: ${INPUT}" >&2; exit 1; }
  SPEC="$(cat "$INPUT")"
fi

ROW="$(jq -c --arg e "$ENVIRONMENT" '.environments[]? | select(.environment == $e)' <<<"$SPEC" 2>/dev/null || true)"
[ -n "$ROW" ] || fail "no Environments row for '${ENVIRONMENT}' in the input (expected 'smaqit spec environments --json' output)"
# Same rule as the generated deploy.yml's VERIFY and installer/infra_spec.go verifiesInCluster().
STEP="$(jq -r 'if .reachability == "private-mesh" or .ingressMode == "platform-owned"
               then "Verify in cluster" else "Verify over HTTPS" end' <<<"$ROW")"

# A failed read is a failed check, never an empty result.
if ! RUN="$(gh run view "$RUN_ID" --json headSha,status,conclusion,jobs 2>/dev/null)" \
   || ! jq -e 'type == "object" and has("headSha") and has("jobs")' <<<"$RUN" >/dev/null 2>&1; then
  fail "could not read run ${RUN_ID} with 'gh run view'; check the run ID and gh's repository and token"
fi

STATUS="$(jq -r '.status // ""' <<<"$RUN")"
[ "$STATUS" = "completed" ] || fail "run ${RUN_ID} has not finished (status '${STATUS}'); wait for it, then run this check again"
CONCLUSION="$(jq -r '.conclusion // ""' <<<"$RUN")"
[ "$CONCLUSION" = "success" ] || fail "run ${RUN_ID} concluded '${CONCLUSION}', not success"

HEAD_SHA="$(jq -r '.headSha // ""' <<<"$RUN")"
[ -n "$HEAD_SHA" ] || fail "run ${RUN_ID} reports no headSha"
if [ "$HEAD_SHA" != "$EXPECTED" ] && [[ "$HEAD_SHA" != "$EXPECTED"* ]]; then
  # The dispatcher deploys the branch tip at dispatch time, which may already contain later merges.
  if ! COMPARE="$(gh api "repos/{owner}/{repo}/compare/${EXPECTED}...${HEAD_SHA}" --jq .status 2>/dev/null)"; then
    fail "could not compare ${EXPECTED} with run ${RUN_ID}'s head ${HEAD_SHA} through the GitHub API"
  fi
  case "$COMPARE" in
    identical|ahead) ;;
    *) fail "wrong SHA — run ${RUN_ID} deployed ${HEAD_SHA}, which does not contain ${EXPECTED} (compare status '${COMPARE}')" ;;
  esac
fi

DEPLOY_JOB="$(jq -c '[.jobs[]? | select(.name == "deploy")][0] // empty' <<<"$RUN")"
[ -n "$DEPLOY_JOB" ] || fail "run ${RUN_ID} has no 'deploy' job; is it a run of the generated k3s deploy.yml?"

step_conclusion() {
  jq -r --arg s "$1" '[.steps[]? | select(.name == $s)][0] | if . == null then "missing" else (.conclusion // "unfinished") end' <<<"$DEPLOY_JOB"
}

check_step() {
  local name="$1" result
  result="$(step_conclusion "$name")"
  case "$result" in
    success) ;;
    missing) fail "step '${name}' is missing from run ${RUN_ID}'s deploy job; regenerate deploy.yml with smaqit.infrastructure-cicd-generate" ;;
    skipped) fail "step '${name}' was skipped in run ${RUN_ID}; the workflow and the '${ENVIRONMENT}' Environments row disagree on how to verify (regenerate deploy.yml, or check that the run deployed '${ENVIRONMENT}')" ;;
    *) fail "step '${name}' concluded '${result}' in run ${RUN_ID}" ;;
  esac
}

check_step "Verify revision"
check_step "$STEP"
echo "ci-evidence: PASS — run ${RUN_ID} deployed ${HEAD_SHA} (contains ${EXPECTED}); 'Verify revision' and '${STEP}' succeeded"
