---
status: In Progress
created: "2026-09-11"
mode: Assisted
started: "2026-09-12"
---

# k3s App-Deployment Skill With Routing and CI/CD

## Description

smaqit's Deployment phase can only deploy an application to a VM over rsync. Every deploy skill
in the product (`smaqit.infrastructure-deploy-rsync*`) assumes SSH access to a host, an nginx
vhost, an `__APP_DIR__` on disk, and, for Phase 5, a generated `deploy.yml` that rsyncs and
restarts. A downstream infrastructure repo now runs a hardened, live-verified k3s platform that
hands each onboarded application a **Namespace-scoped kubeconfig** and nothing else: no SSH, no
host filesystem, no nginx, no Terraform on the app's side. An application targeting that platform
has no deploy skill to route to, and greenfield's no-match path would synthesize a *new rsync*
skill pointed at rsync exemplars with nginx and `__APP_DIR__` baked in — exactly the wrong thing.

This task adds the missing **app-side** deployment capability, end to end, as one supported
product change:

1. A new deploy skill, `smaqit.infrastructure-deploy-k3s-app`: given a scoped kubeconfig for a
   target environment, lint the app's manifests against the platform's guardrails, apply them
   into the app's own Namespace, wait for rollout, stamp the Deployment with the commit SHA, and
   verify externally over the Ingress host. It never requests onboarding, never touches anything
   cluster-scoped, and never reads pod logs.
2. A new `provisioning_mode` value, `existing-k3s`, resolved by `smaqit.input-deployment` and
   branched on by `smaqit.new-greenfield-project` and `smaqit.feature-new` so the k3s family is
   selected and every VM-specific step (Terraform, VM bootstrap, nginx vhost, `VM_HOST`) is
   skipped — the same way `existing-shared` skips Terraform today.
3. A `k3s` generation mode in `smaqit.infrastructure-cicd-generate`, emitting a `deploy.yml`
   that runs the same apply-and-verify sequence from GitHub Actions using the kubeconfig held as
   a GitHub Environment secret, so Phase 5's PR-gated production deploy works against a k3s
   target.

This is the **app side** of the k3s golden path. The **platform side** — the registry file and
converge workflow that create the Namespace, RBAC, Pod Security, NetworkPolicy, quota, and issue
the kubeconfig — is task 116's separate skill, `smaqit.infrastructure-onboard-k3s-app`, so the
pair reads as onboard/deploy. Cluster provisioning
itself is a third, separate concern (task 115). This task depends on neither: it consumes an
already-onboarded Namespace and requires the platform contract only as declared parameters in
the application's own Infrastructure spec.

