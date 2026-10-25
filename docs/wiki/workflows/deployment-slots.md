# Deployment Slots

The deployment phase runs a fixed, ordered sequence of **slots** (variation points). Each slot has a contract and a default filler skill. A project binds any slot to its own skill, or skips it, in `.smaqit/slots.yml`. An unbound slot uses the default for the resolved `provisioning_mode`.

## Why Slots

Before slots, the infrastructure wiring — credentials, CI generation, provisioning, host bootstrap, deploy, verify, TLS — was hardcoded per `provisioning_mode` inside two app-centric orchestrator skills (`smaqit.project-new` and `smaqit.feature-new`). That worked for an app deploying onto its own VM, but a repository that owns machines rather than an app (a platform repository) had no way in: every canonical infrastructure skill assumed an app-owned VM over public SSH, and adding one more `provisioning_mode` branch per repository shape does not scale.

Slots keep the determinism smaqit relies on — the order is fixed and resolution is mechanical — while letting each repository own the parts that genuinely differ. This follows the thin-dispatcher precedent of the k3s onboarding skills: the repository that owns a mechanism defines its contract; smaqit defines the hand-off points.

## Slot Catalogue

One environment run walks these slots in order. Day-two slots never run as part of a full sequence.

The contracts come from the resolver's catalogue (`installer/slots.go`); `smaqit slots contract
[--slot=<slot>] [--mode=<mode>] [--json]` prints the same text. A filler that meets a slot's contract,
and its mode's note, replaces the default with nothing else to change — that is how a platform ships
its own fillers for apps to bind.

<!-- smaqit:slot-contracts:begin (generated from installer/slots.go; go test -run TestSlotContractsMatchWiki -update) -->
| # | Slot | Purpose | Inputs | Handoff | Completion signal |
|---|------|---------|--------|---------|-------------------|
| 1 | `platform-prereqs` | Prepare platform-level prerequisites the target needs before credentials or onboarding | Infrastructure spec, environment | Reachable target network | Prerequisite check passes |
| 2 | `access-credentials` | Load the credentials needed to reach the provider, platform, and repository | Vault paths for the mode | Credentials in Vault (never in artifacts) | Required paths populated |
| 3 | `platform-onboarding` | Request tenancy on a platform-owned target and gate on its approval | Platform repo, registry file, environment | Approved onboarding (merged request) | Request merged; stop while pending |
| 4 | `target-credentials` | Load credentials the target issues or authorizes for this app | Onboarding result or machine registration | Target credential in Vault (for example a per-app SSH key) | Credential populated and usable |
| 5 | `environment-config` | Configure repository and environment secrets and variables for CI/CD | Credentials, target facts | Repository and environment configuration | Expected names present |
| 6 | `artifact-build` | Author and build deployable artifacts (container image, manifests) | Stack spec, source tree | Image and manifests | Image built (and pushed when required) |
| 7 | `iac-generate` | Generate Infrastructure as Code from the consolidated specifications | Infrastructure and Stack specs | IaC configuration with credential references | IaC written and structurally valid |
| 8 | `ci-generate` | Generate or update the CI/CD workflows that deploy the target | Mode, Infrastructure spec | Workflow files | Workflows written and linted |
| 9 | `provision` | Provision or reconcile the target infrastructure through the guarded IaC tool | IaC configuration, credentials | Provisioned target (host, fixed IP) | Guarded plan/apply succeeds |
| 10 | `host-bootstrap` | Bootstrap the host operating system and runtime the app deploys onto | Target host, credentials | Host ready for deploys | Bootstrap checks pass |
| 11 | `app-deploy` | Deploy the application onto the target | Artifacts, target, credentials | Running revision | Deploy completes |
| 12 | `verify` | Verify the deployed revision is healthy and serving | Target URL, expected revision | Verification report | All checks pass |
| 13 | `tls` | Issue and install TLS for a declared custom domain | Domain, DNS, target | Certificate in place | HTTPS smoke test passes |
| — | `replace` | Replace a target in place (day-two operation) | Target identity | Replacement target | Replacement verified |
| — | `teardown` | Destroy the target through the guarded IaC tool (day-two operation) | IaC state | Removed target | Guarded destroy succeeds |

**`existing-k3s`** — what a filler must do in each slot:

- **`platform-prereqs`:** None. A filler must not act on the target: the deploy job joins any mesh itself.
- **`access-credentials`:** Loads only what this machine needs to request onboarding (the default: the `platform-repo` token). Never a deploy credential: the deploy authenticates with the CI job's OIDC identity.
- **`platform-onboarding`:** Inputs: the spec's `Platform Repo` and `Registry File Path`, the environment's `Machine Slug`, and this repository's `owner/repo` and numeric ID. Hands off a merged registry request; stops while pending. Never dispatches the platform's workflows: the first deploy's namespace guard confirms the binding (403 until the platform reconciles).
- **`target-credentials`:** None: the target issues the app nothing. A filler must not store a deploy credential.
- **`environment-config`:** None: nothing is synced, and GitHub creates the machine-slug Environment on the first run that references it. A filler may set non-secret variables or protection rules only.
- **`artifact-build`:** Authors the `Dockerfile` and `deployment/k3s/deployment.yaml` and `service.yaml`: a ClusterIP Service, Pod Security `restricted`, an explicit numeric `runAsUser`, `DEPLOY_SHA`/`DEPLOY_TIME` as pod environment, no `imagePullSecrets`, no Ingress. Runs on this machine and never pushes: the workflow's `build` job does.
- **`iac-generate`:** None: the app owns no infrastructure on the platform.
- **`ci-generate`:** Writes a dispatch-only `deploy.yml` with an `environment` input of machine slugs. Its `deploy` job runs in the machine-slug GitHub Environment with `permissions` `contents: read` and `id-token: write`; joins the mesh when the row names a mesh secret; connects through the spec's `Cluster Connect` action; and runs the guard, lint, apply, rollout, and exec check plus the steps "Verify revision" and "Verify in cluster" or "Verify over HTTPS" (`verify` reads them by name). Also writes the Ingress and probe manifests the spec's allowed modes need.
- **`provision`:** None; refused on `ci-bootstrap`. A binding that would run on this machine is refused (`SLOTS-LOCAL-DEPLOY-REFUSED`).
- **`host-bootstrap`:** None; refused on `ci-bootstrap` and `existing-cicd-pr`. The platform owns the host.
- **`app-deploy`:** The deploy runs in the workflow; the agent reads this filler only to trigger and monitor it. Inputs: the deployment path, the environment's machine slug, and the intended commit (merge or pushed SHA). Triggers the run (dispatch at the default branch, or the merged PR's dispatcher), selects it by workflow, event, commit, and creation time, stops when more than one run matches, and waits for it to finish. Hands the run's ID to `verify`; a failed run stops the walk. Never deploys from this machine.
- **`verify`:** Decides from the deploy run's record only (default: deploy-verify `ci-evidence`), given the run's ID and the intended commit. Never reaches the target from this machine.
- **`tls`:** None. A binding is refused (`SLOTS-LOCAL-DEPLOY-REFUSED`): no path runs it in CI, and the platform serves TLS or awaits the app's `Certificate` inside the deploy.
- **`replace`:** None. A binding is refused (`SLOTS-LOCAL-DEPLOY-REFUSED`): no path runs it in CI.
- **`teardown`:** None. A binding is refused (`SLOTS-LOCAL-DEPLOY-REFUSED`): offboarding is the platform's process.
<!-- smaqit:slot-contracts:end -->

`tls` defaults to none in every full sequence; it runs as a single-slot run (project-new's conditional Phase 6) or when bound, and its filler requires a custom domain declared in the Infrastructure specification.

## Default Fillers by Provisioning Mode

`provisioning_mode` (resolved by `smaqit.input-deployment`) selects the default profile:

| Slot | `provision` / `existing-owned` | `existing-shared` / `existing-unmanaged` | `existing-k3s` |
|------|-------------------------------|------------------------------------------|----------------|
| `platform-prereqs` | none | none | none (the generated `deploy.yml` joins a private mesh itself) |
| `access-credentials` | vault-loader `access` | vault-loader `access` | vault-loader `access` |
| `platform-onboarding` | none | none | request-k3s-onboarding `request` |
| `target-credentials` | none | vault-loader `target` | none (the deploy job connects with its own OIDC identity) |
| `environment-config` | repo-config `sync` | repo-config `sync` | none (no deploy secret; GitHub creates the machine-slug Environment) |
| `artifact-build` | none | none | `image-build-push*` family `build` |
| `iac-generate` | agent builtin | none | none |
| `ci-generate` | cicd-generate `generate` | cicd-generate `generate` | cicd-generate `generate` |
| `provision` | provision-cyso `apply` | none | none |
| `host-bootstrap` | vm-bootstrap `bootstrap` | vm-bootstrap `bootstrap` | none |
| `app-deploy` | `deploy-rsync*` family `deploy` | `deploy-rsync*` family `deploy` | `deploy-k3s-app*` family `deploy`, performed only by the generated `deploy.yml` in CI |
| `verify` | deploy-verify `verify` | deploy-verify `verify` | deploy-verify `ci-evidence` (every path existing-k3s plans) |
| `tls` | domain-tls `issue` | domain-tls `issue` | none (`own-tls`: the Certificate is awaited inside `app-deploy`; otherwise the platform serves TLS) |
| `replace` | none | none | none |
| `teardown` | agent builtin destroy (`provision` only) | none | none |

A **family** default (e.g. `deploy-rsync*`) lists every installed skill with that name prefix that declares the slot; the orchestrator picks one by the declared stack, or synthesizes one when none matches. A family member whose `metadata.smaqit-slots` cannot be read is skipped with a plan warning rather than an error.

Skill names above drop the `smaqit.infrastructure-` prefix for brevity.

## Deployment Paths

The deployment path adjusts the profile:

| Path | Used by | Effect |
|------|---------|--------|
| `standard` | Local runs (project-new VM dev sweep) | Every slot executes locally; `environment-config` and `tls` default to none. Transitional, for VM modes only: refused for `existing-k3s` |
| `existing-cicd-pr` | `smaqit.feature-new` | `iac-generate`, `ci-generate`, `provision`, `host-bootstrap` are **refused** (the PR path never regenerates workflows or provisions directly); `platform-prereqs`, `platform-onboarding` and `tls` default to none; `artifact-build` and `app-deploy` are executed by CI and monitored; `existing-k3s`'s `verify` defaults to deploy-verify `ci-evidence` |
| `ci-push` | Project-new production | `platform-prereqs`, `iac-generate`, `ci-generate`, `host-bootstrap`, `tls` default to none (workflows were committed in the dev sweep; under `provision`/`existing-owned` cloud-init configures the production host, and under `existing-shared`/`existing-unmanaged` the host is already bootstrapped); `artifact-build`, `provision`, `app-deploy` are executed by CI and monitored; `existing-k3s`'s `verify` defaults to deploy-verify `ci-evidence` |
| `ci-bootstrap` | Project-new Phase 4 for `existing-k3s` (any CI-only mode's first deploy to an environment) | `provision` and `host-bootstrap` are **refused**; `platform-prereqs` and `tls` default to none; the onboarding PR's credential, onboarding, `artifact-build` (authoring) and `ci-generate` run locally; `app-deploy` is executed by CI; `verify` defaults to deploy-verify `ci-evidence`. The agent lands the generated workflows and manifests through a PR without the deploy marker, waits for the human merge, then dispatches `deploy.yml` for the environment's machine slug. Refused for VM modes |

Two rules hold on every path:

- **Single-slot runs** (`--slot=<slot>`) ignore the path's full-sequence defaults — that is how `tls` and `teardown` run on demand — but never its refusals. A path's per-mode default (`ci-evidence` for `existing-k3s` `verify` on the CI paths) still applies to a single-slot run.
- **No split CI runs:** the slots a pipeline performs run as one pipeline run, so every locally executed slot must come before the first CI-executed slot (or be `verify`, which runs after; on the CI paths `existing-k3s`'s `ci-evidence` only reads the finished run's record). A local binding that would land between CI-executed slots — e.g. a bound `host-bootstrap` between CI `provision` and CI `app-deploy` under `ci-push` — is refused with `SLOTS-BINDING-REFUSED`.

## CI-Only and Transitional Modes

A deploy to a target runs in the CI/CD pipeline (GitHub Actions by default); the operator's machine
only configures (credentials, onboarding requests, environment secrets), generates (workflows, IaC,
manifests), authors artifacts, and reads the pipeline's evidence. The slots that act on the target
are `provision`, `host-bootstrap`, `app-deploy`, `tls`, `replace`, and `teardown`.

| Mode | Deploys from | Enforcement |
|------|--------------|-------------|
| `existing-k3s` | CI only: `ci-bootstrap` for an environment's first deploy, then `ci-push` or `existing-cicd-pr` | `smaqit slots plan` refuses any target-acting slot that would run locally — a default, a binding, or a single-slot run — with `SLOTS-LOCAL-DEPLOY-REFUSED`. There is no break-glass path |
| `provision`, `existing-owned`, `existing-shared`, `existing-unmanaged` | Transitional: the local `standard` path for the dev sweep, CI for production | None yet. Their workflows can't deploy a non-production environment, write the nginx vhost, `docker-compose.yml` or `.env`, bootstrap the host, issue TLS, or cover the Python stacks; until they can, the local path stays |

The Deployment agent drives every CI deploy through one contract, the ci-monitor seam: trigger the
run (push, PR merge, or dispatch), select it by commit, wait for it to finish, and read its verify
evidence. `gh` against GitHub Actions implements it; another CI system would need its own workflow
templates and its own implementation of those four operations.

Because `existing-k3s` runs no target-acting slot locally, a project binding of one (for example its
own `teardown` or `tls` skill) is refused on every path: no path runs those slots in CI.

## Configuring Slots

`smaqit init` creates a commented `.smaqit/slots.yml` once and never overwrites it; `smaqit update` and `smaqit uninstall` both preserve it. With the file untouched, every slot uses its default and behaviour is unchanged.

```yaml
schemaVersion: 1
phase: deploy
slots:
  target-credentials:
    skill: smaqit.infrastructure-vault-loader
    operation: register
  environment-config:
    skill: smaqit.infrastructure-repo-config
    operation: sync-machine
  provision:
    skill: platform.machine-provision
    operation: apply
    environments: [prod]
  host-bootstrap:
    skill: platform.k3s-baseline
  tls: none
