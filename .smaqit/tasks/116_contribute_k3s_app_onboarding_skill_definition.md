---
status: Completed
created: "2026-09-11"
mode: Assisted
started: "2026-09-12"
completed: "2026-09-12"
---

# Contribute a k3s App-Onboarding Skill Definition

## Description

**Corrected mid-review (2026-09-12) — see Notes for the full reassessment.** The original framing
of this task was wrong: it took one specific downstream infra repo's onboarding implementation
(a git-committed per-machine registry file plus a converge GitHub Actions workflow, with its own
Namespace/RBAC/PSA/NetworkPolicy/ResourceQuota specifics and its own six real bugs) and shipped it
as a smaqit product skill as if that implementation were a universal contract every k3s-onboarding
infra repo should follow. That inverts the machine-monorepo pattern's own principle: **each
infra-owning repo defines and owns its own onboarding contract.** That knowledge is repo-specific,
not something smaqit should centralize, prescribe, or hardcode.

This task now ships `smaqit.infrastructure-onboard-k3s-app` as a **thin dispatcher**, not a
mechanism: when a project's Infrastructure spec targets a k3s cluster owned by another
(infrastructure-owning) repository, this skill recognizes that hand-off point and defers entirely
to that repo's own onboarding skill, workflow, or instructions — never assuming or replicating its
registry format, RBAC scheme, or workflow shape. The detailed mechanism material originally
authored here (the specific downstream implementation's mechanics and its six gotchas) has been
deleted from smaqit entirely — it was one org's own implementation detail, not smaqit's to own.

**Still explicitly out of scope:** stack-detection routing and wiring into
`smaqit.new-greenfield-project`/`smaqit.feature-new` as a selectable deployment target.

## Issue Triage Context

**Mode:** Skip
**Technologies:** GitHub Actions/repo conventions, smaqit skill-definition conventions
**Platforms/Environments:** None — this skill contains no cluster-specific mechanics at all
**Features/Integrations:** `installer/main_test.go` (skill-count assertions), `make -C installer prepare`/`test`/`smoke-test`, `smaqit.infrastructure-vault-loader` (where the deferred-to credential eventually lands)
**Versions/Constraints:** Compiled `skills/smaqit.infrastructure-onboard-k3s-app/` skill only — no definitions-only source file (deleted, per Design Decisions); no stack-detection routing or greenfield/feature-new wiring; no infra-repo-specific mechanics anywhere in the shipped skill

## Design Decisions

- **Thin dispatcher, not a mechanism.** The skill's entire content is: recognize that the target
  k3s cluster is owned by another repo, and defer to that repo's own onboarding process. It never
  assumes a registry format, an RBAC scheme, a workflow shape, or any other infra-repo-specific
  detail — those vary per infra repo and are that repo's own contract to define, never smaqit's.
- **No definitions-only source file.** The original detailed mechanism (one specific downstream
  repo's real implementation, including its own six gotchas) has been deleted from smaqit
  entirely — `.smaqit/definitions/skills/smaqit.infrastructure-onboard-k3s-app.md` no longer
  exists. It documented one org's implementation detail, not a generalizable smaqit contract; it
  had no business being shipped, contributed, or kept as reference material here.
- **No `validated`/`validated-stack` metadata.** Those fields signal "validated against a specific
  downstream stack," which doesn't apply to a generic, infra-repo-agnostic dispatcher — matches the
  frontmatter convention already used by process/config skills like
  `smaqit.infrastructure-repo-config` and `smaqit.infrastructure-vault-loader` (plain
  `metadata.version` only).
- **Compiled directly** (no separate definitions-first period) — the design is narrow and
  low-risk enough not to need a synthesis/proof period.
- **Routing/wiring into `smaqit.new-greenfield-project`/`smaqit.feature-new` remains out of scope**
  for this task — this skill only recognizes the hand-off point; actually wiring it into the phase
  flow as a selectable target is separate, larger work.
- **A cross-session design conflict was found and deliberately left unresolved here.** Task 118
  (child of 117, a concurrent session's work) was designed against this task's *original* detailed
  version, treating it as a legitimate "platform operator's direct-access playbook" that task 118's
  own PR-based, no-direct-access alternative complements. That assumption no longer holds now that
  the detailed mechanism is deleted. The user has taken ownership of reconciling task 117/118's
  design and implementation separately — not addressed by this task.

## Implementation Steps

1. Delete `.smaqit/definitions/skills/smaqit.infrastructure-onboard-k3s-app.md` — the detailed,
   one-org-specific mechanism has no place in smaqit.
2. Rewrite `skills/smaqit.infrastructure-onboard-k3s-app/SKILL.md` as a thin dispatcher:
   - **Description** (frontmatter) — when to use: a project's Infrastructure spec targets a k3s
     cluster owned by another repo; this skill recognizes that and defers, never prescribes.
   - **Pre-conditions** — the target cluster is owned by a different repo; that repo is
     identifiable (from the Infrastructure spec or the operator).
   - **Steps** — identify the infra repo; locate its own onboarding mechanism (skill, workflow, or
     docs) without assuming its shape; follow it exactly as that repo defines it; store whatever
     credential it issues via this project's own `smaqit.infrastructure-vault-loader` conventions.
   - **Scope** — does not prescribe or replicate any infra repo's mechanics; does not provision the
     cluster; does not deploy this app's own workloads.
   - **Failure Handling** — the infra repo's mechanism can't be located; it issues no usable
     credential.
3. Update the `CHANGELOG.md` entry (already promoted to `## [3.4.0]` on this branch) to describe
   the thin dispatcher, not the deleted mechanism.
4. Re-run `make -C installer prepare`, `make -C installer test`, and `make -C installer smoke-test`
   to confirm the rewritten skill still installs and tests cleanly (no change to the skill-count
   assertions — still one new skill directory).
5. Grep the rewritten skill for any literal project, repository, or machine name — none should
   exist at all in the thin version.
6. Push the correction to the existing `task/116-...` branch — PR #87 stays open, updated in place,
   not closed/reopened.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [x] `.smaqit/definitions/skills/smaqit.infrastructure-onboard-k3s-app.md` does not exist —
      deleted, no detailed mechanism content anywhere in smaqit
- [x] `skills/smaqit.infrastructure-onboard-k3s-app/SKILL.md` exists as a thin dispatcher: no
      registry schema, RBAC/PSA/NetworkPolicy specifics, or gotchas tied to any specific infra
      repo's implementation — only recognition-and-defer logic
- [x] Zero real project, repository, or machine names anywhere in the skill
- [x] `metadata` carries only `version` (no `validated`/`validated-stack`), matching the
      process/config skill convention
- [x] `installer/main_test.go`'s two hardcoded skill-count assertions remain at 28 and
      `make -C installer test` passes
- [x] `make -C installer smoke-test` passes, confirming the skill installs cleanly
- [x] No changes are made to `smaqit.input-deployment`, `smaqit.new-greenfield-project`,
      `smaqit.feature-new`, or `smaqit.infrastructure-cicd-generate` — routing/wiring remains a
      separate follow-up task

## Findings

**Implementation approach:**
- First pass (superseded): authored the definitions file and compiled skill as a detailed mechanism mirroring one specific downstream repo's real onboarding implementation (registry schema, RBAC/PSA/NetworkPolicy specifics, six gotchas). PR #87 opened on this basis.
- During PR review, the user identified this as the wrong design: a generic smaqit product skill cannot prescribe one org's specific onboarding mechanics as if universal — each infra-owning repo owns its own contract. Corrected in place on the same branch: deleted the definitions file entirely, rewrote the compiled skill as a thin dispatcher (identify the infra repo → locate its own onboarding mechanism → follow it as defined → store the resulting credential via `smaqit.infrastructure-vault-loader`), dropped `validated`/`validated-stack` metadata (doesn't apply to a generic dispatcher), and rewrote the `CHANGELOG.md` entry to match.
- Re-ran `make -C installer prepare`, `make -C installer test`, and `make -C installer smoke-test` after the rewrite — all pass. No change needed to the skill-count assertions (still one new skill directory).

**Decisions made:**
- Detailed mechanism content is deleted from smaqit entirely, not kept as reference material — it documented one org's implementation, not a generalizable pattern.
- The existing branch/PR was corrected in place rather than closed and reopened.
- A cross-session design conflict was discovered (task 118, built by a concurrent session, assumed this task's *original* detailed version was a legitimate "direct-access playbook" it complements) and deliberately left unresolved by this task — the user is reconciling task 117/118 separately.

**Blockers encountered:**
- The task was materially misscoped on its first pass, caught only during PR review rather than during planning — see Implementation approach.

**Follow-up identified:**
- Task 117/118's cross-session design needs reconciliation now that this task's detailed mechanism no longer exists for task 118 to reference — owned by the user, not tracked as a new task here.

## Files to Create / Modify

| File | Action |
|------|--------|
| `.smaqit/definitions/skills/smaqit.infrastructure-onboard-k3s-app.md` | Delete — created in the superseded first pass, removed on correction |
| `skills/smaqit.infrastructure-onboard-k3s-app/SKILL.md` | Rewrite — thin dispatcher, replacing the superseded detailed-mechanism version |
| `installer/main_test.go` | Unchanged from first pass — still bumped 27→28 |
| `docs/wiki/workflows/testing-smaqit.md` | Unchanged from first pass — still bumped |
| `CHANGELOG.md` | Rewrite the `[3.4.0]` entry to describe the thin dispatcher |

## Notes

**Reassessment (2026-09-12, during PR review):** the original task and its first implementation
pass were wrong. They took one specific downstream infra repo's real onboarding implementation
(registry file + converge workflow, with that repo's own RBAC/PSA/NetworkPolicy specifics and six
real bugs) and shipped it as a smaqit product skill as though it were a universal contract. That
inverts the machine-monorepo pattern's own principle that each infra-owning repo defines and owns
its own onboarding contract. Corrected to a thin dispatcher that only recognizes the hand-off point
and defers to the infra repo's own process — see Design Decisions and Findings for the full
correction. The detailed mechanism content is deleted from smaqit entirely, not preserved as
reference material.

A related, unresolved consequence: task 118 (a concurrent session's work, child of task 117) was
designed assuming this task's *original* detailed version was a legitimate, complementary
"direct-access playbook." That assumption no longer holds. Reconciling task 117/118 is the user's
own follow-up, not part of this task.