Everything platform-specific — Namespace name, ingress class, ClusterIssuer name, hostnames per
environment, kubeconfig secret name, the quota and Pod Security limits manifests must fit — is an
input read from the app's Infrastructure spec. Zero downstream project, repository, or machine
names appear in any product artifact this task produces.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Kubernetes (Deployment/Service/Ingress, Pod Security Admission `restricted`, ResourceQuota/LimitRange, NetworkPolicy, namespace-scoped RBAC), k3s, cert-manager (`Certificate`, `ClusterIssuer`), Traefik ingress class, kubectl, GitHub Actions (Environment secrets, `workflow_dispatch`)
**Platforms/Environments:** A k3s cluster that issues namespace-scoped kubeconfigs (the downstream platform's onboarded test Namespace for live verification; never a real application)
**Features/Integrations:** `smaqit.new-greenfield-project` Phase 4 Step 6 (task 087 dynamic stack detection / synthesis) and Phase 5; `smaqit.feature-new` provisioning-mode resolution; `smaqit.input-deployment`; `smaqit.infrastructure-cicd-generate` template variants; `smaqit.infrastructure-deploy-verify` (reused unchanged); `smaqit.infrastructure-repo-config`; installer payload (`installer/Makefile` compiles `skills-shared/` and `skills-claude/` from `skills/`)
**Versions/Constraints:** kubectl must match the cluster's minor version within the supported skew; the scoped credential is long-lived and grants only the onboarding contract's RBAC allowlist (no `pods/log`, no `pods/exec`, no cluster scope) — verification must not depend on reading logs

## Design Decisions

- **Name and pairing:** `smaqit.infrastructure-deploy-k3s-app` (app side) paired with task 116's
  `smaqit.infrastructure-onboard-k3s-app` (platform side). Named for the platform vocabulary used
  everywhere else in this product (`STK-K3S-PLATFORM`, task 116), even though the skill's content
  is plain Kubernetes and deliberately carries no k3s-only assumption, so it also works against
  any cluster that issues scoped kubeconfigs.
- **One task, definition through compiled skill, routing, CI/CD, payload, and docs** — the
  task-106 "reconcile into product" shape, not task 116's definition-only shape. The
  definition-first split exists so a downstream project can contribute before the framework
  owner validates; here the framework owner is authoring directly, so the intermediate handoff
  has no audience, and a definition alone yields nothing runnable.
- **Routing signal is a new `provisioning_mode` value, `existing-k3s`**, not a new orthogonal
  `deployment_target` field. `provisioning_mode` already answers "who owns the target"; a k3s
  Namespace is always platform-owned, like `existing-shared`, with a different deploy mechanism.
  One new value; every existing callout site branches on it. No Infrastructure spec template
  change: the mode is elicited by `smaqit.input-deployment`, and the target values live in the
  spec's existing Compute, Networking, Secrets, and Constraints sections.
- **Step 6 becomes family-aware.** The mode selects the skill *family*; the declared stack is
  matched *within* that family; no-match synthesis uses that family's own exemplars and inherited
  context. The rsync family's four required-inherited-context items are VM-specific and must not
  leak into the k3s family. The k3s family's own four, inherited verbatim by any synthesized
  member:
  1. `__APP_SLUG__` token — the Namespace is always `app-__APP_SLUG__`, never a free-form name.
  2. A shared Ingress template (bundled asset) that sets `ingressClassName` and the cert-manager
     `cluster-issuer` annotation from spec values — never a hand-written Ingress per app.
  3. The deploy stamp is delivered as pod environment (`DEPLOY_SHA`, `DEPLOY_TIME`) set at apply
     time, preserving the existing health-endpoint contract (`{sha, deployedAt}`) so
     `smaqit.infrastructure-deploy-verify` is reused **unchanged** with `--url https://<host>`.
  4. A preflight guard (bundled `namespace-guard.sh`) that reads the kubeconfig's context, runs
     `kubectl auth can-i --list` in the target Namespace, and refuses to proceed if any
     cluster-scoped verb is permitted or the Namespace doesn't match `app-__APP_SLUG__` —
     the k3s analog of `plan-guard.sh`/`ownership-guard.sh`, guarding against ever running with
     a broader credential than the contract issues.
- **Manifests are linted before apply, not debugged after.** A bundled `manifest-lint.sh` checks
  every Pod template for the `restricted` Pod Security fields (`runAsNonRoot`,
  `allowPrivilegeEscalation: false`, `capabilities.drop: [ALL]`, `seccompProfile:
  RuntimeDefault`, no `privileged`/`hostPath`/host namespaces) and for explicit resource
  requests/limits that fit the declared quota, and for the absence of any cluster-scoped kind.
  A rejected manifest fails fast with the offending field named, because the platform's
  admission rejection after the fact is terse and the app's credential cannot read logs to
  diagnose it.
- **Verification is external.** `kubectl rollout status` for the Deployment, the `Certificate`
  object's `Ready` condition (in the RBAC allowlist), then `smaqit.infrastructure-deploy-verify`
  over HTTPS on the Ingress host. `deploy-verify`'s own scope note that it does not verify TLS
  validity stands; the `Certificate` Ready check covers that gap here.
- **Private image pull is a declared input, not an assumption.** The cluster must pull the app
  image; a private registry needs an `imagePullSecret` in the Namespace. The onboarding contract
  grants `secrets: create`, so the skill creates it from a declared registry credential
  (`REGISTRY_USERNAME`/`REGISTRY_TOKEN` or equivalent) when the spec names a private registry, and
  skips it for a public image. The credential source is the app's own GitHub Environment secret,
  never anything platform-owned.
- **Two environments, one skill.** The dev sweep (greenfield Phase 4) deploys into the *test*
  Namespace; Phase 5 CI/CD deploys into the *prod* Namespace. The skill takes an environment
  argument and reads that environment's kubeconfig secret reference and ingress host from the
  spec. Both Namespaces must already be onboarded — the skill's pre-condition, not its job.
- **Kubeconfig becomes a new Vault-managed app credential; the Vault spec was stale.**
  `smaqit.infrastructure-vault-loader`'s namespace convention currently defines only `ssh`,
  `github`, and `machine` under `secret/apps/<app-slug>/*` — it predates this platform's credential
  type and has no concept of a kubeconfig, even though `smaqit.feature-new`'s Vault-loading step
  needs to load one per environment for `existing-k3s`. This task extends the convention rather
  than routing around it: `secret/apps/<app-slug>/kubeconfig` holds two fields, `test` and `prod`
  (mirroring how `ssh` already holds `private_key`/`public_key` as two fields on one path), since
  the platform issues a distinct scoped kubeconfig per Namespace/environment. The value is always
  received out-of-band from the platform's own onboarding hand-off (a workflow artifact or its
  external secrets store, per task 116) and pasted in — `load-credentials.sh` never talks to a
  cluster to obtain or validate it. Rotation follows the same out-of-band shape: the platform's
  credential rotation is destructive (task 116's known gap), so `rotate-credential.sh`'s
  `apps/<app-slug>/kubeconfig` support re-prompts for a freshly platform-reissued value rather than
  generating one locally, unlike `ssh`'s local keypair regeneration.
- **`cicd-generate` gets a `k3s` mode via a new template variant**, honoring its own rule that
  more variance means a new variant, not a templating engine. `deploy.yml.k3s.template`: a single
  `deploy` job that checks out, sets `KUBECONFIG` from the environment secret, runs the same
  guard → lint → apply → rollout → verify sequence, and never touches Terraform or SSH.
  `post-merge-deploy.yml.template` is reused unchanged; no `provision.yml`; nothing vendored.
- **Never request onboarding, never read logs, never go cluster-scoped.** If the kubeconfig
  secret for the target environment is absent, fail loudly with a pointer to the platform's
  onboarding playbook. The direction stays apps → infrastructure.
- **Product artifacts are anonymized** exactly as task 116 requires: no real downstream project,
  repository, or machine name in the definition, the compiled skill, templates, docs, or tests.

## Implementation Steps

1. Read `.smaqit/definitions/skills/smaqit.infrastructure-deploy-rsync-python-tornado.md` (definition
   shape), `skills/smaqit.infrastructure-deploy-rsync/SKILL.md` (compiled shape and bundled
   `scripts/` layout), `.smaqit/tasks/087_*.md` and `106_*.md` (routing and reconcile precedents),
   and `skills/smaqit.new-greenfield-project/SKILL.md` Phase 4 Steps 3–6, Phase 5, and its Gotchas.
2. Author `.smaqit/definitions/skills/smaqit.infrastructure-deploy-k3s-app.md` mirroring the tornado
   definition's sections (Description, Provenance [anonymized], Required-inherited-context [the
   four k3s-family items above], Steps [Pre-conditions + Steps], Output, Scope, Completion,
   Failure Handling, Gotchas, Allowed Tools, Examples).
