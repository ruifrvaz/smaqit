---
name: smaqit.feature-deploy
description: Use when deploying a feature to an existing project's production environment through its already-proven CI/CD pipeline. Reads specs to identify the stack, target, and provisioning mode before any infrastructure work. Handles all three provisioning_mode branches (provision, existing-owned, existing-shared) with infrastructure readiness, CI/CD-triggered deploy, verification, amendment gate, and optional release tagging. No dev-VM sweep, no deploy-rsync skills — the Nth deploy on infrastructure already proven. Also use when the user says "deploy this", "push to production", "ship this feature", or asks to deploy to a project that already has CI/CD workflows from its initial greenfield run.
metadata:
  version: "1.0.0"
---

# Feature Deploy — Standalone Post-MVP Deployment

## Steps

### Pre-conditions

All items below must be satisfied before starting the Pre-phase. If any fail, stop and direct the user to `smaqit.new-greenfield-project`.

- [ ] Infrastructure spec exists under `specs/infrastructure/` with `status: deployed`
- [ ] CI/CD workflows exist (`.github/workflows/deploy.yml` at minimum)
- [ ] `gh` CLI authenticated (`gh auth login`)
- [ ] Local Vault initialised and running with this project's credential paths populated

---

### Pre-phase — Feature Identification

Identify what is being deployed before any infrastructure work begins.

1. Read `specs/stack/platform-stack.md` — the authoritative stack declaration. Do not re-derive from the filesystem.
2. Read `specs/infrastructure/*.md` — determine deployment topology, target environment, and current status.
3. Resolve `provisioning_mode`:
   - If Infrastructure spec shows `status: deployed`, default to `existing-owned`. This overrides `smaqit.input-deployment`'s generic `provision` default, which is correct for greenfield's brand-new-project case but wrong here.
   - Only fall through to `smaqit.input-deployment`'s standard elicitation if genuinely ambiguous (e.g. session context mentions co-hosting on another project's VM — resolves to `existing-shared`).
   - If resolution falls through to `provision` despite pre-conditions passing: stop and flag — this contradicts the existence of a deployed Infrastructure spec.
4. Confirm CI/CD workflows exist (`.github/workflows/deploy.yml` at minimum).
5. **Gate:** Stack identified, `provisioning_mode` resolved, CI/CD confirmed present.

---

### Phase 1 — Infrastructure Readiness

Applies all three `provisioning_mode` branches.

1. Invoke `smaqit.infrastructure-vault-loader`. Confirm Vault is running, unsealed, and credential paths are populated.
   - **`existing-shared`:** only `secret/apps/<app-slug>/github` is loaded; `cyso`/`tfstate` are never prompted for. Then run `bootstrap-app-to-machine.sh <app-slug> <machine-slug>` to populate `secret/apps/<app-slug>/ssh`.
2. Invoke `smaqit.infrastructure-provision-cyso`.
   - **`existing-owned`:** `terraform apply` is expected to no-op (gated by `plan-guard.sh`) — idempotent behaviour.
   - **`existing-shared`:** **skip this step entirely.** This project never provisions Terraform for a VM it doesn't own.
3. Invoke `smaqit.infrastructure-vm-bootstrap` with the target IP (from Terraform `fixed_ip` output, or manually-set `VM_HOST` for existing-shared).
4. **Gate:** SSH to VM succeeds; all infrastructure preconditions confirmed.

---

### Phase 2 — Deploy via CI/CD

Push to main, trigger the existing CI/CD pipeline, monitor, and verify.

1. Invoke `/smaqit.deployment` agent for spec validation, frontmatter updates, and deployment report. Pass the resolved `provisioning_mode` and target context from session context.
2. Invoke `smaqit.infrastructure-repo-config` to sync secrets from Vault to GitHub.
   - **`existing-shared`:** restricted mode — skips `tfstate`/`cyso` (absent paths), syncs only `ssh` + `github`-derived secrets. Additionally: `gh variable set VM_HOST --body <shared-vm-ip>`.
3. Push to main:
   ```
   git push origin main
   ```
   The existing `deploy.yml` workflow triggers automatically.
   - **`existing-shared`:** `deploy.yml` has only a `deploy` job (no `provision` job) — generated that way by `cicd-generate`'s `deploy-only` mode during greenfield.
4. Monitor the pipeline:
   ```
   gh run watch
   ```
   Wait for the workflow run to complete.
5. Invoke `smaqit.infrastructure-deploy-verify` against production. If any check fails, stop and report.
6. Run the amendment gate:
   ```
   bash [SMAQIT_SKILLS_DIR]/smaqit.feature-deploy/scripts/check-amendments.sh specs/
   ```
   If matches found, review each against the feature's task files and resolve before continuing.
7. **Gate:** CI/CD run completes successfully. `deploy-verify` reports all checks PASS. Amendment gate clear.

---

### Phase 3 — Close-out

