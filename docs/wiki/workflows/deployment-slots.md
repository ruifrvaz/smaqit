# Deployment Slots

The deployment phase runs a fixed, ordered sequence of **slots** (variation points). Each slot has a contract and a default filler skill. A project binds any slot to its own skill, or skips it, in `.smaqit/slots.yml`. An unbound slot uses the default for the resolved `provisioning_mode`.

## Why Slots

Before slots, the infrastructure wiring — credentials, CI generation, provisioning, host bootstrap, deploy, verify, TLS — was hardcoded per `provisioning_mode` inside two app-centric orchestrator skills (`smaqit.new-greenfield-project` and `smaqit.feature-new`). That worked for an app deploying onto its own VM, but a repository that owns machines rather than an app (a platform repository) had no way in: every canonical infrastructure skill assumed an app-owned VM over public SSH, and adding one more `provisioning_mode` branch per repository shape does not scale.

Slots keep the determinism smaqit relies on — the order is fixed and resolution is mechanical — while letting each repository own the parts that genuinely differ. This follows the thin-dispatcher precedent of the k3s onboarding skills: the repository that owns a mechanism defines its contract; smaqit defines the hand-off points.

## Slot Catalogue

One environment run walks these slots in order. Day-two slots never run as part of a full sequence.

| # | Slot | Purpose | Inputs | Handoff | Completion signal |
|---|------|---------|--------|---------|-------------------|
| 1 | `platform-prereqs` | Platform-level prerequisites the target needs first (e.g. network or mesh access) | Infrastructure spec, environment | Reachable target network | Prerequisite check passes |
| 2 | `access-credentials` | Credentials to reach the provider, platform, and repository | Vault paths for the mode | Credentials in Vault (never in artifacts) | Required paths populated |
| 3 | `platform-onboarding` | Request tenancy on a platform-owned target and gate on approval | Platform repo, registry file, environment | Approved onboarding (merged request) | Request merged; stop while pending |
| 4 | `target-credentials` | Credentials the target issues or authorizes for this app | Onboarding result or machine registration | Target credential in Vault (kubeconfig, per-app SSH key) | Credential populated and usable |
| 5 | `environment-config` | Configure repository/environment secrets and variables for CI/CD | Credentials, target facts | Repository and environment configuration | Expected names present |
| 6 | `artifact-build` | Author and build deployable artifacts | Stack spec, source tree | Image and manifests | Image built (and pushed when required) |
| 7 | `iac-generate` | Generate Infrastructure as Code from consolidated specs | Infrastructure and Stack specs | IaC configuration with credential references | IaC written and structurally valid |
| 8 | `ci-generate` | Generate or update the deploying CI/CD workflows | Mode, Infrastructure spec | Workflow files | Workflows written and linted |
| 9 | `provision` | Provision or reconcile the target through the guarded IaC tool | IaC configuration, credentials | Provisioned target (host, fixed IP) | Guarded plan/apply succeeds |
| 10 | `host-bootstrap` | Bootstrap the host OS and runtime | Target host, credentials | Host ready for deploys | Bootstrap checks pass |
| 11 | `app-deploy` | Deploy the application | Artifacts, target, credentials | Running revision | Deploy completes |
| 12 | `verify` | Verify the deployed revision is healthy and serving | Target URL, expected revision | Verification report | All checks pass |
| 13 | `tls` | Issue and install TLS for a declared custom domain | Domain, DNS, target | Certificate in place | HTTPS smoke test passes |
| — | `replace` | Replace a target in place (day two) | Target identity | Replacement target | Replacement verified |
| — | `teardown` | Destroy the target through the guarded IaC tool (day two) | IaC state | Removed target | Guarded destroy succeeds |