3. Compile `skills/smaqit.infrastructure-deploy-k3s-app/`: `SKILL.md` plus `scripts/namespace-guard.sh`,
   `scripts/manifest-lint.sh`, `assets/ingress.yaml.template`, and a reference
   `assets/deployment.yaml.template` showing a `restricted`-compliant Pod template with the stamp
   environment variables. Steps: resolve environment → load kubeconfig → guard → optional
   `imagePullSecret` → lint → `kubectl apply -n app-__APP_SLUG__` → `rollout status` →
   `Certificate` Ready → `smaqit.infrastructure-deploy-verify --url https://<host>
   --expected-sha <sha>`.
4. `skills/smaqit.input-deployment/SKILL.md`: add `existing-k3s` to the `provisioning_mode` list with
   the same shape as the `existing-unmanaged` entry; document that downstream skills branch on it.
5. `skills/smaqit.new-greenfield-project/SKILL.md`: add `existing-k3s` to the Provisioning Mode
   section; add `→ existing-k3s:` callouts to Phase 4 Step 3 (invoke `cicd-generate` in `k3s`
   mode), Step 4 (skip Terraform), Step 5 (skip VM bootstrap), Step 6 (family selection: match the
   declared stack within `smaqit.infrastructure-deploy-k3s-app*`; no-match synthesis points at the
   k3s exemplar and passes the k3s family's four inherited items), and Phase 5 Step 4 (`deploy.yml`
   has a single kubectl-driven `deploy` job); add the k3s family's inherited-context list to
   Gotchas beside the rsync family's; bump `metadata.version`.
