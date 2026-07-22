---
name: smaqit.feature-new
description: >-
  Use when adding a post-MVP feature to a project that has already completed a `smaqit.new-greenfield-project`
  run (or equivalent) and has a deployed target. Applies greenfield's task-per-phase discipline and amendment
  gate to iterative feature work — spec revalidation, development, deployment, validation, close-out — without
  requirements extraction, from-scratch specs, or a dev-VM sweep. Defaults deployment to the existing target
  instead of provisioning a new VM. Also use when the user says "add a feature to this project", "iterate on
  the deployed app", or asks for post-MVP work and the project already has an Infrastructure spec with
  `status: deployed`.
metadata:
  version: "1.0.0"
---

# Post-MVP Feature Workflow

## Steps

### Pre-conditions

All items below must be satisfied before starting. When re-entering at a later phase, confirm only the items for that phase and all earlier phases.

**Always required**
- [ ] Project has already completed at least one MVP cycle — an Infrastructure spec exists under `specs/infrastructure/`
- [ ] `gh` CLI authenticated (`gh auth login`)
- [ ] Feature requirements described in session context (no `assets/raw/` sweep — this is not a from-scratch requirements extraction)

**Required before Phase 3 (Deployment), deploy-now path only**
- [ ] Local Vault initialised and running with this project's existing credential paths populated (`smaqit.infrastructure-vault-loader` already completed during the original MVP cycle — re-run only if credentials have rotated or expired)

If no Infrastructure spec exists yet under `specs/infrastructure/` — this project has not been through an MVP cycle — stop and flag to the user that `smaqit.new-greenfield-project` is the correct skill instead. Do not proceed silently.

---

### Phase 0 — Task Creation (Entry Point)

The operator triggers this phase manually and sets execution and deployment decisions before any work begins.

1. Decide execution mode:
   - **Assisted** — operator is present at each gate; phases do not advance without explicit approval.
   - **Autonomous** — all phases run sequentially without gate interruptions; operator reviews at Phase 5.
2. Decide deploy-now vs. defer for this feature cycle:
   - **Deploy-now** — Phase 3 pushes to production and verifies before this feature is considered fully closed.
   - **Defer** — Phase 3's implementation and gate work completes, but the actual push is batched behind other in-flight feature work; the Deployment-phase task stays open (not `Completed`) with a note pointing at what it's batched behind.
3. Invoke `smaqit.task-create` once for each of Phases 1–5 (Spec Revalidation, Development, Deployment, Validation, Close-out).
4. **Gate:** All task files created in `.smaqit/tasks/`. Operator confirms mode, deploy-now/defer, and approves the task set.

### Phase 1 — Spec Revalidation

Not from-scratch spec generation — reuses the Incremental Development model and the Incremental Spec Updates decision table so specs stay a single source of truth as the feature lands.

1. Invoke `smaqit.task-start` for the Phase 1 task.
2. Invoke `/smaqit.business` → `/smaqit.functional` → `/smaqit.stack` → `/smaqit.infrastructure` → `/smaqit.coverage` as needed. At each, apply the Incremental Spec Updates decision table in [references/spec-lifecycle-reference.md](references/spec-lifecycle-reference.md).
3. Run `smaqit plan --phase=develop` (no `--regen`) to confirm scope — see the Incremental Plan Resolution table in the same reference for exact behavior.
4. For any spec needing only a status bump with no content change (e.g. re-confirming `implemented` still holds after this feature's tests pass), use `smaqit.spec-status-update` rather than re-invoking a full spec agent.
5. **Gate:** All touched specs have `status: draft` (new/updated) or their existing status confirmed still accurate. User reviews and approves the touched spec set.
6. Invoke `smaqit.task-complete` for the Phase 1 task.

### Phase 2 — Development

1. Invoke `smaqit.task-start` for the Phase 2 task.
2. Invoke `/smaqit.development` agent to implement all specs with `status: draft`, explicitly instructing it to use the canonical `<!-- amendment: DATE — description -->` tag for any spec divergence (package mismatch, config change, structural adaptation) — not a prose blockquote. `check-amendments.sh` (Phase 3) matches the literal substring `amendment:`, case-sensitively; a differently-worded or differently-cased note will not be caught.
3. **Gate:** Build passes. All this feature's acceptance criteria met. Development agent sets touched specs to `status: implemented`.
4. Invoke `smaqit.task-complete` for the Phase 2 task, ensuring any amendments are captured under `Decisions made`.

### Phase 3 — Deployment

No dev-VM sweep. This is the Nth feature landing on a deploy pipeline the project's original MVP cycle already proved — deploys straight to the resolved existing target.

1. Resolve `provisioning_mode` before invoking `smaqit.input-deployment`: if `specs/infrastructure/*.md` already declares a target with `status: deployed`, default `provisioning_mode` to `existing-owned` — this overrides `smaqit.input-deployment`'s own generic `provision` default, which is correct for its primary caller (`smaqit.new-greenfield-project`'s brand-new-project case) but wrong here. Only fall through to `smaqit.input-deployment`'s standard elicitation if genuinely ambiguous (e.g. session context mentions co-hosting on a VM another project owns — resolves to `existing-shared` instead).
   → If resolution falls through to `provision` (no Infrastructure spec shows a deployed target) — this contradicts the Pre-conditions check already passed; stop and flag to the user rather than silently provisioning a new VM.