1. Confirm deployment report written to `.smaqit/reports/`.
2. Re-run amendment scan (belt-and-suspenders):
   ```
   bash [SMAQIT_SKILLS_DIR]/smaqit.feature-deploy/scripts/check-amendments.sh specs/
   ```
3. Release tagging is conditional:
   - **Standalone invocation** (not part of a `smaqit.feature-new` cycle): invoke `smaqit.release-analysis` → `smaqit.release-approval` → `smaqit.release-prepare-files` → `smaqit.release-git-local` (or `smaqit.release-git-pr`).
   - **Invoked from within `smaqit.feature-new`:** skip — `feature-new` owns release tagging for its own cycle.
4. **Final output:** Application running in production; deployment verified; specs at `status: deployed`; release tagged (if standalone).

---

## Output

- Deployed and verified application in production
- Deployment report in `.smaqit/reports/`
- All specs updated to `status: deployed`
- Tagged release (if standalone invocation)

## Scope

- Covers post-MVP deployment for projects that have already completed a greenfield run and have existing CI/CD workflows.
- Covers all three `provisioning_mode` values (`provision`, `existing-owned`, `existing-shared`).
- Does NOT cover first-time deployment — use `smaqit.new-greenfield-project`.
- Does NOT include a dev-VM sweep — deploys straight to the resolved production target.
- Does NOT use deploy-rsync skills — deployment goes through the existing CI/CD pipeline.
- Does NOT modify `agents/deployment.md` — invokes it as-is for spec validation and reports.
- Release tagging is skipped when invoked from within `smaqit.feature-new`.

## Examples

**Input:** Project has a deployed MVP. User has implemented a feature (specs updated, code committed). User says "deploy this to production."

**Output:** Pre-phase reads stack spec (Node.js + Vite/React) and infra spec (status: deployed, Cyso VM). Resolves `provisioning_mode` to `existing-owned`. Phase 1 confirms Vault credentials loaded, `provision-cyso` no-ops (`plan-guard` confirms no changes), `vm-bootstrap` confirms VM ready. Phase 2 invokes deployment agent, syncs secrets via `repo-config`, pushes to main, monitors `deploy.yml` workflow, runs `deploy-verify` (all checks PASS), runs amendment gate (clear). Phase 3 confirms report written, re-scans amendments, tags release v1.3.0.

## Gotchas

- **Provisioning mode default-override lives here, not in `smaqit.input-deployment`.** `smaqit.input-deployment`'s default of `provision` is correct for greenfield's brand-new-project case. The override — defaulting to `existing-owned` when an Infrastructure spec shows `status: deployed` — is this skill's own Pre-phase logic. Do not modify `smaqit.input-deployment`.
- **CI/CD must already exist.** This skill does not generate workflows — `smaqit.infrastructure-cicd-generate` ran during greenfield. If `.github/workflows/deploy.yml` is absent, stop and flag.
- **No deploy-rsync skills.** Post-MVP deploys go through CI/CD (GitHub Actions). The deploy-rsync family is for greenfield's dev-VM sweep only.
- **Amendment gate runs every time.** `check-amendments.sh` is shipped with this skill at `scripts/check-amendments.sh`. Run it in Phase 2 (post-deploy) and again in Phase 3 (close-out).
- **`/smaqit.deployment` agent is NOT modified.** It is invoked as-is for spec validation, frontmatter updates, and deployment report. The CI/CD push/monitor/verify sequence is this skill's own prose.

## Completion

- [ ] Pre-phase: stack identified, `provisioning_mode` resolved, CI/CD confirmed
- [ ] Phase 1: vault-loader confirmed, infrastructure ready (branched correctly on `provisioning_mode`)
- [ ] Phase 2: CI/CD run succeeded, `deploy-verify` PASS, amendment gate clear
- [ ] Phase 3: deployment report written, amendment re-scan clear, release tagged (if standalone)

## Failure Handling

| Situation | Action |
|-----------|--------|
| Required input not provided | Request the missing information before proceeding. |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification. |
| Subagent invocation fails | Report the failure with context; do not silently retry. |
| Output artifact already exists | Confirm with user before overwriting. |
| No Infrastructure spec with `status: deployed` | Stop. Flag that `smaqit.new-greenfield-project` must complete first. |
| No CI/CD workflows exist | Stop. Flag that greenfield's `cicd-generate` step wasn't completed. |
| `provisioning_mode` resolves to `provision` despite pre-conditions passing | Stop. Flag the contradiction before provisioning a new VM. |
| `deploy-verify` fails | Stop. Report the failing check. Do not mark deployment as complete. |
| Amendment gate reports unresolved matches | Stop at Phase 2 gate. Resolve or explicitly accept each annotation before continuing. |
| CI/CD workflow fails | Stop. Report the failure. Do not proceed to verify or close-out. |
| Vault not running or credentials missing | Stop. Direct user to run `smaqit.infrastructure-vault-loader` first. |