```

This is a platform repository — one that owns machines rather than an app (see
[Platform Repositories](platform-repositories.md) for the pattern, a machine-verification checklist,
and the gotchas found on live machines). It binds two shipped operations and two of its own skills:

- **`target-credentials` → vault-loader `register`** runs `register-machine.sh <machine-slug>`:
  it populates `secret/machines/<machine-slug>/{base-ssh,cyso,tfstate,metadata}` the first time
  and keeps whatever already exists. After `provision` applies, the machine's address is
  backfilled with `register-machine.sh <machine-slug> --host <ip>`.
- **`environment-config` → repo-config `sync-machine`** runs `sync-machine-env.sh`: it creates a
  GitHub Environment named after the machine slug, preflights the token, and syncs the machine's
  credentials into it. It runs before `provision`, so on a new machine the `FIXED_IP` variable is
  skipped; once the address is backfilled, re-run the slot alone
  (`smaqit slots plan --phase=deploy --mode=provision --env=prod --slot=environment-config`) to set it.

Both shipped skills declare two operations for their slot, so these bindings must name the
operation; `smaqit slots plan` refuses a binding that leaves it out (`SLOTS-OPERATION-INVALID`).

A bound skill must declare the slot in its own frontmatter:

```yaml
metadata:
  version: "1.0.0"
  smaqit-slots: "provision=apply, replace=replace"
