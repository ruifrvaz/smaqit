#!/usr/bin/env bash
# Deterministic guardrail for smaqit.infrastructure-provision-cyso.
#
# Runs `terraform plan`, then inspects the machine-readable plan for any
# resource change whose actions include "delete" (covers both a pure destroy
# and a destroy-then-create replacement). Exits non-zero and names the
# specific resource(s) if any are found, so `terraform apply` never runs
# against an unreviewed destructive plan.
#
# A deliberate replacement is the one exception, and only when named: each
# --allow-replace address is passed to `terraform plan` as -replace=<address>,
# and the guard then accepts a planned replacement of exactly that address
# (delete plus create, in either order). It still blocks every other delete,
# a pure destroy of a listed address, and a listed address the plan does not
# replace (a mistyped address would otherwise apply as a plain update).
#
# Usage: plan-guard.sh [terraform-working-dir] [--allow-replace <address>]... [--out <planfile>]
#   --allow-replace <address>  Plan with -replace=<address> and accept that replacement. Repeatable.
#   --out <planfile>           Keep the inspected binary plan at <planfile>, so the caller applies
#                              exactly what was guarded: `terraform apply <planfile>`. Without it,
#                              the plan is discarded and a later `terraform apply` plans again.
#   Exit 0 — plan is clean (no changes), additive/update-only, or replaces only the
#            listed addresses. Safe to apply.
#   Exit 1 — terraform plan itself failed, OR the plan would delete/replace anything
#            not explicitly allowed. Caller must NOT run `terraform apply`.
#   Exit 2 — usage error.

set -euo pipefail

usage() {
  echo "Usage: plan-guard.sh [terraform-working-dir] [--allow-replace <address>]... [--out <planfile>]" >&2
  exit 2
}

TF_DIR="."
TF_DIR_SET=0
ALLOW=()
OUT=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    --allow-replace)
      [ "$#" -ge 2 ] && [ -n "$2" ] || usage
      ALLOW+=("$2"); shift 2 ;;
    --out)
      [ "$#" -ge 2 ] && [ -n "$2" ] || usage
      OUT="$2"; shift 2 ;;
    -*) usage ;;
    *)
      [ "$TF_DIR_SET" -eq 0 ] || usage
      TF_DIR="$1"; TF_DIR_SET=1; shift ;;
  esac
done

command -v terraform >/dev/null 2>&1 || { echo "plan-guard: terraform not found on PATH" >&2; exit 1; }
command -v jq >/dev/null 2>&1 || { echo "plan-guard: jq not found on PATH" >&2; exit 1; }

# Resolve --out against the caller's directory before changing into TF_DIR.
if [ -n "$OUT" ]; then
  case "$OUT" in
    /*) ;;
    *) OUT="$(pwd)/$OUT" ;;
  esac
  PLAN_BIN="$OUT"
  rm -f "$PLAN_BIN"
else
  PLAN_BIN="$(mktemp -t tfplan.XXXXXX)"
fi
PLAN_JSON="$(mktemp -t tfplan.XXXXXX.json)"
cleanup() {
  rm -f "$PLAN_JSON"
  [ -n "$OUT" ] || rm -f "$PLAN_BIN"
}
trap cleanup EXIT

# Any failure discards a kept plan, so a blocked plan can never be applied by mistake.
discard_kept_plan() {
  [ -z "$OUT" ] || rm -f "$PLAN_BIN"
}

REPLACE_ARGS=()
for address in ${ALLOW[@]+"${ALLOW[@]}"}; do
  REPLACE_ARGS+=("-replace=${address}")
done

pushd "$TF_DIR" >/dev/null

set +e
terraform plan -detailed-exitcode -input=false -out="$PLAN_BIN" ${REPLACE_ARGS[@]+"${REPLACE_ARGS[@]}"}
PLAN_EXIT=$?
set -e

if [ "$PLAN_EXIT" -eq 1 ]; then
  echo "plan-guard: terraform plan failed (exit 1) — see output above" >&2
  popd >/dev/null
  discard_kept_plan
  exit 1
fi

if [ "$PLAN_EXIT" -eq 0 ]; then
  popd >/dev/null
  if [ "${#ALLOW[@]}" -gt 0 ]; then
    echo "plan-guard: BLOCKED — --allow-replace was given but the plan has no changes, so nothing was replaced:" >&2
    printf '  - %s\n' "${ALLOW[@]}" >&2
    echo "Check the resource address(es) against \`terraform state list\`." >&2
    discard_kept_plan
    exit 1
  fi
  echo "plan-guard: PASS — terraform plan reports no changes"
  exit 0
fi

# PLAN_EXIT == 2: the plan has changes. Inspect them for delete/replace actions.
terraform show -json "$PLAN_BIN" > "$PLAN_JSON"
popd >/dev/null

ALLOW_JSON="$(printf '%s\n' ${ALLOW[@]+"${ALLOW[@]}"} | jq -R . | jq -s 'map(select(length > 0))')"

# A replacement is delete plus create, in either order (create-before-destroy plans it as
# ["create","delete"]). Anything else that includes "delete" is a destroy.
VERDICT="$(jq -r --argjson allow "$ALLOW_JSON" '
  def is_replace: (.change.actions | sort) == ["create","delete"];
  [.resource_changes[]? | select(.change.actions | any(. == "delete"))] as $deletes
  | ($deletes | map(select((.address as $a | $allow | any(. == $a)) and is_replace) | .address)) as $replaced
  | {
      blocked: ($deletes | map(select((.address as $a | $replaced | any(. == $a)) | not)
                 | "  - \(.address): \(.change.actions | join(","))")),
      missing: ($allow | map(select(. as $a | $replaced | any(. == $a) | not)) | map("  - \(.)")),
      replaced: ($replaced | map("  - \(.)"))
    }
' "$PLAN_JSON")"

BLOCKED="$(jq -r '.blocked | join("\n")' <<<"$VERDICT")"
MISSING="$(jq -r '.missing | join("\n")' <<<"$VERDICT")"
REPLACED="$(jq -r '.replaced | join("\n")' <<<"$VERDICT")"

if [ -n "$BLOCKED" ] || [ -n "$MISSING" ]; then
  if [ -n "$BLOCKED" ]; then
    echo "plan-guard: BLOCKED — this plan would delete and/or replace the following resource(s):" >&2
    echo "$BLOCKED" >&2
  fi
  if [ -n "$MISSING" ]; then
    echo "plan-guard: BLOCKED — --allow-replace named address(es) this plan does not replace:" >&2
    echo "$MISSING" >&2
  fi
  echo "" >&2
  echo "Do not apply this plan unattended. Review it manually, confirm the destroy/replace is" >&2
  echo "intentional, and only then apply outside of this guarded path." >&2
  discard_kept_plan
  exit 1
fi

if [ -n "$REPLACED" ]; then
  echo "plan-guard: PASS — plan replaces only the explicitly allowed resource(s):"
  echo "$REPLACED"
else
  echo "plan-guard: PASS — plan is additive/update-only (no delete or replace actions)"
fi
[ -z "$OUT" ] || echo "plan-guard: guarded plan kept at ${PLAN_BIN} — apply it with: terraform apply ${PLAN_BIN}"
exit 0
