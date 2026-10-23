---
status: Not Started
created: "2026-10-20"
---

# All Deploys Go Through CI/CD Workflows (existing-k3s Enforced, VM Modes Transitional)

## Description

Make "deploys to a target run only in the CI/CD system (GitHub Actions by default); bootstrap and
configuration run locally" a framework rule, and enforce it in the slot resolver for existing-k3s,
whose generated workflow already deploys every environment (one dispatch-only `deploy.yml`,
machine-slug GitHub Environments, in-CI mesh join and verification). project-new's Phase 4 becomes a
CI deploy to the test environment through a new `ci-bootstrap` path. Local `mesh-check`, local
`kubectl` deploys and local in-cluster verification are removed.

Today the local `standard` path is the default everywhere (`smaqit slots plan` without `--path`,
`smaqit.input-deployment`, and so a bare `/smaqit.deployment`), project-new's Phase 4 deploys a dev
target from the operator's machine, and deploy-k3s-app's own gotchas already say to prefer the
generated workflow even for a first deploy, because a local run exercises a different code path than
production (in live use that hid an image-tag casing bug).

VM modes keep the `standard` local path as a declared, transitional exception: their workflows don't
yet support a non-production environment, write the nginx vhost, `docker-compose.yml` or `.env`,
bootstrap the host, issue TLS, or cover the Python stacks. A follow-up task closes those gaps.

## Issue Triage Context

