---
status: In Progress
created: "2026-10-20"
mode: Assisted
started: "2026-10-20"
---

# CI-Driven Deploys Run No Local Target Checks

## Description

On the `existing-cicd-pr` and `ci-push` deployment paths, the existing-k3s slot plan still runs two
slots on the operator's machine: `platform-prereqs` (`smaqit.infrastructure-deploy-k3s-app`
`mesh-check`) and `verify` (`smaqit.infrastructure-deploy-verify` in in-cluster mode). Both need the
operator's machine on the platform's mesh, and in-cluster verify also needs the scoped kubeconfig
exported locally. The generated `deploy.yml` already does all of it in CI: it joins the mesh, checks
that the Deployment runs the commit's image ("Verify revision") and runs the in-cluster probe or the
HTTPS check ("Verify in cluster" / "Verify over HTTPS").

Task 122 meant `mesh-check` for local deployments only (deploy-k3s-app's own text says "In CI the
generated `deploy.yml` joins the mesh itself, so this operation runs only for local deployments"), but
it added no CI-path rule, so the resolver plans it `@local` on both CI paths. A downstream project's
first feature deploy on `existing-cicd-pr` stopped at slot 1 because the operator's machine was not on
the mesh, although nothing in that deploy needed local mesh access.

This task makes `platform-prereqs` default to none on both CI paths and makes `verify` default, for
existing-k3s on both CI paths, to a new deploy-verify operation `ci-evidence` that decides from the
deploy run's own recorded steps. `ci-evidence` was planned in task 137; it moves here, and 137 keeps
only its credential and runtime-secret work.

## Issue Triage Context

**Mode:** Auto
**Technologies:** GitHub CLI, GitHub Actions, k3s
**Platforms/Environments:** None
**Features/Integrations:** gh run view --json jobs, gh api compare, deploy slot plan, existing-k3s CI deploy paths, deploy-verify
**Versions/Constraints:** smaqit v3.18.0 baseline, MINOR release, no backwards compatibility

## Design Decisions

- **One task owns both halves** (Rui, 2026-10-20): the mesh-check default and the `ci-evidence` verify ship together; `ci-evidence` leaves task 137, which keeps credentials and runtime secrets.
- **`ci-evidence` covers every existing-k3s row on CI paths** (Rui, 2026-10-20): private-mesh, platform-owned and public rows alike, checking whichever of "Verify in cluster" (in-cluster rows: private-mesh or platform-owned) and "Verify over HTTPS" (external rows) applies. VM modes keep their local HTTP verify.
- **`platform-prereqs` is none on CI paths, not CI-executed:** a CI-executed slot at order 1 would put the local credential slots inside the CI span, which `refuseSplitCIRun` refuses. Only existing-k3s fills the slot, so no other mode changes.
- **`verify` stays a local-executor slot:** `ci-evidence` only reads `gh run view`, so the split-CI-run rule and `TestSlotPlanLocalSlotsPrecedeCI`'s verify exemption stay correct.
- **Per-path, per-mode default:** `pathRule` is keyed by path only, so `slotDef` gains a per-path, per-mode default table (path → mode → filler) that the resolver's default branch consults before `Defaults[mode]`; `DefaultNone` still applies after it. Bindings still override it; a single-slot `--slot=verify` run on a CI path resolves the same way. A single-slot `--slot=platform-prereqs` run on a CI path still resolves `mesh-check`, because single-slot runs ignore `DefaultNone` and the request is explicit.
- **`--expected-sha` is the intended commit, proven by ancestry** (Rui, 2026-10-20): the agent passes the merge SHA on `existing-cicd-pr` and the pushed SHA on `ci-push`, never the run's own `headSha` (on the dispatcher path that would compare the run with itself, since `post-merge-deploy.yml` dispatches at `main`'s tip). `ci-evidence` passes when `headSha` equals it, or when `gh api repos/{owner}/{repo}/compare/<expected>...<headSha>` reports `identical` or `ahead`; `behind` or `diverged` fails as a wrong SHA. Other `verify` fillers keep `--expected-sha <deploy-run-headSha>`, because the app reports the SHA that was deployed.
- **Evidence only from GitHub's run record:** `gh run view <id> --json headSha,status,conclusion,jobs` plus the one compare call; `status` distinguishes an unfinished run from a failed one. No kubeconfig, Vault, kubectl, mesh or log reading decides the result. Steps are looked up only in the `deploy` job, because GitHub's implicit steps exist in both jobs.
- **Row input like `mesh-check.sh`** (Rui, 2026-10-20): `ci-evidence.sh <environments.json|-> <environment> --run-id <id> --expected-sha <sha>`; the skill pipes `smaqit spec environments --env=<env> --json` into it. The applicable step follows `verifiesInCluster()`, the same rule as the workflow's `VERIFY`.
- **The target machine is not checked:** `--json jobs` doesn't show which machine a run targeted; the agent's existing run correlation (default prod slug on the dispatcher path, `-f environment` on `ci-push`) stays responsible for it.
- **No backwards compatibility, no opt-out flag;** the `standard` path (local deploys: mesh-check, in-cluster verify) is unchanged. A project binding that names deploy-verify for `verify` without `operation` now fails with `SLOTS-OPERATION-INVALID` (the skill declares two operations); the CHANGELOG says so.