```

The value is a single string of `slot=operation` pairs because the agentskills.io specification limits `metadata` values to strings. Project-owned fillers live in the project's own skill directories (`.agents/skills/`, `.claude/skills/`, `.github/skills/`), which are searched before the global ones.

### Platform fillers

A platform can ship its own fillers for the apps it hosts, so an app follows the platform's
conventions without smaqit knowing them: the platform publishes skills that declare the slots they
fill (`metadata.smaqit-slots`), and each app repository binds them in its `.smaqit/slots.yml`. A
filler replaces the default with nothing else to change as long as it meets the slot's contract and
the mode's note (`smaqit slots contract --slot=<slot> --mode=<mode>`).

For `existing-k3s` the useful substitution points are:

- **`platform-onboarding`** — the platform's own registry request (its entry format, its
  repository fields).
- **`artifact-build`** — the platform's image and manifest conventions (still a ClusterIP Service,
  Pod Security `restricted`, no `imagePullSecrets`).
- **`ci-generate`** — the platform's own deploy workflow, as long as it keeps the steps `verify`
  reads by name ("Verify revision", and "Verify in cluster" or "Verify over HTTPS").
- **`app-deploy`** — how the deploy run is triggered and monitored. The deploy itself always runs in
  the workflow `ci-generate` wrote; the agent reads this filler only to dispatch the run, select it
  by commit, wait for it, and hand its ID to `verify`.

`target-credentials`, `environment-config`, and `tls` are none for `existing-k3s`, and a binding of
`tls`, `replace`, or `teardown` is refused: no path runs them in CI.

## Inspecting and Validating

```bash
smaqit slots plan --phase=deploy --mode=existing-k3s --env=test --path=ci-bootstrap
smaqit slots contract --slot=app-deploy --mode=existing-k3s
smaqit slots plan --phase=deploy --mode=existing-owned --env=prod --path=existing-cicd-pr --json
smaqit slots plan --phase=deploy --mode=provision --env=dev --slot=teardown
smaqit slots validate
```

`smaqit validate` also checks `.smaqit/slots.yml` when it exists. Failures are reported together:

| Code | Meaning |
|------|---------|
| `SLOTS-CONFIG-INVALID` | The file is not valid YAML or violates the schema |
| `SLOTS-UNKNOWN-SLOT` | A binding names a slot that is not in the catalogue |
| `SLOTS-SKILL-MISSING` | A bound (or default) skill is not installed |
| `SLOTS-SLOT-UNDECLARED` | The skill does not declare the slot in `metadata.smaqit-slots` |
| `SLOTS-OPERATION-INVALID` | The operation is not declared, or is ambiguous |
| `SLOTS-BINDING-REFUSED` | The deployment path refuses a binding for that slot |
| `SLOTS-LOCAL-DEPLOY-REFUSED` | A CI-only mode (`existing-k3s`) would act on the target from this machine; plan it on a CI path |
| `SLOTS-SKILL-INVALID` | A skill's frontmatter or slot declaration is malformed |
| `SLOTS-ARGUMENT-INVALID` | A missing or unknown command-line argument |
| `SLOTS-SPEC-MISSING` | `existing-k3s`: no Infrastructure spec declares an `## Environments` table, or it has no row for `--env` |
| `SLOTS-SPEC-FIELD-MISSING` | `existing-k3s`: the row (or a Constraints row it depends on) leaves a required field empty |
| `SLOTS-SPEC-FIELD-INVALID` | `existing-k3s`: a field has a value outside its allowed set, or a value its mode forbids |
| `SLOTS-SPEC-AMBIGUOUS` | `existing-k3s`: more than one Infrastructure spec declares an `## Environments` table |