`tls` defaults to none in every full sequence; it runs as a single-slot run (greenfield's conditional Phase 6) or when bound, and its filler requires a custom domain declared in the Infrastructure specification.

## Default Fillers by Provisioning Mode

`provisioning_mode` (resolved by `smaqit.input-deployment`) selects the default profile:

| Slot | `provision` / `existing-owned` | `existing-shared` / `existing-unmanaged` | `existing-k3s` |
|------|-------------------------------|------------------------------------------|----------------|
| `platform-prereqs` | none | none | deploy-k3s-app `mesh-check` (passes without probing for a `public` environment) |
| `access-credentials` | vault-loader `access` | vault-loader `access` | vault-loader `access` |
| `platform-onboarding` | none | none | request-k3s-onboarding `request` |
| `target-credentials` | none | vault-loader `target` | vault-loader `target` |
| `environment-config` | repo-config `sync` | repo-config `sync` | repo-config `sync` |
| `artifact-build` | none | none | `image-build-push*` family `build` |
| `iac-generate` | agent builtin | none | none |
| `ci-generate` | cicd-generate `generate` | cicd-generate `generate` | cicd-generate `generate` |
| `provision` | provision-cyso `apply` | none | none |
| `host-bootstrap` | vm-bootstrap `bootstrap` | vm-bootstrap `bootstrap` | none |
| `app-deploy` | `deploy-rsync*` family `deploy` | `deploy-rsync*` family `deploy` | `deploy-k3s-app*` family `deploy` |
| `verify` | deploy-verify `verify` | deploy-verify `verify` | deploy-verify `verify` |
| `tls` | domain-tls `issue` | domain-tls `issue` | none (`own-tls`: the Certificate is awaited inside `app-deploy`; otherwise the platform serves TLS) |
| `replace` | none | none | none |
| `teardown` | agent builtin destroy (`provision` only) | none | none |

A **family** default (e.g. `deploy-rsync*`) lists every installed skill with that name prefix that declares the slot; the orchestrator picks one by the declared stack, or synthesizes one when none matches. A family member whose `metadata.smaqit-slots` cannot be read is skipped with a plan warning rather than an error.

Skill names above drop the `smaqit.infrastructure-` prefix for brevity.

## Deployment Paths

The deployment path adjusts the profile:

| Path | Used by | Effect |
|------|---------|--------|
| `standard` | Local runs (greenfield dev sweep) | Every slot executes locally; `environment-config` and `tls` default to none |
| `existing-cicd-pr` | `smaqit.feature-new` | `iac-generate`, `ci-generate`, `provision`, `host-bootstrap` are **refused** (the PR path never regenerates workflows or provisions directly); `platform-onboarding` and `tls` default to none; `artifact-build` and `app-deploy` are executed by CI and monitored |
| `ci-push` | Greenfield production | `iac-generate`, `ci-generate`, `host-bootstrap`, `tls` default to none (workflows were committed in the dev sweep; under `provision`/`existing-owned` cloud-init configures the production host, and under `existing-shared`/`existing-unmanaged` the host is already bootstrapped); `artifact-build`, `provision`, `app-deploy` are executed by CI and monitored |

Two rules hold on every path:

- **Single-slot runs** (`--slot=<slot>`) ignore the path's full-sequence defaults — that is how `tls` and `teardown` run on demand — but never its refusals.
- **No split CI runs:** the slots a pipeline performs run as one pipeline run, so every locally executed slot must come before the first CI-executed slot (or be `verify`, which runs after). A local binding that would land between CI-executed slots — e.g. a bound `host-bootstrap` between CI `provision` and CI `app-deploy` under `ci-push` — is refused with `SLOTS-BINDING-REFUSED`.

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

## Inspecting and Validating

```bash
smaqit slots plan --phase=deploy --mode=existing-k3s --env=test
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

- **Reachability.** A `public` machine exposes its Kubernetes API and the app's Ingress to the
  internet. A `private-mesh` machine is reachable only from peers of a private mesh network
  (WireGuard-based meshes are common): its API and app hostnames live in the mesh's private DNS
  zone, and CI reaches them only after joining the mesh with a per-app setup key.
- **Ingress mode.** `own-tls` is the original behaviour: the app's Ingress asks cert-manager for
  its own certificate. `platform-default-tls` means the platform already serves one certificate
  (often a DNS-01 wildcard) as its ingress controller's default, so the app's Ingress has no `tls`
  section. `platform-owned` means the platform runs the Ingress itself, for example behind its
  single sign-on gate, and the app ships none.

The default fillers read the row, not binding parameters: `mesh-check` confirms mesh access for
`private-mesh`; `ci-generate` renders one `deploy.yml` whose steps branch on the dispatched
machine's row (mesh join, Ingress, Certificate wait); `app-deploy` applies the Ingress the mode
calls for; and `verify` checks the app from inside the cluster with a short-lived probe Job
whenever the deployer can't reach it from outside (`private-mesh` or `platform-owned`). The probe's
verdict is the Job's own status, so it works even where the platform denies `pods/log`.

Three spec-wide Constraints rows apply to every environment:
- `Probe Client`: `curl`, `wget`, or `python3`, whichever the app's image already contains.
- `Container Registry`: `<host>, private` makes the generated workflow create the pull Secret.
- `Deployer Exec`: defaults to `denied`. The deploy then fails when the deployer could exec into
  pods, checked both by asking the API and by a real `kubectl exec` that must be refused. A
  platform that grants exec on purpose is recorded as `allowed`, which turns the failure into a
  warning.

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
