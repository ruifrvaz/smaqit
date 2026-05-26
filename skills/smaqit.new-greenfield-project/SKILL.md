---
name: smaqit.new-greenfield-project
description: Use when orchestrating the complete SDLC for a new project — from raw project assets to a running production application accessible via browser. Covers requirements extraction, specification (business, functional, stack, infrastructure, coverage), task creation, development, IaC generation + dev environment sweep (local provisioning + deploy + verify), CI/CD production deployment, optional domain/TLS, validation, and tagged release. Re-entrant: use the pre-condition checklist to resume at any phase. Also use when the user says "take this from zero to prod", "run the full smaqit pipeline", "deploy a new project end-to-end", or when starting implementation on a freshly initialized repository.
metadata:
  version: "1.0.0"
---

# Project: Zero to Production

## Steps

### Pre-conditions

All items below must be satisfied before starting. When re-entering at a later phase, confirm only the items for that phase and all earlier phases.

**Always required**
- [ ] Raw project assets in `assets/raw/` (code, docs, requirements)
- [ ] `gh` CLI authenticated (`gh auth login`)
- [ ] Git repository created on GitHub (public or private)

**Required before Phase 4 (Dev Sweep)**
- [ ] Cloud account available (Cyso or equivalent)
- [ ] Application credential created in cloud portal; loaded into local Vault at `secret/<project-slug>/cyso`
- [ ] Object Storage state bucket created (with separate state keys for dev and prod); S3 keys loaded into Vault at `secret/<project-slug>/tfstate`
- [ ] SSH keypair generated (passphrase-free); both keys loaded into Vault at `secret/<project-slug>/ssh`
- [ ] Fine-grained PAT with `variables:write` loaded into Vault at `secret/<project-slug>/github`
- [ ] Local Vault initialised and running on 127.0.0.1:8200 (`smaqit.infrastructure-vault-loader` one-time setup complete)

<!-- amendment: 2026-05-25 — Phase 4 pre-conditions updated to require local Vault as credential source. Manual exports, OpenRC file, and SSH key disk paths removed. smaqit.infrastructure-vault-loader is now the gate before Phase 4 execution. -->

**Required before Phase 6 (Domain/TLS)**
- [ ] Domain purchased at registrar
- [ ] DNS A record set to VM fixed IP

---

### Phase 0 — Task Creation (Entry Point)

The operator triggers this phase manually and sets execution mode before any work begins.

1. Decide execution mode:
   - **Assisted** — operator is present at each gate; phases do not advance without explicit approval.
   - **Autonomous** — all phases run sequentially without gate interruptions; operator reviews at Phase 8.
2. Invoke `smaqit.task-create` once for each of Phases 1–7 (include Phase 6 if domain/TLS is planned). Each task covers one phase; acceptance criteria are sourced from the respective phase gate.
3. **Gate:** All task files created in `.smaqit/tasks/`. Operator confirms mode and approves the task set.

### Phase 1 — Requirements Extraction

1. Invoke `smaqit.task-start` for the Phase 1 task.
2. Invoke `smaqit.requirements-extract`.
3. Review the flagged ambiguities with the user. Resolve any that would block specification (e.g., conflicting data model shapes, undefined score ranges).
4. **Gate:** Confirm extracted inventory is sufficient to proceed to specs.
5. Invoke `smaqit.task-complete` for the Phase 1 task.

### Phase 2 — Specification

Run each spec agent sequentially. Each agent reads the previous layer's output and invokes its own input skill internally.

1. Invoke `smaqit.task-start` for the Phase 2 task.
2. Invoke `@smaqit.business` agent.
3. Invoke `@smaqit.functional` agent.
4. Invoke `@smaqit.stack` agent.
5. Invoke `@smaqit.infrastructure` agent.
6. Invoke `@smaqit.coverage` agent.
7. **Gate:** All specs have `status: draft` and acceptance criteria written. User reviews and approves the full spec set.
8. Invoke `smaqit.task-complete` for the Phase 2 task.

### Phase 3 — Development

1. Invoke `smaqit.task-start` for the Phase 3 task.
2. Invoke `@smaqit.development` agent to implement all specs with `status: draft`.
3. If any spec requires amendment to proceed: amend the spec in-place with an `amendment:` annotation and continue. Structural divergences that change architecture must be paused for operator approval before continuing.
4. **Gate:** Build passes (backend and frontend). All MVP acceptance criteria met. Development agent sets specs to `status: implemented`.
5. Invoke `smaqit.task-complete` for the Phase 3 task, ensuring any amendments are captured under `Decisions made`.

### Phase 4 — Dev Environment Sweep

Validates the full infrastructure and deployment approach on a dedicated dev VM before committing to CI/CD.

