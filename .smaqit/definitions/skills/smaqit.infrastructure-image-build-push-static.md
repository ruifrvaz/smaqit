# smaqit.infrastructure-image-build-push-static

## Description

Containerizes a static HTML/CSS/JS site (no build step of its own) for deployment onto an
`existing-k3s` platform: authors the `Dockerfile`/`.dockerignore` and the
`deployment.yaml`/`service.yaml` manifest pair, baking in the one fact a k3s target requires that
a generic container guide would not — an explicit numeric `runAsUser` matching the base image's
real UID, without which the kubelet rejects the pod at admission time. Fills exactly the
Dockerfile/manifest-authoring gap between `smaqit.infrastructure-cicd-generate`'s `k3s` template
(which builds and pushes the image generically, but authors no application content) and
`smaqit.infrastructure-deploy-k3s-app` (which assumes Deployment/Service manifests already exist
locally).

## Provenance

- `synthesized: true`
- `contributed-for-project: magnificah-website`
- `contributed-date: 2026-09-14`
- `contributed-stack`: static HTML/CSS/vanilla-JS site, no build step, deployed as a container
  onto an `existing-k3s` Namespace-per-app platform (Traefik ingress, cert-manager, GHCR image
  registry)
- **Origin**: task 005 (K3s app onboarding & containerized dev deployment on magnificah-test-01)
  confirmed via direct audit that `smaqit.infrastructure-cicd-generate`'s `k3s` mode and
  `smaqit.infrastructure-deploy-k3s-app` had zero coverage for turning a static site into a
  running container image — no Dockerfile authoring, no manifest authoring, no `runAsUser`
  guidance. Every fact in this skill was proven live against a real cluster (`magnificah-test-01`)
  across six iterations, not guessed. Reconciled into canonical smaqit as task 121 (mirroring task
  106's reconciliation precedent), which narrowed the original downstream skill's scope: the
  build-and-push CI job and the image-pull-secret workflow step both moved to
  `smaqit.infrastructure-cicd-generate`'s `deploy.yml.k3s.template` itself (stack-agnostic, so
  every k3s project gets them for free), leaving this skill scoped to the Dockerfile/manifest
  authoring that still needs stack judgment.

## Required-inherited-context

Reuses these existing k3s-family conventions rather than reinventing them:
- The Namespace is always `app-__APP_SLUG__`; manifests live at `deployment/k3s/deployment.yaml`
  and `deployment/k3s/service.yaml` as separate files — `smaqit.infrastructure-cicd-generate`'s
  generated `deploy.yml`'s `kubectl apply`/`manifest-lint.sh` calls reference them individually,
  not as a combined manifest.
- The Ingress is never hand-authored by this skill — it is rendered once from
  `smaqit.infrastructure-deploy-k3s-app`'s own `assets/ingress.yaml.template`, using the
  Infrastructure spec's declared Ingress Class/`ClusterIssuer`/host, and checked in at
  `deployment/k3s/ingress.yaml`. No standalone `certificate.yaml` — cert-manager's ingress-shim
  creates that resource automatically from the Ingress's `cert-manager.io/cluster-issuer`
  annotation.
- The image build-and-push job (login, lowercase-tag, push by commit SHA, `__IMAGE__`
  substitution) and the conditional `imagePullSecret` creation step both already live in
  `smaqit.infrastructure-cicd-generate`'s `assets/deploy.yml.k3s.template` — this skill does not
  author or duplicate either.
- `DEPLOY_SHA`/`DEPLOY_TIME` are stamped into `deployment.yaml` as pod environment by CI (`sed`,
  at apply time) — never a file, since this credential has no host filesystem.

## Steps

1. **Author the Dockerfile** at the repository root, using `nginxinc/nginx-unprivileged:1-alpine`
   as the base image — not plain `nginx:alpine`. The unprivileged variant listens on 8080 as a
   non-root user (`uid=101`) out of the box, matching the platform's Pod Security `restricted`
   requirement without any custom `nginx.conf`; plain `nginx:alpine` runs as root on port 80 and
   fails `manifest-lint.sh`'s non-root check outright.
   - `COPY` the already-built static site directory verbatim into `/usr/share/nginx/html/` — no
     `RUN` step that compiles, transpiles, bundles, or otherwise transforms it.
   - Switch to `USER root` before writing the baked health file (`COPY` always writes as root,
     but `RUN` respects the image's current `USER`, and the unprivileged image's default non-root
     user cannot write into `/usr/share/nginx/html/`), then back to `USER nginx` afterward so the
     image's own default runtime user stays non-root.
   - Bake a `health.json` file with the deploy commit SHA templated in via a `DEPLOY_SHA` build
     `ARG` at image-build time — never at container-start time, and never checked into the site's
     own versioned source tree. Distinct from the `DEPLOY_SHA` pod-environment stamp CI sets at
     apply time (both exist and serve different purposes).
   - Author a `.dockerignore` excluding non-runtime sources (`.git`, `.smaqit`, `.github`, `docs`,
     `specs`, `content`, `assets/raw`, test directories, markdown files).

