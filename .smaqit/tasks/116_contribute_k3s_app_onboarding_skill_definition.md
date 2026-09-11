---
status: Not Started
created: "2026-09-11"
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
`.smaqit/definitions/skills/smaqit.infrastructure-deploy-k3s.md`, following the exact structural
precedent of `.smaqit/definitions/skills/smaqit.infrastructure-deploy-rsync-python-tornado.md` and
the task-106 contribution model: a definitions-only handoff, not a compiled skill. Compiling it
into a supported `skills/smaqit.infrastructure-deploy-k3s/` product capability (stack-detection
routing, global payload, tests, docs) is deliberately out of scope here — a separate follow-up task
does that, mirroring task 106's own two-step split.

## Issue Triage Context

**Mode:** Skip
**Technologies:** Kubernetes RBAC/PodSecurity/NetworkPolicy/ResourceQuota, k3s, GitHub Actions (`workflow_dispatch`, Environment protection rules), smaqit skill-definition conventions
**Platforms/Environments:** None — this task produces a definition file only, no live infrastructure is touched
**Features/Integrations:** `.smaqit/definitions/skills/` (definition authoring), `smaqit.infrastructure-deploy-rsync-python-tornado` (structural precedent), `smaqit.create-skill`/L2 compiler (eventual consumer of this artifact, not invoked by this task)
**Versions/Constraints:** Definition file only — no compiled `skills/smaqit.infrastructure-deploy-k3s/` directory; Provenance must not name the real downstream project, repo, or machines

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
- **Compilation into a supported product skill is a separate follow-up task**, not this task's
  scope — matches task 106's own split between contributing the definition and later reconciling
  it into `skills/`.

## Implementation Steps

1. Read `.smaqit/definitions/skills/smaqit.infrastructure-deploy-rsync-python-tornado.md` in full
   as the structural reference, and `.smaqit/tasks/106_reconcile_python_tornado_rsync_deployment_skill.md`
   for how a prior contribution-to-canonical task was scoped and worded.
2. Author `.smaqit/definitions/skills/smaqit.infrastructure-deploy-k3s.md`:
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
4. Do not compile a `skills/smaqit.infrastructure-deploy-k3s/` directory, touch stack-detection
   routing, or touch the installer/release pipeline in this task — that is a separate follow-up,
   matching task 106's own split.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] `.smaqit/definitions/skills/smaqit.infrastructure-deploy-k3s.md` exists, matching
      `smaqit.infrastructure-deploy-rsync-python-tornado.md`'s section structure
- [ ] Provenance section contains no real project, repository, or machine names
- [ ] Documents every real bug and fix listed in Implementation Steps' Gotchas item, not just the
      final correct mechanism
- [ ] Documents the `workflow_dispatch`-must-exist-on-the-default-branch constraint and the
      GitHub-Environment-branch-policy-applies-per-name gotcha
- [ ] Notes `pods/log` RBAC access and non-destructive credential rotation as known gaps not yet
      incorporated
- [ ] No compiled `skills/smaqit.infrastructure-deploy-k3s/` directory is produced by this task —
      definition file only

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
| `.smaqit/definitions/skills/smaqit.infrastructure-deploy-k3s.md` | Create |

## Notes

Source material was contributed by a downstream project's infrastructure repo, hardened across
four rounds of real production use (registry-based reconciliation, PSA/NetworkPolicy/quota
hardening, then a second-machine extension) rather than synthesized once and left unproven.
Compilation into a supported `skills/smaqit.infrastructure-deploy-k3s/` product capability —
stack-detection routing, global payload, automated tests, documentation — is intentionally a
separate follow-up task, matching task 106's own two-step precedent.
