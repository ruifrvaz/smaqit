#!/usr/bin/env bash
# Manifest lint for smaqit.infrastructure-deploy-k3s-app.
#
# Rejects a manifest set BEFORE any `kubectl apply`, rather than letting the platform's
# admission control reject it after the fact — the scoped credential this skill uses cannot
# read pod logs, so a post-hoc admission failure is much harder to diagnose than a fast,
# explicit local rejection naming the offending field.
#
# Checks, across every Pod-template-bearing manifest given:
#   1. Pod Security "restricted" fields: runAsNonRoot, allowPrivilegeEscalation: false,
#      capabilities.drop: [ALL], seccompProfile.type: RuntimeDefault; no privileged: true, no
#      hostPath volumes, no hostNetwork/hostPID/hostIPC.
#   2. A numeric runAsUser wherever runAsNonRoot is true — a base image whose own USER directive
#      names a user (not a numeric UID) otherwise passes this lint but fails at admission time
#      with CreateContainerConfigError, since the kubelet cannot resolve a named user to verify
#      non-root itself (docker run never surfaces this — it is Kubernetes-only).
#   3. Explicit CPU/memory requests AND limits on every container.
#   4. No cluster-scoped `kind` anywhere in the manifest set.
#
# Usage: manifest-lint.sh <manifest.yaml> [<manifest.yaml> ...]
#   Exit 0 — every check passes across every given manifest.
#   Exit 1 — at least one check failed; the offending file and field are named on stderr.
#
# Requires: python3 with PyYAML (yaml module). If PyYAML is unavailable, this script fails
# loudly rather than silently skipping checks — a lint that can't run is not a passed lint.

set -euo pipefail

if [ "$#" -eq 0 ]; then
  echo "Usage: $0 <manifest.yaml> [<manifest.yaml> ...]" >&2
  exit 1
fi

command -v python3 >/dev/null 2>&1 || { echo "manifest-lint: python3 not found on PATH" >&2; exit 1; }

python3 - "$@" <<'PYEOF'
import sys
import yaml

CLUSTER_SCOPED_KINDS = {
    "Namespace", "Node", "ClusterRole", "ClusterRoleBinding", "PersistentVolume",
    "CustomResourceDefinition", "StorageClass", "PriorityClass", "APIService",
    "MutatingWebhookConfiguration", "ValidatingWebhookConfiguration",
}

failures = []


def fail(path, doc_index, kind, name, msg):
    failures.append(f"{path} (doc {doc_index}, {kind or '?'}/{name or '?'}): {msg}")


def pod_spec_containers(spec):
    containers = list(spec.get("containers") or [])
    containers += list(spec.get("initContainers") or [])
    return containers


# NOTE: allowPrivilegeEscalation, capabilities, and privileged are fields of a container's
# SecurityContext only — they do not exist on a Pod's (top-level) SecurityContext in the
# Kubernetes API. runAsNonRoot and seccompProfile exist at both levels and a container-level
# value overrides the pod-level one. Effective-value merging below reflects that inheritance
# instead of requiring container-only fields at the pod level, which no valid manifest can ever
# satisfy.


def check_container_security_context(pod_sc, csc, label, path, doc_index, kind, name):
    pod_sc = pod_sc or {}
    csc = csc or {}

    effective_run_as_non_root = csc.get("runAsNonRoot", pod_sc.get("runAsNonRoot"))
    if effective_run_as_non_root is not True:
        fail(path, doc_index, kind, name, f"{label}: runAsNonRoot must be true (pod- or container-level)")
    else:
        effective_run_as_user = csc.get("runAsUser", pod_sc.get("runAsUser"))
        if not isinstance(effective_run_as_user, int) or isinstance(effective_run_as_user, bool):
            fail(
                path, doc_index, kind, name,
                f"{label}: runAsNonRoot is true but no numeric runAsUser is set (pod- or "
                "container-level) — a base image whose own USER directive names a user, not a "
                "numeric UID, fails admission with CreateContainerConfigError",
            )

    if csc.get("allowPrivilegeEscalation") is not False:
        fail(path, doc_index, kind, name, f"{label}: allowPrivilegeEscalation must be false")

    caps = (csc.get("capabilities") or {}).get("drop") or []
    if "ALL" not in caps:
        fail(path, doc_index, kind, name, f"{label}: capabilities.drop must include ALL")

    effective_seccomp = csc.get("seccompProfile") or pod_sc.get("seccompProfile") or {}
    if effective_seccomp.get("type") != "RuntimeDefault":
        fail(path, doc_index, kind, name, f"{label}: seccompProfile.type must be RuntimeDefault (pod- or container-level)")

    if csc.get("privileged") is True:
        fail(path, doc_index, kind, name, f"{label}: privileged must not be true")


def check_pod_spec(spec, path, doc_index, kind, name):
    if spec.get("hostNetwork") is True:
        fail(path, doc_index, kind, name, "hostNetwork must not be true")
    if spec.get("hostPID") is True:
        fail(path, doc_index, kind, name, "hostPID must not be true")
    if spec.get("hostIPC") is True:
        fail(path, doc_index, kind, name, "hostIPC must not be true")
    for volume in spec.get("volumes") or []:
        if "hostPath" in volume:
            fail(path, doc_index, kind, name, f"volume '{volume.get('name')}' uses hostPath — not permitted")

    pod_sc = spec.get("securityContext") or {}

    for container in pod_spec_containers(spec):
        cname = container.get("name", "?")
        csc = container.get("securityContext") or {}
        check_container_security_context(pod_sc, csc, f"container '{cname}' securityContext", path, doc_index, kind, name)

        resources = container.get("resources") or {}
        requests = resources.get("requests") or {}
        limits = resources.get("limits") or {}
        for field in ("cpu", "memory"):
            if field not in requests:
                fail(path, doc_index, kind, name, f"container '{cname}' missing resources.requests.{field}")
            if field not in limits:
                fail(path, doc_index, kind, name, f"container '{cname}' missing resources.limits.{field}")


def find_pod_template_spec(doc):
    kind = doc.get("kind")
    if kind == "Pod":
        return doc.get("spec")
    template = (doc.get("spec") or {}).get("template")
    if template:
        return template.get("spec")
    return None


for path in sys.argv[1:]:
    with open(path) as f:
        raw = f.read()
    try:
        docs = [d for d in yaml.safe_load_all(raw) if d]
    except yaml.YAMLError as exc:
        print(f"manifest-lint: FAIL — {path} is not valid YAML: {exc}", file=sys.stderr)
        sys.exit(1)

    for i, doc in enumerate(docs):
        kind = doc.get("kind")
        name = (doc.get("metadata") or {}).get("name")

        if kind in CLUSTER_SCOPED_KINDS:
            fail(path, i, kind, name, f"cluster-scoped kind '{kind}' is not permitted in an app manifest set")
            continue

        pod_spec = find_pod_template_spec(doc)
        if pod_spec:
            check_pod_spec(pod_spec, path, i, kind, name)

if failures:
    print("manifest-lint: FAIL — rejected before any kubectl apply:", file=sys.stderr)
    for item in failures:
        print(f"  - {item}", file=sys.stderr)
    sys.exit(1)

print("manifest-lint: PASS — all manifests satisfy restricted Pod Security, explicit resource limits, and no cluster-scoped kind.")
PYEOF
