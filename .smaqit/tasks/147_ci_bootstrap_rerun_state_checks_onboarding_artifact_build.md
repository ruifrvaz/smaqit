---
status: Not Started
created: "2026-10-25"
---

# ci-bootstrap Re-run on an Onboarded existing-k3s App: State Checks for Onboarding and Artifact Build

## Description

Regenerating an already-onboarded existing-k3s app's `deploy.yml` (for example after an Infrastructure
spec amendment such as task 145's move to the CI identity) requires `ci-bootstrap`: it is the only
existing-k3s path whose defaults run `ci-generate` (`existing-cicd-pr` refuses it, `ci-push` defaults
it to none). But `ci-bootstrap` also re-runs the first-deploy slots, and neither checks whether its
work is already done:

- `platform-onboarding` (`smaqit.infrastructure-request-k3s-onboarding` `request`): Step 2 detects only
  an **open** PR on `onboard-<slug>-<machine>`. Once that PR has merged, a re-run recreates the branch,
  re-appends the entry to the registry file and opens a duplicate PR. Only the operator having to
  supply `entry_content` (and the credential) prevents it.
- `artifact-build` (family `smaqit.infrastructure-image-build-push*`): the only shipped candidate is the
  static-site skill. An app on another stack that already has its own `Dockerfile` and
  `deployment/k3s/` manifests has no "already authored" or "stack not covered" outcome, only the
  candidate's "confirm before overwriting".

Found by a downstream platform migrating its test deploys to task 145's CI identity (smaqit v3.21.0).

## Issue Triage Context

**Mode:** Auto
**Technologies:** GitHub CLI, GitHub Actions, Kubernetes (k3s), Go, Bash
**Platforms/Environments:** None
**Features/Integrations:** deployment slots, ci-bootstrap, request-k3s-onboarding, image-build-push family, slot contracts
**Versions/Constraints:** smaqit v3.21.0, breaking fix with no opt-out flag, no project `.smaqit/slots.yml` override required

## Design Decisions

- **State checks, not a new path:** the affected slots check state first and record "already
  satisfied" — the "every operation checks first and skips what is done" rule — rather than adding a
  re-bootstrap path or letting `ci-push` regenerate workflows. Every re-run becomes safe with no extra
  path to choose. `ci-generate` still runs and regenerates `deploy.yml`.
- **Breaking fix, no opt-out:** follows the no-backwards-compatibility principle; consuming projects need
  no `.smaqit/slots.yml` override.
- **Schema-free onboarding detection:** the onboarding skill still never interprets the registry schema;
  the "already onboarded" check is a merged PR on the onboarding branch, or a plain text match of this
  repository's numeric ID (printed in Step 1) in the registry file on the platform's default branch.
- **Overlap with task 137:** both edit request-k3s-onboarding's credential and early steps (137 retires
  the platform-repo PAT in favour of the operator's own `gh` login). The read-only state check uses the
  operator's own `gh` login, consistent with 137; whichever task merges second rebases onto the other.

## Implementation Steps

1. **Slot contracts (`installer/slots.go`):** add a "satisfied when" clause to `slotDef`, printed by
   `smaqit slots contract` and mirrored into the wiki catalogue by the existing equality test
   (`TestSlotContractsMatchWiki`):
   - `platform-onboarding` (existing-k3s): satisfied when the onboarding is already on the platform's
     default branch;
   - `artifact-build` (existing-k3s): satisfied when `Dockerfile`, `deployment/k3s/deployment.yaml` and
     `service.yaml` exist and pass deploy-k3s-app's `manifest-lint.sh`.
2. **Deployment agent (`agents/deployment.md`):** before running a local slot that has a "satisfied
   when" clause, evaluate it; if it holds, record `already satisfied` with its evidence and skip the
   filler. If `artifact-build`'s artifacts are missing and no family candidate fits the declared stack,
   return `NEEDS-INPUT` instead of offering a candidate that doesn't fit.
3. **request-k3s-onboarding:** a state check before `entry_content` is requested and before any write
   credential is loaded — an open PR goes to the Step 5 gate (as today); a merged PR on `$BRANCH`, or
   the registry file on the default branch containing this repository's numeric ID, means satisfied.
   Only otherwise ask for the entry.
4. **Tests:** contract text and the wiki equality test; a resolver or agent-text test for the skip;
   stub-`gh` cases for the open-PR, merged-PR and ID-on-default-branch states.
5. Definitions first (`.smaqit/definitions/skills/`), then skill text by hand; then
   `docs/wiki/workflows/deployment-slots.md`, `.smaqit/compendium.md`, CHANGELOG, and release.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] `smaqit slots contract` prints a "satisfied when" clause for `platform-onboarding` and
      `artifact-build` (existing-k3s), and the wiki catalogue is kept equal by test
- [ ] The Deployment agent skips a satisfied local slot and records its evidence; `artifact-build` with
      no fitting candidate and no artifacts returns `NEEDS-INPUT`
- [ ] request-k3s-onboarding detects a merged PR or the repository's numeric ID on the default branch
      before asking for `entry_content` or loading any write credential; stub-`gh` tests cover each state
- [ ] A `ci-bootstrap` re-run on an onboarded app opens no platform PR, overwrites no `Dockerfile` or
      manifests, and regenerates `deploy.yml`
- [ ] No downstream names in canonical text; no opt-out flag; wiki, compendium and CHANGELOG updated;
      released

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
| installer/slots.go, installer/slots_test.go | Modify |
| agents/deployment.md | Modify |
| skills/smaqit.infrastructure-request-k3s-onboarding/SKILL.md, .smaqit/definitions/skills/smaqit.infrastructure-request-k3s-onboarding.md | Modify |
| installer/skill_scripts_test.go (stub-gh cases) | Modify |
| docs/wiki/workflows/deployment-slots.md, .smaqit/compendium.md, CHANGELOG.md | Modify |

## Notes

- Reported by a peer session working in a downstream platform repository, 2026-10-25. That platform
  worked around it by answering the onboarding prompt with "already onboarded".
- See task 137 for the concurrent change to request-k3s-onboarding's authentication.
