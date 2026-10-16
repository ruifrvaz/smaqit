---
name: smaqit.new-greenfield-project
description: 'Use when taking a new project from raw assets to a running production application: "take this from zero to prod", "run the full smaqit pipeline", or starting a freshly initialized repository. Orchestrates requirements extraction, all five specification layers, development, a dev-environment sweep, CI/CD production deployment, optional domain/TLS, validation, and release. Re-entrant at any phase.'
metadata:
  version: "2.4.0"
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

Applicability below depends on `provisioning_mode` (resolved by `smaqit.input-deployment` — see "Provisioning Mode and Deployment Slots" above Phase 4). Items marked `[provision/existing-owned]` apply when this project provisions or owns the target VM's Terraform state; items marked `[existing-shared/existing-unmanaged]` apply when targeting a VM this project does not manage via its own Terraform state — either because a different project owns it (`existing-shared`, co-hosted) or because nobody's Terraform manages it at all (`existing-unmanaged`, dedicated); items marked `[existing-k3s]` apply to a Namespace-scoped cluster target. A project that binds its own fillers in `.smaqit/slots.yml` follows those fillers' own pre-conditions for the slots it binds.

- [ ] `[provision/existing-owned]` Cloud account available (Cyso or equivalent)
- [ ] `[provision/existing-owned]` Application credential created in cloud portal; loaded into local Vault at `secret/<project-slug>/cyso`
- [ ] `[provision/existing-owned]` Object Storage state bucket created (with separate state keys for dev and prod); S3 keys loaded into Vault at `secret/<project-slug>/tfstate`
- [ ] `[existing-shared/existing-unmanaged]` The target machine is known by machine-slug: for `existing-shared` it must already be registered (its `secret/machines/<machine-slug>/base-ssh` exists); for `existing-unmanaged` it typically registers for the *first* time. Phase 4's `target-credentials` slot then bootstraps this app's own SSH keypair at `secret/apps/<app-slug>/ssh` via `bootstrap-app-to-machine.sh` (see `smaqit.infrastructure-vault-loader`)
- [ ] `[existing-shared/existing-unmanaged]` The production host is already bootstrapped (Phase 5's `ci-push` path runs no `host-bootstrap` by default; bind one in `.smaqit/slots.yml` if a separate production host needs it)
- [ ] `[all modes]` Fine-grained PAT with `variables:write` loaded into Vault at `secret/apps/<app-slug>/github` (legacy flat-scheme projects: `secret/<project-slug>/github`)
- [ ] `[existing-k3s]` Fine-grained PAT with `contents:write` + `pull_requests:write` on the platform repository only, loaded at `secret/apps/<app-slug>/platform-repo` by the `access-credentials` load (`smaqit.infrastructure-vault-loader`)
- [ ] `[existing-shared/existing-unmanaged]` Target VM's fixed IP known; will be set via `gh variable set VM_HOST` rather than read from a Terraform output
- [ ] `[existing-k3s]` Infrastructure spec Constraints declare `Platform Repo`, `Registry File Path`, `Namespace Quota`, and `Container Registry`, and its `## Environments` table has a complete row for test and prod (`Machine Slug`, `Host`, `Reachability`, `Ingress Mode`, `Ingress Class`, `ClusterIssuer`, `Health Path`, mesh fields for `private-mesh`), plus the mesh-join and `Probe Client` Constraints rows those rows need — all sourced from the platform team's onboarding documentation; `smaqit spec environments` passes. No SSH keypair, cloud account, or Terraform state is needed
- [ ] `[existing-k3s]` For a `private-mesh` environment: this machine can join the platform's mesh (for local slots and verification), and the platform has issued the app's CI setup key
- [ ] `[all modes]` `smaqit slots plan --phase=deploy --mode=<mode> --env=<dev-or-test-env>` resolves without `SLOTS-*` errors (an untouched `.smaqit/slots.yml` uses every default)
- [ ] `[all modes]` Local Vault initialised and running on 127.0.0.1:8200 (`smaqit.infrastructure-vault-loader` one-time setup complete)

<!-- amendment: 2026-05-25 — Phase 4 pre-conditions updated to require local Vault as credential source. Manual exports, OpenRC file, and SSH key disk paths removed. smaqit.infrastructure-vault-loader is now the gate before Phase 4 execution. -->

**Required before Phase 6 (Domain/TLS)**
- [ ] Domain purchased at registrar
- [ ] DNS A record set to VM fixed IP (VM modes; `existing-k3s` uses the Ingress host the platform routes — for a `private-mesh` environment, a name in the mesh's private zone that the platform owns, so there is no public DNS or registrar step)

---

### Agent Invocation and Return Signals (applies to every phase)

Invoke smaqit agents by name: through the Task tool on Claude Code, `runSubagent` on GitHub Copilot, or by spawning the named subagent on Codex. A phase agent's slash command (`/smaqit.development`, `/smaqit.deployment`, `/smaqit.validation`) is equivalent where it exists. Specification agents have no slash command on Claude Code.

Pass every phase agent an explicit parameter block. It holds `mode` (the Phase 0 execution mode), `specification_mode`, `spec_paths`, and any deployment parameters, because a subagent does not see this conversation. Specifications the user approved in Phase 2 are always handed on as `specification_mode: prevalidated` with exact `spec_paths`; otherwise the phase agent would find them still `draft` and generate them again.

When an agent returns `NEEDS-INPUT` or `REVIEW-REQUESTED`, ask the user what it lists. Then resume that agent with the answers as `answers`, or invoke it again with the same parameters plus `answers` and the signal's `approved_spec_paths`. Repeat until it returns a final result. Never answer a signal on the user's behalf.

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

Run each specification agent sequentially. Each agent reads the previous layer's output and applies its own input skill. An agent that cannot ask the user returns `NEEDS-INPUT` instead of writing anything; handle it as described in Agent Invocation and Return Signals.

1. Invoke `smaqit.task-start` for the Phase 2 task.
2. Invoke the `smaqit.business` specification agent.
3. Invoke the `smaqit.functional` specification agent.
4. Invoke the `smaqit.stack` specification agent.
5. Invoke the `smaqit.infrastructure` specification agent.
6. Invoke the `smaqit.coverage` specification agent.
7. Invoke `smaqit.design-validate` for every generated pair. Confirm `smaqit design validate` passes and the active agent has opened each PNG; PlantUML-source reading is not a visual-review fallback.
8. **Gate:** All specs have `status: draft`, acceptance criteria written, and current visually approved same-layer design pairs. User reviews and approves the full spec/design set. Record the approved specification paths per layer: they are the **Phase 2 spec set** that later phases hand on as `spec_paths`.
9. Invoke `smaqit.task-complete` for the Phase 2 task.

### Phase 3 — Development

1. Invoke `smaqit.task-start` for the Phase 3 task.
2. Invoke the `smaqit.development` agent with `mode` from Phase 0, `specification_mode: prevalidated`, and the Phase 2 Business, Functional, and Stack specs as `spec_paths`. It implements them without invoking the specification agents again.
3. If any spec requires amendment to proceed: amend the spec in-place with an `amendment:` annotation and continue. Structural divergences that change architecture must be paused for operator approval before continuing.
4. **Gate:** Build passes (backend and frontend). All MVP acceptance criteria met. Development agent sets specs and their linked designs to `status: implemented`, respecting the least-advanced linked-spec rule, and reruns `smaqit design validate`.
5. Invoke `smaqit.task-complete` for the Phase 3 task, ensuring any amendments are captured under `Decisions made`.

### Provisioning Mode and Deployment Slots (applies to Phases 4–6)

Before Phase 4 begins, `smaqit.input-deployment` resolves `provisioning_mode` — one of:

- **`provision`** (default) — this project provisions its own new VM via Terraform.
- **`existing-owned`** — redeploying to a VM this project's own Terraform state already manages.
- **`existing-shared`** — targeting a VM a *different* project owns and manages via its own Terraform state (co-hosting).
- **`existing-unmanaged`** — targeting a VM dedicated to this project (not co-hosted) but never managed by any Terraform state — provisioned out-of-band and staying that way by design.
- **`existing-k3s`** — targeting an already-onboarded, Namespace-scoped Kubernetes cluster (k3s or equivalent) that hands the app a scoped kubeconfig and nothing else: no SSH, no host filesystem, no nginx, no Terraform on the app's side. The target is platform-owned, like `existing-shared`, but the deploy mechanism is entirely different — `kubectl apply` into the app's own Namespace via the `smaqit.infrastructure-deploy-k3s-app` family, never rsync/SSH/nginx.
- An app hosted on a **platform repository's** machine (a repository that owns machines rather than an app) is not a separate mode: it is `existing-k3s` for a Kubernetes machine that admits tenants by onboarding, or `existing-shared` for a VM the platform's Terraform owns. The platform repository provisions, baselines, and verifies its machines through its own slot fillers; this project never runs them.

This skill never invokes an infrastructure skill directly in Phases 4–6. Each phase invokes `/smaqit.deployment` with a **slot plan**: a fixed, ordered sequence of deployment slots (`platform-prereqs`, `access-credentials`, `platform-onboarding`, `target-credentials`, `environment-config`, `artifact-build`, `iac-generate`, `ci-generate`, `provision`, `host-bootstrap`, `app-deploy`, `verify`, `tls`; day-two `replace`, `teardown`). `provisioning_mode` selects each slot's default filler; the project's `.smaqit/slots.yml` may bind any slot to its own skill or skip it. `smaqit.input-deployment` resolves the plan with `smaqit slots plan --phase=deploy --mode=<mode> --env=<env> --path=<path> --json` — deterministic, never re-derived by judgment. Run that command yourself at any point to see exactly which skill fills which slot.

### Phase 4 — Dev Environment Sweep

Validates the full infrastructure and deployment approach on a dedicated dev target before committing to CI/CD.

1. Invoke `smaqit.task-start` for the Phase 4 task.
2. Invoke `smaqit.input-deployment` with the dev environment name (`dev`; for `existing-k3s`, the test environment whose machine-slug the Infrastructure spec declares) and `deployment_path: standard`. It resolves `provisioning_mode` and the slot plan. Stop on any `SLOTS-*` failure — fix `.smaqit/slots.yml` or the bound skill first.
3. If the plan warns that a family slot (`app-deploy`, or `artifact-build` for `existing-k3s`) has no installed candidate matching the declared stack, run **Family filler synthesis** (below) for that family, then re-resolve the plan.
4. Invoke `/smaqit.deployment` with the step 2 parameter block (`deployment_path: standard`, the resolved `slot_plan`, the dev environment) plus `mode` from Phase 0 and `specification_mode: prevalidated` with the Phase 2 Infrastructure and Stack specs as `spec_paths`, and — as caller context for the `iac-generate` builtin — Terraform state key `dev/terraform.tfstate`. The agent walks the plan locally, slot by slot, reading each filler's `SKILL.md`. With no project bindings, the defaults reproduce the dev sweep per mode:
   - **`provision`/`existing-owned`:** `access-credentials` (vault-loader for the mode and the project's credential scheme) → `iac-generate` (Terraform in `deployment/terraform/`, state key `dev/terraform.tfstate`) → `ci-generate` (`smaqit.infrastructure-cicd-generate`, full mode) → `provision` (`smaqit.infrastructure-provision-cyso`; note the `fixed_ip` output; for `existing-owned` the guarded apply is expected to no-op) → `host-bootstrap` (`smaqit.infrastructure-vm-bootstrap` on the `fixed_ip`) → `app-deploy` (`smaqit.infrastructure-deploy-rsync*`) → `verify`.
   - **`existing-shared`/`existing-unmanaged`:** `access-credentials` (only `secret/apps/<app-slug>/github`) → `target-credentials` (`bootstrap-app-to-machine.sh <app-slug> <machine-slug>`; for `existing-shared` against the machine's already-registered `base-ssh`; for `existing-unmanaged` typically its fresh-registration branch, using this project's own slug as `owner_project`) → `ci-generate` (`deploy-only` mode: no Terraform, no `provision.yml`) → `host-bootstrap` (on the target's known fixed IP, not a Terraform output) → `app-deploy` → `verify`. No `provision` slot: this project never runs Terraform for a VM it doesn't own or nobody manages.
   - **`existing-k3s`:** `platform-prereqs` (`smaqit.infrastructure-deploy-k3s-app` `mesh-check`: passes for a `public` test environment; for `private-mesh`, confirms this machine reaches the API over the mesh, or stops with instructions) → `access-credentials` (`github` + `platform-repo`, no kubeconfig yet) → `platform-onboarding` (`smaqit.infrastructure-request-k3s-onboarding` for the test machine-slug, gated on its PR merging) → `target-credentials` (the test machine-slug's kubeconfig, pasted from the platform's onboarding hand-off) → `artifact-build` (a matched `smaqit.infrastructure-image-build-push*` skill authors the `Dockerfile` and manifests) → `ci-generate` (`k3s` mode: a `build` job then a kubectl-driven `deploy` job reading `KUBECONFIG` from a GitHub Environment secret) → `app-deploy` (`smaqit.infrastructure-deploy-k3s-app*` against the test Namespace) → `verify` (`https://<test-host>`). No `provision`, `host-bootstrap`, nginx, or `VM_HOST` anywhere.
   - If onboarding reports **pending**, the agent stops at that slot: stop this task and re-invoke later — the onboarding skill re-checks the same PR rather than opening a new one.
   - `existing-shared` (not the first site on the VM): the nginx vhost must be name-based only, never `default_server` — every rsync-family skill enforces this through the shared `smaqit.infrastructure-deploy-rsync/scripts/write-vhost.sh`. `existing-unmanaged` VMs are dedicated, so the same script's live inspection correctly resolves to `default_server`.
   - `existing-k3s` first runs hit two known preconditions documented in the fillers' own Gotchas: `manifest-lint.sh` fails on a missing Ingress or probe manifest until it is rendered once, and a freshly generated `deploy.yml` cannot be `workflow_dispatch`ed until it exists on the default branch. `ci-generate` renders one `deploy.yml` for every environment from the Environments table, so Phase 5 reuses it unchanged.
5. If any infrastructure or stack spec required amendment to proceed: amend in-place with an `amendment:` annotation.
6. Commit the generated artifacts: `git add deployment/ .github/workflows/ && git commit -m "ci: add infrastructure and CI/CD workflows"`. Without an `iac-generate` slot (`existing-shared`/`existing-unmanaged`/`existing-k3s`) there is no `deployment/terraform/`; commit `.github/workflows/` (and, for `existing-k3s`, the Kubernetes manifests under `deployment/`).
7. **Gate:** Every slot in the plan completed or was skipped as planned; `verify` reports all checks PASS on the dev target; artifacts committed.
8. Invoke `smaqit.task-complete` for the Phase 4 task, ensuring any amendments are captured under `Decisions made`.
9. *(Optional)* Tear down the dev target with a single-slot run: resolve `smaqit slots plan --phase=deploy --mode=<mode> --env=dev --slot=teardown`, then invoke `/smaqit.deployment` with that plan and `specification_mode: prevalidated` with the Phase 2 Infrastructure and Stack specs as `spec_paths`. The default destroys only under `provision` (guarded destroy of the dev state). Every other mode defaults to none: `existing-owned` keeps a VM this project intends to retain, `existing-shared`/`existing-unmanaged` have no Terraform state of their own, and an `existing-k3s` test slug is offboarded through the platform's own process.

**Family filler synthesis** (Phase 4 Step 3) — for a family slot with no installed candidate:

- **Precondition:** read the declared stack from `specs/stack/platform-stack.md` — the authoritative source. Do not re-derive the stack independently from the filesystem.
- **Family:** the slot plan names it — `smaqit.infrastructure-deploy-rsync*` (VM modes), `smaqit.infrastructure-deploy-k3s-app*` (`existing-k3s`), or `smaqit.infrastructure-image-build-push*` (`existing-k3s` artifacts). A family's required-inherited-context items never leak into another family.
- a. Report what was checked: the declared stack, which candidates in the family were compared against it, and why none matched.
- b. Check whether `smaqit.create-skill` is available: it and its `smaqit.L2` compiler dependency are installed globally by smaqit-adk (not project-local), so check that both are present among this environment's available skills/agents.
- c. **If available (primary path):** invoke `smaqit.create-skill` with a name derived from the declared stack (e.g. `smaqit.infrastructure-deploy-rsync-<stack-slug>`, `smaqit.infrastructure-deploy-k3s-app-<stack-slug>`), explicitly pointed at the family's existing skill(s) as reference exemplars, and that family's four required-inherited-context items (see Gotchas) as explicit input. The new skill must declare the family's slot in `metadata.smaqit-slots` (e.g. `smaqit-slots: "app-deploy=deploy"`), or the slot plan will not list it as a candidate.
- d. **If unavailable (manual fallback):** author the `SKILL.md` by hand, following the family's existing skill's structural shape (Pre-conditions, numbered Steps, Output, Scope, Gotchas, Completion, Failure Handling), applying that family's four required-inherited-context items and the same `metadata.smaqit-slots` declaration.
- e. **Human checkpoint:** present the synthesized skill to the user before it is used. Skippable only in Autonomous mode with the user's prior, explicit sign-off for that mode.
- f. Re-resolve the slot plan; the synthesized skill now appears as the family candidate the agent selects.
- Tag any synthesized skill with provenance metadata (e.g. `synthesized`, `synthesized-for-project`, `synthesized-date`) marking it a candidate for a future reconciliation into canonical `smaqit` once proven by real use.

### Phase 5 — Production Deployment via CI/CD

Uses the workflows committed in Phase 4. Configures production credentials, pushes to main, and monitors the triggered pipeline — all through the Deployment agent's `ci-push` path.

1. Invoke `smaqit.task-start` for the Phase 5 task.
2. Invoke `smaqit.input-deployment` with environment `prod` (for `existing-k3s`, the prod machine-slug's environment) and `deployment_path: ci-push`. Stop on any `SLOTS-*` failure.
3. Invoke `/smaqit.deployment` with `deployment_path: ci-push`, the resolved `slot_plan`, `mode` from Phase 0, and `specification_mode: prevalidated` with the Phase 2 Infrastructure and Stack specs as `spec_paths`. The agent runs the local slots, then pushes and monitors:
   - **Local slots:** `access-credentials`; for `existing-k3s`, `platform-onboarding` for the prod machine-slug (gated on merge; **pending** stops this task until it merges — and a merge alone isn't enough: the platform's own hand-off must deliver the prod kubeconfig); `target-credentials` (`existing-shared`/`existing-unmanaged`: per-app SSH key; `existing-k3s`: prod kubeconfig — stops this task until the platform has handed it off); `environment-config` (`smaqit.infrastructure-repo-config`: full secret sync for `provision`/`existing-owned`; restricted sync plus `gh variable set VM_HOST` for `existing-shared`/`existing-unmanaged`; for `existing-k3s`, `KUBECONFIG` (plus, for a `private-mesh` machine, the spec-named mesh setup-key secret with `--mesh-secret`, loaded with `MESH_SETUP_KEY=1`) on the GitHub Environment named after the prod machine slug (`sync-k3s-env.sh`), plus `REGISTRY_USERNAME`/`REGISTRY_TOKEN` when the spec's `Container Registry` row declares a private registry — never `VM_HOST`).
   - **CI slots:** the agent pushes `main`, selects the `deploy.yml` `push` run whose `headSha` matches the pushed commit, and watches that run. `existing-k3s`'s `deploy.yml` is dispatch-only, so the agent dispatches it for the prod machine slug and watches that `workflow_dispatch` run instead. The pipeline performs `provision` (full mode's `provision` job, `provision`/`existing-owned` only), `artifact-build` (`existing-k3s` `build` job), and `app-deploy` (`deploy` job).
   - **`verify`:** `smaqit.infrastructure-deploy-verify --expected-sha <deploy-run-headSha>` against production (for `existing-k3s`, `--environment prod`: over HTTPS for a `public` row whose Ingress the app owns, in-cluster for `private-mesh` or `platform-owned`).
   - The agent sets infrastructure specs to `status: deployed` and writes the deployment report with the slot resolution.
4. If any spec required amendment during deployment: amend in-place with an `amendment:` annotation.
5. **Gate:** CI/CD run completes successfully. `verify` reports all checks PASS. Health endpoint returns the deployed SHA. Infrastructure specs are `status: deployed`.
6. Invoke `smaqit.task-complete` for the Phase 5 task, ensuring any amendments are captured under `Decisions made`.

### Phase 6 — Domain + TLS (conditional)

Execute only if the domain and DNS pre-conditions are met.

1. Invoke `smaqit.task-start` for the Phase 6 task.
2. Resolve a single-slot plan: `smaqit slots plan --phase=deploy --mode=<mode> --env=prod --slot=tls --json`. For VM modes the default filler is `smaqit.infrastructure-domain-tls`; for `existing-k3s` it is `none` — TLS is already handled by the prod Environments row's Ingress Mode (`own-tls`: the Ingress `Certificate` was issued and awaited inside `app-deploy`; `platform-default-tls`/`platform-owned`: the platform serves TLS), so record that and skip to the gate.
3. Invoke `/smaqit.deployment` with that single-slot plan and `specification_mode: prevalidated` with the Phase 2 Infrastructure and Stack specs as `spec_paths`.
4. **Gate:** HTTPS accessible, HTTP redirects to HTTPS, auto-renewal dry-run passes (VM modes); for `existing-k3s`, the Ingress `Certificate` is `Ready` (`own-tls`) or the `verify` slot passed (`platform-default-tls`/`platform-owned`, whose TLS the platform owns).
5. Invoke `smaqit.task-complete` for the Phase 6 task.

If skipped: application is accessible at `http://<fixed_ip>`. Document as an open item and continue to Phase 7.

### Phase 7 — Validation

1. Invoke `smaqit.task-start` for the Phase 7 task.
2. Invoke the `smaqit.validation` agent with `mode` from Phase 0, `specification_mode: prevalidated`, and the Phase 2 Coverage specs as `spec_paths`.
3. If any spec is found inconsistent with the live system: amend in-place with an `amendment:` annotation.
4. **Gate:** All validation checks pass. User signs off.
5. Invoke `smaqit.task-complete` for the Phase 7 task, ensuring any amendments are captured under `Decisions made`.

### Phase 8 — Release

1. Confirm all phase tasks (1–7) are closed in `PLANNING.md`. If any remain open, resolve before continuing.
2. Run the amendment scan: `bash [SMAQIT_SKILLS_DIR]/smaqit.new-greenfield-project/scripts/check-amendments.sh specs/`. If the script reports matches, review each `amendment:` annotation against the `Blockers encountered` and `Follow-up identified` fields of the relevant phase task and confirm all are resolved or accepted. If no matches are found, skip this step entirely.
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

- Covers the single-app, two-environment path: dev VM or dev Namespace (Phase 4 local sweep) + production VM or production Namespace (Phase 5 CI/CD).
- Covers five `provisioning_mode` values (`provision`, `existing-owned`, `existing-shared`, `existing-unmanaged`, `existing-k3s`) as default slot profiles — see "Provisioning Mode and Deployment Slots" above Phase 4. A project's own fillers bound in `.smaqit/slots.yml` replace the defaults slot by slot; this skill's phase structure is unchanged by them. Does NOT cover two independent Terraform states both managing resources on the same VM; if a project genuinely needs its own state on a VM another project also has opinions about, that is out of scope here and needs its own design.
- Does NOT handle database schema migrations. The current project uses SQLite with append-only schema changes.
- Does NOT cover post-MVP feature cycles. Use `smaqit.feature-new` for iterative feature work after this skill completes.
- Phase 6 (domain/TLS) is conditional on domain purchase — a human action outside the system.

## Gotchas

- **Spec amendment protocol** — when an implementation phase must diverge from a spec (package mismatch, config change, structural adaptation): amend the spec in-place with an `amendment:` annotation describing what changed and why. Tactical divergences (versions, minor config) proceed autonomously. Structural divergences (data model, architecture) require operator approval before continuing. At `task-complete` time, the amendment is captured in `Decisions made`. Phase 8 runs `check-amendments.sh` to detect any open annotations; if none are found the review step is skipped.
- **Source path contract** — both deploy skills (`smaqit.infrastructure-deploy-rsync` and `smaqit.infrastructure-deploy-rsync-python-nextjs`) and generated CI/CD workflows assume `backend/` and `frontend/` as local source directories, and `__APP_DIR__` as the remote deploy path. The `smaqit.stack` specification agent (Phase 2) must declare these exact paths in the stack spec. If the declared stack matches no candidate the slot plan lists for `app-deploy`, the Family filler synthesis procedure under Phase 4 applies — it is not a "default and adapt" fallback.
- **Required-inherited-context for synthesized deploy skills** — any deploy skill synthesized under Phase 4's Family filler synthesis (via `smaqit.create-skill` or manual authoring) must inherit its **own family's** conventions rather than reinventing them, and never borrow the other family's:
  - **rsync family** (`smaqit.infrastructure-deploy-rsync*` — `provision`/`existing-owned`/`existing-shared`/`existing-unmanaged`): (1) the `__APP_DIR__` token convention for the remote deploy path; (2) delegating nginx vhost writing to the shared `smaqit.infrastructure-deploy-rsync/scripts/write-vhost.sh` — never a new `default_server`-vs-name-based implementation; (3) reusing `smaqit.infrastructure-hook-post-deploy-stamp`'s deploy-stamp pattern; (4) reusing `plan-guard.sh`/`ownership-guard.sh` for any Terraform-touching step, never a bare `terraform apply`.
  - **k3s family** (`smaqit.infrastructure-deploy-k3s-app*` — `existing-k3s`): (1) the `__APP_SLUG__` token convention — the Namespace is always `app-__APP_SLUG__`, never a free-form name; (2) delegating Ingress rendering to the shared `assets/ingress.yaml.template` / `ingress-default-tls.yaml.template` per the Environments row's Ingress Mode — never a hand-written per-app Ingress, and none at all for `platform-owned`; (3) delivering the deploy stamp as pod environment (`DEPLOY_SHA`/`DEPLOY_TIME`) set at apply time, never a file — there is no host filesystem this credential can write to; (4) reusing the bundled `namespace-guard.sh` pattern (refuse on any cluster-scoped permission or a Namespace mismatch) before any `manifest-lint.sh`/`kubectl apply`.
  Pass the selected family's four items explicitly as required context to the synthesis step — do not leave them to be rediscovered independently each time, and do not mix the two families' items.
- **Phases 4–6 never call infrastructure skills directly.** Every infrastructure step is a slot the Deployment agent executes from the resolved plan. Calling vault-loader, provision-cyso, a deploy skill, or domain-tls by hand bypasses the project's bindings and the recorded slot resolution; if a step seems missing, inspect `smaqit slots plan` rather than adding a direct call.
- **`existing-k3s` onboarding must be requested and merged before the deploy skill can run.** `smaqit.infrastructure-deploy-k3s-app`'s own Pre-conditions require an already-onboarded Namespace; nothing in this skill or that one ever performs onboarding directly. `smaqit.infrastructure-request-k3s-onboarding` opens a PR against the platform-owned infrastructure repo's registry file and gates on merge (Phase 4 for the test machine-slug, Phase 5 for prod) — a human-review gate, not a busy-poll loop. A "pending" result is not a failure: stop the current task and resume later once the platform team merges it.
- **Separate Terraform state keys** — dev and production must use different state keys (e.g. `dev/terraform.tfstate` vs `prod/terraform.tfstate`). Using the same key causes state conflicts and unintended VM replacement.
- **`GITHUB_TOKEN` reserved name** — enforced in Phase 5 via `smaqit.infrastructure-repo-config`. Never set an env var named `GITHUB_TOKEN` in any workflow to a PAT.
- **PR body sentinel** — Coding Agent must include `smaqit:deploy` as a line in any PR body to trigger post-merge deployment. Must be set at PR creation time, not via label.
- **Build-time Vite vars** — `VITE_DEMO_MODE` and any `VITE_*` vars must be passed to the frontend build step, not at runtime.
- **Floating IP (Cyso)** — use `fixed_ip` from Terraform outputs, not the floating IP. The floating IP does not route on Cyso's flat network.
- **CI/CD idempotency** — if dev VM was not torn down in Phase 4, the `terraform apply` in the Phase 5 CI/CD pipeline will show no changes and skip provisioning, deploying to the existing VM. Ensure the production state key points to a fresh state if a new VM is required.
- **Re-entry** — resume from the first incomplete phase. IaC generation in Phase 4 is idempotent; re-running it overwrites generated files but does not affect cloud resources.
- **Context collapse / phase re-read** — in long sessions the conversation is summarised by the model. Summaries capture phase names and outcomes but not the exact tool calls each phase requires. On resume, the agent operates from the summary's shorthand (e.g. "Phase 7: smoke tests PASS") rather than the SKILL.md instruction set, and substitutes a cheaper action it already ran (e.g. `smaqit.infrastructure-deploy-verify` curl checks from Phase 5) for the correct one (`/smaqit.validation` agent). Mitigation: at every phase boundary, re-read this SKILL.md (`read_file` the full steps for the upcoming phase) before executing any step. Do not rely on session memory or conversation summaries as a substitute for the canonical instruction set.

## Examples

**Input:** `assets/raw/code.txt` contains a React prototype. User invokes `/zero.to.prod`.
**Output:** All 9 phases completed. Application running at `https://<domain>/`, tagged as v1.0.0 on GitHub, all specs `status: deployed`, MVP task closed.

## Completion

- [ ] Phase 0: all phase tasks created, execution mode confirmed
- [ ] Phase 1: requirements extracted, ambiguities resolved
- [ ] Phase 2: all specs and minimal same-layer PlantUML/PNG design pairs drafted, visually reviewed, validated, and approved
- [ ] Phase 3: implementation complete, specs set to `implemented`
- [ ] Phase 4: dev slot plan resolved and executed by the Deployment agent (target provisioned or onboarded, deployed, and verified); IaC artifacts committed
- [ ] Phase 5: `ci-push` slot plan executed; CI/CD run correlated to the pushed SHA succeeded; production verified
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
| `smaqit slots plan` reports a `SLOTS-*` failure | Stop before invoking the Deployment agent. Fix the binding in `.smaqit/slots.yml` or the bound skill's `metadata.smaqit-slots`. |
| Platform onboarding reports pending | Stop the current phase task; re-invoke later. The onboarding filler re-checks the same PR. |
| Phase 6 skipped (no domain) | Document as open item. Continue to Phase 7. |
