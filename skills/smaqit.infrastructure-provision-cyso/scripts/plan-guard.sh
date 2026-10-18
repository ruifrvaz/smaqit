#!/usr/bin/env bash
# Deterministic guardrail for smaqit.infrastructure-provision-cyso.
#
# Runs `terraform plan`, then inspects the machine-readable plan for any
# resource change whose actions include "delete" (covers both a pure destroy
# and a destroy-then-create replacement). Exits non-zero and names the
# specific resource(s) if any are found, so `terraform apply` never runs
# against an unreviewed destructive plan.
#
# Two deliberate exceptions exist, and only when named:
#   - Each --allow-replace address is passed to `terraform plan` as -replace=<address>,
#     and the guard accepts a planned replacement of exactly that address (delete plus
#     create, in either order).
#   - Each --allow-delete address adds nothing to the plan; the guard accepts a planned
#     pure destroy (actions exactly ["delete"]) of exactly that address, such as a
#     reviewed removal of one firewall rule.
# The guard still blocks every other delete, a listed address the plan acts on the other
# way (a replace where a destroy was allowed, or the reverse), and a listed address the
# plan does not touch (a mistyped address would otherwise pass silently). An address may
# be named by only one of the two flags.
#
# Usage: plan-guard.sh [terraform-working-dir] [--allow-replace <address>]... [--allow-delete <address>]... [--out <planfile>]
#   --allow-replace <address>  Plan with -replace=<address> and accept that replacement. Repeatable.
#   --allow-delete <address>   Accept a pure destroy of <address>. Repeatable.
#   --out <planfile>           Keep the inspected binary plan at <planfile>, so the caller applies
#                              exactly what was guarded: `terraform apply <planfile>`. Without it,
#                              the plan is discarded and a later `terraform apply` plans again.
#   Exit 0 — plan is clean (no changes), additive/update-only, or replaces and destroys
#            only the listed addresses, each as listed. Safe to apply.
#   Exit 1 — terraform plan itself failed, OR the plan would delete/replace anything
#            not explicitly allowed, OR a listed address was not replaced/destroyed as
#            listed. Caller must NOT run `terraform apply`.
#   Exit 2 — usage error.

set -euo pipefail

usage() {
  echo "Usage: plan-guard.sh [terraform-working-dir] [--allow-replace <address>]... [--allow-delete <address>]... [--out <planfile>]" >&2
  exit 2
}

TF_DIR="."
TF_DIR_SET=0
ALLOW=()
ALLOW_DELETE=()
OUT=""
while [ "$#" -gt 0 ]; do
  case "$1" in
    --allow-replace)
      [ "$#" -ge 2 ] && [ -n "$2" ] || usage
      ALLOW+=("$2"); shift 2 ;;
    --allow-delete)
      [ "$#" -ge 2 ] && [ -n "$2" ] || usage
      ALLOW_DELETE+=("$2"); shift 2 ;;
    --out)
      [ "$#" -ge 2 ] && [ -n "$2" ] || usage
      OUT="$2"; shift 2 ;;
    -*) usage ;;
    *)
      [ "$TF_DIR_SET" -eq 0 ] || usage
      TF_DIR="$1"; TF_DIR_SET=1; shift ;;
  esac
done

# One address, one intent: the guard never picks between a replacement and a destroy.
for address in ${ALLOW_DELETE[@]+"${ALLOW_DELETE[@]}"}; do
  for replaced in ${ALLOW[@]+"${ALLOW[@]}"}; do
    if [ "$address" = "$replaced" ]; then
      echo "plan-guard: ${address} is named by both --allow-replace and --allow-delete; choose one" >&2
      usage
    fi
  done
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
  if [ "${#ALLOW[@]}" -gt 0 ] || [ "${#ALLOW_DELETE[@]}" -gt 0 ]; then
    echo "plan-guard: BLOCKED — the plan has no changes, so nothing was replaced or destroyed:" >&2
    for address in ${ALLOW[@]+"${ALLOW[@]}"}; do echo "  - ${address} (--allow-replace)" >&2; done
    for address in ${ALLOW_DELETE[@]+"${ALLOW_DELETE[@]}"}; do echo "  - ${address} (--allow-delete)" >&2; done
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