6. `skills/smaqit.infrastructure-vault-loader/`: fix the stale namespace convention. Add
   `secret/apps/<app-slug>/kubeconfig` (fields `test`, `prod`) to `SKILL.md`'s namespace-convention
   table and Gotchas; add an `existing-k3s` branch to `scripts/load-credentials.sh` that prompts
   for and stores both per-environment kubeconfigs verbatim (out-of-band paste — never a cluster
   call); add `apps/<app-slug>/kubeconfig` support to `scripts/rotate-credential.sh` (re-prompts
   for a freshly platform-reissued value, since kubeconfig rotation is destructive and
   platform-side per task 116, not locally regenerated like `ssh`).
7. `skills/smaqit.feature-new/SKILL.md`: extend provisioning-mode resolution (Steps 5–6) and the
   per-mode Vault/repo-config callouts: for `existing-k3s`, load only the app's `github` secret
   plus the now-real `secret/apps/<app-slug>/kubeconfig` (`test`/`prod`) via step 6's new
   vault-loader support; `smaqit.infrastructure-repo-config` writes the environment's kubeconfig
   as a `KUBECONFIG` secret on each GitHub Environment and sets no `VM_HOST`.
8. `skills/smaqit.infrastructure-cicd-generate/`: add `k3s` mode, `assets/deploy.yml.k3s.template`,
   the `__APP_SLUG__` substitution, and Output/Scope/Completion/Gotchas entries for the new mode.
9. `skills/smaqit.infrastructure-repo-config/SKILL.md`: document the `KUBECONFIG` environment secret
   and the optional registry credential for `existing-k3s` targets.
10. Build the installer (`installer/Makefile`) and confirm the new skill lands in both
    `installer/skills-shared/` and `installer/skills-claude/`; run the existing Go test suite.
11. Live-verify against the downstream platform's onboarded **test** Namespace with a throwaway
    slug, never a real application: positive path end to end (guard passes with the scoped
    kubeconfig, lint passes, apply, rollout complete, `Certificate` Ready, `deploy-verify` PASS on
    health, SHA, and SPA root over HTTPS); negative paths (`namespace-guard.sh` refuses a
    cluster-admin kubeconfig; `manifest-lint.sh` rejects a `privileged: true` Pod and an
    over-quota request before any apply); private-registry pull via a created `imagePullSecret`;
    then run the generated `deploy.yml` once from GitHub Actions against the same test Namespace
    using the `KUBECONFIG` environment secret and confirm the identical verify result. Tear down
    by removing the throwaway slug through the platform's own offboarding path.
12. Update `CHANGELOG.md` and user-facing docs for the new supported deployment target and the
    new provisioning mode, including the refreshed Vault namespace convention.
13. Confirm task 116's skill is still named `smaqit.infrastructure-onboard-k3s-app` (that task
    adopted the name itself when it started) so the onboard/deploy pair cannot collide; this task
    makes no change to task 116.

## Known Issues Triage
**Triaged:** 2026-09-12
**Tools searched:** Kubernetes, kubectl, k3s, cert-manager, Traefik, GitHub Actions
**Result:** Advisory

### Blocking Issues
_None._

