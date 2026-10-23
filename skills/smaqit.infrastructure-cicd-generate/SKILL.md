---
name: smaqit.infrastructure-cicd-generate
description: 'Use when generating or updating GitHub Actions CI/CD workflows: the VM set (deploy.yml, provision.yml, post-merge-deploy.yml) for Node.js + React via Terraform and Docker Compose, or the k3s two-job deploy.yml + post-merge-deploy.yml, or when an operator invokes /cicd.generate. Inside smaqit.project-new the deployment agent generates them itself. Never generates copilot-setup-steps.yml.'
metadata:
  version: "3.3.0"
  smaqit-slots: "ci-generate=generate"
---

# CI/CD Workflow Generator

Generates workflows by **copying real, pre-validated template files** from `assets/` and
substituting one token (`k3s` mode: rendering one template with `scripts/render-k3s-deploy.sh`
from the Infrastructure spec's Environments table) — it does not synthesize YAML from a prose
description. This is
deliberate: every Gotcha below used to be something an agent had to remember to reproduce
correctly on every generation; now it's encoded once, in a template that's already been
validated with `actionlint`/`yamllint`/`shellcheck` (Task 085).

## Generation Mode

- **`full`** (default) — this project provisions and manages its own VM's Terraform state.
  Generates all 3 workflows with a `provision` job.
- **`deploy-only`** — `provisioning_mode: existing-shared` or `existing-unmanaged` (see
  `smaqit.project-new` Phase 4/5): no Terraform step is generated at all, and
  `provision.yml` is not generated, whether because this project deploys onto a VM a *different*
  project owns and provisions (`existing-shared`, co-hosting), or because the target VM is
  dedicated to this project but nobody's Terraform manages it (`existing-unmanaged`, provisioned
  out-of-band). Both modes produce the identical generated output — the underlying reason there's
  no Terraform step doesn't change what gets written.
- **`k3s`** — `provisioning_mode: existing-k3s` (see `smaqit.project-new` Phase 4/5):
  the target is a Namespace-scoped Kubernetes/k3s cluster, not a VM at all. No Terraform, no SSH,
  no nginx, no `VM_HOST` — `deploy.yml` is dispatch-only and has a `build` job (builds the
  repository's own `Dockerfile`, pushes to GHCR tagged by commit SHA with the repository name
  lowercased) followed by a `deploy` job. One `deploy.yml` serves every environment: it carries
  each machine's Environments row (`K3S_TARGETS`), resolves the dispatched machine's row, joins the
  mesh first for a `private-mesh` machine, and then runs the same guard → lint → apply → rollout →
  `Certificate` (own-tls only) → verify sequence `smaqit.infrastructure-deploy-k3s-app` documents.
  Verification runs inside the cluster (a probe Job) for a `private-mesh` or `platform-owned`
  machine and over HTTPS otherwise. `provision.yml` is not generated; `post-merge-deploy.yml` is
  reused unchanged and is the only automatic trigger. This workflow is the only way an
  existing-k3s app deploys: nothing deploys from the operator's machine, so this mode also renders
  the Kubernetes manifests the workflow applies (Step 6).

GitHub Actions is the CI/CD system these templates target. The Deployment agent drives a generated
workflow through one contract, the ci-monitor seam: trigger the run (push, PR merge, or dispatch),
select it by commit, wait for it to finish, and read its verify evidence. `gh` implements that
contract today; another CI system would need its own templates and its own implementation of the
same four operations.

## Steps

1. **Read existing workflows** from `.github/workflows/`. If workflows exist and are substantially complete, offer to diff rather than overwrite.

