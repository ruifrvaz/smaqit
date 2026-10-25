#!/usr/bin/env bash
# Pre-flight namespace guard for smaqit.infrastructure-deploy-k3s-app.
#
# Defense-in-depth against ever running with a broader credential than the platform's
# onboarding contract grants. The kubeconfig the platform's connect action writes should be
# restricted to exactly one Namespace (app-__APP_SLUG__) with no cluster-scoped verb permitted. This script refuses
# to proceed if either assumption doesn't hold — it is the k3s family's analog of the
# Terraform family's plan-guard.sh/ownership-guard.sh.
#
# Usage: namespace-guard.sh <kubeconfig-path> <expected-namespace> [--deployer-exec denied|allowed]
#   --deployer-exec  The spec's `Deployer Exec` Constraints row (default: denied). Under denied, a
#                    credential that may exec into pods fails the guard; under allowed it only warns.
#   Exit 0 — the kubeconfig's current context targets exactly <expected-namespace>, no
#            cluster-scoped verb is permitted, and exec into pods is denied (or allowed by the spec).
#   Exit 1 — namespace mismatch, a cluster-scoped verb is permitted, exec is permitted under
#            denied, the API is unreachable or rejects the credential, a permission check gives
#            no clear answer, or a required tool/arg is missing. Do NOT proceed to lint/apply on
#            a non-zero exit.
#   Exit 2 — usage error.
#
# Fails closed: only an explicit "no" from the API counts as denied. An unreachable API (for
# example a mesh-only API before the mesh is joined) is a FAIL, never a PASS. Because a freshly
# joined mesh can resolve the API's name before its port answers, a connection failure is retried
# for a bounded window (NAMESPACE_GUARD_REACH_ATTEMPTS, default 6, NAMESPACE_GUARD_REACH_WAIT
# seconds apart, default 5); a credential the API rejects (Unauthorized/Forbidden) is not.
#
# Exec is checked with `--subresource=exec`. Never write `can-i create pods/exec`: kubectl reads
# `pods/exec` as TYPE/NAME ("a pod named exec"), so any deployer allowed to create pods gets "yes".

set -euo pipefail

usage() {
  echo "Usage: $0 <kubeconfig-path> <expected-namespace> [--deployer-exec denied|allowed]" >&2
  exit 2
}

KUBECONFIG_PATH="${1:-}"
EXPECTED_NAMESPACE="${2:-}"
[ -n "$KUBECONFIG_PATH" ] && [ -n "$EXPECTED_NAMESPACE" ] || usage
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

command -v kubectl >/dev/null 2>&1 || { echo "namespace-guard: kubectl not found on PATH" >&2; exit 1; }

if [ ! -f "$KUBECONFIG_PATH" ]; then
  echo "namespace-guard: kubeconfig not found at $KUBECONFIG_PATH" >&2
  exit 1
fi

export KUBECONFIG="$KUBECONFIG_PATH"

case "$EXPECTED_NAMESPACE" in
  app-*) ;;
  *)
    echo "namespace-guard: FAIL — expected namespace '$EXPECTED_NAMESPACE' does not match the" >&2
    echo "                 app-__APP_SLUG__ convention. Refusing to proceed." >&2
    exit 1
    ;;
esac

# ── Namespace check: the kubeconfig's current context must target exactly the expected one ────

CURRENT_NAMESPACE="$(kubectl config view --minify -o jsonpath='{..namespace}' 2>/dev/null || true)"

if [ -z "$CURRENT_NAMESPACE" ]; then
  echo "namespace-guard: FAIL — kubeconfig's current context declares no namespace. A" >&2
  echo "                 Namespace-scoped kubeconfig must pin its context namespace explicitly." >&2
  exit 1
fi

if [ "$CURRENT_NAMESPACE" != "$EXPECTED_NAMESPACE" ]; then
  echo "namespace-guard: FAIL — kubeconfig context namespace '$CURRENT_NAMESPACE' does not match" >&2
  echo "                 expected '$EXPECTED_NAMESPACE'. Refusing to proceed — this may be the" >&2
  echo "                 wrong environment's kubeconfig, or a credential broader than intended." >&2
  exit 1
fi

echo "namespace-guard: namespace OK — context targets '$EXPECTED_NAMESPACE'"

# ── Reachability: the API must answer and accept the credential before any check means anything ──

REQUEST_TIMEOUT="${NAMESPACE_GUARD_REQUEST_TIMEOUT:-15s}"
ERR_FILE="$(mktemp)"
trap 'rm -f "$ERR_FILE"' EXIT

REACH_ATTEMPTS="${NAMESPACE_GUARD_REACH_ATTEMPTS:-6}"
REACH_WAIT="${NAMESPACE_GUARD_REACH_WAIT:-5}"
attempt=1
while :; do
  set +e
  RULES="$(kubectl auth can-i --list -n "$EXPECTED_NAMESPACE" --request-timeout="$REQUEST_TIMEOUT" 2>"$ERR_FILE")"
  REACH=$?
  set -e
  [ "$REACH" -eq 0 ] && break
  if grep -qiE 'unauthorized|forbidden' "$ERR_FILE"; then
    echo "namespace-guard: FAIL — the Kubernetes API rejected this credential:" >&2
    sed 's/^/                 /' "$ERR_FILE" | head -5 >&2
    echo "                 Unauthorized: the platform did not accept this job's identity (its" >&2
    echo "                 audience, repository, or GitHub Environment). Forbidden: the identity is" >&2
    echo "                 not bound to the Namespace yet — the onboarding has not reconciled." >&2
    exit 1
  fi
  if [ "$attempt" -ge "$REACH_ATTEMPTS" ]; then
    echo "namespace-guard: FAIL — the Kubernetes API is unreachable after ${attempt} attempt(s):" >&2
    sed 's/^/                 /' "$ERR_FILE" | head -5 >&2
    echo "                 For a mesh-only API, check that the mesh join step ran and that the" >&2
    echo "                 platform's mesh key reached this job." >&2
    exit 1
  fi
  echo "namespace-guard: API not answering yet (attempt ${attempt}/${REACH_ATTEMPTS}); retrying in ${REACH_WAIT}s" >&2
  attempt=$((attempt + 1))
  sleep "$REACH_WAIT"