**Mode:** Auto
**Technologies:** GitHub Actions, GitHub CLI, k3s, Kubernetes, Go
**Platforms/Environments:** None
**Features/Integrations:** deployment slot resolver, ci-bootstrap path, workflow_dispatch, existing-k3s CI deploys, project-new Phase 4, deploy-verify ci-evidence
**Versions/Constraints:** smaqit v3.19.0 baseline (after PR #110), MINOR with action-required note, no backwards compatibility, workflow_dispatch requires the workflow on the default branch

## Design Decisions

- **The rule** (Rui, 2026-10-20): a deploy to a target runs only in the CI/CD system. Locally:
  credentials, onboarding PRs, environment syncs, workflow/IaC/manifest generation, artifact
  authoring, and reading CI evidence. Target-acting slots: `provision`, `host-bootstrap`,
  `app-deploy`, `tls`, `replace`, `teardown`.
- **Scope** (Rui, 2026-10-20): enforced for existing-k3s now. VM modes keep `standard` as a declared
  transitional exception until a follow-up task closes their workflow gaps.
- **No local break-glass path for CI-only modes** (Rui, 2026-10-20).
- **project-new Phase 4 deploys the test environment through CI** (Rui, 2026-10-20): the generated
  workflows and manifests land through a PR *without* the deploy sentinel (`post-merge-deploy.yml`
  dispatches with no environment input, so it would deploy the default prod machine); after the
  human merge the agent dispatches `deploy.yml` with `-f environment=<test machine slug>`, selects
  the run by commit, and verifies with `ci-evidence` against the merge SHA. Phase 5 is unchanged
  apart from k3s wording.
- **GitHub Actions only, with a named seam** (Rui, 2026-10-20): the "ci-monitor" contract (trigger,
  select the run by commit, wait, read verify evidence) is documented in the Deployment agent and
  wiki, with `gh` as its default implementation; no abstraction code.
- **No backwards compatibility:** `mesh-check` (operation and script), k3s in-cluster verify and the
  local k3s deploy instructions are removed, not deprecated; the CHANGELOG carries an
  action-required note.
- **A project binding of a target-acting slot under existing-k3s is refused** like a default would
  be; platform repositories are VM-mode projects today and their applies already run through gated
  workflows, so they are unaffected.
- **Depends on task 142** (PR #110: `ci-evidence`, per-path defaults on `slotDef`).

## Implementation Steps

*Phase A: rule and resolver (`installer/slots.go`)*
1. Add a per-slot target-acting flag (`provision`, `host-bootstrap`, `app-deploy`, `tls`, `replace`,
   `teardown`) and a per-mode CI-only flag (`existing-k3s` only; VM modes transitional).
2. `resolveSlotPlan` refuses any target-acting entry with a local executor under a CI-only mode —
   full sequence, single-slot run, or binding — with a new `SLOTS-LOCAL-DEPLOY-REFUSED` diagnostic
   naming the CI paths to use. Configuration slots, `artifact-build` (authoring), `ci-generate`,
   `iac-generate`, and `verify` (`ci-evidence`) stay local.
3. Add `ci-bootstrap` to `deploymentPaths` and `deploymentPathRules`: local configuration slots,
   `artifact-build`, `ci-generate`; `app-deploy` CI-executed; `provision` and `host-bootstrap`
   refused; `tls` none by default; `verify` → `ci-evidence` for existing-k3s via task 142's
   per-path default. `refuseSplitCIRun` stays as is.
4. existing-k3s `platform-prereqs` defaults to none on every path.

*Phase B: agent and skills (depends on A)*
5. `agents/deployment.md`: the rule ("never deploy to a target from this machine under a CI-only
   mode"); a `ci-bootstrap` path bullet and step list (local slots → commit generated workflows and
   manifests → PR without the deploy sentinel → pause for the human merge → dispatch for the
   environment's machine slug → select the run by commit → watch → `ci-evidence` with the merge
   SHA); `standard` marked transitional for VM modes only; the ci-monitor seam.
6. `skills/smaqit.input-deployment/SKILL.md`: default `deployment_path` by mode and repo state —
   existing-k3s → `ci-push` when `deploy.yml` is on the default branch, `ci-bootstrap` otherwise; VM
   modes → `standard`; explicit caller paths win.
7. `skills/smaqit.project-new/SKILL.md`: Phase 4 for existing-k3s uses `ci-bootstrap` against the
   test environment and gates on the CI run's evidence; VM modes keep the local sweep, marked
   transitional; remove the `private-mesh` local mesh precondition (L34); update Phase 5 k3s wording.
   Family filler synthesis for existing-k3s is limited to `image-build-push*` (artifact authoring),
   since `app-deploy` is performed by CI.
8. `skills/smaqit.infrastructure-deploy-k3s-app/`: remove the local `deploy` instructions
   (`credential-file.sh` kubeconfig, local `kubectl`), the `mesh-check` operation, its slot
   declaration and `scripts/mesh-check.sh`; the skill describes what the generated `deploy.yml`
   performs and requires. `skills/smaqit.infrastructure-deploy-verify/`: remove the in-cluster
   `verify` mode (existing-k3s always verifies with `ci-evidence`; VM HTTP mode stays).
9. `skills/smaqit.infrastructure-cicd-generate/SKILL.md`: the k3s workflow is the only k3s deploy
   path; mention the ci-monitor seam. Version bumps for every changed skill.

*Phase C: docs*
10. `framework/` (the deployment rule, in AGENTS/PHASES), `docs/wiki/workflows/deployment-slots.md`
    (paths table with `ci-bootstrap`, a CI-only/transitional table, existing-k3s Targets section),
    `.smaqit/compendium.md` ("local applies are supported by design" narrowed to VM modes; a new entry
    for the rule and the seam), `README.md`'s deploy paragraph, `templates/AGENTS.md.template` only if
    its slot summary becomes wrong.

*Phase D: tasks*
11. ~~Edit `.smaqit/tasks/137_existing_k3s_credential_slots_scripts_only_vault.md`: drop its
    `standard`-path local in-cluster verify step (step 5) and the related notes.~~ Done
    2026-10-23, when task 137 was rescoped to follow task 145 (OIDC replaces kubeconfig).
12. Create a follow-up task "VM modes: CI-only deploys" listing the gaps found during planning:
    multi-environment VM `deploy.yml` (environment input, per-environment GitHub Environments instead
    of repo-level `VM_HOST`); nginx vhost, `docker-compose.yml` and `.env` written in CI; `vm-bootstrap`
    and certbot/TLS as gated workflows; Python stack templates; app SSH key bootstrap without a local
    SSH step; the first `terraform apply` in CI (including the base-key vs app-key keypair mismatch
    between `terraform-env.sh` and the CI `VM_SSH_PUBLIC_KEY`); day-two `teardown`/`replace` as gated
    dispatches; VM verify from CI evidence.

*Phase E: tests and release*
13. `installer/slots_test.go`: existing-k3s `standard` refused (full, single-slot, binding);
    `ci-bootstrap` plans for test/prod/dev rows; `platform-prereqs` none on every k3s path; VM
    `standard` unchanged; split-run rule on `ci-bootstrap`; `TestDefaultFillersDeclareTheirSlots`
    after `mesh-check`'s removal; update `TestSlotPlanDefaultsMatchCurrentBehaviour` and
    `installer/infra_spec_test.go`'s `TestSlotPlanExistingK3sGate` for k3s. Remove
    `TestMeshCheckRequiresTheMeshForPrivateTargets` (`installer/k3s_scripts_test.go`).
    `scripts/smoke-test-installer.sh`: k3s plans pass `--path`; assert `standard` is refused for k3s.
14. Deliberate breaks (drop the CI-only flag; allow a local `app-deploy` binding for k3s; remove
    `ci-bootstrap`'s `provision` refusal) are each caught.
15. `make -C installer test` and `make -C installer smoke-test`; CHANGELOG with an action-required
    note (existing-k3s no longer deploys locally; `mesh-check` and in-cluster verify removed; a
    target-acting k3s binding is refused); MINOR release via task-complete.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] The resolver refuses any target-acting slot (`provision`, `host-bootstrap`, `app-deploy`, `tls`, `replace`, `teardown`) with a local executor under existing-k3s, as a full sequence, a single-slot run or a binding, with `SLOTS-LOCAL-DEPLOY-REFUSED`; VM modes resolve exactly as before
- [ ] A `ci-bootstrap` path plans existing-k3s as local configuration, artifact authoring and `ci-generate`, then `app-deploy @ci` and `verify` `ci-evidence`; it refuses `provision` and `host-bootstrap`; existing-k3s `platform-prereqs` defaults to none on every path
- [ ] `smaqit.input-deployment` defaults existing-k3s to `ci-push` when `deploy.yml` is on the default branch and `ci-bootstrap` otherwise; VM modes keep `standard`; explicit caller paths win
- [ ] The Deployment agent documents `ci-bootstrap` (local slots, workflows landed through a PR without the deploy sentinel, human merge, dispatch for the environment's machine slug, run selected by commit, `ci-evidence` with the merge SHA) and names the ci-monitor seam with GitHub Actions as the default
- [ ] project-new Phase 4 deploys existing-k3s's test environment through `ci-bootstrap` and gates on the CI run's evidence; the local mesh precondition is gone; VM Phase 4 is marked transitional
- [ ] deploy-k3s-app has no local deploy instructions and no `mesh-check` operation or script; deploy-verify has no in-cluster mode; no shipped skill tells an agent to deploy to an existing-k3s target from the local machine
- [ ] `framework/`, the deployment-slots wiki, the compendium and the README state the rule, the k3s enforcement and the VM transitional exception
- [ ] Task 137 drops its local-verify step; a follow-up task for VM-mode CI-only deploys exists, listing the workflow gaps
- [ ] `make -C installer test` and `make -C installer smoke-test` pass; CHANGELOG has an action-required note; MINOR release published via the task PR

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
| installer/slots.go | Modify |
| installer/slots_test.go | Modify |
| installer/infra_spec_test.go | Modify |
| installer/k3s_scripts_test.go | Modify |
| agents/deployment.md | Modify |
| skills/smaqit.input-deployment/SKILL.md | Modify |
| skills/smaqit.project-new/SKILL.md | Modify |
| skills/smaqit.infrastructure-deploy-k3s-app/SKILL.md | Modify |
| skills/smaqit.infrastructure-deploy-k3s-app/scripts/mesh-check.sh | Delete |
| skills/smaqit.infrastructure-deploy-verify/SKILL.md | Modify |
| skills/smaqit.infrastructure-cicd-generate/SKILL.md | Modify |
| framework/AGENTS.md, framework/PHASES.md | Modify |
| docs/wiki/workflows/deployment-slots.md | Modify |
| .smaqit/compendium.md | Modify |
| README.md | Modify |
| templates/AGENTS.md.template | Modify (if needed) |
| scripts/smoke-test-installer.sh | Modify |
| .smaqit/tasks/137_existing_k3s_credential_slots_scripts_only_vault.md | Modify |
| CHANGELOG.md | Modify |

## Notes

- Planned with `task.plan` on 2026-10-20 after discussing why `mesh-check.sh` says "for local runs".
  Three discovery agents mapped project-new and the workflow sets, the resolver paths and agent
  defaults, and every skill that acts on a target locally.
- Start after PR #110 (task 142) merges. Task 143 (Claude Code bench for `ci-evidence`) is unaffected.
- Workflow-on-default-branch constraint: GitHub's `workflow_dispatch` API returns 404 for a workflow
  that exists only on a non-default branch (cicd-generate SKILL.md L217-222), which is why
  `ci-bootstrap` lands the workflows through a PR before dispatching.
