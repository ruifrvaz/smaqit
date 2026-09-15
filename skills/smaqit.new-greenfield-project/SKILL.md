---
name: smaqit.new-greenfield-project
description: >-
  Use when orchestrating the complete SDLC for a new project — from raw project assets to a
  running production application accessible via browser. Covers requirements extraction,
  specification (business, functional, stack, infrastructure, coverage), task creation,
  development, IaC generation + dev environment sweep (local provisioning + deploy + verify),
  CI/CD production deployment, optional domain/TLS, validation, and tagged release. Re-entrant:
  use the pre-condition checklist to resume at any phase. Also use when the user says "take this
  from zero to prod", "run the full smaqit pipeline", "deploy a new project end-to-end", or when
  starting implementation on a freshly initialized repository.
metadata:
  version: "1.8.0"
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

Applicability below depends on `provisioning_mode` (resolved by `smaqit.input-deployment` — see "Provisioning Mode" under Phase 4/5). Items marked `[provision/existing-owned]` apply when this project provisions or owns the target VM's Terraform state; items marked `[existing-shared/existing-unmanaged]` apply when targeting a VM this project does not manage via its own Terraform state — either because a different project owns it (`existing-shared`, co-hosted) or because nobody's Terraform manages it at all (`existing-unmanaged`, dedicated).

- [ ] `[provision/existing-owned]` Cloud account available (Cyso or equivalent)
- [ ] `[provision/existing-owned]` Application credential created in cloud portal; loaded into local Vault at `secret/<project-slug>/cyso`
- [ ] `[provision/existing-owned]` Object Storage state bucket created (with separate state keys for dev and prod); S3 keys loaded into Vault at `secret/<project-slug>/tfstate`
- [ ] `[all modes]` This app has its own distinct SSH keypair at `secret/apps/<app-slug>/ssh`, bootstrapped against the target machine's `secret/machines/<machine-slug>/base-ssh` credential via `bootstrap-app-to-machine.sh` (see `smaqit.infrastructure-vault-loader`) — for `existing-shared`, the target machine must already be registered (its `base-ssh` already exists); for `existing-unmanaged`, the machine is typically registering for the *first* time, so this step exercises `bootstrap-app-to-machine.sh`'s fresh-registration branch instead
- [ ] `[all modes]` Fine-grained PAT with `variables:write` loaded into Vault at `secret/<project-slug>/github`
- [ ] `[existing-shared/existing-unmanaged]` Target VM's fixed IP known; will be set via `gh variable set VM_HOST` rather than read from a Terraform output
- [ ] `[all modes]` Local Vault initialised and running on 127.0.0.1:8200 (`smaqit.infrastructure-vault-loader` one-time setup complete)

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
2. Invoke `/smaqit.business` agent.
3. Invoke `/smaqit.functional` agent.
4. Invoke `/smaqit.stack` agent.
5. Invoke `/smaqit.infrastructure` agent.
6. Invoke `/smaqit.coverage` agent.
7. Invoke `smaqit.design-validate` for every generated pair. Confirm `smaqit design validate` passes and the active agent has opened each PNG; PlantUML-source reading is not a visual-review fallback.
8. **Gate:** All specs have `status: draft`, acceptance criteria written, and current visually approved same-layer design pairs. User reviews and approves the full spec/design set.
9. Invoke `smaqit.task-complete` for the Phase 2 task.

### Phase 3 — Development

1. Invoke `smaqit.task-start` for the Phase 3 task.
2. Invoke `/smaqit.development` agent to implement all specs with `status: draft`.
3. If any spec requires amendment to proceed: amend the spec in-place with an `amendment:` annotation and continue. Structural divergences that change architecture must be paused for operator approval before continuing.
4. **Gate:** Build passes (backend and frontend). All MVP acceptance criteria met. Development agent sets specs and their linked designs to `status: implemented`, respecting the least-advanced linked-spec rule, and reruns `smaqit design validate`.
5. Invoke `smaqit.task-complete` for the Phase 3 task, ensuring any amendments are captured under `Decisions made`.

### Provisioning Mode (applies to Phases 4 and 5)