to_json_list() { printf '%s\n' "$@" | jq -R . | jq -s 'map(select(length > 0))'; }
ALLOW_JSON="$(to_json_list ${ALLOW[@]+"${ALLOW[@]}"})"
ALLOW_DELETE_JSON="$(to_json_list ${ALLOW_DELETE[@]+"${ALLOW_DELETE[@]}"})"

# A replacement is delete plus create, in either order (create-before-destroy plans it as
# ["create","delete"]). A destroy is exactly ["delete"]. Each listed address must be acted on
# exactly as listed; a missing one is reported with what the plan does to it instead.
VERDICT="$(jq -r --argjson allow "$ALLOW_JSON" --argjson allow_delete "$ALLOW_DELETE_JSON" '
  def is_replace: (.change.actions | sort) == ["create","delete"];
  def is_destroy: .change.actions == ["delete"];
  def listed($list): .address as $a | $list | any(. == $a);
  (.resource_changes // []) as $changes
  | [$changes[] | select(.change.actions | any(. == "delete"))] as $deletes
  | ($deletes | map(select(listed($allow) and is_replace) | .address)) as $replaced
  | ($deletes | map(select(listed($allow_delete) and is_destroy) | .address)) as $destroyed
  | def planned($a): ([$changes[] | select(.address == $a) | .change.actions | join(",")] | first)
                     // "not in the plan";
    def missing($list; $done): $list | map(select(. as $a | $done | any(. == $a) | not))
                               | map("  - \(.) (planned: \(planned(.)))");
  {
      blocked: ($deletes | map(select(.address as $a | ($replaced + $destroyed) | any(. == $a) | not)
                 | "  - \(.address): \(.change.actions | join(","))")),
      missing_replace: missing($allow; $replaced),
      missing_delete: missing($allow_delete; $destroyed),
      replaced: ($replaced | map("  - \(.)")),
      destroyed: ($destroyed | map("  - \(.)"))
    }
' "$PLAN_JSON")"

BLOCKED="$(jq -r '.blocked | join("\n")' <<<"$VERDICT")"
MISSING_REPLACE="$(jq -r '.missing_replace | join("\n")' <<<"$VERDICT")"
MISSING_DELETE="$(jq -r '.missing_delete | join("\n")' <<<"$VERDICT")"
REPLACED="$(jq -r '.replaced | join("\n")' <<<"$VERDICT")"
DESTROYED="$(jq -r '.destroyed | join("\n")' <<<"$VERDICT")"

if [ -n "$BLOCKED" ] || [ -n "$MISSING_REPLACE" ] || [ -n "$MISSING_DELETE" ]; then
  if [ -n "$BLOCKED" ]; then
    echo "plan-guard: BLOCKED — this plan would delete and/or replace the following resource(s):" >&2
    echo "$BLOCKED" >&2
  fi
  if [ -n "$MISSING_REPLACE" ]; then
    echo "plan-guard: BLOCKED — --allow-replace named address(es) this plan does not replace:" >&2
    echo "$MISSING_REPLACE" >&2
  fi
  if [ -n "$MISSING_DELETE" ]; then
    echo "plan-guard: BLOCKED — --allow-delete named address(es) this plan does not purely destroy:" >&2
    echo "$MISSING_DELETE" >&2
  fi
  echo "" >&2
  echo "Do not apply this plan unattended. Review it manually, confirm the destroy/replace is" >&2
  echo "intentional, and only then apply outside of this guarded path." >&2
  discard_kept_plan
  exit 1
fi

if [ -n "$REPLACED" ] || [ -n "$DESTROYED" ]; then
  if [ -n "$REPLACED" ]; then
    echo "plan-guard: PASS — plan replaces only the explicitly allowed resource(s):"
    echo "$REPLACED"
  fi
  if [ -n "$DESTROYED" ]; then
    echo "plan-guard: PASS — plan destroys only the explicitly allowed resource(s):"
    echo "$DESTROYED"
  fi
else
  echo "plan-guard: PASS — plan is additive/update-only (no delete or replace actions)"
fi
[ -z "$OUT" ] || echo "plan-guard: guarded plan kept at ${PLAN_BIN} — apply it with: terraform apply ${PLAN_BIN}"
exit 0