## Implementation Steps

*Phase A: resolver (`installer/slots.go`)*
1. `deploymentPathRules` (L199): add `platform-prereqs` to `DefaultNone` for `existing-cicd-pr` and `ci-push`. Only existing-k3s fills the slot, so no other mode changes.
2. `slotDef` (L42): add a per-path, per-mode default table (path → mode → filler). `resolveSlotPlan`'s default branch (L669) reads it before `Defaults[opts.Mode]`; `DefaultNone` still applies after. The `verify` def (L160) uses it: existing-k3s on `existing-cicd-pr` and `ci-push` → `skillFiller(skillDeployVerify, "ci-evidence")`. Bindings still win; `--slot=verify` resolves the same way. Parallel with Phase B.

*Phase B: `ci-evidence` (deploy-verify)*
3. Create `skills/smaqit.infrastructure-deploy-verify/scripts/ci-evidence.sh` (the skill's first script). Usage `ci-evidence.sh <environments.json|-> <environment> --run-id <id> --expected-sha <sha>`; exit 0 pass, 1 fail, 2 usage; `set -euo pipefail` and the jq guard, as in `mesh-check.sh`. Applicable step from the row (`verifiesInCluster()` rule: private-mesh or platform-owned → "Verify in cluster", else "Verify over HTTPS"). Reads `gh run view <id> --json headSha,status,conclusion,jobs`. Checks in order, each with a distinct FAIL line: run finished (`status` completed); run concluded `success`; SHA — `headSha` equals expected, else `gh api repos/{owner}/{repo}/compare/<expected>...<headSha>` `identical`/`ahead` passes, `behind`/`diverged` fails as a wrong SHA; `deploy` job present; "Verify revision" `success`; applicable verify step `success` (separate FAIL for failed, skipped — spec and workflow disagree on the verify mode — and missing). Steps looked up only in the `deploy` job. PASS line names the run ID, `headSha` and the step checked. Step names come from `skills/smaqit.infrastructure-cicd-generate/assets/deploy.yml.k3s.template` (L250, L264, L292).
4. deploy-verify `SKILL.md`: frontmatter `smaqit-slots: "verify=verify, verify=ci-evidence"`, version 1.3.0 → 1.4.0, description mentions verifying from the CI run; a "Choose the operation" lead-in (the slot plan decides); `verify` keeps its HTTP and in-cluster modes; new `ci-evidence` section (inputs, the piped `smaqit spec environments --env=<env> --json` call, check order); Output, Scope (no kubeconfig, Vault, kubectl or logs), example, gotcha (pass the intended commit, not the run's `headSha`), Completion, failure rows (unfinished run, wrong SHA, step failed/skipped/missing, compare API refused).

*Phase C: instruction text (depends on A and B)*
5. `agents/deployment.md`: L27 and L28 — the `verify` slot runs its planned operation; L76 — add the `ci-evidence` case; L242 and L250 — for `ci-evidence`, pass `--run-id <deploy run ID>` and `--expected-sha <intended commit>` (merge SHA on `existing-cicd-pr`, pushed SHA on `ci-push`); other fillers keep `--expected-sha <deploy-run-headSha>`.
6. `skills/smaqit.feature-new/SKILL.md` L134 (note the CI-path defaults), L157, L161 (verify by its planned operation); `skills/smaqit.project-new/SKILL.md` L34 (mesh precondition only for environments deployed locally — the `standard`-path dev sweep), L151-153, L156; `skills/smaqit.infrastructure-deploy-k3s-app/SKILL.md` L71-72 (the plan skips `mesh-check` on CI paths), L340 failure row, L17, L322, version 2.1.1 → 2.1.2. Version bumps for feature-new and project-new. `skills/smaqit.input-deployment/SKILL.md` needs no change (its plan summary is generic).
7. Docs: `docs/wiki/workflows/deployment-slots.md` defaults table (L41, L52, with a CI-path note), paths table (L68-69), L74's verify note, existing-k3s Targets section (L168-173); `.smaqit/compendium.md` L226 and L409. `templates/AGENTS.md.template` needs no change.

*Phase D: task 137 (parallel with C)*
8. Edit `.smaqit/tasks/137_existing_k3s_credential_slots_scripts_only_vault.md`: remove the `ci-evidence` mentions at L23-24 and L39, the design decision at L59-62, Step 5 (L84-85), the test bullet at L95, the AC at L110-111, and `installer/slots.go` from the L138 Files row; narrow the L19-20 deploy-verify note and the L143 row to the `standard` path's local in-cluster verify; add a line naming task 142 as the owner of CI-path verification. 137's PLANNING row doesn't mention verify, so it stays.

*Phase E: tests and release*
9. `installer/slots_test.go`: `TestSlotPlanDeploymentPaths` — existing-k3s on both CI paths for `test` (private-mesh), `prod` (platform-owned) and `dev` (public own-tls): `platform-prereqs` none, `verify` `profile smaqit.infrastructure-deploy-verify:ci-evidence @local`; VM modes on CI paths still `verify:verify`. `TestSlotPlanSingleSlotRuns` — a CI-path `--slot=verify` case. A binding-override case (a test skill bound to `verify` on a CI path wins). Extend `TestDefaultFillersDeclareTheirSlots` to walk the per-path table. `TestSlotPlanDefaultsMatchCurrentBehaviour` stays unchanged.
10. `installer/skill_scripts_test.go`: extend `stubGh` with a `run view` branch (JSON from `STUB_GH_RUN_JSON`) and an `api …/compare/…` branch (status from `STUB_GH_COMPARE_STATUS`). Table test for `ci-evidence.sh`, rows from `specEnvironmentsJSON`: pass in-cluster, pass external, pass on an ancestor (`ahead`); fail on `behind`, `diverged`, run failed, unfinished run, revision step failed, applicable step failed/skipped/missing, no `deploy` job; usage error exit 2. Assert the only `gh` argv are `run view` and the compare call.
11. `scripts/smoke-test-installer.sh`: after the existing smoke-k3s spec, `slots plan --phase=deploy --mode=existing-k3s --env=test --path=existing-cicd-pr --json`; assert `"operation": "ci-evidence"` and no `mesh-check`. The deployment bench fixture pins only the `standard` path and needs no change.
12. `make -C installer test` (runs prepare, so tests see the SKILL.md change) and `make -C installer smoke-test`; deliberate breaks (drop `verify=ci-evidence`, drop the `DefaultNone` entry, accept a skipped step, accept `behind`) each caught. CHANGELOG `[Unreleased]` → Changed: CI-path existing-k3s deploys no longer need the operator's machine on the mesh or a local kubeconfig; action required: a binding that names deploy-verify for `verify` without `operation` now fails with `SLOTS-OPERATION-INVALID`. Release via task-complete (`Prepare release vX.Y.Z`, MINOR — expected v3.19.0; check open PR claims).

## Known Issues Triage
**Triaged:** 2026-10-20
**Tools searched:** GitHub CLI (`cli/cli`), GitHub Actions (`actions/starter-workflows`), k3s (`k3s-io/k3s`)
**Result:** Advisory

### Blocking Issues
- None

### Advisory Issues
- [#9228 `gh run list` and `gh run view` fails when the run's workflow is defined via Organization ruleset](https://github.com/cli/cli/issues/9228) — `cli/cli` — open — bug, priority-2, core, discuss, gh-run. Not this task's case (the generated `deploy.yml` is a repository workflow), but `ci-evidence.sh` must treat a failing `gh run view` as a FAIL, never as an empty result.
- [#10853 `gh run view --log -j <job_id>` erroneously returns no stdout/stderr, exits 0](https://github.com/cli/cli/issues/10853) — `cli/cli` — open — bug, needs-triage, platform. `ci-evidence` reads no logs; listed only because it shares the `run view` feature.
- [#11559 Gzipped responses are not decoded so cannot be unmarshaled as JSON](https://github.com/cli/cli/issues/11559) — `cli/cli` — open — enhancement, needs-triage, gh-api. Partial match on `gh api`; the compare endpoint returns plain JSON.

### Historical (Closed)
- [#13972 `gh run list` reports a run as `queued` while its jobs are running](https://github.com/cli/cli/issues/13972) — `cli/cli` — closed. Supports checking `status` = `completed` rather than inferring completion from jobs.
- [#11311 v2.75 fails downloading logs from workflows with skipped jobs](https://github.com/cli/cli/issues/11311) — `cli/cli` — closed. Logs only; not used.
- [#14253 `gh pr checks` omits a failing check run, and `statusCheckRollup` reports SUCCESS for the commit](https://github.com/cli/cli/issues/14253) — `cli/cli` — closed. Not used (evidence comes from the run's own jobs).

### Unresolvable Tools
- None

### Omitted Tools
- jq — in the research block but not a task-context technology

### Search Warnings
- Categorization limitation: GitHub Actions resolved to `actions/starter-workflows`, which doesn't own step-conclusion semantics (`actions/runner` does); its open and closed searches returned no items. Issue dates were not captured in this run.

## Acceptance Criteria

- [ ] On `existing-cicd-pr` and `ci-push`, the existing-k3s plan resolves `platform-prereqs` to none and `verify` to `smaqit.infrastructure-deploy-verify:ci-evidence @local` for private-mesh, platform-owned and public rows; the `standard` path and every VM mode resolve exactly as before; a project binding still overrides both slots
- [ ] `ci-evidence.sh` passes only when `gh run view <id> --json headSha,status,conclusion,jobs` shows a finished, successful run whose `headSha` is the intended commit or contains it (`gh api` compare `identical` or `ahead`), and whose `deploy` job has a successful "Verify revision" and a successful applicable verify step for the row; it fails with a distinct line for an unfinished run, a failed run, a wrong SHA (`behind` or `diverged`), and a failed, skipped or missing step (stub `gh` JSON tests); it reads no kubeconfig, Vault or logs
- [ ] deploy-verify declares `ci-evidence` in `metadata.smaqit-slots` and `TestDefaultFillersDeclareTheirSlots` passes
- [ ] `agents/deployment.md`, feature-new, project-new (including its mesh precondition), input-deployment where needed, and deploy-k3s-app's mesh-check text describe `verify` by its planned operation and no longer require the operator's machine on the mesh for CI paths
- [ ] `docs/wiki/workflows/deployment-slots.md` and `.smaqit/compendium.md` reflect the CI-path defaults
- [ ] Task 137 no longer carries the `ci-evidence` decision, step or criterion, and points to 142 for CI-path verification
- [ ] `make -C installer test` and `make -C installer smoke-test` pass; CHANGELOG updated, including the action-required note for deploy-verify bindings without `operation`; MINOR release published via the task PR

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
| installer/skill_scripts_test.go | Modify |
| skills/smaqit.infrastructure-deploy-verify/SKILL.md | Modify |
| skills/smaqit.infrastructure-deploy-verify/scripts/ci-evidence.sh | Create |
| skills/smaqit.infrastructure-deploy-k3s-app/SKILL.md | Modify |
| skills/smaqit.feature-new/SKILL.md | Modify |
| skills/smaqit.project-new/SKILL.md | Modify |
| skills/smaqit.input-deployment/SKILL.md | Modify (if needed) |
| agents/deployment.md | Modify |
| docs/wiki/workflows/deployment-slots.md | Modify |
| .smaqit/compendium.md | Modify |
| templates/AGENTS.md.template | Modify (if needed) |
| scripts/smoke-test-installer.sh | Modify |
| .smaqit/tasks/137_existing_k3s_credential_slots_scripts_only_vault.md | Modify |
| CHANGELOG.md | Modify |

## Notes

- Found when a downstream project's feature-new Phase 3 deploy on `existing-cicd-pr` stopped at `platform-prereqs` (`mesh-check: FAIL`), with the deploy itself running entirely in CI.
- `installer/infra_spec.go`'s `verifiesInCluster()` (L59-61) already classifies rows (private-mesh or platform-owned → in-cluster); `ci-evidence` must choose the same step the rendered workflow runs.