Before Phase 4 begins, `smaqit.input-deployment` resolves `provisioning_mode` — one of:

- **`provision`** (default) — this project provisions its own new VM via Terraform.
- **`existing-owned`** — redeploying to a VM this project's own Terraform state already manages.
- **`existing-shared`** — targeting a VM a *different* project owns and manages via its own Terraform state (co-hosting).
- **`existing-unmanaged`** — targeting a VM dedicated to this project (not co-hosted) but never managed by any Terraform state — provisioned out-of-band and staying that way by design.
- **`existing-k3s`** — targeting an already-onboarded, Namespace-scoped Kubernetes cluster (k3s or equivalent) that hands the app a scoped kubeconfig and nothing else: no SSH, no host filesystem, no nginx, no Terraform on the app's side. The target is platform-owned, like `existing-shared`, but the deploy mechanism is entirely different — `kubectl apply` into the app's own Namespace via the `smaqit.infrastructure-deploy-k3s-app` family, never rsync/SSH/nginx.

Phase 4 and Phase 5 steps below are written for `provision`. Where a step differs under `existing-owned`, `existing-shared`, `existing-unmanaged`, or `existing-k3s`, that difference is called out immediately under the step as `→ existing-owned:` / `→ existing-shared:` / `→ existing-unmanaged:` / `→ existing-k3s:`. Steps with no callout are identical across all five modes. `existing-unmanaged` mirrors `existing-shared` almost everywhere (no Terraform, `deploy-only` CI/CD, restricted Vault/repo-config) except machine registration (typically fresh rather than already-registered) and the nginx vhost step (no co-hosting to guard against, so no callout at all — see Step 6 below). `existing-k3s` shares the "no Terraform, no `VM_HOST`" shape with `existing-shared`/`existing-unmanaged` but replaces every VM/SSH/nginx-specific step outright with its own kubeconfig-based equivalent — it is not a variant of the rsync family, so where the rsync-family callouts above don't apply, `existing-k3s` gets its own.

### Phase 4 — Dev Environment Sweep

Validates the full infrastructure and deployment approach on a dedicated dev VM before committing to CI/CD.

1. Invoke `smaqit.task-start` for the Phase 4 task.
2. Invoke `smaqit.infrastructure-vault-loader`. Confirm Vault is running, unsealed, and all `secret/<project-slug>/*` paths are populated. Do not proceed until confirmed.
   → **`existing-shared`:** only `secret/apps/<app-slug>/github` is loaded here; `cyso`/`tfstate` are never prompted for — they live at `secret/machines/<machine-slug>/*`, owned by whichever project provisioned the machine. Then run `bootstrap-app-to-machine.sh <app-slug> <machine-slug>` against the target machine's already-registered `base-ssh` credential to populate `secret/apps/<app-slug>/ssh`.
   → **`existing-unmanaged`:** same as `existing-shared` — only `github` is loaded here. The difference is what `bootstrap-app-to-machine.sh` does next: the target machine is typically registering for the *first* time (no prior `base-ssh`), so this exercises the script's fresh-registration branch — it prompts for host/provider/owner_project (use this project's own slug for `owner_project`, since there is no other project involved), generates a keypair, and defers installing the public key to the operator rather than assuming one is already trusted.
   → **`existing-k3s`:** no SSH, no VM machine bootstrap. Only `secret/apps/<app-slug>/github` and the test environment's registered machine-slug's `secret/apps/<app-slug>/<machine-slug>/kubeconfig` (for this dev sweep) are loaded, via `smaqit.infrastructure-vault-loader`'s `existing-k3s` branch (`MACHINE_SLUG=<test-machine-slug>`) — an out-of-band paste of the platform-issued kubeconfig, never a cluster call. `cyso`/`tfstate`/`ssh` are never prompted for, and there is no `secret/machines/<machine-slug>/*` counterpart for the k3s machine-slug either.
