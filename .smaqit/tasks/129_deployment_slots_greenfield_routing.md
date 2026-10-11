---
status: Completed
completed: "2026-10-11"
mode: Assisted
started: "2026-10-11"
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
**Triaged:** 2026-10-11
**Tools searched:** GitHub CLI (`cli/cli`), Terraform (`hashicorp/terraform`), k3s (`k3s-io/k3s`)
**Result:** Advisory

### Advisory Issues
- [#12786 Proposal: Display Commit SHA in `gh run watch` output](https://github.com/cli/cli/issues/12786) — `cli/cli` — labels: enhancement, gh-run, stale. `gh run watch` does not show the head SHA; `ci-push` must correlate the run via `gh run list --json headSha,event,workflowName` before watching it.

### Historical (Closed)
- [#5631 gh run watch --for <ref>](https://github.com/cli/cli/issues/5631) — `cli/cli` — closed 2022-10-20. No ref filter on watch; select the run ID first.

### Omitted Tools
- GitHub Actions — no repository to search (product docs in the research map)
- Kubernetes — covered by the k3s search; no task-specific API change

## Acceptance Criteria

- [x] Greenfield Phases 4–6 invoke the deployment agent (Phase 4 `standard`, Phase 5 `ci-push`, Phase 6 single-slot `tls`) instead of infrastructure skills directly
- [x] `agents/deployment.md` supports `deployment_path: ci-push`
- [x] k3s `target-credentials` runs after `platform-onboarding`; default `teardown` applies only to `provision`
- [x] With no `.smaqit/slots.yml`, the resolved sequence per mode matches current greenfield behaviour

## Findings

**Implementation approach:**
- Greenfield Phase 4 resolves a `standard` plan for the dev environment and makes one Deployment-agent call; the per-mode callouts became a description of what each default profile executes
- Phase 5 resolves a `ci-push` plan for prod and makes one agent call; the agent runs credential, onboarding, and environment-config slots locally, pushes, correlates the `deploy.yml` push run by `headSha`, and verifies with `--expected-sha`
- Phase 6 is a single-slot `tls` run; optional Phase 4 teardown is a single-slot `teardown` run
- The match-or-synthesize procedure moved into a "Family filler synthesis" subsection keyed off the plan's no-candidate warning; synthesized skills must declare their slot
- `agents/deployment.md` gained the `ci-push` path; `smaqit.input-deployment` names it as greenfield's Phase 5 path

**Decisions made:**
- `ci-push` defaults `host-bootstrap` to none instead of a CI executor: the full CI template never bootstraps, cloud-init configures the production host, and current Phase 5 never ran `vm-bootstrap`
- Greenfield bumped to 2.0.0 because orchestration ownership of every infrastructure step moved to the Deployment agent
- k3s now loads the kubeconfig after onboarding in both phases (fixes the old Phase 4 ordering)
- Phase structure and task bracketing unchanged; new Gotcha forbids direct infrastructure-skill calls in Phases 4–6

**Blockers encountered:**
- None

**Follow-up identified:**
- Whether a production VM should run `vm-bootstrap` at all is a pre-existing greenfield question, not introduced here
- An existing `<!-- amendment: ... -->` comment in greenfield Pre-conditions may match Phase 8's amendment scan
- `cicd-generate`'s "workflows generated during Phase 5" wording remains stale (task 123 drift list)

## Files to Create / Modify

| File | Action |
|------|--------|
| `skills/smaqit.new-greenfield-project/SKILL.md` | Modify |
| `agents/deployment.md` | Modify (`ci-push` path) |

## Notes

Child tasks inherit their active parent's branch, worktree, and workflow mode. Only a standalone or parent task owns Git lifecycle cleanup.
