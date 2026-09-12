---
status: PR Open
created: "2026-09-11"
mode: Assisted
started: "2026-09-12"
pr: 87
---

# Contribute a k3s App-Onboarding Skill Definition

## Description

A downstream project's app-agnostic infrastructure repo (self-hosted, single-server k3s per
machine) has a mature, live-verified app-onboarding mechanism, hardened across four rounds of real
use: a git-committed per-machine registry file plus a converge GitHub Actions workflow that, for
every registered app slug, creates a Namespace, least-privilege RBAC (ServiceAccount/Role/
RoleBinding), PSA `restricted` enforcement, a default-deny NetworkPolicy, a ResourceQuota/
LimitRange, and issues a scoped long-lived kubeconfig handed off via a workflow artifact plus an
external secrets store.

This task turns that contributed pattern into a proper Skill Definition file at
`.smaqit/definitions/skills/smaqit.infrastructure-onboard-k3s-app.md`, following the exact structural
precedent of `.smaqit/definitions/skills/smaqit.infrastructure-deploy-rsync-python-tornado.md`, and
then compiles and ships it as a real, standalone-invocable product skill at
`skills/smaqit.infrastructure-onboard-k3s-app/` (frontmatter, global payload inclusion, installer
tests, docs) — unlike the tornado precedent's initial single-synthesis state, this mechanism is
already hardened across four rounds of real production use, so there is no reason to withhold
compilation pending further validation.

**Explicitly out of scope even after compilation:** stack-detection routing and wiring into
`smaqit.new-greenfield-project`/`smaqit.feature-new` as a selectable deployment target. This skill
performs app *onboarding* (Namespace/RBAC/kubeconfig issuance for a tenant of an existing cluster)
— a fundamentally different, cluster/machine-repo-side concern from an app's own *deployment*
(building and applying the app's own workloads using the issued kubeconfig, analogous to the
`smaqit.infrastructure-deploy-rsync*` family). No hardened source material covers that app-side
deploy mechanism or the new `deployment_target_type` input it would require; it is a separate,
future task requiring its own design work, not a mechanical follow-up like compilation is here.

## Issue Triage Context

**Mode:** Skip
**Technologies:** Kubernetes RBAC/PodSecurity/NetworkPolicy/ResourceQuota, k3s, GitHub Actions (`workflow_dispatch`, Environment protection rules), smaqit skill-definition conventions
**Platforms/Environments:** None — this task produces a definition file only, no live infrastructure is touched
**Features/Integrations:** `.smaqit/definitions/skills/` (definition authoring), `smaqit.infrastructure-deploy-rsync-python-tornado` (structural precedent, both definitions-only and compiled forms), `installer/main_test.go` (skill-count assertions), `make -C installer prepare`/`test`/`smoke-test`
**Versions/Constraints:** Definition file plus compiled `skills/smaqit.infrastructure-onboard-k3s-app/` skill, both shipped by this task; no stack-detection routing or greenfield/feature-new wiring; Provenance must not name the real downstream project, repo, or machines

## Design Decisions

- **Mirror `smaqit.infrastructure-deploy-rsync-python-tornado.md`'s exact section shape**
  (Description, Provenance, Steps [Pre-conditions + Steps], Output, Scope, Completion, Failure
  Handling, Gotchas, Allowed Tools, Examples) rather than inventing a new definition-file
  structure.
- **Provenance is anonymized**, matching that same reference file's pattern (no real project/repo/
  machine name, anywhere — including in this task's own Notes/Description), and framed as a
  deliberate mature contribution (hardened across four rounds of real production use) rather than
  an ad-hoc single-use dev-sweep synthesis.
- **Two known gaps in the source mechanism — RBAC read access to `pods/log`, and non-destructive
  in-place credential rotation — are noted as not yet incorporated**, not blockers; the contributed
  mechanism is complete and proven without them.
- **Compilation happens in this task, not a follow-up.** Unlike the tornado skill (contributed
  unproven, explicitly held back pending real-world validation), this mechanism is already
  hardened across four rounds of production use — withholding compilation would serve no purpose.
  Compilation follows the tornado skill's own documented before/after shape: add YAML frontmatter
  (`name`, `description`, `metadata.version`/`validated`/`validated-stack`), promote Pre-conditions
  to a top-level `##` heading, and fold `Provenance`/`Required-inherited-context` into
  `metadata`/inline Steps prose rather than keeping them as standalone sections (this skill has no
  shared-family "required-inherited-context" to begin with, since it isn't part of the
  `deploy-rsync*` family).
- **Routing/wiring remains a separate follow-up task.** Compiling this skill makes it
  standalone-invocable; it does not make it a selectable alternative inside
  `smaqit.new-greenfield-project`/`smaqit.feature-new`. That requires a new `deployment_target_type`
  input, a Phase 4 Step 6-equivalent branch point, dedicated CI/CD workflow generation (this
  skill's registry+converge shape doesn't fit `smaqit.infrastructure-cicd-generate`'s VM/SSH/rsync
  templates — it is structurally closer to task 115's `smaqit.infrastructure-tenant-reconcile`
  pattern), and a net-new app-side deploy mechanism with no hardened source material behind it.
  None of that is mechanical the way compilation is, so it stays out of this task.