## existing-k3s Targets

An `existing-k3s` app's environments can differ in two ways that change every deployment step,
so each environment declares them in the Infrastructure spec's `## Environments` table instead of
in `.smaqit/slots.yml` binding `params`:

- **Reachability** describes the app. A `public` app is reachable from the internet and verified
  over HTTPS; a `private-mesh` app is reachable only from peers of a private mesh network
  (WireGuard-based meshes are common), with its hostnames in the mesh's private DNS zone, and is
  verified in-cluster.
- **Mesh access** describes the API. A row that names a `Mesh Setup-Key Secret` (always for
  `private-mesh`; for `public` when the platform keeps its API mesh-only even though the app is
  public) makes CI join the mesh first, with the platform's shared, per-machine key — a secret the
  platform holds, for example an organisation secret, never one the app stores.
- **Ingress mode.** `own-tls` is the original behaviour: the app's Ingress asks cert-manager for
  its own certificate. `platform-default-tls` means the platform already serves one certificate
  (often a DNS-01 wildcard) as its ingress controller's default, so the app's Ingress has no `tls`
  section. `platform-owned` means the platform runs the Ingress itself, for example behind its
  single sign-on gate, and the app ships none.

The default fillers read the row, not binding parameters. `ci-generate` renders one `deploy.yml`
whose steps branch on the dispatched machine's row (mesh join, Ingress, Certificate wait), plus the
Ingress and probe manifests the environments need. That workflow performs `app-deploy`: it joins
the mesh when the row names a mesh secret, connects with the job's own GitHub Actions OIDC identity
through the platform's `Cluster Connect` action (the platform binds the Namespace to this repository
and the machine-slug Environment, so the app holds no kubeconfig), applies the Ingress the mode calls for, and checks the app from inside
the cluster with a short-lived probe Job whenever the runner can't reach it from outside
(`private-mesh` or `platform-owned`). The probe's verdict is the Job's own status, so it works even
where the platform denies `pods/log`.

