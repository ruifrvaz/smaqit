---
status: In Progress
created: "2026-10-14"
mode: Assisted
started: "2026-10-18"
---

# plan-guard.sh: `--allow-delete <address>` for a Reviewed Pure Destroy

## Description

`plan-guard.sh` (`smaqit.infrastructure-provision-cyso`) blocks every planned delete. Its one
exception is a replacement named with `--allow-replace <address>` (task 115). A guarded repository
has no way to apply a deliberate, reviewed removal of a single resource, such as:

- a downstream platform repository closing a break-glass firewall opening it had added (the open is
  a create, but the close is a pure destroy of that security-group rule);
- the same repository retiring a public port (removing one ingress rule);
- any other one-resource removal a reviewed PR asks for.

Today such a change can only be applied by going around the guard, which the deployment agent
forbids ("every plan/apply goes through the project's guard script"). The `teardown` slot covers
destroying a whole target, not one resource.

The fix is to add `--allow-delete <address>`, the mirror of `--allow-replace`: accept a planned pure
destroy of exactly the listed address, and keep blocking everything else.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Terraform, jq, Bash, Go
**Platforms/Environments:** None
**Features/Integrations:** plan-guard.sh, smaqit.infrastructure-provision-cyso, terraform show -json resource_changes, platform-repositories guide
**Versions/Constraints:** Terraform 1.15.x JSON plan format, guard must stay fail-closed, repeatable flags, no change to --allow-replace semantics

## Design Decisions

- **Exact address, pure destroy only:** `--allow-delete <address>` accepts a `resource_changes` entry
  whose actions are exactly `["delete"]` at that address. Repeatable.
- **Fail closed:**
  - a listed address the plan replaces instead of destroying → BLOCKED;
  - a listed address the plan doesn't touch → BLOCKED (a mistyped address must not pass silently);
  - every unlisted delete or replace → BLOCKED, as today.
- **No `-replace=` for `--allow-delete`:** unlike `--allow-replace`, it adds nothing to the
  `terraform plan` arguments. It only widens what the inspection accepts.
- **Combinable** with `--allow-replace` and `--out`; the PASS message names each allowed delete and
  each allowed replacement.
- **One address, one intent:** the same address given to both `--allow-delete` and `--allow-replace`
  is a usage error (exit 2), not a choice the guard makes.
- **No-change plans:** a plan with no changes and any `--allow-delete` given → BLOCKED, the same as
  `--allow-replace` today (nothing was destroyed, so the address is wrong or already gone).
- **Out of scope:** `forget` (a `removed` block with `destroy = false`) is not a delete action and is
  never matched by the guard's delete filter; `--allow-delete` neither needs nor affects it.

## Implementation Steps

1. `skills/smaqit.infrastructure-provision-cyso/scripts/plan-guard.sh`:
   - parse a repeatable `--allow-delete`;
   - extend the jq verdict so a pure-`delete` entry at a listed address is allowed;
   - report a listed address that wasn't destroyed (`missing`) and one that was replaced instead;
   - block on a no-change plan when `--allow-delete` is given; reject an address in both lists;
   - make the BLOCKED messages name the flag that was not satisfied (today they say `--allow-replace` only);
   - update the usage, the header comment and the exit-code text.
2. Regenerate installer staging with `make -C installer prepare`. `installer/skills-shared/` and
   `installer/skills-claude/` are gitignored build output and are never edited or copied into by hand.
3. `installer/skill_scripts_test.go`, next to `TestPlanGuardAllowsOnlyTheNamedReplacement`, test:
   - the allowed pure delete passes;
   - an unlisted delete is blocked;
   - a listed address that is replaced is blocked;
   - a listed address absent from the plan is blocked;
   - `--allow-delete` combined with `--allow-replace` and `--out`;
   - a no-change plan with `--allow-delete` is blocked;
   - bad usage (`--allow-delete` with no value; one address in both lists).
4. `skills/smaqit.infrastructure-provision-cyso/SKILL.md` and
   `docs/wiki/workflows/platform-repositories.md` (Governance, "Every plan goes through the guard"):
   document `--allow-delete` beside `--allow-replace`, with the pattern that a gated dispatch takes
   the address as an input.
5. CHANGELOG and release.

## Known Issues Triage
**Triaged:** 2026-10-18
**Tools searched:** Terraform (`hashicorp/terraform`), jq (`jqlang/jq`)
**Result:** Advisory

### Advisory Issues
- [#29618 JSON plan output gives less sensitivity detail than CLI for blocks backed by sets](https://github.com/hashicorp/terraform/issues/29618) — `hashicorp/terraform` — opened 2021-09-21 — bug, confirmed. Affects `sensitive_values` detail only, not `resource_changes[].change.actions`, which the guard reads.
- [#33112 Resource block plan detail is hidden due to sensitive values since 1.4](https://github.com/hashicorp/terraform/issues/33112) — `hashicorp/terraform` — opened 2023-04-28 — bug, cli. Human-readable plan rendering only; the JSON actions array is unaffected.

### Historical (Closed)
- [#38641 JSON Plan: resources being removed from state ('forget') are included in planned_values](https://github.com/hashicorp/terraform/issues/38641) — `hashicorp/terraform` — closed 2026-06-04. A `removed` block with `destroy = false` plans the action `forget`, not `delete`. It is outside the guard's delete filter and outside `--allow-delete`'s scope (not a destroy).

### Unresolvable Tools
- Bash — GNU Bash is not hosted on GitHub; the registry lookup matched an unrelated repository (`dylanaraps/pure-bash-bible`)

### Omitted Tools
- Go — test harness only (`testing`, `os/exec`); no task feature dimension to search in `golang/go`

## Acceptance Criteria

- [ ] `plan-guard.sh --allow-delete <address>` passes a plan whose only delete is a pure destroy of
      that address, and blocks unlisted deletes, a listed address that is replaced, and a listed
      address the plan doesn't destroy
- [ ] `--allow-delete` is repeatable and combines with `--allow-replace` and `--out`; without it,
      behaviour is unchanged (existing guard tests still pass)
- [ ] Go tests cover every case above; `go test ./...` passes
- [ ] The provision-cyso SKILL.md and the platform-repositories guide document the flag
- [ ] Released, with a CHANGELOG entry

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- TBD

**Decisions made:**
- TBD

**Blockers encountered:**
- TBD

**Follow-up identified:**
- TBD

## Files to Create / Modify

| File | Action |
|------|--------|
| `skills/smaqit.infrastructure-provision-cyso/scripts/plan-guard.sh` | Modify |
| `installer/skill_scripts_test.go` | Modify |
| `skills/smaqit.infrastructure-provision-cyso/SKILL.md` | Modify |
| `docs/wiki/workflows/platform-repositories.md` | Modify |
| `CHANGELOG.md` | Modify |

## Notes

- Found in a downstream platform repository's slot-conformance audit (2026-10-14). It blocks that
  repository's break-glass close and a planned public-port retirement. Its consuming task re-vendors
  the guard and adds an `allow_delete` workflow input once this releases.
- The copy in `installer/skills-shared/` differed from `skills/` on 2026-10-14. That was stale
  build output: `skills/` is the source, and `make prepare` regenerates both staging trees from it.