2. Invoke `smaqit.task-start` for the Phase 3 task.
3. Run the amendment gate: `bash [SMAQIT_SKILLS_DIR]/smaqit.new-greenfield-project/scripts/check-amendments.sh specs/`. Reused by reference — do not fork or duplicate the script into this skill's own directory (same precedent as `smaqit.infrastructure-deploy-rsync/scripts/write-vhost.sh` being shared across the deploy-skill family). Run this every time, deploy-now or defer: an unresolved amendment blocks `Completed` on this phase's task even in `defer` mode, so a deferred deploy doesn't quietly lose track of what it's deferred until. If the script reports matches, review each `amendment:` annotation against the `Blockers encountered`/`Follow-up identified` fields of the Phase 2 task and confirm all are resolved or accepted before continuing.
4. **If deploy-now:**
   a. Push (or open a PR, per this project's CI/CD convention) — reuse the existing CI/CD pipeline generated during the original MVP cycle; do not regenerate Terraform IaC or re-run `smaqit.infrastructure-cicd-generate`, that already exists for a project past MVP.
   b. Monitor the pipeline: `gh run watch` (or the project's PR-merge equivalent).
   c. Invoke `smaqit.infrastructure-deploy-verify` against the target. If any check fails, stop and report.
   d. **Gate:** Amendment gate clear (or resolved). CI/CD run completes successfully. `deploy-verify` reports all checks PASS.
   e. Invoke `smaqit.task-complete` for the Phase 3 task, ensuring any amendments are captured under `Decisions made`.
5. **If defer:**
   a. Record in the Phase 3 task what this deploy is batched behind (e.g. a later task ID it will ship alongside).
   b. Leave the Phase 3 task open — set status `Blocked` (or equivalent), not `Completed` — even though the amendment gate already ran clean in step 3.
   c. Do not invoke `smaqit.task-complete` for this task; Phase 5 (Close-out) reads this open-with-reason state, not a completion.

### Phase 4 — Validation

1. Invoke `smaqit.task-start` for the Phase 4 task.
2. Invoke `/smaqit.validation` agent.
3. If any spec is found inconsistent with the live system: amend in-place with the canonical `amendment:` annotation.
4. **Gate:** All validation checks pass. User signs off.
5. Invoke `smaqit.task-complete` for the Phase 4 task, ensuring any amendments are captured under `Decisions made`.

### Phase 5 — Close-out

1. Confirm all phase tasks (1–4, and Phase 3 under deploy-now) are closed in `PLANNING.md`. Phase 3 may legitimately remain open under `defer` — confirm it carries an explicit reason and what it's batched behind, not just an open status with no note.
2. Re-run the amendment scan (belt-and-suspenders, matches greenfield Phase 8): `bash [SMAQIT_SKILLS_DIR]/smaqit.new-greenfield-project/scripts/check-amendments.sh specs/`. If no matches, skip the review step.
3. Release tagging is a parameter, not unconditional: only invoke `smaqit.release-analysis` → `smaqit.release-approval` → `smaqit.release-prepare-files` → `smaqit.release-git-local` (or `smaqit.release-git-pr`) if this run actually deployed (deploy-now path, Phase 3 closed `Completed`). Under `defer`, skip release tagging entirely — it belongs to whichever future run actually performs the batched deploy.
4. Invoke `smaqit.task-complete` for the Phase 5 task.

---

## Output

- Feature implemented, specs updated in place (not regenerated from scratch)
- Deployed to the existing target (deploy-now) or cleanly deferred with a recorded reason (defer)
- No unresolved `amendment:` annotations at close-out
- All phase tasks closed in `PLANNING.md`, or explicitly open-with-reason (Phase 3 under `defer`)

## Scope

- Covers post-MVP iterative feature work on a project that already has a deployed target. Requires an existing Infrastructure spec with `status: deployed` (or resolves `provisioning_mode` to `existing-shared` for a co-hosted target another project manages).
- Does NOT cover greenfield project setup — no requirements extraction, no from-scratch 5-layer spec generation, no dev-VM sweep. Use `smaqit.new-greenfield-project` for that.
- Does NOT retrofit `smaqit.new-greenfield-project` itself to invoke this skill for its own post-Phase-8 iterative work — that is a sensible future task once this skill is proven, not in scope here.
- Does NOT build an automated "is this project past MVP" detector. The Pre-conditions check (Infrastructure spec with a deployed target) and the `provisioning_mode` resolution fallback (flag to the user if it resolves to `provision`) are the only maturity signals used.

## Gotchas

- **`check-amendments.sh` matches a bare substring, not the canonical tag format.** The script (`grep -rl "amendment:" <dir>`) is case-sensitive and matches the literal substring `amendment:` anywhere in a file — it does not require the full `<!-- amendment: DATE — description -->` HTML-comment form. A note using different wording or capitalization (e.g. `**Amendment (...)**`) will not match. Phase 2 instructs the canonical lowercase tag specifically so the gate fires.
- **`provisioning_mode` default-override lives in this skill, not in `smaqit.input-deployment`.** `smaqit.input-deployment`'s own default of `provision` is correct for its primary caller (`smaqit.new-greenfield-project`'s brand-new-project case); do not modify that skill. The override — defaulting to `existing-owned` when an Infrastructure spec already shows a deployed target — is this skill's own Phase 3 Step 1 logic.
- **No dev-VM sweep, by design.** A dev-environment validation pass makes sense for a brand-new deploy pipeline being proven for the first time; it is wasted cloud spend and time for the Nth feature landing on a pipeline already proven. If `provisioning_mode` ever resolves to `provision` here, that is itself a signal this project isn't actually past MVP — stop and flag rather than silently provisioning.
- **Amendment gate runs every time, deploy-now or defer.** Deferred deploys still run the gate at Phase 3 — deferring the push must not mean deferring visibility into unresolved spec divergence.
- **`check-amendments.sh` is reused by reference, never forked.** Invoke it via `[SMAQIT_SKILLS_DIR]/smaqit.new-greenfield-project/scripts/check-amendments.sh` — the placeholder this repo's compile step (`scripts/generate-agents.py`) resolves per platform. A hand-written relative path (`../smaqit.new-greenfield-project/...`) is not how any existing cross-skill reference in this repo works and will not survive compilation.
- **Context collapse / phase re-read** — in long sessions the conversation is summarised by the model, and summaries capture phase names and outcomes but not the exact tool calls each phase requires. At every phase boundary, re-read this SKILL.md for the upcoming phase before executing any step rather than relying on a conversation summary.

## Examples

**Input:** Project has a deployed MVP (Infrastructure spec `status: deployed`). User asks to add an Identity & Access accounts/login feature, deploy-now.
**Output:** Phase 1 updates/creates the touched specs and confirms scope via `smaqit plan --phase=develop`. Phase 2 implements with any divergence (e.g. a new runtime dependency, a new application secret) recorded via the canonical `amendment:` tag. Phase 3 resolves `provisioning_mode` to `existing-owned`, runs the amendment gate (which now correctly catches the recorded divergence and blocks until it's resolved), then deploys and verifies. Phase 4 validates. Phase 5 confirms all phases closed, re-scans for amendments, and tags a release since this run deployed.

## Completion

- [ ] Phase 0: all phase tasks created, execution mode and deploy-now/defer confirmed
- [ ] Phase 1: touched specs revalidated per the Incremental Spec Updates decision table, scope confirmed via `smaqit plan --phase=develop`
- [ ] Phase 2: implementation complete, specs set to `implemented`, any divergence recorded via canonical `amendment:` tag
- [ ] Phase 3: `provisioning_mode` resolved with the existing-target-first override; amendment gate run and clear; deploy-now path verified, or defer path recorded with an open task and reason
- [ ] Phase 4: validation complete
- [ ] Phase 5: all phase tasks confirmed closed or explicitly deferred-with-reason; amendment re-scan clear; release tagged only if this run deployed

## Failure Handling

| Situation | Action |
|-----------|--------|
| No Infrastructure spec with `status: deployed` exists (Pre-conditions) | Stop. Flag that `smaqit.new-greenfield-project` is the correct skill instead. |
| `provisioning_mode` resolves to `provision` despite Pre-conditions passing | Stop. Flag to the user before provisioning a new VM. |
| Amendment gate reports unresolved matches | Stop at the Phase 3 gate. Resolve or explicitly accept each annotation before continuing, in both deploy-now and defer paths. |
| `deploy-verify` fails | Stop. Report the failing check. Do not mark the Phase 3 task as complete. |
| Spec agent returns incomplete output | Re-run with additional context or user clarification. Do not advance with incomplete specs. |
| Subagent invocation fails | Report the failure with context; do not silently retry. |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification. |
