---
status: PR Open
pr: 107
created: "2026-10-18"
mode: Assisted
started: "2026-10-18"
---

# plan-guard BLOCKED Guidance: Stop and Report, Never Take an Allow-Flag Address From the Guard's Output

## Description

When `plan-guard.sh` blocks a plan, every piece of guidance a coding agent reads at that moment
points the wrong way:

1. **The guard's own BLOCKED message** ends with "Review it manually, confirm the destroy/replace is
   intentional, and only then apply outside of this guarded path." That is the text an agent reads at
   the moment it decides what to do next, and it suggests bypassing the guard.
2. **The provision-cyso failure table** (added in task 135, v3.15.0) says a reviewed removal can
   "re-run with `--allow-delete <address>`", without saying where that address may come from.
3. **The deployment agent's guard invariant** requires every plan/apply to go through the guard, but
   says nothing about what to do when the guard blocks.

Together they let an agent turn a blocked side-effect delete or replacement into an allowed one: copy
the address from the BLOCKED output into `--allow-delete` or `--allow-replace` and re-run. The guard
verifies that the plan matches the flags, not that the flags match anyone's intent, so it passes.
The same gap has applied to `--allow-replace` since task 115.

This task fixes the wording only. A human-confirmation gate (terminal confirmation, break-glass, or
CI-only applies for Environment targets) was assessed on 2026-10-18 and not adopted: downstream
projects guard their own apply paths.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Terraform, Bash
**Platforms/Environments:** None
**Features/Integrations:** plan-guard.sh BLOCKED message, smaqit.infrastructure-provision-cyso failure handling, deployment agent guard invariant, smaqit.infrastructure-cicd-generate gated dispatch input, platform-repositories guide
**Versions/Constraints:** Terraform 1.15.x, wording only, no change to the guard's verdict logic, flags or exit codes

## Design Decisions

- **One rule, stated everywhere an agent reads it:** on BLOCKED, stop and report the plan; an
  allow-flag address must be one a human named before the plan ran (the request, the reviewed PR, a
  gated dispatch input), never one taken from the guard's output.
- **The guard's message carries the rule**, because it is the only text guaranteed to be in front of
  the agent at the decision point. It must no longer suggest applying outside the guard.
- **Both flags:** the rule covers `--allow-replace` as well as `--allow-delete`.
- **Wording only:** the verdict logic, the flags and the exit codes are unchanged; existing guard tests
  must pass untouched apart from the message assertion added here.
- **Patch release:** no behaviour change.

## Implementation Steps

1. `skills/smaqit.infrastructure-provision-cyso/scripts/plan-guard.sh`: replace the BLOCKED footer
   ("Do not apply this plan unattended … apply outside of this guarded path") with the stop-and-report
   rule, naming both allow flags.
2. `skills/smaqit.infrastructure-provision-cyso/SKILL.md`: state the rule in the guard step and in the
   two failure-table rows; bump the skill's patch version.
3. `agents/deployment.md`: add the rule to the guard invariant (every plan/apply through the guard).
4. `skills/smaqit.infrastructure-cicd-generate/SKILL.md`: the gated "replace/remove this resource"
   inputs take a human-typed address only; bump the skill's patch version.
5. `docs/wiki/workflows/platform-repositories.md` and `.smaqit/compendium.md`: one sentence each.
6. `installer/skill_scripts_test.go`: assert the BLOCKED output carries the stop rule and no longer
   says "outside of this guarded path".
7. `make -C installer prepare`, `make test`, `make smoke-test`; CHANGELOG.

## Known Issues Triage
**Triaged:** 2026-10-18
**Tools searched:** Terraform (`hashicorp/terraform`, session-cached result from task 135's triage)
**Result:** Advisory

### Advisory Issues
- [#29618 JSON plan output gives less sensitivity detail than CLI for blocks backed by sets](https://github.com/hashicorp/terraform/issues/29618) — `hashicorp/terraform` — opened 2021-09-21 — bug, confirmed. Sensitivity detail only; does not affect the actions the guard reads or the message it prints.
- [#33112 Resource block plan detail is hidden due to sensitive values since 1.4](https://github.com/hashicorp/terraform/issues/33112) — `hashicorp/terraform` — opened 2023-04-28 — bug, cli. Human-readable rendering only.

### Unresolvable Tools
- Bash — GNU Bash is not hosted on GitHub; the registry lookup matched an unrelated repository

## Acceptance Criteria

- [x] The guard's BLOCKED output tells the reader to stop and report, forbids taking an allow-flag address from that output, and no longer suggests applying outside the guard
- [x] provision-cyso, cicd-generate and the deployment agent state the same rule, for both `--allow-delete` and `--allow-replace`
- [x] A Go test asserts the BLOCKED wording; all existing guard tests pass unchanged
- [x] No change to the guard's verdict logic, flags or exit codes
- [x] Released as a patch, with a CHANGELOG entry

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- Replaced the guard's BLOCKED footer ("only then apply outside of this guarded path") with a stop-and-report rule naming both allow flags; the script diff touches only `echo` lines.
- Stated the same rule in provision-cyso (a "Where an allowed address comes from" paragraph using the `count`-index trap as its example, plus both failure-table rows), the Deployment agent's guard invariant, cicd-generate's gated replace/remove inputs, the platform-repositories guide and the compendium.
- Added `TestPlanGuardBlockedOutputSaysStopAndReport`, covering an unlisted delete and a listed address the plan doesn't destroy; restoring the old footer makes it fail. All other guard tests pass unchanged.

**Decisions made:**
- Wording only: a human-confirmation gate (terminal confirmation, break-glass, CI-only applies for Environment targets) was assessed and not adopted; downstream projects guard their own apply paths.
- The rule covers `--allow-replace` as well as `--allow-delete`, since the same gap has existed since task 115.
- Skill patch bumps: provision-cyso 1.10.0 → 1.10.1, cicd-generate 3.2.0 → 3.2.1.

**Blockers encountered:**
- None.

**Follow-up identified:**
- Deferred by the user: CI-only applies for GitHub Environment targets (credential placement, separate test state, OIDC where the provider supports it).
- Wording guides an agent but cannot stop one that ignores it; only credential placement would make local applies to Environment targets impossible.

## Files to Create / Modify

| File | Action |
|------|--------|
| `skills/smaqit.infrastructure-provision-cyso/scripts/plan-guard.sh` | Modify |
| `skills/smaqit.infrastructure-provision-cyso/SKILL.md` | Modify |
| `agents/deployment.md` | Modify |
| `skills/smaqit.infrastructure-cicd-generate/SKILL.md` | Modify |
| `docs/wiki/workflows/platform-repositories.md` | Modify |
| `.smaqit/compendium.md` | Modify |
| `installer/skill_scripts_test.go` | Modify |
| `CHANGELOG.md` | Modify |

## Notes

- Found on 2026-10-18 while assessing a Bench suite for task 135's `--allow-delete`. A live Terraform
  probe showed the count-index trap: removing one element of a `count` list plans a delete of the
  *last* index, so the BLOCKED output names an address nobody asked to remove.
- Out of scope, deferred by the user: CI-only applies for GitHub Environment targets (credential
  placement, separate test state, OIDC where the provider supports it).
