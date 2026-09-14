---
status: Not Started
created: "2026-09-15"
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

## Implementation Steps

1. Read the downstream project's task 005 file and its synthesized skill in full (see
   Description) before writing anything — this task generalizes proven content, it does
   not re-derive it from scratch.
2. Reconcile the image-build-push skill into `skills/` (Finding 1), including its
   Dockerfile-authoring guidance, the lowercase-tag fix baked into the generated build
   step (Finding 2), and a `.smaqit/definitions/skills/` provenance file matching the
   existing convention (see `smaqit.infrastructure-deploy-k3s-app`'s own definition file
   for the shape).
3. Extend `smaqit.infrastructure-cicd-generate`'s `k3s`-mode template/generation logic so
   a generated `deploy.yml` includes the build-and-push job by default (or via a clearly
   documented opt-in, if a case exists for a k3s app that doesn't need one) — do not
   leave this as a separate, easy-to-forget manual step for the operator to remember to
   add.
4. Extend `manifest-lint.sh` for Finding 3: fail with a named-field error (matching the
   script's existing error style) when the effective `runAsNonRoot` is `true` and no
   effective numeric `runAsUser` is set at either pod or container level.
5. Add the `organizations/<org-slug>/github-package-read` path to
   `smaqit.infrastructure-vault-loader`'s documented path table and prompt flow
   (Finding 4); extend `smaqit.infrastructure-repo-config`'s `existing-k3s` branch to
   sync it to `REGISTRY_USERNAME`/`REGISTRY_TOKEN` per environment when a private
   registry is declared.
6. Add Finding 5's and Finding 6's explanations to `smaqit.infrastructure-cicd-generate`'s
   own Gotchas section, and cross-reference both from `smaqit.new-greenfield-project`'s
   k3s Phase 4/5 sequence text.
7. Reframe `smaqit.infrastructure-deploy-k3s-app`'s Examples/Steps per Finding 7:
   `workflow_dispatch` first, local invocation kept as a documented fallback.
8. Rebuild and reinstall (`cd installer && make build && ./dist/smaqit-dev --install-global`)
   and verify the new skill and lint behavior are live.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] A canonical `smaqit.infrastructure-image-build-push-static` (or generalized name,
      per Design Decisions) skill exists under `skills/`, with a matching
      `.smaqit/definitions/skills/` provenance file
- [ ] The k3s-mode CI generation produces a build-and-push job by default, with the
      image reference lowercased before use — verified by generating it fresh for a
      throwaway project slug and inspecting the output, not just reading the template
- [ ] `manifest-lint.sh` rejects a manifest with `runAsNonRoot: true` and no effective
      `runAsUser`, at both pod-level-only and container-level-only test cases, and still
      passes every existing case unchanged
- [ ] `smaqit.infrastructure-vault-loader`'s path table documents
      `organizations/<org-slug>/github-package-read`; `smaqit.infrastructure-repo-config`'s
      `existing-k3s` branch syncs it to `REGISTRY_USERNAME`/`REGISTRY_TOKEN` when a
      private registry is declared
- [ ] `smaqit.infrastructure-cicd-generate`'s Gotchas section explains both the
      Ingress-must-be-rendered-first precondition and the `workflow_dispatch`
      default-branch requirement; `smaqit.new-greenfield-project`'s k3s sequence
      cross-references both
- [ ] `smaqit.infrastructure-deploy-k3s-app`'s Examples present `workflow_dispatch` as
      the default first-deploy path, local invocation retained only as a documented
      fallback
- [ ] Rebuilt and reinstalled; `~/.claude/skills/` reflects every change live

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
| `skills/smaqit.infrastructure-image-build-push-static/` (or generalized name) | Create |
| `.smaqit/definitions/skills/smaqit.infrastructure-image-build-push-static.md` | Create |
| `skills/smaqit.infrastructure-cicd-generate/SKILL.md`, `assets/deploy.yml.k3s.template` | Modify |
| `skills/smaqit.infrastructure-deploy-k3s-app/scripts/manifest-lint.sh`, `SKILL.md` | Modify |
| `skills/smaqit.infrastructure-vault-loader/SKILL.md` | Modify |
| `skills/smaqit.infrastructure-repo-config/SKILL.md` | Modify |
| `skills/smaqit.new-greenfield-project/SKILL.md` | Modify |

## Notes

Source material: a downstream project's task 005 (`.smaqit/tasks/005_phase4_dev_environment_sweep.md`)
and its synthesized project-local skill. Reconciliation pattern mirrors task 106 (Python/Tornado
rsync skill reconciliation). Whether to broaden the image-build-push skill beyond
"static site, no build step of its own" is an open generalization question — resolve it
during implementation once the exact scope of a second real (non-static) use case is
clearer, or leave it explicitly static-scoped and record the broader case as follow-up.