3. Invoke `/smaqit.deployment` agent with context: generate all IaC artifacts — Terraform files in `deployment/terraform/` and GitHub Actions workflow files in `.github/workflows/` using `smaqit.infrastructure-cicd-generate` patterns as reference; Terraform state key `dev/terraform.tfstate`; do not trigger deployment execution.
   → **`existing-shared`/`existing-unmanaged`:** invoke `smaqit.infrastructure-cicd-generate` in `deploy-only` mode. No Terraform files are generated for this project; `provision.yml` is not generated and `deploy.yml` has a single `deploy` job. Identical for both modes — the generated workflow doesn't need to know *why* there's no Terraform, only that there isn't any.
   → **`existing-k3s`:** invoke `smaqit.infrastructure-cicd-generate` in `k3s` mode. No Terraform files are generated; `provision.yml` is not generated; `deploy.yml` has a `build` job (pushes the app's own `Dockerfile` to GHCR, tagged by commit SHA) followed by a kubectl-driven `deploy` job (guard → lint → apply → rollout → verify) reading `KUBECONFIG` from a GitHub Environment secret. No SSH, no nginx, no `VM_HOST` anywhere in the generated output. The `Dockerfile` itself is authored by Step 6 below's matched/synthesized image-build skill, not by this generation step — see that skill's own Gotchas, and `smaqit.infrastructure-cicd-generate`'s own Gotchas, for two preconditions a first-time `k3s`-mode project will hit: `manifest-lint.sh` fails on a missing `ingress.yaml` until it's rendered once (Step 6 below), and a freshly generated `deploy.yml` cannot be `workflow_dispatch`'d until it exists on the default branch.
4. Invoke `smaqit.infrastructure-provision-cyso` with dev environment variables. Note the `fixed_ip` output.
   → **`existing-owned`:** `terraform apply` is expected to no-op (gated by `plan-guard.sh`) — this is correct, idempotent behavior, not a failure. The existing `fixed_ip` remains unchanged.
   → **`existing-shared`/`existing-unmanaged`:** **skip this step entirely.** This project never provisions Terraform for a VM it doesn't own or doesn't manage. Set the target IP via `gh variable set VM_HOST` instead (see Phase 5).
   → **`existing-k3s`:** **skip this step entirely.** There is no VM to provision — the target is a Namespace inside an already-onboarded cluster, not a host. No `VM_HOST` is ever set for this mode; the deploy target is the kubeconfig's own embedded cluster endpoint.
5. Invoke `smaqit.infrastructure-vm-bootstrap` with the dev VM `fixed_ip`.
   → **`existing-shared`/`existing-unmanaged`:** use the manually-set target IP in place of a Terraform `fixed_ip` output.
   → **`existing-k3s`:** **skip this step entirely.** There is no VM to bootstrap — no data volume, no `ubuntu` user, no nginx, no Docker group. The Namespace's own onboarding (RBAC, Pod Security label, NetworkPolicy, quota) already happened on the platform side before this project ever runs.
6. Resolve and invoke the deploy skill for the project's stack. The `provisioning_mode` selects
   the deploy skill **family** first; the declared stack is matched *within* that family. The
   rsync family's four required-inherited-context items are VM-specific and must never leak into
   the k3s family, and vice versa.
   - **Precondition:** read the declared stack from `specs/stack/platform-stack.md` — the
     authoritative source. Do not re-derive the stack independently from the filesystem.
   - **Family selection:**
     - `provision`/`existing-owned`/`existing-shared`/`existing-unmanaged` → the
       `smaqit.infrastructure-deploy-rsync*` family.
     - `existing-k3s` → the `smaqit.infrastructure-deploy-k3s-app*` family.
   - **Judgment:** compare the declared stack against the currently-installed skills in the
     selected family (by description/metadata) and identify whether any matches.
   - **Matched** → invoke it. This is a lookup against whatever skills currently exist in the
     selected family (e.g. `smaqit.infrastructure-deploy-rsync` for Node.js + Vite/React,
     `smaqit.infrastructure-deploy-rsync-python-nextjs` for Python/FastAPI + Next.js,
     `smaqit.infrastructure-deploy-k3s-app` for a plain Kubernetes/k3s target) — it scales to
     however many skills exist in either family without this step needing to be edited every
     time one is added.
   - **No match** →
     a. Report what was checked: the declared stack, which skills in the selected family were
        compared against it, and why none matched.
     b. Check whether `smaqit.create-skill` is available: it and its `smaqit.L2` compiler
        dependency are installed globally by smaqit-adk (not project-local), so check that both
        are present among this environment's available skills/agents.
     c. **If available (primary path):** invoke `smaqit.create-skill` with a name derived from the
        declared stack (e.g. `smaqit.infrastructure-deploy-rsync-<stack-slug>` for the rsync
        family, `smaqit.infrastructure-deploy-k3s-app-<stack-slug>` for the k3s family),
        explicitly pointed at the selected family's own existing skill(s) as reference exemplars
        (rsync family: `smaqit.infrastructure-deploy-rsync` and
        `smaqit.infrastructure-deploy-rsync-python-nextjs`; k3s family:
        `smaqit.infrastructure-deploy-k3s-app`), and that family's own four
        required-inherited-context items (see Gotchas) as explicit input — never the other
        family's.
     d. **If unavailable (manual fallback):** author the `SKILL.md` by hand, following the
        selected family's existing skill's structural shape (Pre-conditions, numbered Steps,
        Output, Scope, Gotchas, Completion, Failure Handling), applying that family's own four
        required-inherited-context items manually.
     e. **Human checkpoint:** present the synthesized skill to the user before invoking it.
        Skippable only in Autonomous mode with the user's prior, explicit sign-off for that mode.
     f. Invoke the synthesized skill to perform the actual deploy.
   - Tag any synthesized skill with provenance metadata (e.g. `synthesized`,
     `synthesized-for-project`, `synthesized-date`) marking it a candidate for a future
     reconciliation into canonical `smaqit` once proven by real use.
   → **`existing-shared`:** if this is not the first site on the VM, the nginx vhost must be name-based only — never `default_server`. Every deploy skill in the family — matched or synthesized — calls the same `smaqit.infrastructure-deploy-rsync/scripts/write-vhost.sh` to enforce this; it is not stack-specific.
   → **`existing-unmanaged`:** no callout applies here — this is a co-hosting concern, not a Terraform-management one, and `existing-unmanaged` VMs are dedicated (not co-hosted) by definition. `write-vhost.sh`'s own live inspection of the VM's existing nginx sites will correctly resolve to `default_server`, same as `provision`/`existing-owned`. Do not apply the `existing-shared` callout above to this mode.
   → **`existing-k3s`:** the deploy skill's own Pre-conditions assume the app's `Dockerfile` and
     Deployment/Service manifests already exist locally — it never authors either. Author them
     first via a matched or synthesized image-build skill (e.g.
     `smaqit.infrastructure-image-build-push-static` for a static site with no build step of its
     own), following the same match/synthesize judgment as the deploy skill itself. Then, before
     invoking the deploy skill, invoke `smaqit.infrastructure-request-k3s-onboarding` for the test
     environment's registered machine-slug and gate on its PR merging — the deploy skill's own
     Pre-conditions also require the Namespace to already be onboarded, and this is the step that
     requests it. If the gate reports "pending," stop this task here (re-invoke later — it
     re-checks the same PR rather than re-creating it) rather than proceeding to the deploy skill
     against a Namespace that doesn't exist yet. Once merged, invoke the matched or synthesized
     `smaqit.infrastructure-deploy-k3s-app*` skill against the `test` Namespace/kubeconfig. There
     is no nginx vhost step at all — routing is the shared Ingress template (rendered with
     `__APP_SLUG__`, the declared ingress class, and `ClusterIssuer`), not a VM vhost. The skill's
     own `namespace-guard.sh` and `manifest-lint.sh` run before any `kubectl apply`, and the skill
     invokes `smaqit.infrastructure-deploy-verify` itself as its final step — Step 7 below still
     runs, but only re-confirms what the deploy skill already checked.
7. Invoke `smaqit.infrastructure-deploy-verify` against the dev VM. If any check fails, stop and fix before continuing.
   → **`existing-k3s`:** target is `https://<test-host>` (the Ingress host from the Infrastructure spec), not a VM IP — `smaqit.infrastructure-deploy-k3s-app` already invoked this as its own last step; this confirms the same result rather than re-deriving it independently.
8. If any infrastructure or stack spec required amendment to proceed: amend in-place with an `amendment:` annotation.
9. Commit all generated IaC artifacts: `git add deployment/ .github/workflows/ && git commit -m "ci: add infrastructure and CI/CD workflows"`.
   → **`existing-shared`/`existing-unmanaged`:** no `deployment/terraform/` directory exists to commit; commit `.github/workflows/` only.
   → **`existing-k3s`:** no `deployment/terraform/` directory exists to commit; commit `.github/workflows/` (and any Kubernetes manifests under `deployment/`) only.
10. **Gate:** All `deploy-verify` checks PASS on dev VM. IaC artifacts committed.
11. Invoke `smaqit.task-complete` for the Phase 4 task, ensuring any amendments are captured under `Decisions made`.
12. *(Optional)* Tear down dev VM: run `terraform destroy` using dev state to avoid ongoing cloud costs.
   → **`existing-shared`:** not applicable — there is no Terraform state for this project to destroy, and the VM is owned by another project regardless.
   → **`existing-unmanaged`:** not applicable — there is no Terraform state for this project (or anyone else's) to destroy for this VM.
   → **`existing-k3s`:** not applicable — there is no VM or Terraform state at all. If a throwaway slug was used for this sweep, offboard it through the platform's own offboarding path instead.

### Phase 5 — Production Deployment via CI/CD

Uses IaC artifacts from Phase 4. Configures production secrets, pushes to main, and monitors the triggered pipeline.

1. Invoke `smaqit.task-start` for the Phase 5 task.
2. Invoke `smaqit.infrastructure-vault-loader`. Confirm Vault is running and all credential paths are populated.
   → **`existing-shared`/`existing-unmanaged`:** only `ssh` and `github` are required, as in Phase 4 step 2.
   → **`existing-k3s`:** before this step, invoke `smaqit.infrastructure-request-k3s-onboarding` for the prod environment's registered machine-slug and gate on its PR merging — the platform team only hands off the prod kubeconfig once that PR is approved, and this step's Vault paste has nothing to paste until then. If the gate reports "pending," stop this task here (re-invoke later — it re-checks the same PR rather than re-creating it). Once merged and the kubeconfig received, `github` and the prod environment's registered machine-slug's `secret/apps/<app-slug>/<machine-slug>/kubeconfig` are required, populated the same way the test machine-slug's was in Phase 4 step 2. No `ssh`, `cyso`, or `tfstate` at any point.
3. Invoke `smaqit.infrastructure-repo-config` to sync all production secrets from Vault to GitHub Secrets (cloud credentials, Terraform backend, SSH key).
   → **`existing-shared`/`existing-unmanaged`:** `repo-config` runs in restricted mode — it detects the absent `tfstate`/`cyso` Vault paths and skips syncing those secrets cleanly (not a hard failure), syncing only `ssh` + `github`-derived secrets. Additionally run `gh variable set VM_HOST --body <target-vm-ip>` — same variable as the default path, just set manually since this project has no Terraform output to derive `VM_HOST` from (either because another project's Terraform owns it, or because nobody's does).
   → **`existing-k3s`:** `repo-config` writes the `prod` kubeconfig as a `KUBECONFIG` secret on the `prod` GitHub Environment (not a plain repository secret — see `smaqit.infrastructure-repo-config`); if the spec's `Container Registry` Constraints row declares a private registry, it also syncs `REGISTRY_USERNAME`/`REGISTRY_TOKEN` from the shared `secret/organizations/<org-slug>/github-package-read` (see `smaqit.infrastructure-vault-loader`'s path convention). No `VM_HOST` is ever set, and no `tfstate`/`cyso`/`ssh`-derived secret is synced.
4. Push to main: `git push origin main`. The `deploy.yml` workflow triggers automatically (provision job → deploy job).
   → **`existing-shared`/`existing-unmanaged`:** `deploy.yml` has only a `deploy` job (no `provision` job) — generated that way by `smaqit.infrastructure-cicd-generate`'s `deploy-only` mode in Phase 4.
   → **`existing-k3s`:** `deploy.yml` has a `build` job (pushes the app's image to GHCR) followed by a kubectl-driven `deploy` job (guard → lint → apply → rollout → `Certificate` → verify) — generated that way by `smaqit.infrastructure-cicd-generate`'s `k3s` mode in Phase 4. No `provision` job, no SSH, no Terraform reference anywhere in the generated workflow.
5. Monitor the pipeline: `gh run watch` — wait for the workflow run to complete.
6. Invoke `smaqit.infrastructure-deploy-verify` against the production VM. If any check fails, stop and report.
   → **`existing-k3s`:** target is `https://<prod-host>` (the Ingress host from the Infrastructure spec) — the `deploy.yml` job already invoked this internally; this confirms the same result.
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
2. Invoke `/smaqit.validation` agent.
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
- Covers five `provisioning_mode` values (`provision`, `existing-owned`, `existing-shared`, `existing-unmanaged`, `existing-k3s`) — see "Provisioning Mode" above Phase 4. Does NOT cover two independent Terraform states both managing resources on the same VM; if a project genuinely needs its own state on a VM another project also has opinions about, that is out of scope here and needs its own design.
- Does NOT handle database schema migrations. The current project uses SQLite with append-only schema changes.
- Does NOT cover post-MVP feature cycles. Use `smaqit.feature-new` for iterative feature work after this skill completes.
- Phase 6 (domain/TLS) is conditional on domain purchase — a human action outside the system.

## Gotchas

- **Spec amendment protocol** — when an implementation phase must diverge from a spec (package mismatch, config change, structural adaptation): amend the spec in-place with an `amendment:` annotation describing what changed and why. Tactical divergences (versions, minor config) proceed autonomously. Structural divergences (data model, architecture) require operator approval before continuing. At `task-complete` time, the amendment is captured in `Decisions made`. Phase 8 runs `check-amendments.sh` to detect any open annotations; if none are found the review step is skipped.
- **Source path contract** — both deploy skills (`smaqit.infrastructure-deploy-rsync` and `smaqit.infrastructure-deploy-rsync-python-nextjs`) and generated CI/CD workflows assume `backend/` and `frontend/` as local source directories, and `__APP_DIR__` as the remote deploy path. The `/smaqit.stack` agent (Phase 2) must declare these exact paths in the stack spec. If the declared stack matches neither existing deploy skill, Phase 4 Step 6's judgment/synthesis procedure applies — it is not a "default and adapt" fallback; see Step 6 above for the full precondition → judgment → synthesis sequence.
- **Required-inherited-context for synthesized deploy skills** — any deploy skill synthesized under Phase 4 Step 6's no-match path (via `smaqit.create-skill` or manual authoring) must inherit its **own family's** conventions rather than reinventing them, and never borrow the other family's:
  - **rsync family** (`smaqit.infrastructure-deploy-rsync*` — `provision`/`existing-owned`/`existing-shared`/`existing-unmanaged`): (1) the `__APP_DIR__` token convention for the remote deploy path; (2) delegating nginx vhost writing to the shared `smaqit.infrastructure-deploy-rsync/scripts/write-vhost.sh` — never a new `default_server`-vs-name-based implementation; (3) reusing `smaqit.infrastructure-hook-post-deploy-stamp`'s deploy-stamp pattern; (4) reusing `plan-guard.sh`/`ownership-guard.sh` for any Terraform-touching step, never a bare `terraform apply`.
  - **k3s family** (`smaqit.infrastructure-deploy-k3s-app*` — `existing-k3s`): (1) the `__APP_SLUG__` token convention — the Namespace is always `app-__APP_SLUG__`, never a free-form name; (2) delegating Ingress rendering to the shared `assets/ingress.yaml.template` — never a hand-written per-app Ingress; (3) delivering the deploy stamp as pod environment (`DEPLOY_SHA`/`DEPLOY_TIME`) set at apply time, never a file — there is no host filesystem this credential can write to; (4) reusing the bundled `namespace-guard.sh` pattern (refuse on any cluster-scoped permission or a Namespace mismatch) before any `manifest-lint.sh`/`kubectl apply`.
  Pass the selected family's four items explicitly as required context to the synthesis step — do not leave them to be rediscovered independently each time, and do not mix the two families' items.
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
