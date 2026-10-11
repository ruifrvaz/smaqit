---
status: Not Started
created: "2026-10-11"
parent: "125"
---

# Deployment Slots C: Route new-greenfield-project Phases 4–6 Through the Deployment Agent

## Description

Child C of task 125, depends on child B (128). `smaqit.new-greenfield-project` currently calls almost
every infrastructure skill directly and invokes the deployment agent only once, for IaC/CI generation
(Phase 4 step 3). This child routes Phases 4–6 through the agent's slot plan: Phase 4 (dev sweep) as a
local `standard` run, Phase 5 (prod) as a new agent-owned `deployment_path: ci-push`, and Phase 6 as a
single-slot `tls` run — collapsing the per-`provisioning_mode` callouts into profile defaults and fixing
the k3s credential ordering and the `existing-owned` teardown hazard on the way.

## Issue Triage Context

**Mode:** Auto
**Technologies:** GitHub Actions, Terraform, Kubernetes, k3s
**Platforms/Environments:** Claude Code, GitHub Copilot, Codex CLI
**Features/Integrations:** CI/CD push-to-main deploy, correlated workflow-run monitoring, Terraform guarded plan/apply
**Versions/Constraints:** smaqit 3.6.1, with no `.smaqit/slots.yml` the resolved per-mode sequence must match current greenfield behaviour

## Design Decisions

- Inherits task 125's Design Decisions (new `deployment_path: ci-push`; `teardown` default only for `provision`; k3s `target-credentials` after `platform-onboarding`).
- **Phase structure and task bracketing stay unchanged** — only the infrastructure steps inside each phase move into agent calls.

## Implementation Steps

1. Phase 4 — one deployment-agent call, `env=dev`, `deployment_path: standard`, all executors local; per-mode callouts replaced by profile defaults; k3s credential ordering fixed; `teardown` defaults only for `provision`.
2. Phase 5 — one deployment-agent call, `env=prod`, `deployment_path: ci-push` (push to main, correlated run watch, `smaqit.infrastructure-deploy-verify`, `status: deployed`); `provision`/`host-bootstrap`/`app-deploy` are `ci` executors.
3. `agents/deployment.md` — add the `ci-push` deployment path (Execution Parameters, Steps 5–6), mirroring `existing-cicd-pr` without the PR/merge gate.
4. Phase 6 — single-slot `tls` run, conditional on a declared domain; k3s profile `none`.
5. Update the skill's Pre-conditions (k3s-tagged items; drop the "all modes need SSH" contradiction), Scope, Gotchas, Examples, Completion; bump `metadata.version`.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] Greenfield Phases 4–6 invoke the deployment agent (Phase 4 `standard`, Phase 5 `ci-push`, Phase 6 single-slot `tls`) instead of infrastructure skills directly
- [ ] `agents/deployment.md` supports `deployment_path: ci-push`
- [ ] k3s `target-credentials` runs after `platform-onboarding`; default `teardown` applies only to `provision`
- [ ] With no `.smaqit/slots.yml`, the resolved sequence per mode matches current greenfield behaviour

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
| `skills/smaqit.new-greenfield-project/SKILL.md` | Modify |
| `agents/deployment.md` | Modify (`ci-push` path) |

## Notes

Child tasks inherit their active parent's branch, worktree, and workflow mode. Only a standalone or parent task owns Git lifecycle cleanup.
