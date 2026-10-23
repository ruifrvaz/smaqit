---
status: PR Open
pr: 112
created: "2026-10-20"
started: "2026-10-23"
mode: Assisted
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
**Triaged:** 2026-10-23
**Tools searched:** GitHub CLI (`cli/cli`), k3s (`k3s-io/k3s`), Kubernetes (`kubernetes/kubernetes`)
**Result:** Advisory

### Blocking Issues
- None

### Advisory Issues
- [#10861 `gh workflow run` initiates two jobs](https://github.com/cli/cli/issues/10861) — `cli/cli` — opened 2025-04-24 — more-info-needed. `ci-bootstrap` selects the deploy run by commit after dispatching; that selection must not assume exactly one matching run (pick the dispatch-event run for the workflow, or fail on ambiguity).
- [#3521 Use current branch as `--ref` for `gh workflow run`](https://github.com/cli/cli/issues/3521) — `cli/cli` — opened 2021-04-27 — enhancement, actions, gh-workflow. `ci-bootstrap` dispatches at the default branch after the merge; pass `--ref` explicitly rather than relying on a default.
- [#8910 `gh pr ready` doesn't work with GitHub Actions' GITHUB_TOKEN secret](https://github.com/cli/cli/issues/8910) — `cli/cli` — opened 2024-04-02 — bug, needs-investigation. Not this task's dimension (the agent opens the PR with the operator's `gh` login, not in a workflow).
- [#142376 Exec drops end of stdout when client reads slowly](https://github.com/kubernetes/kubernetes/issues/142376) — `kubernetes/kubernetes` — opened 2026-09-24 — kind/bug, sig/node, sig/api-machinery, sig/cli, triage/accepted. Unaffected: `exec-refused.sh` stays in the CI workflow and decides from a Forbidden refusal, not from stdout content.

### Historical (Closed)
- [#4001 workflow run: Return the ID](https://github.com/cli/cli/issues/4001) — `cli/cli` — closed. With [#12968 `gh workflow run` changes output uncontrollably](https://github.com/cli/cli/issues/12968) (closed), supports selecting the run by commit through `gh run list` rather than parsing `gh workflow run` output.
- [#9781 `gh workflow run --ref [ref]` ignoring `ref`](https://github.com/cli/cli/issues/9781) — `cli/cli` — closed. Verify the selected run's `headSha` (as `ci-evidence` already does) rather than trusting the dispatch ref.
- [#14477 workflow run -f silently drops boolean workflow_dispatch inputs](https://github.com/cli/cli/issues/14477) — `cli/cli` — closed (unable-to-reproduce). The `environment` input is a string; not affected.
- [#142024 Namespace-scoped `RoleBinding` can grant write permissions on the namespace itself](https://github.com/kubernetes/kubernetes/issues/142024) — `kubernetes/kubernetes` — closed. Tangential (platform RBAC); no change here.

### Unresolvable Tools
- None

### Omitted Tools
- GitHub Actions — resolved to `actions/starter-workflows` (workflow templates, not an Actions bug tracker); excluded as in task 121's triage
- Go — only the standard library and `go test` are used; no feature dimension to search

### Search Warnings
- Categorization limitation: `k3s-io/k3s` open search returned no items and its closed results were unrelated; the task removes local `kubectl` use, so k3s/Kubernetes exposure is minimal. Closed-issue dates were not captured in this run.

## Acceptance Criteria

- [x] The resolver refuses any target-acting slot (`provision`, `host-bootstrap`, `app-deploy`, `tls`, `replace`, `teardown`) with a local executor under existing-k3s, as a full sequence, a single-slot run or a binding, with `SLOTS-LOCAL-DEPLOY-REFUSED`; VM modes resolve exactly as before
- [x] A `ci-bootstrap` path plans existing-k3s as local configuration, artifact authoring and `ci-generate`, then `app-deploy @ci` and `verify` `ci-evidence`; it refuses `provision` and `host-bootstrap`; existing-k3s `platform-prereqs` defaults to none on every path
- [x] `smaqit.input-deployment` defaults existing-k3s to `ci-push` when `deploy.yml` is on the default branch and `ci-bootstrap` otherwise; VM modes keep `standard`; explicit caller paths win
- [x] The Deployment agent documents `ci-bootstrap` (local slots, workflows landed through a PR without the deploy sentinel, human merge, dispatch for the environment's machine slug, run selected by commit, `ci-evidence` with the merge SHA) and names the ci-monitor seam with GitHub Actions as the default
- [x] project-new Phase 4 deploys existing-k3s's test environment through `ci-bootstrap` and gates on the CI run's evidence; the local mesh precondition is gone; VM Phase 4 is marked transitional
- [x] deploy-k3s-app has no local deploy instructions and no `mesh-check` operation or script; deploy-verify has no in-cluster mode; no shipped skill tells an agent to deploy to an existing-k3s target from the local machine
- [x] `framework/`, the deployment-slots wiki, the compendium and the README state the rule, the k3s enforcement and the VM transitional exception
- [x] Task 137 drops its local-verify step; a follow-up task for VM-mode CI-only deploys exists, listing the workflow gaps
- [x] `make -C installer test` and `make -C installer smoke-test` pass; CHANGELOG has an action-required note; MINOR release published via the task PR

## Findings

**Implementation approach:**
- Resolver: `slotDef` gained `TargetActing` (provision, host-bootstrap, app-deploy, tls, replace, teardown) and a `ciOnlyModes` set holding existing-k3s; `resolveSlotPlan` refuses a local executor on a target-acting slot under a CI-only mode with `SLOTS-LOCAL-DEPLOY-REFUSED`, at the one place the executor is set, so defaults, single-slot runs and bindings are all covered.
- New `ci-bootstrap` path: refuses `provision`/`host-bootstrap`, defaults `platform-prereqs` and `tls` to none, CI-executes only `app-deploy`; `verify` gets `ci-evidence` through the existing `PathDefaults`. `platform-prereqs` has no default for any mode now; `validateSlotPlanOptions` refuses `ci-bootstrap` for VM modes.
- Skills: deploy-k3s-app 3.0.0 describes the sequence the generated `deploy.yml` performs (no local deploy, no `mesh-check`, script deleted); deploy-verify 2.0.0 (no in-cluster mode); cicd-generate 3.3.0 renders the Ingress and probe manifests (new Step 6) and names the ci-monitor seam; input-deployment 1.10.0 picks the path from mode and `git cat-file -e origin/<default>:.github/workflows/deploy.yml`; project-new 3.2.0 Phase 4 uses `ci-bootstrap` for k3s; image-build-push-static 1.2.2 text.
- Deployment agent: the rule, `standard` marked transitional, a `ci-bootstrap` step list, the ci-monitor contract, and a stop on ambiguous run selection. Framework (`AGENTS.md` CI-only profiles, `PHASES.md` Where Deployment Executes), the deployment-slots wiki, compendium, README and project `AGENTS.md` template updated; CHANGELOG `[Unreleased]` carries three action-required notes.
- Tests: defaults per mode (VM on `standard`, k3s on `ci-bootstrap`), all three CI paths for private-mesh/platform-owned/public rows, a new `TestSlotPlanCIOnlyModeRefusesLocalDeploys` (full, single-slot and bound tls/teardown/replace on every path, VM bindings unaffected, `ci-bootstrap` binding refusal, VM modes refused on `ci-bootstrap`); `TestMeshCheckRequiresTheMeshForPrivateTargets` removed; smoke test asserts the refusal and CI-executed `app-deploy` with `ci-evidence` on every CI path. `make test` and `make smoke-test` pass; three deliberate breaks each caught.

**Decisions made:**
- Manifest rendering moved to cicd-generate's k3s mode (Rui, 2026-10-23): with `app-deploy` CI-only, the agent never reads its filler, so nothing else would render `ingress*.yaml`/`probe-job.yaml`; cicd-generate already reads the Environments table and vendors deploy-k3s-app's scripts, and runs after `artifact-build` authored `deployment.yaml`. The rendered files are generated output, overwritten each run.
- `ci-bootstrap` is refused for VM modes (`SLOTS-ARGUMENT-INVALID`): their workflows can't deploy a test environment yet, so the path couldn't work there.
- A k3s binding of `tls`, `replace` or `teardown` is refused on every path, because no path runs them in CI; offboarding stays the platform's process.
- `artifact-build` stays local on `ci-bootstrap`: image-build-push-static only authors files and runs a local `docker build` check; pushing is the CI `build` job.
- Run selection after a dispatch treats more than one candidate as ambiguous and stops (from triage: cli/cli #10861).
- Version bumps: majors for deploy-k3s-app and deploy-verify (operations removed), minors for cicd-generate, input-deployment and project-new, patch for image-build-push-static. Skill definitions under `.smaqit/definitions/skills/` were not edited, following task 142.
- Task 137 was rescoped earlier the same day (step 11), when OIDC was chosen to replace kubeconfig in task 145.

**Blockers encountered:**
- None. An Edit in `agents/deployment.md` duplicated three `ci-push` bullets under the new `ci-bootstrap` block; caught on review and removed.

**Follow-up identified:**
- Task 146 (created): VM modes CI-only deploys, listing the workflow gaps.
- Task 145 still has to remove the kubeconfig load/sync and `credential-file.sh`'s kubeconfig entry, which nothing local reads any more.
- The `.smaqit/definitions/skills/` definitions for deploy-k3s-app, deploy-verify and cicd-generate now lag further (no `mesh-check` removal, no `ci-evidence`, no Step 6).
- No live `ci-bootstrap` run yet; the next downstream existing-k3s first deploy is its live check.

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
