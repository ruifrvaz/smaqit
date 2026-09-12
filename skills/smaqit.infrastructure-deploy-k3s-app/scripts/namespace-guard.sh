#!/usr/bin/env bash
# Pre-flight namespace guard for smaqit.infrastructure-deploy-k3s-app.
#
# Defense-in-depth against ever running with a broader credential than the platform's
# onboarding contract issues. A Namespace-scoped kubeconfig should be restricted to exactly
# one Namespace (app-__APP_SLUG__) with no cluster-scoped verb permitted. This script refuses
# to proceed if either assumption doesn't hold — it is the k3s family's analog of the
# Terraform family's plan-guard.sh/ownership-guard.sh.
#
# Usage: namespace-guard.sh <kubeconfig-path> <expected-namespace>
#   Exit 0 — the kubeconfig's current context targets exactly <expected-namespace>, and
#            `kubectl auth can-i --list` in that namespace grants no cluster-scoped verb.
#   Exit 1 — namespace mismatch, a cluster-scoped verb is permitted, or a required tool/arg
#            is missing. Do NOT proceed to lint/apply on a non-zero exit.

set -euo pipefail

KUBECONFIG_PATH="${1:-}"
EXPECTED_NAMESPACE="${2:-}"

if [ -z "$KUBECONFIG_PATH" ] || [ -z "$EXPECTED_NAMESPACE" ]; then
  echo "Usage: $0 <kubeconfig-path> <expected-namespace>" >&2
  exit 1
fi

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

# ── Cluster-scoped verb check: `kubectl auth can-i --list` must grant nothing cluster-scoped ──
# `--list` with a namespace set reports namespaced rules; passing --all-namespaces or omitting
# -n would report cluster-wide, which is exactly what must never be permitted. Query explicitly
# for common cluster-scoped resources in addition to the general list, since a compact RBAC
# allowlist may not always surface every relevant rule under --list alone.

echo "namespace-guard: checking for cluster-scoped permissions in '$EXPECTED_NAMESPACE'..."

CLUSTER_SCOPED_KINDS=(namespaces nodes clusterroles clusterrolebindings persistentvolumes
  customresourcedefinitions storageclasses)

FOUND_CLUSTER_SCOPE=false
for KIND in "${CLUSTER_SCOPED_KINDS[@]}"; do
  for VERB in get list watch create update patch delete deletecollection; do
    set +e
    kubectl auth can-i "$VERB" "$KIND" -n "$EXPECTED_NAMESPACE" >/dev/null 2>&1
    ALLOWED=$?
    set -e
    if [ "$ALLOWED" -eq 0 ]; then
      echo "namespace-guard: FAIL — credential permits '$VERB' on cluster-scoped resource '$KIND'." >&2
      FOUND_CLUSTER_SCOPE=true
    fi
  done
done

# Belt-and-suspenders: `--list` itself, restricted to the namespace, must not surface any rule
# whose resource has no namespace qualifier attached (a namespaced credential's --list output is
# entirely namespace column "" is expected only for the namespaced verbs already covered above;
# this call is kept for visibility in the report, not as the sole signal).
kubectl auth can-i --list -n "$EXPECTED_NAMESPACE" 2>/dev/null || true

if [ "$FOUND_CLUSTER_SCOPE" = "true" ]; then
  echo "namespace-guard: REFUSING to proceed — this credential is broader than the onboarding" >&2
  echo "                 contract should issue. Do not lint or apply. Report this to the" >&2
  echo "                 platform side rather than working around it." >&2
  exit 1
fi

echo "namespace-guard: PASS — no cluster-scoped verb permitted; namespace matches '$EXPECTED_NAMESPACE'."
exit 0