2. **Author `deployment/k3s/deployment.yaml`** (Deployment) and a companion
   `deployment/k3s/service.yaml`, matching `smaqit.infrastructure-deploy-k3s-app`'s own
   `assets/deployment.yaml.template` shape (`containerPort: 8080`, the full `restricted`
   securityContext block, explicit resource requests/limits), with one addition this skill's own
   hard-won fact requires:
   - **Set `runAsUser` explicitly** (the base image's known numeric UID — `101` for
     `nginxinc/nginx-unprivileged`) in the pod-level `securityContext`. Without it, the kubelet
     throws `CreateContainerConfigError`: *"container has runAsNonRoot and image has non-numeric
     user (nginx), cannot verify user is non-root"* — the image's own `USER` directive names a
     user (`nginx`), not a numeric UID, and the kubelet refuses to run the container just to
     resolve that name. `docker run` locally never surfaces this; it is a Kubernetes-only
     admission-time check. Confirm the base image's real UID first
     (`docker run --rm <image> id`) rather than assuming `101` for a different base image.
   - If the image is pushed to a **private** registry (the default for GHCR packages pushed from
     a private repository — visibility is inherited from the repo, not opt-in), also add
     `imagePullSecrets: - name: app-registry` to the pod spec — matching the Secret that
     `deploy.yml.k3s.template`'s own (already-generic) commented-out step creates when uncommented.

## Output

- `Dockerfile`, `.dockerignore` at the repository root
- `deployment/k3s/deployment.yaml`, `deployment/k3s/service.yaml`
- A working, verified container image, pullable and runnable inside the target Namespace's Pod
  Security `restricted` constraints

## Scope

- Does not render or author the Ingress — that remains `smaqit.infrastructure-deploy-k3s-app`'s
  own template-rendering step.
- Does not author or modify any CI/CD workflow — the build-and-push job and the conditional
  `imagePullSecret` step both live in `smaqit.infrastructure-cicd-generate`'s
  `deploy.yml.k3s.template`.
- Does not create or manage the shared registry-read credential's lifecycle — provisioning/
  rotating it is an operator action via `smaqit.infrastructure-vault-loader`/
  `smaqit.infrastructure-repo-config`.
- Does not cover non-static stacks (a backend with its own build step) — scoped deliberately
  narrow to a static site with no build step, per explicit user direction when this skill was
  originally synthesized.
- Does not decide whether the image should be public or private — that is a per-project call;
  this skill covers both paths (public: skip the `imagePullSecrets` field; private: include it).

## Completion

- [ ] `docker build` succeeds locally and the container runs as the expected non-root UID
      (`docker run --rm --user <uid> ...`)
- [ ] `/health.json` (or whatever path the project's own contract defines) returns the correct
      baked SHA
- [ ] `manifest-lint.sh` passes against `deployment.yaml` and `service.yaml`
- [ ] The Deployment's pod spec sets an explicit numeric `runAsUser` matching the base image's
      real UID
- [ ] If private: `imagePullSecrets: - name: app-registry` is present in the pod spec

## Failure Handling

| Situation | Action |
|-----------|--------|
| `docker build`'s health-file `RUN` step fails with a permission error | The base image's default non-root user can't write into the target directory after `COPY` (which always writes as root) — add `USER root` before the `RUN`, then switch back to the image's own non-root user afterward. |
| Pod stuck in `CreateContainerConfigError`, message mentions "non-numeric user" | Set `runAsUser` explicitly in the securityContext to the base image's real numeric UID (`docker run --rm <image> id`) — never leave `runAsNonRoot: true` alone with a named-user base image. |
| Pod stuck in `ImagePullBackOff` | Check whether the registry package is private (GHCR inherits the source repo's visibility by default) — either make it public or add `imagePullSecrets` to the manifest, backed by a valid, tested pull credential from `smaqit.infrastructure-vault-loader`'s shared registry-read path. |
| `manifest-lint.sh` fails with `FileNotFoundError` on `ingress.yaml` | Expected until the Ingress has actually been rendered once (Required-inherited-context) — render and commit it before the workflow can pass, never fabricate a placeholder host. |

## Gotchas

- `docker run` locally will never catch the `runAsUser`/named-user kubelet check — it is
  Kubernetes-admission-specific. Don't trust a clean local `docker run` as proof the manifest will
  schedule cleanly.
- GHCR package visibility silently inherits the source repository's visibility — a private repo
  produces a private package by default, with no explicit opt-in step, which then blocks the
  cluster's pull unless `imagePullSecrets` is added.
- The build-and-push CI job and the `imagePullSecret` workflow step are **not** this skill's job —
  they already live generically in `smaqit.infrastructure-cicd-generate`'s
  `deploy.yml.k3s.template`. Do not re-author either here even if a downstream project's earlier
  synthesized version of this skill included them.

## Allowed Tools

Bash(docker:*), Bash(kubectl:*), Read, Write, Edit

## Examples

**Input:** A static site project (`provisioning_mode: existing-k3s`) has run
`smaqit.infrastructure-cicd-generate` in `k3s` mode (which already builds and pushes the image),
but has no Dockerfile and no application manifests yet.

**Output:** `Dockerfile`/`.dockerignore` authored (`nginxinc/nginx-unprivileged:1-alpine`, baked
`/health.json`); `deployment/k3s/deployment.yaml`/`service.yaml` authored with explicit
`runAsUser: 101` and, if the image is private, `imagePullSecrets: - name: app-registry`.
