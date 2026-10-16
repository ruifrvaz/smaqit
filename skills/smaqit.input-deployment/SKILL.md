---
name: smaqit.input-deployment
description: Validate and elicit deployment phase parameters from session context before deployment begins. Invoke automatically before starting the Deployment phase to confirm or default execution preferences.
metadata:
  version: "1.8.0"
---

# Deployment Input

Validate session context for deployment execution preferences before the Deployment phase begins. All parameters have defaults — the phase can always proceed. Only elicit when a parameter is explicitly referenced but unspecified, or when ambiguity would cause incorrect execution.

## Steps

1. **Extract from session context** — Take any parameter the caller already passed as resolved and never ask about it again. Scan for any explicit deployment target overrides, verification preferences, or output requirements
2. **Apply defaults** — Use standard defaults for all unspecified parameters (listed below)
3. **Resolve blocking parameters** — A parameter is blocking only if it is referenced but missing, or if the deployment target cannot be resolved from Infrastructure specs. If you can ask the user directly, ask. Otherwise return `NEEDS-INPUT` (see below), listing every blocking parameter at once
4. **Resolve the slot plan** — Run `smaqit slots plan` with the resolved parameters (see Slot Plan below); skip when the caller already passed a slot plan
5. **Output the parameter block** — Emit the Phase Parameters block (see Output), including the slot plan. A caller passes it to the Deployment agent unchanged. A Deployment agent that applied this skill itself, because no block was passed, uses it directly

## Parameters

### Mode
**Default:** `autonomous`  
**Elicit when:** The request mentions reviewing or approving the Infrastructure specification before deployment, but does not make clear whether it should pause for review  
**Question:** "Should the Infrastructure specification pause for your review and approval before deployment continues (assisted), or run straight through (autonomous)?"

A request that clearly asks to review the specification resolves to `assisted` without asking. The PR merge gate under `deployment_path: existing-cicd-pr` is a human gate in either mode.

### Specification Mode
**Default:** `orchestrate`. The value is `prevalidated`, with exact `spec_paths`, only when the caller passes it (e.g. `smaqit.feature-new`)  
**Elicit when:** Never — the caller determines it

### Deployment Target
**Default:** Target environment as defined in Infrastructure specs  
**Elicit when:** Multiple deployment targets exist and the active target is ambiguous  
**Question:** "Any target-specific preferences? (e.g., deploy to staging first, use blue-green strategy)"

### Verification
**Default:** Run standard post-deployment health checks as defined in Coverage specs  
**Elicit when:** User has indicated they want to skip or extend verification  
**Question:** "How should deployment be verified? (run smoke tests, skip health checks, full suite)"

### Output Preferences
**Default:** Standard output; deployment steps and results shown  
**Elicit when:** Sensitive environment where log scrubbing is needed  
**Question:** "How should output be displayed? (verbose, scrub sensitive data from logs)"

### Provisioning Mode
**Default:** `provision`  
**Elicit when:** Session context mentions co-hosting, an existing/shared VM, a dedicated VM that will never be Terraform-managed, a Kubernetes/k3s cluster the app is onboarded onto, or names infrastructure another project already owns, and the mode cannot be resolved unambiguously from that context  
**Question:** "Is this a new VM, a redeploy of your own existing VM, a VM another project already manages, a dedicated VM that's provisioned out-of-band and will never be Terraform-managed, or a Namespace on a platform-owned k3s cluster that hands the app a kubeconfig?"

