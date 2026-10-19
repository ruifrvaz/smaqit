---
status: Not Started
created: "2026-10-20"
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
**Features/Integrations:** gh run view --json jobs, deploy slot plan, existing-k3s CI deploy paths, deploy-verify
**Versions/Constraints:** smaqit v3.18.0 baseline, MINOR release, no backwards compatibility

## Design Decisions

- **One task owns both halves** (Rui, 2026-10-20): the mesh-check default and the `ci-evidence` verify ship together; `ci-evidence` leaves task 137, which keeps credentials and runtime secrets.
- **`ci-evidence` covers every existing-k3s row on CI paths** (Rui, 2026-10-20): private-mesh, platform-owned and public rows alike, checking whichever of "Verify in cluster" (in-cluster rows: private-mesh or platform-owned) and "Verify over HTTPS" (external rows) applies. VM modes keep their local HTTP verify.
- **`platform-prereqs` is none on CI paths, not CI-executed:** a CI-executed slot at order 1 would put the local credential slots inside the CI span, which `refuseSplitCIRun` refuses. Only existing-k3s fills the slot, so no other mode changes.
- **`verify` stays a local-executor slot:** `ci-evidence` only reads `gh run view`, so the split-CI-run rule and `TestSlotPlanLocalSlotsPrecedeCI`'s verify exemption stay correct.
- **Per-path, per-mode default:** `pathRule` is keyed by path only, so `slotDef` gains a per-path, per-mode default override that the resolver's default branch consults before `Defaults[mode]`. Bindings still override it; a single-slot `--slot=verify` run on a CI path resolves the same way.
- **Evidence only from the compact run record:** `gh run view <id> --json headSha,conclusion,jobs`; no kubeconfig, Vault, mesh or log reading decides the result.
- **No backwards compatibility, no opt-out flag;** the `standard` path (local deploys: mesh-check, in-cluster verify) is unchanged.

## Implementation Steps

*Phase A: resolver*
1. `installer/slots.go`: add `platform-prereqs` to `DefaultNone` for `existing-cicd-pr` and `ci-push` in `deploymentPathRules`.
2. Add a per-path, per-mode default override to `slotDef` and consult it in `resolveSlotPlan`'s default branch before `Defaults[mode]` (and before `DefaultNone`). Use it for `verify`: existing-k3s on both CI paths → `skillFiller(skillDeployVerify, "ci-evidence")`. Single-slot runs honour it.

*Phase B: `ci-evidence`*
3. `skills/smaqit.infrastructure-deploy-verify/`: new operation `ci-evidence` with `scripts/ci-evidence.sh`, declared in `metadata.smaqit-slots` next to `verify=verify` (confirm the frontmatter format allows two operations for one slot; extend the parser and `TestDefaultFillersDeclareTheirSlots` if not). Inputs: `--run-id`, `--expected-sha`, `--environment` (reads the row through `smaqit spec environments --json` to choose the applicable step). PASS only when `headSha` equals the expected SHA, the run concluded `success`, "Verify revision" succeeded and the applicable verify step succeeded; distinct FAIL lines for a wrong SHA, a failed, skipped or missing step, and an unfinished run. Step names come from `skills/smaqit.infrastructure-cicd-generate/assets/deploy.yml.k3s.template`.
4. deploy-verify SKILL.md: operation section, inputs, outputs, failure table rows, example; version bump.

*Phase C: instruction text*
5. `agents/deployment.md` (L27, L28, L76, L242, L250): the `verify` slot runs its planned operation; for `ci-evidence`, pass the deploy run's ID and `headSha`.
6. `skills/smaqit.feature-new/SKILL.md` (L134, L157, L161, L300); `skills/smaqit.project-new/SKILL.md` (L34 mesh precondition limited to local paths, L151-153, L156); `skills/smaqit.input-deployment/SKILL.md` if its plan summary needs it; `skills/smaqit.infrastructure-deploy-k3s-app/SKILL.md` mesh-check text (L57-72, L340) states the plan skips it on CI paths. Version bumps.
7. Docs: `docs/wiki/workflows/deployment-slots.md` (defaults table, paths table, L74's verify note, existing-k3s Targets section), `.smaqit/compendium.md` (~L409), `templates/AGENTS.md.template` L39 only if its slot summary becomes wrong.

*Phase D: task 137*
8. Edit `.smaqit/tasks/137_existing_k3s_credential_slots_scripts_only_vault.md`: remove the `ci-evidence` design decision (L59-62), Implementation Step 5, the AC at L110-111 and the `installer/slots.go` / deploy-verify Files rows; keep its verify note only for the `standard` path's local in-cluster verify under operator-scripts; add a line naming task 142 as the owner of CI-path verification. Update 137's PLANNING.md row if its wording mentions verify.

*Phase E: tests and release*
9. `installer/slots_test.go`: extend `TestSlotPlanDeploymentPaths` with existing-k3s on both CI paths for the `test` (private-mesh), `prod` (platform-owned) and `dev` (public own-tls) fixture rows — `platform-prereqs` none, `verify` `deploy-verify:ci-evidence @local`; a binding-override case; a `--slot=verify` CI-path case. `TestSlotPlanDefaultsMatchCurrentBehaviour` (standard path) stays unchanged.
10. `installer/skill_scripts_test.go`: `ci-evidence.sh` against stub `gh` JSON — pass for in-cluster, pass for external, wrong SHA, failed step, applicable step skipped, missing step, unfinished run.
11. `scripts/smoke-test-installer.sh`: assert the existing-k3s `existing-cicd-pr` plan; update the deployment bench fixture's stub plan if it pins CI-path verify.
12. `make -C installer test` and `make -C installer smoke-test`; CHANGELOG `[Unreleased]` (Changed, with a note that CI-path existing-k3s deploys no longer need the operator's machine on the mesh or a local kubeconfig); release via task-complete (`Prepare release vX.Y.Z`, MINOR).

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] On `existing-cicd-pr` and `ci-push`, the existing-k3s plan resolves `platform-prereqs` to none and `verify` to `smaqit.infrastructure-deploy-verify:ci-evidence @local` for private-mesh, platform-owned and public rows; the `standard` path and every VM mode resolve exactly as before; a project binding still overrides both slots
- [ ] `ci-evidence.sh` passes only when `gh run view <id> --json headSha,conclusion,jobs` shows the expected SHA, a successful run, a successful "Verify revision" and the row's applicable verify step succeeded, and fails with a distinct line for a wrong SHA, a failed, skipped or missing step, or an unfinished run (stub `gh` JSON tests); it reads no kubeconfig, Vault or logs
- [ ] deploy-verify declares `ci-evidence` in `metadata.smaqit-slots` and `TestDefaultFillersDeclareTheirSlots` passes
- [ ] `agents/deployment.md`, feature-new, project-new (including its mesh precondition), input-deployment where needed, and deploy-k3s-app's mesh-check text describe `verify` by its planned operation and no longer require the operator's machine on the mesh for CI paths
- [ ] `docs/wiki/workflows/deployment-slots.md` and `.smaqit/compendium.md` reflect the CI-path defaults
- [ ] Task 137 no longer carries the `ci-evidence` decision, step or criterion, and points to 142 for CI-path verification
- [ ] `make -C installer test` and `make -C installer smoke-test` pass; CHANGELOG updated; MINOR release published via the task PR

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