Nothing deploys from the operator's machine, so it never needs mesh access or a local kubeconfig.
`verify` is always deploy-verify's `ci-evidence`, which reads the deploy run with
`gh run view --json headSha,status,conclusion,jobs` and passes only when the run succeeded, its head
is the intended commit (the PR merge SHA on `existing-cicd-pr` and `ci-bootstrap`, or the pushed SHA) or a later commit that contains it,
and its `deploy` job's "Verify revision" step and the row's verify step ("Verify in cluster" or
"Verify over HTTPS") both succeeded.

Spec-wide Constraints rows apply to every environment:
- `Cluster Connect Action`, `Cluster Connect Machine Input`, `Cluster Connect Namespace Input`
  (required): the platform's pinned action and its input names. It owns the API address, the
  cluster CA, the token audience, and the token's refresh, so a CA change is a new action tag.
- `Allowed Ingress Modes`: the exposure the platform permits, as a comma list (absent: all three).
  An Environments row using any other mode is refused. A platform that owns every Ingress declares
  `platform-owned`. `manifest-lint.sh` separately refuses any Service that isn't ClusterIP, so a
  manifest that would bypass the Ingress fails before apply rather than at the platform's admission.
- `Probe Client`: `curl`, `wget`, or `python3`, whichever the app's image already contains.
- `Deployer Exec`: defaults to `denied`. The deploy then fails when the deployer could exec into
  pods, checked both by asking the API and by a real `kubectl exec` that must be refused. A
  platform that grants exec on purpose is recorded as `allowed`, which turns the failure into a
  warning.

The platform pulls the app's images itself, so the workflow creates no pull Secret and the
manifests carry no `imagePullSecrets`.

There is no default. `smaqit slots plan --mode=existing-k3s` refuses to plan an environment whose
row is incomplete, and `smaqit spec environments [--env=<env>] [--json]` validates and prints the
table — the same output the fillers consume. The trade-off is deliberate: a spec written before
this table existed must be updated before its next deployment, in exchange for every k3s skill
agreeing on one source of truth for what each machine is.

## Invariants

Every filler — default or project-owned — runs under the deployment phase's invariants: IaC plan/apply only through the project's guard script, no out-of-band changes to IaC-managed resources, and credential references only. When a filler's instructions would break one, the deployment agent stops and reports the conflict instead of following them.

## Further Reading

- [Extending smaqit](extending-smaqit.md) — Where framework extension points live
- [Quickstart](quickstart.md) — Using smaqit in a project