## Implementation Steps

1. Read `.smaqit/definitions/skills/smaqit.infrastructure-deploy-rsync-python-tornado.md` in full
   as the structural reference, and `.smaqit/tasks/106_reconcile_python_tornado_rsync_deployment_skill.md`
   for how a prior contribution-to-canonical task was scoped and worded.
2. Author `.smaqit/definitions/skills/smaqit.infrastructure-onboard-k3s-app.md`:
   - **Description** — when to use this skill: per-app Namespace/RBAC/PSA/NetworkPolicy/Quota
     onboarding onto a self-hosted k3s cluster via a registry file + converge workflow.
   - **Provenance** — anonymized (e.g. `synthesized-for-project: [a downstream project]`), framed
     as a mature, multi-round-hardened contribution, not a one-off synthesis.
   - **Steps** (Pre-conditions + Steps) — the registry schema (opaque app slugs), the converge
     workflow's mechanics (Namespace, least-privilege Role/RoleBinding/ServiceAccount, PSA
     `restricted`, default-deny NetworkPolicy, ResourceQuota/LimitRange, a fail-closed token/CA
     wait before assembling the kubeconfig), and the credential handoff (workflow artifact +
     external secrets store, never committed).
   - **Output / Scope** — explicitly not cluster provisioning (a separate provisioning skill's
     job) and not the onboarded app's own deployment.
   - **Gotchas** — every real bug found in the source mechanism: a `while read`-loop-plus-SSH bug
     that silently consumed stdin and truncated multi-slug runs to just the first entry; a `ca.crt`
     jsonpath-quoting bug; the requirement for a fail-closed wait on token/CA availability before
     assembling a kubeconfig; a self-minted-token gap (a workload can mint its own extra token
     Secret under a broad `secrets: create` grant); the GitHub constraint that `workflow_dispatch`
     must already exist on the default branch before it can be dispatched at all, forcing an
     interim landing PR for a brand-new workflow file; and the discovery that a GitHub Environment
     deployment-branch policy applies per environment **name**, regardless of which job or trigger
     references it — silently blocking a PR-triggered read-only job that happens to share an
     environment name with a separately-gated apply job.
   - **Known gaps not yet incorporated** — read-only RBAC access to `pods/log` (debugging a
     container's log output), and non-destructive in-place credential rotation (recreating the
     onboarded ServiceAccount object to invalidate every previously-issued token via its UID,
     without touching the Namespace or the app's own workloads) — both are in-flight in the source
     project and deliberately not folded in yet.
   - **Completion / Failure Handling / Allowed Tools / Examples** — filled per the reference
     file's shape.
3. Grep the finished file for any literal project, repository, or machine name to confirm
   generalization and Provenance anonymization are both clean.
4. Compile `skills/smaqit.infrastructure-onboard-k3s-app/SKILL.md` from the definitions file:
   - Add YAML frontmatter (`name: smaqit.infrastructure-onboard-k3s-app`, a genericized
     `description`, `metadata.version`/`validated`/`validated-stack`).
   - Promote `Pre-conditions` from a nested subsection to its own top-level `##` heading, matching
     the compiled tornado skill's shape.
   - Drop `Provenance` and `Required-inherited-context` as standalone sections (this skill has no
     shared-family inherited context); fold their substance into `metadata` fields and inline
     Steps/Gotchas prose instead.
   - Keep Steps, Output, Scope, Gotchas, Completion, Failure Handling, Examples, Allowed Tools,
     generalizing any remaining synthesis-specific phrasing to steady-state descriptive language.
5. Run `make -C installer prepare` to regenerate `installer/skills-shared/` and
   `installer/skills-claude/` with the new skill directory.
6. Bump the two hardcoded skill-count assertions in `installer/main_test.go` from 27 to 28:
   `TestRemoveEmbeddedSkillDirsPreservesUnownedSharedContent` and
   `TestSharedSkillsServeCopilotAndCodex`. Bump the matching count in
   `docs/wiki/workflows/testing-smaqit.md` if it enumerates the current total.
7. Run `make -C installer test` (`go vet` + `go test`) and `make -C installer smoke-test` to
   confirm the new skill installs cleanly to the shared global path with its
   `[SMAQIT_SKILLS_DIR]` placeholder resolved.
8. Add a `CHANGELOG.md` entry for the new skill.
9. Do not touch stack-detection routing, `smaqit.input-deployment`, `smaqit.new-greenfield-project`,
   `smaqit.feature-new`, or `smaqit.infrastructure-cicd-generate` in this task — routing/wiring is
   a separate, future task (see Design Decisions).

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [x] `.smaqit/definitions/skills/smaqit.infrastructure-onboard-k3s-app.md` exists, matching
      `smaqit.infrastructure-deploy-rsync-python-tornado.md`'s section structure
- [x] Provenance section contains no real project, repository, or machine names
- [x] Documents every real bug and fix listed in Implementation Steps' Gotchas item, not just the
      final correct mechanism
- [x] Documents the `workflow_dispatch`-must-exist-on-the-default-branch constraint and the
      GitHub-Environment-branch-policy-applies-per-name gotcha
- [x] Notes `pods/log` RBAC access and non-destructive credential rotation as known gaps not yet
      incorporated
- [x] `skills/smaqit.infrastructure-onboard-k3s-app/SKILL.md` exists, compiled with YAML frontmatter
      (`name`, `description`, `metadata.version`/`validated`/`validated-stack`) and a top-level
      `Pre-conditions` heading, matching the compiled tornado skill's shape
- [x] `installer/main_test.go`'s two hardcoded skill-count assertions are bumped 27→28 and
      `make -C installer test` passes
- [x] `make -C installer smoke-test` passes, confirming the skill installs cleanly with its
      `[SMAQIT_SKILLS_DIR]` placeholder resolved
- [x] No changes are made to `smaqit.input-deployment`, `smaqit.new-greenfield-project`,
      `smaqit.feature-new`, or `smaqit.infrastructure-cicd-generate` — routing/wiring remains a
      separate follow-up task

## Findings

**Implementation approach:**
- Authored `.smaqit/definitions/skills/smaqit.infrastructure-onboard-k3s-app.md` mirroring the tornado definitions file's exact section shape, covering all six real bugs from the source mechanism (stdin-consuming `while read`+`ssh` loop, `ca.crt` jsonpath escaping, fail-closed token/CA wait, self-minted-token gap, `workflow_dispatch`-on-default-branch constraint, per-name Environment branch-policy scoping) plus the two known gaps.
- Compiled `skills/smaqit.infrastructure-onboard-k3s-app/SKILL.md` following the compiled tornado skill's actual section order — frontmatter with `name`/`description`/`metadata.version`/`validated`/`validated-stack`, `Pre-conditions` promoted to a top-level heading, `Provenance`/`Required-inherited-context` dropped as standalone sections and folded into metadata plus inline prose.
- Ran `make -C installer prepare`, bumped both hardcoded skill-count assertions in `installer/main_test.go` (27→28) plus the matching count in `docs/wiki/workflows/testing-smaqit.md`, and confirmed `make -C installer test` and `make -C installer smoke-test` both pass.
- Added a `CHANGELOG.md` entry under `[Unreleased]/Added`.

**Decisions made:**
- Scope was widened mid-planning (via `task.plan`) from the original definitions-only handoff to also compiling and shipping the skill as a real product capability — the mechanism's four-rounds-of-production-hardening removed the rationale for withholding compilation the way the unproven tornado skill was originally held back.
- Renamed the skill from `smaqit.infrastructure-deploy-k3s` to `smaqit.infrastructure-onboard-k3s-app` before implementation began, to correctly signal app onboarding (Namespace/RBAC/kubeconfig issuance) as distinct from app deployment.
- Stack-detection routing and wiring into `smaqit.new-greenfield-project`/`smaqit.feature-new` were deliberately kept out of scope — Discovery confirmed this requires a net-new app-side deploy mechanism and a new routing input with no hardened source material behind it, unlike the mechanical, low-risk compilation done here.

**Blockers encountered:**
- None. A concurrent session created task 117 mid-implementation, covering exactly the deferred routing/app-deploy work; it landed cleanly on `main` via its own rebase and did not affect this task's worktree or implementation.

**Follow-up identified:**
- Task 117 ("k3s App-Deployment Skill With Routing and CI/CD") already covers the deferred routing/wiring/app-side-deploy work identified in this task's Design Decisions — no new follow-up task needed.

## Files to Create / Modify

| File | Action |
|------|--------|
| `.smaqit/definitions/skills/smaqit.infrastructure-onboard-k3s-app.md` | Create |
| `skills/smaqit.infrastructure-onboard-k3s-app/SKILL.md` | Create — compiled skill |
| `installer/main_test.go` | Modify — bump skill-count assertions 27→28 |
| `docs/wiki/workflows/testing-smaqit.md` | Modify — bump skill count if enumerated |
| `CHANGELOG.md` | Modify — new skill entry |

## Notes

Source material was contributed by a downstream project's infrastructure repo, hardened across
four rounds of real production use (registry-based reconciliation, PSA/NetworkPolicy/quota
hardening, then a second-machine extension) rather than synthesized once and left unproven.
Compilation into a supported `skills/smaqit.infrastructure-onboard-k3s-app/` product capability
(frontmatter, global payload, automated tests, documentation) is part of this task, given the
mechanism's proven track record. Stack-detection routing and wiring into
`smaqit.new-greenfield-project`/`smaqit.feature-new` as a selectable deployment target remain a
separate, future task — that requires designing a net-new app-side deploy mechanism (building and
applying the app's own workloads via the kubeconfig this skill issues) with no hardened source
material behind it, unlike the onboarding mechanism this task ships.