2. **Resolve configuration values** from Infrastructure specs and `AGENTS.md` (or legacy platform-specific instructions). Only one substitution is needed per mode — the templates otherwise use fixed, documented secret/variable names (see the table below and `smaqit.infrastructure-repo-config`):

   | Token | Resolves to | Default if not found | Applies to |
   |-------|-------------|------------------------|------------|
   | `__APP_DIR__` | The remote application directory on the VM (e.g. `/opt/myapp`) | `/opt/<project-slug>` — derive `<project-slug>` from `AGENTS.md` or legacy platform instructions, same convention as `smaqit.infrastructure-vault-loader` | `full`, `deploy-only` |
   | `__APP_SLUG__` | The app slug — the Namespace is always `app-__APP_SLUG__` | `<project-slug>`, same derivation as above | `k3s` |

   `k3s` mode takes everything else from the Infrastructure spec's `## Environments` table — machine
   slugs (also the GitHub Environment names), hosts, reachability, Ingress mode, Ingress class,
   ClusterIssuer, health path, and the mesh-join action — and its `Container Registry` and
   `Deployer Exec` Constraints rows, through `smaqit spec environments --json`,
   which refuses an incomplete table with `SLOTS-SPEC-*` lines. Never fill a missing value in by hand;
   the fix belongs in the spec.

   All other references (`VM_HOST`, `VM_SSH_KEY`, `TF_BACKEND_ACCESS_KEY`, `OS_APPLICATION_CREDENTIAL_ID`, `GH_TERRAFORM_TOKEN`, `VITE_DEMO_MODE`, `KUBECONFIG`, etc.) are baked into the templates verbatim — do not re-derive or rename them per project. Every secret reaches a step through `env:` (or an action's `with:`), never `${{ secrets.* }}` inside a `run:` script. A `k3s` workflow reads each machine's host from its rendered `K3S_TARGETS` entry, not from an `APP_HOST` variable.

3. **Select and copy the template variant** for the active generation mode:

   | File | `full` | `deploy-only` | `k3s` |
   |------|--------|----------------|-------|
   | `deploy.yml` | `assets/deploy.yml.full.template` | `assets/deploy.yml.deploy-only.template` | `assets/deploy.yml.k3s.template`, rendered by `scripts/render-k3s-deploy.sh` (Step 4) |
   | `provision.yml` | `assets/provision.yml.template` | not generated | not generated |
   | `post-merge-deploy.yml` | `assets/post-merge-deploy.yml.template` | `assets/post-merge-deploy.yml.template` (identical in all modes) | `assets/post-merge-deploy.yml.template` (identical in all modes) |

4. **Substitute the mode's tokens** in the copied `deploy.yml`: `__APP_DIR__` for `full`/`deploy-only`,
   with the resolved value from Step 2 (`provision.yml` and `post-merge-deploy.yml` don't reference
   it). For `k3s`, render instead of substituting — never edit the template's tokens by hand:

   ```bash
   ENVIRONMENTS="$(mktemp)"
   smaqit spec environments --json > "$ENVIRONMENTS"
   bash <this-skill>/scripts/render-k3s-deploy.sh "$ENVIRONMENTS" <app-slug> > .github/workflows/deploy.yml
   ```

   The dispatch options list every environment's machine slug; the default (the machine
   `post-merge-deploy.yml` deploys) is the `prod` environment's, or the only environment's — pass
   `--default-env <environment>` otherwise. The mesh-join step is kept, with the spec's pinned
   action and input names, only when some environment is `private-mesh`. The `Create
   imagePullSecret` step is kept, with the registry host, only when the spec's `Container Registry`
   is private. The spec's `Deployer Exec` (default `denied`) becomes the `deploy` job's
   `DEPLOYER_EXEC`. The renderer refuses (exit 1, reason on stderr) two environments sharing a
   machine slug, an unknown default, a `private-mesh` environment without the spec's mesh access
   rows, a private registry without a host, or an unknown exec policy; it writes nothing on refusal.

5. **Vendor the guard scripts a GitHub-hosted runner needs on disk** — the runner has no access
   to the smaqit skill directories, so any script the generated workflow shells out to must be
   checked into the target repo, not referenced at its smaqit-local path:
   - **`full` only** (Terraform-touching):
     - `smaqit.infrastructure-provision-cyso/scripts/plan-guard.sh` → `deployment/scripts/plan-guard.sh`
     - `smaqit.infrastructure-provision-cyso/scripts/ownership-guard.sh` → `deployment/scripts/ownership-guard.sh`
   - **`deploy-only`:** nothing to vendor — no Terraform, and the rsync/SSH steps are inline `run:`
     commands in the template, not calls to a separate script.
   - **`k3s` only:**
     - `smaqit.infrastructure-deploy-k3s-app/scripts/namespace-guard.sh` → `deployment/k3s/scripts/namespace-guard.sh`
     - `smaqit.infrastructure-deploy-k3s-app/scripts/manifest-lint.sh` → `deployment/k3s/scripts/manifest-lint.sh`
     - `smaqit.infrastructure-deploy-k3s-app/scripts/exec-refused.sh` → `deployment/k3s/scripts/exec-refused.sh`

   All vendored scripts are **generated output, overwritten on every run** — the source skill
   directory's copies are canonical. Never hand-edit a vendored copy expecting it to survive the
   next generation.