**Values:**
- `provision` — this project provisions its own new VM via Terraform. Unchanged, current behavior. Default whenever nothing in session context suggests otherwise.
- `existing-owned` — redeploying to a VM this project's own Terraform state already manages. Same default slot profile as `provision` except `teardown` (none); `terraform apply` is expected to no-op.
- `existing-shared` — targeting a VM a *different* project owns and manages via its own Terraform state (e.g., co-hosting a second app on a VM another project provisioned). Skips Terraform provisioning for this project entirely; downstream skills (`smaqit.infrastructure-vault-loader`, `smaqit.infrastructure-repo-config`, `smaqit.infrastructure-cicd-generate`) branch on this value — see `smaqit.new-greenfield-project` Phase 4/5.
- `existing-unmanaged` — a VM dedicated to this project (not co-hosted with anything else) but never managed by any Terraform state — provisioned out-of-band (manually, or via a provider's own tooling) and staying that way by design. Skips Terraform entirely, exactly like `existing-shared`, but is not a co-hosting scenario: the VM is this project's alone, just outside Terraform's view. The same downstream skills branch on this value — see `smaqit.new-greenfield-project` Phase 4/5 for the two points where it diverges from `existing-shared` (machine registration typically starts fresh rather than assuming an already-registered machine; the nginx vhost gets `default_server` since there's no co-hosting to guard against).
- `existing-k3s` — targeting an already-onboarded, Namespace-scoped Kubernetes cluster (k3s or equivalent) that hands the app a scoped kubeconfig and nothing else: no SSH, no host filesystem, no nginx, no Terraform on the app's side. Like `existing-shared`/`existing-unmanaged`, the target is platform-owned and Terraform provisioning is skipped entirely for this project — but the deploy *mechanism* is different (`kubectl apply` into the app's own Namespace via `smaqit.infrastructure-deploy-k3s-app`, not rsync/SSH/nginx). Every VM-specific step (Terraform, VM bootstrap, nginx vhost, `VM_HOST`) is skipped, replaced by kubeconfig-based deployment — see `smaqit.new-greenfield-project` Phase 4/5 and `smaqit.feature-new` for the exact callouts. Downstream skills branch on this value: `smaqit.infrastructure-vault-loader` (loads `secret/apps/<app-slug>/<machine-slug>/kubeconfig` instead of `ssh`, one machine per environment), `smaqit.infrastructure-cicd-generate` (`k3s` generation mode), and `smaqit.infrastructure-repo-config` (writes a `KUBECONFIG` Environment secret, sets no `VM_HOST`). **Resolving to this value means the Infrastructure spec must already carry the Constraints rows `Platform Repo`, `Registry File Path`, `Namespace Quota`, and `Container Registry`, and a complete `## Environments` row for the target environment — `Machine Slug`, `Host`, `Reachability` (`public` | `private-mesh`), `Ingress Mode` (`own-tls` | `platform-default-tls` | `platform-owned`), `Ingress Class`, `ClusterIssuer`, `Health Path`, the mesh fields and Constraints rows for `private-mesh`, and `Probe Client` when verification runs in-cluster — sourced from the platform team's own onboarding documentation, never invented locally.** The slot plan enforces this: `smaqit slots plan --mode=existing-k3s` refuses an incomplete row with `SLOTS-SPEC-*` lines, which are surfaced like any other `SLOTS-*` refusal. The fix belongs in the Infrastructure spec (through `smaqit.infrastructure`), never in a guessed value.
- **An app hosted on a platform repository's machine** — a repository that owns machines rather than an app — is not a separate mode. Use `existing-k3s` when the machine runs Kubernetes and admits tenants by onboarding (requested with `smaqit.infrastructure-request-k3s-onboarding`), or `existing-shared` when it is a VM the platform's Terraform owns. The platform repository provisions, baselines, and verifies the machine through its own fillers bound in its own `.smaqit/slots.yml`; the app project never binds or runs them, and takes the onboarding contract's fields from the platform's own documentation.

Output the resolved `provisioning_mode` in the parameter block (see Output) so the Deployment agent and later phases receive it without asking again.

### Environment
**Default:** The environment named by the caller (e.g. `dev` for a greenfield dev sweep, `prod` for production); otherwise the single target environment in Infrastructure specs
**Elicit when:** Infrastructure specs declare several environments and the caller named none
**Question:** "Which environment is this deployment for? (e.g. test, prod)"

### Deployment Path
**Default:** `standard`, or the `deployment_path` the caller passes (`existing-cicd-pr` from `smaqit.feature-new`, `ci-push` from `smaqit.new-greenfield-project` Phase 5)
**Elicit when:** Never — the calling skill or agent invocation determines it

## Slot Plan

Deployment runs a fixed sequence of slots (variation points). The project may bind slots to its own skills in `.smaqit/slots.yml`; unbound slots use the defaults for the resolved `provisioning_mode`. Resolution is deterministic and belongs to the CLI, not to judgment:

```
smaqit slots plan --phase=deploy --mode=<provisioning_mode> --env=<environment> --path=<deployment_path> --json
```

Add `--slot=<slot>` for a single-slot (day-two) run such as `teardown` or `replace`.

- **Exit 0:** put the JSON plan in the parameter block as the `slot_plan` the Deployment agent executes. Summarize it for the user as one line per slot (slot, source, filler, executor), and list any warnings — e.g. a family slot with no installed candidate, which the orchestrating skill must match or synthesize before deployment.
- **Nonzero exit:** stop. Surface every `SLOTS-*` line verbatim, whatever its code. Do not substitute a default wiring; the fix belongs in `.smaqit/slots.yml` or the bound skill.
- **Caller already passed a slot plan:** do not run the command again — one resolution per deployment.

## Readiness Condition

No required sections — defaults are always valid. Output the parameter block once the slot plan resolves, unless the deployment target cannot be determined from Infrastructure specs.

## Output

```
PHASE-PARAMETERS
phase: deploy
mode: autonomous
specification_mode: orchestrate
spec_paths: []
approved_spec_paths: []
provisioning_mode: provision
environment: <environment>
deployment_path: standard
slot_plan: <JSON output of smaqit slots plan>
verification: standard health checks
output: standard
```

The caller passes this block to the `smaqit.deployment` agent as its Execution Parameters. A caller adds `approved_spec_paths` and `answers` when it resumes or re-invokes the agent after a returned signal.

## NEEDS-INPUT Return

When step 3 cannot reach the user, end with this block instead of the Output block:

```
NEEDS-INPUT
source: smaqit.input-deployment
phase: deploy
written: nothing
questions:
  - id: <parameter name, e.g. provisioning_mode>
    question: "<the parameter's question above>"
    why: "<what makes it blocking>"
next: Ask the user these questions, then apply this skill again (or resume the agent) with the answers as `answers`, keyed by id.
```

A `SLOTS-*` failure is not a question for the user to answer: stop and surface it, as described under Slot Plan.
