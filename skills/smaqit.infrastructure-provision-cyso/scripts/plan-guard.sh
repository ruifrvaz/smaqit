#!/usr/bin/env bash
# Deterministic guardrail for smaqit.infrastructure-provision-cyso.
#
# Runs `terraform plan`, then inspects the machine-readable plan for any
# resource change whose actions include "delete" (covers both a pure destroy
# and a destroy-then-create replacement). Exits non-zero and names the
# specific resource(s) if any are found, so `terraform apply` never runs
# against an unreviewed destructive plan.
#
# Usage: plan-guard.sh [terraform-working-dir]
#   Exit 0 — plan is clean (no changes) or additive/update-only. Safe to apply.
#   Exit 1 — terraform plan itself failed, OR the plan would delete/replace
#            one or more resources. Caller must NOT run `terraform apply`.

set -euo pipefail

TF_DIR="${1:-.}"

command -v terraform >/dev/null 2>&1 || { echo "plan-guard: terraform not found on PATH" >&2; exit 1; }
command -v jq >/dev/null 2>&1 || { echo "plan-guard: jq not found on PATH" >&2; exit 1; }

PLAN_BIN="$(mktemp -t tfplan.XXXXXX)"
PLAN_JSON="$(mktemp -t tfplan.XXXXXX.json)"
trap 'rm -f "$PLAN_BIN" "$PLAN_JSON"' EXIT

pushd "$TF_DIR" >/dev/null

set +e
terraform plan -detailed-exitcode -input=false -out="$PLAN_BIN"
PLAN_EXIT=$?
set -e

if [ "$PLAN_EXIT" -eq 0 ]; then
  echo "plan-guard: PASS — terraform plan reports no changes"
  popd >/dev/null
  exit 0
fi

if [ "$PLAN_EXIT" -eq 1 ]; then
  echo "plan-guard: terraform plan failed (exit 1) — see output above" >&2
  popd >/dev/null
  exit 1
fi

# PLAN_EXIT == 2: the plan has changes. Inspect them for delete/replace actions.
terraform show -json "$PLAN_BIN" > "$PLAN_JSON"

DESTRUCTIVE="$(jq -r '
  [.resource_changes[]? | select(.change.actions | any(. == "delete"))]
  | map("  - \(.address): \(.change.actions | join(","))")
  | join("\n")
' "$PLAN_JSON")"

popd >/dev/null

if [ -n "$DESTRUCTIVE" ]; then
  echo "plan-guard: BLOCKED — this plan would delete and/or replace the following resource(s):" >&2
  echo "$DESTRUCTIVE" >&2
  echo "" >&2
  echo "Do not apply this plan unattended. Review it manually, confirm the destroy/replace is" >&2
  echo "intentional, and only then apply outside of this guarded path." >&2
  exit 1
fi

echo "plan-guard: PASS — plan is additive/update-only (no delete or replace actions)"
exit 0