6. **`k3s` only — render the manifests the environments need** from
   `smaqit.infrastructure-deploy-k3s-app`'s shared templates (its Shared Templates section), into
   `deployment/k3s/`. `deployment.yaml` and `service.yaml` must already exist there (authored by the
   `artifact-build` slot, which runs first); the probe takes its user from them:

   | Rendered when some environment is… | Template (`smaqit.infrastructure-deploy-k3s-app/assets/`) | Written as |
   |---|---|---|
   | `own-tls` | `ingress.yaml.template` | `deployment/k3s/ingress.yaml` |
   | `platform-default-tls` | `ingress-default-tls.yaml.template` | `deployment/k3s/ingress-default-tls.yaml` |
   | `private-mesh` or `platform-owned` | `probe-job.yaml.template` | `deployment/k3s/probe-job.yaml` |

   ```bash
   K3S=<smaqit-skills-dir>/smaqit.infrastructure-deploy-k3s-app/assets
   sed -e "s/__APP_SLUG__/${APP_SLUG}/g" "$K3S/ingress.yaml.template" > deployment/k3s/ingress.yaml
   sed -e "s/__APP_SLUG__/${APP_SLUG}/g" -e "s/__RUN_AS_USER__/<runAsUser from deployment.yaml>/" \
     -e "s#__PROBE_FETCH__#<fetch command for the spec's Probe Client>#" \
     "$K3S/probe-job.yaml.template" > deployment/k3s/probe-job.yaml
   ```
   Substitute only `__APP_SLUG__`, `__RUN_AS_USER__`, and `__PROBE_FETCH__`; leave the
   per-environment tokens (`__HOST__`, `__INGRESS_CLASS__`, `__CLUSTER_ISSUER__`, `__IMAGE__`,
   `__HEALTH_PATH__`) in place — the workflow stamps the dispatched machine's values before linting.
   The fetch command for each `Probe Client` is in deploy-k3s-app's table. Render nothing for
   `platform-owned` beyond the probe: the platform owns that Ingress. Like the vendored scripts,
   the rendered manifests are generated output, overwritten on every run.