1. Invoke `smaqit.task-start` for the Phase 4 task.
2. Invoke `smaqit.infrastructure-vault-loader`. Confirm Vault is running, unsealed, and all `secret/<project-slug>/*` paths are populated. Do not proceed until confirmed.
3. Invoke `@smaqit.deployment` agent with context: generate all IaC artifacts — Terraform files in `deployment/terraform/` and GitHub Actions workflow files in `.github/workflows/` using `smaqit.infrastructure-cicd-generate` patterns as reference; Terraform state key `dev/terraform.tfstate`; do not trigger deployment execution.
4. Invoke `smaqit.infrastructure-provision-cyso` with dev environment variables. Note the `fixed_ip` output.
5. Invoke `smaqit.infrastructure-vm-bootstrap` with the dev VM `fixed_ip`.
6. Invoke `smaqit.infrastructure-deploy-rsync` to deploy to the dev VM.
7. Invoke `smaqit.infrastructure-deploy-verify` against the dev VM. If any check fails, stop and fix before continuing.
8. If any infrastructure or stack spec required amendment to proceed: amend in-place with an `amendment:` annotation.
9. Commit all generated IaC artifacts: `git add deployment/ .github/workflows/ && git commit -m "ci: add infrastructure and CI/CD workflows"`.
10. **Gate:** All `deploy-verify` checks PASS on dev VM. IaC artifacts committed.
11. Invoke `smaqit.task-complete` for the Phase 4 task, ensuring any amendments are captured under `Decisions made`.
12. *(Optional)* Tear down dev VM: run `terraform destroy` using dev state to avoid ongoing cloud costs.

### Phase 5 — Production Deployment via CI/CD

Uses IaC artifacts from Phase 4. Configures production secrets, pushes to main, and monitors the triggered pipeline.

1. Invoke `smaqit.task-start` for the Phase 5 task.
2. Invoke `smaqit.infrastructure-vault-loader`. Confirm Vault is running and all credential paths are populated.
3. Invoke `smaqit.infrastructure-repo-config` to sync all production secrets from Vault to GitHub Secrets (cloud credentials, Terraform backend, SSH key).
4. Push to main: `git push origin main`. The `deploy.yml` workflow triggers automatically (provision job → deploy job).
5. Monitor the pipeline: `gh run watch` — wait for the workflow run to complete.
6. Invoke `smaqit.infrastructure-deploy-verify` against the production VM. If any check fails, stop and report.
7. If any spec required amendment during deployment: amend in-place with an `amendment:` annotation.
8. **Gate:** CI/CD run completes successfully. `deploy-verify` reports all checks PASS. Health endpoint returns correct SHA. Deployment agent sets infrastructure specs to `status: deployed`.
9. Invoke `smaqit.task-complete` for the Phase 5 task, ensuring any amendments are captured under `Decisions made`.

### Phase 6 — Domain + TLS (conditional)

Execute only if the domain and DNS pre-conditions are met.

1. Invoke `smaqit.task-start` for the Phase 6 task.
2. Invoke `smaqit.infrastructure-domain-tls`.
3. **Gate:** HTTPS accessible, HTTP redirects to HTTPS, auto-renewal dry-run passes.
4. Invoke `smaqit.task-complete` for the Phase 6 task.

If skipped: application is accessible at `http://<fixed_ip>`. Document as an open item and continue to Phase 7.

### Phase 7 — Validation

1. Invoke `smaqit.task-start` for the Phase 7 task.
2. Invoke `@smaqit.validation` agent.
3. If any spec is found inconsistent with the live system: amend in-place with an `amendment:` annotation.
4. **Gate:** All validation checks pass. User signs off.
5. Invoke `smaqit.task-complete` for the Phase 7 task, ensuring any amendments are captured under `Decisions made`.

### Phase 8 — Release

1. Confirm all phase tasks (1–7) are closed in `PLANNING.md`. If any remain open, resolve before continuing.
2. Run the amendment scan: `bash .github/skills/smaqit.new-greenfield-project/scripts/check-amendments.sh specs/`. If the script reports matches, review each `amendment:` annotation against the `Blockers encountered` and `Follow-up identified` fields of the relevant phase task and confirm all are resolved or accepted. If no matches are found, skip this step entirely.
3. Invoke `smaqit.release-analysis` → `smaqit.release-approval` → `smaqit.release-prepare-files`.
4. Invoke `smaqit.release-git-local` (or `smaqit.release-git-pr` for PR-based releases).
5. **Final output:** Application running at `https://<domain>/` (or `http://<fixed_ip>/` if Phase 6 was skipped), with a tagged release on GitHub.

---

## Output

