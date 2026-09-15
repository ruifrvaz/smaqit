---
name: smaqit.infrastructure-image-build-push-static
description: Use when containerizing a static HTML/CSS/vanilla-JS site (no build step of its own) for deployment onto an `existing-k3s` platform. Authors the `Dockerfile`/`.dockerignore` (`nginxinc/nginx-unprivileged:1-alpine` base, baked `health.json` with the deploy SHA) and the `deployment/k3s/deployment.yaml`/`service.yaml` manifest pair, setting an explicit numeric `runAsUser` matching the base image's real UID — without it the kubelet rejects the pod at admission with a non-numeric-user `CreateContainerConfigError`. Adds `imagePullSecrets` when the image is private. Fills the Dockerfile/manifest-authoring gap between `smaqit.infrastructure-cicd-generate`'s `k3s` template (builds and pushes the image, authors no application content) and `smaqit.infrastructure-deploy-k3s-app` (assumes manifests already exist). Does not render the Ingress, touch CI/CD, or manage registry credentials.
metadata:
  version: "1.0.0"
---

# Containerize a Static Site for k3s Deployment

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
   securityContext block, explicit resource requests/limits), with one addition:
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
  narrow to a static site with no build step.
- Does not decide whether the image should be public or private — that is a per-project call;
  this skill covers both paths (public: skip the `imagePullSecrets` field; private: include it).

## Examples

**Input:** A static site project (`provisioning_mode: existing-k3s`) has run
`smaqit.infrastructure-cicd-generate` in `k3s` mode (which already builds and pushes the image),
but has no Dockerfile and no application manifests yet.

**Output:** `Dockerfile`/`.dockerignore` authored (`nginxinc/nginx-unprivileged:1-alpine`, baked
`/health.json`); `deployment/k3s/deployment.yaml`/`service.yaml` authored with explicit
`runAsUser: 101` and, if the image is private, `imagePullSecrets: - name: app-registry`.

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
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| Output artifact already exists | Confirm with user before overwriting |
| `docker build`'s health-file `RUN` step fails with a permission error | The base image's default non-root user can't write into the target directory after `COPY` (which always writes as root) — add `USER root` before the `RUN`, then switch back to the image's own non-root user afterward. |
| Pod stuck in `CreateContainerConfigError`, message mentions "non-numeric user" | Set `runAsUser` explicitly in the securityContext to the base image's real numeric UID (`docker run --rm <image> id`) — never leave `runAsNonRoot: true` alone with a named-user base image. |
| Pod stuck in `ImagePullBackOff` | Check whether the registry package is private (GHCR inherits the source repo's visibility by default) — either make it public or add `imagePullSecrets` to the manifest, backed by a valid, tested pull credential from `smaqit.infrastructure-vault-loader`'s shared registry-read path. |
| `manifest-lint.sh` fails with `FileNotFoundError` on `ingress.yaml` | Expected until the Ingress has actually been rendered once — render and commit it (via `smaqit.infrastructure-deploy-k3s-app`) before the workflow can pass, never fabricate a placeholder host. |

## Allowed Tools

Bash(docker:*), Bash(kubectl:*), Read, Write, Edit