done

# ── Cluster-scoped verb check: no cluster-scoped verb may be permitted ──────────────────────────
# `kubectl auth can-i` prints "yes" (exit 0) or "no" (exit 1). Any other result — an error, a
# timeout, an empty answer — is not a "no", so it fails the guard instead of passing it. Query
# common cluster-scoped resources explicitly, since a compact RBAC allowlist may not surface every
# relevant rule under --list alone.

echo "namespace-guard: checking for cluster-scoped permissions in '$EXPECTED_NAMESPACE'..."

CLUSTER_SCOPED_KINDS=(namespaces nodes clusterroles clusterrolebindings persistentvolumes
  customresourcedefinitions storageclasses)

FOUND_CLUSTER_SCOPE=false
UNDETERMINED=false
for KIND in "${CLUSTER_SCOPED_KINDS[@]}"; do
  for VERB in get list watch create update patch delete deletecollection; do
    set +e
    ANSWER="$(kubectl auth can-i "$VERB" "$KIND" -n "$EXPECTED_NAMESPACE" --request-timeout="$REQUEST_TIMEOUT" 2>"$ERR_FILE")"
    RC=$?
    set -e
    ANSWER="$(printf '%s' "$ANSWER" | head -1 | awk '{print $1}')"
    if [ "$RC" -eq 0 ] && [ "$ANSWER" = "yes" ]; then
      echo "namespace-guard: FAIL — credential permits '$VERB' on cluster-scoped resource '$KIND'." >&2
      FOUND_CLUSTER_SCOPE=true
    elif [ "$RC" -ne 0 ] && [ "$ANSWER" = "no" ]; then
      :
    else
      echo "namespace-guard: FAIL — could not determine whether '$VERB' on '$KIND' is permitted (exit $RC): $(head -1 "$ERR_FILE")" >&2
      UNDETERMINED=true
    fi
  done
done

# ── Exec check: the deployer must not be able to exec into pods ─────────────────────────────────
# `create` is the verb current API servers check for pods/exec; `get` is the one older ones used.

echo "namespace-guard: checking whether this credential may exec into pods..."

EXEC_PERMITTED=false
for VERB in create get; do
  set +e
  ANSWER="$(kubectl auth can-i "$VERB" pods --subresource=exec -n "$EXPECTED_NAMESPACE" --request-timeout="$REQUEST_TIMEOUT" 2>"$ERR_FILE")"
  RC=$?
  set -e
  ANSWER="$(printf '%s' "$ANSWER" | head -1 | awk '{print $1}')"
  if [ "$RC" -eq 0 ] && [ "$ANSWER" = "yes" ]; then
    EXEC_PERMITTED=true
    if [ "$DEPLOYER_EXEC" = "allowed" ]; then
      echo "namespace-guard: WARNING — credential permits '$VERB' on pods/exec (allowed by the spec's Deployer Exec row)." >&2
      [ "${GITHUB_ACTIONS:-}" != "true" ] || echo "::warning::namespace-guard: the deployer may exec into pods ('$VERB' pods/exec); allowed by the spec's Deployer Exec row"
    else
      echo "namespace-guard: FAIL — credential permits '$VERB' on pods/exec." >&2
    fi
  elif [ "$RC" -ne 0 ] && [ "$ANSWER" = "no" ]; then
    :
  else
    echo "namespace-guard: FAIL — could not determine whether '$VERB' on pods/exec is permitted (exit $RC): $(head -1 "$ERR_FILE")" >&2
    UNDETERMINED=true
  fi
done

# Kept for visibility in the report, not as the sole signal.
printf '%s\n' "$RULES"

if [ "$FOUND_CLUSTER_SCOPE" = "true" ]; then
  echo "namespace-guard: REFUSING to proceed — this credential is broader than the onboarding" >&2
  echo "                 contract should issue. Do not lint or apply. Report this to the" >&2
  echo "                 platform side rather than working around it." >&2
  exit 1
fi
if [ "$EXEC_PERMITTED" = "true" ] && [ "$DEPLOYER_EXEC" != "allowed" ]; then
  echo "namespace-guard: REFUSING to proceed — this credential may exec into pods. Ask the platform" >&2
  echo "                 side to remove pods/exec from the deployer's Role, or, if the grant is" >&2
  echo "                 deliberate, record it with the spec's \`Deployer Exec: allowed\` row." >&2
  exit 1
fi
if [ "$UNDETERMINED" = "true" ]; then
  echo "namespace-guard: REFUSING to proceed — a permission check gave no clear answer, so the" >&2
  echo "                 credential's scope is unproven. Fix the connection and run again." >&2
  exit 1
fi

if [ "$EXEC_PERMITTED" = "true" ]; then
  echo "namespace-guard: PASS — no cluster-scoped verb permitted; exec into pods permitted (Deployer Exec: allowed); namespace matches '$EXPECTED_NAMESPACE'."
else
  echo "namespace-guard: PASS — no cluster-scoped verb or pods/exec permitted; namespace matches '$EXPECTED_NAMESPACE'."
fi
exit 0