- Running production application accessible via browser
- Tagged git release with updated `CHANGELOG.md`
- All specs at `status: deployed`
- MVP task closed in `PLANNING.md`

## Scope

- Covers the single-app, two-environment path: dev VM (Phase 4 local sweep) + production VM (Phase 5 CI/CD).
- Does NOT handle database schema migrations. The current project uses SQLite with append-only schema changes.
- Does NOT cover post-MVP feature cycles. Use individual smaqit agents for iterative feature work after this skill completes.
- Phase 6 (domain/TLS) is conditional on domain purchase — a human action outside the system.

## Gotchas

- **Spec amendment protocol** — when an implementation phase must diverge from a spec (package mismatch, config change, structural adaptation): amend the spec in-place with an `amendment:` annotation describing what changed and why. Tactical divergences (versions, minor config) proceed autonomously. Structural divergences (data model, architecture) require operator approval before continuing. At `task-complete` time, the amendment is captured in `Decisions made`. Phase 8 runs `check-amendments.sh` to detect any open annotations; if none are found the review step is skipped.
- **Source path contract** — `smaqit.infrastructure-deploy-rsync` and generated CI/CD workflows assume `backend/` and `frontend/` as local source directories. The `@smaqit.stack` agent (Phase 2) must declare these exact paths in the stack spec. If the project type differs (e.g. api-only), instruct the stack agent explicitly.
- **Separate Terraform state keys** — dev and production must use different state keys (e.g. `dev/terraform.tfstate` vs `prod/terraform.tfstate`). Using the same key causes state conflicts and unintended VM replacement.
- **`GITHUB_TOKEN` reserved name** — enforced in Phase 5 via `smaqit.infrastructure-repo-config`. Never set an env var named `GITHUB_TOKEN` in any workflow to a PAT.
- **PR body sentinel** — Coding Agent must include `smaqit:deploy` as a line in any PR body to trigger post-merge deployment. Must be set at PR creation time, not via label.
- **Build-time Vite vars** — `VITE_DEMO_MODE` and any `VITE_*` vars must be passed to the frontend build step, not at runtime.
- **Floating IP (Cyso)** — use `fixed_ip` from Terraform outputs, not the floating IP. The floating IP does not route on Cyso's flat network.
- **CI/CD idempotency** — if dev VM was not torn down in Phase 4, the `terraform apply` in the Phase 5 CI/CD pipeline will show no changes and skip provisioning, deploying to the existing VM. Ensure the production state key points to a fresh state if a new VM is required.
- **Re-entry** — resume from the first incomplete phase. IaC generation in Phase 4 is idempotent; re-running it overwrites generated files but does not affect cloud resources.
- **Context collapse / phase re-read** — in long sessions the conversation is summarised by the model. Summaries capture phase names and outcomes but not the exact tool calls each phase requires. On resume, the agent operates from the summary's shorthand (e.g. "Phase 7: smoke tests PASS") rather than the SKILL.md instruction set, and substitutes a cheaper action it already ran (e.g. `smaqit.infrastructure-deploy-verify` curl checks from Phase 5) for the correct one (`@smaqit.validation` agent). Mitigation: at every phase boundary, re-read this SKILL.md (`read_file` the full steps for the upcoming phase) before executing any step. Do not rely on session memory or conversation summaries as a substitute for the canonical instruction set.

## Examples

**Input:** `assets/raw/code.txt` contains a React prototype. User invokes `/zero.to.prod`.
**Output:** All 9 phases completed. Application running at `https://himcorp.com/`, tagged as v1.0.0 on GitHub, all specs `status: deployed`, MVP task closed.

## Completion

- [ ] Phase 0: all phase tasks created, execution mode confirmed
- [ ] Phase 1: requirements extracted, ambiguities resolved
- [ ] Phase 2: all specs drafted and approved
- [ ] Phase 3: implementation complete, specs set to `implemented`
- [ ] Phase 4: dev VM provisioned, deployed, and verified; IaC artifacts committed
- [ ] Phase 5: CI/CD pipeline succeeded, production verified
- [ ] Phase 6: TLS live (or documented as open item)
- [ ] Phase 7: validation complete
- [ ] Phase 8: all phase tasks confirmed closed, release tagged, application accessible via browser

## Failure Handling

| Situation | Action |
|-----------|--------|
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| Output artifact already exists | Confirm with user before overwriting |
| Phase fails with a hard blocker | Stop at the gate. Report the blocker. Do not advance to the next phase. |
| Specification agent returns incomplete output | Re-run the agent with additional context or user clarification. Do not advance with incomplete specs. |
| `deploy-verify` fails | Stop. Report the failing check. Do not mark deployment as complete. |
| Phase 6 skipped (no domain) | Document as open item. Continue to Phase 7. |
