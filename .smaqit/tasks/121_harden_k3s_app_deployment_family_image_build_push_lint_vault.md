---
status: PR Open
pr: 91
created: "2026-09-15"
mode: Assisted
started: "2026-09-15"
---

# Harden the k3s App-Deployment Skill Family: Image Build-Push, Lint, and Vault Conventions

## Description

Found live while onboarding the first real `existing-k3s` app end-to-end (a downstream
project's task 005, containerizing and deploying a static site onto its onboarded
Namespace on a real cluster). The onboarding-request and deploy-mechanics skills
themselves (tasks 116-120) are sound — every gap found here sits either upstream of them
(turning source into a pushed image) or in an underspecified convention they silently
assumed already existed (a shared pull credential, a required-before-first-run
precondition). Seven findings, six live-verified defects/gaps plus one design
recommendation, all discovered in a single review pass across ~10 PRs and six
live-diagnosed failures (`ImagePullBackOff`, `CreateContainerConfigError`, a rejected
Docker tag, a 404'd `workflow_dispatch`, a `manifest-lint.sh` `FileNotFoundError`).

Full non-secret evidence: the downstream project's own task file
(`.smaqit/tasks/005_phase4_dev_environment_sweep.md`, its Notes section) and its
synthesized, project-local skill (`smaqit.infrastructure-image-build-push-static`,
`.claude/skills/` in that repo) — read both in full before starting; this task's job is
to generalize and reconcile that proven pattern into canonical smaqit, not re-derive it.

### Finding 1 — No skill anywhere authors a Dockerfile, an image build/push CI step, or a baked health-endpoint convention

`smaqit.infrastructure-cicd-generate`'s `k3s` mode assumes a real, already-pushed image
reference already sits in the checked-in `deployment.yaml` — confirmed by reading its own
`assets/deploy.yml.k3s.template` in full, which goes straight from checkout to
`kubectl apply` with no build step. `smaqit.infrastructure-deploy-k3s-app` likewise
assumes "the app's Deployment/Service manifests present locally" as a precondition,
never producing them. Nothing in the whole framework authors a Dockerfile, builds an
image, pushes it to a registry, or establishes a `/health` endpoint convention. Confirmed
by exhaustive grep across the entire installation for `Dockerfile`, `docker build`,
`docker push`, `containeriz*`, `ghcr`, `docker hub` — zero matches outside unrelated
VM/Compose deploy families.

### Finding 2 — Docker image tags must be lowercase; `github.repository` isn't

The downstream project's first `workflow_dispatch` run failed at the build step:
`invalid tag "Magnificah/magnificah-website:...": repository name must be lowercase`.
`${{ github.repository }}` preserves the real org/repo casing; Docker rejects a
mixed-case reference outright. A one-line `tr '[:upper:]' '[:lower:]'` fix — but only if
whatever generates the build step remembers to include it. This is Finding 1's own
implementation detail, not a separate skill, but is called out on its own since it must
be baked into the generated script itself (mechanized), never left as a Gotcha note a
future agent has to remember to apply by hand.

### Finding 3 — `manifest-lint.sh` doesn't catch the single most common way `runAsNonRoot: true` fails at admission time

The downstream project's pod sat in `CreateContainerConfigError`: *"container has
runAsNonRoot and image has non-numeric user (nginx), cannot verify user is non-root."*
Any base image whose own `USER` directive names a user (not a numeric UID — true of
`nginxinc/nginx-unprivileged` and plausibly many other minimal images) triggers this the
same way, and `docker run` locally never surfaces it — it is a Kubernetes-only
admission-time check. `manifest-lint.sh`'s own stated design philosophy is "rejects a
manifest set BEFORE any `kubectl apply` runs... a fast, explicit local rejection naming
the offending field" — this exact failure mode is a natural, checkable extension of that
same philosophy: `runAsNonRoot: true` set (pod- or container-level) with no explicit
numeric `runAsUser` at either level is a deterministic future failure, not a maybe.

### Finding 4 — No documented Vault convention for a shared, cross-app GHCR pull credential

GHCR packages pushed from a private repository default to private visibility with no
opt-in step. `smaqit.infrastructure-deploy-k3s-app`'s own template already has a
commented-out `imagePullSecret` step (Step 4) referencing `REGISTRY_USERNAME`/
`REGISTRY_TOKEN`, but nothing in `smaqit.infrastructure-vault-loader` or
`smaqit.infrastructure-repo-config` documents where that credential should actually live.
The downstream project worked this out live: a **shared**, read-only pull credential
(`secret/organizations/<org-slug>/github-package-read`, a classic PAT scoped to exactly
`read:packages` — fine-grained PATs have unreliable Packages API support) reused across
every app on every machine in the org, never minted per app. This convention doesn't
exist anywhere in smaqit's own Vault path documentation today.

### Finding 5 — `manifest-lint.sh` errors on `deployment/k3s/ingress.yaml` are correct but silent about *why*

The Ingress is deliberately rendered once, late (per `smaqit.infrastructure-deploy-k3s-app`
Step 5), not checked in from the start — correct, since it needs a real, DNS-resolving
hostname that shouldn't be guessed. But nothing in `smaqit.infrastructure-cicd-generate`'s
own docs says a fresh `k3s`-mode generation's `deploy.yml` **will** fail
`manifest-lint.sh` with a bare `FileNotFoundError` until that render has happened once —
an agent hitting this cold has no signal that it's expected, not a bug in the generated
template.

### Finding 6 — `workflow_dispatch` cannot target a workflow that only exists on a non-default branch

GitHub's own API 404s. Not something smaqit's generated files can prevent structurally,
but nothing in `smaqit.infrastructure-cicd-generate` or `smaqit.new-greenfield-project`'s
k3s Phase 4/5 sequence text warns that a first-time test dispatch may need the workflow
landed via a small, separate, earlier PR before the task's own PR merges.

### Finding 7 — The k3s family's own docs frame the first deploy as a manual, local, interactive step

`smaqit.infrastructure-deploy-k3s-app`'s own Examples section describes Phase 4's dev
sweep as: load the kubeconfig locally, run guard/lint/apply directly. This is worse than
just a style preference — every one of Findings 2/3/4 was actually *caught* only because
the downstream project's user insisted every deploy, including the first, go through the
real CI workflow (`workflow_dispatch`) instead. A manual local first deploy exercises a
different code path (the agent's own shell commands) than what will run in production
(the generated workflow) and would have hidden at least the lowercase-tag bug (a local
shell wouldn't have hit `github.repository`'s casing at all). Recommend reframing the
k3s family's default recommended path to always go through the generated workflow, local
invocation kept only as a documented fallback for environments with no CI at all.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Docker, Kubernetes (k3s), GitHub Container Registry, GitHub Actions, HashiCorp Vault
**Platforms/Environments:** None
**Features/Integrations:** k3s app deployment, container image build/push, Pod Security admission, Vault credential conventions
**Versions/Constraints:** None

## Design Decisions

- **Finding 1 is the anchor; Findings 2-4 are its implementation details, not separate
  skills.** Reconcile the downstream project's proven, project-local skill
  (`smaqit.infrastructure-image-build-push-static`) into canonical smaqit, generalizing
  its name and scope only as far as evidence supports — it was scoped to "a static site
  with no build step of its own"; broadening to cover a stack with its own build step
  (a compiled backend image, say) is explicitly out of scope here without a second
  real example to generalize from. Name it `smaqit.infrastructure-image-build-push`
  only if the static-only scoping is dropped in this task; otherwise keep
  `-static` and note the generalization as this task's own Follow-up.
- **Mechanize everything mechanizable; only fall back to a Gotcha/doc note where the
  fix genuinely cannot be baked into a script, template, or check** (Findings 2, 3 are
  mechanized; 5, 6 are documentation, since GitHub's own platform behavior and a
  deliberate late-render design aren't things a template can prevent).
- **Finding 3's lint addition must warn/fail on `runAsNonRoot` without `runAsUser`
  looking at both pod- and container-level `securityContext`, respecting the same
  inheritance semantics `manifest-lint.sh` already implements for its other checks**
  (container-level overrides pod-level; check the effective value, not just presence at
  one level).
- **Finding 4's Vault convention is additive, not a rename.** Document
  `organizations/<org-slug>/github-package-read` as a new, named path in
  `smaqit.infrastructure-vault-loader`'s existing path table (alongside `platform-repo`
  and `<machine-slug>/kubeconfig`), and wire `smaqit.infrastructure-repo-config`'s
  `existing-k3s` branch to sync it to `REGISTRY_USERNAME`/`REGISTRY_TOKEN` per
  environment when the Infrastructure spec declares a private registry — mirroring how
  `KUBECONFIG` is already synced per-environment there.
- **Finding 7 changes a recommended default, not a hard requirement** — keep local
  invocation documented as a valid fallback (e.g., no CI configured yet), but reorder
  the skill's own Examples/Steps so `workflow_dispatch` is presented first.
- **Reconciliation for Finding 1 goes through `smaqit.create-skill`, not hand-authored
  files.** It writes `.smaqit/definitions/skills/smaqit.infrastructure-image-build-push-static.md`
  and invokes `smaqit.L2` to compile `skills/smaqit.infrastructure-image-build-push-static/SKILL.md`.
  Since `smaqit.create-skill` infers from a name + context, an explicit fidelity check
  against the downstream source is required afterward (base image choice, the `USER
  root`→`USER nginx` ordering for baking `health.json`, the `runAsUser: 101` fact), plus
  a manual correction pass on the Provenance frontmatter (`synthesized: true`,
  `contributed-for-project: magnificah-website`, `contributed-date: 2026-09-14`) — the
  compiler cannot infer those from the skill name alone.
- **Findings 1/2's CI wiring lives in `assets/deploy.yml.k3s.template`, not in the
  reconciled skill.** Confirmed live: `smaqit.infrastructure-cicd-generate` is
  stack-agnostic and `assets/deployment.yaml.template` already carries an `__IMAGE__`
  token with nothing that ever substitutes it — the generated `deploy.yml.k3s.template`
  stamps `__DEPLOY_SHA__`/`__DEPLOY_TIME__` but has no `build` job and no `__IMAGE__`
  step at all. Adding the build-and-push job (with the lowercase-tag fix) to the
  template itself, not the skill, means every future `k3s`-mode project gets it for
  free; the reconciled skill narrows to what still needs stack judgment (Dockerfile
  authoring, health-endpoint baking, the manifest's `runAsUser`).
- **New scope, found during planning: the Infrastructure spec template has no field
  today for "this spec declares a private registry."** `smaqit.infrastructure-repo-config`
  Step 5 already says to sync a registry credential "if the spec names a private
  registry," but nothing makes that condition checkable. Add a conditional Constraints
  row (`Container Registry`, `existing-k3s` only, combining registry host + visibility,
  e.g. `ghcr.io, private`) to `templates/specs/infrastructure.template.md` plus a MUST
  rule in `agents/infrastructure.md`, mirroring task 120's precedent for other
  existing-k3s fields — one fact per row, no inference.
- **No new test harness invented for `manifest-lint.sh`.** No skill-bundled script in
  this repo has automated tests today; verify the new `runAsUser` check with throwaway
  fixture manifests (pod-level-only, container-level-only, missing entirely, and the
  pre-existing passing case) rather than introducing a test framework for one script.

## Implementation Steps

**Phase A — Reconcile the skill (Finding 1), via `smaqit.create-skill`**
1. Invoke `smaqit.create-skill` for `smaqit.infrastructure-image-build-push-static`,
   feeding it the proven downstream content (Dockerfile on
   `nginxinc/nginx-unprivileged:1-alpine`, the `USER root`→`USER nginx` ordering for
   baking `health.json`, `deployment.yaml`/`service.yaml` with explicit
   `runAsUser: 101`, Gotchas/Failure Handling) as the specification input — excluding
   the CI-wiring steps that move to Phase B.
2. Fidelity check the compiled SKILL.md against the downstream source (see Design
   Decisions), and manually correct the Provenance frontmatter fields the compiler
   cannot infer.

**Phase B — Mechanize build-and-push into the template (Findings 1, 2)**
3. Add a `build` job to `skills/smaqit.infrastructure-cicd-generate/assets/deploy.yml.k3s.template`,
   preceding `deploy` (`needs: build`): `docker/login-action@v3` against `ghcr.io`
   (`GITHUB_TOKEN`, `permissions: packages: write`), lowercase the repo
   (`tr '[:upper:]' '[:lower:]'`) before tagging by `github.sha`, output the image
   reference.
4. Extend the existing "Stamp manifest" step to also substitute `__IMAGE__` in
   `deployment.yaml` using the `build` job's output.
5. Update `skills/smaqit.infrastructure-cicd-generate/SKILL.md`'s Steps/Output/Gotchas
   to document the `build` job as always-generated for `k3s` mode; bump
   `metadata.version`.

**Phase C — `runAsUser` lint check (Finding 3)**
6. Extend `check_container_security_context` in
   `skills/smaqit.infrastructure-deploy-k3s-app/scripts/manifest-lint.sh`: mirror the
   existing `runAsNonRoot`/`seccompProfile` effective-value pattern — when effective
   `runAsNonRoot` is `true` and no effective numeric `runAsUser` is set at either level,
   fail with a named-field message.
7. Verify manually with throwaway fixture manifests (see Design Decisions); record the
   verification transcript under this task's own Findings.

**Phase D — Shared registry-credential convention (Finding 4)**
8. Add the `Container Registry` Constraints row to `templates/specs/infrastructure.template.md`
   and a MUST rule to `agents/infrastructure.md` (see Design Decisions — new scope).
9. Add `secret/organizations/<org-slug>/github-package-read` to the Vault path table in
   `skills/smaqit.infrastructure-vault-loader/SKILL.md` (a new `organizations/`
   namespace, sibling to `apps/`/`machines/`, populated manually — org-scoped, not
   per-app, so it does not go through `load-credentials.sh`'s per-app prompt flow).
10. Extend `skills/smaqit.infrastructure-repo-config/SKILL.md` Step 5 to sync
    `REGISTRY_USERNAME`/`REGISTRY_TOKEN` per environment from that specific Vault path
    when the new Constraints row declares a private registry.

**Phase E — Documentation (Findings 5, 6, 7)**
11. Add two Gotchas to `smaqit.infrastructure-cicd-generate/SKILL.md` (Ingress-must-
    render-first; `workflow_dispatch` default-branch requirement); cross-reference both
    from `smaqit.new-greenfield-project/SKILL.md`'s k3s Phase 4/5 text.
12. Reorder `smaqit.infrastructure-deploy-k3s-app/SKILL.md`'s Examples so the generated-
    `deploy.yml`/`workflow_dispatch` example comes first, local invocation second as a
    documented fallback.

**Phase F — Verify**
13. Rebuild and reinstall (`cd installer && make build && ./dist/smaqit-dev --install-global`);
    confirm the new skill and the updated `manifest-lint.sh`/templates are live under
    `~/.claude/skills/`.

## Known Issues Triage
**Triaged:** 2026-09-15
**Tools searched:** k3s (k3s-io/k3s), Docker (moby/moby), HashiCorp Vault (hashicorp/vault)
**Result:** Historical

### Blocking Issues
- None.

### Advisory Issues
- None.

### Historical (Closed)
- [#36080 docker: invalid reference format: repository name must be lowercase](https://github.com/moby/moby/issues/36080) — `moby/moby` — closed 2018-02-26 (`kind/question`). Confirms Finding 2's lowercase-tag requirement is long-standing, documented Docker behavior, not a regression — closed as a question rather than a bug, consistent with the task's own fix being a one-line `tr` normalization rather than a workaround for a Docker defect.

### Unresolvable Tools
- GitHub Container Registry — resolve helper returned `stacksimplify/docker-hub-to-github-container-registry` (an unrelated third-party tutorial repo, not an official GHCR issue tracker); GHCR is a hosted GitHub product feature with no dedicated open-source repository of its own.
- GitHub Actions — resolve helper returned `actions/starter-workflows` (community workflow templates, not a GitHub Actions bug tracker); same exclusion this project's own downstream precedent (magnificah task 005) already applied.

### Omitted Tools
- None — three repositories searched, within the five-repository limit.

### Search Warnings
- None.

## Acceptance Criteria

- [x] A canonical `smaqit.infrastructure-image-build-push-static` (or generalized name,
      per Design Decisions) skill exists under `skills/`, with a matching
      `.smaqit/definitions/skills/` provenance file
- [x] The k3s-mode CI generation produces a build-and-push job by default (baked into
      `assets/deploy.yml.k3s.template` itself, not the reconciled skill), with the
      image reference lowercased before use — verified by generating it fresh for a
      throwaway project slug and inspecting the output, not just reading the template
- [x] `manifest-lint.sh` rejects a manifest with `runAsNonRoot: true` and no effective
      `runAsUser`, at both pod-level-only and container-level-only test cases, and still
      passes every existing case unchanged
- [x] `templates/specs/infrastructure.template.md` declares a `Container Registry`
      Constraints row (`existing-k3s` only) naming registry host + visibility
- [x] `smaqit.infrastructure-vault-loader`'s path table documents
      `organizations/<org-slug>/github-package-read`; `smaqit.infrastructure-repo-config`'s
      `existing-k3s` branch syncs it to `REGISTRY_USERNAME`/`REGISTRY_TOKEN` when the new
      Constraints row declares a private registry
- [x] `smaqit.infrastructure-cicd-generate`'s Gotchas section explains both the
      Ingress-must-be-rendered-first precondition and the `workflow_dispatch`
      default-branch requirement; `smaqit.new-greenfield-project`'s k3s sequence
      cross-references both
- [x] `smaqit.infrastructure-deploy-k3s-app`'s Examples present `workflow_dispatch` as
      the default first-deploy path, local invocation retained only as a documented
      fallback
- [x] Rebuilt and reinstalled; `~/.claude/skills/` reflects every change live

## Findings

**Implementation approach:**
- Phase A: reconciled the skill via `smaqit.create-skill` → `smaqit.L2` (definition file → compiled SKILL.md) rather than hand-authoring either file, per the plan's adjustment; verified fidelity against the downstream source and manually corrected the Provenance frontmatter afterward.
- Phase B: added a `build` job to `deploy.yml.k3s.template` (GHCR login, lowercase-tag fix, push by SHA) and extended the existing stamp step to substitute `__IMAGE__`; validated with `actionlint`/`shellcheck` and a fresh throwaway-slug generation, confirmed correct token substitution.
- Phase C: extended `manifest-lint.sh`'s `check_container_security_context` to require an effective numeric `runAsUser` wherever effective `runAsNonRoot` is true, mirroring the function's existing pod/container inheritance pattern; verified with 5 fixture manifests (pod-level, container-level, missing, boolean-guard, non-pod regression).
- Phase D: added the `Container Registry` Constraints row and MUST rule, documented the org-scoped Vault path, and sharpened `repo-config`'s sync step to name that exact path.
- Phase E: added the two Gotchas to `cicd-generate`, cross-referenced them plus the new image-build skill from `new-greenfield-project`'s k3s sequence, and reordered `deploy-k3s-app`'s Examples to CI-first.
- Phase F: bumped two hardcoded skill-count assertions in `installer/main_test.go` (29→30) after `make prepare` regenerated the embed with the new skill; `go test`/`go vet` pass; rebuilt and ran `--install-global`, verified all four changes live under `~/.claude/skills/`.

**Decisions made:**
- Kept the skill name `-static`-suffixed (no second real non-static example to generalize from yet).
- Build-and-push CI wiring lives in the generic template, not the skill — discovered during planning: `assets/deployment.yaml.template` already carried an unfilled `__IMAGE__` token, confirming the gap was template-level, not skill-level.
- Added an Infrastructure-spec `Container Registry` Constraints row (new scope, not in the original task text) since `repo-config`'s existing "if the spec names a private registry" condition had no field to key off.
- No new test harness for `manifest-lint.sh` — verified via throwaway fixtures, consistent with no skill script in this repo having automated tests today.
- `organizations/<org-slug>/github-package-read` is deliberately excluded from `rotate-credential.sh`'s supported paths (org-scoped, not app/machine-scoped) — rotation is the same manual `vault kv put`/`delete` used to populate it.

**Blockers encountered:**
- `smaqit.create-skill`'s documented output paths (`.agents/skills/`, `.claude/skills/`) are the convention for a downstream consumer project, not this repo — this repo has neither directory and its own canonical skill source is `skills/<name>/SKILL.md`. Resolved by invoking `smaqit.L2` directly against a definition file placed at this repo's own convention instead of following `smaqit.create-skill`'s literal output-path instructions.
- The generic `smaqit.create-skill` validator (`validate-skill.go`) flagged a "Use when..." description opening as an anti-pattern; confirmed two sibling canonical skills already fail the same rule, so kept the description consistent with its siblings rather than making it an outlier.

**Follow-up identified:**
- Whether to broaden `smaqit.infrastructure-image-build-push-static` beyond a static site with no build step of its own remains open, per the task's own Notes — resolve once a second real (non-static) use case exists.
- `rotate-credential.sh` does not support `organizations/<org-slug>/github-package-read` — acceptable today since rotation is rare and manual, but worth adding if this credential type starts needing frequent rotation.

## Files to Create / Modify

| File | Action |
|------|--------|
| `skills/smaqit.infrastructure-image-build-push-static/` (or generalized name) | Create |
| `.smaqit/definitions/skills/smaqit.infrastructure-image-build-push-static.md` | Create |
| `skills/smaqit.infrastructure-cicd-generate/SKILL.md`, `assets/deploy.yml.k3s.template` | Modify |
| `skills/smaqit.infrastructure-deploy-k3s-app/scripts/manifest-lint.sh`, `SKILL.md` | Modify |
| `skills/smaqit.infrastructure-vault-loader/SKILL.md` | Modify |
| `skills/smaqit.infrastructure-repo-config/SKILL.md` | Modify |
| `skills/smaqit.new-greenfield-project/SKILL.md` | Modify |
| `templates/specs/infrastructure.template.md`, `agents/infrastructure.md` | Modify |

## Notes

Source material: a downstream project's task 005 (`.smaqit/tasks/005_phase4_dev_environment_sweep.md`)
and its synthesized project-local skill. Reconciliation pattern mirrors task 106 (Python/Tornado
rsync skill reconciliation). Whether to broaden the image-build-push skill beyond
"static site, no build step of its own" is an open generalization question — resolve it
during implementation once the exact scope of a second real (non-static) use case is
clearer, or leave it explicitly static-scoped and record the broader case as follow-up.
