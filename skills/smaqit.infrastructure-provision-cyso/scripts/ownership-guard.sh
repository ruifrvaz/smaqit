#!/usr/bin/env bash
# Pre-flight ownership guard for smaqit.infrastructure-provision-cyso.
#
# Defense-in-depth for a direct/manual invocation of this skill that bypasses
# smaqit.new-greenfield-project's provisioning_mode branching. If a target VM
# is already declared (via an explicit argument, the VM_HOST env var, or the
# VM_HOST repository variable) but this project's own Terraform state has no
# matching openstack_compute_instance_v2 resource, provisioning would silently
# create a second VM nobody asked for. Stop instead and point at the
# existing-shared path. VM_HOST is a GitHub Actions *variable* (see
# smaqit.infrastructure-repo-config) — unlike a secret, it can be read back
# directly via `gh variable get`, so this script doesn't depend on the caller
# having already exported it into the environment.
#
# Usage: ownership-guard.sh [terraform-working-dir] [explicit-target-ip]
#   Exit 0 — no target VM declared yet (fresh provision), OR the declared
#            target is already owned by a matching resource in this
#            project's Terraform state. Safe to proceed.
#   Exit 1 — a target VM is declared but not owned by this project's state.
#            Do NOT run `terraform apply`. Use provisioning_mode: existing-shared
#            instead (see smaqit.new-greenfield-project Phase 4/5 and
#            smaqit.infrastructure-vault-loader / smaqit.infrastructure-repo-config
#            for the restricted-path handling that mode needs).

set -euo pipefail

TF_DIR="${1:-.}"
EXPLICIT_TARGET="${2:-}"

command -v terraform >/dev/null 2>&1 || { echo "ownership-guard: terraform not found on PATH" >&2; exit 1; }

# ── Resolve the declared target, if any ────────────────────────────────────────
# Priority: explicit argument > VM_HOST env var (already exposed by the
# workflow's env: block, avoiding an extra API call) > `gh variable get`
# (works standalone — no env var plumbing required by the caller).

TARGET="$EXPLICIT_TARGET"

if [ -z "$TARGET" ] && [ -n "${VM_HOST:-}" ]; then
  TARGET="$VM_HOST"
fi

if [ -z "$TARGET" ] && command -v gh >/dev/null 2>&1; then
  set +e
  TARGET="$(gh variable get VM_HOST 2>/dev/null)"
  set -e
fi

if [ -z "$TARGET" ]; then
  echo "ownership-guard: PASS — no target VM declared (VM_HOST variable unset). Fresh provision."
  exit 0
fi

echo "ownership-guard: target VM declared: $TARGET"

# ── Check whether this project's Terraform state owns a matching resource ─────

pushd "$TF_DIR" >/dev/null

set +e
INSTANCE_ADDRS="$(terraform state list 2>/dev/null | grep 'openstack_compute_instance_v2' || true)"
set -e

if [ -z "$INSTANCE_ADDRS" ]; then
  popd >/dev/null
  echo "ownership-guard: BLOCKED — VM_HOST is set to $TARGET, but this project's Terraform" >&2
  echo "state has no openstack_compute_instance_v2 resource at all." >&2
  echo "" >&2
  echo "This looks like a deploy onto a VM another project owns and manages via its own" >&2
  echo "Terraform state (co-hosting), not a VM this project should provision." >&2
  echo "Use provisioning_mode: existing-shared instead — see smaqit.new-greenfield-project" >&2
  echo "Phase 4/5 for the branch, and skip this skill entirely for that path." >&2
  exit 1
fi

MATCHED=false
while IFS= read -r ADDR; do
  [ -z "$ADDR" ] && continue
  set +e
  STATE_IP="$(terraform state show "$ADDR" 2>/dev/null | grep -E '^\s*(access_ip_v4|fixed_ip|access_ip_v6)\s*=' | head -1 | sed -E 's/^[^=]*=\s*"?([^"]*)"?.*/\1/')"
  set -e
  if [ -n "$STATE_IP" ] && [ "$STATE_IP" = "$TARGET" ]; then
    MATCHED=true
    break
  fi
done <<< "$INSTANCE_ADDRS"

popd >/dev/null

if [ "$MATCHED" = "true" ]; then
  echo "ownership-guard: PASS — $TARGET is already owned by this project's Terraform state."
  echo "terraform apply on this state is expected to be an idempotent no-op (existing-owned redeploy)."
  exit 0
fi

echo "ownership-guard: BLOCKED — VM_HOST is set to $TARGET, but no openstack_compute_instance_v2" >&2
echo "resource in this project's Terraform state has a matching IP." >&2
echo "" >&2
echo "This looks like a deploy onto a VM another project owns and manages via its own" >&2
echo "Terraform state (co-hosting), not a VM this project should provision." >&2
echo "Use provisioning_mode: existing-shared instead — see smaqit.new-greenfield-project" >&2
echo "Phase 4/5 for the branch, and skip this skill entirely for that path." >&2
exit 1