7. **Write files to `.github/workflows/`** — 3 workflow files + 2 vendored scripts (`full`); 2
   workflow files + 0 vendored scripts for `deploy-only` (`deploy.yml` + `post-merge-deploy.yml`,
   no `provision.yml`, no Terraform to guard); 2 workflow files + 3 vendored scripts + the rendered
   manifests for `k3s` (`deployment/k3s/scripts/namespace-guard.sh`, `manifest-lint.sh`, and
   `exec-refused.sh` — `deploy.yml`'s `deploy` job is kubectl-driven, not rsync/SSH-driven, but
   still needs its guard/lint/exec-check scripts on disk — and Step 6's Ingress and probe files).

8. **Report** the list of files written with a brief description of each, including which generation mode was used.

## Output

- **`full`:** `.github/workflows/deploy.yml` (provision + deploy sequential jobs), `provision.yml` (PR-only terraform plan with PR comment), `post-merge-deploy.yml` (body sentinel trigger for deploy.yml); `deployment/scripts/plan-guard.sh` and `deployment/scripts/ownership-guard.sh` vendored
- **`deploy-only`:** `.github/workflows/deploy.yml` (single deploy job, no Terraform), `post-merge-deploy.yml` (body sentinel trigger for deploy.yml) — no `provision.yml`, nothing vendored
- **`k3s`:** `.github/workflows/deploy.yml` (dispatch-only; `build` job pushing to GHCR, then a kubectl-driven `deploy` job, one at a time per machine: resolve target → mesh join (private-mesh) → kubeconfig → guard (including exec) → imagePullSecret (private registry) → lint → apply → rollout → exec refused → `Certificate` (own-tls) → revision → in-cluster or HTTPS verify → kubeconfig removed, reading `KUBECONFIG` from a GitHub Environment secret), `post-merge-deploy.yml` (body sentinel trigger for deploy.yml); `deployment/k3s/scripts/namespace-guard.sh`, `manifest-lint.sh`, and `exec-refused.sh` vendored; `deployment/k3s/ingress.yaml`, `ingress-default-tls.yaml`, and `probe-job.yaml` rendered as the environments need — no `provision.yml`, no SSH or Terraform reference anywhere in the generated output

## Scope

- Does NOT set up GitHub Secrets or Variables — use `smaqit.infrastructure-repo-config` for that.
- Does NOT provision infrastructure — Terraform runs inside the generated workflow (`full` mode only; `deploy-only` and `k3s` modes have no Terraform at all).
- Does NOT handle matrix builds or monorepos — the `full`/`deploy-only` templates assume exactly `backend/` + `frontend/` as source directories (same assumption `smaqit.infrastructure-deploy-rsync` makes). This assumption does not apply to `k3s` mode, which deploys a container image, not source artifacts.
- **`k3s` mode only:** the generated `build` job builds and pushes whatever `Dockerfile` already exists at the repository root — it does NOT author the `Dockerfile` itself (see `smaqit.infrastructure-image-build-push-static` or an equivalent for the app's own stack) and does NOT manage registry credentials beyond `GITHUB_TOKEN`'s built-in push scope (a private image's pull-side `imagePullSecret` step is rendered from the spec's `Container Registry` row; its `REGISTRY_USERNAME`/`REGISTRY_TOKEN` are synced by `smaqit.infrastructure-repo-config`).
- Does NOT template anything beyond `__APP_DIR__` (`full`/`deploy-only`). `k3s` mode's one renderer, `scripts/render-k3s-deploy.sh`, fills only the app slug, the dispatch options and default, `K3S_TARGETS`, `DEPLOYER_EXEC`, the optional mesh-join step, and the optional imagePullSecret step with its registry host; everything else varies at run time on the dispatched machine's row. Step 6's manifests take only the app slug, the probe user, and the probe fetch command. It is not a general templating engine: if a project needs more variance than that, treat it as a new template variant.
- **`k3s` mode only:** does NOT author `deployment.yaml` or `service.yaml` — the `artifact-build` slot does (`smaqit.infrastructure-image-build-push-*`).

## Examples

**Input:** New project with no CI/CD workflows, `provisioning_mode: provision`. Operator invokes `/cicd.generate`.
**Output:** 3 files written to `.github/workflows/` (`deploy.yml` with sequential provision→deploy jobs, `provision.yml` as PR-only plan with PR comment, `post-merge-deploy.yml` with body sentinel), plus `deployment/scripts/plan-guard.sh` and `deployment/scripts/ownership-guard.sh` vendored.

**Input:** Co-hosted deploy, `provisioning_mode: existing-shared`. Operator invokes `/cicd.generate`.
**Output:** 2 files written to `.github/workflows/` (`deploy.yml` with a single deploy job, `post-merge-deploy.yml`) — no `provision.yml`, nothing vendored, since there's no Terraform to guard.

**Input:** Second, dedicated deploy target provisioned out-of-band, `provisioning_mode: existing-unmanaged`. Operator invokes `/cicd.generate`.
**Output:** Identical file set to the `existing-shared` example above — the generated workflow doesn't encode *why* there's no Terraform, only that there isn't any.

**Input:** Namespace-scoped Kubernetes/k3s target, `provisioning_mode: existing-k3s`, test on a `private-mesh` machine and prod on a `public` one. Operator invokes `/cicd.generate`.
**Output:** 2 files written to `.github/workflows/` (`deploy.yml`, rendered from the spec's Environments table, with a `build` job that pushes the app's image to GHCR followed by a kubectl-driven `deploy` job reading `KUBECONFIG` from a GitHub Environment secret, `post-merge-deploy.yml`), plus `deployment/k3s/scripts/namespace-guard.sh`, `manifest-lint.sh`, and `exec-refused.sh` vendored, and `deployment/k3s/probe-job.yaml` (the private-mesh test machine verifies in-cluster) and `ingress.yaml` (prod, if `own-tls`) rendered — no `provision.yml`, no SSH or Terraform reference anywhere in the generated output.

## Gotchas

These are now encoded directly in the `assets/*.template` files — kept here as maintainer-facing
documentation of *why* the templates look the way they do, for whoever edits the templates
themselves. They are no longer instructions an agent needs to reconstruct from memory on every
generation.

- **`GITHUB_TOKEN` is reserved** — GitHub Actions runner injects its own short-lived installation token under `GITHUB_TOKEN` before any step executes, overwriting any `env.GITHUB_TOKEN` set by the workflow. The templates use `TF_VAR_github_token`; Terraform auto-maps `TF_VAR_*` → `var.*`.
- **PR body sentinel, not label** — Copilot Coding Agent tokens lack `issues:write`. `post-merge-deploy.yml.template` uses `contains(github.event.pull_request.body, 'smaqit:deploy')`, not a label check. Do not change this to a label trigger.
- **`provision.yml.template` also carries `TF_VAR_github_token`** — omitting it causes interactive prompt hangs in PR preview plans.
- **`VITE_*` vars are build-time** — `VITE_DEMO_MODE` (and any other `VITE_*` var) is set as `env:` on the frontend build step specifically, never the deploy step. Vite bakes them at compile time; they are not available at runtime. Do not move this env block to a later step when editing the template.
- **`.terraform.lock.hcl` must be committed** — do NOT add it to `.gitignore`. If CI fails with provider version errors, check that this file is tracked in git.
- **Never remove the plan-guard/ownership-guard steps from `deploy.yml.full.template` or `provision.yml.template`** — a bare `terraform apply -auto-approve` with no gate in front of it is exactly the failure mode Task 084/085 exist to prevent. A `user_data` edit (even a comment-only one) forces VM replacement, and an ungated `apply` in CI carries it out unattended.
- **The guard and the apply in `deploy.yml.full.template` are coupled** — the guard writes the plan it inspected (`plan-guard.sh deployment/terraform --out deployment/terraform/tfplan`) and the apply step applies exactly that file (`terraform apply tfplan`). Don't change the apply back to `terraform apply -auto-approve`: that plans again, so CI could apply changes the guard never saw. If the state changes between the two steps, Terraform rejects the saved plan as stale and the job fails — the intended outcome. The vendored `plan-guard.sh` must support `--out`, which re-vendoring on every run guarantees.
- **`systemctl reload-or-restart nginx`, never a bare `reload`** — cloud-init's `enable` only arranges start-on-next-boot, so nginx may never have actually started. This is baked into both `deploy.yml` templates; don't simplify it back to `reload`.
- **`{{ ... }}` vs `__TOKEN__`** — the substitution tokens this skill uses (`__APP_DIR__`, `__APP_SLUG__`) are deliberately not double-curly-brace shaped, to avoid colliding visually (and, for `yamllint`, semantically — see `yamllint`'s `braces` rule) with GitHub Actions' own `${{ ... }}` expression syntax, which appears throughout every template. Keep any future token in the same `__TOKEN__` shape.
- **`k3s` mode never touches Terraform, SSH, or nginx** — resist the temptation to add any of
  the `full`/`deploy-only` guard-vendoring or `VM_HOST` logic to `deploy.yml.k3s.template` by
  analogy; the k3s target has no VM at all, so none of that applies. `deploy.yml.k3s.template`'s
  `deploy` job sets `KUBECONFIG` from the environment secret and calls
  `smaqit.infrastructure-deploy-k3s-app`'s own guard/lint/apply/verify sequence directly.
- **The `k3s` manifests are rendered here, not by a deploy.** There is no local k3s deploy to render
  them, so Step 6 writes `deployment/k3s/ingress.yaml` (own-tls), `ingress-default-tls.yaml`
  (platform-default-tls), and `probe-job.yaml` (in-cluster verification) with their
  per-environment tokens (`__HOST__`, `__INGRESS_CLASS__`, `__CLUSTER_ISSUER__`, `__IMAGE__`,
  `__HEALTH_PATH__`) left in; the workflow stamps the dispatched machine's values before linting.
  A manifest that was never rendered and committed makes `manifest-lint.sh` fail with a bare
  `FileNotFoundError` — regenerate, never fabricate a placeholder manifest to make the workflow pass.
- **`k3s` GitHub Environments are named after machine slugs, not `test`/`prod`.** Each target
  machine's Environment holds its own `KUBECONFIG` (and, for `private-mesh`, its mesh setup-key
  secret), matching the app's Vault path `secret/apps/<app-slug>/<machine-slug>/kubeconfig`. A
  project that already uses `test`/`prod` Environments must rename them (or re-sync into new ones
  with `smaqit.infrastructure-repo-config`) when it regenerates.
- **The `k3s` workflow is dispatch-only.** `post-merge-deploy.yml` dispatches it after a marked
  merge; a `push` trigger as well deploys the same merge twice. Don't add one back.
- **`k3s` verification depends on where the runner can reach the app.** Every run checks that the
  Deployment runs this commit's image (`ghcr.io/<repo>:<sha>`). A `public` machine whose Ingress
  the app owns is then checked over HTTPS on the spec's health path and `/`, retrying for 60
  seconds. A `private-mesh` machine (the runner's mesh peer reaches only the API) or a
  `platform-owned` Ingress (e.g. behind the platform's SSO gate) is checked by a probe Job in the
  Namespace, which requests the health path over cluster DNS; its verdict is the Job's status, so
  no `pods/log` grant is needed. The probe needs `create`/`delete` on `jobs` in the Namespace.
- **The mesh-join step is the spec's, not this skill's.** Its `uses:` reference, version pin, and
  input names come verbatim from the spec's `Mesh Join Action`, `Mesh Setup-Key Input`, and `Mesh
  Wait-For Input` rows; the setup key comes from the machine's Environment secret the spec names.
  It runs before `kubectl` is installed and waits for the API's mesh name.
- **No secret or event value inside a `run:` script.** `${{ secrets.X }}` or
  `${{ github.event.* }}` written into a script becomes part of the script's text before it runs:
  a secret lands in the rendered script on disk, and an event value (a branch name, a PR body) can
  inject shell. Pass both through the step's `env:` and use `"$VAR"`. Never put a secret on a
  command line either (`--docker-password=…`, `--password …`): the runner's process list shows
  argv. The pull Secret is built by `jq` from `env.*` and handed to `kubectl` through a file
  descriptor (`--from-file=.dockerconfigjson=<(…)`). A Go test fails any template that breaks this.
- **Credential files are owner-only and always removed.** The kubeconfig is written with
  `umask 077` to `$RUNNER_TEMP/kubeconfig` and shredded by a final `if: always()` step; the VM
  templates do the same for `~/.ssh/deploy_key`. A self-hosted runner keeps its disk between jobs.
- **`post-merge-deploy.yml` runs with no checkout.** `gh` normally infers the repository from the
  git remote, so without `--repo` it fails with `fatal: not a git repository` and a marked merge
  never deploys (seen live). Keep `--repo "$GH_REPO"`.
- **`kubectl auth can-i <verb> pods/exec` asks the wrong question.** kubectl parses its argument as
  `TYPE`, `TYPE/NAME`, or a non-resource URL, so `pods/exec` means "a pod named exec" and any
  deployer allowed to create pods gets `yes`. Use `--subresource=exec`, as `namespace-guard.sh`
  does. A step you add to check a subresource must do the same.
- **`bash -e` ends a check on the answer it was waiting for.** GitHub runs every `run:` step with
  `bash -e`, so `ANSWER="$(kubectl auth can-i …)"` (exit 1 means "no") or a `kubectl exec` that must
  be refused ends the step before it is evaluated — the first live private-mesh deploy failed on
  exactly this. Bracket such a command with `set +e` … `RC=$?` … `set -e` and decide from the
  captured output, or call a script that does (`exec-refused.sh`).
- **One `k3s` deploy per machine at a time.** The `deploy` job's `concurrency` group is keyed on the
  dispatched machine with `cancel-in-progress: false`, so a second dispatch waits instead of
  rolling out over the first. Don't set `cancel-in-progress: true`: it would stop a rollout midway.
- **A newly generated `deploy.yml` cannot be `workflow_dispatch`'d until it exists on the default
  branch** — GitHub's `workflow_dispatch` API 404s for a workflow file that only exists on a
  non-default branch. That is why an existing-k3s project's first deploy (the `ci-bootstrap`
  deployment path) lands `deploy.yml`, the vendored scripts, and the manifests through a PR
  *without* the deploy marker, and dispatches only after the merge: `post-merge-deploy.yml`
  dispatches with no environment input, so a marked first PR would deploy the default (prod)
  machine instead of the test one.

### Workflows outside these templates (machine repositories)

A repository that owns machines rather than an app writes its own provisioning workflows — this
skill has no template for them. Two mistakes each cost a failed live run there; any Terraform
workflow, templated or not, must avoid them:

- **Terraform S3-backend credentials belong in job-level `env:`**, never step-level `env:` on the
  `terraform init` step alone. Steps do not inherit a sibling step's `env:`, and every `terraform`
  command except `validate` reads the backend — `plan` and `apply` included. Step-scoped
  credentials fail the first later step with `No valid credential sources found`. The app
  templates here (`provision.yml.template`, `deploy.yml.full.template`) already set them at job
  level. On a non-AWS S3 endpoint such as Cyso Object Storage, keep the backend's `skip_*`
  settings as `smaqit.infrastructure-provider-cyso` documents them; Terraform has had regressions
  with third-party S3 backends (hashicorp/terraform#37130).
- **A gated apply against an existing machine needs a destroy guard, not only a human reading the
  plan.** A `user_data`, `image_id`, or `key_pair` diff forces the instance to be destroyed and
  recreated. Run `plan` through `plan-guard.sh` (vendored from
  `smaqit.infrastructure-provision-cyso`) before `apply`, and set
  `lifecycle { prevent_destroy = true }` on the instance and its data volumes, so that a
  replacement fails the plan instead of running in CI.
- **A deliberate replacement goes through the same guard, named.** When a gated workflow offers a
  "replace this resource" input (the only way to rebuild an instance whose `user_data` is under
  `ignore_changes`), pass it as `plan-guard.sh <dir> --allow-replace "$INPUT" --out tfplan` and
  apply with `terraform apply tfplan`. The guard plans with `-replace=`, accepts only that
  replacement, and keeps the exact plan it inspected; a bare `terraform plan -replace=… && apply`
  bypasses it. `prevent_destroy` on that resource must be lifted for the run, or the plan fails.
- **A deliberate removal of one resource goes through the guard the same way.** A gated workflow
  that offers a "remove this resource" input (closing a temporary firewall opening, retiring one
  port) passes it as `plan-guard.sh <dir> --allow-delete "$INPUT" --out tfplan`. The flag adds
  nothing to the plan: the reviewed configuration change produces the destroy, and the guard
  accepts only a pure destroy of that address. Pass the input to the step through `env:`, never
  interpolated into the script.
- **Both inputs are typed by a human, never filled from guard output.** A workflow or agent must not
  derive the replace/remove address from a BLOCKED plan; when the guard blocks, the run fails and a
  human decides. Never add a step that retries the guard with addresses parsed from its own output.

## Completion

- [ ] Existing workflows read (or confirmed absent)
- [ ] Generation mode determined (`full`, `deploy-only`, or `k3s`)
- [ ] `__APP_DIR__` resolved for `full`/`deploy-only` (from Infrastructure specs, or defaulted to `/opt/<project-slug>`), or `smaqit spec environments --json` passing for `k3s`
- [ ] Correct template variant copied for `deploy.yml` (`full` vs `deploy-only` vs `k3s`)
- [ ] The mode's token substituted in the copied `deploy.yml` (`k3s`: rendered by `scripts/render-k3s-deploy.sh`, exit 0)
- [ ] `provision.yml` copied verbatim (`full` only); NOT generated in `deploy-only` or `k3s` mode
- [ ] `post-merge-deploy.yml` copied verbatim — all modes, identical content
- [ ] `deployment/scripts/plan-guard.sh` and `deployment/scripts/ownership-guard.sh` vendored (`full` only); `deployment/k3s/scripts/namespace-guard.sh`, `manifest-lint.sh`, and `exec-refused.sh` vendored (`k3s` only); nothing vendored for `deploy-only`
- [ ] `k3s` only: the Ingress file for each Ingress mode in use and `probe-job.yaml` (when any environment verifies in-cluster) rendered from deploy-k3s-app's templates, per-environment tokens left in
- [ ] `k3s` mode only: generated `deploy.yml` is dispatch-only, has a `build` job preceding `deploy`, lists every environment's machine slug, keeps the mesh-join step exactly when some environment is `private-mesh`, and keeps the imagePullSecret step exactly when the `Container Registry` is private
- [ ] Report delivered, stating which generation mode was used

## Failure Handling

| Situation | Action |
|-----------|--------|
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| Output artifact already exists | Confirm with user before overwriting |
| A step the project added to the generated `deploy.yml` checks a permission with `kubectl auth can-i <verb> <resource>/<subresource>` and gets `yes` despite the platform denying it | kubectl read `<resource>/<subresource>` as a named resource. Rewrite it with `--subresource=<name>` (see Gotchas); don't trust or act on the misleading `yes`. |
| A step the project added ends on a command that was expected to fail (a `can-i` "no", a refused `kubectl exec`) before its verdict runs | GitHub's `bash -e` stopped it. Bracket the command with `set +e` … `RC=$?` … `set -e` and decide from the captured result, or call the vendored `exec-refused.sh`. Never `\|\| true`, which also turns a real error into a pass. |
| Workflows already exist and differ | Show a diff summary; ask whether to overwrite or merge |
| Infrastructure spec not available | Default `__APP_DIR__` to `/opt/<project-slug>` (`full`/`deploy-only`); note the value used in the report. `k3s` cannot render without the spec's Environments table — stop and name it |
| `smaqit spec environments` reports `SLOTS-SPEC-*` lines (`k3s`) | Stop and surface them verbatim; the fix belongs in the Infrastructure spec, never in a hand-filled workflow |
| `render-k3s-deploy.sh` refuses (`k3s`) | Surface its stderr; fix the spec (one machine slug per environment, mesh access rows for `private-mesh`) or pass `--default-env` |
| A template's content needs to differ from the shipped `assets/*.template` for this specific project | Treat that as a reason to add a new template variant to this skill (or flag it as a limitation), not to hand-edit the copied output — hand edits silently diverge from the validated template on the next re-generation |
| `deploy-only` or `k3s` mode requested but a `provision.yml` already exists from a prior `full` generation | Delete the stale `provision.yml` and the vendored `deployment/scripts/*.sh` — leaving them in place would plan/guard Terraform that no longer matches this project's `provisioning_mode` |
| `deployment/scripts/plan-guard.sh` or `ownership-guard.sh` already exist and differ from the skill directory's current copies | Overwrite — they are generated output, regenerated on every run (see Step 5) |
| `deployment/k3s/scripts/namespace-guard.sh`, `manifest-lint.sh`, or `exec-refused.sh` already exist and differ from `smaqit.infrastructure-deploy-k3s-app`'s current copies | Overwrite — same generated-output rule as the Terraform guards |
| `deployment/k3s/deployment.yaml` missing when Step 6 needs the probe's `runAsUser` (`k3s`) | Stop. The `artifact-build` slot authors it first; run that slot (or the app stack's image-build-push skill), then generate again. Never guess a UID |
| A rendered Ingress or `probe-job.yaml` already exists and differs (`k3s`) | Overwrite — it is generated output. A project-specific change belongs in the spec or a new template variant |
| `k3s` mode requested but the project's existing `deploy.yml` references SSH/Terraform (stale, from a prior `full`/`deploy-only` generation) | Overwrite with `assets/deploy.yml.k3s.template`, remove the stale `deployment/scripts/*.sh` and `provision.yml` if present, and vendor `deployment/k3s/scripts/*.sh` instead — a mode switch means the target itself changed; do not attempt to merge old VM-specific steps into the new job |