### Advisory Issues
- [#13791 Kubernetes Ingress: missing TLS secret aborts loading of remaining spec.tls[] entries](https://github.com/traefik/traefik/issues/13791) — `traefik/traefik` — opened 2026-08-27 — `area/provider/k8s/ingress`, `kind/bug/possible` — confirmed only for an Ingress with multiple `spec.tls[]` entries sharing one manifest (a common cert-manager multi-domain pattern per the report); this task's shared Ingress template is single-host-per-app, so it does not currently hit this path — worth re-checking if the template ever grows multi-host support.
- [#137774 Ineffective resource name restrictions in role](https://github.com/kubernetes/kubernetes/issues/137774) — `kubernetes/kubernetes` — opened 2026-03-16 — `kind/bug`, `sig/api-machinery` — an etcd-exhaustion DoS via unbounded Role/ClusterRole name length, but the exploit path requires permission to create Role/ClusterRole objects. This task's `namespace-guard.sh` design already refuses any cluster-scoped verb and the platform (task 116) is the only actor that ever creates RBAC objects, so this skill's own credential cannot reach the vulnerable path — informational only.

### Historical (Closed)
- [#134929 ServiceAccount with RoleBinding only can delete its namespace](https://github.com/kubernetes/kubernetes/issues/134929) — `kubernetes/kubernetes` — closed 2026-03-17 — directly relevant precedent for this task's least-privilege RBAC assumption; worth a sanity check during live verification that the onboarded ServiceAccount's RoleBinding cannot delete its own Namespace on the target cluster's Kubernetes version.
- [#12506 ACME HTTP-01 challenge returns 404 with IngressClass traefik and cert-manager integration](https://github.com/traefik/traefik/issues/12506) — `traefik/traefik` — closed 2026-01-15 — known integration gotcha for exactly this task's stack (Traefik ingress class + cert-manager ACME); the closed issue's resolution is worth reviewing if `Certificate` never reaches `Ready` during live verification.

### Unresolvable Tools
_None — all named tools resolved to a repository._

### Omitted Tools
_None — 5 resolved repositories (`kubernetes/kubernetes` also covers `kubectl`), within the 5-repository limit._

### Search Warnings
_None._

### Notes
- `cert-manager/cert-manager` and `k3s-io/k3s` searches returned results, but none confirmed a relevant match to this task's platform/feature dimensions (ClusterIssuer/Certificate-Ready reliability, or namespace-scoped-kubeconfig behavior) even loosely — omitted as noise rather than listed.
- `GitHub Actions` resolved via the deterministic helper to `actions/starter-workflows`, which returned no results for `workflow_dispatch`/environment-protection terms. This is expected: GitHub's `workflow_dispatch`-must-exist-on-default-branch and per-environment-name branch-policy behaviors (already documented in this task's own Design Decisions and in task 116) are platform behavior, not something tracked in that repo's issue tracker — recorded as a categorization limitation, not a Clear result for that dimension.

## Acceptance Criteria

- [ ] `.smaqit/definitions/skills/smaqit.infrastructure-deploy-k3s-app.md` exists, mirroring the tornado definition's section structure, with an anonymized Provenance
- [ ] `skills/smaqit.infrastructure-deploy-k3s-app/` is compiled with `SKILL.md`, `scripts/namespace-guard.sh`, `scripts/manifest-lint.sh`, and the Ingress and Deployment templates; the installer build places it in both `installer/skills-shared/` and `installer/skills-claude/`, and the existing Go tests pass
- [ ] `smaqit.input-deployment` accepts `provisioning_mode: existing-k3s`; `smaqit.new-greenfield-project` and `smaqit.feature-new` carry `→ existing-k3s:` callouts at every step that differs, and Step 6 selects the k3s family (stack matched within it; synthesis uses the k3s exemplar and the k3s family's four inherited items, never the rsync family's)
- [ ] `smaqit.infrastructure-vault-loader`'s namespace convention is no longer stale: `secret/apps/<app-slug>/kubeconfig` (`test`/`prod` fields) is documented in `SKILL.md`, `load-credentials.sh` has a working `existing-k3s` branch that loads/stores it, `rotate-credential.sh` supports re-prompting for a reissued value at that path, and `smaqit.feature-new`'s `existing-k3s` Vault step actually loads a credential that exists (no dangling reference)
- [ ] `smaqit.infrastructure-cicd-generate` emits, in `k3s` mode, `deploy.yml` (single kubectl-driven job from a `KUBECONFIG` Environment secret) and `post-merge-deploy.yml` only — no `provision.yml`, nothing vendored, no SSH or Terraform reference
- [ ] Live, against an onboarded test Namespace with a throwaway slug: `namespace-guard.sh` passes with the scoped kubeconfig; `manifest-lint.sh` passes; apply and `rollout status` succeed; the `Certificate` reaches `Ready`; `smaqit.infrastructure-deploy-verify --url https://<host> --expected-sha <sha>` reports PASS on health, SHA, and SPA root — with `deploy-verify` itself unmodified
- [ ] Live negative cases: `namespace-guard.sh` refuses a cluster-admin kubeconfig; `manifest-lint.sh` rejects a `privileged: true` Pod template and an over-quota resource request, each before any `kubectl apply` runs
- [ ] A private-registry image is pulled successfully via an `imagePullSecret` the skill created from a declared registry credential
- [ ] The generated `deploy.yml` runs end to end from GitHub Actions against the test Namespace using the `KUBECONFIG` Environment secret and reaches the same verify result as the local sweep
- [ ] At no point does the skill or generated workflow read pod logs, apply a cluster-scoped object, or request onboarding
- [ ] The throwaway slug is offboarded through the platform's own path afterward; nothing is left in the cluster
- [ ] `CHANGELOG.md` and user docs describe the new target and mode; no product artifact names a real downstream project, repository, or machine

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
| `.smaqit/definitions/skills/smaqit.infrastructure-deploy-k3s-app.md` | Create |
| `skills/smaqit.infrastructure-deploy-k3s-app/SKILL.md` | Create |
| `skills/smaqit.infrastructure-deploy-k3s-app/scripts/namespace-guard.sh` | Create |
| `skills/smaqit.infrastructure-deploy-k3s-app/scripts/manifest-lint.sh` | Create |
| `skills/smaqit.infrastructure-deploy-k3s-app/assets/ingress.yaml.template` | Create |
| `skills/smaqit.infrastructure-deploy-k3s-app/assets/deployment.yaml.template` | Create |
| `skills/smaqit.input-deployment/SKILL.md` | Modify (add `existing-k3s`) |
| `skills/smaqit.infrastructure-vault-loader/SKILL.md` | Modify (stale namespace convention: add `kubeconfig` field) |
| `skills/smaqit.infrastructure-vault-loader/scripts/load-credentials.sh` | Modify (`existing-k3s` branch: load/store per-environment kubeconfig) |
| `skills/smaqit.infrastructure-vault-loader/scripts/rotate-credential.sh` | Modify (support `apps/<app-slug>/kubeconfig`) |
| `skills/smaqit.new-greenfield-project/SKILL.md` | Modify (mode, Phase 4/5 callouts, family-aware Step 6, Gotchas, version bump) |
| `skills/smaqit.feature-new/SKILL.md` | Modify (mode resolution, Vault/repo-config callouts) |
| `skills/smaqit.infrastructure-cicd-generate/SKILL.md` | Modify (`k3s` mode) |
| `skills/smaqit.infrastructure-cicd-generate/assets/deploy.yml.k3s.template` | Create |
| `skills/smaqit.infrastructure-repo-config/SKILL.md` | Modify (`KUBECONFIG` secret, registry credential) |
| `installer/` payload | Verify (compiled from `skills/` by the Makefile; no manual list to edit) |
| `CHANGELOG.md`, user docs | Modify |

## Notes

- Origin: planned 2026-09-11 with the downstream platform's owner after establishing that task
  116 covers only the *platform side* (onboarding) and that nothing in the product deploys an
  application into a scoped Kubernetes Namespace. The decisions above (one task; the name;
  `existing-k3s` as the routing signal; CI/CD generation in scope) were each confirmed
  explicitly during planning.
- This task supersedes the downstream repo's abandoned "handoff document" task: rather than staging
  a requirements document for later transfer, the capability is authored directly here.
- Task 116 adopted the name `smaqit.infrastructure-onboard-k3s-app` on its own when it started
  (2026-09-12), so the platform-side/app-side pair is already legible without any change from this
  task. This task does not touch task 116 and does not depend on it.
- Task 115 still names "Tenant Reconcile", the pre-k3s tenancy model the downstream platform has
  since retired entirely; it likely needs re-scoping against task 116 before it starts. Flagged,
  not acted on here.
- Two known platform gaps are inherited as constraints, not fixed here: the scoped credential
  cannot read `pods/log` (verification is designed external for this reason), and credential
  rotation is destructive (irrelevant to this skill, which only consumes the credential).
- Vault kubeconfig scope was added 2026-09-12 during pre-start assessment: `smaqit.feature-new`'s
  planned `existing-k3s` Vault step referenced a `kubeconfig` credential that
  `smaqit.infrastructure-vault-loader`'s namespace convention had no concept of — a genuinely
  stale spec, not a new decision — so fixing it was folded into this task rather than split out,
  per explicit user confirmation.
